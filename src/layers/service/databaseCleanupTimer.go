package services

import (
	"kukus/nam/v2/layers/data"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// This file provides a specific implementation of a TimerJob for database cleanup tasks.
type DatabaseCleanupTimer struct {
	Logger      *slog.Logger
	DbPool      *pgxpool.Pool
	Name        string
	Description string
	Enabled     bool
	Interval    time.Duration
	Timer       *time.Timer
}

func NewDatabaseCleanupTimer(interval time.Duration, logger *slog.Logger, dbPool *pgxpool.Pool) *DatabaseCleanupTimer {
	return &DatabaseCleanupTimer{
		Logger:      logger.With("timer", "DatabaseCleanupTimer"),
		DbPool:      dbPool,
		Name:        "Database Cleanup Timer",
		Description: "Performs routine cleanup tasks on the database, squashing the healthcheck_results table.",
		Enabled:     false,
		Interval:    interval,
	}
}

// Start begins the execution of the timer job.
func (t *DatabaseCleanupTimer) Start() {
	timer := time.NewTimer(t.Interval)
	t.Timer = timer
	go func() {
		for {
			<-timer.C
			if t.Enabled {
				t.Run()
			}
			timer.Reset(t.Interval)
		}
	}()
}

// Stop halts the execution of the timer job.
func (t *DatabaseCleanupTimer) Stop() {
	if t.Timer != nil {
		t.Logger.Info("Stopping timer job")
		t.Timer.Stop()
	}
}

// Run executes the timer job's task.
func (t *DatabaseCleanupTimer) Run() {
	res, err := data.CleanUpDatabase(t.DbPool)
	if err != nil {
		t.Logger.Error("Database cleanup failed", "error", err)
	} else {
		t.Logger.Info("Database cleanup succeeded", "result", res)
	}
}

// Enable activates the timer job, allowing it to run at its scheduled intervals.
func (t *DatabaseCleanupTimer) Enable() {
	t.Enabled = true
	t.Logger.Info("Enabled timer job")
}

// Disable deactivates the timer job, preventing it from running until re-enabled.
func (t *DatabaseCleanupTimer) Disable() {
	t.Enabled = false
	t.Logger.Info("Disabled timer job")
}

// IsEnabled checks whether the timer job is currently enabled.
func (t *DatabaseCleanupTimer) IsEnabled() bool {
	return t.Enabled
}

// GetName returns the name of the timer job.
func (t *DatabaseCleanupTimer) GetName() string {
	return t.Name
}

// GetDescription returns a brief description of the timer job's purpose.
func (t *DatabaseCleanupTimer) GetDescription() string {
	return t.Description
}

// GetInterval returns the interval at which the timer job is scheduled to run.
func (t *DatabaseCleanupTimer) GetInterval() time.Duration {
	return t.Interval
}
