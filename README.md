# What
This is a TFTP parser, client, and daemon. This began as a simple project to explore parsing a binary protocol.

The parser is unit tested.

# Overview
TFTP is the Trivial File Transfer Protocol defined in [RFC 1350](https://datatracker.ietf.org/doc/html/rfc1350).
It allows for two nodes to act as a client/server to read or write files, with no user authentication or listing of directories.

RFC 1305: "Notice that both machines involved in a transfer are considered senders and receivers.  One sends data and receives acknowledgments, the other sends acknowledgments and receives data."

TFTP requires the recipient to acknowledge every packet before the sender can send the next.

TFTP is typically implemented on top of UDP, with clients initially sending requests to server port 69. The server then responds from a randomly chosen ephemeral port for all subsequent packets in that transfer, allowing multiple concurrent sessions.

There are details to explore in the implementation of a client and server, including handling duplicate packets,
retransmitting packets, timeouts when ACKs are missing, and determining correct error codes.

# How to Use
## Dependencies
```bash
git clone git@github.com:zcroft27/tftp.git tftp
cd tftp
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)" # install homebrew if not already.
brew install task # optional, enables use of Taskfile for manual testing.
brew install go
go mod tidy
```
## Operation
```bash
sudo task server > server.log 2>&1 & # sudo since binding to privileged port 69, & to run in the background. CAREFUL: sudo first time will ask for password, use `fg` to enter it. Or, just use a different shell for this.
task generate-test-file # creates cmd/tftpd/tftp-root/test.txt with ~280 MiB of random ascii data.
task get # will retrieve the generated data into test.txt.
task put # will write the generated data into cmd/tftpd/tftp-root/written-to.txt.
```

## Cleanup
```bash
sudo lsof -i :69 # view the server process!
fg # become the server process...
ctrl+c # SIGINT
```