package orchestrator

import (
    //"os"
    "context"
    "fmt"
    "log"
    "net/http"

	"github.com/google/go-github/v53/github"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v3"
	"encoding/json"
	"bytes"
    
	"github.com/balaji-balu/margo-hello-world/pkg/deployment"
//"github.com/balaji-balu/margo-hello-world/ent"
    //fsmloader "github.com/balaji-balu/margo-hello-world/internal/fsm"
)

func (lo *LocalOrchestrator) applyDeployment(ctx context.Context, yamlData []byte) {
    //log.Println("✅ Updated desiredstate.yaml from CO registry")

    fmt.Println("applyDeployment.Current State:", lo.machine.Current())

    fmt.Println("New State:", lo.machine.Current())
   
    var dply deployment.ApplicationDeployment
    if err := yaml.Unmarshal(yamlData, &dply); err != nil {
        log.Println("❌ Failed to parse desiredstate.yaml:", err)
        return
    }
    log.Println("dply:", dply)
    //return 

    log.Printf("✅ Parsed Deployment: %s (%d components)\n", 
        dply.Metadata.Name, len(dply.Spec.DeploymentProfile.Components))

    for _, p := range dply.Spec.Parameters {
        log.Println("Parameter:", p.Name)
        if p.Name == "SiteId" {
            for _, t := range p.Targets {
                log.Println("trget:", t.Pointer)
                //log.Println("components:", t.Components)
                for _, c := range t.Components {
                    log.Println("component:", c)
                    //hosts.append(c)
                }
            }   
        }
    }   
    for _, c := range dply.Spec.DeploymentProfile.Components {
        log.Printf(" - Component: %s (image: %s)\n", c.Name)
        log.Printf("properties: %+v\n", c.Properties)
        log.Println("repository:", c.Properties.Repository)
        log.Println("revision:", c.Properties.Revision)
        log.Println("wait:", c.Properties.Wait)
        log.Println("timeout:", c.Properties.Timeout)
        log.Println("packageLocation:", c.Properties.PackageLocation)
        log.Println("keyLocation:", c.Properties.KeyLocation)

        // TODO: Apply logic here — e.g., send to Edge Node(s)
        if err := lo.machine.Event(ctx, "verify_success"); err != nil {
            fmt.Println("❌ Error:", err)
            return
        }
        fmt.Println("→:", lo.machine.Current())

        req := deployment.DeployRequest{
	    	AppName: dply.Metadata.Name,
            Image:   c.Properties.Repository,
            Token:   lo.Config.Token,
            Revision: c.Properties.Revision,
        }
        if err := PostDeployment(req, lo.Hosts, c); err != nil {
            log.Println("1. PostDeployment Error:", err)
            if err := lo.machine.Event(ctx, "edge_rejected"); err != nil {
                fmt.Println("❌ Error:", err)
                return
            }
            // fmt.Println("→:", lo.machine.Current())
            return
        }

        if err := lo.machine.Event(ctx, "edge_accepted"); err != nil {
            return
        }

        fmt.Println("→", lo.machine.Current())
    }
}

func PostDeployment(req deployment.DeployRequest, hostids []string, c deployment.Component) error{

    // req := deployment.DeployRequest{
	// 	AppName: appName,
    //     Image:   c.Properties.Repository,
    //     Token:   os.Getenv("GITHUB_TOKEN"),
    //     Revision: c.Properties.Revision,

    //     //AppName: req.AppName,
    //     //Image:   req.Image
    //     //Token:   req.Token,
	// }
    body, _ := json.Marshal(req)
    
    for _, hostid := range hostids {
        log.Println("hostid:", hostid)

        url:= fmt.Sprintf("%s/deploy", hostid)
        log.Println("url:", url)
        //.Println("req:", req)
        resp, err := http.Post(url, "application/json", bytes.NewReader(body))
        if err != nil {
            log.Println("Error:", err)
            return err
        }
        defer resp.Body.Close()

        if resp.StatusCode != http.StatusOK {
            log.Println("2.PostDeployment Error:", resp.Status)
            return err 
        }
    }
    return nil
}

func fetchYAMLFromGitHub(ctx context.Context, token, owner, repo, path string) ([]byte, error) {
    ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
    tc := oauth2.NewClient(ctx, ts)
    client := github.NewClient(tc)

    fileContent, _, _, err := client.Repositories.GetContents(ctx, owner, repo, path, nil)
    if err != nil {
        return nil, err
    }

    if fileContent == nil {
        return nil, fmt.Errorf("no file found at path %s (is it a directory?)", path)
    }
    content, err := fileContent.GetContent()
    if err != nil {
        return nil, err
    }

    return []byte(content), nil
}