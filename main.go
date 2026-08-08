package main

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"splitnow/db"
	"splitnow/internal/config"
	"splitnow/internal/handlers"
)

func main() {
	cfg := config.Load()

	// connnecting to the splitnow db with root user (for now)
	err := db.Connect(cfg.DBURL)
	if err != nil {
		println("Failed to connect to db, reason: ", err)
	}

	store := db.Store{}

	redisClient := redis.NewClient(
		&redis.Options{Addr: cfg.RedisAddr})

	r := gin.Default()
	r.GET("/health", handlers.Health)
	r.GET("/messages", handlers.GetMessages(store))
	r.POST("/rooms", handlers.CreateRoom(store))
	r.GET("/rooms", handlers.GetRooms(store))
	r.GET("/ws", handlers.WS(store, redisClient))
	r.Run(":" + cfg.Port)
}
