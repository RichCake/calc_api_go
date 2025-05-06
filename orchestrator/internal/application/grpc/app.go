package grpc

import (
	"fmt"
	"log/slog"
	"net"

	"google.golang.org/grpc"

	"github.com/RichCake/calc_api_go/orchestrator/internal/config"
	grpctasks "github.com/RichCake/calc_api_go/orchestrator/internal/grpc"
	"github.com/RichCake/calc_api_go/orchestrator/internal/services/expression"
)

func RunGRPCServer(expressionService *expression.ExpressionService, config config.Config) {
	host := "localhost"
	port := config.GRPCPort

	addr := fmt.Sprintf("%s:%s", host, port)
	lis, err := net.Listen("tcp", addr)

	if err != nil {
		panic(err)
	}
	
	slog.Info("tcp listener started", "port", port)
	grpcServer := grpc.NewServer()
	grpctasks.Register(grpcServer, expressionService)
	if err := grpcServer.Serve(lis); err != nil {
		panic(err)
	}
}