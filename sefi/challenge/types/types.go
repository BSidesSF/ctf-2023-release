package types

import (
	_ "encoding/json"
)

type ClientConfig struct {
	APIEndpoint string `json:"api_endpoint"`
	RequestKey  string `json:"request_key"`
	ClientID    string `json:"client_id"`
}

type Sample struct {
	Freqs     int    `json:"samples_per_timeslot"`
	TimeSlots int    `json:"time_slots"`
	Samples   []byte `json:"samples"`
	StartTime int    `json:"start_time"`
}

type WorkUnitRequest struct {
	ClientID      string `json:"client_id"`
	UnitsFinished int    `json:"units_finished"`
}

type WorkUnitResponse struct {
	ClientID string  `json:"client_id"`
	WorkUnit *Sample `json:"work_unit"`
}
