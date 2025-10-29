# EOS (Edge Orchestration System) - Design Document

## Module Breakdown
| Component | Modules | Description |
|-----------|---------|-------------|
| CO | api, controller, fleetstore, margo_client, otel_gateway | API server, deployment logic, fleet storage, Margo communication |
| LO | server, fleetcache, nodeclient, otel_agent | gRPC server for CO, local deployer for EN |
| EN | server, executor, reporter | Receives and executes workloads, reports status |

## gRPC Service Design (Margo API)
```protobuf
service Margo {
  rpc DeployFleet(FleetRequest) returns (FleetResponse);
  rpc ReportStatus(StatusReport) returns (Ack);
}

service EdgeNode {
  rpc ApplyDeploy(DeployNodeRequest) returns (DeployNodeResponse);
  rpc ReportFromEdge(StatusReport) returns (Ack);
}
```

## YAML Fleet Definition Example
```yaml
apiVersion: v1
kind: Fleet
metadata:
  name: hello-fleet
spec:
  nodes:
    - id: edge1
      containers:
        - name: hello
          image: alpine
          command: ["echo", "hello world"]
```

## Resilience & Retry Logic
- LO retry policy for unreachable ENs with exponential backoff.
- CO retry for LO communication failures.
- Local caching for offline operation; replay once connected.

## Error Handling
- Structured gRPC error codes
- Observability pipeline captures logs, traces, metrics
- Local compensating actions (rollback, restart EN)

## Extensibility Hooks
- GitOps Controller for repo → fleet sync
- ML/Policy Agent for optimized deployment decisions
- Dapr sidecars for multi-language runtime