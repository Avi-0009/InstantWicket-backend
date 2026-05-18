DROP INDEX IF EXISTS idx_users_name;
DROP INDEX IF EXISTS idx_users_phone_no;

DROP INDEX IF EXISTS idx_player_stats_user_id;
DROP INDEX IF EXISTS idx_player_stats_runs;
DROP INDEX IF EXISTS idx_player_stats_wickets;

DROP INDEX IF EXISTS idx_matches_status;
DROP INDEX IF EXISTS idx_matches_created_by;
DROP INDEX IF EXISTS idx_matches_team_a;
DROP INDEX IF EXISTS idx_matches_team_b;
DROP INDEX IF EXISTS idx_matches_date;

DROP INDEX IF EXISTS idx_match_players_match_id;
DROP INDEX IF EXISTS idx_match_players_player_id;
DROP INDEX IF EXISTS idx_match_players_team_id;

DROP INDEX IF EXISTS idx_innings_match_id;
DROP INDEX IF EXISTS idx_innings_batting_team;
DROP INDEX IF EXISTS idx_innings_bowling_team;

DROP INDEX IF EXISTS idx_balls_over;
DROP INDEX IF EXISTS idx_balls_created_at;
DROP INDEX IF EXISTS idx_balls_out_player;
DROP INDEX IF EXISTS idx_balls_innings_over_ball;

DROP INDEX IF EXISTS idx_player_match_stats_player;
DROP INDEX IF EXISTS idx_player_match_stats_team;
DROP INDEX IF EXISTS idx_player_match_stats_runs;
DROP INDEX IF EXISTS idx_player_match_stats_wickets;

DROP INDEX IF EXISTS idx_live_match_stats_innings;