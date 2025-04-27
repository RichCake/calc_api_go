package storage

//
// В этом модуле низкоуровневая логика взаимодействия
// со списком задач и выражений
//
// Все методы понятны и без моих комментариев
//

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/RichCake/calc_api_go/internal/models"
	"github.com/RichCake/calc_api_go/internal/services/calculation"

	_ "github.com/mattn/go-sqlite3"
)

type Storage struct {
	db *sql.DB
}

func NewStorage() *Storage {
	ctx := context.TODO()

	db, err := sql.Open("sqlite3", "store.db")
	if err != nil {
		panic(err)
	}

	err = db.PingContext(ctx)
	if err != nil {
		panic(err)
	}

	if err = createTables(ctx, db); err != nil {
		panic(err)
	}
	return &Storage{
		db: db,
	}
}

func (s *Storage) Close() error {
	return s.db.Close()
}

func (s *Storage) SaveExpression(expression *models.Expression) (int, error) {
	ctx := context.TODO()
	treeBytes, err := calculation.SerializeTree(*expression.BinaryTree)
	if err != nil {
		return 0, err
	}

	if expression.ID == 0 {
		q := `
		INSERT INTO expressions (status, result, binary_tree_bytes)
		VALUES ($1, $2, $3)
		`
		res, err := s.db.ExecContext(ctx, q, expression.Status, expression.Result, treeBytes)
		if err != nil {
			return 0, err
		}
		lastID, err := res.LastInsertId()
		if err != nil {
			return 0, err
		}
		expression.ID = int(lastID)
		return int(lastID), nil
	}

	q := `
	UPDATE expressions
	SET status = $1, result = $2, binary_tree_bytes = $3
	WHERE expression_id = $4
	`
	_, err = s.db.ExecContext(ctx, q, expression.Status, expression.Result, treeBytes, expression.ID)
	if err != nil {
		return 0, err
	}
	return expression.ID, nil
}

func (s *Storage) SaveTask(task *models.Task) (int, error) {
	ctx := context.TODO()
	nanos := task.OperationTime.Nanoseconds()

	if task.ID == 0 {
		q := `
		INSERT INTO tasks (status, arg1, arg2, operation, operation_time, expression_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		`
		res, err := s.db.ExecContext(ctx, q, task.Status, task.Arg1, task.Arg2, task.Operation, nanos, task.ExpressionID)
		if err != nil {
			return 0, err
		}
		lastID, err := res.LastInsertId()
		if err != nil {
			return 0, err
		}
		task.ID = int(lastID)
		return task.ID, nil
	}

	q := `
	UPDATE tasks
	SET status = $1, arg1 = $2, arg2 = $3, operation = $4, operation_time = $5, expression_id = $6
	WHERE task_id = $7
	`
	_, err := s.db.ExecContext(ctx, q, task.Status, task.Arg1, task.Arg2, task.Operation, nanos, task.ExpressionID, task.ID)
	if err != nil {
		return 0, err
	}
	return task.ID, nil
}

func (s *Storage) GetExpressions() ([]models.Expression, error) {
	var expressions []models.Expression
	var q = "SELECT expression_id, status, result FROM expressions"
	ctx := context.TODO()
	rows, err := s.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		e := models.Expression{}
		err := rows.Scan(&e.ID, &e.Status, &e.Result)
		if err != nil {
			return nil, err
		}
		expressions = append(expressions, e)
	}

	return expressions, nil
}

func (s *Storage) GetTasks() []models.Task {
	var tasks []models.Task
	var q = "SELECT task_id, status, arg1, arg2, operation, operation_time, expression_id FROM tasks"
	ctx := context.TODO()
	rows, err := s.db.QueryContext(ctx, q)
	if err != nil {
		return nil
	}
	defer rows.Close()

	for rows.Next() {
		t := models.Task{}
		var nanoseconds int64
		err := rows.Scan(&t.ID, &t.Status, &t.Arg1, &t.Arg2, &t.Operation, &nanoseconds, &t.ExpressionID)
		t.OperationTime = time.Duration(nanoseconds)
		if err != nil {
			return nil
		}
		tasks = append(tasks, t)
	}

	return tasks
}

func (s *Storage) GetPendingTask() (models.Task, error) {
	var task models.Task
	var q = `
	SELECT task_id, status, arg1, arg2, operation, operation_time, expression_id 
	FROM tasks
	WHERE status = $1
	LIMIT 1
	`
	ctx := context.TODO()
	var nanoseconds int64
	err := s.db.QueryRowContext(ctx, q, "pending").Scan(&task.ID, &task.Status, &task.Arg1, &task.Arg2, &task.Operation, &nanoseconds, &task.ExpressionID)
	task.OperationTime = time.Duration(nanoseconds)
	if errors.Is(err, sql.ErrNoRows) {
		return task, ErrItemNotFound
	} else if err != nil {
		return task, err
	}
	fmt.Println(task.Status)
	return task, nil
}

func (s *Storage) DeleteTask(task_id int) error {
	var q = "DELETE tasks WHERE task_id = $1"
	ctx := context.TODO()
	_, err := s.db.ExecContext(ctx, q, task_id)
	if err != nil {
		return err
	}
	return nil
}

// Удаление всех задач, связанных с выражением
func (s *Storage) DeleteTaskByExpressionID(expression_id int) error {
	var q = "DELETE tasks WHERE expression_id = $1"
	ctx := context.TODO()
	_, err := s.db.ExecContext(ctx, q, expression_id)
	if err != nil {
		return err
	}
	return nil
}

func (s *Storage) GetTask(task_id int) (models.Task, error) {
	var task models.Task
	var q = `
	SELECT task_id, status, arg1, arg2, operation, operation_time, expression_id 
	FROM tasks
	WHERE task_id = $1
	`
	ctx := context.TODO()
	var nanoseconds int64
	err := s.db.QueryRowContext(ctx, q, task_id).Scan(&task.ID, &task.Status, &task.Arg1, &task.Arg2, &task.Operation, &nanoseconds, &task.ExpressionID)
	task.OperationTime = time.Duration(nanoseconds)
	if errors.Is(err, sql.ErrNoRows) {
		return task, ErrItemNotFound
	} else if err != nil {
		return task, err
	}
	return task, nil
}

func (s *Storage) GetExpression(expression_id int) (models.Expression, error) {
	var expression models.Expression
	var q = `
	SELECT expression_id, status, result, binary_tree_bytes
	FROM expressions
	WHERE expression_id = $1
	`
	ctx := context.TODO()
	var treeBytes []byte
	err := s.db.QueryRowContext(ctx, q, expression_id).Scan(&expression.ID, &expression.Status, &expression.Result, &treeBytes)
	if errors.Is(err, sql.ErrNoRows) {
		return expression, ErrItemNotFound
	} else if err != nil {
		return expression, err
	}
	tree, err := calculation.DeserializeTree(treeBytes)
	if err != nil {
		return expression, err
	}
	expression.BinaryTree = &tree
	return expression, nil
}
