package services

import (
	"context"
	"kukus/nam/v2/layers/data"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

// TODO: Logging

type HealthcheckService struct {
	Database          *data.Database
	ListenerConnector *pgx.Conn                     // Connector for listening to healthcheck changes
	Status            string                        // Status of the healthcheck service, e.g., "running", "stopped"
	Observers         map[uint]*HealthcheckObserver // Map of healthchecks being monitored with their observers
}

func NewHealthcheckService(database *data.Database) *HealthcheckService {
	hcs := HealthcheckService{
		Database:  database,
		Observers: make(map[uint]*HealthcheckObserver),
	}
	go hcs.Start()
	return &hcs
}

func (hcs *HealthcheckService) Start() error {
	hcs.Status = "starting"
	// Initialize the listener connector
	conn, err := hcs.Database.Pool.Acquire(context.Background())
	if err != nil {
		hcs.Status = "error"
		return err
	}
	hcs.ListenerConnector = conn.Conn()
	// Start listening for notifications
	_, err = hcs.ListenerConnector.Exec(context.Background(), "LISTEN healthcheck_changes")
	if err != nil {
		hcs.Status = "error"
		return err
	}
	hcs.Status = "running"
	// Sync existing healthchecks from the database
	hcs.SyncObservers("")
	go hcs.ListenForChanges()
	return nil
}

func (hcs *HealthcheckService) ListenForChanges() {
	for {
		// Wait for notifications
		notification, err := hcs.ListenerConnector.WaitForNotification(context.Background())
		if err != nil {
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
	if len(parts) < 2 {
		// full sync
		// Clear existing observers
		for id := range hcs.Observers {
			if observer, exists := hcs.Observers[id]; exists {
				observer.TimerCancel() // Cancel the timer for each observer
			}
		}
		healthchecks, err := data.GetHealthChecksAll(hcs.Database.Pool)
		if err != nil {
			hcs.Status = "error"
			return err
		}
		hcs.Observers = make(map[uint]*HealthcheckObserver)
		for _, hc := range *healthchecks {
			obj := HealthcheckObserver{
				Healthcheck: &hc,
				Timer:       time.NewTimer(hc.CheckInterval),
			}
			hcs.Observers[*hc.ID] = &obj
			// Get port from application definition
			// Get server instances for the application
			// Get the URL together
			targets, err := data.GetHealthcheckTargets(hcs.Database.Pool, *hc.ID)
			if err != nil {
				hcs.Status = "error"
				return err
			}
			for _, target := range *targets {
				// TODO: Support HTTPS
				obj.TargetURLs = append(obj.TargetURLs, "http://"+target.Hostname+":"+strconv.Itoa(int(target.Port))+target.Url)
			}
			hcs.Observers[*hc.ID].Start()
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
			hcs.Observers[uint(id)].Start()
		case "UPDATE":
			// Update existing observer
			if observer, exists := hcs.Observers[uint(id)]; exists {
				hc, err := data.GetHealthCheckById(hcs.Database.Pool, uint(id))
				if err != nil {
					hcs.Status = "error"
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
					hcs.Status = "error"
					return err
				}
				obj := HealthcheckObserver{
					Healthcheck: hc,
					Timer:       time.NewTimer(hc.CheckInterval),
				}
				hcs.Observers[uint(id)] = &obj
				hcs.Observers[uint(id)].Start()
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
	Healthcheck *data.Healthcheck  // The healthcheck being observed
	Timer       *time.Timer        // Timer for periodic checks
	TimerCancel context.CancelFunc // Cancel function for the timer
	TargetURLs  []string           // URLs to check, derived from the healthcheck
}

func (hco HealthcheckObserver) Start() {
	hco.Timer = time.NewTimer(hco.Healthcheck.CheckInterval)
	hco.TimerCancel = func() {
		if !hco.Timer.Stop() {
			<-hco.Timer.C // Drain the channel if the timer was already fired
		}
	}
	go func() {
		for {
			select {
			case <-hco.Timer.C:
				// Perform the healthcheck for all associated applications, on all associated servers
				for _, url := range hco.TargetURLs {
					result, err := hco.Healthcheck.PerformCheck(url)
					if err != nil {
						// Happens only if there is something wrong on the network layer
						println("Healthcheck failed:", err.Error())
					}
					println("Healthcheck result:", result)
					// TODO: Save the result to the cache
					// Reset the timer for the next check
				}
				hco.Timer.Reset(hco.Healthcheck.CheckInterval)
			}
		}
	}()
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
