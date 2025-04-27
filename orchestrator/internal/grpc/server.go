package grpc

import (
	"context"

	orchestrator "github.com/RichCake/calc_api_go/protos/gen/go/orchestrator"
	"google.golang.org/grpc"
)

type serverAPI struct {
	orchestrator.UnimplementedTasksServer
}

func Register(gRPC *grpc.Server) {
	orchestrator.RegisterTasksServer(gRPC, &serverAPI{})
}

func (s *serverAPI) SendTask(
	ctx context.Context,
	req *orchestrator.SendTaskRequest,
) (*orchestrator.SendTaskResponse, error) {
	panic("aaaAAa")
}

func (s *serverAPI) ReceiveTask(
	ctx context.Context,
	req *orchestrator.ReceiveTaskRequest,
) (*orchestrator.ReceiveTaskResponse, error) {
	panic("aaaAAa")
}
