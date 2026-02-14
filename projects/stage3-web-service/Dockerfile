FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/server ./cmd/server

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /bin/server /app/server
COPY configs/config.yaml /app/configs/config.yaml

EXPOSE 8081
ENTRYPOINT ["/app/server"]
