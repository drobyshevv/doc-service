FROM golang:1.25 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# service
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o auth-service ./cmd/auth/server

# migrator
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o auth-migrator ./cmd/auth/migrator


FROM alpine:latest

WORKDIR /app

RUN apk --no-cache add ca-certificates

COPY --from=builder /app/auth-service .
COPY --from=builder /app/auth-migrator .

COPY configs ./configs
COPY migrations ./migrations

EXPOSE 8082

CMD ["./auth-service"]