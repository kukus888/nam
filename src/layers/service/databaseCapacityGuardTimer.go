package services

import (
	"kukus/nam/v2/layers/data"
	"log/slog"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DatabaseCapacityGuardTimer checks database usage against configured limits,
// then runs smart cleanup and fallback oldest-first deletion when needed.
type DatabaseCapacityGuardTimer struct {
	mu          sync.RWMutex
	Logger      *slog.Logger
	DbPool      *pgxpool.Pool
	Name        string
	Description string
	Enabled     bool
	Interval    time.Duration
	Timer       *time.Timer
}

const (
	fallbackDeleteBatchRows int64 = 5000
	fallbackMaxBatches      int   = 8
)

func NewDatabaseCapacityGuardTimer(interval time.Duration, logger *slog.Logger, dbPool *pgxpool.Pool) *DatabaseCapacityGuardTimer {
	return &DatabaseCapacityGuardTimer{
		Logger:      logger.With("timer", "DatabaseCapacityGuardTimer"),
		DbPool:      dbPool,
		Name:        "Database Capacity Guard",
		Description: "Monitors DB size and protects capacity by smart cleanup first, then oldest-first hard deletion of healthcheck_results.",
		Enabled:     false,
		Interval:    interval,
	}
}

func (t *DatabaseCapacityGuardTimer) Start() {
	timer := time.NewTimer(t.GetInterval())
	t.Timer = timer
	go func() {
		for {
			<-timer.C
			if t.IsEnabled() {
				t.Run()
			}
			timer.Reset(t.GetInterval())
		}
	}()
}

func (t *DatabaseCapacityGuardTimer) Stop() {
	if t.Timer != nil {
		t.Logger.Info("Stopping timer job")
		t.Timer.Stop()
	}
}

func (t *DatabaseCapacityGuardTimer) Run() {
	cfg, err := data.GetDatabaseCapacityGuardConfig(t.DbPool)
	if err != nil {
		t.Logger.Error("Capacity guard skipped: invalid configuration", "error", err)
		return
	}

	if cfg.AllocatedBytes <= 0 {
		t.Logger.Error("Capacity guard skipped: db_capacity_allocated_bytes must be greater than zero")
		return
	}

	dbSize, err := data.GetDatabaseSize(t.DbPool)
	if err != nil {
		t.Logger.Error("Capacity guard failed to read database size", "error", err)
		return
	}
	usagePercent := (float64(*dbSize) / float64(cfg.AllocatedBytes)) * 100
	if usagePercent < cfg.WarnPercent {
		t.Logger.Debug("Capacity guard OK", "db_size_bytes", *dbSize, "allocated_bytes", cfg.AllocatedBytes, "usage_percent", usagePercent)
		return
	}

	t.Logger.Warn("Capacity guard threshold reached", "db_size_bytes", *dbSize, "allocated_bytes", cfg.AllocatedBytes, "usage_percent", usagePercent, "warn_percent", cfg.WarnPercent, "critical_percent", cfg.CriticalPercent)

	cleanupResult, cleanupErr := data.CleanUpDatabase(t.DbPool)
	if cleanupErr != nil {
		t.Logger.Error("Capacity guard smart cleanup failed", "error", cleanupErr)
	} else {
		t.Logger.Info("Capacity guard smart cleanup complete", "result", cleanupResult)
	}

	updatedDbSize, err := data.GetDatabaseSize(t.DbPool)
	if err != nil {
		t.Logger.Error("Capacity guard failed to read post-cleanup database size", "error", err)
		return
	}
	updatedUsagePercent := (float64(*updatedDbSize) / float64(cfg.AllocatedBytes)) * 100
	if updatedUsagePercent < cfg.CriticalPercent {
		t.Logger.Info("Capacity guard resolved after smart cleanup", "db_size_bytes", *updatedDbSize, "usage_percent", updatedUsagePercent)
		return
	}

	deletedTotal := int64(0)
	for batch := 0; batch < fallbackMaxBatches; batch++ {
		deletedRows, delErr := data.DeleteOldestHealthcheckResultsBatch(t.DbPool, fallbackDeleteBatchRows)
		if delErr != nil {
			t.Logger.Error("Capacity guard fallback deletion failed", "error", delErr, "batch_size", fallbackDeleteBatchRows)
			break
		}
		if *deletedRows == 0 {
			t.Logger.Info("Capacity guard fallback deletion found no more rows")
			break
		}

		deletedTotal += *deletedRows
		currentDbSize, sizeErr := data.GetDatabaseSize(t.DbPool)
		if sizeErr != nil {
			t.Logger.Error("Capacity guard failed to read database size after fallback batch", "error", sizeErr)
			break
		}

		currentUsagePercent := (float64(*currentDbSize) / float64(cfg.AllocatedBytes)) * 100
		t.Logger.Warn("Capacity guard fallback batch executed", "deleted_rows_batch", *deletedRows, "deleted_rows_total", deletedTotal, "db_size_bytes", *currentDbSize, "usage_percent", currentUsagePercent)

		if currentUsagePercent < cfg.WarnPercent {
			t.Logger.Info("Capacity guard resolved after fallback deletion", "db_size_bytes", *currentDbSize, "usage_percent", currentUsagePercent, "deleted_rows_total", deletedTotal)
			return
		}
	}

	finalDbSize, err := data.GetDatabaseSize(t.DbPool)
	if err != nil {
		t.Logger.Error("Capacity guard failed final database size read", "error", err)
		return
	}
	finalUsagePercent := (float64(*finalDbSize) / float64(cfg.AllocatedBytes)) * 100
	t.Logger.Error("Capacity guard finished but usage still high", "db_size_bytes", *finalDbSize, "usage_percent", finalUsagePercent, "deleted_rows_total", deletedTotal, "batch_size", fallbackDeleteBatchRows, "max_batches", fallbackMaxBatches)
}

func (t *DatabaseCapacityGuardTimer) SetInterval(interval time.Duration) {
	if interval <= 0 {
		t.Logger.Warn("Rejected invalid timer interval update", "interval", interval.String())
		return
	}
	t.mu.Lock()
	t.Interval = interval
	if t.Timer != nil {
		t.Timer.Reset(interval)
	}
	t.mu.Unlock()
	t.Logger.Info("Updated timer interval", "interval", interval.String())
}

func (t *DatabaseCapacityGuardTimer) Enable() {
	t.mu.Lock()
	t.Enabled = true
	t.mu.Unlock()
	t.Logger.Info("Enabled timer job")
}

func (t *DatabaseCapacityGuardTimer) Disable() {
	t.mu.Lock()
	t.Enabled = false
	t.mu.Unlock()
	t.Logger.Info("Disabled timer job")
}

func (t *DatabaseCapacityGuardTimer) IsEnabled() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.Enabled
}

func (t *DatabaseCapacityGuardTimer) GetName() string {
	return t.Name
}

func (t *DatabaseCapacityGuardTimer) GetDescription() string {
	return t.Description
}

func (t *DatabaseCapacityGuardTimer) GetInterval() time.Duration {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.Interval
}
