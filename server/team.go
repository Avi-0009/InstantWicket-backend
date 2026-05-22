package server

import (
	"github.com/Avi-0009/InstantWicket-backend/handler"
	"github.com/Avi-0009/InstantWicket-backend/middleware"

	"github.com/gin-gonic/gin"
)

func TeamRoutes(r *gin.Engine) {

	team := r.Group("/v1/teams")
	{
		team.GET("", handler.GetTeams)
		team.GET("/:team_id", handler.GetTeam)
	}
	protected := team.Group(
		"",
		middleware.AuthMiddleware(),
	)
	{
		//protected.POST("", handler.CreateTeam)
		protected.PUT("/:team_id", handler.UpdateTeam)
	}
}

//POST   /v1/teams
//GET    /v1/teams
//GET    /v1/teams/:team_id
//PUT    /v1/teams/:team_id
//DELETE /v1/teams/:team_id
