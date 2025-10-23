package main

import (
	"fmt"
	"net"
)

func Receive() {
	// conn, err := net.ListenUDP("udp", &net.UDPAddr{Port: 12345})
	// if err != nil {
	// 	panic("failed to open connection")
	// }
	loopback := net.IPv4(byte(127), byte(0), byte(0), byte(1))
	netconn, err := net.ListenTCP("tcp", &net.TCPAddr{IP: loopback, Port: 12345})
	if err != nil {
		panic("failed to open netconn")
	}

	fmt.Println("Listening...")

	buffer := make([]byte, 1024)
	conn, err := netconn.Accept()
	if err != nil {
		panic("failed to read from UDP")
	}
	n, err := conn.Read(buffer)
	if err != nil {
		panic("failed to read from conn")
	}
	fmt.Printf("Read %d byte\n", n)

	fmt.Println("message: ", string(buffer))
}

func main() {
	Receive()
}
