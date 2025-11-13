package profileselector

import (
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// Models --------------------------------------------------

type Scope string

const (
	ScopeSite   Scope = "site"
	ScopeRegion Scope = "region"
	ScopeGlobal Scope = "global"
)

type PlacementStrategy string

const (
	StrategyBestFit  PlacementStrategy = "best-fit"
	StrategyBalanced PlacementStrategy = "balanced"
	StrategySpread   PlacementStrategy = "spread"
	StrategyAffinity PlacementStrategy = "affinity"
)

// Profile defines how to select devices/sites
type Profile struct {
	ID          string             `yaml:"id" json:"id"`
	Name        string             `yaml:"name" json:"name"`
	Scope       Scope              `yaml:"scope" json:"scope"`
	Strategy    PlacementStrategy  `yaml:"strategy" json:"strategy"`
	Constraints []Constraint       `yaml:"constraints" json:"constraints"`
	Preferences []Preference       `yaml:"preferences" json:"preferences"`
}

type Constraint struct {
	Key      string      `yaml:"key" json:"key"`
	Operator string      `yaml:"operator" json:"operator"` // >=, <=, ==, >, <
	Value    interface{} `yaml:"value" json:"value"`
}

type Preference struct {
	Key    string  `yaml:"key" json:"key"`
	Weight float64 `yaml:"weight" json:"weight"`
}

// ApplicationDescription (minimal subset used for matching)
type ApplicationDescription struct {
	Name      string                 `json:"name" yaml:"name"`
	Resources map[string]interface{} `json:"resources" yaml:"resources"` // e.g. cpu.cores, memory.mb, gpu.count
}

// DeviceCapabilities models an EN's capabilities
type DeviceCapabilities struct {
	ID       string             `json:"id" yaml:"id"`
	SiteID   string             `json:"siteId" yaml:"siteId"`
	RegionID string             `json:"regionId" yaml:"regionId"`
	CPU      ResourceCPU        `json:"cpu" yaml:"cpu"`
	Memory   ResourceMemory     `json:"memory" yaml:"memory"`
	GPU      ResourceGPU        `json:"gpu" yaml:"gpu"`
	Network  ResourceNetwork    `json:"network" yaml:"network"`
	Meta     map[string]string  `json:"meta,omitempty" yaml:"meta,omitempty"`
}

type ResourceCPU struct {
	Cores float64 `json:"cores" yaml:"cores"`
}

type ResourceMemory struct {
	MB float64 `json:"mb" yaml:"mb"`
}

type ResourceGPU struct {
	Count float64 `json:"count" yaml:"count"`
}

type ResourceNetwork struct {
	LatencyMS float64 `json:"latency_ms" yaml:"latency_ms"`
	Bandwidth float64 `json:"bandwidth_mbps" yaml:"bandwidth_mbps"`
}

// SiteCapabilities groups devices
type SiteCapabilities struct {
	ID      string               `json:"id" yaml:"id"`
	Region  string               `json:"region" yaml:"region"`
	Devices []DeviceCapabilities `json:"devices" yaml:"devices"`
	Meta    map[string]string    `json:"meta,omitempty" yaml:"meta,omitempty"`
}

// SelectionResult is the canonical output
type SelectionResult struct {
	Scope    Scope   `json:"scope" yaml:"scope"`
	RegionID string  `json:"regionId" yaml:"regionId"`
	SiteID   string  `json:"siteId" yaml:"siteId"`
	DeviceID string  `json:"deviceId" yaml:"deviceId"`
	Score    float64 `json:"score" yaml:"score"`
	Reason   string  `json:"reason,omitempty" yaml:"reason,omitempty"`
}

// SelectionContext carries inputs to the selector
type SelectionContext struct {
	Profile Profile
	App     ApplicationDescription
	// Depending on scope, one or more of these will be used
	Sites  []SiteCapabilities
	Region string // optional region filter
	Site   string // optional site filter
}

// Selector interface --------------------------------------

type Selector interface {
	Select(ctx SelectionContext) ([]SelectionResult, error)
}

// HierarchicalSelector orchestrates scope-aware selection
type HierarchicalSelector struct {
	deviceSelector *DeviceSelector
}

func NewHierarchicalSelector() *HierarchicalSelector {
	return &HierarchicalSelector{deviceSelector: NewDeviceSelector()}
}

func (h *HierarchicalSelector) Select(ctx SelectionContext) ([]SelectionResult, error) {
	scope := ctx.Profile.Scope
	if scope == "" {
		scope = ScopeSite // default to site
	}

	switch scope {
	case ScopeSite:
		return h.selectWithinSite(ctx)
	case ScopeRegion:
		return h.selectAcrossRegion(ctx)
	case ScopeGlobal:
		return h.selectGlobally(ctx)
	default:
		return nil, fmt.Errorf("unknown scope: %s", scope)
	}
}

// Site-level selection: expects ctx.Site or single-site ctx.Sites
func (h *HierarchicalSelector) selectWithinSite(ctx SelectionContext) ([]SelectionResult, error) {
	var targetSite *SiteCapabilities
	if ctx.Site != "" {
		for _, s := range ctx.Sites {
			if s.ID == ctx.Site {
				targetSite = &s
				break
			}
		}
		if targetSite == nil {
			return nil, fmt.Errorf("site %s not found in context", ctx.Site)
		}
	} else {
		if len(ctx.Sites) == 0 {
			return nil, errors.New("no sites provided for site-scope selection")
		}
		// pick first site if not specified
		targetSite = &ctx.Sites[0]
	}

	deviceResults, err := h.deviceSelector.SelectDevices(ctx.App, targetSite.Devices, ctx.Profile)
	if err != nil {
		return nil, err
	}

	var out []SelectionResult
	for _, dr := range deviceResults {
		out = append(out, SelectionResult{
			Scope:    ScopeSite,
			RegionID: targetSite.Region,
			SiteID:   targetSite.ID,
			DeviceID: dr.DeviceID,
			Score:    dr.Score,
			Reason:   dr.Reason,
		})
	}
	return out, nil
}

// Region-level selection: evaluate each site in region, pick best devices per site then aggregate
func (h *HierarchicalSelector) selectAcrossRegion(ctx SelectionContext) ([]SelectionResult, error) {
	if len(ctx.Sites) == 0 {
		return nil, errors.New("no sites provided for region selection")
	}

	// filter by region if provided
	var regionSites []SiteCapabilities
	for _, s := range ctx.Sites {
		if ctx.Region == "" || s.Region == ctx.Region {
			regionSites = append(regionSites, s)
		}
	}
	if len(regionSites) == 0 {
		return nil, fmt.Errorf("no sites found for region: %s", ctx.Region)
	}

	var aggregated []SelectionResult

	// For each site, compute top device score
	for _, s := range regionSites {
		deviceResults, _ := h.deviceSelector.SelectDevices(ctx.App, s.Devices, ctx.Profile)
		if len(deviceResults) == 0 {
			continue
		}
		best := deviceResults[0]
		aggregated = append(aggregated, SelectionResult{
			Scope:    ScopeRegion,
			RegionID: s.Region,
			SiteID:   s.ID,
			DeviceID: best.DeviceID,
			Score:    best.Score,
			Reason:   best.Reason,
		})
	}

	// sort aggregated by score
	sort.Slice(aggregated, func(i, j int) bool { return aggregated[i].Score > aggregated[j].Score })
	return aggregated, nil
}

// Global selection: evaluate across all sites and regions
func (h *HierarchicalSelector) selectGlobally(ctx SelectionContext) ([]SelectionResult, error) {
	if len(ctx.Sites) == 0 {
		return nil, errors.New("no sites provided for global selection")
	}

	var global []SelectionResult
	for _, s := range ctx.Sites {
		deviceResults, _ := h.deviceSelector.SelectDevices(ctx.App, s.Devices, ctx.Profile)
		for i, d := range deviceResults {
			// optionally limit per-site top N; here we include top 3 per site
			if i >= 3 {
				break
			}
			global = append(global, SelectionResult{
				Scope:    ScopeGlobal,
				RegionID: s.Region,
				SiteID:   s.ID,
				DeviceID: d.DeviceID,
				Score:    d.Score,
				Reason:   d.Reason,
			})
		}
	}

	sort.Slice(global, func(i, j int) bool { return global[i].Score > global[j].Score })
	return global, nil
}

// DeviceSelector implements device-level matching and scoring

type deviceMatch struct {
	DeviceID string
	Score    float64
	Reason   string
}

type DeviceSelector struct {
}

func NewDeviceSelector() *DeviceSelector { return &DeviceSelector{} }

// SelectDevices filters devices by constraints and ranks them by preferences
func (ds *DeviceSelector) SelectDevices(app ApplicationDescription, devices []DeviceCapabilities, profile Profile) ([]deviceMatch, error) {
	var matches []deviceMatch

	for _, d := range devices {
		ok, r := ds.satisfiesConstraints(app, d, profile.Constraints)
		if !ok {
			continue
		}
		score := ds.computeScore(app, d, profile.Preferences)
		matches = append(matches, deviceMatch{DeviceID: d.ID, Score: score, Reason: r})
	}

	// sort descending
	sort.Slice(matches, func(i, j int) bool { return matches[i].Score > matches[j].Score })
	return matches, nil
}

// satisfiesConstraints checks all constraints and returns (ok, reason)
func (ds *DeviceSelector) satisfiesConstraints(app ApplicationDescription, d DeviceCapabilities, constraints []Constraint) (bool, string) {
	for _, c := range constraints {
		key := strings.ToLower(c.Key)
		switch key {
		case "cpu.cores":
			if !compareFloat(d.CPU.Cores, c.Operator, toFloat(c.Value)) {
				return false, fmt.Sprintf("cpu.cores %v %s %v failed", d.CPU.Cores, c.Operator, c.Value)
			}
		case "memory.mb":
			if !compareFloat(d.Memory.MB, c.Operator, toFloat(c.Value)) {
				return false, fmt.Sprintf("memory.mb %v %s %v failed", d.Memory.MB, c.Operator, c.Value)
			}
		case "gpu.count":
			if !compareFloat(d.GPU.Count, c.Operator, toFloat(c.Value)) {
				return false, fmt.Sprintf("gpu.count %v %s %v failed", d.GPU.Count, c.Operator, c.Value)
			}
		default:
			// allow matching against metadata keys like meta.power, meta.location.zone etc.
			if v, ok := d.Meta[c.Key]; ok {
				// only support equality/inequality for meta for now
				if c.Operator == "==" && v != fmt.Sprintf("%v", c.Value) {
					return false, fmt.Sprintf("meta.%s %s %v failed", c.Key, c.Operator, c.Value)
				}
				if c.Operator == "!=" && v == fmt.Sprintf("%v", c.Value) {
					return false, fmt.Sprintf("meta.%s %s %v failed", c.Key, c.Operator, c.Value)
				}
			}
		}
	}
	return true, "constraints satisfied"
}

func compareFloat(a float64, operator string, b float64) bool {
	switch operator {
	case ">=":
		return a >= b
	case "<=":
		return a <= b
	case ">":
		return a > b
	case "<":
		return a < b
	case "==":
		return a == b
	case "!=":
		return a != b
	default:
		return false
	}
}

func toFloat(v interface{}) float64 {
	switch t := v.(type) {
	case int:
		return float64(t)
	case int64:
		return float64(t)
	case float64:
		return t
	case float32:
		return float64(t)
	case string:
		// try parse but fallback to 0
		var f float64
		fmt.Sscanf(t, "%f", &f)
		return f
	default:
		return 0
	}
}

// computeScore combines preference weights into a single score. Preferences may be positive (prefer more) or negative (prefer less e.g. latency).
func (ds *DeviceSelector) computeScore(app ApplicationDescription, d DeviceCapabilities, prefs []Preference) float64 {
	score := 0.0
	// baseline: reward cpu/memory/gpu surplus
	cpuSurplus := math.Max(0, d.CPU.Cores-deduceAppResource(app, "cpu.cores"))
	memSurplus := math.Max(0, d.Memory.MB-deduceAppResource(app, "memory.mb"))
	gpuSurplus := math.Max(0, d.GPU.Count-deduceAppResource(app, "gpu.count"))

	// normalize by simple heuristics
	score += cpuSurplus * 0.1
	score += (memSurplus / 1024.0) * 0.05
	score += gpuSurplus * 0.5

	for _, p := range prefs {
		switch strings.ToLower(p.Key) {
		case "network.latency_ms":
			// lower latency -> higher score; weight expected to be negative if user prefers lower latency
			if d.Network.LatencyMS <= 0 {
				score += 0
			} else {
				score += (1.0 / (1.0 + d.Network.LatencyMS)) * p.Weight * 100.0
			}
		case "memory.mb":
			score += d.Memory.MB * p.Weight / 1024.0
		case "site.energy_cost":
			// energy cost expected in meta as string float
			if v, ok := d.Meta["energy_cost"]; ok {
				var ec float64
				fmt.Sscanf(v, "%f", &ec)
				score += (-ec) * p.Weight
			}
		default:
			// try meta key
			if v, ok := d.Meta[p.Key]; ok {
				var fv float64
				fmt.Sscanf(v, "%f", &fv)
				score += fv * p.Weight
			}
		}
	}

	// clamp
	if math.IsNaN(score) || math.IsInf(score, 0) {
		score = 0
	}
	return score
}

func deduceAppResource(app ApplicationDescription, key string) float64 {
	k := strings.ToLower(key)
	// simple path: resources map contains keys like "cpu.cores" or nested maps; we handle flat map only for now
	if app.Resources == nil {
		return 0
	}
	if v, ok := app.Resources[k]; ok {
		return toFloat(v)
	}
	return 0
}

// YAML helpers -------------------------------------------

func LoadProfileFromFile(path string) (Profile, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return Profile{}, err
	}
	var p Profile
	if err := yaml.Unmarshal(b, &p); err != nil {
		return Profile{}, err
	}
	return p, nil
}

