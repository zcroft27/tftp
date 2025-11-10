package main

import (
	"flag"
	"log"
	"tftp/internal/server"
)

func main() {
	port := flag.Int("port", 69, "Port to listen on")
	root := flag.String("root", "./tftp-root", "Root directory for file transfers")
	flag.Parse()

	srv := server.New(*port, *root)
	log.Printf("Starting TFTP server on port %d...", *port)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
