package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptrace"
	"time"

	"github.com/JoelTeoGom/go-l4-load-balancer/agent/internal/config"
)

func main() {

	cfg := config.NewConfig()
	client := http.Client{}
	address := fmt.Sprintf("http://%s:9000/register-node", cfg.Address())
	req, err := http.NewRequest("POST", address, nil)
	if err != nil {
		log.Fatalf("Request creation failed: %v", err)
	}

	fmt.Println(address)
	start := time.Now()
	trace := &httptrace.ClientTrace{
		DNSDone: func(info httptrace.DNSDoneInfo) {
			log.Printf("DNS took: %v", time.Since(start))
		},
		GotConn: func(info httptrace.GotConnInfo) {
			log.Printf("Connection took: %v, Reused: %v", time.Since(start), info.Reused)
		},
	}
	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()
	log.Printf("Status: %s, Total time: %v", resp.Status, time.Since(start))
}

// func createContainerNetworkRules() error {

// }

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
