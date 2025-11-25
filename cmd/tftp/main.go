package main

import (
	"flag"
	"fmt"
	"log"
	"tftp/internal/client"
)

func Transfer(hostPath string, remotePath string, hostToRemote bool) {

}

var validModes = map[string]struct{}{"get": {}, "put": {}}

func main() {

	mode := flag.String("mode", "put", "To write (put) to or read (get) a file from remote.")
	remoteAddress := flag.String("remote-address", "", "Remote server address")
	local := flag.String("host-path", "", "The path on the host to read from or write to.")
	remote := flag.String("remote-path", "", "The path on remote to read from or write to.")

	flag.Parse()

	err := validateFlags(mode, remote, remoteAddress, local)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	cli := client.New(*remoteAddress)

	fmt.Println("host: ", *local)

	operations := map[string]func(string, string) error{
		"get": cli.Get,
		"put": cli.Put,
	}

	op := operations[*mode] // Validated safe in validateFlags.
	err = op(*remote, *local)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("finished with success")
}

func validateFlags(mode, remote, remoteAddress, local *string) error {
	if _, ok := validModes[*mode]; !ok {
		return fmt.Errorf("mode %s is not valid", *mode)
	}

	if *remote == "" {
		return fmt.Errorf("remote path %s must not be empty", *remote)
	}

	if *local == "" {
		return fmt.Errorf("local path %s must not be empty", *local)
	}

	if *remoteAddress == "" {
		return fmt.Errorf("remote address %s must not be empty", *local)
	}

	return nil
}
