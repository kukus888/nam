package main

import (
	"fmt"

	"gorm.io/gorm"
)

var ObservedTopology Topology

var AppConfiguration ApplicationConfiguration
var Database *gorm.DB

type Topology interface{}

func main() {
	// Load the application configuration and start vital components. Failure to start results in a panic.
	AppConfiguration = LoadAndParseConfiguration("config.yaml")
	DbStart()

	Database.AutoMigrate(
		&TopologyNode{},
		&Proxy{},
		&F5{},
		&F5Egress{},
		&Nginx{},
		&NginxEgress{},
		&ApplicationDefinition{},
		&Server{},
		&ApplicationInstance{},
	)
	//InitWebServer()

	// Rundeck POC
	rdckCli := NewRundeckClient("http://localhost", "cN3EWNUG8rT4n5YAQLtwOPSX2gWpSuzQ")
	info, err := rdckCli.GetSystemInfo()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Rundeck System Info: %+v\n", info)
}
