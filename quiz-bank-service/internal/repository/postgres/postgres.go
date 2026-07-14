package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/richardktran/realtime-quiz/quiz-bank-service/pkg/model"
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

func (r *Repository) GetAll(ctx context.Context) ([]model.Question, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, question, options, correct_answer, time_limit
		FROM questions
		ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanQuestions(rows)
}

func (r *Repository) GetByIDs(ctx context.Context, ids []string) ([]model.Question, error) {
	if len(ids) == 0 {
		return []model.Question{}, nil
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, question, options, correct_answer, time_limit
		FROM questions
		WHERE id = ANY($1)
		ORDER BY id`, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanQuestions(rows)
}

func (r *Repository) GetRandom(ctx context.Context, count int) ([]model.Question, error) {
	if count <= 0 {
		return []model.Question{}, nil
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, question, options, correct_answer, time_limit
		FROM questions
		ORDER BY RANDOM()
		LIMIT $1`, count)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanQuestions(rows)
}

func scanQuestions(rows *sql.Rows) ([]model.Question, error) {
	questions := make([]model.Question, 0)

	for rows.Next() {
		var q model.Question
		var optionsJSON []byte

		if err := rows.Scan(&q.ID, &q.Question, &optionsJSON, &q.CorrectAnswer, &q.TimeLimit); err != nil {
			return nil, err
		}

		if err := json.Unmarshal(optionsJSON, &q.Options); err != nil {
			return nil, err
		}

		questions = append(questions, q)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return questions, nil
}
