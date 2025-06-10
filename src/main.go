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
	"strconv"
	"strings"
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
	// Initialise pgx database
	pgxConnStr := "postgres://" + App.Configuration.Database.User + ":" + App.Configuration.Database.Password + "@" + App.Configuration.Database.Host + ":" + strconv.Itoa(App.Configuration.Database.Port) + "/" + App.Configuration.Database.Name
	// Initialise database connection
	db, err := data.NewDatabase(pgxConnStr + "?application_name=nam")
	if err != nil {
		panic("Unable to initialise database connection: " + err.Error())
	}
	// Try connecting to the database
	if err := db.Pool.Ping(context.Background()); err != nil {
		panic("Unable to connect to the database: " + err.Error())
	}
	App.Database = db
	pgxConnStrSafe := strings.Replace(pgxConnStr, App.Configuration.Database.Password, "*****", -1) // Hide password in logs
	log.Info("Successfully initialised database connection", "dsn", pgxConnStrSafe)

	if App.Configuration.WebServer.Enabled {
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
