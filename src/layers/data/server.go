package data

import (
	"context"
	"strconv"

	"github.com/jackc/pgx/v5"
)

type ServerDAO struct {
	ID       uint   `json:"id"`
	Alias    string `json:"alias"`
	Hostname string `json:"hostname"`
}

func (s ServerDAO) TableName() string {
	return "server"
}

func (s ServerDAO) ApiName() string {
	return "server"
}

// Inserts Server into Database. New ID is stored in the referenced Server struct.
// Does not roll back transaction, this is merely a facade for an insert statement
func (s ServerDAO) DbInsert(tx pgx.Tx) (*uint, error) {
	var id uint
	err := tx.QueryRow(context.Background(), "INSERT INTO server (alias, hostname) VALUES ($1, $2) RETURNING id", s.Alias, s.Hostname).Scan(&id)
	return &id, err
}

// Deletes specified ServerDAO
func (s ServerDAO) Delete(tx pgx.Tx) (*int, error) {
	var ra = 0
	com, err := tx.Exec(context.Background(), "DELETE FROM server WHERE id = $1", s.ID)
	if err != nil {
		tx.Rollback(context.Background())
		return nil, err
	}
	ra += int(com.RowsAffected())
	return &ra, nil
}

func (s *ServerDAO) GetUsingApplicationInstances(tx pgx.Tx) ([]ApplicationInstanceDAO, error) {
	idstr := strconv.Itoa(int(s.ID))
	return DbQueryTypeWithParams(tx, ApplicationInstanceDAO{}, DbFilter{
		Column:   "server_id",
		Operator: DbOperatorEqual,
		Value:    idstr,
	})
}
