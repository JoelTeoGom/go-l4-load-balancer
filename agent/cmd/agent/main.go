package main

func main() {

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
