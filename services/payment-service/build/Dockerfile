FROM golang:1.27.2-alpine as builder

WORKDIR /app

COPY services/payment-service/go.mod services/payment-service/go.sum ./

RUN go mod download

COPY services/payment-service/ .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/payment-service ./cmd

FROM alpine:3.24

WORKDIR /app

COPY --from=builder /bin/payment-service .

ENTRYPOINT ["/app/payment-service"]