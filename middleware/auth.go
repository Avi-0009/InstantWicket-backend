package middleware

import (
	"net/http"
	"strings"

	"github.com/Avi-0009/InstantWicket-backend/database"
	"github.com/Avi-0009/InstantWicket-backend/models"
	"github.com/Avi-0009/InstantWicket-backend/utils"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {

	return func(c *gin.Context) {

		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(
			authHeader,
			"Bearer ",
		)

		claims, err := utils.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		var session models.Session

		query := `SELECT id, user_id FROM user_sessions WHERE id = $1 AND archived_at IS NULL;`
		err = database.DB.Get(&session, query, claims.SessionID)

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Session expired or logged out"})
			c.Abort()
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("sessionID", claims.SessionID)

		c.Next()
	}
}
