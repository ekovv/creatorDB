package creator

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"time"
)

type Creator struct {
	log     *slog.Logger
	storage Storage
}

type Storage interface {
	SaveConnection(ctx context.Context, user, login string, password []byte, dbName, dbType string, connectionString string) error
	GetConnection(ctx context.Context, user, login string, dbName, dbType string) (string, []byte, error)
}

func New(log *slog.Logger, storage Storage) *Creator {
	return &Creator{
		log:     log,
		storage: storage,
	}
}

func (c *Creator) CreateDB(ctx context.Context, user, login, password, dbName, dbType string) (string, error) {
	const op = "create.CreateDB"
	log := c.log.With("op", op)

	lastConnString, pass, err := c.storage.GetConnection(ctx, user, login, dbName, dbType)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {

		} else {
			return "", fmt.Errorf("error checking last connection")
		}
	}
	if pass != nil {
		err = bcrypt.CompareHashAndPassword(pass, []byte(password))
		if err == nil {
			return lastConnString, nil
		}
	}

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return "", fmt.Errorf("failed to create client: %v", err)
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
		return "", fmt.Errorf("error starting container")
	}

	time.Sleep(15 * time.Second)

	var port string
	inspect, err := cli.ContainerInspect(ctx, resp.ID)
	if err != nil {
		return "", fmt.Errorf("failed to inspect")
	}
	ports, ok := inspect.NetworkSettings.Ports["5432/tcp"]
	if ok && len(ports) > 0 {
		port = ports[0].HostPort
	}

	connStr := ""
	if dbType == "postgresql" {
		connStr = fmt.Sprintf("postgres://%s:%s@localhost:%s/%s?sslmode=disable", login, password, port, dbName)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	err = c.storage.SaveConnection(ctx, user, login, hashedPassword, dbName, dbType, connStr)
	if err != nil {
		return "", fmt.Errorf("error saving connection")
	}

	log.Info("creating database")
	return connStr, nil
}
