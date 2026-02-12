.PHONY: help test lint fmt vet race clean docker-up docker-down

# 默认目标
help: ## 显示帮助信息
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# === 开发 ===

fmt: ## 格式化所有 Go 代码
	@find . -name "*.go" -exec gofmt -w {} \;

vet: ## 静态分析
	@for dir in $$(find projects -name "go.mod" -exec dirname {} \;); do \
		echo "==> vet $$dir"; \
		(cd $$dir && go vet ./...); \
	done

lint: ## golangci-lint 检查
	@for dir in $$(find projects -name "go.mod" -exec dirname {} \;); do \
		echo "==> lint $$dir"; \
		(cd $$dir && golangci-lint run ./...); \
	done

# === 测试 ===

test: ## 运行所有项目测试
	@for dir in $$(find projects -name "go.mod" -exec dirname {} \;); do \
		echo "==> test $$dir"; \
		(cd $$dir && go test ./...); \
	done

race: ## 运行测试（开启竞争检测）
	@for dir in $$(find projects -name "go.mod" -exec dirname {} \;); do \
		echo "==> race test $$dir"; \
		(cd $$dir && go test -race ./...); \
	done

bench: ## 运行基准测试
	@for dir in $$(find projects -name "go.mod" -exec dirname {} \;); do \
		echo "==> bench $$dir"; \
		(cd $$dir && go test -bench=. -benchmem ./...); \
	done

cover: ## 生成测试覆盖率报告
	@for dir in $$(find projects -name "go.mod" -exec dirname {} \;); do \
		echo "==> coverage $$dir"; \
		(cd $$dir && go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out -o coverage.html); \
	done

# === Docker ===

docker-up: ## 启动 Docker 服务（需指定 PROJECT，如 make docker-up PROJECT=phase2-mini-redis）
	@if [ -z "$(PROJECT)" ]; then echo "Usage: make docker-up PROJECT=<project-dir>"; exit 1; fi
	docker-compose -f projects/$(PROJECT)/docker-compose.yml up -d

docker-down: ## 停止 Docker 服务
	@if [ -z "$(PROJECT)" ]; then echo "Usage: make docker-down PROJECT=<project-dir>"; exit 1; fi
	docker-compose -f projects/$(PROJECT)/docker-compose.yml down

# === 清理 ===

clean: ## 清理构建产物
	@find . -name "*.test" -delete
	@find . -name "*.out" -delete
	@find . -name "*.html" -path "*/coverage.html" -delete
	@find . -name "bin" -type d -exec rm -rf {} + 2>/dev/null || true
