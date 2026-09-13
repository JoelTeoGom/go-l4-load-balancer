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
	node := dp.registry.ObtainRandomNode()
	if node == nil {
		fmt.Println("No available nodes")
		return
	}

	newConn, err := net.Dial("tcp", node.Address())
	if err != nil {
		fmt.Println(err)
		return
	}
	defer newConn.Close()

	for {
		// go func() {
		// 	responseBuffer := make([]byte, 1024)
		// 	num, err := newConn.Read(responseBuffer)

		// }()

		readBuffer := make([]byte, 1024)
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
