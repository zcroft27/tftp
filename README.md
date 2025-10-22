# What
This is a TFTP parser. This began as a simple project to explore parsing a binary protocol. TFTP is a straightforward protocol that I may make a client and server for in the future.

The parser is unit tested.

# Overview
TFTP is the Trivial File Transfer Protocol defined in [RFC 1350](https://datatracker.ietf.org/doc/html/rfc1350).
It allows for two nodes to act as a client/server to read or write files, with no user authentication or listing of directories.

RFC 1305: "Notice that both machines involved in a transfer are considered senders and receivers.  One sends data and receives acknowledgments, the other sends acknowledgments and receives data."

TFTP requires the recipient to acknowledge every packet before the sender can send the next.

TFTP is typically implemented on top of UDP, with clients initially sending requests to server port 69. The server then responds from a randomly chosen ephemeral port for all subsequent packets in that transfer, allowing multiple concurrent sessions.

There are details to explore in the implementation of a client and server, including handling duplicate packets,
retransmitting packets, timeouts when ACKs are missing, and determining correct error codes.

# Intentions
I'd like to build an actual client and server for this protocol.