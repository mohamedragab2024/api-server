package utils

import "os"

type Config struct {
	EtcdNodes string
}

func NewConfig() *Config {
	return &Config{
		EtcdNodes: os.Getenv("ETCD_NODES"),
	}
}
