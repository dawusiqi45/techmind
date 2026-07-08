# TechMind 通用构建镜像
# 构建 Server: docker build --target server -t techmind-server .
# 构建 Worker: docker build --target worker -t techmind-worker .

FROM golang:1.25-alpine AS builder

ENV GOPROXY=https://goproxy.cn,direct

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# 构建 Server
FROM builder AS server-builder
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o techmind-server ./cmd/server

# 构建 Worker
FROM builder AS worker-builder
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o techmind-worker ./cmd/worker

# Server 最终镜像
FROM alpine:latest AS server
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=server-builder /app/techmind-server /app/techmind-server
COPY config/ ./config/
EXPOSE 8080
CMD ["/app/techmind-server"]

# Worker 最终镜像
FROM alpine:latest AS worker
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=worker-builder /app/techmind-worker /app/techmind-worker
COPY config/ ./config/
CMD ["/app/techmind-worker"]