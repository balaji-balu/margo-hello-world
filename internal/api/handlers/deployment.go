package handlers

import (
    "context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/balaji/hello/pkg"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gopkg.in/yaml.v3"

	pb "github.com/balaji/hello/proto_generated"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	"github.com/gin-gonic/gin"
    "github.com/balaji/hello/ent"
    "github.com/balaji/hello/ent/deploymentprofile"
    "github.com/balaji/hello/ent/component"
    "github.com/balaji/hello/pkg/application"
    "github.com/balaji/hello/pkg/deployment"
    
    "net/http"	
    "github.com/google/go-github/v55/github"
    "golang.org/x/oauth2"

    //"github.com/joho/godotenv"
    "path/filepath"
)

// server implements CentralOrchestrator gRPC interface
type server struct {
    pb.UnimplementedCentralOrchestratorServer
}

type HostMapping struct {
    SiteID  string   `json:"SiteID"`
    HostIDs []string `json:"HostIDs"`
}

type App struct {
    AppID     string        `json:"AppID"`
    ProfileID string        `json:"ProfileID"`
    Sites     []HostMapping `json:"Sites"`
}

func init() {
    // err := godotenv.Load("../.env") // relative path to project root
    // if err != nil {
    //     log.Println("xzzxzxzx No .env file found, reading from system environment")
    // }
}

func (s *server) ReportStatus(ctx context.Context, req *pb.StatusReport) (*pb.DeployResponse, error) {
    status := req.Statuses[0]
    log.Printf("[CO] ReportStatus received for deployment %s, node=%s, status=%s",
        req.DeploymentId, 
        status.NodeId, 
        status.Status.String())
    return &pb.DeployResponse{
        DeploymentId: req.DeploymentId,
        Message:      "Status received by CO",
    }, nil
}

// deployFleet reads a YAML file and sends deploy request to LO
func deployFleet(ctx context.Context, yamlPath, loAddr string) error {

    workflow := "policy-sync"

    tracer := otel.Tracer("eos/orchestrator")
    ctx, span := tracer.Start(ctx, "ScheduleDeployment")
    defer span.End()

    log.Printf("[CO] reading fleet from %s", yamlPath)
    b, err := os.ReadFile(yamlPath)
    if err != nil {
        return fmt.Errorf("read yaml: %v", err)
    }

    var inp pkg.Deployment
    // var inp struct {
    //     Metadata struct {
    //         Name string `yaml:"name"`
    //     } `yaml:"metadata"`
    //     Spec struct {
    //         Nodes []struct {
    //             Id       string `yaml:"id"`
    //             Services []struct {
    //                 Name      string `yaml:"name"`
    //                 Container struct {
    //                     Image   string   `yaml:"image"`
    //                     Command []string `yaml:"command"`
    //                 } `yaml:"container"`
    //             } `yaml:"services"`
    //         } `yaml:"nodes"`
    //     } `yaml:"spec"`
    // }

    if err := yaml.Unmarshal(b, &inp); err != nil {
        return fmt.Errorf("yaml unmarshal: %v", err)
    }

    fleet := &pb.Fleet{Name: inp.Metadata.Name}
    for _, n := range inp.Spec.Nodes {
        node := &pb.NodeSpec{Id: n.Id}
        span.SetAttributes(
            attribute.String("eos.node", n.Id),
            attribute.String("eos.workflow", workflow),
            attribute.String("margo.spec", "v0.7-pre"),
        )
        for _, s := range n.Services {
            svc := &pb.ServiceSpec{
                Name: s.Name,
                Container: &pb.ContainerSpec{
                    Image:   s.Container.Image,
                    Command: s.Container.Command,
                },
            }
            node.Services = append(node.Services, svc)
        }
        fleet.Nodes = append(fleet.Nodes, node)
    }

    deploymentID := fmt.Sprintf("dep-%d", time.Now().Unix())
    req := &pb.DeployRequest{DeploymentId: deploymentID, Fleet: fleet}

    conn, err := grpc.Dial(loAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        return fmt.Errorf("dial lo: %v", err)
    }
    defer conn.Close()
    client := pb.NewLocalOrchestratorClient(conn)

    ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
    defer cancel()

    log.Printf("[CO] sending deploy to LO at %s", loAddr)
    resp, err := client.ReceiveDeploy(ctx, req)
    if err != nil {
        return fmt.Errorf("send deploy to lo: %v", err)
    }

    log.Printf("[CO] LO responded: %s, deployment_id=%s", resp.Message, resp.DeploymentId)
    return nil
}

func buildDeployParameters(sites []HostMapping) []deployment.Parameter {
    var params []deployment.Parameter

    for _, site := range sites {
        param := deployment.Parameter{
            Name:"SiteId",
            Value: site.SiteID, // This can represent site or contextual value
            Targets: []deployment.Target{
                {
                    Pointer:    fmt.Sprintf("/sites/%s", site.SiteID),
                    Components: site.HostIDs, // hostIDs act as components or nodes
                },
            },
        }
        params = append(params, param)
    }

    return params
}

func buildDeploymentComponents(ents []*ent.Component) []deployment.Component {
    comps := make([]deployment.Component, 0, len(ents))
    for _, c := range ents {
        comps = append(comps, deployment.Component{
            Name: c.Name,
            Properties: application.ComponentProperties{
                Repository:      c.Properties.Repository,
                Revision:        c.Properties.Revision,
                Wait:            c.Properties.Wait,
                Timeout:         c.Properties.Timeout,
                PackageLocation: c.Properties.PackageLocation,
                KeyLocation:     c.Properties.KeyLocation,
            },
        })
    }
    return comps
}

