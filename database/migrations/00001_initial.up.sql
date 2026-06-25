CREATE EXTENSION IF NOT EXISTS pgcrypto;



CREATE TABLE users (
                       id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

                       name TEXT NOT NULL,

                       phone_no TEXT UNIQUE NOT NULL,

                       password TEXT,

                       created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

                       updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

                       archived_at TIMESTAMP
);



CREATE TABLE user_sessions (
                               id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

                               user_id UUID NOT NULL,

                               created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

                               archived_at TIMESTAMP,

                               FOREIGN KEY (user_id)
                                   REFERENCES users(id)
);



CREATE TABLE player_stats (
                              id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

                              user_id UUID UNIQUE NOT NULL,

                              batting_style TEXT,

                              bowling_style TEXT,

                              career_matches INTEGER DEFAULT 0,

                              career_runs INTEGER DEFAULT 0,

                              career_wickets INTEGER DEFAULT 0,

                              career_catches INTEGER DEFAULT 0,

                              career_runouts INTEGER DEFAULT 0,

                              career_stumpings INTEGER DEFAULT 0,

                              career_fours INTEGER DEFAULT 0,

                              career_sixes INTEGER DEFAULT 0,

                              strike_rate DECIMAL(5,2) DEFAULT 0,

                              economy DECIMAL(5,2) DEFAULT 0,

                              last_played DATE,

                              created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

                              updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

                              FOREIGN KEY (user_id)
                                  REFERENCES users(id)
);



CREATE TABLE teams (
                       id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

                       name TEXT NOT NULL,

                       created_by UUID NOT NULL,

                       created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

                       updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

                       FOREIGN KEY (created_by)
                           REFERENCES users(id)
);



CREATE TABLE matches (
                         id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

                         team_a_id UUID NOT NULL,

                         team_b_id UUID NOT NULL,

                         toss_winner_team_id UUID,

                         toss_decision TEXT,

                         match_date TIMESTAMP,

                         allow_common_player BOOLEAN DEFAULT TRUE,

                         allow_solo_batting BOOLEAN DEFAULT TRUE,

                         overs_limit INTEGER NOT NULL,

                         status TEXT DEFAULT 'ongoing',

                         winner_team_id UUID,

                         man_of_match UUID,

                         worst_player UUID,

                         umpire_id UUID,

                         created_by UUID NOT NULL,

                         created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

                         updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

                         FOREIGN KEY (team_a_id)
                             REFERENCES teams(id),

                         FOREIGN KEY (team_b_id)
                             REFERENCES teams(id),

                         FOREIGN KEY (toss_winner_team_id)
                             REFERENCES teams(id),

                         FOREIGN KEY (winner_team_id)
                             REFERENCES teams(id),

                         FOREIGN KEY (man_of_match)
                             REFERENCES player_stats(id),

                         FOREIGN KEY (worst_player)
                             REFERENCES player_stats(id),

                         FOREIGN KEY (umpire_id)
                             REFERENCES users(id),

                         FOREIGN KEY (created_by)
                             REFERENCES users(id),

                         CHECK (
                             toss_decision IN (
                                               'bat',
                                               'bowl'
                                 )
                                 OR toss_decision IS NULL
                             ),

                         CHECK (
                             status IN (
                                        'ongoing',
                                        'completed',
                                        'abandoned'
                                 )
                             )
);



CREATE TABLE match_players (
                               id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

                               match_id UUID NOT NULL,

                               team_id UUID NOT NULL,

                               player_id UUID NOT NULL,

                               is_common_player BOOLEAN DEFAULT FALSE,

                               is_captain BOOLEAN DEFAULT FALSE,

                               is_wicket_keeper BOOLEAN DEFAULT FALSE,

                               is_retired BOOLEAN DEFAULT FALSE,

                               returned_to_play BOOLEAN DEFAULT FALSE,

                               created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

                               UNIQUE(match_id, player_id),

                               FOREIGN KEY (match_id)
                                   REFERENCES matches(id),

                               FOREIGN KEY (team_id)
                                   REFERENCES teams(id),

                               FOREIGN KEY (player_id)
                                   REFERENCES player_stats(id)
);



CREATE TABLE innings (
                         id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

                         match_id UUID NOT NULL,

                         innings_no INTEGER NOT NULL,

                         batting_team_id UUID NOT NULL,

                         bowling_team_id UUID NOT NULL,

                         striker_id UUID,

                         non_striker_id UUID,

                         bowler_id UUID,

                         total_runs INTEGER DEFAULT 0,

                         total_wickets INTEGER DEFAULT 0,

                         total_extras INTEGER DEFAULT 0,

                         legal_balls INTEGER DEFAULT 0,

                         target_runs INTEGER,

                         status TEXT DEFAULT 'ongoing',

                         started_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

                         completed_at TIMESTAMP,

                         created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

                         updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

                         FOREIGN KEY (match_id)
                             REFERENCES matches(id),

                         FOREIGN KEY (batting_team_id)
                             REFERENCES teams(id),

                         FOREIGN KEY (bowling_team_id)
                             REFERENCES teams(id),

                         FOREIGN KEY (striker_id)
                             REFERENCES player_stats(id),

                         FOREIGN KEY (non_striker_id)
                             REFERENCES player_stats(id),

                         FOREIGN KEY (bowler_id)
                             REFERENCES player_stats(id)
);



