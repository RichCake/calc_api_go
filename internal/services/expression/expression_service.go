package expression

import (
	"log/slog"
	"strconv"
	"time"

	"github.com/RichCake/calc_api_go/internal/models"
	"github.com/RichCake/calc_api_go/internal/services/calculation"
	"github.com/RichCake/calc_api_go/internal/storage"
)

// ExpressionService содержит логику обработки выражений
type ExpressionService struct {
	storage *storage.Storage
}

// Конструктор сервиса выражений
func NewExpressionService(s *storage.Storage) *ExpressionService {
	return &ExpressionService{storage: s}
}

// Обработка нового выражения
func (s *ExpressionService) ProcessExpression(expressionStr string) (int, error) {
	slog.Info("ExpressionService.ProcessExpression: Начало обработки выражения", "expression", expressionStr)
	postfix, err := calculation.ToPostfix(expressionStr)
	if err != nil {
		slog.Error("ExpressionService.ProcessExpression: Ошибка при переводе в постфиксную запись")
		return 0, err
	}
	slog.Info("ExpressionService.ProcessExpression: Выражение переведено в постфиксную запись", "expression", postfix)

	newExpression := models.Expression{
		Status: "processing",
		BinaryTree: calculation.BuildTree(postfix),
	}
	slog.Info("ExpressionService.ProcessExpression: Построено бинарное дерево", "BinaryTree", newExpression.BinaryTree)

	// Добавляем в хранилище и получаем ID
	expressionID := s.storage.AddExpression(newExpression)
	slog.Info("ExpressionService.ProcessExpression: Выражение добавлено в хранилище", "id", expressionID)

	// Создаём задачи для spare-узлов
	spareNodes := newExpression.BinaryTree.FindSpareNodes()
	slog.Info("ExpressionService.ProcessExpression: Получен список свободных узлов", "spareNodes", spareNodes)
	for _, node := range spareNodes {
		s.createTaskForSpareNode(node, expressionID)
	}
	slog.Info("ExpressionService.ProcessExpression: Конец обработки выражения", "id", expressionID)
	return expressionID, nil
}

// Создание задачи для spare-узлов
func (s *ExpressionService) createTaskForSpareNode(node *calculation.TreeNode, expressionID int) {
	slog.Info("ExpressionService.createTaskForSpareNode: Начала создания задачи для узла", "node", node, "expressionID", expressionID)
	arg1, _ := strconv.ParseFloat(node.Left.Val, 64)
	arg2, _ := strconv.ParseFloat(node.Right.Val, 64)
	slog.Info("ExpressionService.createTaskForSpareNode: Извлечены аргументы для задачи", "arg1", arg1, "arg2", arg2)

	task := models.Task{
		ExpressionID:  expressionID,
		Status:        "pending",
		Arg1:          arg1,
		Arg2:          arg2,
		Operation:     node.Val,
		OperationTime: getOperationTime(node.Val),
	}
	slog.Info("ExpressionService.createTaskForSpareNode: Сформирована задача", "task", task)

	taskID := s.storage.AddTask(task)
	slog.Info("ExpressionService.createTaskForSpareNode: Задача добавлена в хранилище", "taskID", taskID)
	node.TaskID = taskID
	slog.Info("ExpressionService.createTaskForSpareNode: Вершине присвоена ID задачи", "node", node, "taskID", taskID)
}

func getOperationTime(operation string) time.Duration {
	switch operation {
	case "+":
		return time.Second
	case "-":
		return time.Second
	case "*":
		return 2 * time.Second
	case "/":
		return 2 * time.Second
	default:
		return 0
	}
}

// Получение списка выражений
func (s *ExpressionService) GetExpressions() []models.Expression {
	slog.Info("ExpressionService.GetExpressions: Выдаем список выражений")
	return s.storage.GetExpressions()
}

func (s *ExpressionService) GetPendingTask() *models.Task {
	return s.storage.GetPendingTask()
}

func (s *ExpressionService) ProcessIncomingTask(task_id int, result float64) {
	slog.Info("ExpressionService.ProcessIncomingTask: Начало обработки входящей задачи", "task_id", task_id, "result", result)
	task := s.storage.FindTaskByID(task_id)
	slog.Info("ExpressionService.ProcessIncomingTask: Найдена задача по task_id", "task_id", task_id, "task", task)
	task.Status = "done"
	slog.Info("ExpressionService.ProcessIncomingTask: Задаче установлен статус done", "task", task)
	expression := s.storage.FindExpressionByID(task.ExpressionID)
	slog.Info("ExpressionService.ProcessIncomingTask: Найдено выражение для задачи", "task.ExpressionID", task.ExpressionID, "expression", expression)
	s.storage.DeleteTask(task_id)
	slog.Info("ExpressionService.ProcessIncomingTask: Задача удалена", "task_id", task_id)
	parent_task_node, node := expression.BinaryTree.FindParentAndNodeByTaskID(task_id)
	slog.Info("ExpressionService.ProcessIncomingTask: Найден узел и родитель узла для задачи", "task_id", task_id, "node", node, "parent_task_node", parent_task_node)
	expression.BinaryTree.ReplaceNodeWithValue(node, result)
	slog.Info("ExpressionService.ProcessIncomingTask: Узел заменен на значение", "expression.BinaryTree", expression.BinaryTree)
	if parent_task_node == nil {
		slog.Info("ExpressionService.ProcessIncomingTask: У узла нет родителя, завершаем вычисление выражения")
		s.SolveExpression(expression, result)
		return
	}
	if parent_task_node.IsSpare() {
		slog.Info("ExpressionService.ProcessIncomingTask: Из родителя можно сделать задачу")
		s.createTaskForSpareNode(parent_task_node, expression.ID)
	}
	slog.Info("ExpressionService.ProcessIncomingTask: Из родителя нельзя сделать задачу")
}

func (s *ExpressionService) SolveExpression(expression *models.Expression, result float64) {
	expression.Result = result
	expression.Status = "solve"
	expression.BinaryTree = nil
	slog.Info("ExpressionService.SolveExpression: Выражение решено", "expression", expression)
}
