package main

import (
	"fmt"
)

var AppConfiguration ApplicationConfiguration

func main() {
	// Load the application configuration and start vital components. Failure to start results in a panic.
	AppConfiguration = LoadAndParseConfiguration("config.yaml")
	DbStart()
	//DB.InsertServer(Server{Hostname: "testhostname01", Alias: "Testalias01"})
	//DB.InsertServer(Server{Hostname: "testhostname02", Alias: "Testalias02"})

	InitWebServer()

	// Rundeck POC
	rdckCli := NewRundeckClient("http://localhost", "cN3EWNUG8rT4n5YAQLtwOPSX2gWpSuzQ")
	info, err := rdckCli.GetSystemInfo()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Rundeck System Info: %+v\n", info)
}
