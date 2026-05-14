package server

import (
	"github.com/gin-gonic/gin"
)

func NewServer() *gin.Engine {
	r := gin.Default()
	r.Use(gin.Recovery())
	UserRoutes(r)
	return r
}
