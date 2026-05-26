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
			Status  string `db:"status"`
			MatchID string `db:"match_id"`
		}

		err := tx.Get(&inningsData, `SELECT status, match_id FROM innings WHERE id = $1 FOR UPDATE`, input.InningsID)
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
			    striker_id = $4,
			    non_striker_id = $5,
			    bowler_id = $6,
			    last_updated = CURRENT_TIMESTAMP
			WHERE match_id = $7`,
			totalRunsOnBall, wicketIncrement, legalBallIncrement,
			input.StrikerID, input.NonStrikerID, input.BowlerID, inningsData.MatchID,
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
			WHERE match_id = $5 AND player_id = $6`,
			input.RunsFromBat, ballsFacedIncrement, fours, sixes, inningsData.MatchID, input.StrikerID,
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
			WHERE match_id = $6 AND player_id = $7`,
			runsConceded, legalBallIncrement, bowlerWickets, wides, noBalls, inningsData.MatchID, input.BowlerID,
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
				WHERE match_id = $4 AND player_id = $5`,
				catches, runouts, stumpings, inningsData.MatchID, *input.FielderID,
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
			lms.match_id, 
			lms.current_score, 
			lms.wickets, 
			lms.legal_balls,
			
			lms.striker_id,
			COALESCE(su.name, '') AS striker_name,
			COALESCE(spms.runs_scored, 0) AS striker_runs,
			COALESCE(spms.balls_played, 0) AS striker_balls,

			lms.non_striker_id,
			COALESCE(nsu.name, '') AS non_striker_name,

			lms.bowler_id,
			COALESCE(bu.name, '') AS bowler_name,
			COALESCE(bpms.runs_conceded, 0) AS bowler_runs,
			COALESCE(bpms.wickets_taken, 0) AS bowler_wickets

		FROM live_match_stats lms
		LEFT JOIN player_stats sps ON lms.striker_id = sps.id
		LEFT JOIN users su ON sps.user_id = su.id
		LEFT JOIN player_match_stats spms ON lms.striker_id = spms.player_id AND lms.match_id = spms.match_id

		LEFT JOIN player_stats nsps ON lms.non_striker_id = nsps.id
		LEFT JOIN users nsu ON nsps.user_id = nsu.id

		LEFT JOIN player_stats bps ON lms.bowler_id = bps.id
		LEFT JOIN users bu ON bps.user_id = bu.id
		LEFT JOIN player_match_stats bpms ON lms.bowler_id = bpms.player_id AND lms.match_id = bpms.match_id

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
		err := tx.Get(&inningsID, `
				INSERT INTO innings(
				                    match_id, innings_no, batting_team_id, bowling_team_id,
				                    striker_id, non_striker_id, bowler_id, target_runs, status
				) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'ongoing')
				RETURNING id`,
			req.MatchID, req.InningsNo, req.BattingTeamID, req.BowlingTeamID,
			req.StrikerID, req.NonStrikerID, req.BowlerID, req.TargetRuns,
		)
		if err != nil {
			return err
		}
		_, err = tx.Exec(`INSERT INTO live_match_stats(match_id, innings_id,
    striker_id, non_striker_id, bowler_id, current_over, legal_balls, current_score, wickets, required_runs)
	VALUES ($1, $2, $3, $4, $5, 0, 0, 0, 0, $6)
	ON CONFLICT (match_id) DO UPDATE SET
	                          innings_id = EXCLUDED.innings_id,
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
			req.MatchID, inningsID, req.StrikerID, req.NonStrikerID, req.BowlerID, req.TargetRuns)
		return err
	})
	return inningsID, err
}

func CompleteInnings(c context.Context, inningsID string) error {
	_, err := database.DB.ExecContext(c, `UPDATE innings
SET status = 'completed', updated_at = CURRENT_TIMESTAMP WHERE id = $1`, inningsID)
	return err
}
