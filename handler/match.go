package handler

import (
	"net/http"

	"github.com/Avi-0009/InstantWicket-backend/database/dbHelper"
	"github.com/Avi-0009/InstantWicket-backend/models"
	"github.com/gin-gonic/gin"
)

func CreateMatch(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var input models.StartLiveMatchRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	teamAID, err := dbHelper.CreateTeam(input.TeamAName, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create Team A"})
		return
	}

	teamBID, err := dbHelper.CreateTeam(input.TeamBName, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create Team B"})
		return
	}

	tossWinnerTeamID := teamAID
	if input.TossWinner == "B" {
		tossWinnerTeamID = teamBID
	}

	matchInput := models.CreateMatch{
		TeamAID:           teamAID,
		TeamBID:           teamBID,
		TossWinnerTeamID:  tossWinnerTeamID,
		TossDecision:      input.TossDecision,
		AllowCommonPlayer: input.AllowCommonPlayer,
		AllowSoloBatting:  input.AllowSoloBatting,
		OversLimit:        input.OversLimit,
		UmpireID:          input.UmpireID,
	}

	matchID, err := dbHelper.CreateMatch(matchInput, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Match and Teams created successfully",
		"match_id": matchID,
	})
}

func GetMatches(c *gin.Context) {
	matches, err := dbHelper.GetMatches()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"matches": matches})
}

func GetMatchByID(c *gin.Context) {
	matchID := c.Param("match_id")
	if matchID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "match_id is required"})
		return
	}
	match, err := dbHelper.GetMatchByID(matchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "match not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"match": match})
}
