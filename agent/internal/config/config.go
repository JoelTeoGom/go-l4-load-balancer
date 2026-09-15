package config

import "os"

type Config struct {
	lbAddr      string
	servicePort string
}

func NewConfig() *Config {
	return &Config{
		lbAddr:      os.Getenv("LB_ADDRESS"),
		servicePort: os.Getenv("SVC_PORT"),
	}
}

func (c *Config) LbAddress() string {
	return c.lbAddr
}

func (c *Config) SvcPort() string {
	return c.servicePort
}
