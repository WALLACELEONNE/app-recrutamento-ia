DROP INDEX IF EXISTS idx_turns_session;
DROP INDEX IF EXISTS idx_sessions_candidate;

DROP TABLE IF EXISTS session_turns;
DROP TABLE IF EXISTS interview_sessions;
DROP TABLE IF EXISTS candidates;
DROP TABLE IF EXISTS jobs;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS organizations;

DROP TYPE IF EXISTS role_type;
DROP TYPE IF EXISTS session_status;
