package services

import (
	"errors"
	"log/slog"
)

// Contains basic service definition and functions for service management.

type Service interface {
	GetName() string
	GetDescription() string
	GetStatus() string
	Start() error
	Stop() error
	IsRunning() bool
}

type ServiceManager struct {
	services map[string]Service
	logger   slog.Logger
}

func NewServiceManager(logger slog.Logger) *ServiceManager {
	return &ServiceManager{
		services: make(map[string]Service),
		logger:   logger,
	}
}
func (sm *ServiceManager) RegisterService(svc Service) {
	sm.services[svc.GetName()] = svc
}
func (sm *ServiceManager) StartService(name string) error {
	svc, exists := sm.services[name]
	if !exists {
		return nil // or return an error if you prefer
	}
	if svc.IsRunning() {
		return nil // or return an error if you prefer
	}
	return svc.Start()
}
func (sm *ServiceManager) StopService(name string) error {
	svc, exists := sm.services[name]
	if !exists {
		return nil // or return an error if you prefer
	}
	if !svc.IsRunning() {
		return nil // or return an error if you prefer
	}
	return svc.Stop()
}
func (sm *ServiceManager) GetServiceStatus(name string) (string, error) {
	svc, exists := sm.services[name]
	if !exists {
		return "", errors.New("service not found")
	}
	return svc.GetStatus(), nil
}
