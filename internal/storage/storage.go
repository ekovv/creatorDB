package storage

import (
	"context"
	"creatorDB/internal/config"
	"database/sql"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

type Storage struct {
	conn *sql.DB
}

func NewPostgresDBStorage(config config.Config) (*Storage, error) {
	db, err := sql.Open("postgres", config.DB)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to db %w", err)
	}
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to create migrate driver, %w", err)
	}
	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"smartTables", driver)
	if err != nil {
		return nil, fmt.Errorf("failed to migrate: %w", err)
	}
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return nil, fmt.Errorf("failed to do migrate %w", err)
	}
	s := &Storage{
		conn: db,
	}

	return s, s.CheckConnection()
}

func (s *Storage) CheckConnection() error {
	if err := s.conn.Ping(); err != nil {
		return fmt.Errorf("failed to connect to db %w", err)
	}

	return nil
}

func (s *Storage) SaveConnection(ctx context.Context, login string, password []byte, dbName, dbType string, connectionString string) error {
	query := `INSERT INTO users (login, password, dbName, dbType, connectionString) VALUES ($1, $2, $3, $4, $5)`
	_, err := s.conn.ExecContext(ctx, query, login, password, dbName, dbType, connectionString)
	if err != nil {
		return err
	}
	return nil
}

func (s *Storage) GetConnection(ctx context.Context, login string, dbName, dbType string) (string, []byte, error) {
	var connectionString string
	var password []byte
	query := `SELECT connectionString, password FROM users WHERE login = $1 AND dbName = $2 AND dbType = $3;`
	err := s.conn.QueryRowContext(ctx, query, login, dbName, dbType).Scan(&connectionString, &password)
	if err != nil {
		return "", nil, err
	}

	return connectionString, password, nil
}
