package data

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func (role Role) DbInsert(pool *pgxpool.Pool) (*uint64, error) {
	tx, err := pool.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(context.Background())

	err = tx.QueryRow(context.Background(), `
		INSERT INTO " " (name, description) 
		VALUES ($1, $2) RETURNING id;
	`, role.Name, role.Description).Scan(&role.Id)
	if err != nil {
		return nil, err
	}
	return &role.Id, tx.Commit(context.Background())
}

func GetRoleById(pool *pgxpool.Pool, id int) (*Role, error) {
	tx, err := pool.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(context.Background())

	rows, err := tx.Query(context.Background(), `
		SELECT * FROM "role" WHERE id = $1;
	`, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil // No user found
	} else if err != nil {
		return nil, err
	}

	role, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByNameLax[Role])
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil // No role found
	} else if err != nil {
		return nil, err
	}

	return &role, tx.Commit(context.Background())
}

func GetRoleByName(pool *pgxpool.Pool, name string) (*Role, error) {
	tx, err := pool.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(context.Background())

	rows, err := tx.Query(context.Background(), `
		SELECT * FROM "role" WHERE name = $1;
	`, name)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil // No role found
	} else if err != nil {
		return nil, err
	}

	role, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByNameLax[Role])
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil // No role found
	} else if err != nil {
		return nil, err
	}

	return &role, tx.Commit(context.Background())
}

func GetUserRolesByUserId(pool *pgxpool.Pool, userId uint64) ([]Role, error) {
	tx, err := pool.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(context.Background())

	rows, err := tx.Query(context.Background(), `
		SELECT r.* FROM "role" r
		JOIN user_role_mapping urm ON r.id = urm.role_id
		WHERE urm.user_id = $1;
	`, userId)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil // No roles found
	} else if err != nil {
		return nil, err
	}

	var roles []Role
	for rows.Next() {
		var role Role
		if err := rows.Scan(&role.Id, &role.Name, &role.Description); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}

	return roles, tx.Commit(context.Background())
}
