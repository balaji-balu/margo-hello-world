package main

import (
    "flag"
    "fmt"
    "time"

    "github.com/balaji-balu/margo-hello-world/internal/testenv"
)

func main() {
    dur := flag.Duration("simulate-network", 0, "duration to simulate network loss on loopback (e.g. 30s)")
    flag.Parse()
    if *dur == 0 {
        fmt.Println("no duration provided")
        return
    }
    if err := testenv.SimulateNetworkLoss(*dur); err != nil {
        fmt.Println("simulate network error:", err)
    } else {
        fmt.Println("network simulation complete")
    }
    time.Sleep(1 * time.Second)
}
