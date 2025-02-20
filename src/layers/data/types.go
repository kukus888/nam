package data

import (
	"context"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
)

type Proxy struct {
	ID      uint
	Ingress uint
	Egress  TopologyNode
}

type F5 struct {
	ID      uint
	Ingress TopologyNode
}

type F5Egress struct {
	ID     uint
	Egress TopologyNode
}

type Nginx struct {
	ID      uint
	Ingress TopologyNode
}

type NginxEgress struct {
	ID     uint
	Egress TopologyNode
}

// HealthcheckDAO represents a healthcheck element in the topology yaml file
// This is called when a healthcheck is needed
type HealthcheckDAO struct {
	ID             uint          `json:"id"`
	Url            string        `json:"url"`
	Timeout        time.Duration `json:"timeout"`
	CheckInterval  time.Duration `json:"check_interval"`
	ExpectedStatus int           `json:"expected_status"`
}

func (s HealthcheckDAO) TableName() string {
	return "healthcheck"
}

func (s HealthcheckDAO) ApiName() string {
	return "healthchecks"
}

// Inserts Healthcheck into Database. New ID is stored in the referenced Healthcheck struct.
// Does not roll back transaction, this is merely a facade for an insert statement
func (s HealthcheckDAO) DbInsert(tx pgx.Tx) (*uint, error) {
	var id uint
	err := tx.QueryRow(context.Background(), "INSERT INTO healthcheck (url, timeout, check_interval, expected_status) VALUES ($1, $2, $3, $4) RETURNING id", s.Url, s.Timeout, s.CheckInterval, s.ExpectedStatus).Scan(&id)
	return &id, err
}

// Gets all ApplicationDefinition objects using this healthcheck
func (hc *HealthcheckDAO) GetUsingApplicationDefinitions(tx pgx.Tx) ([]ApplicationDefinitionDAO, error) {
	idstr := strconv.Itoa(int(hc.ID))
	return DbQueryTypeWithParams(tx, ApplicationDefinitionDAO{}, DbFilter{
		Column:   "healthcheck_id",
		Operator: DbOperatorEqual,
		Value:    idstr,
	})
}
