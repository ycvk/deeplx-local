## deeplx-local

用于提供给沉浸式翻译的在本地运行 deeplx 的工具。

通过并发请求存在`url.txt`内的 deeplx 的翻译接口，来获取低延迟、可用的url。


初步实现了负载均衡，延迟越低响应越快的接口会被优先使用。

### 使用方法
#### 本地运行
1. 填入`url.txt`内的 deeplx 的翻译接口，可以只填`ip:port`，也可以填写完整的url。
2. `go build -o deeplx .`来编译。
3. 启动编译后的程序，翻译地址为 `http://localhost:62155/translate` ，端口可自行修改。

#### Docker Compose运行
1. 填入`url.txt`内的 deeplx 的翻译接口，可以只填`ip:port`，也可以填写完整的url。
2. `docker compose up -d`来构建并启动容器，`docker-compose.yml`中的配置和端口可自行更改。

#### Docker 运行
`docker run -itd -p 62155:62155 neccen/deeplx-local:latest`