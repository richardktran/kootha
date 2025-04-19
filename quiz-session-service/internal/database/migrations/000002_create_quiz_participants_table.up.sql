CREATE TABLE IF NOT EXISTS quiz_participants (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    quiz_session_id VARCHAR(255) NOT NULL,
    FOREIGN KEY (quiz_session_id) REFERENCES quiz_sessions(id)
);
