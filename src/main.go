package main

import (
	"fmt"
	data "kukus/nam/v2/layers/data"
	services "kukus/nam/v2/layers/service"
	"time"

	"github.com/gin-gonic/gin"
)

type Application struct {
	Engine        *gin.Engine
	Database      *data.Database
	Configuration ApplicationConfiguration
	Services      *services.ServiceManager
}

var App Application

func main() {
	// Load the application configuration and start vital components. Failure to start results in a panic.
	App.Configuration = LoadAndParseConfiguration("config.yaml")

	// Initialise pgx database TODO: pgx dsn from yaml
	db, err := data.NewDatabase("postgres://postgres:heslo123@localhost:5432/postgres?application_name=nam")
	if err != nil {
		panic("Unable to initialise database connection: " + err.Error())
	}
	App.Database = db

	// Init WS
	go InitWebServer(&App)

	// Init Services
	App.Services = services.NewServiceManager()
	healthcheckService := services.NewHealthcheckService(App.Database)
	App.Services.RegisterService(healthcheckService)
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
