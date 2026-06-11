package dbHelper

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

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

		var currentMatchStats struct {
			LegalBalls       int `db:"legal_balls"`
			PartnershipRuns  int `db:"partnership_runs"`
			PartnershipBalls int `db:"partnership_balls"`
		}
		err = tx.Get(&currentMatchStats, `SELECT legal_balls, partnership_runs, partnership_balls FROM live_match_stats WHERE match_id = $1`, inningsData.MatchID)
		if err != nil {
			return err
		}

		totalRunsOnBall := input.RunsFromBat + input.Extras
		// Calculate exact partnership state at the end of this ball
		ballPartRuns := currentMatchStats.PartnershipRuns + totalRunsOnBall
		ballPartBalls := currentMatchStats.PartnershipBalls
		if input.IsLegalBall {
			ballPartBalls++
		}
		if input.IsWicket {
			ballPartRuns = 0
			ballPartBalls = 0
		}

		_, err = tx.Exec(`
			INSERT INTO balls (
				innings_id, over_number, ball_number, is_legal_ball, 
				runs_from_bat, extras, extra_type, total_runs, is_wicket, wicket_type, 
				fielder_id, striker_id, non_striker_id, bowler_id, out_player_id,
				partnership_runs, partnership_balls
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17
			)`,
			input.InningsID, input.OverNumber, input.BallNumber, input.IsLegalBall,
			input.RunsFromBat, input.Extras, input.ExtraType, totalRunsOnBall, input.IsWicket, input.WicketType,
			input.FielderID, input.StrikerID, input.NonStrikerID, input.BowlerID, input.OutPlayerID,
			ballPartRuns, ballPartBalls, // <-- Saves the history so Undo works perfectly!
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

		// strike rotation code here
		var currentLegalBalls int
		err = tx.Get(&currentLegalBalls, `SELECT legal_balls FROM live_match_stats WHERE match_id = $1`, inningsData.MatchID)
		if err != nil {
			return err
		}

		// calc physical runs (Runs from bat + Byes + Leg Byes)
		physicalRuns := input.RunsFromBat
		if input.ExtraType != nil {
			if *input.ExtraType == "bye" || *input.ExtraType == "leg_bye" {
				physicalRuns = input.Extras
			} else if *input.ExtraType == "wide" || *input.ExtraType == "no_ball" {
				// If they ran on a Wide/No Ball, subtract the 1 penalty run to find actual crossing runs
				physicalRuns = input.Extras - 1
			}
		}
		swappedForRuns := physicalRuns%2 != 0

		// check if the over is completed
		isOverComplete := false
		if input.IsLegalBall {
			isOverComplete = (currentMatchStats.LegalBalls+1)%6 == 0
		}

		// check for the next striker and non-striker
		// we convert StrikerID to a pointer by using the '&' address operator
		nextStrikerID := &input.StrikerID
		nextNonStrikerID := input.NonStrikerID

		// swap if they ran odd runs OR the over ended, but NOT both(XOR logic)
		// only rotate the strike if there is actually a non-striker! (Ignores Solo Batting)
		if nextNonStrikerID != nil {
			if swappedForRuns != isOverComplete {
				temp := nextStrikerID
				nextStrikerID = nextNonStrikerID
				nextNonStrikerID = temp
			}
		}

		// set the out player to nil so the frontend knows they are gone
		if input.IsWicket && input.OutPlayerID != nil {
			if nextStrikerID != nil && *nextStrikerID == *input.OutPlayerID {
				nextStrikerID = nil
			}
			if nextNonStrikerID != nil && *nextNonStrikerID == *input.OutPlayerID {
				nextNonStrikerID = nil
			}
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
			    partnership_runs = CASE WHEN $2 > 0 THEN 0 ELSE partnership_runs + $1 END,
			    partnership_balls = CASE WHEN $2 > 0 THEN 0 ELSE partnership_balls + $3 END,
			    last_updated = CURRENT_TIMESTAMP
			WHERE match_id = $9`,
			totalRunsOnBall, wicketIncrement, legalBallIncrement,
			nextStrikerID, nextNonStrikerID, input.BowlerID, inningsData.BattingTeamID, inningsData.BowlingTeamID, inningsData.MatchID,
		)
		if err != nil {
			return err
		}

		// updating batsmam's stats here
		ballsFacedIncrement := 0
		if input.IsLegalBall {
			ballsFacedIncrement = 1 // Only count as ball faced if it's a legal delivery (Standard, Bye, Leg Bye)
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

		if input.IsWicket && input.OutPlayerID != nil {
			_, err = tx.Exec(`
		UPDATE player_match_stats 
		SET is_out = true, updated_at = CURRENT_TIMESTAMP
		WHERE match_id = $1 AND player_id = $2 AND team_id = $3`,
				inningsData.MatchID, *input.OutPlayerID, inningsData.BattingTeamID,
			)
			if err != nil {
				return err
			}
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
			lms.match_id, lms.innings_id, lms.batting_team_id, lms.bowling_team_id, lms.current_score, lms.wickets, lms.legal_balls, lms.required_runs, lms.partnership_runs, lms.partnership_balls,
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
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// fetch the last 15 balls
	if board.InningsID != "" {
		var rawBalls []struct {
			RunsFromBat int     `db:"runs_from_bat"`
			Extras      int     `db:"extras"`
			ExtraType   *string `db:"extra_type"`
			IsWicket    bool    `db:"is_wicket"`
		}

		// Query the last 15 balls for this innings
		ballQuery := `
			SELECT runs_from_bat, extras, extra_type, is_wicket
			FROM balls
			WHERE innings_id = $1
			ORDER BY over_number DESC, ball_number DESC
			LIMIT 15
		`
		database.DB.Select(&rawBalls, ballQuery, board.InningsID)

		// Reverse the loop so the oldest ball is on the left and newest is on the right
		board.RecentBalls = make([]string, 0, len(rawBalls))
		for i := len(rawBalls) - 1; i >= 0; i-- {
			b := rawBalls[i]
			outcome := ""
			if b.IsWicket && b.ExtraType != nil && *b.ExtraType == "no_ball" {
				outcome = fmt.Sprintf("%dnb W", b.RunsFromBat+b.Extras)
			} else if b.IsWicket && b.ExtraType != nil && *b.ExtraType == "wide" {
				outcome = fmt.Sprintf("%dwd W", b.Extras)
			} else if b.IsWicket {
				outcome = "W"
			} else if b.ExtraType != nil && *b.ExtraType == "wide" {
				outcome = fmt.Sprintf("%dwd", b.Extras)
			} else if b.ExtraType != nil && *b.ExtraType == "no_ball" {
				outcome = fmt.Sprintf("%dnb", b.RunsFromBat+b.Extras)
			} else if b.ExtraType != nil && *b.ExtraType == "bye" {
				outcome = fmt.Sprintf("%db", b.RunsFromBat+b.Extras)
			} else if b.ExtraType != nil && *b.ExtraType == "leg_bye" {
				outcome = fmt.Sprintf("%dlb", b.RunsFromBat+b.Extras)
			} else {
				outcome = fmt.Sprintf("%d", b.RunsFromBat)
			}

			board.RecentBalls = append(board.RecentBalls, outcome)
		}
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
		//if !allowSoloBatting && req.NonStrikerID == nil {
		//	return errors.New("non_striker_id is required")
		//}
		// safely convert Go pointers to untyped nil for the SQL driver
		safeString := func(s *string) interface{} {
			if s == nil {
				return nil
			}
			return *s
		}
		safeInt := func(i *int) interface{} {
			if i == nil {
				return nil
			}
			return *i
		}

		err = tx.Get(&inningsID, `
				INSERT INTO innings(
				                    match_id, innings_no, batting_team_id, bowling_team_id, striker_id, non_striker_id, bowler_id, target_runs, status
				) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'ongoing')
				RETURNING id`,
			req.MatchID, req.InningsNo, req.BattingTeamID, req.BowlingTeamID, safeString(req.StrikerID), safeString(req.NonStrikerID), safeString(req.BowlerID), safeInt(req.TargetRuns),
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
			req.MatchID, inningsID, req.BattingTeamID, req.BowlingTeamID, safeString(req.StrikerID), safeString(req.NonStrikerID), safeString(req.BowlerID), safeInt(req.TargetRuns))
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
			COALESCE(pms.fours, 0) AS fours, 
			COALESCE(pms.sixes, 0) AS sixes, 
			pms.is_out,
			CASE 
				WHEN pms.is_out = true THEN 'Out'
				WHEN (lms.striker_id = pms.player_id OR lms.non_striker_id = pms.player_id) AND lms.batting_team_id = pms.team_id THEN 
					CASE WHEN m.status = 'completed' THEN 'Not out' ELSE 'Batting' END
				WHEN pms.balls_played > 0 OR pms.runs_scored > 0 THEN 'Not out'
				WHEN m.status = 'completed' THEN 'Did not bat'
				ELSE 'Yet to bat'
			END AS batting_status,
			COALESCE(pms.runs_conceded, 0) AS runs_conceded,
			COALESCE(pms.wickets_taken, 0) AS wickets_taken,
			COALESCE(pms.balls_bowled, 0) AS balls_bowled,
			COALESCE(mc.calculated_maidens, 0) AS maidens,
			COALESCE(pms.wides, 0) AS wides,
			COALESCE(pms.no_balls, 0) AS no_balls,			
			COALESCE(pms.catches, 0) AS catches,
			COALESCE(pms.runouts, 0) AS runouts,
			COALESCE(pms.stumpings, 0) AS stumpings
		FROM player_match_stats pms
		JOIN player_stats ps ON pms.player_id = ps.id
		JOIN users u ON ps.user_id = u.id
		JOIN matches m ON m.id = pms.match_id
		LEFT JOIN live_match_stats lms ON lms.match_id = pms.match_id 
		LEFT JOIN (
			SELECT bowler_id, COUNT(*) as calculated_maidens
			FROM (
				SELECT b.bowler_id, b.innings_id, b.over_number
				FROM balls b
				JOIN innings i ON i.id = b.innings_id
				WHERE i.match_id = $1
				GROUP BY b.bowler_id, b.innings_id, b.over_number
				HAVING SUM(b.total_runs) = 0 AND SUM(CASE WHEN b.is_legal_ball THEN 1 ELSE 0 END) = 6
			) AS valid_overs
			GROUP BY bowler_id
		) mc ON mc.bowler_id = pms.player_id
		WHERE pms.match_id = $1
	`

	err := database.DB.Select(&scorecard, query, matchID)
	if scorecard == nil {
		scorecard = []models.PlayerScorecard{}
	}
	return scorecard, err
}

