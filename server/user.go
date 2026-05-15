package server

import (
	"github.com/Avi-0009/InstantWicket-backend/handler"
	"github.com/gin-gonic/gin"
)

func UserRoutes(r *gin.Engine) {

	user := r.Group("/v1")
	{
		user.POST("/register", handler.Register)
		user.POST("/login", handler.LoginUser)
	}
}
