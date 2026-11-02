package data

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func (s ServiceOfferingDAO) TableName() string {
	return "service_offering"
}

func (s ServiceOfferingDAO) ApiName() string {
	return "offerings"
}

func GetServiceOfferingsAll(pool *pgxpool.Pool) (*[]ServiceOfferingListItem, error) {
	tx, err := pool.Begin(context.Background())
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(context.Background())

	var offerings []ServiceOfferingListItem
	err = pgxscan.Select(context.Background(), tx, &offerings, `
		SELECT
			so.id,
			so.name,
			so.description,
			so.created_by_user_id,
			so.updated_by_user_id,
			so.created_at,
			so.updated_at,
			COUNT(soad.application_definition_id)::INT AS member_count
		FROM service_offering so
		LEFT JOIN service_offering_application_definition soad
			ON soad.service_offering_id = so.id
		GROUP BY so.id
		ORDER BY so.name ASC
	`)
	if err != nil {
		return nil, err
	}

	return &offerings, tx.Commit(context.Background())
}

func GetServiceOfferingById(pool *pgxpool.Pool, id uint64) (*ServiceOfferingDAO, error) {
	if id == 0 {
		return nil, errors.New("id is required")
	}
	tx, err := pool.Begin(context.Background())
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(context.Background())

	var offering ServiceOfferingDAO
	err = pgxscan.Get(context.Background(), tx, &offering, `
		SELECT
			id,
			name,
			description,
			created_by_user_id,
			updated_by_user_id,
			created_at,
			updated_at
		FROM service_offering
		WHERE id = $1
	`, id)
	if err != nil {
		return nil, err
	}

	return &offering, tx.Commit(context.Background())
}

