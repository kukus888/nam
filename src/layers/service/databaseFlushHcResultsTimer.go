package services

import (
	"kukus/nam/v2/layers/data"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// This file provides a specific implementation of a TimerJob for database cleanup tasks.
type DatabaseHealthCheckResultFlusher struct {
	Logger      *slog.Logger
	DbPool      *pgxpool.Pool
	Name        string
	Description string
	Enabled     bool
	Interval    time.Duration
}

func NewDatabaseHealthCheckResultFlusher(logger *slog.Logger, dbPool *pgxpool.Pool) *DatabaseHealthCheckResultFlusher {
	return &DatabaseHealthCheckResultFlusher{
		Logger:      logger.With("timer", "DatabaseHealthCheckResultFlusher"),
		DbPool:      dbPool,
		Name:        "Database Health Check Result Flusher",
		Description: "Flushes the healthcheck_results table. This deletes every healthcheck result you have. Use only for one-time cleanup or in very long intervals. This cannot be enabled.",
		Enabled:     false,
		Interval:    time.Hour * 24 * 7 * 52 * 69, // Default to 69 years
	}
}

// Start begins the execution of the timer job.
func (t *DatabaseHealthCheckResultFlusher) Start() {
	// This timer is not meant to be started, but we implement this method to satisfy the TimerJob interface.
}

// Stop halts the execution of the timer job.
func (t *DatabaseHealthCheckResultFlusher) Stop() {
	// This timer is not meant to be stopped, but we implement this method to satisfy the TimerJob interface.
}

// Run executes the timer job's task.
func (t *DatabaseHealthCheckResultFlusher) Run() {
	res, err := data.FlushHealthCheckResults(t.DbPool)
	if err != nil {
		t.Logger.Error("Database flush of healthcheck_results table failed", "error", err)
	} else {
		t.Logger.Info("Database flush of healthcheck_results table succeeded", "result", res)
	}
}

// Enable activates the timer job, allowing it to run at its scheduled intervals.
func (t *DatabaseHealthCheckResultFlusher) Enable() {
	t.Enabled = false
}

// Disable deactivates the timer job, preventing it from running until re-enabled.
func (t *DatabaseHealthCheckResultFlusher) Disable() {
	t.Enabled = false
}

// IsEnabled checks whether the timer job is currently enabled.
func (t *DatabaseHealthCheckResultFlusher) IsEnabled() bool {
	return t.Enabled
}

// GetName returns the name of the timer job.
func (t *DatabaseHealthCheckResultFlusher) GetName() string {
	return t.Name
}

// GetDescription returns a brief description of the timer job's purpose.
func (t *DatabaseHealthCheckResultFlusher) GetDescription() string {
	return t.Description
}

// GetInterval returns the interval at which the timer job is scheduled to run.
func (t *DatabaseHealthCheckResultFlusher) GetInterval() time.Duration {
	return t.Interval
}
