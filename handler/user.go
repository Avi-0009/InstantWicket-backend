package handler

import (
	"net/http"

	"github.com/Avi-0009/InstantWicket-backend/database/dbHelper"
	"github.com/Avi-0009/InstantWicket-backend/models"
	"github.com/Avi-0009/InstantWicket-backend/utils"
	"github.com/gin-gonic/gin"
)

func Register(c *gin.Context) {

	var input models.RegisterUser

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	isUserExist, err := dbHelper.IsUserExist(
		input.PhoneNo,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if isUserExist {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User already exists"})
		return
	}

	hashedPassword, err := utils.HashPassword(
		input.Password,
	)

	if err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	err = dbHelper.CreateUser(input.Name, input.PhoneNo, hashedPassword)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully"})
}

func LoginUser(c *gin.Context) {

	var input models.LoginUser

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	user, err := dbHelper.GetUserByPhoneNo(
		input.PhoneNo,
	)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid phone number or password"})
		return
	}

	isPasswordCorrect := utils.CheckPassword(user.Password, input.Password)

	if !isPasswordCorrect {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid phone number or password"})
		return
	}

	sessionID, err := dbHelper.CreateUserSession(user.ID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
		return
	}

	token, err := utils.GenerateToken(user.ID, sessionID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"token":   token,
		"user": gin.H{
			"id":       user.ID,
			"name":     user.Name,
			"phone_no": user.PhoneNo,
		},
	})
}

func LogoutUser(c *gin.Context) {

	sessionID := c.GetString("sessionID")

	if sessionID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}

	err := dbHelper.DeleteUserSession(
		sessionID,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to logout"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Logout successful"})
}

func ResetPassword(c *gin.Context) {
	var input models.ResetPassword
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	err := dbHelper.UpdatePasswordByPhone(input.PhoneNo, input.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
	}
	c.JSON(http.StatusOK, gin.H{"message": "Password reset successful"})
}
