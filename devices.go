package zigbee2mqtt

type CommonDevice struct {
	LinkQuality float64 `json:"linkquality"`
	Battery     float64 `json:"battery"`
	Voltage     float64 `json:"voltage"`
}

type SonoffContactSensor struct {
	CommonDevice
	Contact    bool `json:"contact"`
	Tamper     bool `json:"contact"`
	BatteryLow bool `json:"battery_low"`
}

type AqaraMotionSensor struct {
	CommonDevice
	Occupancy         bool    `json:"occupancy"`
	TriggerIndicator  bool    `json:"trigger_indicator"`
	MotionSensitity   string  `json:"motion_sensitivity"`
	DetectionInterval float64 `json:"detection_interval"`
	TemperatureC      float64 `json:"device_temperature"`
	Illuminance       float64 `json:"illuminance"`
	PowerOutageCount  int     `json:"power_outage_count"`
}

type SonoffTemperatureSensor struct {
	CommonDevice
	Humidity     float64 `json:"humidity"`
	TemperatureC float64 `json:"temperature"`
}

type ThirdRealityMotionSensor struct {
	CommonDevice
	Occupancy  bool `json:"occupancy"`
	Tamper     bool `json:"tamper"`
	BatteryLow bool `json:"battery_low"`
}
