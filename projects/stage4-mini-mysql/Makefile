.PHONY: build run test race bench perf stress chaos fmt vet clean docker-build docker-run docker-down

APP_NAME := mini-mysql
BUILD_DIR := bin

build:
	go build -o $(BUILD_DIR)/$(APP_NAME) ./cmd/server

run: build
	./$(BUILD_DIR)/$(APP_NAME)

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
	docker build -t $(APP_NAME):latest .

docker-run:
	docker-compose up -d

docker-down:
	docker-compose down