func LoadProfilesFromFile(path string) ([]Profile, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var p []Profile
	if err := yaml.Unmarshal(b, &p); err != nil {
		return nil, err
	}
	return p, nil
}

// Simple usage example (documented function - not executed here)
func ExampleUsage() {
	// 1) load profile
	// p, _ := LoadProfileFromFile("configs/profiles/gpu-region-balance.yaml")

	// 2) build selection context with app and sites
	// ctx := SelectionContext{Profile: p, App: appDesc, Sites: sites}

	// 3) selector
	// sel := NewHierarchicalSelector()
	// results, _ := sel.Select(ctx)

	// 4) use top result(s) to populate DesiredState per-site/device
}


# Margo Profile Selection Architecture and Runbook

## 1. Overview

This document describes the **Profile Selection** subsystem in a Margo-compliant Edge Orchestration environment. It provides conceptual understanding, design rationale, and operational steps for selecting appropriate deployment targets (sites, regions, or devices) based on the application's declared requirements and the available device capabilities.

The module is designed for integration into the **Central Orchestrator (CO)** and 
operates within a multi-layer orchestration hierarchy (CO → LO → EN).

---

## 2. Purpose

The Profile Selector is responsible for mapping **Application Descriptions** (defined per [Margo Application Package Spec](https://specification.margo.org/specification/application-package/application-description/)) to **Device Capabilities** (defined per [Margo Management Interface Device Capabilities](https://specification.margo.org/specification/margo-management-interface/device-capabilities/)) during deployment request processing.

It ensures that:
- Application resource requirements (CPU, memory, GPU, network, etc.) are satisfied.
- Deployments respect scope (site, region, or global).
- Placement strategy (best-fit, balanced, spread, or affinity) is followed.

---

## 3. Context in Deployment Flow

1. **CLI** registers the application using the Margo-compliant Application Description.
2. **CO** parses and persists the application metadata.
3. **CLI** issues a deployment request (Desired State as per [Margo Desired State Spec](https://specification.margo.org/specification/margo-management-interface/desired-state/)).
4. **CO** triggers the Profile Selector to choose the optimal device(s) or site(s).
5. The resulting deployment targets are embedded into the Desired State YAML and stored in Git.
6. **LO** observes site-specific Desired State updates and delegates deployment to **EN**.
7. **EN** executes the installation and reports progress through the Deployment Status pipeline.

---

## 4. Hierarchical Selection Design

The Profile Selector operates in a **multi-scope hierarchical model**:

| Scope | Description | Typical Owner |
|--------|--------------|----------------|
| **Site** | Selects devices within a single site. | Local Orchestrator (LO) |
| **Region** | Selects across multiple sites in a region. | Regional CO or CO agent |
| **Global** | Selects across all regions. | Central Orchestrator (CO) |

---

## 5. Profile Model

```go
type Scope string
const (
    ScopeSite   Scope = "site"
    ScopeRegion Scope = "region"
    ScopeGlobal Scope = "global"
)

type PlacementStrategy string
const (
    StrategyBestFit   PlacementStrategy = "best-fit"
    StrategyBalanced  PlacementStrategy = "balanced"
    StrategySpread    PlacementStrategy = "spread"
    StrategyAffinity  PlacementStrategy = "affinity"
)

type Constraint struct {
    Key      string `json:"key"`
    Operator string `json:"operator"`
    Value    interface{} `json:"value"`
}

type Preference struct {
    Key    string  `json:"key"`
    Weight float64 `json:"weight"`
}

type Profile struct {
    ID          string            `json:"id"`
    Name        string            `json:"name"`
    Scope       Scope             `json:"scope"`
    Strategy    PlacementStrategy `json:"strategy"`
    Constraints []Constraint      `json:"constraints,omitempty"`
    Preferences []Preference      `json:"preferences,omitempty"`
}
```

Profiles define **what kind of sites or devices** should host an application and **how** to prioritize among matching candidates.

---

## 6. SelectionContext

```go
type SelectionContext struct {
    Scope     model.Scope
    Profile   model.Profile
    Sites     []model.SiteCapabilities
    Devices   []model.DeviceCapabilities
    RegionID  string
    SiteID    string
}
```

This context object provides the necessary runtime data for the selector, including available devices, the region/site scope, and the target profile.

---

## 7. Hierarchical Selector

The main entry point is the **HierarchicalSelector**, which orchestrates selection according to the given scope.

```go
func (s *HierarchicalSelector) Select(app model.ApplicationDescription, ctx SelectionContext) ([]model.SelectionResult, error) {
    switch ctx.Profile.Scope {
    case model.ScopeSite:
        return s.selectWithinSite(app, ctx)
    case model.ScopeRegion:
        return s.selectAcrossRegion(app, ctx)
    case model.ScopeGlobal:
        return s.selectGlobally(app, ctx)
    default:
        return nil, fmt.Errorf("invalid scope: %v", ctx.Profile.Scope)
    }
}
```

Each of the sub-functions implements the filtering and ranking logic:

- **Site Selector:** Evaluates individual devices against constraints and preferences.
- **Region Selector:** Aggregates results from multiple sites in a region.
- **Global Selector:** Combines results from all regions to form a global ranked list.

---

## 8. Example Profile YAMLs

### Region-level GPU Intensive Application
```yaml
id: gpu-region-balance
name: GPU Balanced Profile
scope: region
strategy: balanced
constraints:
  - key: "gpu.count"
    operator: ">="
    value: 1
preferences:
  - key: "network.latency_ms"
    weight: -0.5
  - key: "site.energy_cost"
    weight: -0.2
```

### Site-level Lightweight Sensor Application
```yaml
id: sensor-local
name: Low Power Local Profile
scope: site
strategy: best-fit
constraints:
  - key: "cpu.cores"
    operator: "<="
    value: 2
preferences:
  - key: "location.zone"
    weight: 0.8
```

---

## 9. Output Example

For a region-level selection, the selector may return ranked results:

```json
[
  {
    "scope": "site",
    "regionId": "region-west",
    "siteId": "shop-floor-A",
    "deviceId": "edge-node-01",
    "score": 0.92
  },
  {
    "scope": "site",
    "regionId": "region-west",
    "siteId": "shop-floor-B",
    "deviceId": "edge-node-03",
    "score": 0.81
  }
]
```

---

## 10. Integration with Deployment Flow

| Step | Component | Action |
|------|------------|--------|
| 1️⃣ | CLI | `edgectl deploy app digitron-orchestrator --profile gpu-region-balance --region west` |
| 2️⃣ | CO | Loads app + profile + topology → runs `HierarchicalSelector` |
| 3️⃣ | CO | Updates **DesiredState YAML** with selected sites/devices |
| 4️⃣ | GitOps | Commits YAML per-site into Git |
| 5️⃣ | LO | Detects change and deploys locally |
| 6️⃣ | EN | Executes install and updates DeploymentStatus |

---

## 11. Operational Runbook

### 11.1 Pre-Deployment
- Ensure all **device capabilities** are reported to CO (via LO).
- Validate that **profiles** are registered and accessible.
- Check that **application descriptions** correctly define resource requirements.

### 11.2 Deployment Execution
- CO loads the profile and constructs the `SelectionContext`.
- HierarchicalSelector executes and ranks devices/sites.
- CO updates the DesiredState YAML and commits to Git.
- LO and EN automatically act on the changes.

### 11.3 Monitoring & Troubleshooting
- Verify selection logs under CO logs: `profileselector.log`.
- Use CLI: `edgectl profile simulate --app <app> --profile <profile>` to dry-run selection.
- If selection fails:
  - Check missing or mismatched constraint keys.
  - Validate device capability reporting (may be stale or missing fields).
  - Review the applied scope and strategy.

---

## 12. Future Enhancements
- Support **dynamic weight tuning** based on telemetry.
- Add **affinity/anti-affinity policies**.
- Integrate with **ML-based capacity forecasting**.
- Provide **REST API endpoints** for profile simulation.

---

## 13. Summary

The **Profile Selector** is a key decision-making component in Margo-based orchestration. It bridges application requirements and device capabilities using a flexible, policy-driven, and hierarchical selection mechanism. 

This document doubles as both **architecture guide** and **runbook** for engineers operating or extending the system.

