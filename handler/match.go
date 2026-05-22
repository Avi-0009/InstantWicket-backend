package handler

import (
	"context"
	"net/http"

	"github.com/Avi-0009/InstantWicket-backend/database"
	"github.com/Avi-0009/InstantWicket-backend/database/dbHelper"
	"github.com/Avi-0009/InstantWicket-backend/models"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
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

	var finalMatchID string

	err := database.WithTransaction(context.Background(), func(tx *sqlx.Tx) error {

		teamAID, err := dbHelper.CreateTeam(tx, input.TeamAName, userID)
		if err != nil {
			return err
		}

		teamBID, err := dbHelper.CreateTeam(tx, input.TeamBName, userID)
		if err != nil {
			return err
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

		matchID, err := dbHelper.CreateMatch(tx, matchInput, userID)
		if err != nil {
			return err
		}
		finalMatchID = matchID
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create Team A"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Match and Teams created successfully",
		"match_id": finalMatchID,
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
