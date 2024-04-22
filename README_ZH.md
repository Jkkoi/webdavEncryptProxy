# Go 代理服务器

该项目用 Go 实现了一个基本的 HTTP/HTTPS 代理服务器。它在 CTR 模式下为 PUT 请求提供 AES 加密功能，为 GET 请求提供解密功能，并提供基本的身份验证和转发机制。

### 安装

要开始安装，请克隆软件源并构建可执行文件：

```bash
git clone [repository-url]
cd [repository-directory]
go build
```

## 使用方法

要启动服务器，可以使用以下命令：

```shell
./main -addr ":8080" -cert "path/to/certfile" -key "path/to/keyfile"
```

### 配置标志

- `-addr`： 设置服务器监听地址（默认：`:8080`）。
- `-cert`： 用于 HTTPS 的 SSL 证书文件。
- `-key`： 用于 HTTPS 的 SSL 密钥文件。

## API 端点

服务器代理请求并根据方法处理请求：

- **PUT**： 在转发前使用 AES CTR 模式加密正文。
- **GET**： 如果请求成功，则对响应进行解密。

需要基本身份验证凭据，并从请求头中进行解析。

## PR

欢迎PR！请随时提交拉取请求。

## 许可证

本项目采用 MIT License GPL V2 许可，详情请参见 LICENSE.md 文件。