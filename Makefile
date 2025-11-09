.PHONY: test

test:
	@echo "🧪 Running integration tests..."
	mkdir -p logs
	go test ./tests/integration -v | tee logs/test.log
