package orchestrator

import (
	"context"
	"log"
	"time"
	"github.com/google/uuid"
	"entgo.io/ent/dialect/sql"

	"github.com/balaji-balu/margo-hello-world/ent"
	"github.com/balaji-balu/margo-hello-world/ent/host"
	"github.com/balaji-balu/margo-hello-world/pkg/model"
	"github.com/balaji-balu/margo-hello-world/ent/deploymentstatus"
	"github.com/balaji-balu/margo-hello-world/ent/deploymentcomponentstatus"
)

func (lo *LocalOrchestrator) GetAllNodes(ctx context.Context) ([]*ent.Host, error) {
	nodes, err := lo.db.Host.Query().All(ctx)
	if err != nil {
		return nil, err
	}
	return nodes, nil
}

func (lo *LocalOrchestrator) CreateEdgeNode(ctx context.Context, 
	h model.HealthMsg) (error) {
	siteId, err := uuid.Parse(h.SiteID)
	if err != nil {
		//log.
		return err
	}
	err = lo.db.Host.
		Create().
		SetHostID(h.NodeID).
		SetSiteID(siteId).
		SetRuntime(h.Runtime).
		SetCPUFree(100 - h.CPUPercent).
		SetStatus("Alive").
		SetLastHeartbeat(time.Now()).
		//SetHostname(h.hostname).
		//SetIPAddress(ip).
		OnConflict(
			sql.ConflictColumns(host.FieldHostID), // field used as unique key
		).
		UpdateNewValues().
		Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}

// SaveDeploymentRecord creates a new deployment status record.
func (lo *LocalOrchestrator) SaveDeploymentRecord(ctx context.Context, 
	dep *model.DeploymentStatus) (*ent.DeploymentStatus, error) {
	id, err := uuid.Parse(dep.DeploymentID)
	//log.Println(id)	
	newDep, err := lo.db.DeploymentStatus.
		Create().
		SetID(id).
		SetState(dep.Status.State). 
		//SetCreatedAt(dep.CreatedAt).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return newDep, nil
}

func (lo *LocalOrchestrator) UpdateDeploymentRecord(
	ctx context.Context, dep model.DeploymentStatus) error {
	// --- 1. Parse deployment ID ---
	id, err := uuid.Parse(dep.DeploymentID)
	//log.Println(id)
	if err != nil {
		return err
	}

	// --- 2. Update parent deployment status ---
	_, err = lo.db.DeploymentStatus.
		Update().
		Where(deploymentstatus.ID(id)).
		SetState(dep.Status.State).
		SetErrorCode(dep.Status.Error.Code).
		SetErrorMessage(dep.Status.Error.Message).
		//SetUpdatedAtNow().
		Save(ctx)
	if err != nil {
		return err
	}
	log.Printf("Updated deployment record %s with state=%s", dep.DeploymentID, dep.Status.State)

	// --- 3. If component list exists, update each component ---
	for _, c := range dep.Components {
		// Try to find and update by name + parent deployment relation
		updated, err := lo.db.DeploymentComponentStatus.
			Update().
			Where(
				deploymentcomponentstatus.HasDeploymentWith(deploymentstatus.ID(id)),
				deploymentcomponentstatus.NameEQ(c.Name),
			).
			SetState(c.State).
			SetErrorCode(c.Error.Code).
			SetErrorMessage(c.Error.Message).
			//SetUpdatedAtNow().
			Save(ctx)
		if err != nil {
			return err
		}

		// If no rows were updated, create new component record
		if updated == 0 {
			_, err := lo.db.DeploymentComponentStatus.
				Create().
				SetName(c.Name).
				SetState(c.State).
				SetErrorCode(c.Error.Code).
				SetErrorMessage(c.Error.Message).
				//SetCreatedAtNow().
				//SetUpdatedAtNow().
				//SetDeploymentID(dep.DeploymentID). // attach to parent
				Save(ctx)
			if err != nil {
				return err
			}
			log.Printf("Created new component record: %s", c.Name)
		} else {
			log.Printf("Updated component record: %s", c.Name)
		}
	}

	return nil
}

// GetAllDeployments returns all deployment records.
func GetAllDeployments(ctx context.Context, db *ent.Client) ([]*ent.DeploymentStatus, error) {
	deployments, err := db.DeploymentStatus.Query().All(ctx)
	if err != nil {
		return nil, err
	}
	return deployments, nil
}
