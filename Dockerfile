FROM golang:1.23-bookworm as builder

RUN apt-get update && apt-get install -y \
    pkg-config \
    libssl-dev

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o kvstore ./cmd/kvstore

FROM debian:bookworm-slim
WORKDIR /app
RUN apt-get update && apt-get install -y \
    libssl3 \
    ca-certificates && \
    rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/kvstore /app/kvstore

EXPOSE 8080
ENTRYPOINT ["/app/kvstore"]
