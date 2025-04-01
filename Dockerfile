# 构建阶段
FROM golang:1.23.2-alpine AS builder

WORKDIR /app

# 设置 GOPROXY 使用国内镜像
ENV GOPROXY=https://goproxy.cn,direct

# 安装依赖
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码和配置文件
COPY . .

COPY 民法典.csv /app/民法典.csv

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# 运行阶段
FROM alpine:latest

WORKDIR /app

# 安装基本工具和时区数据
RUN apk --no-cache add tzdata ca-certificates

# 从构建阶段复制二进制文件和配置文件
COPY --from=builder /app/main .
COPY --from=builder /app/config ./config
COPY --from=builder /app/民法典.csv ./民法典.csv

# 设置时区
ENV TZ=Asia/Shanghai

# 暴露端口
EXPOSE 12349

# 运行应用
CMD ["./main"]