package create

import (
	"context"
	createv1 "github.com/ekovv/protosDB/gen/go/creator"
	"google.golang.org/grpc"
)

type serverAPI struct {
	createv1.UnimplementedCreatorServer
}

func Register(gRPC *grpc.Server) {
	createv1.RegisterCreatorServer(gRPC, &serverAPI{})
}

func (s *serverAPI) CreateDB(ctx context.Context, req *createv1.CreateDBRequest) (*createv1.CreateDBResponse, error) {
	return &createv1.CreateDBResponse{
		ConnectionString: "blablablabla",
	}, nil
}
