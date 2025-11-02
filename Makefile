testenv:
	go run ./cmd/testenv --simulate-network=30s

eo-all:
	go run ./cmd/en --config ./configs/en_tiruvannamalai_host1.yaml &
	go run ./cmd/en --config ./configs/en_tiruvannamalai_host2.yaml &
	go run ./cmd/en --config ./configs/lo_chennai_host1.yaml &
	go run ./cmd/en --config ./configs/lo_chennai_host2.yaml &

lo-all:
	go run ./cmd/lo --config ./configs/lo_tiruvannamalai.yaml &
	go run ./cmd/lo --config ./configs/lo_chennai.yaml &

co-lo-all: start-nats
	go run ./cmd/co --config ./configs/co.yaml &
	$(MAKE) lo-all eo-all
