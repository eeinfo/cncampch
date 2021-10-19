package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

func main() {
	var eg errgroup.Group
	router := gin.Default()
	// 中间件作用于所有的HTTP请求
	router.Use(SetHeader(), LoggerToFile(), gin.Recovery())
	// 注册路由
	router.GET("/healthz", healthzHandler)
	router.GET("/error", errorHandler)
	router.GET("/test", testHandler)
	server := &http.Server{
		Addr:         ":80",
		Handler:      router,
		ReadTimeout:  50 * time.Second,
		WriteTimeout: 100 * time.Second}

	eg.Go(func() error {
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
		return err
	})

	// 实现优雅关闭, 尽量等待请求完成，如果超过60秒请求未完成，则退出
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}
	log.Println("Server exiting")
}

// 测试正常路由
func testHandler(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, nil)
}

// 测试异常路由
func errorHandler(c *gin.Context) {
	panic(errors.New("an error occurred"))
}

// 4.当范文localhost/healthz时，应当返回200
func healthzHandler(c *gin.Context) {
	c.Status(http.StatusOK)
}

// 中间件-将日志写入文件
func LoggerToFile() gin.HandlerFunc {
	logger := Logger()
	return func(c *gin.Context) {
		// 开始时间
		startTime := time.Now()

		// 处理请求
		c.Next()

		// 结束时间
		endTime := time.Now()

		// 执行时间
		latencyTime := endTime.Sub(startTime)

		// 请求方式
		reqMethod := c.Request.Method

		// 请求路由
		reqUri := c.Request.RequestURI

		// 状态码
		statusCode := c.Writer.Status()

		// 请求IP
		clientIP := c.ClientIP()

		//日志格式
		logger.Infof("| %3d | %13v | %15s | %s | %s |",
			statusCode,
			latencyTime,
			clientIP,
			reqMethod,
			reqUri,
		)
	}
}

// 中间件-复制request header到 response header
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

// 记录日志，主要包括客户端IP，相应状态码
func Logger() *logrus.Logger {
	now := time.Now()
	logFilePath := ""
	if dir, err := os.Getwd(); err == nil {
		logFilePath = dir + "/logs/"
	}
	if err := os.MkdirAll(logFilePath, 0777); err != nil {
		fmt.Println(err.Error())
	}
	logFileName := now.Format("2006-01-02") + ".log"
	//日志文件
	fileName := path.Join(logFilePath, logFileName)
	if _, err := os.Stat(fileName); err != nil {
		if _, err := os.Create(fileName); err != nil {
			fmt.Println(err.Error())
		}
	}
	//写入文件
	src, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		fmt.Println("err", err)
	}

	//实例化
	logger := logrus.New()

	//设置输出
	logger.Out = src

	//设置日志级别
	logger.SetLevel(logrus.DebugLevel)

	//设置日志格式
	logger.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
	})
	return logger
}
