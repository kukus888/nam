package services

import (
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// This file provides services for managing timer/cron jobs.

// TimerService manages multiple TimerJobs, allowing for centralized control and monitoring of scheduled tasks within the application. Singleton
type TimerService struct {
	Logger *slog.Logger
	DbPool *pgxpool.Pool
	Jobs   map[int]TimerJob
}

var TimerSvc *TimerService

func NewTimerService(pool *pgxpool.Pool, logger *slog.Logger) {
	if TimerSvc != nil {
		logger.Warn("TimerService already initialized, skipping re-initialization")
		return
	}
	ts := &TimerService{
		Logger: logger.With("service", "TimerService"),
		DbPool: pool,
		Jobs: map[int]TimerJob{
			0: NewDatabaseCleanupTimer(24*time.Hour, logger, pool),
			1: NewDatabaseHealthCheckResultFlusher(logger, pool),
		},
	}
	TimerSvc = ts
}

func GetTimerService() *TimerService {
	if TimerSvc == nil {
		panic("TimerService not initialized! Call NewTimerService first.")
	}
	return TimerSvc
}

// TimerJob represents a scheduled task that can be started, stopped, and run at specified intervals. It includes metadata such as name and description for better management and logging.
type TimerJob interface {
	Start()
	Stop()
	Run()
	Enable()
	Disable()
	IsEnabled() bool
	GetName() string
	GetDescription() string
	GetInterval() time.Duration
}

type TimerImpl struct {
	Name        string
	Description string
	Enabled     bool
	Interval    time.Duration
	Timer       *time.Timer
}

// Start begins the execution of the timer job.
func (t *TimerImpl) Start() {
	// Placeholder for job start logics
}

// Stop halts the execution of the timer job.
func (t *TimerImpl) Stop() {
	// Placeholder for job stop logics
}

// Run executes the timer job's task.
func (t *TimerImpl) Run() {
	// Placeholder for job execution logic
}

// Enable activates the timer job, allowing it to run at its scheduled intervals.
func (t *TimerImpl) Enable() {
	t.Enabled = true
}

// Disable deactivates the timer job, preventing it from running until re-enabled.
func (t *TimerImpl) Disable() {
	t.Enabled = false
}

// IsEnabled checks whether the timer job is currently enabled.
func (t *TimerImpl) IsEnabled() bool {
	return t.Enabled
}

// GetName returns the name of the timer job.
func (t *TimerImpl) GetName() string {
	return t.Name
}

// GetDescription returns a brief description of the timer job's purpose.
func (t *TimerImpl) GetDescription() string {
	return t.Description
}

// GetInterval returns the interval at which the timer job is scheduled to run.
func (t *TimerImpl) GetInterval() time.Duration {
	return t.Interval
}
