# Stage 1: Build
FROM golang:alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o app ./cmd/main.go

# Stage 2: Run
FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/app .

# Jika butuh .env, copy juga:
# COPY --from=builder /app/.env .env

EXPOSE 8080
CMD ["./app"]
