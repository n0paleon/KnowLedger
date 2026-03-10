# Stage 1: Build
FROM golang:1.26.1 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o main cmd/web/main.go

# Stage 2: Runtime
FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/main .

EXPOSE 8080
CMD ["./main"]