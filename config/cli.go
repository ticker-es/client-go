package config

import c "github.com/mtrense/soil/config"

func FlagConnect() c.Applicant {
	return c.Flag("connect", c.Str("localhost:6677"), c.Abbr("c"), c.Description("Server to connect to"), c.Mandatory(), c.Persistent(), c.Env())
}

func FlagCerts() c.Applicant {
	return func(b *c.Command) {
		b.Apply(
			c.Flag("ca-cert", c.Str(""), c.Description("CA certificate used to verify server connection"), c.Persistent(), c.EnvName("ca_cert")),
			c.Flag("client-cert", c.Str(""), c.Description("Client certificate"), c.Persistent(), c.EnvName("client_cert")),
			c.Flag("client-key", c.Str(""), c.Description("Client key"), c.Persistent(), c.EnvName("client_key")),
		)
	}
}
