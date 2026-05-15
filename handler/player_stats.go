package handler

import (
	"fmt"
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	//fmt.Println("2")
	if isExist {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Player stats already exists"})
		return
	}

	err = dbHelper.CreatePlayerStats(userID, input.BattingStyle, input.BowlingStyle)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Player stats created successfully",
	})
}

func GetPlayerStats(c *gin.Context) {
	userID := c.GetString("userID")
	fmt.Println("1")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	fmt.Println("2")

	playerStats, err := dbHelper.GetPlayerStatsByUserID(userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Player stats not found"})
		return
	}
	fmt.Println("3")
	c.JSON(http.StatusOK, gin.H{"player_stats": playerStats})
}
