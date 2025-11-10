package main

import (
	"flag"
	"fmt"
	"log"
)

func Transfer(hostPath string, remotePath string, hostToRemote bool) {

}

var validModes = map[string]func() error{
	"put": writeFile,
	"get": readFile,
}

func main() {

	mode := flag.String("mode", "put", "To write (put) to or read (get) a file from remote.")
	local := flag.String("host-path", "", "The path on the host to read from or write to.")
	remote := flag.String("remote-path", "", "The path on remote to read from or write to.")

	flag.Parse()

	err := validateFlags(mode, remote, local)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	err = validModes[*mode]()
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func writeFile() error {
	return nil
}

func readFile() error {
	return nil
}

func validateFlags(mode, remote, local *string) error {
	if _, ok := validModes[*mode]; !ok {
		return fmt.Errorf("mode %s is not valid", *mode)
	}

	if *remote == "" {
		return fmt.Errorf("remote path %s must not be empty", *remote)
	}

	if *local == "" {
		return fmt.Errorf("local path %s must not be empty", *local)
	}

	return nil
}
