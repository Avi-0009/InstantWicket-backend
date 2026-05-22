package server

import (
	"github.com/Avi-0009/InstantWicket-backend/handler"
	"github.com/Avi-0009/InstantWicket-backend/middleware"
	"github.com/gin-gonic/gin"
)

func PlayerStatsRoutes(r *gin.Engine) {
	playerStats := r.Group("/v1/player_stats")
	{
		playerStats.GET("", handler.GetAllPlayerStats)
		playerStats.GET("/:player_id", handler.GetPlayerStats)
		playerStats.GET("/search", handler.SearchPlayerStats)
	}
	protected := playerStats.Group("", middleware.AuthMiddleware())
	{
		//protected.POST("", handler.CreatePlayerStats)
		protected.PUT("/:player_id", handler.UpdatePlayerStats)
		protected.POST("/guest", handler.AddGuest)
	}
	//playerStats := r.Group("/v1/player_stats", middleware.AuthMiddleware())
	//{
	//	playerStats.POST("", handler.CreatePlayerStats)
	//	playerStats.GET("", handler.GetPlayerStats)
	//	playerStats.PUT("", handler.UpdatePlayerStats)
	//}
}
