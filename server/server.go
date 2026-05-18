package server

import (
	"time"

	"github.com/Avi-0009/InstantWicket-backend/config"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func NewServer() *gin.Engine {
	r := gin.Default()

	// 1. Add CORS Middleware right here
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			config.GetEnv(
				"LOCAL_URL",
				"http://localhost:5173",
			),

			config.GetEnv(
				"FRONTEND_URL",
				"http://192.168.29.139:5173",
			),
		},
		AllowMethods:  []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:  []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders: []string{"Content-Length"},

		AllowCredentials: true,

		MaxAge: 12 * time.Hour,
	}))

	r.Use(gin.Recovery())

	// 2. Load your routes AFTER the middleware
	UserRoutes(r)
	PlayerStatsRoutes(r)
	MatchRoutes(r)
	return r
}
