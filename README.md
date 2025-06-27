# 项目介绍

该项目包含钉钉互动卡片在各种业务场景下的使用例子，具体的卡片使用示例介绍以及**视频演示**参考这篇文档：[卡片使用示例](https://wolai.dingtalk.com/u5hVaGKEbw81Cc9zREDY6H)。

不知道钉钉互动卡片是什么？参考这篇文档：[感知卡片](https://wolai.dingtalk.com/upuadGKed3jjiQ9Y3bbFFQ)。

examples 目录下的每一个文件夹都是一个独立的卡片使用示例，一般包含以下这些内容：

- 卡片例子介绍 `README.md`
- 导出的卡片模板 JSON 文件 `{folder_name}.json`（可用于导入到自己的卡片模板里查看具体模板内容及相关配置，导入步骤如下）
  - 在[卡片平台](https://open-dev.dingtalk.com/fe/card)创建一个卡片模板
  - 右上角菜单按钮中点击 `更多` -> `导入` 按钮
- Python 代码示例 `python/index.py`
- Java 代码示例 `java/src/main/java/com/card/java/*.java`
- Node.js 代码示例 `nodejs/index.ts`
- Golang 代码示例 `golang/index.go`

代码示例中使用到的卡片相关接口文档：

- [创建并投放卡片](https://open.dingtalk.com/document/orgapp/create-and-deliver-cards)
- [更新卡片](https://open.dingtalk.com/document/orgapp/interactive-card-update-interface)
- [AI卡片流式更新](https://open.dingtalk.com/document/orgapp/api-streamingupdate)
- [事件回调](https://open.dingtalk.com/document/orgapp/event-callback-card)

# 如何启动？

所有示例代码都是通过 [服务端 Stream 模式](https://open.dingtalk.com/document/resourcedownload/introduction-to-stream-mode) 启动运行的。钉钉 Stream 模式可以用于多种场景的回调，包括事件订阅、机器人接收消息、卡片回调等。相关教程：[Stream Mode](https://opensource.dingtalk.com/developerpedia/docs/explore/tutorials/stream/overview)。

Stream 模式和 HTTP 模式仅仅是网络请求的方式和创建卡片时的 callbackType 参数不同，创建并投放卡片、更新卡片、对卡片事件回调的逻辑处理并没有任何不同。因此 HTTP 模式接入卡片的业务也可以参考该示例中的代码。

下面以 example/helloworld 为例，验证卡片开发环境是否已经成功配置。该例子需要完成[创建企业内部机器人](https://open.dingtalk.com/document/orgapp/the-creation-and-installation-of-the-application-robot-in-the)的前置流程，不同语言的聊天机器人教程参考：[聊天机器人](https://opensource.dingtalk.com/developerpedia/docs/category/%E8%81%8A%E5%A4%A9%E6%9C%BA%E5%99%A8%E4%BA%BA)。

该示例的视频演示：[hello world](https://wolai.dingtalk.com/89gp6tEDFQaXTM2RqDsd4f)。

以下编程语言（Python、Java、Golang、Bun 或 Node.js）的环境都可以通过 [asdf](https://asdf-vm.com/zh-hans/guide/introduction.html) 来安装并管理版本。

应用 client-id 和 client-secret 的**默认值**会分别从环境变量 `DINGTALK_APP_CLIENT_ID` 和 `DINGTALK_APP_CLIENT_SECRET` 中读取。

## Python

参考 Python 版本：3.10.13

依赖安装：

```bash
cd python
pip install -r requirements.txt
```

执行命令：

```bash
python index.py --client_id <client_id> --client_secret <client_secret>
```

## Java

参考 Java 版本：openjdk-21.0.2

依赖安装：

```bash
cd java
mvn clean install
```

在 `./src/main/resources/application.properties` 中添加应用凭据的配置：

```
dingtalk.app.client-id=<client_id>
dingtalk.app.client-secret=<client_secret>
```

执行命令：

```bash
mvn spring-boot:run
```

## Node.js

可以使用 Bun 直接运行 ts 文件，参考 Bun 版本：1.1.4

依赖安装：

```bash
cd nodejs
bun install
```

执行命令：

```bash
bun run index.ts --clientId <client_id> --clientSecret <client_secret>
```

## Golang

参考 Golang 版本：1.20

依赖安装：

```bash
cd golang
go mod tidy
```

执行命令：

```bash
go run index.go --client_id <client_id> --client_secret <client_secret>
```

# 注意事项

## 没有收到卡片交互组件的回传请求事件

Steram 没有收到卡片交互组件的回传请求事件回调通常可以从以下四个方面进行排查：

1. 检查一下是否注册 topic 为 `/v1.0/card/instances/callback` 的卡片回调。
2. 检查一下创建卡片时是否传入 callbackType="STREAM" 参数。
3. 检查一下创建卡片时用于生成 access_token 的 client-id 和注册回调服务使用的 client-id 是不是同一个。
4. 检查一下同一个 client-id 和 client-secret 是否启动了不止一个 Stream 服务。请务必保证一个 client-id 同一时间只启动一个 Stream 服务。如果有线上 Stream 服务在运行，希望在线下启动 Stream 服务开发调试，可以额外创建一个开发调试用的 client-id，线上线下环境分别设置系统环境变量使用不同的 client-id 进行隔离，避免相互干扰。

## 其它注意事项

[互动卡片 FAQ](https://open.dingtalk.com/document/orgapp/faq-card)


# 启动 FastAPI 服务

app 目录是 FastAPI 服务的启动目录，该服务启动的时候也会启动 stream 服务监听用户给机器人发送消息的事件。

启动后可以在 http://127.0.0.1:8000/docs 调用接口触发卡片投放，需先拷贝 .env.example 文件为 .env 文件，并填入下面这些信息：

- 自己的通义千问 API KEY `DASHSCOPE_API_KEY`
- 钉钉开放平台应用的 `CLIENT_ID` 和 `CLIENT_SECRET`
- 测试用的用户 ID `USER_ID`
- 上报数据到多维表的 webhook `WEBHOOK`

依赖安装：

```bash
pip install -r requirements.txt
```

执行命令：

```bash
uvicorn app.main:app
```

该项目使用到的卡片模板在 `app/card_templates` 目录下。

其中，`app/stream/chatbot_handler.py` 处理机器人接收消息，`app/stream/card_callback_handler.py` 处理卡片回传请求事件，`app/api/routes/notice.py` 主动给用户推送卡片。
