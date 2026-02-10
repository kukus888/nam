package services

import (
	"kukus/nam/v2/layers/data"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// This is an on-demand caching layer for dashboard data. Vibe-coded. Has bugs. But is good enough for now.

// DashboardCacheService provides caching for dashboard data
type DashboardCacheService struct {
	pool  *pgxpool.Pool
	lock  *sync.Mutex
	cache *data.DashboardData
}

var dashboardCache *DashboardCacheService = nil

func NewDashboardCacheService(pool *pgxpool.Pool) {
	dashboardCache = &DashboardCacheService{
		pool: pool,
		lock: &sync.Mutex{},
	}
	StartCacheRefresher()
}

// GetDashboardData returns cached dashboard data
func GetDashboardData(pool *pgxpool.Pool) (*data.DashboardData, error) {
	if dashboardCache == nil {
		NewDashboardCacheService(pool)
		return GetDashboardData(pool)
	}
	if dashboardCache.cache == nil {
		// If cache is not yet populated, fetch data directly from the database
		return data.GetDashboardData(pool)
	}
	return dashboardCache.cache, nil
}

// Periodically refresh the cache in the background
func StartCacheRefresher() {
	go func() {
		for {
			dashboardCache.lock.Lock()
			newData, err := data.GetDashboardData(dashboardCache.pool)
			if err == nil {
				dashboardCache.cache = newData
			}
			dashboardCache.lock.Unlock()
			// Sleep for a while before refreshing again
			time.Sleep(30 * time.Second)
		}
	}()
}
