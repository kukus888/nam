package services

import (
	"context"
	"crypto/tls"
	"kukus/nam/v2/layers/data"
	"log/slog"
	"runtime/pprof"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TODO: Logging

type HealthcheckService struct {
	Database            *data.Database
	ListenerConnectorHc *pgx.Conn                     // Connector for listening to healthcheck changes
	ListenerConnectorAi *pgx.Conn                     // Connector for listening to application instance changes
	Status              string                        // Status of the healthcheck service, e.g., "running", "stopped"
	Observers           map[uint]*HealthcheckObserver // Map of healthchecks being monitored with their observers (keyed by instance ID)
	Logger              *slog.Logger
	TlsConfig           *tls.Config
}

func NewHealthcheckService(database *data.Database, logger *slog.Logger, tlsConfig *tls.Config) *HealthcheckService {
	hcs := HealthcheckService{
		Database:  database,
		Observers: make(map[uint]*HealthcheckObserver),
		Logger:    logger.With("service", "HealthcheckService"),
		TlsConfig: tlsConfig,
	}
	go hcs.Start()
	return &hcs
}

func (hcs *HealthcheckService) Start() error {
	hcs.Status = "starting"
	// Start listening for notifications on healthcheck_change channel
	connHc, err := hcs.Database.Pool.Acquire(context.Background())
	if err != nil {
		hcs.Logger.Error("Failed to acquire database connection for healthcheck listener", "error", err)
		hcs.Status = "error"
		return err
	}
	hcs.ListenerConnectorHc = connHc.Conn()
	_, err = hcs.ListenerConnectorHc.Exec(context.Background(), "LISTEN healthcheck_change")
	if err != nil {
		hcs.Logger.Error("Failed to start listening for healthcheck changes", "error", err)
		hcs.Status = "error"
		return err
	}
	// Start listening for notifications on application_instance_change channel
	connAi, err := hcs.Database.Pool.Acquire(context.Background())
	if err != nil {
		hcs.Logger.Error("Failed to acquire database connection for application instance listener", "error", err)
		hcs.Status = "error"
		return err
	}
	hcs.ListenerConnectorAi = connAi.Conn()
	_, err = hcs.ListenerConnectorAi.Exec(context.Background(), "LISTEN application_instance_change")
	if err != nil {
		hcs.Logger.Error("Failed to start listening for application instance changes", "error", err)
		hcs.Status = "error"
		return err
	}
	hcs.Logger.Info("Database listeners for healthcheck and application instance changes started successfully")
	hcs.Status = "running"
	// Sync existing healthchecks from the database
	hcs.SyncObserversAll()
	go hcs.ListenForHealthcheckChanges()
	go hcs.ListenForApplicationInstanceChanges()
	hcs.Logger.Info("Healthcheck service started successfully")
	return nil
}

// Syncs all observers from the database, overwrites existing ones
func (hcs *HealthcheckService) SyncObserversAll() error {
	// first, clear existing observers
	for id := range hcs.Observers {
		if observer, exists := hcs.Observers[id]; exists {
			observer.TimerCancel() // Cancel the timer for each observer
		}
	}
	// Prepare data, fetch application instances
	ais, err := data.GetAllApplicationInstancesFull(hcs.Database.Pool)
	if err != nil {
		hcs.Logger.Error("Failed to get application instances from database", "error", err)
		hcs.UpdateStatus("error")
		return err
	}
	// Fetch healthchecks
	healthchecks, err := data.GetHealthChecksAll(hcs.Database.Pool)
	if err != nil {
		hcs.Logger.Error("Failed to get healthchecks from database", "error", err)
		hcs.UpdateStatus("error")
		return err
	}
	// Map healthchecks by ID for easy lookup
	healthcheckMap := make(map[uint]*data.Healthcheck, len(*healthchecks))
	for _, hc := range *healthchecks {
		if hc.Id != nil {
			healthcheckMap[*hc.Id] = &hc
		}
	}
	hcs.Observers = make(map[uint]*HealthcheckObserver, len(*ais))
	for _, ai := range *ais {
		if ai.ApplicationDefinition.HealthcheckId != nil && !ai.MaintenanceMode {
			hc := healthcheckMap[*ai.ApplicationDefinition.HealthcheckId]
			if hc == nil { // This should not happen, but just in case
				hcs.Logger.Warn("Healthcheck ID referenced in application definition not found", "application_definition_id", ai.ApplicationDefinition.Id, "healthcheck_id", *ai.ApplicationDefinition.HealthcheckId)
				continue
			}
			hcs.NewObserver(&ai, hc)
		}
	}
	return nil
}

// Listens for changes to application instances and updates observers accordingly
func (hcs *HealthcheckService) ListenForApplicationInstanceChanges() {
	pprof.SetGoroutineLabels(context.WithValue(context.Background(), "component", "healthcheck_service_application_instance_listener"))
	log := hcs.Logger.With("listener", "application_changes")
	for {
		// Wait for notifications
		notification, err := hcs.ListenerConnectorAi.WaitForNotification(context.Background())
		/*
			Notifications that can happen:
			INSERT:{ID} - new application instance created
			UPDATE:{ID} - existing application instance (or its dependencies, such as server/definition) updated
			DELETE:{ID} - existing application instance deleted
		*/
		if err != nil {
			log.Error("Failed to receive notification", "error", err)
			hcs.Status = "error"
			return
		}
		if notification != nil {
			// Handle the notification (e.g., log it, update internal state, etc.)
			log.Info("Received application instance change notification", "payload", notification.Payload)
			// Sync the observers based on the notification payload
			arr := strings.Split(notification.Payload, ":")
			if len(arr) != 2 { // Invalid payload, shouldnt happen
				log.Warn("Invalid notification payload format", "payload", notification.Payload)
				continue
			}
			operation := arr[0]
			idStr := arr[1]
			id, err := strconv.Atoi(idStr)
			if err != nil {
				log.Warn("Invalid application instance ID in notification payload", "payload", notification.Payload)
				continue
			}
			switch operation {
			case "INSERT":
				// Inserted new application instance. Create observer if it has a healthcheck
				ai, err := data.GetApplicationInstanceFullById(hcs.Database.Pool, uint64(id))
				if err != nil {
					log.Error("Failed to get new application instance from database", "application instance_id", id, "error", err)
					continue
				}
				if ai == nil {
					log.Warn("New application instance not found in database", "application instance_id", id)
					continue
				}
				if ai.ApplicationDefinition.HealthcheckId != nil || ai.MaintenanceMode {
					hc, err := data.GetHealthCheckById(hcs.Database.Pool, *ai.ApplicationDefinition.HealthcheckId)
					if err != nil {
						log.Error("Failed to get healthcheck for new application instance from database", "application instance_id", id, "healthcheck_id", *ai.ApplicationDefinition.HealthcheckId, "error", err)
						continue
					}
					if hc == nil {
						log.Warn("Healthcheck for new application instance not found in database", "application instance_id", id, "healthcheck_id", *ai.ApplicationDefinition.HealthcheckId)
						continue
					}
					hcs.NewObserver(ai, hc)
					// New application instance inserted
					log.Info("Application instance inserted", "payload", notification.Payload)
				}
				continue
			case "UPDATE":
				// Updated existing application instance. Need to update the observer for this application instance
				ai, err := data.GetApplicationInstanceFullById(hcs.Database.Pool, uint64(id))
				if err != nil {
					log.Error("Failed to get updated application instance", "application instance_id", id, "error", err)
					continue
				}
				if ai == nil {
					log.Warn("Updated application instance not found in database", "application instance_id", id)
					continue
				}
				if ai.ApplicationDefinition.HealthcheckId == nil {
					// No healthcheck assigned, just remove existing observer if exists
					if observer, exists := hcs.Observers[uint(id)]; exists {
						observer.TimerCancel()
						delete(hcs.Observers, uint(id))
					}
					continue
				}
				hc, err := data.GetHealthCheckById(hcs.Database.Pool, *ai.ApplicationDefinition.HealthcheckId)
				if err != nil {
					log.Error("Failed to get healthcheck for updated application instance from database", "application instance_id", id, "healthcheck_id", *ai.ApplicationDefinition.HealthcheckId, "error", err)
					continue
				}
				if hc == nil {
					log.Warn("Healthcheck for updated application instance not found in database", "application instance_id", id, "healthcheck_id", *ai.ApplicationDefinition.HealthcheckId)
					// Just remove existing observer if exists
					if observer, exists := hcs.Observers[uint(id)]; exists {
						observer.TimerCancel()
						delete(hcs.Observers, uint(id))
					}
					continue
				}
				// Update or create observer for the application instance
				if observer, exists := hcs.Observers[uint(id)]; exists {
					// Remove existing observer
					observer.TimerCancel()
					delete(hcs.Observers, uint(id))
				}
				if !ai.MaintenanceMode {
					// Create new observer
					hcs.NewObserver(ai, hc)
				}
				// Existing application instance updated
				log.Info("Healthcheck updated", "payload", notification.Payload)
			case "DELETE":
				// Deleted existing application instance. Need to stop and remove the observer that monitor this application instance
				// Remove observers for the application instance
				if observer, exists := hcs.Observers[uint(id)]; exists {
					observer.TimerCancel() // Cancel the timer for the observer
					delete(hcs.Observers, uint(id))
				}
				// Existing application instance deleted
				log.Info("Application instance deleted", "payload", notification.Payload)
			}
		}
	}
}

// Listens for notifications from the database about healthcheck changes
func (hcs *HealthcheckService) ListenForHealthcheckChanges() {
	pprof.SetGoroutineLabels(context.WithValue(context.Background(), "component", "healthcheck_service_healthcheck_listener"))
	log := hcs.Logger.With("listener", "healthcheck_changes")
	for {
		// Wait for notifications
		notification, err := hcs.ListenerConnectorHc.WaitForNotification(context.Background())
		/*
			Notifications that can happen:
			INSERT:{ID} - new healthcheck created
			UPDATE:{ID} - existing healthcheck updated
			DELETE:{ID} - existing healthcheck deleted
		*/
		if err != nil {
			log.Error("Failed to receive notification", "error", err)
			hcs.Status = "error"
			return
		}
		if notification != nil {
			// Handle the notification (e.g., log it, update internal state, etc.)
			log.Info("Received healthcheck change notification", "payload", notification.Payload)
			// Sync the observers based on the notification payload
			arr := strings.Split(notification.Payload, ":")
			if len(arr) != 2 { // Invalid payload, shouldnt happen
				log.Warn("Invalid notification payload format", "payload", notification.Payload)
				continue
			}
			operation := arr[0]
			idStr := arr[1]
			id, err := strconv.Atoi(idStr)
			if err != nil {
				log.Warn("Invalid healthcheck ID in notification payload", "payload", notification.Payload)
				continue
			}
			switch operation {
			case "INSERT":
				// Inserted new healthcheck. No need to create observers, because no application instance is linked to it yet
				continue
			case "UPDATE":
				// Updated existing healthcheck. Need to update all observers that monitor this healthcheck
				hc, err := data.GetHealthCheckById(hcs.Database.Pool, uint(id))
				if err != nil {
					log.Error("Failed to get updated healthcheck from database", "healthcheck_id", id, "error", err)
					continue
				}
				if hc == nil {
					log.Warn("Updated healthcheck not found in database", "healthcheck_id", id)
					continue
				}
				ais, err := data.GetAllApplicationInstancesFullByHealthcheckId(hcs.Database.Pool, uint64(id))
				if err != nil {
					log.Error("Failed to get application instances for updated healthcheck", "healthcheck_id", id, "error", err)
					continue
				}
				// Recreate observers for each application instance
				for _, ai := range *ais {
					if observer, exists := hcs.Observers[uint(id)]; exists {
						// Delete existing observer
						observer.TimerCancel()
						delete(hcs.Observers, uint(id))
					}
					// Create new observer
					hcs.NewObserver(&ai, hc)
				}
				// Existing healthcheck updated
				log.Info("Healthcheck updated", "payload", notification.Payload)
			case "DELETE":
				// Deleted existing healthcheck. Need to stop and remove all observers that monitor this healthcheck
				// Get instances using this healthcheck
				ais, err := data.GetAllApplicationInstancesFullByHealthcheckId(hcs.Database.Pool, uint64(id))
				if err != nil {
					log.Error("Failed to get application instances for deleted healthcheck", "healthcheck_id", id, "error", err)
					continue
				}
				// Remove observers for each application instance
				for _, ai := range *ais {
					if observer, exists := hcs.Observers[uint(ai.Id)]; exists {
						observer.TimerCancel() // Cancel the timer for the observer
						delete(hcs.Observers, uint(ai.Id))
					}
				}
				// Existing healthcheck deleted
				log.Info("Healthcheck deleted", "payload", notification.Payload)
			}
		}
	}
}

// Observer is a struct that monitors a specific healthcheck and its associated application instance.
// Each instance has its own observer.
type HealthcheckObserver struct {
	ApplicationInstance *data.ApplicationInstanceFull // The application instance being monitored
	Healthcheck         *data.Healthcheck             // The healthcheck being observed
	Timer               *time.Timer                   // Timer for periodic checks
	TimerCancel         context.CancelFunc            // Cancel function for the timer
	TargetUrl           string                        // URL to check
	Context             context.Context               // Context for managing goroutines
	DbPool              *pgxpool.Pool                 // Database connection pool
	ProbeFunc           func()                        // Function to perform the healthcheck probe
	Logger              *slog.Logger
	TlsConfig           *tls.Config
}

// Creates and starts a new observer for the given application instance and healthcheck
func (hcs *HealthcheckService) NewObserver(ai *data.ApplicationInstanceFull, hc *data.Healthcheck) *HealthcheckObserver {
	observer := HealthcheckObserver{
		ApplicationInstance: ai,
		Healthcheck:         hc,
		Timer:               time.NewTimer(hc.CheckInterval),
	}
	// Get target url together
	observer.TargetUrl = hc.Protocol + "://" + ai.Server.Hostname + ":" + strconv.Itoa(int(ai.ApplicationDefinition.Port)) + hc.ReqUrl
	observer.Logger = hcs.Logger.With("application_instance_id", ai.Id, "application_instance_name", ai.Name, "healthcheck_id", *hc.Id, "healthcheck_name", hc.Name)
	if hc.Protocol == "https" {
		observer.TlsConfig = hcs.TlsConfig.Clone()
	}
	observer.Context = context.WithValue(context.Background(), "component", "healthcheck_observer")
	observer.Context = context.WithValue(observer.Context, "application_instance_id", ai.Id)
	observer.Context = context.WithValue(observer.Context, "healthcheck_id", *hc.Id)
	// Start the observer
	observer.Start(hcs.Database.Pool)
	hcs.Observers[ai.Id] = &observer
	hcs.Logger.Debug("Healthcheck observer started", "id", *hc.Id, "name", hc.Name, "for_application_instance_id", ai.Id, "for_application_instance_name", ai.Name)
	return &observer
}

func (hco *HealthcheckObserver) Start(pool *pgxpool.Pool) {
	hco.DbPool = pool
	hco.Timer = time.NewTimer(hco.Healthcheck.CheckInterval)
	hco.TimerCancel = func() {
		cancelCtx := context.WithValue(context.Background(), "component", "healthcheck_observer_timer_cancel")
		pprof.SetGoroutineLabels(cancelCtx)
		if !hco.Timer.Stop() {
			<-hco.Timer.C // Drain the channel if the timer was already fired
		}
	}
	// Set up the healthcheck probe function
	hco.ProbeFunc = func() {
		// Set goroutine labels for better profiling
		pprof.SetGoroutineLabels(hco.Context)
		// Perform the healthcheck
		result, err := hco.Healthcheck.PerformCheck(hco.TargetUrl, hco.TlsConfig)
		result.ApplicationInstanceID = hco.ApplicationInstance.Id
		if err != nil {
			// Happens only if there is something wrong on the network layer
			hco.Logger.Debug("Healthcheck failed", "instance_id", hco.ApplicationInstance.Id, "url", hco.TargetUrl, "error", err.Error())
		}
		hco.Logger.Debug("Healthcheck result", "instance_id", hco.ApplicationInstance.Id, "is_successful", result.IsSuccessful, "status", result.ResStatus, "response_time", result.ResTime)
		// Insert the result into the database
		_, err = result.DbInsert(hco.DbPool)
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
	if hcs.ListenerConnectorHc != nil {
		// Unlisten and close the connection
		_, err := hcs.ListenerConnectorHc.Exec(context.Background(), "UNLISTEN healthcheck_changes")
		if err != nil {
			hcs.Status = "error"
			return err
		}
		hcs.ListenerConnectorHc.Close(context.Background())
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
