package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/golang/glog"
)

func init() {
	// 初始化日志文件
	f, _ := os.OpenFile("./log.txt", os.O_WRONLY|os.O_CREATE|os.O_SYNC,
		0755)
	os.Stdout = f
	os.Stderr = f
}

func main() {
	flag.Set("v", "4")

	http.HandleFunc("/", rootHandler)

	// 4.当范文localhost/healthz时，应当返回200
	http.HandleFunc("/healthz", healthz)

	glog.V(2).Info("Starting http server...")
	glog.Flush()

	err := http.ListenAndServe(":8001", nil)
	if err != nil {
		defer glog.Flush()
		glog.Fatal(err)
	}
}

// healthz处理函数
func healthz(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "ok\n")
	defer glog.Flush()
	glog.V(2).Info(os.Stdout, fmt.Sprintf("healthz IP: %s, StateCode:%d", getUserIP(w, r), 200))
}

// 根目录处理函数
func rootHandler(w http.ResponseWriter, r *http.Request) {
	// 发生错误后日志记录状态码为500
	defer func() {
		if err := recover(); err != nil {
			defer glog.Flush()
			glog.V(2).Info(os.Stdout, fmt.Sprintf("rootHandler IP: %s, StateCode:%d", getUserIP(w, r), 500))
		}
	}()

	// 1.接收客户端request，并将request中带的header写入response header
	for k, v := range r.Header {
		for _, val := range v {
			w.Header().Set(k, val)
		}
	}
	// 2.读取当前系统的环境变量中VERSION配置，并写入response header
	w.Header().Set("VERSION", os.Getenv("VERSION"))
	io.WriteString(w, "root\n")

	// panic(errors.New("Something wrong!"))

	// 3.Server端记录访问日志包括客户端ip, HTTP返回码，输出到server端标准输出
	defer glog.Flush()
	glog.V(2).Info(os.Stdout, fmt.Sprintf("rootHandler IP: %s, StateCode:%d", getUserIP(w, r), 200))

}

// 获取用户真实IP地址
func getUserIP(w http.ResponseWriter, r *http.Request) string {
	var userIP string
	if len(r.Header.Get("X-Forwarded-For")) > 1 {
		userIP = r.Header.Get("X-Forwarded-For")
	} else if len(r.Header.Get("X-Real-IP")) > 1 {
		userIP = r.Header.Get("X-Real-IP")
	} else {
		if strings.Contains(r.RemoteAddr, ":") {
			userIP = r.RemoteAddr[0:strings.LastIndex(r.RemoteAddr, ":")]
		} else {
			userIP = r.RemoteAddr
		}
	}
	return userIP
}
