package main

import (
	"fmt"
	"net"
)

func Write() {
	loopback := net.IPv4(byte(127), byte(0), byte(0), byte(1))
	fmt.Println("dialing UDP")
	conn, err := net.DialUDP("udp", nil, &net.UDPAddr{IP: loopback, Port: 14444})
	if err != nil {
		panic("failed to dial on udp")
	}
	fmt.Println("dailed UDP")

	n, err := conn.Write([]byte("hello world"))
	if err != nil {
		panic("failed to write on UDP")
	}
	fmt.Printf("Wrote %d bytes\n", n)
}

func main() {
	Write()
}
