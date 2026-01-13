FROM golang:1.21-alpine AS builder

RUN apk add --no-cache git make
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o siem-server ./cmd/api/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

# Создание непривилегированного пользователя
RUN addgroup -g 1000 siem && \
    adduser -D -u 1000 -G siem siem

WORKDIR /app

COPY --from=builder /app/siem-server .

RUN mkdir -p /app/data && \
    chown -R siem:siem /app

USER siem

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/health || exit 1

CMD ["./siem-server", "-port", "8080", "-data", "/app/data"]
