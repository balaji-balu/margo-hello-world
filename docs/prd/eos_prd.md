# EOS (Edge Orchestration System) - Product Requirements Document

## Overview
EOS is a distributed orchestration system for managing workloads across edge environments. It enables centralized control with optional local orchestration for resilience, scalability, and autonomy.

## Objectives
- Simplify deployment of workloads to distributed edge sites.
- Provide centralized visibility and management.
- Support autonomous local operation when disconnected from central control.
- Enable both developer-friendly and operator-friendly interfaces (API, CLI, Web UI).

## Key Features
| Feature | Description |
|---------|-------------|
| Central orchestrator (CO) | Global control plane for all fleets/sites |
| Local orchestrator (LO) | Site-level control plane (optional) |
| Edge nodes (EN) | Hosts workloads and reports status |
| Margo API | Unified control-plane interface (CO↔LO↔EN) |
| Fleet management | Grouping of edge nodes for coordinated deployments |
| Deployment orchestration | Rollout, rollback, and monitoring of workloads |
| Observability | OpenTelemetry-based metrics/logs/traces |
| Connectivity resilience | Local autonomy when network link to CO fails |
| Security | Mutual TLS between CO, LO, and EN; Role-based access (Keycloak) |
| Multi-tenancy | Logical isolation of users and sites |

## Users
- Operators/Admins – manage global fleets, monitor health, configure policies.
- Developers – deploy and test services at the edge.
- Support Engineers – monitor deployments and handle incidents.

## Non-Goals
- EOS does not replace Kubernetes; it complements K8s for constrained edge environments.
- No dependency on persistent internet connectivity at edge.

## Success Metrics
- Deployment success rate > 99%
- End-to-end latency (CO→EN) < 2s under normal network
- Offline operation capability at least 24h
- Seamless recovery after network reconnection
