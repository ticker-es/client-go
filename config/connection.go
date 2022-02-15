package config

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/mtrense/soil/logging"
	"github.com/ticker-es/client-go/client"
	"google.golang.org/grpc/credentials"
	"io/ioutil"
)

func Connect(connect, caCert, clientCert, clientKey string) *client.Client {
	certificates := readClientCerts(clientCert, clientKey)
	cfg := &tls.Config{
		Certificates:     certificates,
		RootCAs:          ReadCACerts(caCert),
		VerifyConnection: verifyConnection,
	}
	cred := credentials.NewTLS(cfg)
	cl := client.NewClient(connect, client.Credentials(cred))
	if err := cl.Connect(); err != nil {
		panic(err)
	}
	return cl
}

func readClientCerts(clientCert, clientKey string) []tls.Certificate {
	var certificates []tls.Certificate
	if cert, err := tls.LoadX509KeyPair(clientCert, clientKey); err == nil {
		certificates = append(certificates, cert)
	} else {
		logging.L().Err(err).Msg("Could not read client certificate/key")
	}
	return certificates
}

func verifyConnection(state tls.ConnectionState) error {
	return nil
}

func ReadCACerts(caCertFiles ...string) *x509.CertPool {
	caCerts := x509.NewCertPool()
	for _, caCertFile := range caCertFiles {
		if caCertData, err := ioutil.ReadFile(caCertFile); err == nil {
			if !caCerts.AppendCertsFromPEM(caCertData) {
				logging.L().Error().Str("filename", caCertFile).Msg("Could not parse CA Certificate from PEM")
			}
		} else {
			logging.L().Err(err).Msg("Could not read CA Certificate")
		}
	}
	return caCerts
}
