https://datatracker.ietf.org/doc/html/rfc1350

# Protocol flow
1. Request to read/write a file, also serving as a request for connection.  
2. If granted, the connection is opened and the file is sent in fixed length blocks of 512 bytes.  
3. Each data packet contains one block of data and must be ACKd before the next packet can be sent.  
4. A data packet of less than 512 bytes signals termination of a transfer.  

# Notes
1. Both machines are considered senders and receivers.
2. One sends data and receives
   acknowledgments, the other sends acknowledgments and receives data.

# Lost packets
1. If a packet gets lost in the network, the intended recipient will timeout and may retransmit his
last packet (which may be data or an acknowledgment), thus causing
the sender of the lost packet to retransmit that lost packet.   
2. The sender has to keep just one packet on hand for retransmission, since
the lock step acknowledgment guarantees that all older packets have
been received.  

# Errors
1. Most errors cause termination of the connection. An error is signalled by sending an error packet.
2. Error packets are not ACKd and not retransmitted. Therefore, recipient may not receive the error packet and timeouts should be in place.

# Packet form
```
---------------------------------------------------
|  Local Medium  |  Internet  |  Datagram  |  TFTP  |
---------------------------------------------------
```