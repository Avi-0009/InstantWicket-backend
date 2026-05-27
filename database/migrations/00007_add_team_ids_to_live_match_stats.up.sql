ALTER TABLE live_match_stats
    ADD COLUMN batting_team_id UUID REFERENCES teams(id);

ALTER TABLE live_match_stats
    ADD COLUMN bowling_team_id UUID REFERENCES teams(id);