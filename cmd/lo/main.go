// file: lo/main.go
package main

import (
	"context"
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"google.golang.org/grpc"
	pb "github.com/balaji/hello/proto_generated"

	"github.com/balaji/hello/pkg/deployment"
	"fmt"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	_ "github.com/lib/pq"  // enables the 'postgres' driver
	"github.com/balaji/hello/ent"
	"github.com/joho/godotenv"
	"net/http"
	"bytes"
	"encoding/json"

	"github.com/balaji/hello/internal/orchestrator"
)

type server struct {
	pb.UnimplementedLocalOrchestratorServer

	nodeAddr map[string]string       // nodeID -> address
	enConns  map[string]*grpc.ClientConn
	enClients map[string]pb.EdgeNodeClient

	statuses map[string]*pb.StatusReport // deploymentID -> merged statuses
	mu       sync.Mutex

	coAddr   string
	coConn   *grpc.ClientConn
	coClient pb.CentralOrchestratorClient
}

func init() {

    err:= godotenv.Load("./.env") // relative path to project root
    if err != nil {
        log.Println("No .env file found, reading from system environment")
    }
}

func newServer(nodeAddr map[string]string, coAddr string) *server {
	return &server{
		nodeAddr:  nodeAddr,
		enConns:   make(map[string]*grpc.ClientConn),
		enClients: make(map[string]pb.EdgeNodeClient),
		statuses:  make(map[string]*pb.StatusReport),
		coAddr:    coAddr,
	}
}

// initialize persistent gRPC connections
func (s *server) initClients() error {
	for id, addr := range s.nodeAddr {
		conn, err := grpc.Dial(addr, grpc.WithInsecure())
		if err != nil {
			return err
		}
		s.enConns[id] = conn
		s.enClients[id] = pb.NewEdgeNodeClient(conn)
		log.Printf("LO: connected to EN %s at %s", id, addr)
	}

	connCo, err := grpc.Dial(s.coAddr, grpc.WithInsecure())
	if err != nil {
		return err
	}
	s.coConn = connCo
	s.coClient = pb.NewCentralOrchestratorClient(connCo)
	log.Printf("LO: connected to CO at %s", s.coAddr)
	return nil
}

func (s *server) ReceiveDeploy(ctx context.Context, req *pb.DeployRequest) (*pb.DeployResponse, error) {
	log.Printf("LO: ReceiveDeploy id=%s fleet=%s", req.DeploymentId, req.Fleet.Name)

	var wg sync.WaitGroup
	for _, node := range req.Fleet.Nodes {
		client, ok := s.enClients[node.Id]
		if !ok {
			log.Printf("LO: unknown node %s", node.Id)
			continue
		}

		per := &pb.DeployRequest{
			DeploymentId: req.DeploymentId,
			Fleet: &pb.Fleet{
				Name:  req.Fleet.Name,
				Nodes: []*pb.NodeSpec{node},
			},
		}

		wg.Add(1)
		go func(id string, c pb.EdgeNodeClient, req *pb.DeployRequest) {
			defer wg.Done()
			ctx2, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()
			resp, err := c.ApplyDeploy(ctx2, req)
			if err != nil {
				log.Printf("LO: apply deploy to EN %s err: %v", id, err)
				return
			}
			log.Printf("LO: EN %s responded: %s", id, resp.Message)
		}(node.Id, client, per)
	}

	wg.Wait()
	return &pb.DeployResponse{DeploymentId: req.DeploymentId, Message: "deployment forwarded to edges"}, nil
}

func (s *server) ReportFromEdge(ctx context.Context, rpt *pb.StatusReport) (*pb.DeployResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	existing, ok := s.statuses[rpt.DeploymentId]
	if !ok {
		// first report
		s.statuses[rpt.DeploymentId] = rpt
	} else {
		// merge new statuses
		existing.Statuses = append(existing.Statuses, rpt.Statuses...)
	}

	go s.forwardToCO(s.statuses[rpt.DeploymentId])
	return &pb.DeployResponse{DeploymentId: rpt.DeploymentId, Message: "status stored"}, nil
}

