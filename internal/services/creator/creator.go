package creator

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
)

type Creator struct {
	log     *slog.Logger
	storage Storage
}

type Storage interface {
	SaveConnection(ctx context.Context, login, password, dbName, dbType string) error
	GetConnection(ctx context.Context, login, password, dbName, dbType string) (string, error)
}

func New(log *slog.Logger, storage Storage) *Creator {
	return &Creator{
		log:     log,
		storage: storage,
	}
}

func (c *Creator) CreateDB(ctx context.Context, login, password, dbName, dbType string) (string, error) {
	const op = "create.CreateDB"
	log := c.log.With(slog.String("op", op))

	connStr := fmt.Sprintf("user=%s password=%s", login, password)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return "", err
	}

	err = c.storage.SaveConnection(ctx, login, password, dbName, dbType)
	if err != nil {
		return "", err
	}

	_, err = db.ExecContext(ctx, "CREATE DATABASE `"+dbName+"`;")

	if err != nil {
		return "", err
	}
	log.Info("creating database")
	return fmt.Sprintf("%s dbname=%s", connStr, dbName), nil
}
