package handler

import (
	"net/http"

	"github.com/Avi-0009/InstantWicket-backend/database/dbHelper"
	"github.com/Avi-0009/InstantWicket-backend/models"
	"github.com/gin-gonic/gin"
)

//func CreateTeam(c *gin.Context) {
//	userID := c.GetString("userID")
//
//	if userID == "" {
//		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
//		return
//	}
//
//	var input models.CreateTeam
//
//	if err := c.ShouldBindJSON(&input); err != nil {
//		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
//		return
//	}
//	teamID, err := dbHelper.CreateTeam(input.Name, userID)
//	if err != nil {
//		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
//		return
//	}
//	c.JSON(http.StatusCreated, gin.H{
//		"message": "Team created successfully",
//		"teamID":  teamID,
//	})
//}

func GetTeams(c *gin.Context) {
	teams, err := dbHelper.GetTeams()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"teams": teams,
	})
}

func GetTeam(c *gin.Context) {
	teamID := c.Param("team_id")
	team, err := dbHelper.GetTeam(teamID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "team not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"team": team,
	})
}

func UpdateTeam(c *gin.Context) {
	teamID := c.Param("team_id")
	var input models.UpdateTeam
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	err := dbHelper.UpdateTeam(teamID, input.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Team updated successfully"})
}

func GetTeamPlayersHandler(c *gin.Context) {
	teamID := c.Param("team_id")
	if teamID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "team_id is required"})
		return
	}
	players, err := dbHelper.GetTeamPlayers(teamID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"players": players,
	})
}
