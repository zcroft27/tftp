package main

import (
	"fmt"
	"net"
)

func Read() {
	loopback := net.IPv4(byte(127), byte(0), byte(0), byte(1))
	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: loopback, Port: 14444})
	if err != nil {
		panic("failed to listen on udp")
	}
	fmt.Println("Reading...")
	buff := make([]byte, 1024)
	conn.Read(buff)
	fmt.Println("received msg: ", string(buff))
}

func main() {
	Read()
}
