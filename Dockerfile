FROM alpine:latest
LABEL authors="ycvk"

COPY deeplx /usr/local/bin/deeplx
COPY url.txt /usr/local/bin/url.txt
#COPY ./fofa/fofa_amd64 /usr/local/bin/fofa
#COPY ./fofa/fofa.sh /usr/local/bin/run_fofa.sh

# 设置执行权限
#RUN chmod +x /usr/local/bin/run_fofa.sh

WORKDIR /usr/local/bin
CMD ["./deeplx"]