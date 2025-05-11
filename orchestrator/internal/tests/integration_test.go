package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/RichCake/calc_api_go/orchestrator/internal/application"
	"github.com/RichCake/calc_api_go/protos/gen/go/orchestrator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestExpressionCalculation(t *testing.T) {
	app := application.New(true)
	go func() {
		err := app.RunServer()
		require.NoError(t, err)
	}()
	defer app.Close()

	time.Sleep(time.Second)

	// Регистрация
	registerResp, err := http.Post("http://localhost:8080/api/v1/register", "application/json",
		strings.NewReader(`{"login": "testuser", "password": "testpass"}`))
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, registerResp.StatusCode)

	// Вход
	loginResp, err := http.Post("http://localhost:8080/api/v1/login", "application/json",
		strings.NewReader(`{"login": "testuser", "password": "testpass"}`))
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, loginResp.StatusCode)

	var loginResult struct {
		AccessToken string `json:"access_token"`
	}
	err = json.NewDecoder(loginResp.Body).Decode(&loginResult)
	require.NoError(t, err)
	require.NotEmpty(t, loginResult.AccessToken)

	// Отправка выражения
	calcReq, err := http.NewRequest("POST", "http://localhost:8080/api/v1/calculate",
		strings.NewReader(`{"expression": "2+2"}`))
	require.NoError(t, err)
	calcReq.Header.Set("Content-Type", "application/json")
	calcReq.Header.Set("Authorization", "Bearer "+loginResult.AccessToken)

	calcResp, err := http.DefaultClient.Do(calcReq)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, calcResp.StatusCode)

	var calcResult struct {
		ID int `json:"id"`
	}
	err = json.NewDecoder(calcResp.Body).Decode(&calcResult)
	require.NoError(t, err)
	require.Greater(t, calcResult.ID, 0)

	// Подключение к gRPC
	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()

	client := proto.NewTasksClient(conn)

	ctx := context.Background()

	// Получение задачи
	taskResp, err := client.SendTask(ctx, &proto.SendTaskRequest{})
	require.NoError(t, err)
	require.NotNil(t, taskResp)

	var result float64
	switch taskResp.Operation {
	case "+":
		result = taskResp.Arg1 + taskResp.Arg2
	case "-":
		result = taskResp.Arg1 - taskResp.Arg2
	case "*":
		result = taskResp.Arg1 * taskResp.Arg2
	case "/":
		if taskResp.Arg2 != 0 {
			result = taskResp.Arg1 / taskResp.Arg2
		} else {
			result = 0
		}
	}

	// Отправка задачи
	_, err = client.ReceiveTask(ctx, &proto.ReceiveTaskRequest{
		Id:     taskResp.Id,
		Result: result,
	})
	require.NoError(t, err)

	// Проверка результата
	statusReq, err := http.NewRequest("GET", fmt.Sprintf("http://localhost:8080/api/v1/expressions/%d", calcResult.ID), nil)
	require.NoError(t, err)
	statusReq.Header.Set("Authorization", "Bearer "+loginResult.AccessToken)

	statusResp, err := http.DefaultClient.Do(statusReq)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, statusResp.StatusCode)

	var statusResult struct {
		ID     int     `json:"id"`
		Status string  `json:"status"`
		Result float64 `json:"result"`
	}
	err = json.NewDecoder(statusResp.Body).Decode(&statusResult)
	require.NoError(t, err)
	assert.Equal(t, "solve", statusResult.Status)
	assert.Equal(t, result, statusResult.Result)
}
