package main

import (
	"log"
	"redpacket/config"
	"redpacket/database"
	"redpacket/handler"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	database.Init(cfg)

	r := gin.Default()

	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	api := r.Group("/api/v1")
	{
		activities := api.Group("/activities")
		{
			activities.POST("", handler.CreateActivity)
			activities.GET("", handler.ListActivities)
			activities.GET("/:id", handler.GetActivity)
			activities.GET("/:id/records", handler.GetRecords)
		}

		redpacket := api.Group("/redpacket")
		{
			redpacket.POST("/grab", handler.GrabRedPacket)
		}

		users := api.Group("/users")
		{
			users.GET("/:user_id/redpackets", handler.GetUserRedPackets)
		}
	}

	log.Printf("Server starting on port %s...", cfg.ServerPort)
	if err := r.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
