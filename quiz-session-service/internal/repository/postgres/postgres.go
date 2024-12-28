package postgres

import (
	"context"
	"database/sql"
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

	return &Repository{
		db: db,
	}, nil
}

func (r *Repository) GetSessionById(ctx context.Context, id string) (*model.QuizSession, error) {
	var session model.QuizSession
	err := r.db.QueryRowContext(ctx, "SELECT id, name, duration FROM quiz_sessions WHERE id=$1", id).Scan(&session.ID, &session.Name, &session.Duration)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}

	return &session, nil
}

func (r *Repository) CreateQuizSession(ctx context.Context, session *model.QuizSession) (*model.QuizSession, error) {
	_, err := r.db.ExecContext(ctx, "INSERT INTO quiz_sessions (id, name, duration) VALUES ($1, $2, $3)", session.ID, session.Name, session.Duration)

	if err != nil {
		return nil, err
	}
	return session, nil
}
