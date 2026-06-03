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
			    partnership_runs = CASE WHEN $2 > 0 THEN 0 ELSE partnership_runs + $1 END,
			    partnership_balls = CASE WHEN $2 > 0 THEN 0 ELSE partnership_balls + $3 END,
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

			if b.IsWicket {
				outcome = "W"
			} else if b.ExtraType != nil && *b.ExtraType == "wide" {
				outcome = fmt.Sprintf("%dwd", b.Extras) // "1wd", etc.
			} else if b.ExtraType != nil && *b.ExtraType == "no_ball" {
				outcome = fmt.Sprintf("%dnb", b.RunsFromBat+b.Extras) // "7nb", etc.
			} else {
				outcome = fmt.Sprintf("%d", b.RunsFromBat) // "0", "4", "6", etc.
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
			COALESCE(pms.maiden_overs, 0) AS maidens,
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

		// fetch the last recorded ball
		var lastBall models.Ball
		err := tx.Get(&lastBall, `
			SELECT id, match_id, innings_id, over_number, ball_number, striker_id, non_striker_id, bowler_id, 
			       is_legal_ball, runs_from_bat, extras, is_wicket, extra_type, wicket_type, out_player_id, 
			       fielder_id, partnership_runs, partnership_balls, created_at 
			FROM balls 
			WHERE match_id = $1 
			ORDER BY id DESC LIMIT 1
		`, matchID)

		if err != nil {
			return fmt.Errorf("no balls found to undo")
		}

		// fetch the second-to-last ball to safely restore the partnership score
		var prevBall struct {
			PartnershipRuns  int `db:"partnership_runs"`
			PartnershipBalls int `db:"partnership_balls"`
		}
		hasPrevBall := false
		err = tx.Get(&prevBall, `
			SELECT partnership_runs, partnership_balls FROM balls 
			WHERE match_id = $1 AND innings_id = $2 AND id != $3 
			ORDER BY id DESC LIMIT 1
		`, matchID, lastBall.InningsID, lastBall.ID)
		if err == nil {
			hasPrevBall = true
		}

		// del the last ball
		_, err = tx.Exec("DELETE FROM balls WHERE id = $1", lastBall.ID)
		if err != nil {
			return err
		}

		// rev batter stats
		batterBalls := 0
		if lastBall.IsLegalBall || (lastBall.ExtraType != nil && *lastBall.ExtraType == "no_ball") {
			batterBalls = 1
		}
		fours, sixes := 0, 0
		if lastBall.RunsFromBat == 4 {
			fours = 1
		}
		if lastBall.RunsFromBat == 6 {
			sixes = 1
		}

		if lastBall.StrikerID != nil {
			_, err = tx.Exec(`
				UPDATE player_match_stats 
				SET runs_scored = GREATEST(runs_scored - $1, 0),
					balls_played = GREATEST(balls_played - $2, 0),
					fours = GREATEST(fours - $3, 0),
					sixes = GREATEST(sixes - $4, 0)
				WHERE match_id = $5 AND player_id = $6
			`, lastBall.RunsFromBat, batterBalls, fours, sixes, matchID, *lastBall.StrikerID)
			if err != nil {
				return err
			}
		}

		// rev bowler stats
		bowlerBalls := 0
		if lastBall.IsLegalBall {
			bowlerBalls = 1
		}
		bowlerRuns := lastBall.RunsFromBat + lastBall.Extras
		if lastBall.ExtraType != nil && (*lastBall.ExtraType == "bye" || *lastBall.ExtraType == "leg_bye") {
			bowlerRuns = 0
		}

		wides, noBalls := 0, 0
		if lastBall.ExtraType != nil {
			if *lastBall.ExtraType == "wide" {
				wides = lastBall.Extras
			}
			if *lastBall.ExtraType == "no_ball" {
				noBalls = lastBall.Extras
			}
		}

		wickets := 0
		if lastBall.IsWicket && lastBall.WicketType != nil && *lastBall.WicketType != "run_out" {
			wickets = 1
		}

		if lastBall.BowlerID != nil {
			_, err = tx.Exec(`
				UPDATE player_match_stats 
				SET runs_conceded = GREATEST(runs_conceded - $1, 0),
					balls_bowled = GREATEST(balls_bowled - $2, 0),
					wides = GREATEST(wides - $3, 0),
					no_balls = GREATEST(no_balls - $4, 0),
					wickets_taken = GREATEST(wickets_taken - $5, 0)
				WHERE match_id = $6 AND player_id = $7
			`, bowlerRuns, bowlerBalls, wides, noBalls, wickets, matchID, *lastBall.BowlerID)
			if err != nil {
				return err
			}
		}
		// rev fielder stats
		if lastBall.IsWicket && lastBall.FielderID != nil && lastBall.WicketType != nil {
			catches, runouts, stumpings := 0, 0, 0
			if *lastBall.WicketType == "caught" {
				catches = 1
			}
			if *lastBall.WicketType == "run_out" {
				runouts = 1
			}
			if *lastBall.WicketType == "stumped" {
				stumpings = 1
			}

			_, err = tx.Exec(`
				UPDATE player_match_stats 
				SET catches = GREATEST(catches - $1, 0), runouts = GREATEST(runouts - $2, 0), stumpings = GREATEST(stumpings - $3, 0)
				WHERE match_id = $4 AND player_id = $5
			`, catches, runouts, stumpings, matchID, *lastBall.FielderID)
			if err != nil {
				return err
			}
		}
		// rev wicket
		if lastBall.IsWicket && lastBall.OutPlayerID != nil {
			_, err = tx.Exec(`
				UPDATE player_match_stats SET is_out = false 
				WHERE match_id = $1 AND player_id = $2 
				AND team_id = (SELECT batting_team_id FROM live_match_stats WHERE innings_id = $3)
			`, matchID, *lastBall.OutPlayerID, lastBall.InningsID)
			if err != nil {
				return err
			}
		}
		// rev live match stats
		totalRuns := lastBall.RunsFromBat + lastBall.Extras
		legalBalls := 0
		if lastBall.IsLegalBall {
			legalBalls = 1
		}
		totalWickets := 0
		if lastBall.IsWicket {
			totalWickets = 1
		}
		partRuns, partBalls := 0, 0
		if hasPrevBall {
			partRuns = prevBall.PartnershipRuns
			partBalls = prevBall.PartnershipBalls
		}

		_, err = tx.Exec(`
			UPDATE live_match_stats
			SET current_score = GREATEST(current_score - $1, 0), legal_balls = GREATEST(legal_balls - $2, 0), wickets = GREATEST(wickets - $3, 0),
				striker_id = $4, non_striker_id = $5, bowler_id = $6, partnership_runs = $7, partnership_balls = $8
			WHERE match_id = $9 AND innings_id = $10
		`, totalRuns, legalBalls, totalWickets, lastBall.StrikerID, lastBall.NonStrikerID, lastBall.BowlerID, partRuns, partBalls, matchID, lastBall.InningsID)
		if err != nil {
			return err
		}
		// If the ball triggered the end of the match, this unlocks it!
		_, err = tx.Exec("UPDATE matches SET status = 'ongoing' WHERE id = $1 AND status = 'completed'", matchID)
		return err
	})
}
