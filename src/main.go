package main

import (
	"fmt"
)

var ObservedTopology Topology

var AppConfiguration ApplicationConfiguration

type Topology interface{}

func main() {
	// Load the application configuration and start vital components. Failure to start results in a panic.
	AppConfiguration = LoadAndParseConfiguration("config.yaml")
	DB := Database{}
	DB.Start()
	//DB.InsertServer(Server{Hostname: "testhostname01", Alias: "Testalias01"})
	//DB.InsertServer(Server{Hostname: "testhostname02", Alias: "Testalias02"})
	s, e := DB.QueryServerAll()
	if e != nil {
		panic(e)
	}
	fmt.Printf("%s %s", s[0].Alias, s[0].Hostname)
	//InitWebServer()

	// Rundeck POC
	rdckCli := NewRundeckClient("http://localhost", "cN3EWNUG8rT4n5YAQLtwOPSX2gWpSuzQ")
	info, err := rdckCli.GetSystemInfo()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Rundeck System Info: %+v\n", info)
}
