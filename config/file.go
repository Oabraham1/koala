package config

import (
	"os"
	"path/filepath"
)

var (
	CAFile                      = configFile("ca.pem")                    // CAFile is the path to the CA file
	ServerCertFile              = configFile("server.pem")                // ServerCertFile is the path to the server certificate file
	ServerKeyFile               = configFile("server-key.pem")            // ServerKeyFile is the path to the server key file
	RootClientCertificateFile   = configFile("root-client.pem")           // RootClientCertificateFile is the path to the root client certificate file
	RootClientKeyFile           = configFile("root-client-key.pem")       // RootClientKeyFile is the path to the root client key file
	NobodyClientCertificateFile = configFile("nobody-client.pem")         // NobodyClientCertificateFile is the path to the nobody client certificate file
	NobodyClientKeyFile         = configFile("nobody-client-key.pem")     // NobodyClientKeyFile is the path to the nobody client key file
	AccessControlModelFile      = configFile("access-control-model.conf") // AccessControlModelFile is the path to the access control model file
	AccessControlPolicyFile     = configFile("access-control-policy.csv") // AccessControlPolicyFile is the path to the access control policy file
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
