package client

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func dialOptionWrapper(opts ...grpc.DialOption) Option {
	return func(c *Client) {
		c.dialOptions = append(c.dialOptions, opts...)
	}
}

func Credentials(cred credentials.TransportCredentials) Option {
	return dialOptionWrapper(grpc.WithTransportCredentials(cred))
}

func AuthenticationToken(token string) Option {
	return func(c *Client) {
		c.authenticationToken = token
	}
}

func AutoAcknowledge() Option {
	return func(c *Client) {
		c.autoAcknowledge = true
	}
}
