package app

import (
	grpcapp "creatorDB/internal/app/grpc"
	"creatorDB/internal/config"
	"creatorDB/internal/services/creator"
	"creatorDB/internal/storage"
	"log/slog"
)

type App struct {
	GRPCSrv *grpcapp.App
}

func New(log *slog.Logger, grpcPort int, config config.Config) *App {
	stM, err := storage.NewPostgresDBStorage(config)
	if err != nil {
		return nil
	}
	service := creator.New(log, stM)

	grpcApp := grpcapp.New(log, grpcPort, service)

	return &App{
		GRPCSrv: grpcApp,
	}
}
