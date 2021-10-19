# syntax=docker/dockerfile:1
# 构建阶段
FROM golang:1.17-alpine AS build

# 设置go mod代理和http server的环境变量VERSION
ENV GO111MODULE=on \
          GOPROXY=https://goproxy.cn,direct \
          VERSION=v1.0

#设置工作目录
WORKDIR /app

# 下载go mod
COPY /http_server/go.mod .
COPY /http_server/go.sum .
RUN go mod download

# 复制源代码
COPY /http_server/*.go .

# 编译go程序
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o httpserver /app
 
# 部署阶段
FROM alpine:latest AS prod

# 复制已编译程序
COPY --from=build /app /http-server 

# 设置国内源，安装curl
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories \
         && apk upgrade --update \
         && apk add curl \
         && rm -rf /var/cache/apk/*

# 设置健康检查
HEALTHCHECK --interval=5s --timeout=3s --retries=3 \
  CMD curl -fs http://localhost/healthz || exit 1

# 将日志文件保存在宿主机
VOLUME /logs

# 容器对外端口设置
EXPOSE 80
# 入口点设置启动程序
ENTRYPOINT ["/http-server/httpserver"]