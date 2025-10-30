package main

import (
    "context"
    "flag"
    "fmt"   
    "net"
    "os"
    "time"
    "log"

    "github.com/balaji-balu/margo-hello-world/ent"
    "github.com/balaji-balu/margo-hello-world/internal/api"
    cfffg "github.com/balaji-balu/margo-hello-world/internal/config"
    fsmloader "github.com/balaji-balu/margo-hello-world/internal/fsm"

    "google.golang.org/grpc"
    "entgo.io/ent/dialect"
    "entgo.io/ent/dialect/sql"
    _ "github.com/lib/pq"  // enables the 'postgres' driver
    "github.com/joho/godotenv"


// telemetry
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
    "go.opentelemetry.io/otel/sdk/resource"
    "go.opentelemetry.io/otel/sdk/trace"
    "go.opentelemetry.io/otel/semconv/v1.12.0"
)

func InitTelemetry(ctx context.Context, serviceName string) func(context.Context) error {
    exporter, err := otlptracegrpc.New(ctx,
        otlptracegrpc.WithInsecure(),
        otlptracegrpc.WithEndpoint("localhost:4317"),
    )
    if err != nil {
        log.Fatalf("Failed to create trace exporter: %v", err)
    }
    tp := trace.NewTracerProvider(
        trace.WithBatcher(exporter, trace.WithBatchTimeout(time.Second)),
        trace.WithResource(resource.NewWithAttributes(
            semconv.SchemaURL,
            semconv.ServiceNameKey.String(serviceName),
        )),
    )
    otel.SetTracerProvider(tp)
    return tp.Shutdown
}


func init() {
    err := godotenv.Load("../../.env") // relative path to project root
    if err != nil {
        log.Println("No .env file found, reading from system environment")
    }
}

func main() {
    ctx := context.Background()
    shutdown := InitTelemetry(ctx, "orchestrator")
    defer shutdown(ctx)

    //config := flag.String("config", "", "config file")
    //node := flag.String("node", "", "node name")

    configPath := flag.String("config", "./configs/co.yaml", "path to config file")
	flag.Parse()

	cfg, err := cfffg.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	log.Printf("✅ Loaded config: site=%s, port=%d, CO URL=%s",
		cfg.Server.Site, cfg.Server.Port, cfg.CO.URL)

    
    grpcPort := flag.String("grpc", ":50051", "CO gRPC listen address")
    //httpPort := flag.String("http", ":8080", "CO HTTP listen address")
    //loAddr := flag.String("lo", "localhost:50052", "Local Orchestrator address")
    flag.Parse()

    //log.Println("config:", *config, "node:", node)
    
    dsn := os.Getenv("DATABASE_URL")
    // if dsn == "" {
    //     dsn = "postgres://postgres:postgres@localhost:5432/orchestration?sslmode=disable"
    // }


	machine, err := fsmloader.LoadFSM("./configs/fsm.yaml", "CO")
	if err != nil {
		log.Fatalf("failed to load FSM: %v", err)
	}

	fmt.Println("CO initial:", machine.Current())
	_ = machine.Event(ctx, "send_request", )
	_ = machine.Event(ctx, "complete")
	_ = machine.Event(ctx, "reset")

    fmt.Println("[CO] connecting to postgres at ", dsn)
    drv, err := sql.Open(dialect.Postgres, dsn)
    if err != nil {
        log.Fatalf("failed connecting to postgres: %v", err)
    }
    client := ent.NewClient(ent.Driver(drv))
    defer client.Close()

    router := api.NewRouter(client)
    log.Println("CO API running on :", cfg.Server.Port)
    if err := router.Run(fmt.Sprintf(":%d", cfg.Server.Port)); err != nil {
        log.Fatal(err)
    }

    // Start gRPC server for callbacks from LO
    go func() {
        lis, err := net.Listen("tcp", *grpcPort)
        if err != nil {
            log.Fatalf("[CO] failed to listen: %v", err)
        }
        s := grpc.NewServer()
        //pb.RegisterCentralOrchestratorServer(s, &server{})
        log.Printf("[CO] gRPC listening on %s", *grpcPort)
        if err := s.Serve(lis); err != nil {
            log.Fatalf("[CO] serve: %v", err)
        }
    }()

/*
    appDesc, err := application.ParseFromFile("../tests/app1.yaml")
    if err != nil {
        log.Fatal(err)
        //return
    }

    if err == nil {
        //fmt.Println("App:", appDesc.Metadata.Name)
        //fmt.Println("Catalog site:", appDesc.Metadata.Catalog.Application.Site)
        //fmt.Println("Deployment profiles:", len(appDesc.DeploymentProfiles))
        if err := Persist(ctx, client, appDesc); err != nil {
            log.Fatal(err)
        }
    }


    appDesc, err = application.ParseFromFile("../tests/app2.yaml")
    if err != nil {
        log.Fatal(err)
        //return
    }
    if err == nil {
        //fmt.Println("App:", appDesc.Metadata.Name)
        //fmt.Println("Catalog site:", appDesc.Metadata.Catalog.Application.Site)
        //fmt.Println("Deployment profiles:", len(appDesc.DeploymentProfiles))
        if err := Persist(ctx, client, appDesc); err != nil {
            log.Fatal(err)
        }
    }



    http.HandleFunc("/app/install", func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPost {
            http.Error(w, "POST required", http.StatusMethodNotAllowed)
            return
        }
    }

    http.HandleFunc("/app/all", func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodGet {
            http.Error(w, "GET required", http.StatusMethodNotAllowed)
            return
        }
    }
    // REST endpoint to trigger deployment
    http.HandleFunc("/deploy", func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPost {
            http.Error(w, "POST required", http.StatusMethodNotAllowed)
            return
        }
        // Option 1: via query param ?file=fleet.yaml
        yamlPath := r.URL.Query().Get("file")

        // Option 2: upload directly as body
        if yamlPath == "" {
            tmpFile := "uploaded-fleet.yaml"
            data, err := io.ReadAll(r.Body)
            if err != nil {
                http.Error(w, "failed to read body", http.StatusBadRequest)
                return
            }
            if err := os.WriteFile(tmpFile, data, 0644); err != nil {
                http.Error(w, "failed to write file", http.StatusInternalServerError)
                return
            }
            yamlPath = tmpFile
        }

        log.Printf("[CO] REST deploy request: %s", yamlPath)
        if err := deployFleet(ctx, yamlPath, *loAddr); err != nil {
            log.Printf("[CO] deploy failed: %v", err)
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        w.WriteHeader(http.StatusOK)
        fmt.Fprintf(w, "Deployment triggered for %s\n", yamlPath)
    })

    go func() {
        log.Printf("[CO] HTTP listening on %s", *httpPort)
        if err := http.ListenAndServe(*httpPort, nil); err != nil {
            log.Fatalf("[CO] http: %v", err)
        }
    }()

    // Wait for Ctrl+C
    sig := make(chan os.Signal, 1)
    signal.Notify(sig, os.Interrupt)
    <-sig
    log.Println("[CO] shutting down")

*/
}




// deployment profile: DeploymentProfiles