package config

import (
	"os"
	"path/filepath"
)

var (
	CAFile         = configFile("ca.pem")
	ServerCertFile = configFile("server.pem")
	ServerKeyFile  = configFile("server-key.pem")
)

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