func (offering ServiceOfferingDAO) DbInsert(pool *pgxpool.Pool) (*uint64, error) {
	tx, err := pool.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		return nil, err
	}

	var insertedID uint64
	err = tx.QueryRow(context.Background(), `
		INSERT INTO service_offering (name, description, created_by_user_id, updated_by_user_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, offering.Name, offering.Description, offering.CreatedByUserID, offering.UpdatedByUserID).Scan(&insertedID)
	if err != nil {
		tx.Rollback(context.Background())
		return nil, err
	}

	return &insertedID, tx.Commit(context.Background())
}

func UpdateServiceOffering(pool *pgxpool.Pool, offering *ServiceOfferingDAO) error {
	if offering.Id == 0 {
		return errors.New("id is required for update")
	}

	tx, err := pool.Begin(context.Background())
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())

	_, err = tx.Exec(context.Background(), `
		UPDATE service_offering
		SET
			name = $1,
			description = $2,
			updated_by_user_id = $3,
			updated_at = NOW()
		WHERE id = $4
	`, offering.Name, offering.Description, offering.UpdatedByUserID, offering.Id)
	if err != nil {
		return err
	}

	return tx.Commit(context.Background())
}

func DeleteServiceOfferingById(pool *pgxpool.Pool, id uint64) error {
	tx, err := pool.Begin(context.Background())
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())

	_, err = tx.Exec(context.Background(), `DELETE FROM service_offering WHERE id = $1`, id)
	if err != nil {
		return err
	}

	return tx.Commit(context.Background())
}

func GetServiceOfferingMembers(pool *pgxpool.Pool, offeringID uint64) (*[]ServiceOfferingMember, error) {
	tx, err := pool.Begin(context.Background())
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(context.Background())

	var members []ServiceOfferingMember
	err = pgxscan.Select(context.Background(), tx, &members, `
		SELECT
			soad.service_offering_id,
			soad.application_definition_id,
			ad.name AS application_name,
			ad.type AS application_type
		FROM service_offering_application_definition soad
		JOIN application_definition ad ON ad.id = soad.application_definition_id
		WHERE soad.service_offering_id = $1
		ORDER BY ad.name ASC
	`, offeringID)
	if err != nil {
		return nil, err
	}

	return &members, tx.Commit(context.Background())
}

func ReplaceServiceOfferingMembers(pool *pgxpool.Pool, offeringID uint64, appDefinitionIDs []uint64) error {
	tx, err := pool.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())

	_, err = tx.Exec(context.Background(), `
		DELETE FROM service_offering_application_definition
		WHERE service_offering_id = $1
	`, offeringID)
	if err != nil {
		return err
	}

	for _, appDefinitionID := range appDefinitionIDs {
		_, err = tx.Exec(context.Background(), `
			INSERT INTO service_offering_application_definition (service_offering_id, application_definition_id)
			VALUES ($1, $2)
			ON CONFLICT (service_offering_id, application_definition_id) DO NOTHING
		`, offeringID, appDefinitionID)
		if err != nil {
			return err
		}
	}

	return tx.Commit(context.Background())
}

func AddServiceOfferingMember(pool *pgxpool.Pool, offeringID uint64, appDefinitionID uint64) error {
	_, err := pool.Exec(context.Background(), `
		INSERT INTO service_offering_application_definition (service_offering_id, application_definition_id)
		VALUES ($1, $2)
		ON CONFLICT (service_offering_id, application_definition_id) DO NOTHING
	`, offeringID, appDefinitionID)
	return err
}

func RemoveServiceOfferingMember(pool *pgxpool.Pool, offeringID uint64, appDefinitionID uint64) error {
	_, err := pool.Exec(context.Background(), `
		DELETE FROM service_offering_application_definition
		WHERE service_offering_id = $1 AND application_definition_id = $2
	`, offeringID, appDefinitionID)
	return err
}

func PreviewServiceOfferingAutoGroup(pool *pgxpool.Pool, ruleType string, pattern string) (*[]ServiceOfferingAutoGroupMatch, error) {
	var appDefs []ApplicationDefinitionDAO
	err := pgxscan.Select(context.Background(), pool, &appDefs, `
		SELECT id, name, port, type, healthcheck_id
		FROM application_definition
		ORDER BY name ASC
	`)
	if err != nil {
		return nil, err
	}

	matches := make([]ServiceOfferingAutoGroupMatch, 0)
	for _, appDef := range appDefs {
		isMatch, err := appDefinitionNameMatchesRule(appDef.Name, ruleType, pattern)
		if err != nil {
			return nil, err
		}
		if !isMatch {
			continue
		}

		offeringNames, err := getOfferingNamesByApplicationDefinitionID(pool, uint64(appDef.Id))
		if err != nil {
			return nil, err
		}

		matches = append(matches, ServiceOfferingAutoGroupMatch{
			ApplicationDefinitionID: uint64(appDef.Id),
			ApplicationName:         appDef.Name,
			ExistingOfferings:       offeringNames,
		})
	}

	return &matches, nil
}

func appDefinitionNameMatchesRule(name string, ruleType string, pattern string) (bool, error) {
	switch ruleType {
	case "prefix":
		return len(pattern) > 0 && len(name) >= len(pattern) && name[:len(pattern)] == pattern, nil
	case "regex":
		r, err := regexp.Compile(pattern)
		if err != nil {
			return false, fmt.Errorf("invalid regex pattern: %w", err)
		}
		return r.MatchString(name), nil
	default:
		return false, errors.New("unsupported rule type, expected 'prefix' or 'regex'")
	}
}

func getOfferingNamesByApplicationDefinitionID(pool *pgxpool.Pool, appDefinitionID uint64) ([]string, error) {
	type offeringNameRow struct {
		Name string `db:"name"`
	}
	rows := make([]offeringNameRow, 0)
	err := pgxscan.Select(context.Background(), pool, &rows, `
		SELECT so.name
		FROM service_offering_application_definition soad
		JOIN service_offering so ON so.id = soad.service_offering_id
		WHERE soad.application_definition_id = $1
		ORDER BY so.name ASC
	`, appDefinitionID)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(rows))
	for _, row := range rows {
		names = append(names, row.Name)
	}
	return names, nil
}

func InsertServiceOfferingAuditEvent(pool *pgxpool.Pool, serviceOfferingID *uint64, action string, details map[string]any, createdByUserID *uint64) error {
	var detailsJSON []byte
	if details != nil {
		encoded, err := json.Marshal(details)
		if err != nil {
			return err
		}
		detailsJSON = encoded
	}

	_, err := pool.Exec(context.Background(), `
		INSERT INTO service_offering_audit (service_offering_id, action, details, created_by_user_id)
		VALUES ($1, $2, $3, $4)
	`, serviceOfferingID, action, detailsJSON, createdByUserID)
	return err
}
