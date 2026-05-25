package handler

import (
	"fmt"
	"net/http"

	"github.com/Avi-0009/InstantWicket-backend/database/dbHelper"
	"github.com/Avi-0009/InstantWicket-backend/models"

	"github.com/gin-gonic/gin"
)

//func CreatePlayerStats(c *gin.Context) {
//
//	userID := c.GetString("userID")
//
//	if userID == "" {
//		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
//		return
//	}
//
//	var input models.CreatePlayerStats
//
//	if err := c.ShouldBindJSON(&input); err != nil {
//		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
//		return
//	}
//
//	isExist, err := dbHelper.IsPlayerStatsExist(
//		userID,
//	)
//	//fmt.Println("1")
//	if err != nil {
//		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
//		return
//	}
//	//fmt.Println("2")
//	if isExist {
//		playerID, err := dbHelper.GetPlayerIDByUserID(userID)
//		if err != nil {
//			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
//			return
//		}
//
//		c.JSON(http.StatusOK, gin.H{
//			"error":     "Player stats already exists",
//			"player_id": playerID,
//		})
//		return
//	}
//
//	playerID, err := dbHelper.CreatePlayerStats(userID, input.BattingStyle, input.BowlingStyle)
//
//	if err != nil {
//		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
//		return
//	}
//
//	c.JSON(http.StatusCreated, gin.H{
//		"message":  "Player stats created successfully",
//		"playerID": playerID,
//	})
//}

func AddGuest(c *gin.Context) {
	var input models.AddGuestPlayer
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	playerID, err := dbHelper.AddGuest(input.Name, input.PhoneNo)
	if err != nil {
		fmt.Println("error in addGuest handler", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add guest"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"message":   "Player stats added successfully",
		"player_id": playerID,
		"name":      input.Name,
		"phone_no":  input.PhoneNo,
	})
}

func GetPlayerStats(c *gin.Context) {
	playerID := c.Param("player_id")
	//fmt.Println("x")
	if playerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Player ID is required"})
		return
	}
	//fmt.Println("y")

	playerStats, err := dbHelper.GetPlayerStatsByPlayerID(playerID)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Player stats not found"})
		return
	}
	fmt.Println("z")
	c.JSON(http.StatusOK, gin.H{"player_stats": playerStats})
}

func GetAllPlayerStats(c *gin.Context) {
	playerStats, err := dbHelper.GetAllPlayersStats()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch player stats"})
		return
	}
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

func SearchPlayerStats(c *gin.Context) {

	query := c.Query("q")

	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "query is required",
		})
		return
	}

	players, err := dbHelper.SearchPlayerStats(query)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"players": players})
}
