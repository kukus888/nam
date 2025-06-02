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
		logger:   *logger.With("component", "ServiceManager"),
	}
}
func (sm *ServiceManager) RegisterService(svc Service) {
	sm.services[svc.GetName()] = svc
	sm.logger.Debug("Service registered", "name", svc.GetName())
}
func (sm *ServiceManager) StartService(name string) error {
	svc, exists := sm.services[name]
	if !exists {
		sm.logger.Warn("Error starting service: Service not found", "name", name)
		return nil
	}
	if svc.IsRunning() {
		sm.logger.Debug("Error starting service: Service already running", "name", name)
		return nil
	}
	return svc.Start()
}
func (sm *ServiceManager) StopService(name string) error {
	svc, exists := sm.services[name]
	if !exists {
		sm.logger.Debug("Error stopping service: Wanted to stop service, but it does not exist", "name", name)
		return nil
	}
	if !svc.IsRunning() {
		sm.logger.Debug("Error stopping service: Service is not running, nothing to stop", "name", name)
		return nil
	}
	return svc.Stop()
}
func (sm *ServiceManager) GetServiceStatus(name string) (string, error) {
	svc, exists := sm.services[name]
	if !exists {
		return "", errors.New("Error getting service status: Service not found: " + name)
	}
	return svc.GetStatus(), nil
}
