package server

import (
	"github.com/Avi-0009/InstantWicket-backend/handler"
	"github.com/Avi-0009/InstantWicket-backend/middleware"

	"github.com/gin-gonic/gin"
)

func UserRoutes(r *gin.Engine) {

	user := r.Group("/v1/auth")
	{
		user.POST("/register", handler.Register)
		user.POST("/login", handler.LoginUser)
		user.PUT("/reset-password", handler.ResetPassword)

		user.POST(
			"/logout",
			middleware.AuthMiddleware(),
			handler.LogoutUser,
		)
	}
}
