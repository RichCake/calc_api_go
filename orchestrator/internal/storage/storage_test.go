package storage

import (
	"testing"

	"github.com/RichCake/calc_api_go/orchestrator/internal/models"
	"github.com/RichCake/calc_api_go/orchestrator/internal/services/calculation"
	"github.com/stretchr/testify/require"
)

func TestSaveAndGetUser(t *testing.T) {
	type testCase struct {
		name      string
		input     *models.User
		expectErr error
	}

	tests := []testCase{
		{
			name: "valid user",
			input: &models.User{
				Login:        "alice",
				PasswordHash: "pass1",
			},
			expectErr: nil,
		},
		{
			name: "duplicate login",
			input: &models.User{
				Login:        "alice",
				PasswordHash: "pass2",
			},
			expectErr: ErrUsernameTaken,
		},
	}

	st := NewStorage(true)
	defer st.Close()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			id, err := st.SaveUser(tc.input)

			if tc.expectErr != nil {
				require.ErrorIs(t, err, tc.expectErr)
				return
			}

			require.NoError(t, err)
			require.Greater(t, id, 0)
			require.Equal(t, id, tc.input.ID)

			userFromDB, err := st.GetUser(tc.input.Login)
			require.NoError(t, err)
			require.Equal(t, tc.input.Login, userFromDB.Login)
			require.Equal(t, tc.input.PasswordHash, userFromDB.PasswordHash)
		})
	}
}

func TestSaveAndGetExpression(t *testing.T) {
	storage := NewStorage(true)
	defer storage.Close()

	tests := []struct {
		name       string
		expression *models.Expression
		wantErr    bool
	}{
		{
			name: "create new expression",
			expression: &models.Expression{
				Status:     "pending",
				Result:     0,
				BinaryTree: &calculation.Tree{},
				UserID:     1,
			},
			wantErr: false,
		},
		{
			name: "update existing expression",
			expression: &models.Expression{
				ID:         1,
				Status:     "completed",
				Result:     42.0,
				BinaryTree: &calculation.Tree{},
				UserID:     1,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := storage.SaveExpression(tt.expression)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			newExpression, err := storage.GetExpression(tt.expression.ID)
			require.NoError(t, err)
			require.Equal(t, *tt.expression, newExpression)
		})
	}
}

func TestSaveAndGetTask(t *testing.T) {
	storage := NewStorage(true)
	defer storage.Close()

	tests := []struct {
		name    string
		task    *models.Task
		wantErr bool
	}{
		{
			name:    "create new task",
			task:    &models.Task{},
			wantErr: false,
		},
		{
			name: "update existing task",
			task: &models.Task{
				ID:     1,
				Status: "аляулу",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := storage.SaveTask(tt.task)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			newTask, err := storage.GetTask(tt.task.ID)
			require.NoError(t, err)
			require.Equal(t, *tt.task, newTask)
		})
	}
}

func TestGetExpressions(t *testing.T) {
	storage := NewStorage(true)
	defer storage.Close()

	user_id := 1

	tests := []struct {
		name        string
		expressions []*models.Expression
		wantErr     bool
	}{
		{
			name:        "load multiple expressions",
			expressions: []*models.Expression{{UserID: user_id}, {UserID: user_id}},
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, expression := range tt.expressions {
				_, err := storage.SaveExpression(expression)
				require.NoError(t, err)
			}
			expressions, err := storage.GetExpressions(user_id)
			require.NoError(t, err)
			require.Equal(t, len(tt.expressions), len(expressions))
		})
	}
}

func TestGetTasks(t *testing.T) {
	storage := NewStorage(true)
	defer storage.Close()

	tests := []struct {
		name    string
		tasks   []*models.Task
		wantErr bool
	}{
		{
			name:    "load multiple tasks",
			tasks:   []*models.Task{{}, {}},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, task := range tt.tasks {
				_, err := storage.SaveTask(task)
				require.NoError(t, err)
			}
			tasks := storage.GetTasks()
			require.Equal(t, len(tt.tasks), len(tasks))
		})
	}
}

func TestGetPendingTask(t *testing.T) {
	storage := NewStorage(true)
	defer storage.Close()

	tests := []struct {
		name    string
		tasks   []*models.Task
		wantErr bool
	}{
		{
			name:    "load multiple tasks",
			tasks:   []*models.Task{{Status: "pending"}, {}},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, task := range tt.tasks {
				_, err := storage.SaveTask(task)
				require.NoError(t, err)
			}
			task, err := storage.GetPendingTask()
			require.NoError(t, err)
			require.Equal(t, "pending", task.Status)
		})
	}
}

func TestDeleteTaskByExpressionID(t *testing.T) {
	storage := NewStorage(true)
	defer storage.Close()

	expression_id := 1

	tests := []struct {
		name    string
		tasks   []*models.Task
		wantErr bool
	}{
		{
			name:    "load multiple tasks",
			tasks:   []*models.Task{{ExpressionID: 1}, {ExpressionID: 2}},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, task := range tt.tasks {
				_, err := storage.SaveTask(task)
				require.NoError(t, err)
			}
			err := storage.DeleteTaskByExpressionID(expression_id)
			require.NoError(t, err)
			tasks := storage.GetTasks()
			require.Equal(t, len(tt.tasks)-1, len(tasks))
		})
	}
}
