package utils

import "os"

type Config struct {
	EtcdNodes string
	ServerUrl string
}

func NewConfig() *Config {
	return &Config{
		EtcdNodes: os.Getenv("ETCD_NODES"),
		ServerUrl: os.Getenv("SERVER_URL"),
	}
}
