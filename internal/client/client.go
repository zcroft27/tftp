package client

type Client struct {
	serverAddr string
}

func New(serverAddr string) *Client {
	return &Client{serverAddr: serverAddr}
}

func (c *Client) Get(remote, local string) error {
	// Send RRQ
	// Receive DATA packets
	// Send ACKs
	return nil
}

func (c *Client) Put(local, remote string) error {
	// Send WRQ
	// Send DATA packets
	// Receive ACKs
	return nil
}
