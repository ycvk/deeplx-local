# 使用适合Go应用的基础镜像
FROM golang:latest as builder

# 设置工作目录
WORKDIR /app

# 复制所有文件到容器中
COPY . .

# 下载依赖
RUN go mod tidy

# 构建应用程序
RUN CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} go build -o deeplx .

# 运行阶段
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/deeplx .
# 确保 url.txt 文件也被复制到容器中
COPY url.txt ./

##COPY ./fofa/fofa_amd64 ./
##COPY ./fofa/fofa.sh ./
#
## 设置执行权限
##RUN chmod +x /usr/local/bin/run_fofa.sh

# 开放端口
EXPOSE 62155

# 运行你的程序
CMD ["./deeplx"]