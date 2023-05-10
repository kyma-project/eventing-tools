package payload

type DTO struct {
	Start string `json:"StartTime"`
	Value int    `json:"Value"`
}

type LegacyEvent struct {
	Data             DTO    `json:"Data"`
	EventType        string `json:"Event-Type"`
	EventTypeVersion string `json:"Event-Type-Version"`
	EventTime        string `json:"Event-Time"`
	EventTracing     bool   `json:"Event-Tracing"`
}
