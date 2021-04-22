package client

import (
	"github.com/ticker-es/client-go/rpc"
	"google.golang.org/grpc"
)

type Client struct {
	address             string
	connection          *grpc.ClientConn
	eventStreamClient   rpc.EventStreamClient
	maintenanceClient   rpc.MaintenanceClient
	authenticationToken string
	autoAcknowledge     bool
}

type Option = func(c *Client)

func NewClient(address string, opts ...Option) (*Client, error) {
	cl := &Client{}
	for _, opt := range opts {
		opt(cl)
	}
	if conn, err := grpc.Dial(address, grpc.WithInsecure()); err != nil {
		return nil, err
	} else {
		cl.connection = conn
		cl.eventStreamClient = rpc.NewEventStreamClient(conn)
		cl.maintenanceClient = rpc.NewMaintenanceClient(conn)
		return cl, nil
	}
}
