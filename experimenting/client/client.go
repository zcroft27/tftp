package main

import (
	"fmt"
	"net"
)

func Write() {
	loopback := net.IPv4(byte(127), byte(0), byte(0), byte(1))
	fmt.Println("is loopback? ", loopback.IsLoopback())
	raddr := net.TCPAddr{Port: 12345, IP: loopback}
	conn, err := net.DialTCP("tcp", nil, &raddr)
	if err != nil {
		panic("failed to dial on TCP conn")
	}
	fmt.Println("About to write")
	msg := []byte("hello world")
	n, err := conn.Write(msg)
	if err != nil {
		panic("failed to write on TCP conn")
	}

	fmt.Println("Wrote: ", n)
}

func main() {
	Write()
}
