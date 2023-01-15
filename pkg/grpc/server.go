package grpc

import (
	"context"
	grpc "github.com/wesen/majuscule/pkg/grpc/api"
)

type Server struct{}

func (s *Server) Complete(ctx context.Context, req *grpc.CompleteRequest) (*grpc.CompleteResponses, error) {
	// Your implementation here
	return nil, nil
}
