package create

import (
	"context"
	createv1 "github.com/ekovv/protosDB/gen/go/creator"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Create interface {
	CreateDB(ctx context.Context, user, login, password, dbName, dbType string) (string, error)
}

type serverAPI struct {
	createv1.UnimplementedCreatorServer
	create Create
}

func Register(gRPC *grpc.Server, create Create) {
	createv1.RegisterCreatorServer(gRPC, &serverAPI{create: create})
}

func (s *serverAPI) CreateDB(ctx context.Context, req *createv1.CreateDBRequest) (*createv1.CreateDBResponse, error) {
	if req.GetLogin() == "" {
		return nil, status.Error((codes.InvalidArgument), "You must provide a login")
	}

	if req.GetPassword() == "" {
		return nil, status.Error((codes.InvalidArgument), "You must provide a password")
	}

	if req.GetDbName() == "" || req.GetDbType() == "" {
		return nil, status.Error((codes.InvalidArgument), "You must provide a name for database")
	}

	if req.GetUser() == "" {
		return nil, status.Error((codes.InvalidArgument), "You must provide a user")
	}

	connectionString, err := s.create.CreateDB(ctx, req.GetUser(), req.GetLogin(), req.GetPassword(), req.GetDbName(), req.GetDbType())
	if err != nil {
		return nil, status.Error(codes.Internal, "internal error creating")
	}

	return &createv1.CreateDBResponse{
		ConnectionString: connectionString,
	}, nil
}
