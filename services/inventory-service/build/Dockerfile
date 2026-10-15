FROM golang:1.27.2-alpine AS builder

WORKDIR /app

COPY services/inventory-service/go.mod services/inventory-service/go.sum ./

RUN go mod download

COPY services/inventory-service/ .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/inventory-service ./cmd

FROM alpine:3.24

WORKDIR /app

COPY --from=builder /bin/inventory-service .

ENTRYPOINT ["/app/inventory-service"]