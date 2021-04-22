package client

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
