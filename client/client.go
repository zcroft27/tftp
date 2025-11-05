package client

import (
	"flag"
	"fmt"
	"log"
)

func Transfer(hostPath string, remotePath string, hostToRemote bool) {

}

var validModes = make(map[string]struct{})

func main() {
	validModes["put"] = struct{}{}
	validModes["get"] = struct{}{}

	mode := flag.String("mode", "put", "To write (put) to or read (get) a file from remote.")
	local := flag.String("host-path", "", "The path on the host to read from or write to.")
	remote := flag.String("remote-path", "", "The path on remote to read from or write to.")

	flag.Parse()

	err := validateFlags(mode, remote, local)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func validateFlags(mode, remote, local *string) error {
	if _, ok := validModes[*mode]; !ok {
		return fmt.Errorf("Mode %s is not valid", *mode)
	}

	if *remote == "" {
		return fmt.Errorf("Remote path %s must not be empty", *remote)
	}

	if *local == "" {
		return fmt.Errorf("Local path %s must not be empty", *local)
	}

	return nil
}
