package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	data "kukus/nam/v2/layers/data"
	services "kukus/nam/v2/layers/service"
	"log/slog"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

type Application struct {
	Engine        *gin.Engine
	Database      *data.Database
	Configuration ApplicationConfiguration
	Services      *services.ServiceManager
	TlsConfig     *tls.Config
}

var App Application

func main() {
	// Load flags
	configFile := flag.String("config", "config.yaml", "Path to the configuration file")
	flag.Parse()

	// Load the application configuration and start vital components. Failure to start results in a panic.
	appCfg, err := LoadAndParseConfiguration(*configFile)
	if err != nil {
		panic("Unable to load configuration: " + err.Error())
	} else {
		App.Configuration = *appCfg
		fmt.Println("Configuration loaded successfully")
	}
	tlsConfig, err := ParseSecrets(App.Configuration)
	if err != nil {
		panic("Unable to parse secrets! " + err.Error())
	}
	App.TlsConfig = tlsConfig

	// Init logging
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: App.Configuration.Logging.SlogLevel,
	}))
	log.Info("Successfully initialised configuration and logging")

	log.Debug("Initializing postgres database connection")
	// Initialise database connection
	db, err := data.NewDatabase(App.Configuration.Database.Dsn)
	if err != nil {
		panic("Unable to initialise database connection: " + err.Error())
	}
	// Try connecting to the database
	if err := db.Pool.Ping(context.Background()); err != nil {
		panic("Unable to connect to the database: " + err.Error())
	}
	// Try migrating the schema
	if err = data.AutoMigrate(appCfg.Database.Dsn); err != nil {
		panic("Unable to migrate database schema: " + err.Error())
	}
	App.Database = db
	log.Info("Successfully initialised database connection and migrated to latest schema")

	if App.Configuration.WebServer.Enabled {
		// Start web server
		go InitWebServer(&App)
	} else {
		log.Warn("Web server is disabled in the configuration, skipping web server initialization")
	}

	// Init Services
	slog.Debug("Initializing services")
	App.Services = services.NewServiceManager(*log)
	if enabled, found := App.Configuration.Services["HealthcheckService"]; enabled && found {
		slog.Info("HealthcheckService is enabled, initializing")
		healthcheckService := services.NewHealthcheckService(App.Database, log.With("component", "HealthcheckService"), App.TlsConfig)
		App.Services.RegisterService(healthcheckService)
	}
	for {
		status, _ := App.Services.GetServiceStatus("HealthcheckService")
		fmt.Println("HealthcheckService status:", status)
		time.Sleep(1 * time.Minute)
	}
	// Rundeck POC
	/*rdckCli := NewRundeckClient("http://localhost", "cN3EWNUG8rT4n5YAQLtwOPSX2gWpSuzQ")
	info, err := rdckCli.GetSystemInfo()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Rundeck System Info: %+v\n", info)*/
}
