package agent

import (
	"context"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	pb "github.com/RichCake/calc_api_go/protos/gen/go/orchestrator"
)

type task struct {
	ID            int           `json:"id"`
	Arg1          float64       `json:"arg1"`
	Arg2          float64       `json:"arg2"`
	Operation     string        `json:"operation"`
	OperationTime time.Duration `json:"operation_time"`
}

type solvedTask struct {
	ID     int     `json:"id"`
	Result float64 `json:"result"`
}

func solveTask(t task) solvedTask {
	solved := solvedTask{ID: t.ID}

	time.Sleep(t.OperationTime)

	switch t.Operation {
	case "+":
		solved.Result = t.Arg1 + t.Arg2
	case "-":
		solved.Result = t.Arg1 - t.Arg2
	case "*":
		solved.Result = t.Arg1 * t.Arg2
	case "/":
		if t.Arg2 == 0 {
			log.Printf("Ошибка: деление на 0 в задаче ID %d\n", t.ID)
			solved.Result = 0
		} else {
			solved.Result = t.Arg1 / t.Arg2
		}
	default:
		log.Printf("Ошибка: неизвестная операция %s в задаче ID %d\n", t.Operation, t.ID)
	}

	return solved
}

func worker(tasks <-chan task, results chan<- solvedTask, wg *sync.WaitGroup) {
	defer wg.Done()

	for t := range tasks {
		timer := time.NewTimer(t.OperationTime)
		<-timer.C
		solved := solveTask(t)
		results <- solved
	}
}

func RunAgent() {
	// Берем из env разные переменные
	taskPort, exists := os.LookupEnv("TASKS_PORT")
	if !exists {
		taskPort = "50051"
	}

	addr := "localhost" + ":" + taskPort
	workerCountStr, exists := os.LookupEnv("AGENT_COMPUTING_POWER")
	if !exists {
		workerCountStr = "10"
	}
	var workerCount int
	if num, err := strconv.Atoi(workerCountStr); err != nil {
		workerCount = num
	} else {
		workerCount = 10
	}
	// устанавливаем соединение с сервером
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// создаем клиента
	client := pb.NewTasksClient(conn)

	inputCh := make(chan task, workerCount)
	outputCh := make(chan solvedTask, workerCount)
	var wg sync.WaitGroup

	// эта горутина постоянно просит задачи
	go func() {
		defer close(inputCh)
		for {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			
			resp, err := client.SendTask(ctx, &pb.SendTaskRequest{})
			cancel()
			if err != nil {
				st, ok := status.FromError(err)
				if ok {
					switch st.Code() {
					case codes.NotFound:
						log.Print("Задач нет, ждем...")
						time.Sleep(time.Second)
					case codes.Unavailable:
						log.Print("Сервер недоступен")
						time.Sleep(time.Second)
					default:
						log.Printf("gRPC ошибка: %v (%s)", st.Message(), st.Code())
						time.Sleep(time.Second)
					}
				} else {
					log.Printf("Неизвестная ошибка: %v", err)
				}
				continue
			}

			t := task{
				ID:            int(resp.Id),
				Arg1:          resp.Arg1,
				Arg2:          resp.Arg2,
				Operation:     resp.Operation,
				OperationTime: time.Duration(resp.OperationTimeMs),
			}
			log.Printf("Получена задача: %+v", t)
			inputCh <- t
		}
	}()

	// Запускаем воркеров
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go worker(inputCh, outputCh, &wg)
	}

	// горутина, которая отправляет решения
	go func() {
		for res := range outputCh {
			log.Printf("Отправляем решение %v", res)
			req := pb.ReceiveTaskRequest{
				Id: int64(res.ID),
				Result: res.Result,
			}

			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			client.ReceiveTask(ctx, &req)
			cancel()
		}
	}()

	wg.Wait()
	close(outputCh)
}
