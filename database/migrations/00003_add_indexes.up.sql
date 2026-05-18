CREATE INDEX IF NOT EXISTS idx_users_name
    ON users(name);

CREATE INDEX IF NOT EXISTS idx_users_phone_no
    ON users(phone_no);



CREATE INDEX IF NOT EXISTS idx_player_stats_user_id
    ON player_stats(user_id);

CREATE INDEX IF NOT EXISTS idx_player_stats_runs
    ON player_stats(career_runs DESC);

CREATE INDEX IF NOT EXISTS idx_player_stats_wickets
    ON player_stats(career_wickets DESC);



CREATE INDEX IF NOT EXISTS idx_matches_status
    ON matches(status);

CREATE INDEX IF NOT EXISTS idx_matches_created_by
    ON matches(created_by);

CREATE INDEX IF NOT EXISTS idx_matches_team_a
    ON matches(team_a_id);

CREATE INDEX IF NOT EXISTS idx_matches_team_b
    ON matches(team_b_id);

CREATE INDEX IF NOT EXISTS idx_matches_date
    ON matches(match_date DESC);



CREATE INDEX IF NOT EXISTS idx_match_players_match_id
    ON match_players(match_id);

CREATE INDEX IF NOT EXISTS idx_match_players_player_id
    ON match_players(player_id);

CREATE INDEX IF NOT EXISTS idx_match_players_team_id
    ON match_players(team_id);



CREATE INDEX IF NOT EXISTS idx_innings_match_id
    ON innings(match_id);

CREATE INDEX IF NOT EXISTS idx_innings_batting_team
    ON innings(batting_team_id);

CREATE INDEX IF NOT EXISTS idx_innings_bowling_team
    ON innings(bowling_team_id);



CREATE INDEX IF NOT EXISTS idx_balls_over
    ON balls(innings_id, over_number);

CREATE INDEX IF NOT EXISTS idx_balls_created_at
    ON balls(created_at DESC);

CREATE INDEX IF NOT EXISTS idx_balls_out_player
    ON balls(out_player_id);

CREATE INDEX IF NOT EXISTS idx_balls_innings_over_ball
    ON balls(innings_id, over_number, ball_number);



CREATE INDEX IF NOT EXISTS idx_player_match_stats_player
    ON player_match_stats(player_id);

CREATE INDEX IF NOT EXISTS idx_player_match_stats_team
    ON player_match_stats(team_id);

CREATE INDEX IF NOT EXISTS idx_player_match_stats_runs
    ON player_match_stats(runs_scored DESC);

CREATE INDEX IF NOT EXISTS idx_player_match_stats_wickets
    ON player_match_stats(wickets_taken DESC);



CREATE INDEX IF NOT EXISTS idx_live_match_stats_innings
    ON live_match_stats(innings_id);