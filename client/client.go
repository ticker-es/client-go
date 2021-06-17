package client

import (
	"github.com/ticker-es/client-go/rpc"
	"google.golang.org/grpc"
)

type Client struct {
	address             string
	insecure            bool
	dialOptions         []grpc.DialOption
	connection          *grpc.ClientConn
	eventStreamClient   rpc.EventStreamClient
	maintenanceClient   rpc.MaintenanceClient
	authenticationToken string
	autoAcknowledge     bool
}

type Option = func(c *Client)

func NewClient(address string, opts ...Option) *Client {
	cl := &Client{
		address: address,
	}
	for _, opt := range opts {
		opt(cl)
	}
	return cl
}

func (s *Client) Connect() error {
	if conn, err := grpc.Dial(s.address, s.dialOptions...); err != nil {
		return err
	} else {
		s.connection = conn
		s.eventStreamClient = rpc.NewEventStreamClient(conn)
		s.maintenanceClient = rpc.NewMaintenanceClient(conn)
		return nil
	}
}
