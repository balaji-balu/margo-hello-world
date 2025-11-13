package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/balaji-balu/margo-hello-world/ent"
	"github.com/balaji-balu/margo-hello-world/internal/api"
	"github.com/balaji-balu/margo-hello-world/internal/config"
	//"github.com/balaji-balu/margo-hello-world/internal/fsmloader"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	_ "github.com/lib/pq" // enables the 'postgres' driver
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"go.uber.org/zap"

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
	err := godotenv.Load("./.env") // relative path to project root
	if err != nil {
		log.Println("No .env file found, reading from system environment")
	}
}

func main() {
	ctx := context.Background()
	shutdown := InitTelemetry(ctx, "orchestrator")
	defer shutdown(ctx)
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err) // Handle any potential error
	}

	// Print the current working directory
	fmt.Println("Current working directory:", dir)

	entries, err := os.ReadDir(".")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Contents of the current directory:")
	for _, entry := range entries {
		// entry.Name() gets the name of the file or directory
		// entry.IsDir() returns true if it is a directory
		fmt.Printf("- %s (Directory: %t)\n", entry.Name(), entry.IsDir())
	}
	configPath := flag.String("config", "./configs/co.yaml", "path to config file")
	flag.Parse()

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("✅ Loaded config: site=%s, port=%d",
		cfg.Server.Site, port)

    // Use NewProduction() for JSON, performance, and sampled logging
    logger, err := zap.NewProduction()
    if err != nil {
        // If Zap fails, fall back to standard log or panic
        log.Fatalf("can't initialize zap logger: %v", err)
    }
    defer logger.Sync() // Ensure all buffered logs are written

	// Inside main()
	// Redirect all calls from the standard library's 'log' package to Zap.
	// This is the single most important step for converting existing logs immediately.
	zap.RedirectStdLog(logger)

	grpcPort := flag.String("grpc", ":50051", "CO gRPC listen address")
	//httpPort := flag.String("http", ":8080", "CO HTTP listen address")
	//loAddr := flag.String("lo", "localhost:50052", "Local Orchestrator address")
	flag.Parse()

	//log.Println("config:", *config, "node:", node)

	//dsn := os.Getenv("DATABASE_URL")
	// if dsn == "" {
	//     dsn = "postgres://postgres:postgres@localhost:5432/orchestration?sslmode=disable"
	// }

	// machine, err := fsmloader.LoadFSM("./configs/fsm.yaml", "CO")
	// if err != nil {
	// 	log.Fatalf("failed to load FSM: %v", err)
	// }

	// fmt.Println("CO initial:", machine.Current())
	// _ = machine.Event(ctx, "send_request", )
	// _ = machine.Event(ctx, "complete")
	// _ = machine.Event(ctx, "reset")
	dsn := os.Getenv("DATABASE_URL")
	fmt.Println("[CO] connecting to postgres at", dsn)

	var drv *sql.Driver
	var err1 error
	for i := 1; i <= 10; i++ {
		drv, err1 = sql.Open(dialect.Postgres, dsn)
		if err1 == nil {
			if err1 = drv.DB().Ping(); err1 == nil {
				fmt.Println("✅ Connected to Postgres")
				break
			}
		}
		fmt.Printf("⏳ Waiting for Postgres (attempt %d)...\n", i)
		time.Sleep(3 * time.Second)
	}
	if err1 != nil {
		log.Fatalf("❌ Failed to connect to Postgres after retries: %v", err)
	}

	client := ent.NewClient(ent.Driver(drv))
	defer client.Close()

	router := api.NewRouter(client, cfg)
	log.Println("CO API running on :", port)
	if err := router.Run(fmt.Sprintf(":%s", port)); err != nil {
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
}


