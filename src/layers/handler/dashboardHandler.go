package handlers

import (
	"kukus/nam/v2/layers/data"
	"strings"

	"github.com/gin-gonic/gin"
)

// GetPageDashboardNew renders the new efficient dashboard
func (pc PageHandler) GetPageDashboard(ctx *gin.Context) {
	dashboardData, err := data.GetDashboardData(pc.Database.Pool)
	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"error": "Unable to get dashboard data", "trace": err.Error()})
		return
	}

	ctx.HTML(200, "pages/dashboard-new", gin.H{
		"Dashboard": dashboardData,
	})
}

// GetDashboardDataAPI returns dashboard data as JSON for HTMX updates
func (pc PageHandler) GetDashboardDataAPI(ctx *gin.Context) {
	// Parse filters from query parameters
	showHealthy := ctx.DefaultQuery("healthy", "true") == "true"
	showUnhealthy := ctx.DefaultQuery("unhealthy", "true") == "true"
	showMaintenance := ctx.DefaultQuery("maintenance", "true") == "true"
	showUnknown := ctx.DefaultQuery("unknown", "true") == "true"
	appNameFilter := strings.ToLower(ctx.DefaultQuery("app_name", ""))
	appTypeFilter := ctx.DefaultQuery("app_type", "")

	dashboardData, err := data.GetDashboardData(pc.Database.Pool)
	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"error": "Unable to get dashboard data", "trace": err.Error()})
		return
	}

	// Apply filters
	var filteredApplications []data.DashboardApplication
	for _, app := range dashboardData.Applications {
		// Apply app name filter
		if appNameFilter != "" && !strings.Contains(strings.ToLower(app.Name), appNameFilter) {
			continue
		}

		// Apply app type filter
		if appTypeFilter != "" && app.Type != appTypeFilter {
			continue
		}

		// Apply health status filter
		includeApp := false
		switch app.HealthStatus {
		case "healthy":
			includeApp = showHealthy
		case "degraded", "unhealthy":
			includeApp = showUnhealthy
		case "unknown":
			includeApp = showUnknown
		}

		// Also check if any instances are in maintenance mode and maintenance filter is on
		if !includeApp && showMaintenance {
			for _, instance := range app.Instances {
				if instance.MaintenanceMode {
					includeApp = true
					break
				}
			}
		}

		if includeApp {
			filteredApplications = append(filteredApplications, app)
		}
	}

	dashboardData.Applications = filteredApplications
	ctx.JSON(200, dashboardData)
}

// GetDashboardComponent returns just the applications list for HTMX replacement
func (pc PageHandler) GetDashboardComponent(ctx *gin.Context) {
	// Parse filters from query parameters
	showHealthy := ctx.DefaultQuery("healthy", "true") == "true"
	showUnhealthy := ctx.DefaultQuery("unhealthy", "true") == "true"
	showMaintenance := ctx.DefaultQuery("maintenance", "true") == "true"
	showUnknown := ctx.DefaultQuery("unknown", "true") == "true"
	appNameFilter := strings.ToLower(ctx.DefaultQuery("app_name", ""))
	appTypeFilter := ctx.DefaultQuery("app_type", "")

	dashboardData, err := data.GetDashboardData(pc.Database.Pool)
	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"error": "Unable to get dashboard data", "trace": err.Error()})
		return
	}

	// Apply filters
	var filteredApplications []data.DashboardApplication
	for _, app := range dashboardData.Applications {
		// Apply app name filter
		if appNameFilter != "" && !strings.Contains(strings.ToLower(app.Name), appNameFilter) {
			continue
		}

		// Apply app type filter
		if appTypeFilter != "" && app.Type != appTypeFilter {
			continue
		}

		// Apply health status filter
		includeApp := false
		switch app.HealthStatus {
		case "healthy":
			includeApp = showHealthy
		case "degraded", "unhealthy":
			includeApp = showUnhealthy
		case "unknown":
			includeApp = showUnknown
		}

		// Also check if any instances are in maintenance mode and maintenance filter is on
		if !includeApp && showMaintenance {
			for _, instance := range app.Instances {
				if instance.MaintenanceMode {
					includeApp = true
					break
				}
			}
		}

		if includeApp {
			filteredApplications = append(filteredApplications, app)
		}
	}

	ctx.HTML(200, "components/dashboard-applications", filteredApplications)
}
