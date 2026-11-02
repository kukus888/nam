package data

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OfferingDashboardData struct {
	Offerings   []OfferingDashboardItem  `json:"offerings"`
	Summary     OfferingDashboardSummary `json:"summary"`
	RefreshedAt time.Time                `json:"refreshed_at"`
}

type OfferingDashboardSummary struct {
	TotalOfferings              int `json:"total_offerings"`
	HealthyOfferings            int `json:"healthy_offerings"`
	ProblematicOfferings        int `json:"problematic_offerings"`
	DownOfferings               int `json:"down_offerings"`
	NoDataOfferings             int `json:"no_data_offerings"`
	OfferingsWithMaintenance    int `json:"offerings_with_maintenance"`
	TotalMemberDefinitions      int `json:"total_member_definitions"`
	TotalHealthyDefinitions     int `json:"total_healthy_definitions"`
	TotalProblemDefinitions     int `json:"total_problem_definitions"`
	TotalDownDefinitions        int `json:"total_down_definitions"`
	TotalNoDataDefinitions      int `json:"total_no_data_definitions"`
	TotalMaintenanceDefinitions int `json:"total_maintenance_definitions"`
}

type OfferingDashboardItem struct {
	ID               uint64                        `json:"id"`
	Name             string                        `json:"name"`
	Description      *string                       `json:"description"`
	Status           string                        `json:"status"`
	HealthyCount     int                           `json:"healthy_count"`
	ProblematicCount int                           `json:"problematic_count"`
	DownCount        int                           `json:"down_count"`
	NoDataCount      int                           `json:"no_data_count"`
	MaintenanceCount int                           `json:"maintenance_count"`
	TotalDefinitions int                           `json:"total_definitions"`
	Hostnames        []string                      `json:"hostnames"`
	Definitions      []OfferingDashboardDefinition `json:"definitions"`
}

type OfferingDashboardDefinition struct {
	ApplicationDefinitionID uint64 `json:"application_definition_id"`
	Name                    string `json:"name"`
	Type                    string `json:"type"`
	Status                  string `json:"status"`
	HealthyInstances        int    `json:"healthy_instances"`
	UnhealthyInstances      int    `json:"unhealthy_instances"`
	NoDataInstances         int    `json:"no_data_instances"`
	MaintenanceInstances    int    `json:"maintenance_instances"`
	TotalInstances          int    `json:"total_instances"`
}

type offeringDashboardDefinitionAgg struct {
	item                 OfferingDashboardDefinition
	totalInstances       int
	healthyInstances     int
	unhealthyInstances   int
	noDataInstances      int
	maintenanceInstances int
	hasInstance          bool
}

type offeringDashboardAgg struct {
	item       *OfferingDashboardItem
	definition map[uint64]*offeringDashboardDefinitionAgg
	hostnames  map[string]bool
}

