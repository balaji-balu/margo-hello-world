package orchestrator

import (
    "context"
    "fmt"
    //"io"
    "log"
    "net/http"
    "os"
    "time"
    "go.uber.org/zap"
)

func (lo *LocalOrchestrator) Poll(ctx context.Context) {

    if err := lo.machine.Event(ctx, "receive_request"); err != nil {
        fmt.Println("❌ Poll: Error:", err)
        lo.logger.Error("Poll: Error:", zap.String("error", err.Error()))
        return
    }

    if len(lo.Hosts) == 0 {
        if err := lo.machine.Event(ctx, "no_edges"); err != nil {
            fmt.Println("❌ Poll: Error")
        }
        return
    }

    lo.logger.Info("Poll:Polling for changes...", 
        zap.String("owner", lo.Config.Owner), 
        zap.String("repo",lo.Config.Repo), 
        zap.String("token",lo.Config.Token))

    yamlData, err := fetchYAMLFromGitHub(ctx, 
            lo.Config.Token, lo.Config.Owner, lo.Config.Repo, lo.Config.Path)
    if err != nil {
            log.Printf("poll error: %v", err)
            return
    } 

    log.Println("✅ New desiredstate.yaml fetched, applying...")
    
    lo.applyDeployment(ctx, yamlData)

    lo.Journal.LastSuccess = time.Now()
}

func (lo *LocalOrchestrator) PollWithETag(ctx context.Context) {
    log.Println("Polling for changes...", lo.Config.Owner, lo.Config.Repo, lo.Config.Token)
    url := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/main/desiredstate.yaml",
        lo.Config.Owner, lo.Config.Repo)
    //https://api.github.com/repos/{owner}/{repo}/contents/{path}

    req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
    if lo.Journal.ETag != "" {
        req.Header.Set("If-None-Match", lo.Journal.ETag)
    }
    if lo.Config.Token != "" {
        req.Header.Set("Authorization", "Bearer "+lo.Config.Token)
    }

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        log.Println("2.Poll error:", err)
        return
    }
    defer resp.Body.Close()

    switch resp.StatusCode {
    case http.StatusNotModified:
        log.Println("No changes detected")
        return
    case http.StatusOK:
        lo.Journal.ETag = resp.Header.Get("ETag")
        //body, _ := io.ReadAll(resp.Body)
        log.Println("✅ New desiredstate.yaml fetched, applying...")
        //lo.ApplyDeployment(body)
        lo.Journal.LastSuccess = time.Now()
    default:
        log.Printf("Unexpected status %d\n", resp.StatusCode)
    }
}

func (lo *LocalOrchestrator) WaitForWebhook(ctx context.Context) {
    log.Println("Webhook mode not yet active; falling back to AdaptivePull")
    //lo.PollWithETag(ctx)
}

func (lo *LocalOrchestrator) ScanLocalInbox(ctx context.Context) {
    inbox := "/opt/lo/inbox"
    entries, err := os.ReadDir(inbox)
    if err != nil {
        log.Println("No local inbox found, skipping")
        return
    }

    for _, f := range entries {
        path := inbox + "/" + f.Name()
        data, err := os.ReadFile(path)
        if err == nil {
            log.Println("Applying offline deployment:", path, data)
            //.ApplyDeployment(data)
        }
    }
}
