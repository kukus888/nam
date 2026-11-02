//go:build integration

package data

import (
	"context"
	"testing"
)

func TestGetHealthcheckLatestResultByApplicationDefinitionIdHandlesMaintenanceWithoutResults(t *testing.T) {
	pool := setupIntegrationDB(t)
	ctx := context.Background()

	var topologyNodeID, serverID, definitionID int
	if err := pool.QueryRow(ctx, `INSERT INTO topology_node (type) VALUES ('application_instance') RETURNING id`).Scan(&topologyNodeID); err != nil {
		t.Fatalf("failed to create topology node: %v", err)
	}
	if err := pool.QueryRow(ctx, `INSERT INTO "server" (hostname) VALUES ('maintenance-test') RETURNING id`).Scan(&serverID); err != nil {
		t.Fatalf("failed to create server: %v", err)
	}
	if err := pool.QueryRow(ctx, `INSERT INTO application_definition (name, type) VALUES ('maintenance-test', 'test') RETURNING id`).Scan(&definitionID); err != nil {
		t.Fatalf("failed to create application definition: %v", err)
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO application_instance (topology_node_id, name, server_id, application_definition_id, maintenance_mode)
		VALUES ($1, 'maintenance-test', $2, $3, TRUE)
	`, topologyNodeID, serverID, definitionID); err != nil {
		t.Fatalf("failed to create application instance: %v", err)
	}

	results, err := GetHealthcheckLatestResultByApplicationDefinitionId(pool, uint64(definitionID))
	if err != nil {
		t.Fatalf("GetHealthcheckLatestResultByApplicationDefinitionId() returned error: %v", err)
	}
	if len(*results) != 1 {
		t.Fatalf("expected one instance result, got %d", len(*results))
	}
	if !(*results)[0].IsMaintenance {
		t.Fatal("expected instance to be marked as in maintenance")
	}
	if (*results)[0].IsSuccessful {
		t.Fatal("expected instance without a healthcheck result to be unsuccessful")
	}
}