func (s *server) forwardToCO(rpt *pb.StatusReport) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	resp, err := s.coClient.ReportStatus(ctx, rpt)
	if err != nil {
		log.Printf("LO: forward to CO err: %v", err)
		return
	}
	log.Printf("LO: forwarded status to CO: %s", resp.Message)
}

func (s *server) shutdown() {
	log.Println("LO: shutting down...")
	for _, conn := range s.enConns {
		conn.Close()
	}
	if s.coConn != nil {
		s.coConn.Close()
	}
}


func (s *server) handleStatus(w http.ResponseWriter, r *http.Request) {
	var report deployment.DeploymentReport
	if err := json.NewDecoder(r.Body).Decode(&report); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	log.Printf("[LO] Status from agent: %s - %s", report.AppName, report.Status)

	// Forward to CO
	body, _ := json.Marshal(report)
	http.Post("http://central-orchestrator:8082/status", "application/json", bytes.NewReader(body))
}

func initDB() *ent.Client {
	dsn := os.Getenv("DATABASE_URL")
    // if dsn == "" {
    //     dsn = "postgres://postgres:postgres@localhost:5432/orchestration?sslmode=disable"
    // }

    fmt.Println("[CO] connecting to postgres at ", dsn)
    drv, err := sql.Open(dialect.Postgres, dsn)
    if err != nil {
        log.Fatalf("failed connecting to postgres: %v", err)
    }
    client := ent.NewClient(ent.Driver(drv))
    defer client.Close()

	return client
}
func main() {
	port := flag.String("port", ":50052", "listen address")
	coAddr := flag.String("co", "localhost:50051", "central orchestrator address")
	config := flag.String("config", "", "config file")
    site := flag.String("site", "", "site name")
	eoport := flag.String("eoport", ":8080", "edge orchestrator port")
	flag.Parse()

	log.Println("config:", *config, "site:", *site)

	nodeAddr := map[string]string{
		"edge1": "localhost:50054",
		"edge2": "localhost:50055",
	}


	// cfg, err := config.LoadConfig("./config.yaml")
	// if err != nil {
	// 	log.Fatal("failed to load config:", err)
	// }

	//fmt.Println("Starting Local Orchestrator with trigger:", cfg.Trigger.Type)

	lo := orchestrator.NewLocalOrchestrator()
	lo.EOPort = *eoport

	//lo.Client = initDB()


	// ✅ Context + signal handling
	ctx, cancel := context.WithCancel(context.Background())
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-stop
		log.Println("LO: shutting down...")
		cancel()
	}()

	// ✅ Start trigger with context
	// trg := trigger.New(cfg)
	// go func() {
	// 	if err := trg.Start(ctx); err != nil {
	// 		log.Printf("trigger stopped with error: %v", err)
	// 	}
	// }()

	for {
		select {
    		case <-ctx.Done():
        		log.Println("LO: main loop exiting...")
        		return // gracefully exit main()
			default:
				mode := lo.DetectMode()
				switch mode {
				case orchestrator.PushPreferred:
					lo.WaitForWebhook(ctx)
				case orchestrator.AdaptivePull:
					lo.Poll(ctx)
				case orchestrator.OfflineDeterministic:
					lo.ScanLocalInbox(ctx)
				default:
					log.Println("Unknown mode, defaulting to AdaptivePull")
					lo.Poll(ctx)
				}
		}

        lo.PersistJournal()
        time.Sleep(5 * time.Second)
    }

	// ✅ Start gRPC server
	lis, err := net.Listen("tcp", *port)
	if err != nil {
		log.Fatalf("LO: listen error: %v", err)
	}

	srv := grpc.NewServer()
	server := newServer(nodeAddr, *coAddr)
	if err := server.initClients(); err != nil {
		log.Fatalf("LO: init clients error: %v", err)
	}

	pb.RegisterLocalOrchestratorServer(srv, server)

	go func() {
		log.Printf("LO listening on %s", *port)
		if err := srv.Serve(lis); err != nil {
			log.Fatalf("LO: serve error: %v", err)
		}
	}()

	// ✅ Wait for cancel (Ctrl+C)
	<-ctx.Done()

	// ✅ Graceful shutdown
	log.Println("LO: stopping gRPC server...")
	srv.GracefulStop()
	server.shutdown()

	time.Sleep(1 * time.Second)
	log.Println("LO: exited cleanly")
}