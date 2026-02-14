FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/mini-mq-broker ./cmd/broker

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /bin/mini-mq-broker /app/mini-mq-broker
EXPOSE 19092
ENTRYPOINT ["/app/mini-mq-broker"]
