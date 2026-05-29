package server

import (
	"github.com/Avi-0009/InstantWicket-backend/handler"
	"github.com/Avi-0009/InstantWicket-backend/middleware"
	"github.com/gin-gonic/gin"
)

func ScoringRoutes(r *gin.Engine) {
	scoring := r.Group("/v1/scoring")
	{
		scoring.GET("/live/:match_id", handler.GetLiveScoreboardHandler)
		scoring.GET("/scorecard/:match_id", handler.GetScoreCardsHandler)
	}

	protected := scoring.Group("", middleware.AuthMiddleware())
	{
		protected.POST("/start", handler.StartInningsHandler)
		protected.POST("/ball", handler.RecordBallHandler)
		protected.POST("/innings/:innings_id/complete", handler.CompleteInningsHandler)
	}
}