CREATE TABLE balls (
                       id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

                       innings_id UUID NOT NULL,

                       over_number INTEGER NOT NULL,

                       ball_number INTEGER NOT NULL,

                       is_legal_ball BOOLEAN DEFAULT TRUE,

                       delivery_type TEXT DEFAULT 'normal',

                       is_free_hit BOOLEAN DEFAULT FALSE,

                       runs_from_bat INTEGER DEFAULT 0,

                       extras INTEGER DEFAULT 0,

                       extra_type TEXT,

                       total_runs INTEGER DEFAULT 0,

                       is_wicket BOOLEAN DEFAULT FALSE,

                       wicket_type TEXT,

                       caught_by UUID,

                       striker_id UUID NOT NULL,

                       non_striker_id UUID,

                       bowler_id UUID NOT NULL,

                       out_player_id UUID,

                       is_runout BOOLEAN DEFAULT FALSE,

                       bounce_count INTEGER DEFAULT 0,

                       is_two_bounce_no_ball BOOLEAN DEFAULT FALSE,

                       bounce_warning_given BOOLEAN DEFAULT FALSE,

                       is_valid BOOLEAN DEFAULT TRUE,

                       created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

                       FOREIGN KEY (innings_id)
                           REFERENCES innings(id),

                       FOREIGN KEY (caught_by)
                           REFERENCES player_stats(id),

                       FOREIGN KEY (striker_id)
                           REFERENCES player_stats(id),

                       FOREIGN KEY (non_striker_id)
                           REFERENCES player_stats(id),

                       FOREIGN KEY (bowler_id)
                           REFERENCES player_stats(id),

                       FOREIGN KEY (out_player_id)
                           REFERENCES player_stats(id),

                       CHECK (
                           delivery_type IN (
                                             'normal',
                                             'wide',
                                             'no_ball',
                                             'two_bounce'
                               )
                           ),

                       CHECK (
                           extra_type IN (
                                          'wide',
                                          'no_ball',
                                          'bye',
                                          'leg_bye'
                               )
                               OR extra_type IS NULL
                           ),

                       CHECK (
                           ball_number >= 1
                               AND ball_number <= 6
                           )
);



CREATE TABLE player_match_stats (
                                    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

                                    match_id UUID NOT NULL,

                                    player_id UUID NOT NULL,

                                    team_id UUID NOT NULL,

                                    runs_scored INTEGER DEFAULT 0,

                                    balls_played INTEGER DEFAULT 0,

                                    fours INTEGER DEFAULT 0,

                                    sixes INTEGER DEFAULT 0,

                                    strike_rate DECIMAL(5,2) DEFAULT 0,

                                    wickets_taken INTEGER DEFAULT 0,

                                    balls_bowled INTEGER DEFAULT 0,

                                    maiden_overs INTEGER DEFAULT 0,

                                    runs_conceded INTEGER DEFAULT 0,

                                    wides INTEGER DEFAULT 0,

                                    no_balls INTEGER DEFAULT 0,

                                    economy DECIMAL(5,2) DEFAULT 0,

                                    catches INTEGER DEFAULT 0,

                                    runouts INTEGER DEFAULT 0,

                                    stumpings INTEGER DEFAULT 0,

                                    dot_balls_played INTEGER DEFAULT 0,

                                    dot_balls_bowled INTEGER DEFAULT 0,

                                    is_duck BOOLEAN DEFAULT FALSE,

                                    is_golden_duck BOOLEAN DEFAULT FALSE,

                                    is_out BOOLEAN DEFAULT FALSE,

                                    batting_position INTEGER,

                                    batting_points INTEGER DEFAULT 0,

                                    bowling_points INTEGER DEFAULT 0,

                                    fielding_points INTEGER DEFAULT 0,

                                    result_points INTEGER DEFAULT 0,

                                    total_points INTEGER DEFAULT 0,

                                    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

                                    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

                                    UNIQUE(match_id, player_id),

                                    FOREIGN KEY (match_id)
                                        REFERENCES matches(id),

                                    FOREIGN KEY (player_id)
                                        REFERENCES player_stats(id),

                                    FOREIGN KEY (team_id)
                                        REFERENCES teams(id)
);



CREATE TABLE live_match_stats (
                                  match_id UUID PRIMARY KEY,

                                  innings_id UUID NOT NULL,

                                  striker_id UUID,

                                  non_striker_id UUID,

                                  bowler_id UUID,

                                  current_over INTEGER DEFAULT 0,

                                  legal_balls INTEGER DEFAULT 0,

                                  current_score INTEGER DEFAULT 0,

                                  wickets INTEGER DEFAULT 0,

                                  required_runs INTEGER,

                                  required_balls INTEGER,

                                  last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

                                  FOREIGN KEY (match_id)
                                      REFERENCES matches(id),

                                  FOREIGN KEY (innings_id)
                                      REFERENCES innings(id),

                                  FOREIGN KEY (striker_id)
                                      REFERENCES player_stats(id),

                                  FOREIGN KEY (non_striker_id)
                                      REFERENCES player_stats(id),

                                  FOREIGN KEY (bowler_id)
                                      REFERENCES player_stats(id)
);



CREATE INDEX idx_balls_innings
    ON balls(innings_id);



CREATE INDEX idx_balls_bowler
    ON balls(bowler_id);



CREATE INDEX idx_balls_striker
    ON balls(striker_id);



CREATE INDEX idx_player_match_stats
    ON player_match_stats(match_id, player_id);