func GetOfferingDashboardData(pool *pgxpool.Pool) (*OfferingDashboardData, error) {
	tx, err := pool.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(context.Background())

	rows, err := tx.Query(context.Background(), `
		SELECT
			so.id AS service_offering_id,
			so.name AS service_offering_name,
			so.description AS service_offering_description,
			ad.id AS application_definition_id,
			ad.name AS application_definition_name,
			ad.type AS application_definition_type,
			ai.id AS application_instance_id,
			ai.maintenance_mode,
			s.hostname AS server_hostname,
			hr.is_successful,
			hr.time_end
		FROM service_offering so
		LEFT JOIN service_offering_application_definition soad ON soad.service_offering_id = so.id
		LEFT JOIN application_definition ad ON ad.id = soad.application_definition_id
		LEFT JOIN application_instance ai ON ai.application_definition_id = ad.id
		LEFT JOIN "server" s ON s.id = ai.server_id
		LEFT JOIN LATERAL (
			SELECT is_successful, time_end
			FROM healthcheck_results hcr
			WHERE hcr.application_instance_id = ai.id
			ORDER BY hcr.time_end DESC
			LIMIT 1
		) hr ON ai.id IS NOT NULL
		ORDER BY so.name, ad.name, ai.name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	offeringAggs := make(map[uint64]*offeringDashboardAgg)

	for rows.Next() {
		var (
			offeringID      uint64
			offeringName    string
			offeringDesc    *string
			appDefID        *uint64
			appDefName      *string
			appDefType      *string
			instanceID      *uint64
			maintenanceMode *bool
			serverHostname  *string
			isSuccessful    *bool
			lastCheckTime   *time.Time
		)

		err = rows.Scan(
			&offeringID,
			&offeringName,
			&offeringDesc,
			&appDefID,
			&appDefName,
			&appDefType,
			&instanceID,
			&maintenanceMode,
			&serverHostname,
			&isSuccessful,
			&lastCheckTime,
		)
		if err != nil {
			return nil, err
		}

		agg, exists := offeringAggs[offeringID]
		if !exists {
			item := &OfferingDashboardItem{
				ID:          offeringID,
				Name:        offeringName,
				Description: offeringDesc,
				Status:      "no_data",
				Definitions: make([]OfferingDashboardDefinition, 0),
				Hostnames:   make([]string, 0),
			}
			agg = &offeringDashboardAgg{
				item:       item,
				definition: make(map[uint64]*offeringDashboardDefinitionAgg),
				hostnames:  make(map[string]bool),
			}
			offeringAggs[offeringID] = agg
		}

		if appDefID == nil || appDefName == nil || appDefType == nil {
			continue
		}

		defAgg, defExists := agg.definition[*appDefID]
		if !defExists {
			defAgg = &offeringDashboardDefinitionAgg{
				item: OfferingDashboardDefinition{
					ApplicationDefinitionID: *appDefID,
					Name:                    *appDefName,
					Type:                    *appDefType,
					Status:                  "no_data",
				},
			}
			agg.definition[*appDefID] = defAgg
		}

		if instanceID == nil || maintenanceMode == nil {
			continue
		}

		defAgg.hasInstance = true
		defAgg.totalInstances++

		if serverHostname != nil && *serverHostname != "" {
			agg.hostnames[strings.ToLower(*serverHostname)] = true
		}

		if *maintenanceMode {
			defAgg.maintenanceInstances++
			continue
		}

		if isSuccessful == nil || lastCheckTime == nil {
			defAgg.noDataInstances++
			continue
		}

		if *isSuccessful {
			defAgg.healthyInstances++
		} else {
			defAgg.unhealthyInstances++
		}
	}

	offerings := make([]OfferingDashboardItem, 0, len(offeringAggs))
	summary := OfferingDashboardSummary{}

	for _, agg := range offeringAggs {
		definitions := make([]OfferingDashboardDefinition, 0, len(agg.definition))

		for _, defAgg := range agg.definition {
			defAgg.item.TotalInstances = defAgg.totalInstances
			defAgg.item.HealthyInstances = defAgg.healthyInstances
			defAgg.item.UnhealthyInstances = defAgg.unhealthyInstances
			defAgg.item.NoDataInstances = defAgg.noDataInstances
			defAgg.item.MaintenanceInstances = defAgg.maintenanceInstances

			defAgg.item.Status = deriveDefinitionStatus(defAgg)
			switch defAgg.item.Status {
			case "healthy":
				agg.item.HealthyCount++
			case "problematic":
				agg.item.ProblematicCount++
			case "down":
				agg.item.DownCount++
			case "no_data":
				agg.item.NoDataCount++
			case "maintenance":
				agg.item.MaintenanceCount++
			}

			definitions = append(definitions, defAgg.item)
		}

		sort.Slice(definitions, func(i, j int) bool {
			return definitions[i].Name < definitions[j].Name
		})
		agg.item.Definitions = definitions
		agg.item.TotalDefinitions = len(definitions)
		agg.item.Status = deriveOfferingStatus(*agg.item)

		hostnames := make([]string, 0, len(agg.hostnames))
		for hostname := range agg.hostnames {
			hostnames = append(hostnames, hostname)
		}
		sort.Strings(hostnames)
		agg.item.Hostnames = hostnames

		summary.TotalMemberDefinitions += agg.item.TotalDefinitions
		summary.TotalHealthyDefinitions += agg.item.HealthyCount
		summary.TotalProblemDefinitions += agg.item.ProblematicCount
		summary.TotalDownDefinitions += agg.item.DownCount
		summary.TotalNoDataDefinitions += agg.item.NoDataCount
		summary.TotalMaintenanceDefinitions += agg.item.MaintenanceCount

		if agg.item.MaintenanceCount > 0 {
			summary.OfferingsWithMaintenance++
		}

		switch agg.item.Status {
		case "healthy":
			summary.HealthyOfferings++
		case "problematic":
			summary.ProblematicOfferings++
		case "down":
			summary.DownOfferings++
		case "no_data":
			summary.NoDataOfferings++
		}

		offerings = append(offerings, *agg.item)
	}

	summary.TotalOfferings = len(offerings)
	sort.Slice(offerings, func(i, j int) bool {
		return offerings[i].Name < offerings[j].Name
	})

	if err = tx.Commit(context.Background()); err != nil {
		return nil, err
	}

	return &OfferingDashboardData{
		Offerings:   offerings,
		Summary:     summary,
		RefreshedAt: time.Now(),
	}, nil
}

func deriveDefinitionStatus(def *offeringDashboardDefinitionAgg) string {
	if !def.hasInstance || def.totalInstances == 0 {
		return "no_data"
	}
	if def.maintenanceInstances == def.totalInstances {
		return "maintenance"
	}
	if def.unhealthyInstances > 0 {
		return "down"
	}
	nonMaintenance := def.totalInstances - def.maintenanceInstances
	if nonMaintenance <= 0 {
		return "maintenance"
	}
	if def.healthyInstances == nonMaintenance && def.noDataInstances == 0 {
		return "healthy"
	}
	if def.healthyInstances > 0 {
		return "problematic"
	}
	if def.noDataInstances > 0 {
		return "no_data"
	}
	return "down"
}

func deriveOfferingStatus(offering OfferingDashboardItem) string {
	if offering.TotalDefinitions == 0 {
		return "no_data"
	}
	if offering.DownCount > 0 {
		return "down"
	}
	if offering.MaintenanceCount == offering.TotalDefinitions {
		return "down"
	}
	if offering.HealthyCount > 0 && offering.HealthyCount+offering.MaintenanceCount == offering.TotalDefinitions {
		return "healthy"
	}
	if offering.NoDataCount == offering.TotalDefinitions {
		return "no_data"
	}
	if offering.HealthyCount > 0 || offering.NoDataCount > 0 || offering.ProblematicCount > 0 {
		return "problematic"
	}
	return "down"
}
