package dbHelper

import (
	"context"
	"errors"

	"github.com/Avi-0009/InstantWicket-backend/database"
	"github.com/Avi-0009/InstantWicket-backend/models"
	"github.com/jmoiron/sqlx"
)

func RecordBall(c context.Context, input models.RecordBallRequest) error {
	return database.WithTransaction(c, func(tx *sqlx.Tx) error {
		var inningsData struct {
			Status           string `db:"status"`
			MatchID          string `db:"match_id"`
			BattingTeamID    string `db:"batting_team_id"`
			BowlingTeamID    string `db:"bowling_team_id"`
			TotalWickets     int    `db:"total_wickets"`
			AllowSoloBatting bool   `db:"allow_solo_batting"`
		}

		err := tx.Get(&inningsData, `SELECT 
    i.status, i.match_id, i.batting_team_id, i.bowling_team_id, i.total_wickets, m.allow_solo_batting 
	FROM innings i
	JOIN matches m ON m.id = i.match_id
	WHERE i.id = $1 FOR UPDATE`, input.InningsID)
		if inningsData.Status != "ongoing" {
			return errors.New("cannot record balls, inning is no longer ongoing")
		}

		var activePlayers int
		err = tx.Get(&activePlayers, `SELECT COUNT(*) FROM match_players
		WHERE match_id = $1 AND team_id = $2 AND is_retired = FALSE`,
			inningsData.MatchID, inningsData.BattingTeamID)

		if err != nil {
			return err
		}
		lastPlayerRemaining := inningsData.TotalWickets >= activePlayers-1
		if lastPlayerRemaining {
			if !inningsData.AllowSoloBatting {
				_, err = tx.Exec(`UPDATE innings SET status = 'completed', completed_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE id = $1`, input.InningsID)
				return err
			}
			if input.NonStrikerID != nil {
				return errors.New("non striker not allowed for last batsman")
			}
		} else {
			if input.NonStrikerID == nil {
				return errors.New("non striker is required")
			}
		}

		if err != nil {
			return err
		}
		if inningsData.Status != "ongoing" {
			return errors.New("cannot record ball: innings is no longer ongoing")
		}

		totalRunsOnBall := input.RunsFromBat + input.Extras
		_, err = tx.Exec(`
			INSERT INTO balls (
				innings_id, over_number, ball_number, is_legal_ball, 
				runs_from_bat, extras, extra_type, total_runs, is_wicket, wicket_type, 
				fielder_id, striker_id, non_striker_id, bowler_id, out_player_id
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
			)`,
			input.InningsID, input.OverNumber, input.BallNumber, input.IsLegalBall,
			input.RunsFromBat, input.Extras, input.ExtraType, totalRunsOnBall, input.IsWicket, input.WicketType,
			input.FielderID, input.StrikerID, input.NonStrikerID, input.BowlerID, input.OutPlayerID,
		)
		if err != nil {
			return err
		}

		// updating innings table here
		legalBallIncrement := 0
		if input.IsLegalBall {
			legalBallIncrement = 1
		}

		wicketIncrement := 0
		if input.IsWicket {
			wicketIncrement = 1
		}

		_, err = tx.Exec(`
			UPDATE innings 
			SET total_runs = total_runs + $1,
			    total_wickets = total_wickets + $2,
			    total_extras = total_extras + $3,
			    legal_balls = legal_balls + $4,
			    updated_at = CURRENT_TIMESTAMP
			WHERE id = $5`,
			totalRunsOnBall, wicketIncrement, input.Extras, legalBallIncrement, input.InningsID,
		)
		if err != nil {
			return err
		}

		// updating live match stats here
		_, err = tx.Exec(`
			UPDATE live_match_stats 
			SET current_score = current_score + $1,
			    wickets = wickets + $2,
			    legal_balls = legal_balls + $3,
			    current_over = (legal_balls + $3) / 6,
			    striker_id = $4,
			    non_striker_id = $5,
			    bowler_id = $6,
			    batting_team_id = $7,
			    bowling_team_id = $8,
			    last_updated = CURRENT_TIMESTAMP
			WHERE match_id = $9`,
			totalRunsOnBall, wicketIncrement, legalBallIncrement,
			input.StrikerID, input.NonStrikerID, input.BowlerID, inningsData.BattingTeamID, inningsData.BowlingTeamID, inningsData.MatchID,
		)
		if err != nil {
			return err
		}

		// updating batsmam's stats here
		ballsFacedIncrement := 1
		if input.ExtraType != nil && *input.ExtraType == "wide" {
			ballsFacedIncrement = 0 // wide don't count as legit ball
		}

		fours, sixes := 0, 0
		if input.RunsFromBat == 4 {
			fours = 1
		}
		if input.RunsFromBat == 6 {
			sixes = 1
		}

		_, err = tx.Exec(`
			UPDATE player_match_stats 
			SET runs_scored = runs_scored + $1,
			    balls_played = balls_played + $2,
			    fours = fours + $3,
			    sixes = sixes + $4,
			    updated_at = CURRENT_TIMESTAMP
			WHERE match_id = $5 AND team_id = $6 AND player_id = $7`,
			input.RunsFromBat, ballsFacedIncrement, fours, sixes, inningsData.MatchID, inningsData.BattingTeamID, input.StrikerID,
		)
		if err != nil {
			return err
		}

		// updating boller's stats here
		runsConceded := totalRunsOnBall
		if input.ExtraType != nil && (*input.ExtraType == "bye" || *input.ExtraType == "leg_bye") {
			runsConceded = 0
		}

		bowlerWickets := 0
		if input.IsWicket && input.WicketType != nil && *input.WicketType != "run_out" {
			bowlerWickets = 1 // got wicket here
		}

		wides, noBalls := 0, 0
		if input.ExtraType != nil && *input.ExtraType == "wide" {
			wides = 1
		}
		if input.ExtraType != nil && *input.ExtraType == "no_ball" {
			noBalls = 1
		}

		_, err = tx.Exec(`
			UPDATE player_match_stats 
			SET runs_conceded = runs_conceded + $1,
			    balls_bowled = balls_bowled + $2,
			    wickets_taken = wickets_taken + $3,
			    wides = wides + $4,
			    no_balls = no_balls + $5
			WHERE match_id = $6 AND team_id = $7 AND player_id = $8`,
			runsConceded, legalBallIncrement, bowlerWickets, wides, noBalls, inningsData.MatchID, inningsData.BowlingTeamID, input.BowlerID,
		)
		if err != nil {
			return err
		}

		// updaing fielders stats here
		if input.IsWicket && input.FielderID != nil {
			catches, runouts, stumpings := 0, 0, 0

			switch *input.WicketType {
			case "caught", "caught_and_bowled":
				catches = 1
			case "run_out":
				runouts = 1
			case "stumped":
				stumpings = 1
			}

			_, err = tx.Exec(`
				UPDATE player_match_stats 
				SET catches = catches + $1,
				    runouts = runouts + $2,
				    stumpings = stumpings + $3,
				    updated_at = CURRENT_TIMESTAMP
				WHERE match_id = $4 AND team_id = $5 AND player_id = $6`,
				catches, runouts, stumpings, inningsData.MatchID, inningsData.BowlingTeamID, *input.FielderID,
			)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

// fetching the live scoreboard data with Player names using joins here

func GetLiveScoreboard(matchID string) (*models.LiveScoreboardResponse, error) {
	var board models.LiveScoreboardResponse

	query := `
		SELECT 
			lms.match_id, lms.innings_id, lms.batting_team_id, lms.bowling_team_id, lms.current_score, lms.wickets, lms.legal_balls, lms.required_runs,
			lms.striker_id,
			COALESCE(su.name, '') AS striker_name, COALESCE(spms.runs_scored, 0) AS striker_runs, COALESCE(spms.balls_played, 0) AS striker_balls,
			lms.non_striker_id,
			COALESCE(nsu.name, '') AS non_striker_name, COALESCE(nspms.runs_scored, 0) AS non_striker_runs, COALESCE(nspms.balls_played, 0) AS non_striker_balls,
			lms.bowler_id,
			COALESCE(bu.name, '') AS bowler_name, COALESCE(bpms.runs_conceded, 0) AS bowler_runs, COALESCE(bpms.wickets_taken, 0) AS bowler_wickets
		FROM live_match_stats lms
		LEFT JOIN player_stats sps ON lms.striker_id = sps.id
		LEFT JOIN users su ON sps.user_id = su.id
		LEFT JOIN player_match_stats spms ON lms.striker_id = spms.player_id AND lms.match_id = spms.match_id AND lms.batting_team_id = spms.team_id
		LEFT JOIN player_stats nsps ON lms.non_striker_id = nsps.id
		LEFT JOIN users nsu ON nsps.user_id = nsu.id
		LEFT JOIN player_match_stats nspms ON lms.non_striker_id = nspms.player_id AND lms.match_id = nspms.match_id AND lms.batting_team_id = nspms.team_id
		LEFT JOIN player_stats bps ON lms.bowler_id = bps.id
		LEFT JOIN users bu ON bps.user_id = bu.id
		LEFT JOIN player_match_stats bpms ON lms.bowler_id = bpms.player_id AND lms.match_id = bpms.match_id AND lms.bowling_team_id = bpms.team_id
		WHERE lms.match_id = $1
	`

	err := database.DB.Get(&board, query, matchID)
	if err != nil {
		return nil, err
	}

	return &board, nil
}

func StartInnings(c context.Context, req models.StartInningsRequest) (string, error) {
	var inningsID string
	err := database.WithTransaction(c, func(tx *sqlx.Tx) error {
		var allowSoloBatting bool

		err := tx.Get(&allowSoloBatting, `SELECT allow_solo_batting FROM matches WHERE id = $1`, req.MatchID)
		if err != nil {
			return err
		}
		if !allowSoloBatting && req.NonStrikerID == nil {
			return errors.New("non_striker_id is required")
		}
		err = tx.Get(&inningsID, `
				INSERT INTO innings(
				                    match_id, innings_no, batting_team_id, bowling_team_id, striker_id, non_striker_id, bowler_id, target_runs, status
				) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'ongoing')
				RETURNING id`,
			req.MatchID, req.InningsNo, req.BattingTeamID, req.BowlingTeamID, req.StrikerID, req.NonStrikerID, req.BowlerID, req.TargetRuns,
		)
		if err != nil {
			return err
		}
		_, err = tx.Exec(`INSERT INTO live_match_stats(match_id, innings_id, batting_team_id, bowling_team_id,
    striker_id, non_striker_id, bowler_id, current_over, legal_balls, current_score, wickets, required_runs)
	VALUES ($1, $2, $3, $4, $5, $6, $7, 0, 0, 0, 0, $8)
	ON CONFLICT (match_id) DO UPDATE SET
	                          innings_id = EXCLUDED.innings_id,
	batting_team_id = EXCLUDED.batting_team_id,
	bowling_team_id = EXCLUDED.bowling_team_id,
	                          striker_id = EXCLUDED.striker_id,
	    					  non_striker_id = EXCLUDED.non_striker_id,
	                          bowler_id = EXCLUDED.bowler_id,
	                          current_over = 0,
	                          legal_balls = 0,
	                          current_score = 0,
	                          wickets = 0,
	                          required_runs = EXCLUDED.required_runs,
	                          last_updated = CURRENT_TIMESTAMP
			`,
			req.MatchID, inningsID, req.BattingTeamID, req.BowlingTeamID, req.StrikerID, req.NonStrikerID, req.BowlerID, req.TargetRuns)
		return err
	})
	return inningsID, err
}

func CompleteInnings(c context.Context, inningsID string) error {
	_, err := database.DB.ExecContext(c, `UPDATE innings
SET status = 'completed', updated_at = CURRENT_TIMESTAMP WHERE id = $1`, inningsID)
	return err
}

func GetMatchScorecard(matchID string) ([]models.PlayerScorecard, error) {
	var scorecard []models.PlayerScorecard
	query := `
		SELECT 
			pms.team_id, pms.player_id, COALESCE(u.name, 'Unknown') AS player_name,
			COALESCE(pms.runs_scored, 0) AS runs_scored,
			COALESCE(pms.balls_played, 0) AS balls_played,
			0 AS fours, 0 AS sixes, false AS is_out, 
			COALESCE(pms.runs_conceded, 0) AS runs_conceded,
			COALESCE(pms.wickets_taken, 0) AS wickets_taken,
			0.0 AS overs_bowled, 0 AS maidens
		FROM player_match_stats pms
		JOIN player_stats ps ON pms.player_id = ps.id
		JOIN users u ON ps.user_id = u.id
		WHERE pms.match_id = $1
	`
	err := database.DB.Select(&scorecard, query, matchID)
	if scorecard == nil {
		scorecard = []models.PlayerScorecard{}
	}
	return scorecard, err
}

func CompleteMatch(matchID string) error {
	query := `UPDATE matches SET status = 'completed', updated_at = CURRENT_TIMESTAMP WHERE id = $1`
	_, err := database.DB.Exec(query, matchID)
	return err
}
