package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
)

type passKey struct {
	Key      string `json:"key"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func handler(w http.ResponseWriter, r *http.Request) {
	/* 解析basic auth，获取密钥
	user 为上游地址
	pass 结构如上
	*/
	//TODO: 其他认证方式
	user, pass, ok := r.BasicAuth()
	if !ok {
		w.Header().Set("WWW-Authenticate", "Basic")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	userBytes, _ := base64.StdEncoding.DecodeString(user)
	passBytes, _ := base64.StdEncoding.DecodeString(pass)
	user = string(userBytes)
	pass = string(passBytes)
	passKeyStruct := &passKey{}
	err := json.Unmarshal([]byte(pass), passKeyStruct)
	if err != nil {
		http.Error(w, "Failed to parse auth", http.StatusInternalServerError)
	}
	key, err := base64.StdEncoding.DecodeString(passKeyStruct.Key)
	if err != nil {
		http.Error(w, "Failed to parse cipher", http.StatusInternalServerError)
	}
	targetURL := user
	r.SetBasicAuth(passKeyStruct.Username, passKeyStruct.Password)

	var RequestBodyReader io.Reader = nil
	if r.Method == http.MethodPut {
		// 配置AES CTR模式
		block, err := aes.NewCipher(key)
		iv := make([]byte, aes.BlockSize)
		// 从 crypto/rand 生成随机 IV
		if _, err := io.ReadFull(rand.Reader, iv); err != nil {
			panic(err)
		}

		ivReader := bytes.NewReader(iv)
		if err != nil {
			http.Error(w, "Failed to create cipher", http.StatusInternalServerError)
			return
		}
		stream := cipher.NewCTR(block, iv)
		encryptedReader := &cipher.StreamReader{
			S: stream,
			R: r.Body,
		}
		combinedReader := io.MultiReader(ivReader, encryptedReader)
		RequestBodyReader = combinedReader

		// 修改 http.content_length_header
		oLength, _ := strconv.ParseInt(r.Header.Get("Content-Length"), 10, 64)
		newLength := oLength + 16
		r.Header.Set("Content-Length", strconv.FormatInt(newLength, 10))
	} else {
		RequestBodyReader = r.Body
	}

	req, err := http.NewRequest(r.Method, targetURL+r.URL.Path, RequestBodyReader)
	if err != nil {
		http.Error(w, "Request creation failed", http.StatusInternalServerError)
		return
	}
	// 复制原始请求的头部信息
	req.Header = make(http.Header)
	for h, val := range r.Header {
		req.Header[h] = val
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Request failed", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	// 构造向目标URL的请求

	var RespBodyWriter io.Writer = nil
	if r.Method == http.MethodGet &&
		// 判定为成功再解密，否则直接转发
		resp.StatusCode == http.StatusOK {
		// 配置AES CTR模式
		block, err := aes.NewCipher(key)
		if err != nil {
			http.Error(w, "Failed to create cipher", http.StatusInternalServerError)
			return
		}
		// 创建一个16字节的缓冲区来存储 IV
		iv := make([]byte, 16)
		// 从 resp.Body 读取16字节到 iv
		_, err = resp.Body.Read(iv)
		fmt.Println(iv, resp.StatusCode)
		if err != nil {
			println("failed to read IV: %v", err)
		}
		stream := cipher.NewCTR(block, iv)
		decryptedReader := &cipher.StreamWriter{
			S: stream,
			W: w,
		}
		RespBodyWriter = decryptedReader
		// 修改 http.content_length_header
		oLength, _ := strconv.ParseInt(resp.Header.Get("Content-Length"), 10, 64)
		newLength := oLength - 16
		resp.Header.Set("Content-Length", strconv.FormatInt(newLength, 10))
		//RespBodyWriter = w
	} else {
		RespBodyWriter = w

	}
	// 复制响应头部到原始响应中
	for h, val := range resp.Header {
		w.Header()[h] = val
	}
	w.WriteHeader(resp.StatusCode)
	_, err = io.Copy(RespBodyWriter, resp.Body)
	if err != nil {
		println(err.Error())
	}

}

func main() {
	http.HandleFunc("/", handler)

	// 定义命令行参数
	var addr string
	var certFile string
	var keyFile string
	var uid string
	flag.StringVar(&addr, "addr", ":8080", "HTTP network address")
	flag.StringVar(&certFile, "cert", "", "SSL certificate file")
	flag.StringVar(&keyFile, "key", "", "SSL key file")
	flag.StringVar(&uid, "uid", "", "Linux Only")
	flag.Parse() // 解析命令行参数

	// 根据是否提供了证书和密钥决定启动HTTP还是HTTPS服务器
	if certFile != "" && keyFile != "" {
		// 检查文件是否存在
		if _, err := os.Stat(certFile); os.IsNotExist(err) {
			log.Fatalf("Certificate file '%s' not found.", certFile)
		}
		if _, err := os.Stat(keyFile); os.IsNotExist(err) {
			log.Fatalf("Key file '%s' not found.", keyFile)
		}

		log.Printf("Starting HTTPS server on %s", addr)
		log.Fatal(http.ListenAndServeTLS(addr, certFile, keyFile, nil))
	} else {
		log.Printf("Starting HTTP server on %s", addr)
		log.Fatal(http.ListenAndServe(addr, nil))
	}
}
