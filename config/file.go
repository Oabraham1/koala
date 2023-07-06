package config

import (
	"os"
	"path/filepath"
)

var (
	// CAFile is the path to the CA file
	CAFile = configFile("ca.pem")
	// ServerCertFile is the path to the server certificate file
	ServerCertFile = configFile("server.pem")
	// ServerKeyFile is the path to the server key file
	ServerKeyFile = configFile("server-key.pem")
	// RootClientCertificateFile is the path to the root client certificate file
	RootClientCertificateFile = configFile("root-client.pem")
	// RootClientKeyFile is the path to the root client key file
	RootClientKeyFile = configFile("root-client-key.pem")
	// NobodyClientCertificateFile is the path to the nobody client certificate file
	NobodyClientCertificateFile = configFile("nobody-client.pem")
	// NobodyClientKeyFile is the path to the nobody client key file
	NobodyClientKeyFile = configFile("nobody-client-key.pem")
	// AccessControlModelFile is the path to the access control model file
	AccessControlModelFile = configFile("access-control-model.conf")
	// AccessControlPolicyFile is the path to the access control policy file
	AccessControlPolicyFile = configFile("access-control-policy.csv")
)

// configFile returns the path to a configuration file
func configFile(name string) string {
	if directory := os.Getenv("CONFIG_DIR"); directory != "" {
		return filepath.Join(directory, name)
	}
	homeDirectory, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	return filepath.Join(homeDirectory, ".koala", name)
}
