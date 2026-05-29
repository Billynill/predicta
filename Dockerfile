# syntax=docker/dockerfile:1

FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /predicta ./cmd/api

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata wget

WORKDIR /app

COPY --from=builder /predicta /app/predicta
COPY api/openapi.yaml /app/api/openapi.yaml
COPY config/employees.json /app/config/employees.json
COPY config/employees.example.json /app/config/employees.example.json

ENV HTTP_PORT=8080 \
    OPENAPI_PATH=/app/api/openapi.yaml \
    EMPLOYEES_FILE=/app/config/employees.json \
    GIN_MODE=release

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD wget -qO- http://127.0.0.1:8080/health || exit 1

CMD ["/app/predicta"]
