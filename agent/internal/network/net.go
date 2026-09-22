package network

import (
	"fmt"
	"net"
	"net/url"
	"os"
)

func ObtainOutboundIP() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()
	return conn.LocalAddr().(*net.UDPAddr).IP.String(), nil
}

func ObtainLoadbalancerIP(target string) (string, error) {
	parsedURL, err := url.Parse(target)
	if err != nil {
		return "", err
	}

	host := parsedURL.Hostname()
	if net.ParseIP(host) == nil {
		return "", fmt.Errorf("target %q does not contain a valid ip", target)
	}

	return host, nil
}

func ResolveHostname() (string, error) {
	return os.Hostname()
}
