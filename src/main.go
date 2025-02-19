package main

import (
	"fmt"
	data "kukus/nam/v2/layers/data"

	"github.com/gin-gonic/gin"
)

type Application struct {
	Engine        *gin.Engine
	Database      *data.Database
	Configuration ApplicationConfiguration
}

var App Application

func main() {
	// Load the application configuration and start vital components. Failure to start results in a panic.
	App.Configuration = LoadAndParseConfiguration("config.yaml")

	// Initialise pgx database TODO: pgx dsn from yaml
	db, err := data.NewDatabase("postgres://postgres:heslo123@localhost:5432/postgres")
	if err != nil {
		panic("Unable to initialise database connection: " + err.Error())
	}
	App.Database = db

	// Init WS
	InitWebServer(App)

	// Rundeck POC
	rdckCli := NewRundeckClient("http://localhost", "cN3EWNUG8rT4n5YAQLtwOPSX2gWpSuzQ")
	info, err := rdckCli.GetSystemInfo()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Rundeck System Info: %+v\n", info)
}
