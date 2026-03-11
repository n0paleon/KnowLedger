# Stage 1: Build
FROM golang:1.26.1 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

RUN apt-get update && apt-get install -y nodejs npm

COPY . .
RUN npm install -D tailwindcss @tailwindcss/cli
RUN npx @tailwindcss/cli -i ./web/static/assets/css/input.css -o ./web/static/assets/css/tailwind.css

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o main cmd/web/main.go

# Stage 2: Runtime
FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/main .

EXPOSE 8080
CMD ["./main"]