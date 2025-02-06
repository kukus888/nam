package main

import (
	"time"
)

// Topology represents a Topology element in the topology yaml file
type TopologyNode struct {
	ID   uint `gorm:"primaryKey"`
	Name string
	Type string // Type of Topology element, e.g. loadbalancer, firewall, etc. \n The Type corresponds to the table name in the database
}

type Proxy struct {
	ID      uint         `gorm:"primaryKey;References:id"`
	Ingress uint         `gorm:"foreignKey:id;References:id"`
	Egress  TopologyNode `gorm:"foreignKey:id;References:id"`
}

type F5 struct {
	ID      uint         `gorm:"primaryKey;References:id"`
	Ingress TopologyNode `gorm:"foreignKey:id;References:id"`
}

type F5Egress struct {
	ID     uint         `gorm:"primaryKey;References:id"`
	Egress TopologyNode `gorm:"foreignKey:ID;References:id"`
}

type Nginx struct {
	ID      uint         `gorm:"primaryKey;References:id"`
	Ingress TopologyNode `gorm:"foreignKey:id;References:id"`
}

type NginxEgress struct {
	ID     uint         `gorm:"primaryKey;References:id"`
	Egress TopologyNode `gorm:"foreignKey:id;References:id"`
}

// ApplicationDefinition represents the definition of an application and its general properties
type ApplicationDefinition struct {
	ID   uint   `yaml:"id" gorm:"primaryKey"`
	Name string `yaml:"name"`
	Port int    `yaml:"port"`
	Type string `yaml:"type"`
}

// ApplicationInstance represents an instance of an application
type ApplicationInstance struct {
	ID         uint                  `gorm:"primaryKey"`
	Server     Server                `gorm:"foreignKey:ID"`
	Definition ApplicationDefinition `gorm:"foreignKey:ID"`
}

type Server struct {
	ID       uint   `gorm:"primaryKey"`
	Name     string `yaml:"name"`
	Hostname string `yaml:"hostname" gorm:"unique"`
}

// Healthcheck represents a healthcheck element in the topology yaml file
// This is called when a healthcheck is needed
type Healthcheck struct {
	ID             uint                  `gorm:"primaryKey"`
	Application    ApplicationDefinition `gorm:"foreignKey:ID"`
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
