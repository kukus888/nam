package data

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DashboardData represents a complete dashboard view with all necessary information
type DashboardData struct {
	Applications []DashboardApplication `json:"applications"`
	Summary      DashboardSummary       `json:"summary"`
}

// DashboardApplication represents an application with all its instances and health status
type DashboardApplication struct {
	Id               uint                `json:"id" db:"application_definition_id"`
	Name             string              `json:"name" db:"application_definition_name"`
	Type             string              `json:"type" db:"application_definition_type"`
	Port             int                 `json:"port" db:"application_definition_port"`
	HealthyCount     int                 `json:"healthy_count"`
	UnhealthyCount   int                 `json:"unhealthy_count"`
	MaintenanceCount int                 `json:"maintenance_count"`
	TotalCount       int                 `json:"total_count"`
	HealthStatus     string              `json:"health_status"` // healthy, degraded, unhealthy, unknown
	Instances        []DashboardInstance `json:"instances"`
}

// DashboardInstance represents an instance with its current health status
type DashboardInstance struct {
	Id              uint       `json:"id" db:"application_instance_id"`
	Name            string     `json:"name" db:"application_instance_name"`
	ServerHostname  string     `json:"server_hostname" db:"server_hostname"`
	ServerAlias     string     `json:"server_alias" db:"server_alias"`
	MaintenanceMode bool       `json:"maintenance_mode" db:"maintenance_mode"`
	IsHealthy       *bool      `json:"is_healthy" db:"is_successful"`
	LastCheckTime   *time.Time `json:"last_check_time" db:"time_end"`
	ResponseTime    *int       `json:"response_time" db:"res_time"`
	ErrorMessage    *string    `json:"error_message" db:"error_message"`
	HasHealthcheck  bool       `json:"has_healthcheck"`
}

// DashboardSummary provides overall system statistics
type DashboardSummary struct {
	TotalApplications     int `json:"total_applications"`
	TotalInstances        int `json:"total_instances"`
	HealthyInstances      int `json:"healthy_instances"`
	UnhealthyInstances    int `json:"unhealthy_instances"`
	MaintenanceInstances  int `json:"maintenance_instances"`
	UnknownInstances      int `json:"unknown_instances"`
	HealthyApplications   int `json:"healthy_applications"`
	DegradedApplications  int `json:"degraded_applications"`
	UnhealthyApplications int `json:"unhealthy_applications"`
}

// GetDashboardData retrieves all dashboard data in a single efficient query
func GetDashboardData(pool *pgxpool.Pool) (*DashboardData, error) {
	tx, err := pool.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(context.Background())

	// Single query to get all application definitions with their instances and latest health results
	rows, err := tx.Query(context.Background(), `
		SELECT 
			ad.id AS application_definition_id,
			ad.name AS application_definition_name,
			ad.type AS application_definition_type,
			ad.port AS application_definition_port,
			ai.id AS application_instance_id,
			ai.name AS application_instance_name,
			ai.maintenance_mode,
			s.hostname AS server_hostname,
			s.alias AS server_alias,
			hr.is_successful,
			hr.time_end,
			hr.res_time,
			hr.error_message,
			CASE WHEN ad.healthcheck_id IS NOT NULL THEN true ELSE false END AS has_healthcheck
		FROM application_definition ad
		LEFT JOIN application_instance ai ON ad.id = ai.application_definition_id
		LEFT JOIN "server" s ON ai.server_id = s.id
		LEFT JOIN LATERAL (
			SELECT is_successful, time_end, res_time, error_message
			FROM healthcheck_results hcr
			WHERE hcr.application_instance_id = ai.id
			ORDER BY hcr.time_end DESC
			LIMIT 1
		) hr ON ai.id IS NOT NULL
		ORDER BY ad.name, ai.name;
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Parse the results and organize them into the dashboard structure
	applicationMap := make(map[uint]*DashboardApplication)
	var applications []DashboardApplication

	for rows.Next() {
		var (
			appId           uint
			appName         string
			appType         string
			appPort         int
			instanceId      *uint
			instanceName    *string
			maintenanceMode *bool
			serverHostname  *string
			serverAlias     *string
			isSuccessful    *bool
			timeEnd         *time.Time
			resTime         *int
			errorMessage    *string
			hasHealthcheck  bool
		)

		err := rows.Scan(
			&appId, &appName, &appType, &appPort,
			&instanceId, &instanceName, &maintenanceMode,
			&serverHostname, &serverAlias,
			&isSuccessful, &timeEnd, &resTime, &errorMessage,
			&hasHealthcheck,
		)
		if err != nil {
			return nil, err
		}

		// Get or create application
		app, exists := applicationMap[appId]
		if !exists {
			app = &DashboardApplication{
				Id:        appId,
				Name:      appName,
				Type:      appType,
				Port:      appPort,
				Instances: []DashboardInstance{},
			}
			applicationMap[appId] = app
		}

		// Add instance if it exists
		if instanceId != nil {
			instance := DashboardInstance{
				Id:              *instanceId,
				Name:            *instanceName,
				ServerHostname:  *serverHostname,
				ServerAlias:     *serverAlias,
				MaintenanceMode: *maintenanceMode,
				IsHealthy:       isSuccessful,
				LastCheckTime:   timeEnd,
				ResponseTime:    resTime,
				ErrorMessage:    errorMessage,
				HasHealthcheck:  hasHealthcheck,
			}
			app.Instances = append(app.Instances, instance)
		}
	}

	// Convert map to slice
	for _, app := range applicationMap {
		applications = append(applications, *app)
	}

	// Calculate application-level health status and counts
	summary := DashboardSummary{}
	for i := range applications {
		app := &applications[i]
		app.TotalCount = len(app.Instances)
		summary.TotalInstances += app.TotalCount

		for _, instance := range app.Instances {
			if instance.MaintenanceMode {
				app.MaintenanceCount++
				summary.MaintenanceInstances++
			} else if instance.IsHealthy == nil {
				summary.UnknownInstances++
			} else if *instance.IsHealthy {
				app.HealthyCount++
				summary.HealthyInstances++
			} else {
				app.UnhealthyCount++
				summary.UnhealthyInstances++
			}
		}

		// Determine application health status
		if app.TotalCount == 0 {
			app.HealthStatus = "unknown"
		} else if app.HealthyCount == app.TotalCount-app.MaintenanceCount {
			app.HealthStatus = "healthy"
			summary.HealthyApplications++
		} else if app.HealthyCount > 0 {
			app.HealthStatus = "degraded"
			summary.DegradedApplications++
		} else {
			app.HealthStatus = "unhealthy"
			summary.UnhealthyApplications++
		}
	}

	summary.TotalApplications = len(applications)

	err = tx.Commit(context.Background())
	if err != nil {
		return nil, err
	}

	return &DashboardData{
		Applications: applications,
		Summary:      summary,
	}, nil
}
