FROM golang:1.26.1 AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /booking-app ./cmd/server

FROM alpine:latest
WORKDIR /app

RUN apk add --no-cache netcat-openbsd curl

RUN curl -fsSL https://github.com/pressly/goose/releases/download/v3.24.1/goose_linux_x86_64 -o /usr/local/bin/goose && \
    chmod +x /usr/local/bin/goose

COPY --from=builder /booking-app ./booking-app

COPY --from=builder /app/pkg/database/pg/migrations ./migrations

COPY --from=builder /app/entrypoint.sh ./entrypoint.sh

RUN chmod +x ./entrypoint.sh

EXPOSE 8080

ENTRYPOINT ["./entrypoint.sh"]
