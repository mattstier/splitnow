package main

import (
	"github.com/gin-gonic/gin"

	"splitnow/db"
	"splitnow/internal/handlers"
)

func main() {
	// connnecting to the splitnow db with root user (for now)
	err := db.Connect("postgres://matestier@/splitnow?host=/var/run/postgresql")
	if err != nil {
		println("Failed to connect to db, reason: ", err)
	}

	r := gin.Default()
	r.GET("/health", handlers.Health)
	r.GET("/messages", handlers.GetMessages)
	r.POST("/rooms", handlers.CreateRoom)
	r.GET("/rooms", handlers.GetRooms)
	r.GET("/ws", handlers.WS)
	r.Run()
}
