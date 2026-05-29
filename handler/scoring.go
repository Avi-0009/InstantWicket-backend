package handler

import (
	"context"
	"net/http"

	"github.com/Avi-0009/InstantWicket-backend/database/dbHelper"
	"github.com/Avi-0009/InstantWicket-backend/models"
	"github.com/gin-gonic/gin"
)

// Handles recording a single ball

func RecordBallHandler(c *gin.Context) {
	// here we're checking for user and umpire
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var input models.RecordBallRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	err := dbHelper.RecordBall(context.Background(), input)
	if err != nil {
		if err.Error() == "cannot record ball: innings is no longer ongoing" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record ball: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Ball recorded successfully",
	})
}

// Handles retrieving the live scorecard

func GetLiveScoreboardHandler(c *gin.Context) {
	matchID := c.Param("match_id")
	if matchID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "match_id is required"})
		return
	}

	scoreboard, err := dbHelper.GetLiveScoreboard(matchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch scoreboard: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, scoreboard)
}

func StartInningsHandler(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req models.StartInningsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	inningsID, err := dbHelper.StartInnings(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start innings: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message":  "Innings started successfully",
		"inningID": inningsID,
	})
}

func CompleteInningsHandler(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	inningsID := c.Param("innings_id")
	if inningsID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "innings_id is required"})
		return
	}
	err := dbHelper.CompleteInnings(c.Request.Context(), inningsID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to complete innings: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Innings completed successfully"})
}

func GetScoreCardsHandler(c *gin.Context) {
	matchID := c.Param("match_id")
	if matchID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "match_id is required"})
		return
	}
	scorecard, err := dbHelper.GetMatchScorecard(matchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch scorecard: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"scorecard": scorecard})
}
