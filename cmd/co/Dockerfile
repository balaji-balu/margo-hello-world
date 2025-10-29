# syntax=docker/dockerfile:1.4

#### --- Build Stage ---
FROM docker.io/library/golang:1.25-alpine AS builder

WORKDIR /app

# Install dependencies (git, curl, bash, and Atlas CLI)
# RUN apk add --no-cache git bash curl build-base && \
#     curl -sSf https://atlasgo.sh | sh

# Copy orchestrator app code (local context = ./co)
COPY . .

# Copy Ent schemas and migrations from parent directory contexts
COPY --from=parent_ent /ent ./ent
COPY --from=parent_migrations /migrations ./migrations
COPY --from=parent_pkg /pkg ./pkg
COPY --from=parent_proto /proto_generated ./proto_generated
COPY --from=parent_test /tests ./tests

# Fetch Go dependencies
COPY --from=parent_ent /go.mod ./go.mod
COPY --from=parent_ent /go.sum ./go.sum
RUN go mod download

# Optionally verify Ent code and migrations
# RUN go generate ./ent

# Compile the orchestrator app
RUN go mod tidy
RUN go build -o orchestrator ./main.go

#### --- Runtime Stage ---
FROM alpine:3.19

WORKDIR /app

# CA certificates, bash, and migration tools if needed
RUN apk add --no-cache ca-certificates bash

# Copy binary and runtime data from builder
COPY --from=builder /app/orchestrator /app/orchestrator
COPY --from=builder /app/migrations /app/migrations
COPY --from=builder /app/ent /app/ent
COPY --from=builder /app/tests /app/tests

# COPY --from=builder /app/*.yaml /app/*.yaml
# COPY --from=builder /app/proto_generated /app/proto_generated
# COPY --from=builder /app/go.mod /app/go.mod

# Environment configuration
ENV DATABASE_URL="postgres://postgres:postgres@postgres:5432/orchestration?sslmode=disable"
ENV OTEL_EXPORTER_OTLP_ENDPOINT="http://otel-collector:4317"
ENV SERVICE_NAME="central-orch"

# Default command; no entrypoint so Compose or K8s can override
CMD ["/app/orchestrator"]
