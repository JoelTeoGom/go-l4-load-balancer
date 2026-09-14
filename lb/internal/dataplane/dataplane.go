package dataplane

import (
	"context"
	"fmt"
	"net"

	"github.com/JoelTeoGom/go-l4-load-balancer/lb/internal/registry"
)

type DataPlane struct {
	registry *registry.Registry
}

func NewDataPlane(registry *registry.Registry) *DataPlane {
	return &DataPlane{
		registry: registry,
	}
}

func (dp *DataPlane) StartDataplane(ctx context.Context, address string) error {
	address = fmt.Sprintf("%s:8080", address)
	fmt.Println("Listening Data plane: ", address)
	ln, err := net.Listen("tcp", address)
	if err != nil {
		fmt.Println(err)
		return err
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}

		go dp.handleConnection(ctx, conn)
	}
}

func (dp *DataPlane) handleConnection(ctx context.Context, conn net.Conn) {
	defer conn.Close()
	node := dp.registry.ObtainRandomNode()
	if node == nil {
		fmt.Println("No available nodes")
		return
	}
	address := fmt.Sprintf("%s:8080", node.Address())
	newConn, err := net.Dial("tcp", address)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer newConn.Close()

	readBuffer := make([]byte, 1024)
	//	writeBuffer := make([]byte, 1024)
	for {
		// go func() {
		// 	responseBuffer := make([]byte, 1024)
		// 	num, err := newConn.Read(responseBuffer)

		// }()

		n, err := conn.Read(readBuffer)
		if err != nil {
			fmt.Println(err)
			return
		}

		_, err = newConn.Write(readBuffer[:n])
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}
