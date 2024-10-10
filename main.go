package main

import (
	"database/sql"
	"fmt"
	"time"
)

type User struct {
	username    string
	lastLoginAt time.Time
}

type UserStore struct {
	db *sql.DB
}

func (u *UserStore) Upsert(username string) error {
	query := `
INSERT INTO users (
  username
) VALUES (
  ?
)
ON DUPLICATE KEY UPDATE
  login_count = login_count + 1;
`

	result, err := u.db.Exec(query, username)
	if err != nil {
		return fmt.Errorf("db.Exec failed for username %q", username)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to retrieve rows affected for upsert of username: %q", username)
	}

	if rowsAffected != 1 {
		return fmt.Errorf("unexpected rowsAffected: %d for upsert of username: %q", rowsAffected, username)
	}

	return nil
}
