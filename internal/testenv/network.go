package testenv

import (
	"os/exec"
	"time"
)

func SimulateNetworkLoss(duration time.Duration) error {
    cmd := exec.Command("bash", "-c", "sudo tc qdisc add dev lo root netem loss 70%")
    cmd.Run()
    time.Sleep(duration)
    exec.Command("bash", "-c", "sudo tc qdisc del dev lo root").Run()
	return nil
}
