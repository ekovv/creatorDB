package creator

import (
	"context"
	"log/slog"
)

type Creator struct {
	log *slog.Logger
}

func New(log *slog.Logger) *Creator {
	return &Creator{
		log: log,
	}
}

func (c *Creator) CreateDB(ctx context.Context, login, password, DbaName, DbType string) (string, error) {
	panic("not implemented")
}
