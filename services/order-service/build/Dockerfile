FROM golang:1.27.2-alpine as builder

WORKDIR /app

COPY services/order-service/go.mod services/order-service/go.sum ./

RUN go mod download

COPY services/order-service/ .

RUN CGO_ENABLED=0 GOOS=linux go build -v -ldflags="-s -w" -o /bin/order-service ./cmd

FROM alpine:3.24

WORKDIR /app

COPY --from=builder /bin/order-service .

ENTRYPOINT ["/app/order-service"]