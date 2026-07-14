ALTER TABLE quiz_participants DROP COLUMN IF EXISTS score;
ALTER TABLE quiz_participants DROP COLUMN IF EXISTS name;
ALTER TABLE quiz_sessions DROP COLUMN IF EXISTS current_index;
ALTER TABLE quiz_sessions DROP COLUMN IF EXISTS question_ids;
ALTER TABLE quiz_sessions DROP COLUMN IF EXISTS status;
ALTER TABLE quiz_sessions DROP COLUMN IF EXISTS host_id;
