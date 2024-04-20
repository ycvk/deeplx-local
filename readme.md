## deeplx-local

用于提供给沉浸式翻译的在本地运行 deeplx 的工具。

通过并发请求存在`url.txt`内的 deeplx 的翻译接口，来获取低延迟、可用的url。


初步实现了负载均衡，延迟越低响应越快的接口会被优先使用。

### 使用方法
1. 填入`url.txt`内的 deeplx 的翻译接口，可以只填`ip:port`，也可以填写完整的url。
2. `go build -o deeplx .`来编译。
3. 启动后的翻译地址为 `http://localhost:62155/translate` ，端口可自行修改。