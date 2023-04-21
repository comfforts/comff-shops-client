package config

import (
	"os"
	"path/filepath"
)

const (
	CAFile         = "ca.pem"
	ServerCertFile = "server.pem"
	ServerKeyFile  = "server-key.pem"
	ClientCertFile = "client.pem"
	ClientKeyFile  = "client-key.pem"
)

type FileType string

const (
	Policy FileType = "Policy"
	Certs  FileType = "Certs"
)

const (
	POLICY_BASE_DIR = "policies"
	CERTS_BASE_DIR  = "certs"
)

func CertFile(fileName string) string {
	return configFile(fileName, Certs)
}

func PolicyFile(fileName string) string {
	return configFile(fileName, Policy)
}

func configFile(fileName string, fileType FileType) string {
	if fileType == Policy {
		if dir := os.Getenv("CERTS_PATH"); dir != "" {
			return filepath.Join(dir, POLICY_BASE_DIR, fileName)
		}
		return filepath.Join(POLICY_BASE_DIR, fileName)
	}

	if dir := os.Getenv("CERTS_PATH"); dir != "" {
		return filepath.Join(dir, CERTS_BASE_DIR, fileName)
	}
	return filepath.Join(CERTS_BASE_DIR, fileName)
}
