.PHONY: build run test lint fmt vet race bench cover clean docker-build docker-run docker-down

APP_NAME := server
BUILD_DIR := bin

# === 构建 ===
build:
	go build -ldflags="-s -w" -o $(BUILD_DIR)/$(APP_NAME) ./cmd/server

run: build
	./$(BUILD_DIR)/$(APP_NAME)

# === 测试 ===
test:
	go test ./... -v -count=1

# 竞争检测：检测并发数据竞争
race:
	go test -race ./... -v

# 基准测试：测量函数性能
bench:
	go test -bench=. -benchmem ./...

# 覆盖率报告
cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# === 代码质量 ===
lint:
	golangci-lint run ./...

fmt:
	gofmt -w .

vet:
	go vet ./...

# === Docker ===
docker-build:
	docker build -t $(APP_NAME):latest .

docker-run:
	docker-compose up -d

docker-down:
	docker-compose down

# === 清理 ===
clean:
	rm -rf $(BUILD_DIR) coverage.out coverage.html
