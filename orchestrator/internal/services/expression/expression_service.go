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
	"log/slog"
	"strconv"
	"time"

	"github.com/RichCake/calc_api_go/orchestrator/internal/config"
	"github.com/RichCake/calc_api_go/orchestrator/internal/models"
	"github.com/RichCake/calc_api_go/orchestrator/internal/services/calculation"
	"github.com/RichCake/calc_api_go/orchestrator/internal/storage"
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
func (s *ExpressionService) ProcessExpression(expressionStr string, user_id int) (int, error) {
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
		UserID: user_id,
	}

	// Добавляем выражение в хранилище
	expressionID, err := s.storage.SaveExpression(&newExpression)
	if err != nil {
		slog.Error("ExpressionService.ProcessExpression: error in storage", "error", err.Error())
		return 0, ErrStorage
	}

	// Ищем вершины у которых дети это числа...
	spareNodes := newExpression.BinaryTree.FindSpareNodes()
	for _, node := range spareNodes {
		// ..., и создаем для них задачи
		task, err := s.createTaskForSpareNode(node, &newExpression)
		if errors.Is(err, ErrZeroDivisionTask) {
			s.closeExpressionWithError(&newExpression, "division by zero")
		} else if err != nil {
			slog.Error("ExpressionService.ProcessExpression: error in service", "error", err.Error())
			return expressionID, ErrService
		}
		_, err = s.storage.SaveTask(&task)
		if err != nil {
			slog.Error("ExpressionService.ProcessExpression: error in storage", "error", err.Error())
			return expressionID, ErrStorage
		}
		node.TaskID = task.ID
	}
	_, err = s.storage.SaveExpression(&newExpression)
	if err != nil {
		slog.Error("ExpressionService.ProcessExpression: error in storage", "error", err.Error())
		return expressionID, ErrStorage
	}
	return expressionID, nil
}

// Создание задачи для свободного узла. Свободный - это узел, у которого оба ребенка - числа
func (s *ExpressionService) createTaskForSpareNode(node *calculation.TreeNode, expression *models.Expression) (models.Task, error) {
	var task models.Task
	arg1, _ := strconv.ParseFloat(node.Left.Val, 64)
	arg2, _ := strconv.ParseFloat(node.Right.Val, 64)

	if arg2 == 0 && node.Val == "/" {
		// если делим на ноль, то закрываем выражение
		return task, ErrZeroDivisionTask
	}

	task = models.Task{
		ExpressionID:  expression.ID,
		Status:        "pending",
		Arg1:          arg1,
		Arg2:          arg2,
		Operation:     node.Val,
		OperationTime: s.getOperationTime(node.Val),
	}
	slog.Info("ExpressionService.createTaskForSpareNode: Task created", "task", task)
	return task, nil
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
func (s *ExpressionService) GetExpressions(user_id int) ([]models.Expression, error) {
	expressions, err := s.storage.GetExpressions(user_id)
	if err != nil {
		slog.Error("ExpressionService.createTaskForSpareNode: error in storage", "error", err.Error())
		return expressions, ErrStorage
	}
	return expressions, nil
}

// Получение выражения по ID
func (s *ExpressionService) GetExpressionByID(id int, user_id int) (models.Expression, error) {
	expression, err := s.storage.GetExpression(id)
	if errors.Is(err, storage.ErrItemNotFound) || expression.UserID != user_id {
		return expression, ErrExpressionNotFound
	} else if err != nil {
		return expression, err
	}
	return expression, nil
}

// Если задача не будет решена, то установит статус в ожидании.
// Может создавать гонку, пока не использую
func (s *ExpressionService) setTimerToTask(task models.Task) {
	timer := time.NewTimer(task.OperationTime + 3*time.Second)
	<-timer.C
	task, _ = s.storage.GetTask(task.ID)
	if task.Status == "in progress" {
		slog.Warn("Reset task because it was not solved")
		task.Status = "pending"
		s.storage.SaveTask(&task)
	}
}

// Этот метод раздает задачу, которая ждет отправки
func (s *ExpressionService) GetPendingTask() (models.Task, error) {
	task, err := s.storage.GetPendingTask()
	if errors.Is(err, storage.ErrItemNotFound) {
		return task, ErrPendingTaskNotFount
	}
	task.Status = "in progress"
	_, err = s.storage.SaveTask(&task)
	if err != nil {
		return task, ErrStorage
	}
	// go s.setTimerToTask(task)
	return task, nil
}

// Обработка входящей задачи. Или по другому: запускается когда агент отправляет результат задачи
func (s *ExpressionService) ProcessIncomingTask(task_id int, result float64) error {
	task, err := s.storage.GetTask(task_id)
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
	task.Status = "done"
	_, err = s.storage.SaveTask(&task)
	if err != nil {
		slog.Error("ExpressionService.ProcessIncomingTask: error in storage", "error", err.Error())
		return ErrStorage
	}
	expression, err := s.storage.GetExpression(task.ExpressionID)
	if err != nil {
		slog.Error("ExpressionService.ProcessIncomingTask: error in storage", "error", err.Error())
		return ErrStorage
	}
	// Здесь самое интересное. Когда пришел результат задачи, мы заменяем вершину задачи на результат...
	parent_task_node, node := expression.BinaryTree.FindParentAndNodeByTaskID(task_id)
	if node == nil {
		// У меня тут фантомно спотыкается программа.
		// Ошибка из-за кривого sqlite. Щас должно быть все ок (пожалуйста)
		s.closeExpressionWithError(&expression, "task_id not found. critical error")
		return ErrService
	}
	expression.BinaryTree.ReplaceNodeWithValue(node, result)
	if parent_task_node == nil {
		// ... если у вершины нет родителя, то это значит, что это корень дерева и выражение решено
		s.solveExpression(&expression, result)
		return nil
	}
	// ... и проверяем, можно ли из родителя сделать задачу
	if parent_task_node.IsSpare() {
		new_task, err := s.createTaskForSpareNode(parent_task_node, &expression)
		if errors.Is(err, ErrZeroDivisionTask) {
			s.closeExpressionWithError(&expression, "division by zero")
		} else if err != nil {
			slog.Error("ExpressionService.ProcessExpression: error in service", "error", err.Error())
			return ErrService
		}
		_, err = s.storage.SaveTask(&new_task)
		if err != nil {
			slog.Error("ExpressionService.ProcessExpression: error in storage", "error", err.Error())
			return ErrStorage
		}
		parent_task_node.TaskID = new_task.ID
	}
	_, err = s.storage.SaveExpression(&expression)
	if err != nil {
		return err
	}
	return nil
}

func (s *ExpressionService) closeExpressionWithError(expression *models.Expression, errorMsg string) {
	expression.Status = "error " + errorMsg
	s.storage.SaveExpression(expression)
	s.storage.DeleteTaskByExpressionID(expression.ID)
}

func (s *ExpressionService) solveExpression(expression *models.Expression, result float64) {
	expression.Result = result
	expression.Status = "solve"
	s.storage.SaveExpression(expression)
}
