# Go Proxy Server

[简体中文](README_ZH.md)

This project implements a basic HTTP/HTTPS proxy server in Go. It offers encryption capabilities with AES in CTR mode for PUT requests and decryption for GET requests, along with basic authentication and forwarding mechanisms.

## Installation

To get started, clone the repository and build the executable:

```bash
git clone [repository-url]
cd [repository-directory]
go build
```

## Usage

To start the server, you can use the following command:

```shell
go run main -addr ":8080" -cert "path/to/certfile" -key "path/to/keyfile"
```

### Configuration Flags

- `-addr`: Set the server listening address (default: `:8080`).
- `-cert`: SSL certificate file for HTTPS.
- `-key`: SSL key file for HTTPS.

## API Endpoints

The server proxies requests and handles them based on the method:

- **PUT**: Encrypts the body using AES CTR mode before forwarding.
- **GET**: Decrypts the response if the request was successful.

Basic authentication credentials are required and parsed from the request header.

## Contributing

Contributions are welcome! Please feel free to submit pull requests.

## License

This project is licensed under the MIT License GPL V2 - see the LICENSE.md file for details.

