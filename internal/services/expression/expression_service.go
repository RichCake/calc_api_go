package expression

//
//
// Этот модуль содержит логику обработки выражений
// ExpressionService взаимодействует со списком заданий
// и выражений через хранилище Storage
//
//

import (
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/RichCake/calc_api_go/internal/config"
	"github.com/RichCake/calc_api_go/internal/models"
	"github.com/RichCake/calc_api_go/internal/services/calculation"
	"github.com/RichCake/calc_api_go/internal/storage"
)

type ExpressionService struct {
	storage    *storage.Storage
	timeConfig config.TimeConfig
}

func NewExpressionService(s *storage.Storage, tc config.TimeConfig) *ExpressionService {
	return &ExpressionService{storage: s, timeConfig: tc}
}

func (s *ExpressionService) Close() {
	s.storage.Close()
}

// Обработчик входящего выражения.
// Он запускается один раз для каждого выражения
func (s *ExpressionService) ProcessExpression(expressionStr string) (int, error) {
	// Первым делом переводим в постфиксную запись
	postfix, err := calculation.ToPostfix(expressionStr)
	if err != nil {
		slog.Error("ExpressionService.ProcessExpression: Error in processing to postfix")
		return 0, err
	}

	// Формируем выражение и здесь же строим бинарное дерево
	newExpression := models.Expression{
		Status:     "processing",
		BinaryTree: calculation.BuildTree(postfix),
	}

	// Добавляем выражение в хранилище
	expressionID, err := s.storage.AddExpression(newExpression)
	newExpression.ID = expressionID
	if err != nil {
		slog.Error("ExpressionService.ProcessExpression: error in storage", "error", err.Error())
		return 0, ErrStorage
	}

	// Ищем вершины у которых дети это числа...
	spareNodes := newExpression.BinaryTree.FindSpareNodes()
	for _, node := range spareNodes {
		// ..., и создаем для них задачи
		err := s.createTaskForSpareNode(node, expressionID)
		if err != nil {
			return expressionID, err
		}
	}
	err = s.storage.SetExpressionTree(newExpression.ID, newExpression.BinaryTree)
	if err != nil {
		return expressionID, err
	}
	return expressionID, nil
}

// Создание задачи для свободного узла. Свободный - это узел, у которого оба ребенка - числа
func (s *ExpressionService) createTaskForSpareNode(node *calculation.TreeNode, expressionID int) error {
	arg1, _ := strconv.ParseFloat(node.Left.Val, 64)
	arg2, _ := strconv.ParseFloat(node.Right.Val, 64)

	if arg2 == 0 && node.Val == "/" {
		// если делим на ноль, то закрываем выражение
		expression, err := s.storage.FindExpressionByID(expressionID)
		if errors.Is(err, storage.ErrItemNotFound) {
			slog.Error("ExpressionService.createTaskForSpareNode: Expression not found", "expressionID", expressionID)
			return ErrService
		} else if err != nil {
			slog.Error("ExpressionService.createTaskForSpareNode: error in service", "error", err.Error())
			return ErrService
		}
		s.closeExpressionWithError(expression, "division by zero")
	}

	task := models.Task{
		ExpressionID:  expressionID,
		Status:        "pending",
		Arg1:          arg1,
		Arg2:          arg2,
		Operation:     node.Val,
		OperationTime: s.getOperationTime(node.Val),
	}
	slog.Info("ExpressionService.createTaskForSpareNode: Task created", "task", task)

	// Добавляем задачу в хранилище
	taskID, err := s.storage.AddTask(task)
	if err != nil {
		slog.Error("ExpressionService.createTaskForSpareNode: error in storage", "error", err.Error())
		return ErrStorage
	}
	node.TaskID = taskID
	// записать выражение в базочку
	err = s.storage.SetTaskID(expressionID, taskID)
	if errors.Is(err, storage.ErrItemNotFound) {
		// TODO. Можно добавить отдельную обработку этой ошибки 
		// но как будто она здесь не случится
		slog.Error("ExpressionService.createTaskForSpareNode: item not found", "error", err.Error())
		return ErrStorage
	} else if err != nil {
		slog.Error("ExpressionService.createTaskForSpareNode: error in service", "error", err.Error())
		return ErrStorage
	}
	return nil
}