func CompleteMatch(c context.Context, matchID string) error {
	return database.WithTransaction(c, func(tx *sqlx.Tx) error {
		// set match to completed
		_, err := tx.Exec(`UPDATE matches SET status = 'completed', updated_at = CURRENT_TIMESTAMP WHERE id = $1`, matchID)
		if err != nil {
			return err
		}
		// Sync maidens before updating career stats
		_, err = tx.Exec(`
			WITH MaidenCounts AS (
				SELECT bowler_id, COUNT(*) as maidens
				FROM (
					SELECT b.bowler_id, b.innings_id, b.over_number, 
						   SUM(b.total_runs) as over_runs, 
						   SUM(CASE WHEN b.is_legal_ball THEN 1 ELSE 0 END) as legal_balls
					FROM balls b
					JOIN innings i ON i.id = b.innings_id
					WHERE i.match_id = $1
					GROUP BY b.bowler_id, b.innings_id, b.over_number
				) over_stats
				WHERE over_runs = 0 AND legal_balls = 6
				GROUP BY bowler_id
			)
			UPDATE player_match_stats pms
			SET maiden_overs = COALESCE(mc.maidens, 0)
			FROM MaidenCounts mc
			WHERE pms.player_id = mc.bowler_id AND pms.match_id = $1
		`, matchID)
		if err != nil {
			return err
		}
		// feed all matchstats data to careerstats
		_, err = tx.Exec(`
			UPDATE player_stats ps
			SET 
				career_matches = ps.career_matches + 1,				
				-- bat
				career_innings_batted = ps.career_innings_batted + CASE WHEN pms.balls_played > 0 OR pms.runs_scored > 0 THEN 1 ELSE 0 END,
				career_runs = ps.career_runs + pms.runs_scored,
				career_balls_faced = ps.career_balls_faced + pms.balls_played,
				career_fours = ps.career_fours + pms.fours,
				career_sixes = ps.career_sixes + pms.sixes,				
				-- scoreMilestone
				career_thirties = ps.career_thirties + CASE WHEN pms.runs_scored >= 30 AND pms.runs_scored < 50 THEN 1 ELSE 0 END,
				career_fifties = ps.career_fifties + CASE WHEN pms.runs_scored >= 50 AND pms.runs_scored < 100 THEN 1 ELSE 0 END,
				career_hundreds = ps.career_hundreds + CASE WHEN pms.runs_scored >= 100 THEN 1 ELSE 0 END,				
				-- ducks & notOut
				career_not_outs = ps.career_not_outs + CASE WHEN pms.is_out = false AND (pms.balls_played > 0 OR pms.runs_scored > 0) THEN 1 ELSE 0 END,
				career_ducks = ps.career_ducks + CASE WHEN pms.runs_scored = 0 AND pms.is_out = true AND pms.balls_played > 0 THEN 1 ELSE 0 END,
				career_golden_ducks = ps.career_golden_ducks + CASE WHEN pms.runs_scored = 0 AND pms.is_out = true AND pms.balls_played = 1 THEN 1 ELSE 0 END,				
				-- highscore
				career_highest_score = GREATEST(ps.career_highest_score, pms.runs_scored),
				-- ball
				career_innings_bowled = ps.career_innings_bowled + CASE WHEN pms.balls_bowled > 0 THEN 1 ELSE 0 END,
				career_wickets = ps.career_wickets + pms.wickets_taken,
				career_runs_conceded = ps.career_runs_conceded + pms.runs_conceded,
				career_balls_bowled = ps.career_balls_bowled + pms.balls_bowled,
				career_maiden_overs = ps.career_maiden_overs + pms.maiden_overs,
				career_wides = ps.career_wides + pms.wides,
				career_no_balls = ps.career_no_balls + pms.no_balls,
				career_best_bowling_runs = CASE 
					WHEN pms.wickets_taken > ps.career_best_bowling_wickets THEN pms.runs_conceded
					WHEN pms.wickets_taken = ps.career_best_bowling_wickets THEN LEAST(ps.career_best_bowling_runs, pms.runs_conceded)
					ELSE ps.career_best_bowling_runs 
				END,
				career_best_bowling_wickets = GREATEST(ps.career_best_bowling_wickets, pms.wickets_taken),
				-- fielding
				career_catches = ps.career_catches + pms.catches,
				career_runouts = ps.career_runouts + pms.runouts,
				career_stumpings = ps.career_stumpings + pms.stumpings,
				-- calc strike rate and eco of player
				strike_rate = CASE 
				    WHEN (ps.career_balls_faced + pms.balls_played) > 0 
				    THEN ((ps.career_runs + pms.runs_scored)::DECIMAL / (ps.career_balls_faced + pms.balls_played)::DECIMAL) * 100 
				    ELSE 0 
				END,
				economy = CASE 
				    WHEN (ps.career_balls_bowled + pms.balls_bowled) > 0 
				    THEN ((ps.career_runs_conceded + pms.runs_conceded)::DECIMAL / ((ps.career_balls_bowled + pms.balls_bowled)::DECIMAL / 6.0))
				    ELSE 0 
				END
			FROM player_match_stats pms
			WHERE ps.id = pms.player_id AND pms.match_id = $1;
		`, matchID)
		return err
	})
}

