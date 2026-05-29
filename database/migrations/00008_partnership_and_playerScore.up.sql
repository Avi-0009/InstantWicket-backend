ALTER TABLE player_stats
    ADD COLUMN career_thirties INT DEFAULT 0;

ALTER TABLE live_match_stats
    ADD COLUMN partnership_runs INT DEFAULT 0,
    ADD COLUMN partnership_balls INT DEFAULT 0;