//go:build poll
package trigger

import (
    "fmt"
    //"io"
    //"net/http"
    //"os"
    "time"
    "log"
    "context"
    "github.com/google/go-github/v53/github"
    "golang.org/x/oauth2"
    "gopkg.in/yaml.v3"
    "github.com/balaji/hello/pkg/deployment"
    //"github.com/balaji/hello/pkg/application"
    //"github.com/balaji/hello/ent/deploymentprofile"
    "encoding/json"
    "bytes"
    "net/http"
)

type PollTrigger struct {
    RepoOwner string
    RepoName  string
    Path      string
    Token     string
    Interval  time.Duration
}

func applyDeployment(ctx context.Context, yamlData []byte) {
    //log.Println("✅ Updated desiredstate.yaml from CO registry")


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
        PostDeployment(dply.Metadata.Name, "hosts[0]", c)
    }
}

func PostDeployment(appName string, hostid string, c deployment.Component) {

    req := deployment.DeployRequest{
		AppName: appName,
        Image:   c.Properties.Repository,
        //Token:   req.Token,
	}
    body, _ := json.Marshal(req)
	http.Post("http://localhost:8081/deploy", "application/json", bytes.NewReader(body))
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

func (p *PollTrigger) Start(ctx context.Context) error {
    //var etag string
    //url := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", p.RepoOwner, p.RepoName, p.Path)
    for {
        yamlData, err := fetchYAMLFromGitHub(ctx, p.Token, p.RepoOwner, p.RepoName, p.Path)
        if err != nil {
                log.Printf("poll error: %v", err)
        } else {
                applyDeployment(ctx, yamlData)
        }
        time.Sleep(p.Interval)
    }   
}

func (p *PollTrigger) Stop() {
	log.Println("[PollTrigger] Stopped polling loop.")
}