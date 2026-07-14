package postgres

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/richardktran/realtime-quiz/user-service/internal/repository"
	"github.com/richardktran/realtime-quiz/user-service/pkg/model"
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

	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL
		)`); err != nil {
		return nil, err
	}

	return &Repository{db: db}, nil
}

func (r *Repository) CreateUser(ctx context.Context, user *model.User) (*model.User, error) {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO users (id, name) VALUES ($1, $2) ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name`,
		user.ID, user.Name,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (*model.User, error) {
	var user model.User
	err := r.db.QueryRowContext(ctx, `SELECT id, name FROM users WHERE id=$1`, id).Scan(&user.ID, &user.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}
