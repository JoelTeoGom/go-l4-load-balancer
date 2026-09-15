package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os/exec"
	"strings"

	"github.com/JoelTeoGom/go-l4-load-balancer/agent/internal/config"
	"github.com/JoelTeoGom/go-l4-load-balancer/agent/internal/registry"
)

func main() {
	ctx := context.Background()

	cfg := config.NewConfig()
	hostName, err := initSetup(cfg.LbAddress(), cfg.SvcPort())
	if err != nil {
		panic(err)
	}

	registry := registry.NewRegistry(cfg.LbAddress(), hostName)
	err = registry.ConnectToCtrlPlane(ctx, cfg.LbAddress(), hostName)
	if err != nil {
		panic(err)
	}

}

func initSetup(target, servicePort string) (string, error) {
	//0. Setting up TCP forwarding to 1
	cmd := exec.Command("sysctl", "-w", "net.ipv4.ip_forward=1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("sysctl: %w: %s", err, out)
	}

	//1. Setting up PREROUTING jump TO KUBE-SERVICES
	rule := "-t nat -A PREROUTING -j KUBE-SERVICES"
	args := strings.Fields(rule) // ["-t","nat","-A","PREROUTING","-j","KUBE-SERVICES"]
	cmd = exec.Command("iptables", args...)
	cmd.CombinedOutput()

	//2.Setting up
	localIP, err := outboundIP(target)
	if err != nil {
		fmt.Println("Lb not up!!!")
		panic(err)
	}

	rule = fmt.Sprintf("-A KUBE-SERVICES -d %s -p tcp --dport %s -j KUBE-SVC", localIP, servicePort)
	args = strings.Fields(rule)
	cmd = exec.Command("iptables", args...)
	cmd.CombinedOutput()

	//3. hostName
	hostName, err := exec.Command("hostname").Output()
	if err != nil {
		log.Fatal(err)
	}

	return string(hostName), nil
}

// quick udp connection to obtain local address quick
func outboundIP(target string) (net.IP, error) {
	conn, err := net.Dial("udp", target)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	return conn.LocalAddr().(*net.UDPAddr).IP, nil
}

//1. Node up -> service systemmd   ---->>>>>> up agent
//1.1 Settup Node SERVICE_AND_NETWORK RULES
//2. Agent ->   ctrlplane_url/node/register
//3. Metrics ----> 5 / 10 secs  ctrlplane_url/node/health
//4. Event   ----> worker escuchando SSE en un channel i ejecuta ordenes create_pod, remove_pod ... ping health

//if pod_create --->>

//apaga   ---> ctrlplane_url/node/unregister

// Cómo lo tiene kube-proxy

// Tres niveles de cadenas. Te lo pongo con un servicio en 10.96.0.10:80 y tres pods.

// Entrada. En PREROUTING se desvía todo a una cadena propia:

// -A PREROUTING -j KUBE-SERVICES

// Por servicio. Una regla que captura el destino y salta a la cadena de ese servicio:

// -A KUBE-SERVICES -d 10.96.0.10/32 -p tcp --dport 80 -j KUBE-SVC-XPGD46QRK7WJZT7O

// El reparto. Aquí está lo que te interesa:

// -A KUBE-SVC-XPGD46QRK7WJZT7O -m statistic --mode random --probability 0.33333333349 -j KUBE-SEP-AAAA
// -A KUBE-SVC-XPGD46QRK7WJZT7O -m statistic --mode random --probability 0.50000000000 -j KUBE-SEP-BBBB
// -A KUBE-SVC-XPGD46QRK7WJZT7O -j KUBE-SEP-CCCC

// Ahí ves lo de las probabilidades acumulativas: 1/3, luego 1/2 de los dos que quedan, y la última sin condición. Cada una reparte un tercio del total.

// Y el destino. Una cadena por endpoint, con el DNAT:

// -A KUBE-SEP-AAAA -p tcp -j DNAT --to-destination 10.244.1.5:8080
