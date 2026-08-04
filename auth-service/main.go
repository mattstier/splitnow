package main

import (
	"github.com/gin-gonic/gin"
)

func Health(ctx *gin.Context) {
	ctx.JSON(200, gin.H{"status": "ok"})
}

func main() {
	router := gin.Default()
	router.GET("/health", Health) 
	router.Run(":8081")
}
