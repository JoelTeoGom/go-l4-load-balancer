package config

import "os"

type Config struct {
	address string
}

func NewConfig() *Config {
	return &Config{
		address: os.Getenv("LB_ADDRESS"),
	}
}

func (c *Config) Address() string {
	return c.address
}
