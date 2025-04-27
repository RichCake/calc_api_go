package storage

import (
	"context"
	"database/sql"
	"errors"

	"github.com/RichCake/calc_api_go/orchestrator/internal/models"

	sqlite "github.com/mattn/go-sqlite3"
)

func (s *Storage) SaveUser(user *models.User) (int, error) {
	ctx := context.TODO()

	if user.ID == 0 {
		q := `
		INSERT INTO users (login, password)
		VALUES ($1, $2)
		`
		res, err := s.db.ExecContext(ctx, q, user.Login, user.Password)
		if errors.Is(err, sqlite.ErrConstraintUnique) {
			return 0, ErrUsernameTaken
		} else if err != nil {
			return 0, err
		}
		lastID, err := res.LastInsertId()
		if err != nil {
			return 0, err
		}
		user.ID = int(lastID)
		return int(lastID), nil
	}

	q := `
	UPDATE users
	SET login = $1, password = $2
	WHERE user_id = $3
	`
	_, err := s.db.ExecContext(ctx, q, user.Login, user.Password, user.ID)
	if err != nil {
		return 0, err
	}
	return user.ID, nil
}

func (s *Storage) GetUser(user_id int) (models.User, error) {
	var user models.User
	var q = `
	SELECT user_id, login, password 
	FROM users
	WHERE user_id = $1
	`
	ctx := context.TODO()
	err := s.db.QueryRowContext(ctx, q, user_id).Scan(&user.ID, &user.Login, &user.Password)
	if errors.Is(err, sql.ErrNoRows) {
		return user, ErrItemNotFound
	} else if err != nil {
		return user, err
	}
	return user, nil
}
