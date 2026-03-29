#STAGE 1
FROM golang:1.26.1-alpine3.23 AS builder

ARG SQLC_VERSION=v1.30.0

WORKDIR /build

COPY go.mod go.sum ./

RUN go mod download

RUN go install github.com/sqlc-dev/sqlc/cmd/sqlc@${SQLC_VERSION}

COPY . .

RUN /go/bin/sqlc generate -f sqlc.yaml

#собираем статический бинарник из исходных файлов 

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o /build/api \
    ./cmd/api

#STAGE 2

FROM alpine:3.23.3

# Устанавливаем CA certificates для HTTPS запросов
# tzdata для корректной работы с timezone

RUN apk --no-cache add ca-certificates tzdata

RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /app

COPY --from=builder /build/api /app/api

COPY --from=builder /build/internal/database/migrations /app/internal/database/migrations

RUN chmod +x /app/api

USER appuser

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

CMD ["/app/api"]