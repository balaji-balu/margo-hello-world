package integration

import (
    "os/exec"
    "testing"
    "time"
    "net/http"
    "io"
	"fmt"
	"os"
)

func start(t *testing.T, name string, path string, args ...string) *exec.Cmd {
    cmd := exec.Command("go", append([]string{"run", path}, args...)...)
cmd.Stdout = os.Stdout
cmd.Stderr = os.Stderr
    if err := cmd.Start(); err != nil {
        t.Fatalf("failed to start %s: %v", name, err)
    }
    time.Sleep(2 * time.Second)
    t.Logf("%s started (pid=%d)", name, cmd.Process.Pid)
    return cmd
}

func TestHelloWorldFlow(t *testing.T) {
    co := start(t, "CO", "./cmd/co/main.go --config=./configs/co.yaml")
    defer co.Process.Kill()
    lo := start(t, "LO", "./cmd/lo/main.go --config=./configs/lo_tiruvannamalai.yaml")
    defer lo.Process.Kill()
    en := start(t, "EN", "./cmd/en/main.go --config=./configs/edge1,yaml")
    defer en.Process.Kill()

    time.Sleep(3 * time.Second)

	if err := waitForHTTP(9101, "/healthz", 5, 2*time.Second); err != nil {
		t.Fatalf("CO not ready: %v", err)
	}	
    // trigger from CO
    resp, err := http.Get("http://localhost:9101/healthz")
    if err != nil {
        t.Fatalf("request to CO failed: %v", err)
    }
    defer resp.Body.Close()
    out, _ := io.ReadAll(resp.Body)

    want := "hello world"
    got := string(out)
    if got != want {
        t.Fatalf("unexpected response: got %q, want %q", got, want)
    }

    t.Log("✅ Hello World orchestration flow succeeded")
}

func waitForHTTP(port int, path string, retries int, delay time.Duration) error {
    url := fmt.Sprintf("http://localhost:%d%s", port, path)
    for i := 0; i < retries; i++ {
        resp, err := http.Get(url)
        if err == nil && resp.StatusCode == 200 {
            resp.Body.Close()
            return nil
        }
        time.Sleep(delay)
    }
    return fmt.Errorf("service at %s not ready", url)
}
