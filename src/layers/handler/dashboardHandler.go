package handlers

import (
	"kukus/nam/v2/layers/data"
	services "kukus/nam/v2/layers/service"
	"log/slog"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
)

func (pc PageHandler) GetPageDashboards(ctx *gin.Context) {
	ctx.HTML(200, "pages/dashboards", gin.H{})
}

// GetPageDashboardNew renders the new efficient dashboard
func (pc PageHandler) GetPageDashboard(ctx *gin.Context) {
	dashboardData, err := data.GetDashboardData(pc.Database.Pool)
	if err != nil {
		slog.Error("Failed to render dashboard page", "error", err, "path", ctx.Request.URL.Path)
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

	dashboardData, err := services.GetDashboardData()
	if err != nil {
		slog.Error("Failed to fetch dashboard API data", "error", err, "path", ctx.Request.URL.Path)
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

	dashboardData, err := services.GetDashboardData()
	if err != nil {
		slog.Error("Failed to render dashboard component", "error", err, "path", ctx.Request.URL.Path)
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

func (pc PageHandler) GetPageOfferingsDashboard(ctx *gin.Context) {
	dashboardData, err := services.GetOfferingDashboardData()
	if err != nil || dashboardData == nil {
		slog.Error("Failed to render offerings dashboard page", "error", err, "path", ctx.Request.URL.Path)
		ctx.AbortWithStatusJSON(500, gin.H{"error": "Unable to get offerings dashboard data", "trace": err})
		return
	}

	ctx.HTML(200, "pages/dashboard-offerings", gin.H{
		"Dashboard": dashboardData,
	})
}

func (pc PageHandler) GetOfferingsDashboardDataAPI(ctx *gin.Context) {
	dashboardData, err := services.GetOfferingDashboardData()
	if err != nil || dashboardData == nil {
		slog.Error("Failed to fetch offerings dashboard API data", "error", err, "path", ctx.Request.URL.Path)
		ctx.AbortWithStatusJSON(500, gin.H{"error": "Unable to get offerings dashboard data", "trace": err})
		return
	}

	filtered := filterAndSortOfferingDashboardItems(dashboardData.Offerings, ctx)
	dashboardData.Offerings = filtered
	ctx.JSON(200, dashboardData)
}

func (pc PageHandler) GetOfferingsDashboardComponent(ctx *gin.Context) {
	dashboardData, err := services.GetOfferingDashboardData()
	if err != nil || dashboardData == nil {
		slog.Error("Failed to render offerings dashboard component", "error", err, "path", ctx.Request.URL.Path)
		ctx.AbortWithStatusJSON(500, gin.H{"error": "Unable to get offerings dashboard data", "trace": err})
		return
	}

	filtered := filterAndSortOfferingDashboardItems(dashboardData.Offerings, ctx)
	ctx.HTML(200, "components/dashboard-offerings", filtered)
}

func filterAndSortOfferingDashboardItems(items []data.OfferingDashboardItem, ctx *gin.Context) []data.OfferingDashboardItem {
	showHealthy := ctx.DefaultQuery("healthy", "true") == "true"
	showProblematic := ctx.DefaultQuery("problematic", "true") == "true"
	showDown := ctx.DefaultQuery("down", "true") == "true"
	showMaintenance := ctx.DefaultQuery("maintenance", "true") == "true"
	showNoData := ctx.DefaultQuery("no_data", "true") == "true"
	nameFilter := strings.ToLower(ctx.DefaultQuery("offering_name", ""))
	hostnameFilter := strings.ToLower(ctx.DefaultQuery("hostname", ""))
	sortBy := ctx.DefaultQuery("sort", "name")

	filtered := make([]data.OfferingDashboardItem, 0)
	for _, item := range items {
		if nameFilter != "" && !strings.Contains(strings.ToLower(item.Name), nameFilter) {
			continue
		}

		if hostnameFilter != "" {
			hostFound := false
			for _, hostname := range item.Hostnames {
				if strings.Contains(strings.ToLower(hostname), hostnameFilter) {
					hostFound = true
					break
				}
			}
			if !hostFound {
				continue
			}
		}

		statusAllowed := false
		switch item.Status {
		case "healthy":
			statusAllowed = showHealthy
		case "problematic":
			statusAllowed = showProblematic
		case "down":
			statusAllowed = showDown
		case "no_data":
			statusAllowed = showNoData
		}
		if !statusAllowed {
			continue
		}

		if !showMaintenance && item.MaintenanceCount > 0 {
			continue
		}

		filtered = append(filtered, item)
	}

	if sortBy == "severity" {
		sort.Slice(filtered, func(i, j int) bool {
			si := offeringSeverityRank(filtered[i].Status)
			sj := offeringSeverityRank(filtered[j].Status)
			if si == sj {
				return filtered[i].Name < filtered[j].Name
			}
			return si < sj
		})
	} else {
		sort.Slice(filtered, func(i, j int) bool {
			return filtered[i].Name < filtered[j].Name
		})
	}

	return filtered
}

func offeringSeverityRank(status string) int {
	switch status {
	case "down":
		return 0
	case "problematic":
		return 1
	case "healthy":
		return 2
	case "no_data":
		return 3
	default:
		return 4
	}
}
