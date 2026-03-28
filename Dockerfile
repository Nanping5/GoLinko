# GoLinko 服务 Dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app

# 安装依赖
RUN apk add --no-cache git

# 复制 go.mod 和 go.sum
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/main.go

# 运行阶段
FROM alpine:latest

WORKDIR /app

# 安装 ca-certificates（用于 HTTPS）
RUN apk --no-cache add ca-certificates tzdata

# 设置时区
ENV TZ=Asia/Shanghai

# 复制二进制文件
COPY --from=builder /app/main .
COPY --from=builder /app/configs ./configs

# 暴露端口
EXPOSE 8080

# 运行
CMD ["./main"]
