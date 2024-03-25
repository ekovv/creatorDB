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

	var image string
	var env []string
	var port nat.Port
	switch dbType {
	case "postgresql":
		image = "postgres"
		env = []string{
			fmt.Sprintf("POSTGRES_USER=%s", login),
			fmt.Sprintf("POSTGRES_PASSWORD=%s", password),
			fmt.Sprintf("POSTGRES_DB=%s", dbName),
		}
		port = "5432/tcp"
	case "mysql":
		image = "mysql"
		env = []string{
			fmt.Sprintf("MYSQL_USER=%s", login),
			fmt.Sprintf("MYSQL_PASSWORD=%s", password),
			fmt.Sprintf("MYSQL_DATABASE=%s", dbName),
		}
		port = "3306/tcp"
	default:
		return "", fmt.Errorf("unsupported database type: %s", dbType)
	}

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: image,
		Env:   env,
		ExposedPorts: nat.PortSet{
			port: struct{}{},
		},
	}, &container.HostConfig{
		PortBindings: nat.PortMap{
			port: []nat.PortBinding{
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

	var portStr string
	inspect, err := cli.ContainerInspect(ctx, resp.ID)
	if err != nil {
		return "", fmt.Errorf("failed to inspect")
	}
	ports, ok := inspect.NetworkSettings.Ports[port]
	if ok && len(ports) > 0 {
		portStr = ports[0].HostPort
	}

	connStr := ""
	switch dbType {
	case "postgresql":
		connStr = fmt.Sprintf("postgres://%s:%s@localhost:%s/%s?sslmode=disable", login, password, portStr, dbName)
	case "mysql":
		connStr = fmt.Sprintf("%s:%s@tcp(localhost:%s)/%s", login, password, portStr, dbName)
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
