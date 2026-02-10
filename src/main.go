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
	"gopkg.in/natefinch/lumberjack.v2"
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
	dbVersions := flag.String("db", "", "Database migration tool: versions, drop, newschema, raw")
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

	// Pivot to db migration micro-tool
	if *dbVersions != "" {
		switch *dbVersions {
		case "versions":
			data.DbMigrationTool(appCfg.Database.Dsn)
		case "drop":
			data.DropEverything(appCfg.Database.Dsn)
		case "newschema":
			data.NewSchema(appCfg.Database.Dsn)
		}
		os.Exit(0) // Exit after running the migration tool
	}

	// Init logging
	rotlog := &lumberjack.Logger{
		Filename:   App.Configuration.Logging.FilePath,
		MaxSize:    App.Configuration.Logging.MaxSize,
		MaxBackups: App.Configuration.Logging.MaxBackups,
		MaxAge:     App.Configuration.Logging.MaxAge,
		Compress:   App.Configuration.Logging.Compress,
	}
	log := slog.New(slog.NewJSONHandler(rotlog, &slog.HandlerOptions{
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
	log.Debug("Initializing services")
	App.Services = services.NewServiceManager(*log)
	if enabled, found := App.Configuration.Services["HealthcheckService"]; enabled && found {
		log.Info("HealthcheckService is enabled, initializing")
		healthcheckService := services.NewHealthcheckService(App.Database, log.With("component", "HealthcheckService"), App.TlsConfig)
		App.Services.RegisterService(healthcheckService)
	}
	for {
		status, _ := App.Services.GetServiceStatus("HealthcheckService")
		fmt.Println("HealthcheckService status:", status)
		time.Sleep(1 * time.Minute)
	}
}
