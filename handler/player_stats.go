package handler

import (
	"net/http"

	"github.com/Avi-0009/InstantWicket-backend/database/dbHelper"
	"github.com/Avi-0009/InstantWicket-backend/models"

	"github.com/gin-gonic/gin"
)

func CreatePlayerStats(c *gin.Context) {

	userID := c.GetString("userID")

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var input models.CreatePlayerStats

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	isExist, err := dbHelper.IsPlayerStatsExist(
		userID,
	)
	//fmt.Println("1")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	//fmt.Println("2")
	if isExist {
		playerID, err := dbHelper.GetPlayerIDByUserID(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"error":     "Player stats already exists",
			"player_id": playerID,
		})
		return
	}

	playerID, err := dbHelper.CreatePlayerStats(userID, input.BattingStyle, input.BowlingStyle)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Player stats created successfully",
		"playerID": playerID,
	})
}

func GetPlayerStats(c *gin.Context) {
	playerID := c.Param("player_id")
	//fmt.Println("1")
	if playerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Player ID is required"})
		return
	}
	//fmt.Println("2")

	playerStats, err := dbHelper.GetPlayerStatsByPlayerID(playerID)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Player stats not found"})
		return
	}
	//fmt.Println("3")
	c.JSON(http.StatusOK, gin.H{"player_stats": playerStats})
}

func UpdatePlayerStats(c *gin.Context) {
	UserID := c.GetString("userID")

	if UserID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	//fmt.Println("1")
	var input models.UpdatePlayerStats
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	//fmt.Println("2")
	err := dbHelper.UpdatePlayerStats(UserID, input.BattingStyle, input.BowlingStyle)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	//fmt.Println("3")
	c.JSON(http.StatusOK, gin.H{"message": "Player stats updated successfully"})
}
