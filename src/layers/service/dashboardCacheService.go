package services

import (
	"kukus/nam/v2/layers/data"
	"log/slog"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// This is an on-demand caching layer for dashboard data. Vibe-coded. Has bugs. But is good enough for now.

// DashboardCacheService provides caching for dashboard data
type DashboardCacheService struct {
	pool   *pgxpool.Pool
	lock   *sync.Mutex
	cache  *data.DashboardData
	logger *slog.Logger
}

var dashboardCache *DashboardCacheService = nil

func NewDashboardCacheService(pool *pgxpool.Pool, logger *slog.Logger) {
	if logger == nil {
		logger = slog.Default()
	}
	dashboardCache = &DashboardCacheService{
		pool:   pool,
		lock:   &sync.Mutex{},
		logger: logger.With("service", "DashboardCacheService"),
	}
	dashboardCache.logger.Info("Initialized dashboard cache service")
	StartCacheRefresher()
}

// GetDashboardData returns cached dashboard data
func GetDashboardData() (*data.DashboardData, error) {
	if dashboardCache == nil {
		// Log warn, but don't panic - this can happen if the service is not initialized yet
		waitPasses := 3
		waitTime := 5 * time.Second
		slog.Warn("Dashboard cache service not initialized, trying to wait for initialization")
		// Wait 3 times 5 seconds, then panic if still not initialized - this is a safeguard against the service not being initialized at all
		for i := 1; i <= waitPasses; i++ {
			time.Sleep(waitTime)
			if dashboardCache != nil {
				dashboardCache.logger.Info("Dashboard cache service initialized after wait", "wait_pass", i)
				return GetDashboardData()
			} else {
				slog.Warn("Dashboard cache service still not initialized after wait", "wait_pass", i, "wait_time_seconds", waitTime.Seconds())
			}
		}
		slog.Error("Dashboard cache service failed to initialize after retries")
		panic("DashboardCacheService not initialized after waiting, cannot fetch dashboard data")
	}
	if dashboardCache.cache == nil {
		// If cache is not yet populated, fetch data directly from the database
		dashboardCache.logger.Debug("Dashboard cache miss, fetching data directly from database")
		return data.GetDashboardData(dashboardCache.pool)
	}
	return dashboardCache.cache, nil
}

// Periodically refresh the cache in the background
func StartCacheRefresher() {
	go func() {
		dashboardCache.logger.Info("Starting dashboard cache refresher", "refresh_interval", (30 * time.Second).String())
		for {
			newData, err := data.GetDashboardData(dashboardCache.pool)
			if err == nil {
				dashboardCache.lock.Lock()
				dashboardCache.cache = newData
				dashboardCache.lock.Unlock()
			} else {
				dashboardCache.logger.Error("Failed to refresh dashboard cache", "error", err)
			}
			// Sleep for a while before refreshing again
			time.Sleep(30 * time.Second)
		}
	}()
}
