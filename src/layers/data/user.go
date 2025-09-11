package data

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateUser(pool *pgxpool.Pool, user UserDTO, roleId uint64) (*uint64, error) {
	if user.Username == "" || user.Email == "" || user.Password == "" {
		return nil, fmt.Errorf("username, email and password are required")
	}
	hashedPassword, err := HashPassword(user.Password)
	if err != nil {
		return nil, fmt.Errorf("error hashing password: %w", err)
	}
	// Create user in the database
	u := User{
		Username:     user.Username,
		Email:        user.Email,
		PasswordHash: hashedPassword,
		RoleId:       roleId,
	}
	id, err := u.DbInsert(pool)
	if err != nil {
		return nil, fmt.Errorf("error creating user: %w", err)
	}
	// Update roles
	return id, nil
}

// Gets all users, and joins with the permissions table
func GetAllUsersFull(pool *pgxpool.Pool) (*[]UserFull, error) {
	tx, err := pool.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(context.Background())

	rows, err := tx.Query(context.Background(), `
		select u.id as user_id, u.username, u.email, r.id as role_id, r.name as role_name, r.color as role_color, r.description as role_description
		from "user" u inner join "role" r on u.role_id = r.id;
	`)
	if err != nil {
		return nil, err
	}

	users, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[UserFull])
	if err != nil {
		return nil, err
	}

	return &users, tx.Commit(context.Background())
}

// Helper insert function. Use CreateUser instead
func (user User) DbInsert(pool *pgxpool.Pool) (*uint64, error) {
	tx, err := pool.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(context.Background())

	err = tx.QueryRow(context.Background(), `
		INSERT INTO "user" (username, password_hash, email, role_id) 
		VALUES ($1, $2, $3, $4) RETURNING id;
	`, user.Username, user.PasswordHash, user.Email, user.RoleId).Scan(&user.Id)
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

func GetUserById(pool *pgxpool.Pool, id uint64) (*User, error) {
	tx, err := pool.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(context.Background())

	rows, err := tx.Query(context.Background(), `
		SELECT * FROM "user" WHERE id = $1;
	`, id)
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

// Removes this user from the database
func (user User) Delete(pool *pgxpool.Pool) error {
	tx, err := pool.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())

	_, err = tx.Exec(context.Background(), `
		DELETE FROM "user" WHERE id = $1;
	`, user.Id)
	if err != nil {
		return err
	}

	return tx.Commit(context.Background())
}

// Updates all the users information, except password, in database
func (user User) UpdateWithoutPassword(pool *pgxpool.Pool) error {
	tx, err := pool.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())

	_, err = tx.Exec(context.Background(), `
		UPDATE "user" SET username = $1, email = $2, role_id = $3 WHERE id = $4;
	`, user.Username, user.Email, user.RoleId, user.Id)
	if err != nil {
		return err
	}

	return tx.Commit(context.Background())
}

func (user User) UpdatePassword(pool *pgxpool.Pool, newPassword string) error {
	hash, err := HashPassword(newPassword)
	if err != nil {
		return err
	}

	tx, err := pool.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())

	_, err = tx.Exec(context.Background(), `
		UPDATE "user" SET password_hash = $1 WHERE id = $2;
	`, hash, user.Id)
	if err != nil {
		return err
	}

	return tx.Commit(context.Background())
}
