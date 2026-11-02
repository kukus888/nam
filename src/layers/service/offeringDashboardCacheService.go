package services

import (
	"kukus/nam/v2/layers/data"
	"log/slog"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type OfferingDashboardCacheService struct {
	pool   *pgxpool.Pool
	lock   *sync.RWMutex
	cache  *data.OfferingDashboardData
	logger *slog.Logger
}

var offeringDashboardCache *OfferingDashboardCacheService

func NewOfferingDashboardCacheService(pool *pgxpool.Pool, logger *slog.Logger) {
	if logger == nil {
		logger = slog.Default()
	}
	offeringDashboardCache = &OfferingDashboardCacheService{
		pool:   pool,
		lock:   &sync.RWMutex{},
		logger: logger.With("service", "OfferingDashboardCacheService"),
	}
	offeringDashboardCache.logger.Info("Initialized offering dashboard cache service")
	startOfferingDashboardCacheRefresher()
}

func GetOfferingDashboardData() (*data.OfferingDashboardData, error) {
	if offeringDashboardCache == nil {
		return nil, nil
	}

	offeringDashboardCache.lock.RLock()
	cached := offeringDashboardCache.cache
	offeringDashboardCache.lock.RUnlock()
	if cached != nil {
		return cached, nil
	}

	fresh, err := data.GetOfferingDashboardData(offeringDashboardCache.pool)
	if err != nil {
		return nil, err
	}

	offeringDashboardCache.lock.Lock()
	offeringDashboardCache.cache = fresh
	offeringDashboardCache.lock.Unlock()
	return fresh, nil
}

func startOfferingDashboardCacheRefresher() {
	go func() {
		offeringDashboardCache.logger.Info("Starting offering dashboard cache refresher", "refresh_interval", (30 * time.Second).String())
		for {
			fresh, err := data.GetOfferingDashboardData(offeringDashboardCache.pool)
			if err != nil {
				offeringDashboardCache.logger.Error("Failed to refresh offering dashboard cache", "error", err)
			} else {
				offeringDashboardCache.lock.Lock()
				offeringDashboardCache.cache = fresh
				offeringDashboardCache.lock.Unlock()
			}
			time.Sleep(30 * time.Second)
		}
	}()
}
