
podman-compose -f docker-compose.2lo-4en.yaml --env-file .env.staging up --scale en=2 -d --build


field.JSON("tags", []string{}).Optional(),

site.f95d34b2-8019-4590-a3ff-ff1e15ecc5d5.deploy.edge1-containerd
f95d34b2-8019-4590-a3ff-ff1e15ecc5d5 edge1-containerd containerd india-south-1 9201}

at root directory
 ent generate ./ent/schema`
 
 atlas migrate diff   --dev-url "docker://postgres/16/test?search_path=public"   --to "ent://ent/schema"   --dir "file://migrations"
 
 atlas migrate apply --url "postgres://postgres:postgres@localhost:5432/orchestration?sslmode=disable" --dir "file://migrations"``

// reset db
atlas schema clean --url "postgres://postgres:postgres@localhost:5432/orchestration?sslmode=disable"
atlas migrate diff initial \
  --to "ent://ent/schema" \
  --dev-url "docker://postgres/16/test?search_path=public"


//pg dump
sudo apt install postgresql-client

pg_dump -h localhost -d orchestration -U postgres -s -F p -E UTF-8 -f ./output.txt


cd /path
protoc --go_out=. --go-grpc_out=. proto/*.proto

another terminal
cd co
go run main.c

another terminal
cd lo
go run main.c

another terminal
cd en
go run main.c

test/1.http

POST http://localhost:8080/deploy?file=fleet.yaml
```
open another terminal 

```bash
docker run -d --name otel-collector   -p 4317:4317   -p 4318:4318   -v $(pwd)/collector-config.yaml:/etc/otel/config.yaml   docker.io/otel/opentelemetry-collector-contrib:0.137.0
```

   ./co -f examples/fleet.yaml -lo localhost:50052
   ./lo -port :50052 -co localhost:50051
   ./en -port :50054 -id edge1 -lo localhost:50052
   ./en -port :50055 -id edge2 -lo localhost:50052

1. Edge Node: listens on port 60052

= Reads the deployment instructions from LO .
- Starts containers locally (like the Python/Go example we wrote).
- Sends status (success/failure) back to LO .

2. Local Orchestrator (LO): listens on port 60051

- Receives deployment instructions from CO .
- Forwards them to the correct edge node(s) to EN.
- Collects deployment status from edge nodes.
- Forwards status back to CO.

3. Central Orchestrator (CO): listens on port 50051

- Sends deployment instructions to LO .
- Receives final status from LO .

### Todo:
- postgres


**Now we have:**
- Traces: EN container executions, LO orchestration, CO orchestration decisions
- Metrics: Deployment counts, node status
- Full observability: CO can see EN activity in near real-time
= Central Orchestrator (CO) → sends fleet manifest
- Local Orchestrator (LO) → parses manifest → deploys only its node’s containers
- Edge Node (EN) → runs multiple containers on request
- Scalable to N nodes, M services per node
- Manifest-driven, so changes automatically propagate

### Testing the Observability Flow

- Run OTEL Collector
- Start EN → LO → CO
- Observe logs in OTEL Collector for:
- Traces for each container run
- Metrics counters for successful deployments


Reference:
- chatgpt "Hello world prototype" https://chatgpt.com/c/68f57881-0720-8321-a714-eef84194b083


- zap() for logging
- offline first
- hierarchical FSM(finite state machine) using loopfsm
- Github access ( or gitlab or bitbucket )
- otel for observability
- webportal and cli 
- co, lo, en (draw the digram)
- atlas and ent and postgres


get started :

1. start postgres, nats containers
at the root directory, run 
$ podman-compose up -d
2. execute postgres psql 
$ docker exec -it postgres psql -U postgres -d orchestration