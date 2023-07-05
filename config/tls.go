package config

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
)

type TLSConfig struct {
	CAFile          string
	CertificateFile string
	KeyFile         string
	ServerAddress   string
	Server          bool
}

func SetupTLSConfiguration(config TLSConfig) (*tls.Config, error) {
	var err error
	tlsConfig := &tls.Config{}
	if config.CertificateFile != "" && config.KeyFile != "" {
		tlsConfig.Certificates = make([]tls.Certificate, 1)
		tlsConfig.Certificates[0], err = tls.LoadX509KeyPair(config.CertificateFile, config.KeyFile)
		if err != nil {
			return nil, err
		}
	}
	if config.CAFile != "" {
		bit, err := os.ReadFile(config.CAFile)
		if err != nil {
			return nil, err
		}
		ca := x509.NewCertPool()
		ok := ca.AppendCertsFromPEM(bit)
		if !ok {
			return nil, fmt.Errorf("failed to parse root certificate: %q", config.CAFile)
		}
		if config.Server {
			tlsConfig.ClientCAs = ca
			tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
		} else {
			tlsConfig.RootCAs = ca
		}
		tlsConfig.ServerName = config.ServerAddress
	}
	return tlsConfig, nil
}
