FROM alpine
ARG TARGETOS
ARG TARGETARCH
WORKDIR /usr/local/bin/
COPY build/deeplx_${TARGETOS}_${TARGETARCH} /usr/local/bin/deeplx
# 确保 url.txt 文件也被复制到容器中
COPY url.txt /usr/local/bin/url.txt

# 开放端口
EXPOSE 62155

# 运行
CMD ["./deeplx"]