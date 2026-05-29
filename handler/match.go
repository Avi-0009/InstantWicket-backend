package handler

import (
	"context"
	"net/http"
	"strconv"

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

		matchID, err := dbHelper.CreateMatch(tx, matchInput, input.TeamAPlayers, input.TeamBPlayers, userID)
		if err != nil {
			return err
		}
		finalMatchID = matchID
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Fetch the fully created match to return to the frontend
	match, err := dbHelper.GetMatchByID(finalMatchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Match created but failed to fetch details"})
		return
	}

	// Fetch all players assigned to this match
	players, _ := dbHelper.GetMatchPlayers(finalMatchID)

	// Initialize as empty arrays so they serialize to [] in JSON instead of null
	match.TeamAPlayers = []models.MatchPlayer{}
	match.TeamBPlayers = []models.MatchPlayer{}

	// Sort the players into their teams
	for _, p := range players {
		if p.TeamID == match.TeamAID {
			match.TeamAPlayers = append(match.TeamAPlayers, p)
		} else if p.TeamID == match.TeamBID {
			match.TeamBPlayers = append(match.TeamBPlayers, p)
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Match and Teams created successfully",
		"match":   match,
	})
}

func GetMatches(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")

	page, _ := strconv.Atoi(pageStr)
	limit, _ := strconv.Atoi(limitStr)

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 10
	}
	offset := (page - 1) * limit

	matches, err := dbHelper.GetMatches(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if matches == nil {
		matches = []models.Match{}
	}

	c.JSON(http.StatusOK, gin.H{
		"matches": matches,
		"page":    page,
		"limit":   limit,
	})
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
	players, _ := dbHelper.GetMatchPlayers(matchID)
	match.TeamAPlayers = make([]models.MatchPlayer, 0)
	match.TeamBPlayers = make([]models.MatchPlayer, 0)

	for _, p := range players {
		if p.TeamID == match.TeamAID {
			match.TeamAPlayers = append(match.TeamAPlayers, p)
		} else if p.TeamID == match.TeamBID {
			match.TeamBPlayers = append(match.TeamBPlayers, p)
		}
	}
	c.JSON(http.StatusOK, gin.H{"match": match})
}

func GetMatchPlayersHandler(c *gin.Context) {
	matchID := c.Param("match_id")
	if matchID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "match_id is required"})
		return
	}
	players, err := dbHelper.GetMatchPlayers(matchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch match player not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"players": players})

}
