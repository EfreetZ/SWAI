.PHONY: build run-server run-client test race bench perf stress chaos fmt vet clean docker-build docker-run docker-down

SERVER_BIN := mini-rpc-server
CLIENT_BIN := mini-rpc-client
BUILD_DIR := bin

build:
	go build -o $(BUILD_DIR)/$(SERVER_BIN) ./cmd/server
	go build -o $(BUILD_DIR)/$(CLIENT_BIN) ./cmd/client

run-server: build
	./$(BUILD_DIR)/$(SERVER_BIN)

run-client: build
	./$(BUILD_DIR)/$(CLIENT_BIN)

test:
	go test ./... -v -count=1

race:
	go test -race ./... -v

bench:
	go test -bench=. -benchmem -run=^$ ./...

perf:
	go test ./test/performance -v -count=1

stress:
	go test ./test/stress -v -count=1

chaos:
	go test ./test/chaos -v -count=1

fmt:
	gofmt -w .

vet:
	go vet ./...

clean:
	rm -rf $(BUILD_DIR)
	go clean -testcache

docker-build:
	docker build -t mini-rpc:latest .

docker-run:
	docker-compose up -d

docker-down:
	docker-compose down
