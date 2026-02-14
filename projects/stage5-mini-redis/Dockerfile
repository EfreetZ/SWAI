FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/mini-redis ./cmd/server

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /bin/mini-redis /app/mini-redis
EXPOSE 16379
ENTRYPOINT ["/app/mini-redis"]
