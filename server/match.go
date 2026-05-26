package server

import (
	"github.com/Avi-0009/InstantWicket-backend/handler"
	"github.com/Avi-0009/InstantWicket-backend/middleware"
	"github.com/gin-gonic/gin"
)

func MatchRoutes(r *gin.Engine) {
	match := r.Group("/v1/matches")
	{
		match.GET("", handler.GetMatches)
		match.GET("/:match_id", handler.GetMatchByID)
		match.GET("/:match_id/players", handler.GetMatchPlayersHandler)
	}
	protected := match.Group("", middleware.AuthMiddleware())
	{
		protected.POST("/match", handler.CreateMatch)
	}
}
