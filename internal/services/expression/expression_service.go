package expression

//
//
// Этот модуль содержит логику обработки выражений
// ExpressionService взаимодействует со списком заданий
// и выражений через хранилище Storage
// 
//

import (
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
	expressionID := s.storage.AddExpression(newExpression)

	// Ищем вершины у которых дети это числа...
	spareNodes := newExpression.BinaryTree.FindSpareNodes()
	for _, node := range spareNodes {
		// ..., и создаем для них задачи
		s.createTaskForSpareNode(node, expressionID)
	}
	return expressionID, nil
}

// Создание задачи для свободного узла. Свободный - это узел, у которого оба ребенка - числа
func (s *ExpressionService) createTaskForSpareNode(node *calculation.TreeNode, expressionID int) {
	arg1, _ := strconv.ParseFloat(node.Left.Val, 64)
	arg2, _ := strconv.ParseFloat(node.Right.Val, 64)

	if arg2 == 0 && node.Val == "/" {
		// если делим на ноль, то закрываем выражение
		s.closeExpressionWithError(s.storage.FindExpressionByID(expressionID), "division by zero")
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
	taskID := s.storage.AddTask(task)
	node.TaskID = taskID
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
func (s *ExpressionService) GetExpressions() []models.Expression {
	return s.storage.GetExpressions()
}

// Получение выражения по ID
func (s *ExpressionService) GetExpressionByID(id int) *models.Expression {
	return s.storage.FindExpressionByID(id)
}

func setTimerToTask() {

}

// Этот метод раздает задачу, которая ждет отправки
func (s *ExpressionService) GetPendingTask() *models.Task {
	task := s.storage.GetPendingTask()
	return task
}

// Обработка входящей задачи. Или по другому: запускается когда агент отправляет результат задачи
func (s *ExpressionService) ProcessIncomingTask(task_id int, result float64) {
	task := s.storage.FindTaskByID(task_id)
	task.Status = "done"
	expression := s.storage.FindExpressionByID(task.ExpressionID)
	// Здесь самое интересное. Когда пришел результат задачи мы заменяем вершину задачи на результат...
	parent_task_node, node := expression.BinaryTree.FindParentAndNodeByTaskID(task_id)
	expression.BinaryTree.ReplaceNodeWithValue(node, result)
	if parent_task_node == nil {
		// ... если у вершины нет родителя, то это значит, что это корень дерева и выражение решено
		s.solveExpression(expression, result)
		return
	}
	// ... и проверяем, можно ли из родителя сделать задачу
	if parent_task_node.IsSpare() {
		s.createTaskForSpareNode(parent_task_node, expression.ID)
	}
}

func (s *ExpressionService) closeExpressionWithError(expression *models.Expression, errorMsg string) {
	expression.Status = "error " + errorMsg
	expression.BinaryTree = nil
	s.storage.DeleteTaskByExpressionID(expression.ID)
}

func (s *ExpressionService) solveExpression(expression *models.Expression, result float64) {
	expression.Result = result
	expression.Status = "solve"
	expression.BinaryTree = nil
}
