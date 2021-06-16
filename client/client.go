package client

import (
	"github.com/ticker-es/client-go/rpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type Client struct {
	address             string
	insecure            bool
	connection          *grpc.ClientConn
	eventStreamClient   rpc.EventStreamClient
	maintenanceClient   rpc.MaintenanceClient
	authenticationToken string
	autoAcknowledge     bool
}

type Option = func(c *Client)

func NewClient(address string, cred credentials.TransportCredentials) (*Client, error) {
	cl := &Client{}
	//for _, opt := range opts {
	//	opt(cl)
	//}
	if conn, err := grpc.Dial(address, grpc.WithTransportCredentials(cred)); err != nil {
		return nil, err
	} else {
		cl.connection = conn
		cl.eventStreamClient = rpc.NewEventStreamClient(conn)
		cl.maintenanceClient = rpc.NewMaintenanceClient(conn)
		return cl, nil
	}
}
