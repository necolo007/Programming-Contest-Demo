# 构建阶段
FROM golang:1.21-alpine AS builder

WORKDIR /app

# 安装依赖
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/main.go

# 运行阶段
FROM alpine:latest

WORKDIR /app

# 安装基本工具和时区数据
RUN apk --no-cache add tzdata ca-certificates

# 从构建阶段复制二进制文件
COPY --from=builder /app/main .

# 设置时区
ENV TZ=Asia/Shanghai

# 暴露端口
EXPOSE 12349

# 运行应用
CMD ["./main"]