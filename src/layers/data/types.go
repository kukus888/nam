package data

import (
	"context"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
)

// Topology represents a Topology element in the topology yaml file
type TopologyNode struct {
	ID   uint
	Type string // Type of Topology element, e.g. loadbalancer, firewall, etc. \n The Type corresponds to the table name in the database
}

// Inserts TopologyNode into Database. New ID is stored in the referenced TopologyNode
// Does not roll back transaction, this is merely a facade for an insert statement
func (tn TopologyNode) DbInsert(tx pgx.Tx) (*uint, error) {
	var id uint
	err := tx.QueryRow(context.Background(), "INSERT INTO topology_node (type) VALUES ($1) RETURNING id", tn.Type).Scan(&id)
	return &id, err
}

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

// ApplicationDefinition represents the definition of an application and its general properties
type ApplicationDefinition struct {
	ID            uint
	Name          string
	Port          int
	Type          string
	HealthcheckId *uint `db:"healthcheck_id"`
}

// Creates new ApplicationDefinition in DB
// Returns the inserted ApplicationDefinition object
func (appDef ApplicationDefinition) DbInsert(tx pgx.Tx) (*uint, error) {
	var resId uint
	var err error
	if appDef.HealthcheckId == nil {
		err = tx.QueryRow(context.Background(), "INSERT INTO application_definition (name, port, type) VALUES ($1, $2, $3) RETURNING id", appDef.Name, appDef.Port, appDef.Type).Scan(&resId)
	} else {
		err = tx.QueryRow(context.Background(), "INSERT INTO application_definition (name, port, type, healthcheck_id) VALUES ($1, $2, $3, $4) RETURNING id", appDef.Name, appDef.Port, appDef.Type, appDef.HealthcheckId).Scan(&resId)
	}
	if err != nil {
		tx.Rollback(context.Background())
		return nil, err
	}
	return &resId, nil
}

// ApplicationInstance represents an instance of an application
type ApplicationInstance struct {
	ID           uint
	ServerId     uint `db:"server_id"`
	DefinitionId uint `db:"application_definition_id"`
	Name         string
}

// Creates new ApplicationInstance in DB, with underlying TopologyNode struct.
// Returns the inserted ApplicationInstance object
func (dto ApplicationInstance) DbInsert(tx pgx.Tx) (*uint, error) {
	// Create instance name first
	// Create underlying topologyNode
	tn := TopologyNode{Type: "application_instance"}
	tnId, err := tn.DbInsert(tx)
	if err != nil {
		tx.Rollback(context.Background())
		return nil, err
	}
	// Insert instance into DB
	dto.ID = *tnId
	var resId uint
	err = tx.QueryRow(context.Background(), "INSERT INTO application_instance (id, name, server_id, application_definition_id) VALUES ($1, $2, $3, $4) RETURNING id", dto.ID, dto.Name, dto.ServerId, dto.DefinitionId).Scan(&resId)
	if err != nil {
		tx.Rollback(context.Background())
		return nil, err
	}
	return &resId, nil
}

type Server struct {
	ID       uint
	Alias    string
	Hostname string
}

// Inserts Server into Database. New ID is stored in the referenced Server struct.
// Does not roll back transaction, this is merely a facade for an insert statement
func (s Server) DbInsert(tx pgx.Tx) (*uint, error) {
	var id uint
	err := tx.QueryRow(context.Background(), "INSERT INTO server (alias, hostname) VALUES ($1, $2) RETURNING id", s.Alias, s.Hostname).Scan(&id)
	return &id, err
}

func (s *Server) GetUsingApplicationInstances(tx pgx.Tx) ([]ApplicationInstance, error) {
	idstr := strconv.Itoa(int(s.ID))
	return DbQueryTypeWithParams(tx, ApplicationInstance{}, DbFilter{
		Column:   "server_id",
		Operator: DbOperatorEqual,
		Value:    idstr,
	})
}

// Healthcheck represents a healthcheck element in the topology yaml file
// This is called when a healthcheck is needed
type Healthcheck struct {
	ID             uint
	Url            string
	Timeout        time.Duration
	Interval       time.Duration
	ExpectedStatus int
}

// Inserts Healthcheck into Database. New ID is stored in the referenced Healthcheck struct.
// Does not roll back transaction, this is merely a facade for an insert statement
func (s Healthcheck) DbInsert(tx pgx.Tx) (*uint, error) {
	var id uint
	err := tx.QueryRow(context.Background(), "INSERT INTO healthcheck (url, timeout, check_interval, expected_status) VALUES ($1, $2, $3, $4) RETURNING id", s.Url, s.Timeout, s.Interval, s.ExpectedStatus).Scan(&id)
	return &id, err
}

// Gets all ApplicationDefinition objects using this healthcheck
func (hc *Healthcheck) GetUsingApplicationDefinitions(tx pgx.Tx) ([]ApplicationDefinition, error) {
	idstr := strconv.Itoa(int(hc.ID))
	return DbQueryTypeWithParams(tx, ApplicationDefinition{}, DbFilter{
		Column:   "healthcheck_id",
		Operator: DbOperatorEqual,
		Value:    idstr,
	})
}
