package orchestrator

import (
	// "log"
	// "math/rand"
	// "time"
)

/*
func PickNodes(nodes []EdgeNode, runtime, region string, max int) []EdgeNode {
	candidates := []EdgeNode{}
	for _, n := range nodes {

		//log.Println("[scheduler] checking node", n.NodeID)
		if !n.Alive {
			//log.Println("[scheduler] node is not alive")
			continue
		}
		if runtime != "" && n.Runtime != runtime {
			//log.Println("[scheduler] node runtime does not match", n.Runtime, runtime)
			continue
		}
		if region != "" && n.Region != region {
			//log.Println("[scheduler] node region does not match", n.Region, region)
			continue
		}
		candidates = append(candidates, n)
	}
	if len(candidates) == 0 {
		log.Println("[scheduler] no matching nodes found")
		return nil
	}

	rand.Seed(time.Now().UnixNano())
	// shuffle
	for i := range candidates {
		j := rand.Intn(i + 1)
		candidates[i], candidates[j] = candidates[j], candidates[i]
	}
	if len(candidates) > max {
		return candidates[:max]
	}
	return candidates
}
*/