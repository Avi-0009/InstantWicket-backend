-- USER SESSIONS EXPIRY

ALTER TABLE user_sessions
    ADD COLUMN expires_at TIMESTAMP NOT NULL
        DEFAULT (NOW() + INTERVAL '24 hours');


-- MATCH PLAYERS COMMON PLAYER SUPPORT

ALTER TABLE match_players
DROP CONSTRAINT IF EXISTS match_players_match_id_player_id_key;

ALTER TABLE match_players
    ADD CONSTRAINT unique_match_team_player
        UNIQUE(match_id, team_id, player_id);


-- PLAYER MATCH STATS COMMON PLAYER SUPPORT

ALTER TABLE player_match_stats
DROP CONSTRAINT IF EXISTS player_match_stats_match_id_player_id_key;

ALTER TABLE player_match_stats
    ADD CONSTRAINT unique_match_team_player_stats
        UNIQUE(match_id, team_id, player_id);