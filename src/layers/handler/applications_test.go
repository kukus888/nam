package handlers

import (
	"testing"

	"kukus/nam/v2/layers/data"
)

func TestAllInstancesInMaintenance(t *testing.T) {
	tests := []struct {
		name      string
		instances []data.ApplicationInstanceFull
		want      bool
	}{
		{name: "no instances"},
		{name: "all in maintenance", instances: []data.ApplicationInstanceFull{{MaintenanceMode: true}, {MaintenanceMode: true}}, want: true},
		{name: "none in maintenance", instances: []data.ApplicationInstanceFull{{}, {}}, want: false},
		{name: "mixed maintenance state", instances: []data.ApplicationInstanceFull{{MaintenanceMode: true}, {}}, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := allInstancesInMaintenance(tt.instances); got != tt.want {
				t.Errorf("allInstancesInMaintenance() = %t, want %t", got, tt.want)
			}
		})
	}
}
