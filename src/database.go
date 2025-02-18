package main

import (
	"context"
	"fmt"
	"strings"
	"unicode"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Database struct {
	Pool *pgxpool.Pool
}

// TODO: Impl DB context

// Golang sometimes sucks, or perhaps i should read about functional programming
var Db Database

// Loads database connection
func DbStart() {
	dsn := "postgres://postgres:heslo123@localhost:5432/postgres"
	p, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		panic(err)
	}
	Db.Pool = p
}

func DbQueryTopologyNodeAll() ([]TopologyNode, error) {
	return DbQueryTypeAll(TopologyNode{})
}

// Returns
func DbQueryTopologyNode(filter ...DbFilter) ([]TopologyNode, error) {
	return DbQueryTypeWithParams(TopologyNode{}, filter...)
}

// Gets all Server instances from database
func DbQueryServerAll() ([]Server, error) {
	return DbQueryTypeAll(Server{})
}

// Gets Server instances from database by ID
func DbQueryServerID(ID string) ([]Server, error) {
	rows, err := Db.Pool.Query(context.Background(), `SELECT * FROM server WHERE id = $1`, ID)
	if err != nil {
		return nil, err
	}
	servers, err := pgx.CollectRows(rows, pgx.RowToStructByName[Server])
	if err != nil {
		return nil, err
	}
	return servers, nil
}

// A type set of what types are permitted to be used with DB (used to permit generics)
type IDatabasable interface {
	TopologyNode | ApplicationDefinition | ApplicationInstance | Server | Healthcheck
}

// Idiotic conversion because postgres doesnt support capital letter in table and column name properly
// Converts type name to table name
// e.g.: ApplicationDefinition => application_definition
// TODO: Cacheable???
func StructToTableName(structName string) string {
	structArr := strings.Split(structName, ".")
	name := structArr[len(structArr)-1]
	outputName := ""
	outputName += strings.ToLower(string(name[0]))
	for index := 1; index < len(name); index++ {
		b := name[index]
		if unicode.IsUpper(rune(b)) {
			outputName += "_" + strings.ToLower(string(b))
		} else {
			outputName += string(b)
		}
	}
	return outputName
}

// Gets all instances of that type from database
func DbQueryTypeAll[T IDatabasable](typ T) ([]T, error) {
	tableName := StructToTableName(fmt.Sprintf("%T", typ))
	query := fmt.Sprintf(`SELECT * FROM "%s"`, tableName)
	rows, err := Db.Pool.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	res, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[T])
	if err != nil {
		return nil, err
	}
	return res, nil
}

type DbOperator string

const (
	DbOperatorAnd                DbOperator = "AND"
	DbOperatorOr                 DbOperator = "OR"
	DbOperatorEqual              DbOperator = "="
	DbOperatorNotEqual           DbOperator = "!="
	DbOperatorGreaterThan        DbOperator = ">"
	DbOperatorLessThan           DbOperator = "<"
	DbOperatorGreaterThanOrEqual DbOperator = ">="
	DbOperatorLessThanOrEqual    DbOperator = "<="
)

// Try to implement injection-safe db filters
type DbFilter struct {
	Column   string
	Operator DbOperator
	Value    string
}

// Checks DbFilter for possible SQL injection attack. Returns error if there is a risk of injection
func (f *DbFilter) CheckInjection() error {
	invalidStrs := `'"'\`
	if strings.ContainsAny(f.Column, invalidStrs) {
		return fmt.Errorf("invalid character in DbFilter Column")
	} else if strings.ContainsAny(f.Value, invalidStrs) {
		return fmt.Errorf("invalid character in DbFilter Value")
	} else {
		return nil
	}
}

func DbQueryTypeWithParams[T IDatabasable](typ T, filters ...DbFilter) ([]T, error) {
	tableName := StructToTableName(fmt.Sprintf("%T", typ))
	query := fmt.Sprintf(`SELECT * FROM %s tab`, tableName)
	if len(filters) > 0 {
		query += " WHERE"
		for i, filter := range filters {
			err := filter.CheckInjection()
			if err != nil {
				return nil, err
			}
			if i > 0 && i < len(filters)-1 {
				query += " AND"
			}
			query += fmt.Sprintf(` tab."%s" %s %s`, filter.Column, filter.Operator, filter.Value)
		}
	}
	rows, err := Db.Pool.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	res, err := pgx.CollectRows(rows, pgx.RowToStructByName[T])
	if err != nil {
		return nil, err
	}
	return res, nil
}

// Creates new ApplicationDefinition in DB, with underlying Nodes
func (appDef ApplicationDefinition) DbInsert() error {
	ads := []ApplicationDefinition{
		appDef,
	}
	// Insert definition into DB
	copyCount, err := Db.Pool.CopyFrom(
		context.Background(),
		pgx.Identifier{"application_definition"},
		[]string{"id", "name", "port", "type", "healthcheck_id"},
		pgx.CopyFromSlice(len(ads), func(i int) ([]any, error) {
			return []any{ads[i].ID, ads[i].Name, ads[i].Port, ads[i].Type, nil}, nil
		}),
	)
	if err != nil {
		return err // TODO: Error handling, unique values, etc
	}
	fmt.Printf("Inserted %d rows\n", copyCount)
	return nil
}

// Inserts a healthcheck object into DB
func (hc Healthcheck) DbInsert() error {
	hcs := []Healthcheck{
		hc,
	}
	copyCount, err := Db.Pool.CopyFrom(
		context.Background(),
		pgx.Identifier{"application_instance"},
		[]string{"url", "timeout", "check_interval", "expected_status"},
		pgx.CopyFromSlice(len(hcs), func(i int) ([]any, error) {
			return []any{hcs[i].Url, hcs[i].Timeout, hcs[i].Interval, hcs[i].ExpectedStatus}, nil
		}),
	)
	if err != nil {
		return err // TODO: Error handling, unique values, etc
	}
	fmt.Printf("Inserted %d rows\n", copyCount)
	return nil
}
