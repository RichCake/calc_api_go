package expression

import (
	"testing"

	"github.com/RichCake/calc_api_go/orchestrator/internal/config"
	"github.com/RichCake/calc_api_go/orchestrator/internal/models"
	"github.com/RichCake/calc_api_go/orchestrator/internal/storage"
	"github.com/stretchr/testify/require"
)

func setUpService() *ExpressionService {
	tc := config.TimeConfig{}
	storage := storage.NewStorage(true)
	service := NewExpressionService(storage, tc)
	return service
}

func TestService(t *testing.T) {
	service := setUpService()

	user_id := 1
	service.storage.SaveUser(&models.User{ID: user_id, Login: "test"})

	tests := []struct {
		name           string
		expression_str string
		expected_task  models.Task
		result         float64
		wantErr        bool
	}{
		{
			name:           "simple expression",
			expression_str: "2 + 2",
			expected_task: models.Task{
				ID:            1,
				Arg1:          2.0,
				Arg2:          2.0,
				ExpressionID:  1,
				Status:        "in progress",
				Operation:     "+",
				OperationTime: 0,
			},
			result:  4,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expression_id, err := service.ProcessExpression(tt.expression_str, user_id)
			require.NoError(t, err)
			newTask, err := service.GetPendingTask()
			require.NoError(t, err)
			require.Equal(t, tt.expected_task, newTask)

			err = service.ProcessIncomingTask(newTask.ID, tt.result)
			require.NoError(t, err)

			newExpression, err := service.GetExpressionByID(expression_id, user_id)
			require.NoError(t, err)
			require.Equal(t, tt.result, newExpression.Result)
		})
	}
}