func UndoLastBall(c context.Context, matchID string) error {
	return database.WithTransaction(c, func(tx *sqlx.Tx) error {

		// fetch last boll
		var lastBall struct {
			ID               string         `db:"id"`
			InningsID        string         `db:"innings_id"`
			StrikerID        sql.NullString `db:"striker_id"`
			NonStrikerID     sql.NullString `db:"non_striker_id"`
			BowlerID         sql.NullString `db:"bowler_id"`
			IsLegalBall      sql.NullBool   `db:"is_legal_ball"`
			RunsFromBat      sql.NullInt64  `db:"runs_from_bat"`
			Extras           sql.NullInt64  `db:"extras"`
			IsWicket         sql.NullBool   `db:"is_wicket"`
			ExtraType        sql.NullString `db:"extra_type"`
			WicketType       sql.NullString `db:"wicket_type"`
			OutPlayerID      sql.NullString `db:"out_player_id"`
			FielderID        sql.NullString `db:"fielder_id"`
			PartnershipRuns  sql.NullInt64  `db:"partnership_runs"`
			PartnershipBalls sql.NullInt64  `db:"partnership_balls"`
		}

		err := tx.Get(&lastBall, `
			SELECT 
				b.id,
				b.innings_id,
				b.striker_id,
				b.non_striker_id,
				b.bowler_id,
				b.is_legal_ball,
				b.runs_from_bat,
				b.extras,
				b.is_wicket,
				b.extra_type,
				b.wicket_type,
				b.out_player_id,
				b.fielder_id,
				b.partnership_runs,
				b.partnership_balls 
			FROM balls b
			JOIN innings i ON i.id = b.innings_id
			WHERE i.match_id = $1
			ORDER BY b.created_at DESC 
			LIMIT 1
		`, matchID)

		if err != nil {
			if err == sql.ErrNoRows {
				return fmt.Errorf("no balls found to undo")
			}
			return err
		}

		// fetch last ball for partnershp
		var prevBall struct {
			PartnershipRuns  sql.NullInt64 `db:"partnership_runs"`
			PartnershipBalls sql.NullInt64 `db:"partnership_balls"`
		}

		err = tx.Get(&prevBall, `
			SELECT b.partnership_runs, b.partnership_balls
			FROM balls b
			JOIN innings i ON i.id = b.innings_id
			WHERE i.match_id = $1
			  AND b.innings_id = $2
			  AND b.id != $3
			ORDER BY b.created_at DESC
			LIMIT 1
		`, matchID, lastBall.InningsID, lastBall.ID)
		hasPrevBall := (err == nil)

		// del the last ball
		_, err = tx.Exec("DELETE FROM balls WHERE id = $1", lastBall.ID)
		if err != nil {
			return err
		}

		// extract values
		isLegal := true
		if lastBall.IsLegalBall.Valid {
			isLegal = lastBall.IsLegalBall.Bool
		}
		runsBat := 0
		if lastBall.RunsFromBat.Valid {
			runsBat = int(lastBall.RunsFromBat.Int64)
		}
		ext := 0
		if lastBall.Extras.Valid {
			ext = int(lastBall.Extras.Int64)
		}
		isWkt := false
		if lastBall.IsWicket.Valid {
			isWkt = lastBall.IsWicket.Bool
		}

		// rev player_match_stats for batter
		batterBalls := 0
		if isLegal {
			batterBalls = 1 // Only revert ball if it was a legal delivery
		}
		fours, sixes := 0, 0
		if runsBat == 4 {
			fours = 1
		}
		if runsBat == 6 {
			sixes = 1
		}

		if lastBall.StrikerID.Valid {
			_, err = tx.Exec(`
				UPDATE player_match_stats 
				SET runs_scored = GREATEST(runs_scored - $1, 0), balls_played = GREATEST(balls_played - $2, 0),
					fours = GREATEST(fours - $3, 0), sixes = GREATEST(sixes - $4, 0),
					updated_at = CURRENT_TIMESTAMP
				WHERE match_id = $5 AND player_id = $6
			`, runsBat, batterBalls, fours, sixes, matchID, lastBall.StrikerID.String)
			if err != nil {
				return err
			}
		}

		// rev player_match_stats for bowler
		bowlerBalls := 0
		if isLegal {
			bowlerBalls = 1
		}
		bowlerRuns := runsBat + ext
		if lastBall.ExtraType.Valid && (lastBall.ExtraType.String == "bye" || lastBall.ExtraType.String == "leg_bye") {
			bowlerRuns = 0
		}

		wides, noBalls := 0, 0
		if lastBall.ExtraType.Valid {
			if lastBall.ExtraType.String == "wide" {
				wides = ext
			}
			if lastBall.ExtraType.String == "no_ball" {
				noBalls = ext
			}
		}

		wickets := 0
		if isWkt && lastBall.WicketType.Valid && lastBall.WicketType.String != "run_out" {
			wickets = 1
		}

		if lastBall.BowlerID.Valid {
			_, err = tx.Exec(`
				UPDATE player_match_stats 
				SET runs_conceded = GREATEST(runs_conceded - $1, 0), balls_bowled = GREATEST(balls_bowled - $2, 0),
					wides = GREATEST(wides - $3, 0), no_balls = GREATEST(no_balls - $4, 0), wickets_taken = GREATEST(wickets_taken - $5, 0)
				WHERE match_id = $6 AND player_id = $7
			`, bowlerRuns, bowlerBalls, wides, noBalls, wickets, matchID, lastBall.BowlerID.String)
			if err != nil {
				return err
			}
		}
		// rev player_match_stats for fielder
		if isWkt && lastBall.FielderID.Valid && lastBall.WicketType.Valid {
			catches, runouts, stumpings := 0, 0, 0
			if lastBall.WicketType.String == "caught" || lastBall.WicketType.String == "caught_and_bowled" {
				catches = 1
			}
			if lastBall.WicketType.String == "run_out" {
				runouts = 1
			}
			if lastBall.WicketType.String == "stumped" {
				stumpings = 1
			}

			_, err = tx.Exec(`
				UPDATE player_match_stats 
				SET catches = GREATEST(catches - $1, 0), runouts = GREATEST(runouts - $2, 0), stumpings = GREATEST(stumpings - $3, 0),
					updated_at = CURRENT_TIMESTAMP
				WHERE match_id = $4 AND player_id = $5
			`, catches, runouts, stumpings, matchID, lastBall.FielderID.String)
			if err != nil {
				return err
			}
		}

		// rev wick
		if isWkt && lastBall.OutPlayerID.Valid {
			_, err = tx.Exec(`
				UPDATE player_match_stats SET is_out = false, updated_at = CURRENT_TIMESTAMP
				WHERE match_id = $1 AND player_id = $2
			`, matchID, lastBall.OutPlayerID.String)
			if err != nil {
				return err
			}
		}

		// prep inning & match stat rev
		totalRuns := runsBat + ext
		legalBalls := 0
		if isLegal {
			legalBalls = 1
		}
		totalWickets := 0
		if isWkt {
			totalWickets = 1
		}

		// rev innings
		_, err = tx.Exec(`
			UPDATE innings 
			SET 
				total_runs = GREATEST(total_runs - $1, 0),
				total_wickets = GREATEST(total_wickets - $2, 0),
				total_extras = GREATEST(total_extras - $3, 0),
				legal_balls = GREATEST(legal_balls - $4, 0),
				updated_at = CURRENT_TIMESTAMP
			WHERE id = $5
		`, totalRuns, totalWickets, ext, legalBalls, lastBall.InningsID)
		if err != nil {
			return err
		}

		// rev live_match_stats
		partRuns, partBalls := 0, 0
		if hasPrevBall {
			if prevBall.PartnershipRuns.Valid {
				partRuns = int(prevBall.PartnershipRuns.Int64)
			}
			if prevBall.PartnershipBalls.Valid {
				partBalls = int(prevBall.PartnershipBalls.Int64)
			}
		}

		// Prepare nullable pointers for the UPDATE
		var sID, nsID, bID interface{} = nil, nil, nil
		if lastBall.StrikerID.Valid {
			sID = lastBall.StrikerID.String
		}
		if lastBall.NonStrikerID.Valid {
			nsID = lastBall.NonStrikerID.String
		}
		if lastBall.BowlerID.Valid {
			bID = lastBall.BowlerID.String
		}

		_, err = tx.Exec(`
			UPDATE live_match_stats
			SET current_score = GREATEST(current_score - $1, 0), legal_balls = GREATEST(legal_balls - $2, 0), wickets = GREATEST(wickets - $3, 0),
			    current_over = GREATEST(legal_balls - $2, 0) / 6,
				striker_id = $4, non_striker_id = $5, bowler_id = $6, partnership_runs = $7, partnership_balls = $8,
				last_updated = CURRENT_TIMESTAMP
			WHERE match_id = $9 AND innings_id = $10
		`, totalRuns, legalBalls, totalWickets, sID, nsID, bID, partRuns, partBalls, matchID, lastBall.InningsID)
		if err != nil {
			return err
		}

		// make status ongoing from completed (If undone ball triggered match completion, this reopens it)
		_, err = tx.Exec("UPDATE matches SET status = 'ongoing' WHERE id = $1 AND status = 'completed'", matchID)
		return err
	})
}
