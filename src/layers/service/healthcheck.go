package services

import (
	"context"
	"kukus/nam/v2/layers/data"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TODO: Logging

type HealthcheckService struct {
	Database          *data.Database
	ListenerConnector *pgx.Conn                     // Connector for listening to healthcheck changes
	Status            string                        // Status of the healthcheck service, e.g., "running", "stopped"
	Observers         map[uint]*HealthcheckObserver // Map of healthchecks being monitored with their observers
	Logger            *slog.Logger
}

func NewHealthcheckService(database *data.Database, logger *slog.Logger) *HealthcheckService {
	hcs := HealthcheckService{
		Database:  database,
		Observers: make(map[uint]*HealthcheckObserver),
		Logger:    logger,
	}
	go hcs.Start()
	return &hcs
}

func (hcs *HealthcheckService) Start() error {
	hcs.Status = "starting"
	// Initialize the listener connector
	conn, err := hcs.Database.Pool.Acquire(context.Background())
	if err != nil {
		hcs.Logger.Error("Failed to acquire database connection for healthcheck listener", "error", err)
		hcs.Status = "error"
		return err
	}
	hcs.ListenerConnector = conn.Conn()
	// Start listening for notifications
	_, err = hcs.ListenerConnector.Exec(context.Background(), "LISTEN healthcheck_changes")
	if err != nil {
		hcs.Logger.Error("Failed to start listening for healthcheck changes", "error", err)
		hcs.Status = "error"
		return err
	}
	hcs.Status = "running"
	// Sync existing healthchecks from the database
	hcs.SyncObservers("")
	go hcs.ListenForChanges()
	hcs.Logger.Info("Healthcheck service started successfully")
	return nil
}

func (hcs *HealthcheckService) ListenForChanges() {
	for {
		// Wait for notifications
		notification, err := hcs.ListenerConnector.WaitForNotification(context.Background())
		if err != nil {
			hcs.Logger.Error("Failed to receive notification", "error", err)
			hcs.Status = "error"
			return
		}
		if notification != nil {
			// Handle the notification (e.g., log it, update internal state, etc.)
			hcs.SyncObservers(notification.Payload)
		}
	}
}

// Syncs the healthchecks from the database
// payload -
func (hcs *HealthcheckService) SyncObservers(payload string) error {
	// Parse payload to get the healthcheck ID and operation
	parts := strings.Split(payload, ":")
	if len(parts) != 2 {
		// full sync
		// Clear existing observers
		for id := range hcs.Observers {
			if observer, exists := hcs.Observers[id]; exists {
				observer.TimerCancel() // Cancel the timer for each observer
			}
		}
		healthchecks, err := data.GetHealthChecksAll(hcs.Database.Pool)
		if err != nil {
			hcs.Logger.Error("Failed to get healthchecks from database", "error", err)
			hcs.UpdateStatus("error")
			return err
		}
		hcs.Observers = make(map[uint]*HealthcheckObserver)
		for _, hc := range *healthchecks {
			obj := HealthcheckObserver{
				Healthcheck:     &hc,
				Timer:           time.NewTimer(hc.CheckInterval),
				TargetInstances: make(map[uint]string), // Initialize the map for target instances
			}
			// Get port from application definition
			// Get server instances for the application
			// Get the URL together
			targets, err := data.GetHealthcheckTargets(hcs.Database.Pool, *hc.ID)
			if err != nil {
				hcs.Logger.Error("Failed to get healthcheck targets", "error", err)
				hcs.UpdateStatus("error")
				return err
			}
			protocol := "http"
			if hc.VerifySSL {
				protocol = "https"
			}
			for _, target := range *targets {
				// TODO: Support HTTPS properly
				obj.TargetInstances[target.ApplicationInstanceID] = protocol + "://" + target.Hostname + ":" + strconv.Itoa(int(target.Port)) + target.Url
			}
			obj.Logger = hcs.Logger.With("healthcheck_id", *hc.ID, "healthcheck_name", hc.Name)
			obj.Start(hcs.Database.Pool)
			hcs.Observers[*hc.ID] = &obj
			hcs.Logger.Debug("Healthcheck observer started", "id", *hc.ID, "name", hc.Name)
		}
	} else {
		id, err := strconv.Atoi(parts[1])
		if err != nil {
			// full sync
			hcs.SyncObservers("")
			return nil
		}
		switch parts[0] {
		case "INSERT":
			// Add new ID to observers
			// TODO: Implement all the logic as in the full sync
			/*
				hc, err := data.GetHealthCheckById(hcs.Database.Pool, uint(id))
				if err != nil {
					hcs.Status = "error"
					return err
				}
				obj := HealthcheckObserver{
					Healthcheck: hc,
					Timer:       time.NewTimer(hc.CheckInterval),
				}
				hcs.Observers[uint(id)] = &obj
				hcs.Observers[uint(id)].Start(hcs.Database.Pool)
			*/
			hcs.SyncObservers("")
			return nil
		case "UPDATE":
			// Update existing observer
			if observer, exists := hcs.Observers[uint(id)]; exists {
				hc, err := data.GetHealthCheckById(hcs.Database.Pool, uint(id))
				if err != nil {
					hcs.UpdateStatus("error")
					return err
				}
				observer.Healthcheck = hc
				// Reset the timer with the new interval
				observer.TimerCancel()
				observer.Timer.Reset(hc.CheckInterval)
				hcs.Observers[uint(id)] = observer
			} else {
				// If observer does not exist, we can treat it as an insert
				hc, err := data.GetHealthCheckById(hcs.Database.Pool, uint(id))
				if err != nil {
					hcs.UpdateStatus("error")
					return err
				}
				obj := HealthcheckObserver{
					Healthcheck:     hc,
					Timer:           time.NewTimer(hc.CheckInterval),
					TargetInstances: make(map[uint]string), // Initialize the map for target instances
				}
				hcs.Observers[uint(id)] = &obj
				hcs.Observers[uint(id)].Start(hcs.Database.Pool)
			}
		case "DELETE":
			// Remove observer
			if observer, exists := hcs.Observers[uint(id)]; exists {
				// Stop the timer and remove the observer
				observer.TimerCancel()
				delete(hcs.Observers, uint(id))
			} else {
				// If observer does not exist, we can ignore this operation
			}
		default:
			// Unknown operation, full sync
			hcs.SyncObservers("")
			return nil
		}
	}

	return nil
}

type HealthcheckObserver struct {
	Healthcheck     *data.Healthcheck  // The healthcheck being observed
	Timer           *time.Timer        // Timer for periodic checks
	TimerCancel     context.CancelFunc // Cancel function for the timer
	TargetInstances map[uint]string    // map of URLs to check, key is the application ID, value is the server URL
	DbPool          *pgxpool.Pool      // Database connection pool
	ProbeFunc       func()             // Function to perform the healthcheck probe
	Logger          *slog.Logger
}

func (hco *HealthcheckObserver) Start(pool *pgxpool.Pool) {
	hco.DbPool = pool
	hco.Timer = time.NewTimer(hco.Healthcheck.CheckInterval)
	hco.TimerCancel = func() {
		if !hco.Timer.Stop() {
			<-hco.Timer.C // Drain the channel if the timer was already fired
		}
	}
	// Set up the healthcheck probe function
	hco.ProbeFunc = func() {
		// Perform the healthcheck for all associated applications, on all associated servers
		results := make([]data.HealthcheckResult, 0)
		for instanceId, url := range hco.TargetInstances {
			result, err := hco.Healthcheck.PerformCheck(url)
			result.ApplicationInstanceID = instanceId
			if err != nil {
				// Happens only if there is something wrong on the network layer
				hco.Logger.Debug("Healthcheck failed", "instance_id", instanceId, "url", url, "error", err.Error())
			}
			hco.Logger.Debug("Healthcheck result", "instance_id", instanceId, "is_successful", result.IsSuccessful, "status", result.ResStatus, "response_time", result.ResTime)
			results = append(results, *result)
		}
		err := data.HealthcheckResultBatchInsert(hco.DbPool, &results)
		if err != nil {
			hco.Logger.Error("Failed to insert healthcheck results into database", "error", err)
		}
		// Reset the timer for the next check
		hco.Timer.Reset(hco.Healthcheck.CheckInterval)
	}
	// Start the healthcheck probe function
	go func() {
		hco.ProbeFunc() // Initial call to start the healthcheck immediately
		for {           // this is okay, ignore linter
			select {
			case <-hco.Timer.C:
				hco.ProbeFunc()
			}
		}
	}()
}

// Sets the new status for the service. Useful for debugging
func (hcs *HealthcheckService) UpdateStatus(newStatus string) {
	hcs.Status = newStatus
}

func (hcs *HealthcheckService) Stop() error {
	hcs.Status = "stopping"
	if hcs.ListenerConnector != nil {
		// Unlisten and close the connection
		_, err := hcs.ListenerConnector.Exec(context.Background(), "UNLISTEN healthcheck_changes")
		if err != nil {
			hcs.Status = "error"
			return err
		}
		hcs.ListenerConnector.Close(context.Background())
	}
	hcs.Status = "stopped"
	return nil
}

func (hcs *HealthcheckService) IsRunning() bool {
	return hcs.Status == "running"
}
func (hcs *HealthcheckService) GetName() string {
	return "HealthcheckService"
}
func (hcs *HealthcheckService) GetDescription() string {
	return "Service that monitors healthchecks and listens for changes in the database."
}
func (hcs *HealthcheckService) GetStatus() string {
	return hcs.Status
}
