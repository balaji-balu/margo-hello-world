package orchestrator

import (
// "context"
// "encoding/json"
// "fmt"
// "log"
// "math/rand"
// "time"

)

/*
func deploy(dep ) {
	// ✅ Extract important info
	deploymentID := dep.Metadata.Annotations.ApplicationID//dep.Metadata.Annotations["id"]
	deploymentType := dep.Spec.DeploymentProfile.Type
	//repo := dep.Spec.DeploymentProfile.Components[0].Properties["repository"]
	//revision := dep.Spec.DeploymentProfile.Components[0].Properties["revision"]

	//dep.Spec.DeploymentProfile.Components[0].Properties["repository"].(string)
	props := dep.Spec.DeploymentProfile.Components[0].Properties
	repoVal := props.Repository
	revisionVal := props.Revision

	log.Printf("[LO] Deployment ID: %s", deploymentID)
	log.Printf("[LO] Deployment Type: %s", deploymentType)
	log.Printf("[LO] Repository: %v", repoVal)
	log.Printf("[LO] Revision: %v", revisionVal)
	//log.Printf("[LO] DeploymentType:%v", )

	if deploymentType == "compose" {
		deploymentType = "containerd"
	}

	req := DeployAPIRequest{
		DeploymentID: deploymentID,
		GitRepoURL: repoVal,
		Revision:     revisionVal,
		//WasmImages:   []string{},
		//ContainerImages: [repoVal],
		//TargetNodes:  []string{},
		RuntimeFilter: deploymentType,
		Region:       "india-south-1",
	}
	req.ContainerImages = append(req.ContainerImages, repoVal)



	fsm.Transition(shared.Deploying)

	// build shared.DeployRequest
	dr := shared.DeployRequest{
		DeploymentID: req.DeploymentID,
		GitRepoURL:   req.GitRepoURL,
		WasmImages:   req.WasmImages,
		ContainerImages: req.ContainerImages,
		Revision:     req.Revision,
		EdgeNodeIDs:  req.TargetNodes,
	}

	// choose targets
	targets := []EdgeNode{}
	if len(req.TargetNodes) > 0 {
		// load nodes from db
		all := GetAllNodes(db)
		for _, n := range all {
			for _, id := range req.TargetNodes {
				if n.NodeID == id {
					targets = append(targets, n);
					break
				}
			}
		}
	} else {
		all := GetAllNodes(db)
		rt := req.RuntimeFilter
		if rt == "" && len(req.WasmImages) > 0 { rt = "wasm" }
		if rt == "" && len(req.ContainerImages) > 0 { rt = "containerd" }
		targets = PickNodes(all, rt, req.Region, 10)
	}

	if len(targets) == 0 {
		//http.Error(w, "no target nodes available", 400)
		log.Println("[LO] no target nodes available")
		return fmt.Errorf("no target nodes available")
	}

	// prepare deployment record map with pending
	rec := map[string]string{}
	for _, t := range targets { rec[t.NodeID] = "pending" }
	SaveDeploymentRecord(db, dr.DeploymentID, rec)

	// send to each node with retries + circuit breaker + jitter
	for _, t := range targets {
		subj := fmt.Sprintf("site.%s.deploy.%s", t.SiteID, t.NodeID)
		payload, _ := json.Marshal(dr)
		attempts := 3
		for i := 0; i < attempts; i++ {
			err := cb.Call(func() error { return nc.Publish(subj, payload) })
			if err == nil {
				log.Printf("[LO] published deploy %s -> %s", dr.DeploymentID, t.NodeID)
				break
			}
			// jittered sleep
			t := time.Duration(200+rand.Intn(400)) * time.Millisecond
			time.Sleep(t)
		}
	}
	return nil
}
*/
