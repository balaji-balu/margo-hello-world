package main

import (
	"context"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	//"os/exec"
	//"path/filepath"
	"time"

	"github.com/balaji-balu/margo-hello-world/pkg/deployment"
	//"github.com/google/go-containerregistry/pkg/authn"
//"github.com/google/go-containerregistry/pkg/name"
	//"github.com/google/go-containerregistry/pkg/v1/remote"
	//"github.com/google/go-containerregistry/pkg/v1/tarball"

	"github.com/joho/godotenv"
	"os"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/content/oci"

	"os/exec"
	"flag"

	//"oras.land/oras-go/v2/content/file"
	"github.com/balaji-balu/margo-hello-world/internal/config"
	"go.uber.org/zap"
	"github.com/balaji-balu/margo-hello-world/internal/fsmloader"
	"github.com/balaji-balu/margo-hello-world/internal/edgenode"

)

var localOrchestratorURL string  // LO endpoint

func init() {
    err := godotenv.Load("./.env") // relative path to project root
    if err != nil {
        log.Println("No .env file found, reading from system environment")
    }
}
/*
func reportStatus(app string, status deployment.DeploymentStatus, msg string, nodeFSM *fsmloader.EdgeNodeFSM) {
	report := deployment.DeploymentReport{
		NodeID:  "edge-node-1",
		AppName: app,
		Status:  status,
		Message: msg,
		State:   nodeFSM.Current(),
		Timestamp:    time.Now().Format(time.RFC3339),
	}
	body, _ := json.Marshal(report)
	url := fmt.Sprintf("%s/deployment_status", localOrchestratorURL)
	http.Post(url, "application/json", bytes.NewReader(body))
}
*/

// register sends this edge node’s URL to the Local Orchestrator.
func register(edgeURL string) error {
	log.Println("Registering edge with Local Orchestrator...")
	payload := map[string]string{
		"edge_url": edgeURL,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	client := &http.Client{Timeout: 5 * time.Second}

	url := fmt.Sprintf("%s/register", localOrchestratorURL)
	log.Println("Sending registration request to", url)

	resp, err := client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("⚠️ Failed to register edge (%s): status=%s", edgeURL, resp.Status)
	} else {
		log.Printf("✅ Successfully registered edge: %s", edgeURL)
	}

	return nil
}

func handleDeploy(en *edgenode.EdgeNode,w http.ResponseWriter, r *http.Request, 
	nodeFSM *fsmloader.EdgeNodeFSM) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", 405)
		return
	}

	log.Println("Received deploy request")

	var req deployment.DeployRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	log.Println("Received deploy request for:", 
		req.AppName, req.Image, req.Token, req.Revision)

	ctx := context.Background()

	// if err := nodeFSM.FSM.Event(ctx, "start_deployment"); err != nil {
    //     //nodeFSM.logger.Error("Cannot start deployment", zap.Error(err))
    //     return
    // }

	nodeFSM.StartDeployment(ctx, []string{req.AppName})

	// ✅ Split repo and tag
	//repoName := "ghcr.io/edge-orchestration-platform/edge-onnx-sample"
	//tag := "8996d9c5b7a689283fbea25b8a6b5757d6b6bc5e"
	image := fmt.Sprintf("%s:%s",req.Image, req.Revision)

	token := os.Getenv("GITHUB_TOKEN") // your GitHub token

	// ✅ Create remote repository
	repo, err := remote.NewRepository(req.Image)
	if err != nil {
		http.Error(w, fmt.Sprintf("invalid repo: %v", err), 400)
		return
	}

	// ✅ Authenticate
	repo.Client = &auth.Client{
		Credential: auth.StaticCredential(repo.Reference.Registry, auth.Credential{
			Username: "balaji",
			Password: token,
		}),
		Cache: auth.NewCache(),
	}

	// ✅ Local OCI store (for caching)
	store, err := oci.New("local-cache")
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to create oci store: %v", err), 500)
		return
	}

	log.Println("Pulling OCI artifact using oras-go ...")

	// ✅ Pass tag separately (not part of repo)
	_, err = oras.Copy(ctx, repo, req.Revision, store, "", oras.DefaultCopyOptions)
	if err != nil {
		log.Printf("Failed to pull OCI image: %v", err)
		http.Error(w, fmt.Sprintf("Failed to pull OCI image: %v", err), 500)
		return
	}

	// Step 2️⃣ — Run the container using Docker
	log.Println("Starting container...", image)
	cmd := exec.Command("docker", "run", "-d", "--name", req.Revision, image)
	out, err := cmd.CombinedOutput()
	if err != nil {
		en.ReportStatus(req.AppName, "failed", 
			fmt.Sprintf("Docker run error: %v, output: %s", err, string(out)),
			nodeFSM.FSM.Current(),
		)
		log.Printf("Failed to start container: %v\n%s", err, string(out))
		if err := nodeFSM.FSM.Event(ctx, "deployment_failed"); err != nil {
			//nodeFSM.logger.Error("deployment failed", zap.Error(err))
			//return
		}

		http.Error(w, fmt.Sprintf("Failed to start container: %v\n%s", err, out), 500)
		return
	}
	containerID := string(bytes.TrimSpace(out))
	log.Printf("✅ Container started: %s", containerID)


	// Step 3️⃣ — Check container status
	statusCmd := exec.Command("docker", "inspect", "-f", "{{.State.Status}}", req.AppName)
	statusOut, err := statusCmd.CombinedOutput()
	if err != nil {
		en.ReportStatus(req.AppName, 
			"failed", fmt.Sprintf("Inspect failed: %v", err), nodeFSM.FSM.Current())
		http.Error(w, fmt.Sprintf("Failed to check container status: %v", err), 500)
		return
	}
	status := string(bytes.TrimSpace(statusOut))
	log.Printf("Container status: %s", status)


	log.Println("✅ OCI image pulled successfully!")
	w.Write([]byte("Deployment image pulled successfully"))

	en.ReportStatus(req.AppName, "success", 
		"Deployment image pulled successfully", nodeFSM.FSM.Current())
	
}

func main() {
	configPath := flag.String("config", "./configs/edge1.yaml", "config file path")
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("❌ Error loading config: %v", err)
	}

	log.Printf("✅ Loaded config: domain=%s, port=%d, LO URL=%s",
		cfg.Server.Domain, cfg.Server.Port, cfg.LO.URL)

	localOrchestratorURL = cfg.LO.URL
	en := edgenode.NewEdgeNode(localOrchestratorURL)

	// Initialize logger and context
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()
	ctx := context.Background()

	// Initialize FSM for this edge node
	nodeFSM := fsmloader.NewEdgeNodeFSM(ctx, "edge-node-1", en, logger)

	// Register with Local Orchestrator
	register(fmt.Sprintf("http://%s:%d", cfg.Server.Domain, cfg.Server.Port))

	// Define HTTP handlers
	http.HandleFunc("/deploy", func(w http.ResponseWriter, r *http.Request) {
		handleDeploy(en, w, r, nodeFSM)
	})

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	fmt.Println("🚀 Edge Node Server started on", addr)

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("❌ Server failed to start: %v", err)
	}
}