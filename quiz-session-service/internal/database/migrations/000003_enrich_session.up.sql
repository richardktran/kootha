-- Enrich quiz_sessions with host, status, questions, and current index
ALTER TABLE quiz_sessions ADD COLUMN IF NOT EXISTS host_id TEXT NOT NULL DEFAULT '';
ALTER TABLE quiz_sessions ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'waiting';
ALTER TABLE quiz_sessions ADD COLUMN IF NOT EXISTS question_ids JSONB NOT NULL DEFAULT '[]';
ALTER TABLE quiz_sessions ADD COLUMN IF NOT EXISTS current_index INT NOT NULL DEFAULT 0;

ALTER TABLE quiz_participants ADD COLUMN IF NOT EXISTS name TEXT NOT NULL DEFAULT '';
ALTER TABLE quiz_participants ADD COLUMN IF NOT EXISTS score INT NOT NULL DEFAULT 0;

CREATE UNIQUE INDEX IF NOT EXISTS quiz_participants_user_session_uidx
  ON quiz_participants (user_id, quiz_session_id);
