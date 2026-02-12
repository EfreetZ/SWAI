# === 构建阶段 ===
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git make

WORKDIR /app

# 利用 Docker 缓存层：先复制依赖文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源码并构建
COPY . .
ARG BUILD_TARGET=./cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/server ${BUILD_TARGET}

# === 运行阶段 ===
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app
COPY --from=builder /app/server .

EXPOSE 8080

ENTRYPOINT ["./server"]
