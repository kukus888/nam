package services

import (
	"kukus/nam/v2/layers/data"
	"time"
)

type HealthcheckService struct {
	Database *data.Database
}

// Healthcheck business object
type Healthcheck struct {
	ID             uint
	Url            string
	Timeout        time.Duration
	CheckInterval  time.Duration
	ExpectedStatus int
}
