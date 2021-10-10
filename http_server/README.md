借助Gin框架的中间件机制，实现日志保存和请求头的复制功能；为了提高服务可用性，实现了优雅关停，在正常终止服务时，等待未完成的服务，如果60秒后仍未执行完，再停止服务。



**1.接收客户端 request，并将 request 中带的 header 写入 response header**

**2.读取当前系统的环境变量中的 VERSION 配置，并写入 response header**

通过实现自定义中间件处理函数，复制请求头到相应头中，并添加系统环境变量VERSION，通过`router.Use(SetHeader())`注册后作用于所有请求。

```go
func SetHeader() gin.HandlerFunc {
	return func(c *gin.Context) {

		// 1.接收客户端request，并将request中带的header写入response header
		for k, v := range c.Request.Header {
			for _, val := range v {
				c.Writer.Header().Set(k, val)
			}
		}

		// 2.读取当前系统的环境变量中VERSION配置，并写入response header
		c.Writer.Header().Set("VERSION", os.Getenv("VERSION"))
		// 处理请求
		c.Next()

	}
}	 

```

**3.Server 端记录访问日志包括客户端 IP，HTTP 返回码，输出到 server 端的标准输出**

与请求头复制一样，日志记录也通过自定义中间件实现特殊记录需求，默认存储在程序目录logs下（`http_server/logs/`），日志文件使用日期作为文件名称。

**4.当访问 localhost/healthz 时，应返回200**

```shell
$ curl -XGET -i localhost/healthz
 
HTTP/1.1 200 OK
Accept: */*
User-Agent: curl/7.68.0
Version: v1.0
Date: Sun, 10 Oct 2021 14:29:56 GMT
Content-Length: 0
```



为了测试程序还简单实现了两个路由，测试正常请求和异常请求。

```shell
$ curl -XGET -i localhost/test

HTTP/1.1 200 OK
Accept: */*
Content-Type: application/json; charset=utf-8
User-Agent: curl/7.68.0
Version: v1.0
Date: Sun, 10 Oct 2021 14:29:53 GMT
Content-Length: 4
```

```shell
$ curl -XGET -i localhost/error

HTTP/1.1 500 Internal Server Error
Accept: */*
User-Agent: curl/7.68.0
Version: v1.0
Date: Sun, 10 Oct 2021 14:29:54 GMT
Content-Length: 0
```