func (s ExpressionService) getOperationTime(operation string) time.Duration {
	switch operation {
	case "+":
		return s.timeConfig.TimeAdd
	case "-":
		return s.timeConfig.TimeSub
	case "*":
		return s.timeConfig.TimeMul
	case "/":
		return s.timeConfig.TimeDiv
	default:
		return 0
	}
}

// Получение списка выражений из хранилища
func (s *ExpressionService) GetExpressions() ([]models.Expression, error) {
	expressions, err := s.storage.GetExpressions()
	if err != nil {
		slog.Error("ExpressionService.createTaskForSpareNode: error in storage", "error", err.Error())
		return expressions, ErrStorage
	}
	return expressions, nil
}

// Получение выражения по ID
func (s *ExpressionService) GetExpressionByID(id int) (models.Expression, error) {
	expression, err := s.storage.FindExpressionByID(id)
	if errors.Is(err, storage.ErrItemNotFound) {
		return expression, ErrExpressionNotFound
	} else if err != nil {
		return expression, err
	}
	return expression, nil
}

// Если задача не будет решена, то установит статус в ожидании
func (s *ExpressionService) setTimerToTask(task models.Task) {
	timer := time.NewTimer(task.OperationTime + time.Second)
	<-timer.C
	if task.Status == "in progress" {
		slog.Warn("Reset task because it was not solved")
		s.storage.SetTaskStatus(task.ID, "pending")
	}
}

// Этот метод раздает задачу, которая ждет отправки
func (s *ExpressionService) GetPendingTask() (models.Task, error) {
	task, err := s.storage.GetPendingTask()
	if errors.Is(err, storage.ErrItemNotFound) {
		return task, ErrPendingTaskNotFount
	} 
	go s.setTimerToTask(task)
	return task, nil
}

// Обработка входящей задачи. Или по другому: запускается когда агент отправляет результат задачи
func (s *ExpressionService) ProcessIncomingTask(task_id int, result float64) error {
	task, err := s.storage.FindTaskByID(task_id)
	if errors.Is(err, storage.ErrItemNotFound) {
		return ErrTaskNotFound
	} else if err != nil {
		slog.Error("ExpressionService.ProcessIncomingTask: error in storage", "error", err.Error())
		return err
	}
	// Если воркер долго решал задачу и она ушла новому, но старый все же отправил решение
	if task.Status == "done" {
		slog.Warn("ExpressionService.ProcessIncomingTask: receive task that already solved")
		return nil
	}
	err = s.storage.SetTaskStatus(task.ID, "done")
	if err != nil {
		slog.Error("ExpressionService.ProcessIncomingTask: error in storage", "error", err.Error())
		return ErrStorage
	}
	expression, err := s.storage.FindExpressionByID(task.ExpressionID)
	if err != nil {
		slog.Error("ExpressionService.ProcessIncomingTask: error in storage", "error", err.Error())
		return ErrStorage
	}
	// Здесь самое интересное. Когда пришел результат задачи, мы заменяем вершину задачи на результат...
	parent_task_node, node := expression.BinaryTree.FindParentAndNodeByTaskID(task_id)
	fmt.Println("result FindParentAndNodeByTaskID", parent_task_node, node)
	expression.BinaryTree.ReplaceNodeWithValue(node, result)
	if parent_task_node == nil {
		// ... если у вершины нет родителя, то это значит, что это корень дерева и выражение решено
		s.solveExpression(expression, result)
		return nil
	}
	// ... и проверяем, можно ли из родителя сделать задачу
	if parent_task_node.IsSpare() {
		err = s.createTaskForSpareNode(parent_task_node, expression.ID)
		if err != nil {
			return err
		}
	}
	err = s.storage.SetExpressionTree(expression.ID, expression.BinaryTree)
	if err != nil {
		return err
	}
	return nil
}

func (s *ExpressionService) closeExpressionWithError(expression models.Expression, errorMsg string) {
	status := "error " + errorMsg
	s.storage.SetExpressionStatus(expression.ID, status)
	s.storage.DeleteTaskByExpressionID(expression.ID)
}

func (s *ExpressionService) solveExpression(expression models.Expression, result float64) {
	status := "solve"
	s.storage.SetExpressionResult(expression.ID, result)
	s.storage.SetExpressionStatus(expression.ID, status)
}
