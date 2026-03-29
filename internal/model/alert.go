package model

import "time"

type Alert struct {
	ID        string    `json:"id"`
	Source    string    `json:"source"`    // kj90x/manual/ai
	Level     int       `json:"level"`     // 1-5
	Location  string    `json:"location"`
	GasType   string    `json:"gas_type"`
	Value     float64   `json:"value"`
	Status    string    `json:"status"`    // open/acknowledged/resolved
	CreatedAt time.Time `json:"created_at"`
}

type Command struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`      // broadcast/call/sms/ws
	Target    string    `json:"target"`
	Content   string    `json:"content"`
	Status    string    `json:"status"`    // pending/executing/done/failed
	CreatedAt time.Time `json:"created_at"`
}
