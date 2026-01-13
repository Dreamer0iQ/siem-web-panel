FROM golang:1.21-alpine AS builder
RUN apk add --no-cache git make
WORKDIR /app
COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o siem-agent ./cmd/agent/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata bash
RUN addgroup -g 1000 siem && \
    adduser -D -u 1000 -G siem siem

WORKDIR /app

COPY --from=builder /app/siem-agent .

COPY configs/agent.yaml /app/configs/
RUN mkdir -p /app/logs && \
    chown -R siem:siem /app

USER siem

CMD ["./siem-agent", "-config", "/app/configs/agent.yaml"]
