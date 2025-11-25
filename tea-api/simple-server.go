//go:build ignore
// +build ignore

// 该文件为简单测试服务器示例，默认不参与构建，避免与正式 main 冲突。
package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	fmt.Println("启动简单测试服务器...")

	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
			"status":  "success",
		})
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "茶心阁API服务运行正常",
			"status":  "healthy",
		})
	})

	fmt.Println("服务器启动在 :8081")
	log.Fatal(r.Run(":8081"))
}
