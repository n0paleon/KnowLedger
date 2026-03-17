# Stage 1: Tailwind CSS builder
FROM alpine:latest AS tw-builder

ARG TAILWIND_VERSION=v4.2.1
RUN wget -qO /usr/local/bin/tailwindcss \
    https://github.com/tailwindlabs/tailwindcss/releases/download/${TAILWIND_VERSION}/tailwindcss-linux-x64 \
    && chmod +x /usr/local/bin/tailwindcss

# Stage 2: Build
FROM golang:1.26.1 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY --from=tw-builder /usr/local/bin/tailwindcss /usr/local/bin/tailwindcss

COPY . .

# Build CSS
RUN tailwindcss -i ./web/static/assets/css/input.css \
                -o ./web/static/assets/css/tailwind.css \
                --minify

# Build web
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o main cmd/web/main.go

# Build admincli
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o admincli cmd/admincli/main.go

# Stage 3: Runtime
FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/main .
COPY --from=builder /app/admincli .

EXPOSE 3000
CMD ["./main"]
