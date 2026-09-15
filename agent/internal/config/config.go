package config

import "os"

type Config struct {
	lbAddr    string
	agentPort string
}

func NewConfig() *Config {
	return &Config{
		lbAddr:    os.Getenv("LB_ADDRESS"),
		agentPort: os.Getenv("AGENT_PORT"),
	}
}

func (c *Config) LbAddress() string {
	return c.lbAddr
}

func (c *Config) AgentPort() string {
	return c.agentPort
}
