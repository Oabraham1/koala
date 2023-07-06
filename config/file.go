package config

import (
	"os"
	"path/filepath"
)

var (
	CAFile                      = configFile("ca.pem")
	ServerCertFile              = configFile("server.pem")
	ServerKeyFile               = configFile("server-key.pem")
	RootClientCertificateFile   = configFile("root-client.pem")
	RootClientKeyFile           = configFile("root-client-key.pem")
	NobodyClientCertificateFile = configFile("nobody-client.pem")
	NobodyClientKeyFile         = configFile("nobody-client-key.pem")
	AccessControlModelFile      = configFile("access-control-model.conf")
	AccessControlPolicyFile     = configFile("access-control-policy.csv")
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
