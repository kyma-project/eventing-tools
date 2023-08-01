package config

import (
	"flag"
	"fmt"
	"time"
)

// Config represents environment config.
type Config struct {
	ServerAddress       string
	PublishHost         string        `config:"publish_host"`
	MaxIdleConns        int           `config:"max_idle_conns"`
	MaxConnsPerHost     int           `config:"max_conns_per_host"`
	MaxIdleConnsPerHost int           `config:"max_idle_conns_per_host"`
	IdleConnTimeout     time.Duration `config:"idle_conn_timeout"`
}

func New() *Config {
	c := new(Config)
	flag.StringVar(&c.ServerAddress, "addr", ":8888", "HTTP Server listen address.")
	flag.Parse()
	return c
}

func (c *Config) String() string {
	return fmt.Sprintf(
		"ServerAddress: %v PublishHost: %v MaxIdleConns: %v MaxConnsPerHost : %v MaxIdleConnsPerHost: %v IdleConnTimeout: %v",
		c.ServerAddress, c.PublishHost, c.MaxIdleConns, c.MaxConnsPerHost, c.MaxIdleConnsPerHost, c.IdleConnTimeout,
	)
}
