package data

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Assigns new role to the user, rewriting the old one
func AssignRoleIdToUser(pool *pgxpool.Pool, userId, roleId uint64) error {
	tx, err := pool.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())

	// Check if the user doesnt have any roles assigned
	var existingRoleId uint64
	err = tx.QueryRow(context.Background(), `
		SELECT role_id FROM user_role_mapping WHERE user_id = $1;
	`, userId).Scan(&existingRoleId)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return err
	}

	if errors.Is(err, pgx.ErrNoRows) {
		_, err = tx.Exec(context.Background(), `
			INSERT INTO user_role_mapping (user_id, role_id) 
			VALUES ($1, $2);
		`, userId, roleId)
		if err != nil {
			return err
		}
	} else {
		// Overwrite the role mapping
		_, err = tx.Exec(context.Background(), `
			UPDATE user_role_mapping SET role_id = $1 WHERE user_id = $2;
		`, roleId, userId)
		if err != nil {
			return err
		}
	}
	return tx.Commit(context.Background())
}

func GetAllRoles(pool *pgxpool.Pool) ([]Role, error) {
	tx, err := pool.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(context.Background())
	rows, err := tx.Query(context.Background(), "SELECT id, name, color, description FROM role")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []Role
	for rows.Next() {
		var role Role
		if err := rows.Scan(&role.Id, &role.Name, &role.Color, &role.Description); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return roles, nil
}

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
