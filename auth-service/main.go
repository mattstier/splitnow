package main

import (
	"github.com/gin-gonic/gin"

	"splitnow/auth-service/db"
	"splitnow/auth-service/internal/handlers"
)

func Health(ctx *gin.Context) {
	ctx.JSON(200, gin.H{"status": "ok"})
}

func main() {
	// connnecting to the splitnow_auth db with root user (for now)
	err := db.Connect("postgres://matestier@/splitnow_auth?host=/var/run/postgresql")
	if err != nil {
		println("Failed to connect to db, reason: ", err)
	}

	store := db.Store{} 

	router := gin.Default()
	router.GET("/health", Health)
	router.POST("/users", handlers.CreateUser(store))
	router.Run(":8081")
}
