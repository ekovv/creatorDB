package creator

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
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
	log := c.log.With("op", op)

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return "", err
	}

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: "postgres",
		Env: []string{
			fmt.Sprintf("POSTGRES_USER=%s", login),
			fmt.Sprintf("POSTGRES_PASSWORD=%s", password),
			fmt.Sprintf("POSTGRES_DB=%s", dbName),
		},
		ExposedPorts: nat.PortSet{
			"5432/tcp": struct{}{},
		},
	}, &container.HostConfig{
		PortBindings: nat.PortMap{
			"5432/tcp": []nat.PortBinding{
				{
					HostIP: "0.0.0.0",
				},
			},
		},
	}, nil, nil, "")
	if err != nil {
		return "", err
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return "", err
	}

	inspect, err := cli.ContainerInspect(ctx, resp.ID)
	if err != nil {
		return "", err
	}

	port := inspect.NetworkSettings.Ports["5432/tcp"][0].HostPort

	connStr := fmt.Sprintf("host=localhost port=%s user=%s password=%s dbname=%s sslmode=disable", port, login, password, dbName)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return "", err
	}

	err = c.storage.SaveConnection(ctx, login, password, dbName, dbType)
	if err != nil {
		return "", err
	}

	_, err = db.ExecContext(ctx, "CREATE DATABASE "+dbName+";")
	if err != nil {
		return "", err
	}
	log.Info("creating database")
	return connStr, nil
}
