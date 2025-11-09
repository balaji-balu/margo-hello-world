package orchestrator

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	bolt "go.etcd.io/bbolt"
)

const nodesBucket = "nodes"
const deploymentsBucket = "deployments"

type EdgeNode struct {
	NodeID   string            `json:"node_id"`
	SiteID   string            `json:"site_id"`
	Runtime  string            `json:"runtime"`
	Region   string            `json:"region"`
	LastSeen time.Time         `json:"last_seen"`
	CPUFree  float64           `json:"cpu_free"`
	Alive    bool              `json:"alive"`
	Labels   map[string]string `json:"labels,omitempty"`
}

func InitDB(path string) *bolt.DB {
	log.Println("InitDB called with path:", path)
	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		log.Println("bolt open failed", err)
		log.Fatal("BoltDB open:", err)
	}
	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(nodesBucket))
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte(deploymentsBucket))
		return err
	})
	log.Println("Returning")
	return db
}

func SaveNode(db *bolt.DB, n EdgeNode) {
	data, _ := json.Marshal(n)
	db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte(nodesBucket)).Put([]byte(n.NodeID), data)
	})
}

func GetAllNodes(db *bolt.DB) []EdgeNode {
	ns := []EdgeNode{}
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(nodesBucket))
		b.ForEach(func(k, v []byte) error {
			var n EdgeNode
			json.Unmarshal(v, &n)
			ns = append(ns, n)
			return nil
		})
		return nil
	})
	return ns
}

// deployment record stored as map[nodeID]statusString
func SaveDeploymentRecord(db *bolt.DB, deployID string, record map[string]string) {
	b, _ := json.Marshal(record)
	db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte(deploymentsBucket)).Put([]byte(deployID), b)
	})
}

func GetAllDeployments(db *bolt.DB) map[string]map[string]string {
	res := map[string]map[string]string{}
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(deploymentsBucket))
		b.ForEach(func(k, v []byte) error {
			var rec map[string]string
			json.Unmarshal(v, &rec)
			res[string(k)] = rec
			return nil
		})
		return nil
	})
	return res
}

func UpdateDeploymentRecord(db *bolt.DB, deployID, nodeID string, success interface{}, message string) {
	// load existing
	var rec map[string]string
	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(deploymentsBucket))
		v := b.Get([]byte(deployID))
		if v == nil {
			rec = map[string]string{nodeID: toStr(success)}
			b2, _ := json.Marshal(rec)
			return b.Put([]byte(deployID), b2)
		}
		json.Unmarshal(v, &rec)
		rec[nodeID] = toStr(success)
		b2, _ := json.Marshal(rec)
		return b.Put([]byte(deployID), b2)
	})
}

func toStr(v interface{}) string {
	switch x := v.(type) {
	case bool:
		if x {
			return "success"
		} else {
			return "failure"
		}
	case string:
		return x
	default:
		return fmt.Sprintf("%v", v)
	}
}