func buildDeploymentProfile( 
    profile *ent.DeploymentProfile, 
    components []*ent.Component) deployment.DeploymentProfile {
    return deployment.DeploymentProfile{
        Type:       profile.Type,
        Components: buildDeploymentComponents(components),
    }
}

func buildApplicationDeployment(
    appDesc *ent.ApplicationDesc,
    profile *ent.DeploymentProfile,
    components []*ent.Component,
    sites []HostMapping,
) deployment.ApplicationDeployment {
   
    return deployment.ApplicationDeployment{
        APIVersion: "margo.edge/v1",
        Kind:       "Deployment",
        Metadata: deployment.Metadata{
            Name: appDesc.Name + "-deployment",
            Labels: map[string]string{
                "app": appDesc.Name,
            },
            Annotations: deployment.Annotations{
                ApplicationID: appDesc.ID,
                ID:            "deployment-1", // could generate UUID here
            },
        },
        Spec: deployment.Spec{
            DeploymentProfile: buildDeploymentProfile(profile, components),
            Parameters: buildDeployParameters(sites),
        },
    }
}

// PushDeploymentYAML converts a deployment struct to YAML and pushes to GitHub
func PushDeploymentYAML(ctx context.Context, token, owner, repo, path, message string, appdply interface{}) error {
    // 1. Marshal to YAML
    yamlBytes, err := yaml.Marshal(appdply)
    if err != nil {
        return fmt.Errorf("failed to marshal YAML: %w", err)
    }

    // 2. Setup GitHub client with token
    ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
    tc := oauth2.NewClient(ctx, ts)
    client := github.NewClient(tc)

    // 3. Check if file exists to get SHA (needed to update)
    fileContent, _, resp, err := client.Repositories.GetContents(ctx, owner, repo, path, nil)
    var sha *string
    if resp != nil && resp.StatusCode == 200 && fileContent != nil {
        sha = fileContent.SHA
    }

    opts := &github.RepositoryContentFileOptions{
        Message: github.String(message),
        Content: yamlBytes,
        SHA:     sha, // nil if creating new file
    }

    if sha == nil {
        _, _, err = client.Repositories.CreateFile(ctx, owner, repo, path, opts)
    } else {
        _, _, err = client.Repositories.UpdateFile(ctx, owner, repo, path, opts)
    }

    if err != nil {
        return fmt.Errorf("failed to push file to GitHub: %w", err)
    }

    log.Println("YAML file successfully pushed to GitHub:", path)
    return nil
}

func CreateDeployment(c *gin.Context, client *ent.Client) {

    var app App
    if err := c.ShouldBindJSON(&app); err != nil {
        log.Printf("Error binding JSON: %v", err)
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    //appID := c.Param("id")
    log.Println("app:", app)
    log.Printf("appID:", app.AppID)

    ctx := context.Background()
    appDesc, err := client.ApplicationDesc.Get(ctx, app.AppID)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "application not found"})
        return
    }
    log.Println("appdesc:", appDesc)    
    profile, err := client.DeploymentProfile.
        Query(). 
        Where(deploymentprofile.AppIDEQ(app.AppID)).
        Only(ctx)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "deployment profile not found"})
        return
    }
    log.Println("profile:", profile)

    log.Println("profile id", profile.ID)
    components, err := client.Component.
        Query(). 
        Where(component.DeploymentProfileIDEQ(profile.ID)).
        All(ctx)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "components  not found"})
        return
    }
    log.Println("components:", components)

    appdply := buildApplicationDeployment(appDesc, profile, components, app.Sites)
    log.Println("appdply:", appdply)

    //yamlData, err := yaml.Marshal(&appdply)

    //os.WriteFile("./desiredstate.yaml", yamlData, 0644)
    //log.Println("yamldata:", yamlData)

    token := os.Getenv("GITHUB_TOKEN")
    
    owner := "edge-orchestration-platform"
    repo := "deployments"
    path := filepath.Join("site-1", profile.ID, "desiredstate.yaml")
    message := "Add new deployment"
    if err := PushDeploymentYAML(ctx, token, owner, repo, path, message, appdply); err != nil {
        log.Fatal(err)
    }
    return


    // var ad application.ApplicationDescription

	// if err := c.ShouldBindJSON(&ad); err != nil {
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	// 	return
	// }

	// ctx := context.Background()
	// if err := Persist(ctx, client, &ad); err != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	// 	return
	// }


}

func GetDeploymentStatus(c *gin.Context, client *ent.Client) {
	//id := c.Param("id")

	// ctx := context.Background()		
	// deployment, err := client.DeploymentProfile.Query().Where(deploymentprofile.ID(id)).Only(ctx)
	// if err != nil {
	// 	c.JSON(http.StatusNotFound, gin.H{"error": "deployment.Get(ctx, id)) not found"})
	// 	return
	// }

	// c.JSON(http.StatusOK, deployment)

	// if err != nil {
	// 	c.JSON(http.StatusNotFound, gin.H{"error": "deployment not found"})
	// 	return
	// }

	// c.JSON(http.StatusOK, deployment)
}