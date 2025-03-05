package storage

//
// В этом модуле низкоуровневая логика взаимодействия
// со списком задач и выражений
//
// Все методы понятны и без моих комментариев
//

import (
	"sync"
	"sync/atomic"

	"github.com/RichCake/calc_api_go/internal/models"
)

type Storage struct {
	mu             sync.Mutex
	taskList       []models.Task
	expressionList []models.Expression
	expressionID   int64
	taskID         int64
}

func NewStorage() *Storage {
	return &Storage{
		taskList:       make([]models.Task, 0),
		expressionList: make([]models.Expression, 0),
	}
}

func (s *Storage) AddExpression(expression models.Expression) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	expression.ID = int(atomic.AddInt64(&s.expressionID, 1))
	s.expressionList = append(s.expressionList, expression)
	return expression.ID
}

func (s *Storage) GetExpressions() []models.Expression {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.expressionList
}

func (s *Storage) AddTask(task models.Task) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	task.ID = int(atomic.AddInt64(&s.taskID, 1))
	s.taskList = append(s.taskList, task)
	return task.ID
}

func (s *Storage) GetPendingTask() *models.Task {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.taskList {
		if s.taskList[i].Status == "pending" {
			s.taskList[i].Status = "in progress"
			return &s.taskList[i]
		}
	}
	return nil
}

func (s *Storage) DeleteTask(task_id int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.taskList {
		if s.taskList[i].ID == task_id {
			s.taskList[i] = s.taskList[len(s.taskList)-1]
			s.taskList = s.taskList[:len(s.taskList)-1]
			break
		}
	}
}

// Удаление всех задач, связанных с выражением
func (s *Storage) DeleteTaskByExpressionID(expression_id int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.taskList {
		if s.taskList[i].ExpressionID == expression_id {
			s.taskList[i] = s.taskList[len(s.taskList)-1]
			s.taskList = s.taskList[:len(s.taskList)-1]
			break
		}
	}
}

func (s *Storage) FindTaskByID(task_id int) *models.Task {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.taskList {
		if s.taskList[i].ID == task_id {
			return &s.taskList[i]
		}
	}
	return nil
}

func (s *Storage) FindExpressionByID(expression_id int) *models.Expression {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.expressionList {
		if s.expressionList[i].ID == expression_id {
			return &s.expressionList[i]
		}
	}
	return nil
}
