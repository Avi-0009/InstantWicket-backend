package dbHelper

import (
	"database/sql"
	"errors"

	"github.com/Avi-0009/InstantWicket-backend/database"
	"github.com/Avi-0009/InstantWicket-backend/models"
	"github.com/jmoiron/sqlx"
)

func CreateMatch(tx *sqlx.Tx, input models.CreateMatch, teamAPlayers []models.MatchPlayerInput, teamBPlayers []models.MatchPlayerInput, createdBy string) (string, error) {
	var matchID string
	dbUmpireID := sql.NullString{
		String: input.UmpireID,
		Valid:  input.UmpireID != "",
	}

	query := `INSERT INTO matches (
                     team_a_id, team_b_id, toss_winner_team_id, toss_decision, allow_common_player, allow_solo_batting, overs_limit, umpire_id, created_by
                     ) VALUES (
                               $1, $2, $3, $4, $5, $6, $7, $8, $9
                     )
                     RETURNING id
`
	err := tx.Get(&matchID, query, input.TeamAID, input.TeamBID, input.TossWinnerTeamID, input.TossDecision, input.AllowCommonPlayer, input.AllowSoloBatting, input.OversLimit, dbUmpireID, createdBy)
	if err != nil {
		return "", err
	}

	upsertPlayerToMatch := func(p models.MatchPlayerInput, teamID string) error {
		playerID := p.ID

		if p.IsCommonPlayer && p.IsCaptain {
			return errors.New("common player can never be captain")
		}

		// If it's a new player, create them in the DB dynamically!
		if playerID == "" {
			var uID string
			err := tx.Get(&uID, `INSERT INTO users (name, phone_no, password) VALUES ($1, $2, 'guest_account') ON CONFLICT (phone_no) WHERE archived_at IS NULL DO UPDATE SET updated_at = CURRENT_TIMESTAMP RETURNING id`, p.Name, p.PhoneNo)
			if err != nil {
				return err
			}

			err = tx.Get(&playerID, `INSERT INTO player_stats (user_id, batting_style, bowling_style) VALUES ($1, 'Right Handed', 'Right Arm Fast') ON CONFLICT (user_id) DO UPDATE SET updated_at = CURRENT_TIMESTAMP RETURNING id`, uID)
			if err != nil {
				return err
			}
		}

		// Insert into match_players with all the roles (Captain and all)
		_, err := tx.Exec(`INSERT INTO match_players (match_id, team_id, player_id, is_common_player, is_captain, is_wicket_keeper) VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT (match_id, team_id, player_id) DO NOTHING`, matchID, teamID, playerID, p.IsCommonPlayer, p.IsCaptain, p.IsWicketKeeper)

		if err != nil {
			return err
		}

		_, err = tx.Exec(`INSERT INTO player_match_stats (match_id, team_id, player_id) VALUES ($1, $2, $3) ON CONFLICT (match_id, team_id, player_id) DO NOTHING`, matchID, teamID, playerID)
		return err
	}

	for _, p := range teamAPlayers {
		if err := upsertPlayerToMatch(p, input.TeamAID); err != nil {
			return "", err
		}
	}

	for _, p := range teamBPlayers {
		if err := upsertPlayerToMatch(p, input.TeamBID); err != nil {
			return "", err
		}
	}

	return matchID, nil
}

func GetMatches(limit, offset int) ([]models.Match, error) {
	var matches []models.Match
	query := `SELECT 
			m.id, m.team_a_id, 
			tA.name AS team_a_name, 
			m.team_b_id, 
			tB.name AS team_b_name, 
			m.toss_winner_team_id, m.toss_decision, m.allow_common_player, m.allow_solo_batting, 
			m.overs_limit, m.status, 
			COALESCE(
				m.winner_team_id,
				CASE 
					WHEN m.status = 'completed' THEN
						CASE 
							WHEN COALESCE((SELECT total_runs FROM innings WHERE match_id = m.id AND batting_team_id = m.team_a_id LIMIT 1), 0) > 
								 COALESCE((SELECT total_runs FROM innings WHERE match_id = m.id AND batting_team_id = m.team_b_id LIMIT 1), 0)
							THEN m.team_a_id
							WHEN COALESCE((SELECT total_runs FROM innings WHERE match_id = m.id AND batting_team_id = m.team_b_id LIMIT 1), 0) > 
								 COALESCE((SELECT total_runs FROM innings WHERE match_id = m.id AND batting_team_id = m.team_a_id LIMIT 1), 0)
							THEN m.team_b_id
							ELSE NULL -- Tie
						END
					ELSE NULL
				END
			) AS winner_team_id,
			m.man_of_match, m.worst_player, 
			m.umpire_id, m.created_by, m.created_at, m.updated_at,
			u_umpire.name AS umpire_name,
			u_creator.name AS creator_name,
			lms.current_score AS live_score,
			lms.wickets AS live_wickets,
			lms.legal_balls AS live_legal_balls,
			lms.required_runs AS target_runs,
			su.name AS striker_name,
			nsu.name AS non_striker_name,
			bu.name AS bowler_name,
			COALESCE((SELECT total_runs FROM innings WHERE match_id = m.id AND batting_team_id = m.team_a_id LIMIT 1), 0) AS team_a_score,
			COALESCE((SELECT total_wickets FROM innings WHERE match_id = m.id AND batting_team_id = m.team_a_id LIMIT 1), 0) AS team_a_wickets,
			COALESCE((SELECT legal_balls FROM innings WHERE match_id = m.id AND batting_team_id = m.team_a_id LIMIT 1), 0) AS team_a_balls,
			COALESCE((SELECT total_runs FROM innings WHERE match_id = m.id AND batting_team_id = m.team_b_id LIMIT 1), 0) AS team_b_score,
			COALESCE((SELECT total_wickets FROM innings WHERE match_id = m.id AND batting_team_id = m.team_b_id LIMIT 1), 0) AS team_b_wickets,
			COALESCE((SELECT legal_balls FROM innings WHERE match_id = m.id AND batting_team_id = m.team_b_id LIMIT 1), 0) AS team_b_balls
		FROM matches m
		JOIN teams tA ON m.team_a_id = tA.id
		JOIN teams tB ON m.team_b_id = tB.id
		LEFT JOIN live_match_stats lms ON m.id = lms.match_id
		LEFT JOIN player_stats sps ON lms.striker_id = sps.id
		LEFT JOIN users su ON sps.user_id = su.id
		LEFT JOIN player_stats nsps ON lms.non_striker_id = nsps.id
		LEFT JOIN users nsu ON nsps.user_id = nsu.id
		LEFT JOIN player_stats bps ON lms.bowler_id = bps.id
		LEFT JOIN users bu ON bps.user_id = bu.id
		LEFT JOIN users u_umpire ON m.umpire_id = u_umpire.id
		LEFT JOIN users u_creator ON m.created_by = u_creator.id
		ORDER BY m.created_at DESC
		LIMIT $1 OFFSET $2`

	err := database.DB.Select(&matches, query, limit, offset)
	if err != nil {
		return nil, err
	}
	return matches, nil
}

