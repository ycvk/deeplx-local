## deeplx-local

用于提供给沉浸式翻译的在本地运行 deeplx 的工具。

通过并发请求存在`url.txt`内的 deeplx 的翻译接口，来获取低延迟、可用的url。

翻译超大文本时，会自动做拆分并行翻译合并处理。

### 一键启动
`docker run --pull=always -itd -p 8080:62155 neccen/deeplx-local:latest`


**翻译地址：**`http://localhost:8080/translate`


添加 `route`环境变量可修改默认翻译地址, 如：

`docker run --pull=always -itd -p 8080:62155 -e route=abc neccen/deeplx-local:latest`

翻译地址为`http://localhost:8080/abc`

### 可选参数
- `route`：默认为`/translate`，可自定义翻译地址。比如设置为 `abc`，则翻译地址为 `http://localhost:8080/abc`
- `360_api_key`：[quake360](https://quake.360.net/quake/#/personal?tab=message)的api_key，可用于每日自动爬取更多的翻译接口。（有每日免费次数）
- `hunter_api_key`：[鹰图](https://hunter.qianxin.com/home/myInfo)的api_key，可用于每日自动爬取更多的翻译接口。(有每日免费次数)
- 提到的`fofa`不想加，不送免费额度，本来想用[Cl0udG0d/Fofa-hack: 非付费会员，fofa数据采集工具](https://github.com/Cl0udG0d/Fofa-hack)偷个懒，发现不传自己的auth会有20条搜索的限制，不整了

### 使用方法

#### 1. Docker 运行
##### 极简命令:
**会自动使用我内置的爬取的deeplx翻译接口**



`docker run -itd -p 8080:62155 neccen/deeplx-local:latest`

翻译地址为 `http://localhost:8080/translate`

##### 完整命令:

`docker run -itd -p 8080:62155 -v /url.txt文件目录:/usr/local/bin/url.txt -e route xxx -e 360_api_key="xxxxx" neccen/deeplx-local:latest`


#### 2. Docker Compose 运行
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
      - route=xxx  # 可选 默认为 /translate
      - 360_api_key=xxxxx  # 可选
      - hunter_api_key=xxxxx  # 可选
    container_name: deeplx
    restart: unless-stopped
```

#### 3. 本地运行
1. 在Release中下载对应平台的二进制文件。
2. 在可执行文件的目录下，新建`url.txt`, 填入`url.txt`内的 deeplx 的翻译接口，可以只填`ip:port`，也可以填写完整的url。
3. 启动程序，翻译地址为 `http://localhost:62155/translate`

#### 4. Windows后台运行
1. 在Release中下载`windows-xxx`标识的发行版，如[windows-v0.1.2](https://github.com/ycvk/deeplx-local/releases/tag/windows-v0.1.2)
2. 解压打开`.exe`文件后，会自动在后台启动，托盘可以看到服务图标![2fa59a5c188a7e02041948d7b6918e83.png](https://i.mji.rip/2024/05/08/2fa59a5c188a7e02041948d7b6918e83.png)
3. **注意防火墙可能会提示是否允许联网**，点击是
4. 不想使用需要关闭时，点击图标，点击`quit`即可
5. 沉浸式翻译地址为`http://localhost:62155/translate`


### Bob翻译插件
看到有人提问Bob如何用上此翻译，我手撸了一个，
配套的Bob翻译插件请看 [ycvk/deeplx-local-bobplugin: 用于自建deeplx服务的bob翻译插件](https://github.com/ycvk/deeplx-local-bobplugin)

### 自行抓取url方法

目前网络上的自建deepl服务有很多，我列举几个开源项目：
- [OwO-Network/DeepLX: DeepL Free API (No TOKEN required)](https://github.com/OwO-Network/DeepLX)
- [xiaozhou26/deeplx-pro](https://github.com/xiaozhou26/deeplx-pro/tree/main)
- [ifyour/deeplx-for-cloudflare: 🔥 Deploy DeepLX on Cloudflare](https://github.com/ifyour/deeplx-for-cloudflare)

1. 分析前者的代码可以发现，它暴露了一个根路径的 `get` 接口，返回固定的响应：
```json
{ 
  "code": 200, 
  "message": "DeepL Free API, Developed by sjlleo and missuo. Go to /translate with POST. http://github.com/OwO-Network/DeepLX"
}
```
以此可以通过各种测绘工具通过这个特征去抓取使用此服务搭建的翻译接口。

以fofa搜索为例，搜索框输入：
```
body='{"code":200,"message":"DeepL Free API, Developed by sjlleo and missuo. Go to /translate with POST. http://github.com/OwO-Network/DeepLX"}'
```
原项目代码如下：

https://github.com/OwO-Network/DeepLX/blob/93a3204eab4366b467ac6e2672b0f6186d435837/main.go#L78-L83

2. 同理，分析后者的代码可以发现，它同样暴露了一个根路径的 `get` 接口，返回固定的响应：
```
Welcome to deeplx-pro
```
以此可以通过各种测绘工具通过这个特征去抓取使用此服务搭建的翻译接口。

以quake搜索为例，搜索框输入：
```
response:"Welcome to deeplx-pro"
```
原项目代码如下：

https://github.com/xiaozhou26/deeplx-pro/blob/70fc070b21d14b136c69ac172a8e060fc547ed9b/server.js#L11-L13


3. 也一样
```
{
"code": 200,
"message": "Free translation API, Use POST method to access /translate."
}
```