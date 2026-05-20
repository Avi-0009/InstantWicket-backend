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
	var input models.CreateMatch
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	matchID, err := dbHelper.CreateMatch(input, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"message":  "Match created successfully",
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
