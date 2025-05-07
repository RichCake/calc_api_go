package grpc

import (
	"context"
	"errors"
	"log/slog"

	"github.com/RichCake/calc_api_go/orchestrator/internal/services/expression"
	orchestrator "github.com/RichCake/calc_api_go/protos/gen/go/orchestrator"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type serverAPI struct {
	orchestrator.TasksServer
	service *expression.ExpressionService
}

func Register(gRPC *grpc.Server, service *expression.ExpressionService) {
	orchestrator.RegisterTasksServer(gRPC, &serverAPI{service: service})
}

func (s *serverAPI) SendTask(
	ctx context.Context,
	req *orchestrator.SendTaskRequest,
) (*orchestrator.SendTaskResponse, error) {
	task, err := s.service.GetPendingTask()
	if errors.Is(err, expression.ErrPendingTaskNotFount) {
		return nil, status.Errorf(codes.NotFound, "no pending task found")
	} else if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get pending task: %v", err)
	}
	slog.Info("GRPC. Send task", "task", task)
	response := orchestrator.SendTaskResponse{
		Id: int64(task.ID),
		Arg1: task.Arg1,
		Arg2: task.Arg2,
		Operation: task.Operation,
		OperationTimeMs: task.OperationTime.Nanoseconds(),
	}
	return &response, nil
}

func (s *serverAPI) ReceiveTask(
	ctx context.Context,
	req *orchestrator.ReceiveTaskRequest,
) (*orchestrator.ReceiveTaskResponse, error) {
	slog.Info("GRPC. Receive task", "request", req)
	s.service.ProcessIncomingTask(int(req.Id), req.Result)
	return &orchestrator.ReceiveTaskResponse{}, nil
}
