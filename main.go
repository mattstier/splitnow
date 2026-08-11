package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"splitnow/db"
	"splitnow/internal/config"
	"splitnow/internal/handlers"
	"splitnow/internal/token"
)

func main() {
	cfg := config.Load()

	// connnecting to the splitnow db with root user (for now)
	err := db.Connect(cfg.DBURL)
	if err != nil {
		println("Failed to connect to db, reason: ", err)
	}

	// loads + parses the public PEM key only
	tm, err := token.New(cfg.JWTPublicKeyPath)
	if err != nil {
		log.Fatal(err)
	}

	store := db.Store{}

	redisClient := redis.NewClient(
		&redis.Options{Addr: cfg.RedisAddr})

	r := gin.Default()
	r.GET("/health", handlers.Health)

	// Endpoints requiring authorization
	r.GET("/messages",
		handlers.RequireAuth(tm, handlers.ExtractFromHeader),
		handlers.GetMessages(store))
	r.POST("/rooms",
		handlers.RequireAuth(tm, handlers.ExtractFromHeader),
		handlers.CreateRoom(store))
	r.GET("/rooms",
		handlers.RequireAuth(tm, handlers.ExtractFromHeader),
		handlers.GetRooms(store))
	r.GET("/rooms/mine",
		handlers.RequireAuth(tm, handlers.ExtractFromHeader),
		handlers.GetUserRooms(store))
	r.POST("/rooms/:room_id/join",
		handlers.RequireAuth(tm, handlers.ExtractFromHeader),
		handlers.AddMember(store))
	r.POST("/rooms/:room_id/leave",
		handlers.RequireAuth(tm, handlers.ExtractFromHeader),
		handlers.RemoveMember(store))

	r.GET("/ws",
		handlers.RequireAuth(tm, handlers.ExtractFromQuery),
		handlers.WS(store, redisClient))

	r.Run(":" + cfg.Port)
}
