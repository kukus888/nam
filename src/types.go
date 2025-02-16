package main

import (
	"time"
)

// Topology represents a Topology element in the topology yaml file
type TopologyNode struct {
	ID   uint
	Name string
	Type string // Type of Topology element, e.g. loadbalancer, firewall, etc. \n The Type corresponds to the table name in the database
}

type Proxy struct {
	ID      uint
	Ingress uint
	Egress  TopologyNode
}

type F5 struct {
	ID      uint
	Ingress TopologyNode
}

type F5Egress struct {
	ID     uint
	Egress TopologyNode
}

type Nginx struct {
	ID      uint
	Ingress TopologyNode
}

type NginxEgress struct {
	ID     uint
	Egress TopologyNode
}

// ApplicationDefinition represents the definition of an application and its general properties
type ApplicationDefinition struct {
	ID            uint
	Name          string
	Port          int
	Type          string
	HealthcheckId *uint
}

// ApplicationInstance represents an instance of an application
type ApplicationInstance struct {
	ID         uint
	Server     Server
	Definition ApplicationDefinition
}

// DTO of ApplicationInstance
type ApplicationInstanceDTO struct {
	Id           uint
	ServerId     uint
	DefinitionId uint
	Name         string
}

type Server struct {
	ID       uint
	Alias    string
	Hostname string
}

// Healthcheck represents a healthcheck element in the topology yaml file
// This is called when a healthcheck is needed
type Healthcheck struct {
	ID             uint
	Application    ApplicationDefinition
	Url            string
	Timeout        time.Duration
	Interval       time.Duration
	ExpectedStatus int
}

type ApplicationType string

var applicationTypes = []ApplicationType{
	"spring",
	"jboss",
	"hazelcast",
}

// ListApplicationTypes returns a list of all possible application types
func ListApplicationTypes() []ApplicationType {
	return applicationTypes
}

// IsValidApplicationType checks if the given application type is valid
func IsValidApplicationType(appType string) bool {
	for _, t := range applicationTypes {
		if string(t) == appType {
			return true
		}
	}
	return false
}
