package config

import (
	"net"
	"os"
)

const (
	defaultCtrlPlanePort = "9000"
	defaultDataPlanePort = "8080"
	defaultEtcdDir       = "./etcd-data"
)

type Config struct {
	ctrlPlaneIP   string
	ctrlPlanePort string
	dataPlaneIP   string
	dataPlanePort string
	etcdDir       string
}

func NewConfig() *Config {
	return &Config{
		ctrlPlaneIP:   os.Getenv("CTRL_PLANE_IP"),
		ctrlPlanePort: getEnvOrDefault("CTRL_PLANE_PORT", defaultCtrlPlanePort),
		dataPlaneIP:   os.Getenv("DATA_PLANE_IP"),
		dataPlanePort: getEnvOrDefault("DATA_PLANE_PORT", defaultDataPlanePort),
		etcdDir:       getEnvOrDefault("ETCD_DIR", defaultEtcdDir),
	}
}

func (c *Config) CtrlPlaneIP() string {
	return c.ctrlPlaneIP
}

func (c *Config) CtrlPlanePort() string {
	return c.ctrlPlanePort
}

// CtrlPlaneAddress returns the "ip:port" the control plane listens on.
func (c *Config) CtrlPlaneAddress() string {
	return net.JoinHostPort(c.ctrlPlaneIP, c.ctrlPlanePort)
}

func (c *Config) DataPlaneIP() string {
	return c.dataPlaneIP
}

func (c *Config) DataPlanePort() string {
	return c.dataPlanePort
}

// DataPlaneAddress returns the "ip:port" the data plane listens on.
func (c *Config) DataPlaneAddress() string {
	return net.JoinHostPort(c.dataPlaneIP, c.dataPlanePort)
}

func (c *Config) EtcdDir() string {
	return c.etcdDir
}

func getEnvOrDefault(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
