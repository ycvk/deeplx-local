# 使用适合Go应用的基础镜像
FROM golang:1.22-alpine as builder
RUN apk update && apk add --no-cache upx make && rm -rf /var/cache/apk/*

# 设置工作目录
WORKDIR /app

# 复制所有文件到容器中
COPY . .

# 下载依赖
RUN go mod tidy

# 构建应用程序
#RUN CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} go build -o deeplx .
RUN make build-all
RUN upx -6 build/deeplx_linux_*

FROM alpine as final
ARG TARGETOS
ARG TARGETARCH
WORKDIR /usr/local/bin/
COPY --from=builder /app/build/deeplx_${TARGETOS}_${TARGETARCH} /usr/local/bin/deeplx
# 确保 url.txt 文件也被复制到容器中
COPY --from=builder /app/url.txt /usr/local/bin/url.txt

# 开放端口
EXPOSE 62155

# 运行
CMD ["./deeplx"]