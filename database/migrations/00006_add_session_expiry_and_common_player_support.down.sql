-- REMOVE SESSION EXPIRY

ALTER TABLE user_sessions
DROP COLUMN IF EXISTS expires_at;



-- REVERT MATCH PLAYERS CONSTRAINT

ALTER TABLE match_players
DROP CONSTRAINT IF EXISTS unique_match_team_player;

ALTER TABLE match_players
    ADD CONSTRAINT match_players_match_id_player_id_key
        UNIQUE(match_id, player_id);



-- REVERT PLAYER MATCH STATS CONSTRAINT

ALTER TABLE player_match_stats
DROP CONSTRAINT IF EXISTS unique_match_team_player_stats;

ALTER TABLE player_match_stats
    ADD CONSTRAINT player_match_stats_match_id_player_id_key
        UNIQUE(match_id, player_id);