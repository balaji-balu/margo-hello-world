package orchestrator

import (
    "log"
    "net"
    "time"
)

type SyncMode int

const (
    PushPreferred SyncMode = iota
    AdaptivePull
    OfflineDeterministic
)

func networkStable() bool {
    conn, err := net.DialTimeout("tcp", "github.com:443", 2*time.Second)
    if err != nil {
        return false
    }
    _ = conn.Close()
    return true
}

func (lo *LocalOrchestrator) DetectMode() SyncMode {
    if networkStable() {
        return AdaptivePull
    }
    if time.Since(lo.Journal.LastSuccess).Hours() > 2 {
        log.Println("No connectivity, switching to Offline mode")
        return OfflineDeterministic
    }
    return AdaptivePull
}
