package data

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func (user User) DbInsert(pool *pgxpool.Pool) (*uint64, error) {
	tx, err := pool.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(context.Background())

	err = tx.QueryRow(context.Background(), `
		INSERT INTO "user" (username, password_hash) 
		VALUES ($1, $2) RETURNING id;
	`, user.Username, user.PasswordHash).Scan(&user.Id)
	if err != nil {
		return nil, err
	}
	return &user.Id, tx.Commit(context.Background())
}

func GetUserByUsername(pool *pgxpool.Pool, username string) (*User, error) {
	tx, err := pool.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(context.Background())

	rows, err := tx.Query(context.Background(), `
		SELECT * FROM "user" WHERE username = $1;
	`, username)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil // No user found
	} else if err != nil {
		return nil, err
	}

	user, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByNameLax[User])
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil // No user found
	} else if err != nil {
		return nil, err
	}

	return &user, tx.Commit(context.Background())
}

func GetUserCount(pool *pgxpool.Pool) (int, error) {
	tx, err := pool.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(context.Background())

	var count int
	err = tx.QueryRow(context.Background(), `
		SELECT COUNT(*) FROM "user";
	`).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, tx.Commit(context.Background())
}

func (user User) HasRole(role string) bool {
	return true // Placeholder for role checking logic
	// TODO: Implement role checking logic based on the user's roles
}