func GetMatchByID(matchID string) (*models.Match, error) {
	var match models.Match
	query := `SELECT 
			m.id, m.team_a_id, 
			tA.name AS team_a_name, 
			m.team_b_id, 
			tB.name AS team_b_name, 
			m.toss_winner_team_id, m.toss_decision, m.allow_common_player, m.allow_solo_batting, 
			m.overs_limit, m.status, 
			COALESCE(
				m.winner_team_id,
				CASE 
					WHEN m.status = 'completed' THEN
						CASE 
							WHEN COALESCE((SELECT total_runs FROM innings WHERE match_id = m.id AND batting_team_id = m.team_a_id LIMIT 1), 0) > 
								 COALESCE((SELECT total_runs FROM innings WHERE match_id = m.id AND batting_team_id = m.team_b_id LIMIT 1), 0)
							THEN m.team_a_id
							WHEN COALESCE((SELECT total_runs FROM innings WHERE match_id = m.id AND batting_team_id = m.team_b_id LIMIT 1), 0) > 
								 COALESCE((SELECT total_runs FROM innings WHERE match_id = m.id AND batting_team_id = m.team_a_id LIMIT 1), 0)
							THEN m.team_b_id
							ELSE NULL -- Tie
						END
					ELSE NULL
				END
			) AS winner_team_id,
			m.man_of_match, m.worst_player, 
			m.umpire_id, m.created_by, m.created_at, m.updated_at,
			u_umpire.name AS umpire_name,
			u_creator.name AS creator_name,
			lms.current_score AS live_score,
			lms.wickets AS live_wickets,
			lms.legal_balls AS live_legal_balls,
			lms.required_runs AS target_runs,
			su.name AS striker_name,
			nsu.name AS non_striker_name,
			bu.name AS bowler_name,
			COALESCE((SELECT total_runs FROM innings WHERE match_id = m.id AND batting_team_id = m.team_a_id LIMIT 1), 0) AS team_a_score,
			COALESCE((SELECT total_wickets FROM innings WHERE match_id = m.id AND batting_team_id = m.team_a_id LIMIT 1), 0) AS team_a_wickets,
			COALESCE((SELECT legal_balls FROM innings WHERE match_id = m.id AND batting_team_id = m.team_a_id LIMIT 1), 0) AS team_a_balls,
			COALESCE((SELECT total_runs FROM innings WHERE match_id = m.id AND batting_team_id = m.team_b_id LIMIT 1), 0) AS team_b_score,
			COALESCE((SELECT total_wickets FROM innings WHERE match_id = m.id AND batting_team_id = m.team_b_id LIMIT 1), 0) AS team_b_wickets,
			COALESCE((SELECT legal_balls FROM innings WHERE match_id = m.id AND batting_team_id = m.team_b_id LIMIT 1), 0) AS team_b_balls
		FROM matches m
		JOIN teams tA ON m.team_a_id = tA.id
		JOIN teams tB ON m.team_b_id = tB.id
		LEFT JOIN live_match_stats lms ON m.id = lms.match_id
		LEFT JOIN player_stats sps ON lms.striker_id = sps.id
		LEFT JOIN users su ON sps.user_id = su.id
		LEFT JOIN player_stats nsps ON lms.non_striker_id = nsps.id
		LEFT JOIN users nsu ON nsps.user_id = nsu.id
		LEFT JOIN player_stats bps ON lms.bowler_id = bps.id
		LEFT JOIN users bu ON bps.user_id = bu.id
		LEFT JOIN users u_umpire ON m.umpire_id = u_umpire.id
		LEFT JOIN users u_creator ON m.created_by = u_creator.id
		WHERE m.id = $1`
	err := database.DB.Get(&match, query, matchID)
	if err != nil {
		return nil, err
	}
	return &match, nil
}

func GetMatchPlayers(matchID string) ([]models.MatchPlayer, error) {
	var players []models.MatchPlayer
	query := `
		SELECT 
			mp.team_id, mp.player_id AS id, u.name, mp.is_common_player, mp.is_captain, mp.is_wicket_keeper, mp.is_retired, mp.returned_to_play
		FROM match_players mp
		JOIN player_stats ps ON mp.player_id = ps.id
		JOIN users u ON ps.user_id = u.id
		WHERE mp.match_id = $1
	`
	err := database.DB.Select(&players, query, matchID)
	if players == nil {
		players = []models.MatchPlayer{}
	}
	return players, err
}
