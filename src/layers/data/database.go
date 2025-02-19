package data

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

// Initializes new pgx database connection with provided connection string
func NewDatabase(dsn string) (*Database, error) {
	p, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, err
	}
	return &Database{Pool: p}, nil
}

// A type set of what types are permitted to be used with DB (used to permit generics)
type IDatabaseQueryable interface {
	DbInsert(pgx.Tx) (*uint, error) // Inserts object into DB. Returns the object's new ID, or an error
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
func DbQueryTypeAll[T IDatabaseQueryable](tx pgx.Tx, typ T) ([]T, error) {
	tableName := StructToTableName(fmt.Sprintf("%T", typ))
	query := fmt.Sprintf(`SELECT * FROM "%s"`, tableName)
	rows, err := tx.Query(context.Background(), query)
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

func DbQueryTypeWithParams[T IDatabaseQueryable](tx pgx.Tx, typ T, filters ...DbFilter) ([]T, error) {
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
	rows, err := tx.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	res, err := pgx.CollectRows(rows, pgx.RowToStructByName[T])
	if err != nil {
		return nil, err
	}
	return res, nil
}
