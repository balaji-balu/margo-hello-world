# EOS (Edge Orchestration System) - Architecture Document

## System Overview
Three-tier orchestration system:
- CO – Central Orchestrator  
- LO – Local Orchestrator (per site)  
- EN – Edge Node (per host)

Users interact via Web Portal or CLI through CO API.

## Components
### Central Orchestrator (CO)
- Exposes REST/GraphQL API for users.
- Uses Margo gRPC API to communicate with LOs.
- Maintains global Fleet Registry and deployment controller.
- Integrates with Keycloak for authentication & authorization.
- Aggregates observability data from all LOs.

### Local Orchestrator (LO)
- One per site.
- Acts as proxy and local controller for edge nodes.
- Communicates with CO via Margo API.
- Communicates with ENs via local gRPC network (overlay or LAN).
- Performs local decision-making when CO link is unavailable.

### Edge Node (EN)
- Executes workloads (containers, scripts, etc.)
- Runs lightweight runtime (e.g., containerd)
- Sends periodic telemetry and async deployment status
- Connects to LO (or CO if no LO)

## Communication Paths
| From | To | Interface | Purpose |
|------|----|-----------|---------|
| CO | LO | gRPC (Margo API) | Deployments, configs, telemetry |
| LO | EN | gRPC | Deployment and status control |
| EN | LO | gRPC callback | Status/telemetry |
| User | CO | REST/GraphQL | Portal/CLI access |

## Deployment Data Flow
1. User → CO API: `POST /deploy/fleet`
2. CO parses YAML → builds `FleetSpec`
3. CO → each LO via `DeployFleet()`
4. LO → each EN via `ApplyDeploy()`
5. EN executes and reports status upstream.
6. CO aggregates and exposes results via `/status`

## Observability & Security
- OTEL: CO = Gateway mode; LO/EN = Agent mode
- Security: mTLS between all gRPC connections; Keycloak for identity

## ASCII Architecture Diagram
```
            +----------------------+
            | Web Portal / CLI     |
            +----------+-----------+
                       |
                       v
             +----------------------+
             | Central Orchestrator |
             |  (CO)                |
             +----------+-----------+
                       | Margo gRPC
        -------------------------------------------------
         |                    |                     |
 +----------------+   +----------------+     +----------------+
 | Local Orchestr.|   | Local Orchestr.|     | Local Orchestr.|
 |   (LO - Site A)|   |   (LO - Site B)| ... |   (LO - Site N)|
 +--------+-------+   +--------+-------+     +--------+-------+
          |                    |                     |
     Overlay or Local Net  Overlay or Local Net  Overlay or Local Net
          |                    |                     |
   +------+------+      +-------+------+       +------+------+
   | EN1 EN2 EN3 |      | EN4 EN5 EN6 |       | EN7 EN8 EN9 |
   +-------------+      +-------------+       +-------------+
```
