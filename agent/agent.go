package agent

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

type task struct {
	ID             int           `json:"id"`
	Arg1           float64       `json:"arg1"`
	Arg2           float64       `json:"arg2"`
	Operation      string        `json:"operation"`
	OperationTime  time.Duration `json:"operation_time"`
}

type solvedTask struct {
	ID     int     `json:"id"`
	Result float64 `json:"result"`
}

func solveTask(t task) solvedTask {
	solved := solvedTask{ID: t.ID}

	// Имитация времени выполнения
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
		solved := solveTask(t)
		results <- solved
	}
}

func RunAgent() {
	os.Setenv("TASK_URL", "http://localhost:8080/internal/task") // ПЕРЕДЕЛАТЬ
	taskURL := os.Getenv("TASK_URL")
	workerCount := 10 // ПЕРЕДЕЛАТЬ

	inputCh := make(chan task, workerCount)
	outputCh := make(chan solvedTask, workerCount)
	var wg sync.WaitGroup

	// Получение задач
	go func() {
		defer close(inputCh)
		for {
			resp, err := http.Get(taskURL)
			if resp == nil {
				log.Print("Сервер не отвечает")
				time.Sleep(time.Second)
				continue
			}
			if resp.StatusCode == http.StatusNotFound {
				log.Print("Задач нет")
				time.Sleep(time.Second)
				continue
			}
			if err != nil {
				log.Printf("Ошибка при получении задачи: %v\n", err)
				continue
			}
			log.Printf("Получен ответ %d от сервера", resp.StatusCode)
			var t task
			if err := json.NewDecoder(resp.Body).Decode(&t); err != nil {
				log.Printf("Ошибка при декодировании JSON: %v\n", err)
			}
			log.Printf("Получена задача %v", t)
			resp.Body.Close()

			inputCh <- t
		}
	}()

	// Запуск воркеров
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go worker(inputCh, outputCh, &wg)
	}

	// Обработка результатов
	go func() {
		for res := range outputCh {
			log.Printf("Отправляем решение %v", res)
			data, err := json.Marshal(res)
			if err != nil {
				log.Printf("Ошибка при маршалинге JSON: %v\n", err)
				continue
			}

			resp, err := http.Post(taskURL, "application/json", bytes.NewReader(data))
			if err != nil {
				log.Printf("Ошибка при отправке результата: %v\n", err)
				continue
			}
			resp.Body.Close()
		}
	}()

	// Ожидание завершения всех горутин
	wg.Wait()
	close(outputCh)
}