package main

import (
	"log"
	"os"
	"os/signal"
	"redpacket/config"
	"redpacket/database"
	"redpacket/handler"
	"redpacket/service"
	"syscall"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	database.Init(cfg)

	stopCh := make(chan struct{})
	service.StartAutoRefundScheduler(30, stopCh)

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
			activities.POST("/:id/refund", handler.TriggerActivityRefund)
		}

		redpacket := api.Group("/redpacket")
		{
			redpacket.POST("/grab", handler.GrabRedPacket)
		}

		refund := api.Group("/refunds")
		{
			refund.POST("/batch", handler.TriggerBatchRefund)
			refund.GET("", handler.ListRefundRecords)
		}

		users := api.Group("/users")
		{
			users.GET("/:user_id/redpackets", handler.GetUserRedPackets)
		}
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		log.Println("Shutting down gracefully...")
		close(stopCh)
		os.Exit(0)
	}()

	log.Printf("Server starting on port %s...", cfg.ServerPort)
	if err := r.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
