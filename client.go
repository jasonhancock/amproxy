package amproxy

import (
	"net"

	"github.com/pkg/errors"
)

// ErrorNotConnected is the error returned when Write is called but the client
// isn't connected to the remote server yet
var ErrorNotConnected = errors.New("You must call Connect() before attempting to write metrics")

// Client is a client that will sign metrics and send to an amproxy server
type Client struct {
	conn      net.Conn
	addr      string
	apiKey    string
	apiSecret string
}

// NewClient creates a new amproxy Client
func NewClient(apiKey, apiSecret, serverAddr string) *Client {
	return &Client{
		addr:      serverAddr,
		apiKey:    apiKey,
		apiSecret: apiSecret,
	}
}

// Write computes the signature for the message and ships it over the wire
func (c *Client) Write(m Message) error {
	if c.conn == nil {
		return ErrorNotConnected
	}

	m.PublicKey = c.apiKey
	m.Signature = m.ComputeSignature(c.apiSecret)
	_, err := c.conn.Write([]byte(m.String() + "\n"))
	return err
}

// Connect connects to the remote amproxy server
func (c *Client) Connect() error {
	if c.conn == nil {
		conn, err := net.Dial("tcp", c.addr)
		if err != nil {
			return errors.Wrap(err, "dialing "+c.addr)
		}
		c.conn = conn
	}
	return nil
}

// Disconnect disconnects from the remote amproxy server
func (c *Client) Disconnect() error {
	if c.conn != nil {
		err := c.conn.Close()
		c.conn = nil
		if err != nil {
			return errors.Wrap(err, "closing connection")
		}
	}
	return nil
}
