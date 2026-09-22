package config

import "os"

type Config struct {
	controlPlaneURL string
	agentPort       string
}

func NewConfig() *Config {
	return &Config{
		controlPlaneURL: os.Getenv("LB_ADDRESS"),
		agentPort:       os.Getenv("AGENT_PORT"),
	}
}

// ControlPlaneURL returns the orchestrator control plane base URL (ex: http://192.168.1.50:9000).
func (c *Config) ControlPlaneURL() string {
	return c.controlPlaneURL
}

func (c *Config) AgentPort() string {
	return c.agentPort
}
