package server

import (
	"github.com/Avi-0009/InstantWicket-backend/handler"
	"github.com/Avi-0009/InstantWicket-backend/middleware"
	"github.com/gin-gonic/gin"
)

func MatchRoutes(r *gin.Engine) {
	match := r.Group("/v1/match")
	{
		match.GET("", handler.GetMatches)
	}
	protected := match.Group("", middleware.AuthMiddleware())
	{
		protected.POST("", handler.CreateMatch)
	}
}
