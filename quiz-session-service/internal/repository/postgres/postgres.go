package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/richardktran/realtime-quiz/quiz-session-service/internal/repository"
	"github.com/richardktran/realtime-quiz/quiz-session-service/pkg/model"
)

type Repository struct {
	db *sql.DB
}

type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

func New(config Config) (*Repository, error) {
	connectionStr := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		config.Host, config.User, config.Password, config.DBName, config.Port)

	db, err := sql.Open("postgres", connectionStr)
	if err != nil {
		return nil, err
	}

	return &Repository{db: db}, nil
}

func (r *Repository) GetSessionById(ctx context.Context, id string) (*model.QuizSession, error) {
	var session model.QuizSession
	var questionIDsJSON []byte

	err := r.db.QueryRowContext(ctx,
		`SELECT id, name, duration, host_id, status, question_ids, current_index
		 FROM quiz_sessions WHERE id=$1`, id,
	).Scan(
		&session.ID,
		&session.Name,
		&session.Duration,
		&session.HostID,
		&session.Status,
		&questionIDsJSON,
		&session.CurrentIndex,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}

	if len(questionIDsJSON) > 0 {
		_ = json.Unmarshal(questionIDsJSON, &session.QuestionIDs)
	}
	if session.QuestionIDs == nil {
		session.QuestionIDs = []string{}
	}

	participants, err := r.ListParticipants(ctx, id)
	if err != nil {
		return nil, err
	}
	session.Participants = participants

	return &session, nil
}

func (r *Repository) CreateQuizSession(ctx context.Context, session *model.QuizSession) (*model.QuizSession, error) {
	if session.Status == "" {
		session.Status = model.StatusWaiting
	}
	questionIDs, err := json.Marshal(session.QuestionIDs)
	if err != nil {
		questionIDs = []byte("[]")
	}

	_, err = r.db.ExecContext(ctx,
		`INSERT INTO quiz_sessions (id, name, duration, host_id, status, question_ids, current_index)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 ON CONFLICT (id) DO NOTHING`,
		session.ID, session.Name, session.Duration, session.HostID, session.Status, questionIDs, session.CurrentIndex,
	)
	if err != nil {
		return nil, err
	}
	return session, nil
}

func (r *Repository) UpdateSession(ctx context.Context, session *model.QuizSession) error {
	questionIDs, err := json.Marshal(session.QuestionIDs)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx,
		`UPDATE quiz_sessions
		 SET name=$2, duration=$3, host_id=$4, status=$5, question_ids=$6, current_index=$7
		 WHERE id=$1`,
		session.ID, session.Name, session.Duration, session.HostID, session.Status, questionIDs, session.CurrentIndex,
	)
	return err
}

func (r *Repository) JoinQuiz(ctx context.Context, sessionId, userId, name string) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO quiz_participants (user_id, quiz_session_id, name, score)
		 VALUES ($1, $2, $3, 0)
		 ON CONFLICT (user_id, quiz_session_id) DO UPDATE SET name = EXCLUDED.name`,
		userId, sessionId, name,
	)
	return err
}

func (r *Repository) ListParticipants(ctx context.Context, sessionId string) ([]model.Participant, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT user_id, name, score FROM quiz_participants WHERE quiz_session_id=$1`, sessionId,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var participants []model.Participant
	for rows.Next() {
		var p model.Participant
		if err := rows.Scan(&p.ID, &p.Name, &p.Score); err != nil {
			return nil, err
		}
		participants = append(participants, p)
	}
	if participants == nil {
		participants = []model.Participant{}
	}
	return participants, rows.Err()
}

func (r *Repository) UpdateHost(ctx context.Context, sessionId, hostId string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE quiz_sessions SET host_id=$2 WHERE id=$1`, sessionId, hostId)
	return err
}
