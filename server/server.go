package server

import (
	"time"

	"github.com/Avi-0009/InstantWicket-backend/config"
	"github.com/Avi-0009/InstantWicket-backend/handler"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func NewServer() *gin.Engine {
	r := gin.Default()

	// Add CORS Middleware right here
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			config.GetEnv(
				"LOCAL_URL",
				"http://localhost:5173",
			),

			config.GetEnv(
				"FRONTEND_URL",
				"http://192.168.29.220:5173",
			),
		},
		AllowMethods:  []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:  []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders: []string{"Content-Length"},

		AllowCredentials: true,

		MaxAge: 7 * 12 * time.Hour,
	}))

	r.Use(gin.Recovery())

	r.GET("/api/ping", handler.PingHandler)

	// Load your routes AFTER the middleware
	UserRoutes(r)
	PlayerStatsRoutes(r)
	MatchRoutes(r)
	TeamRoutes(r)
	ScoringRoutes(r)
	return r
}
