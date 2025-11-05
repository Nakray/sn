package config

type Config struct {
	Database        string            `json:"database"`
	Server          ServerConfig      `json:"server"`
	Monitoring      MonitoringConfig  `json:"monitoring"`
	VK              VKConfig          `json:"vk"`
	RelevanceHours  int               `json:"relevance_hours"`
}

type ServerConfig struct {
	Port int `json:"port"`
}

type MonitoringConfig struct {
	IntervalMinutes int `json:"interval_minutes"`
	Workers         int `json:"workers"`
}

type VKConfig struct {
	APIVersion string `json:"api_version"`
}
