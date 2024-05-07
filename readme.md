## deeplx-local

用于提供给沉浸式翻译的在本地运行 deeplx 的工具。

通过并发请求存在`url.txt`内的 deeplx 的翻译接口，来获取低延迟、可用的url。


初步实现了负载均衡，延迟越低响应越快的接口会被优先使用。

### 一键启动
`docker run -itd -p 8080:62155 neccen/deeplx-local:latest`


翻译地址：`http://localhost:8080`

### 可选参数
- `360_api_key`：[quake360](https://quake.360.net/quake/#/personal?tab=message)的api_key，可用于每日自动爬取更多的翻译接口。（有每日免费次数）
- `hunter_api_key`：[鹰图](https://hunter.qianxin.com/home/myInfo)的api_key，可用于每日自动爬取更多的翻译接口。(有每日免费次数)
- 群友提到的`fofa`不想加，不送免费额度啊，本来想用[Cl0udG0d/Fofa-hack: 非付费会员，fofa数据采集工具](https://github.com/Cl0udG0d/Fofa-hack)偷个懒，发现不传自己的auth会有20条搜索的限制，懒得整了

### 使用方法
#### 编译运行
1. 填入`url.txt`内的 deeplx 的翻译接口，可以只填`ip:port`，也可以填写完整的url。
2. `go build -o deeplx .`来编译。
3. 启动编译后的程序，翻译地址为 `http://localhost:62155/translate` ，端口可自行修改。

#### 本地运行
1. 在Release中下载对应平台的二进制文件。
2. 在可执行文件的目录下，新建`url.txt`, 填入`url.txt`内的 deeplx 的翻译接口，可以只填`ip:port`，也可以填写完整的url。
3. 启动程序，翻译地址为 `http://localhost:62155/translate`

#### Docker Compose 自编译运行
1. 填入`url.txt`内的 deeplx 的翻译接口，可以只填`ip:port`，也可以填写完整的url。
2. `docker compose up -d`来构建并启动容器，`docker-compose.yml`中的配置和端口可自行更改。

#### Docker Compose 运行
```yaml
version: '3.8'
services:
  deeplx:
    image: neccen/deeplx-local:latest
    ports:
      - "62155:62155"
    volumes:
      - /url.txt文件目录:/usr/local/bin/url.txt  # 本地url.txt文件目录,删除此行则使用内置的已经爬取的deeplx翻译接口
    environment:
      - 360_api_key=xxxxx  # 可选
      - hunter_api_key=xxxxx  # 可选
    container_name: deeplx
    restart: unless-stopped
```

#### Docker 运行
##### 完整命令:

`docker run -itd -p 62155:62155 -v /url.txt文件目录:/usr/local/bin/url.txt -e 360_api_key="xxxxx" neccen/deeplx-local:latest`

##### 极简命令: 
**会自动使用我内置的爬取的deeplx翻译接口**



`docker run -itd -p 62155:62155 neccen/deeplx-local:latest`


### Bob翻译插件
看到有人提问Bob如何用上此翻译，我手撸了一个，
配套的Bob翻译插件请看 [ycvk/deeplx-local-bobplugin: 用于自建deeplx服务的bob翻译插件](https://github.com/ycvk/deeplx-local-bobplugin)