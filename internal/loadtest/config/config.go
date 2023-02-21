package config

import (
	"flag"
	"fmt"
	"log"
	"strings"
	"time"
)

// Config represents environment config.
type Config struct {
	ServerAddress        string
	PublishEndpoint      string        `config:"publish_endpoint"`
	UseLegacyEvents      bool          `config:"use_legacy_events"`
	EventSource          string        `config:"event_source"`
	MaxInflightMessages0 int           `config:"max_inflight_messages_0"`
	MaxInflightMessages1 int           `config:"max_inflight_messages_1"`
	EventName0           string        `config:"event_name_0"`
	EventName1           string        `config:"event_name_1"`
	VersionFormat        string        `config:"version_format"`
	GenerateCount0       int           `config:"generate_count_0"`
	GenerateCount1       int           `config:"generate_count_1"`
	EpsStart0            int           `config:"eps_start_0"`
	EpsStart1            int           `config:"eps_start_1"`
	EpsIncrement0        int           `config:"eps_increment_0"`
	EpsIncrement1        int           `config:"eps_increment_1"`
	EpsLimit             int           `config:"eps_limit"`
	Workers              int           `config:"workers"`
	MaxIdleConns         int           `config:"max_idle_conns"`
	MaxConnsPerHost      int           `config:"max_conns_per_host"`
	MaxIdleConnsPerHost  int           `config:"max_idle_conns_per_host"`
	IdleConnTimeout      time.Duration `config:"idle_conn_timeout"`
}

func New() *Config {
	c := new(Config)
	flag.StringVar(&c.ServerAddress, "addr", ":8888", "HTTP Server listen address.")
	flag.Parse()
	return c
}

// IsVersionFormatEmpty returns true if event format is empty.
func (c *Config) IsVersionFormatEmpty() bool {
	return len(strings.TrimSpace(c.VersionFormat)) == 0
}

// IsEmptyEventFormat0 returns true if event name 0 is empty.
func (c *Config) IsEmptyEventFormat0() bool {
	return len(strings.TrimSpace(c.EventName0)) == 0
}

// IsEmptyEventFormat1 returns true if event name 1 is empty.
func (c *Config) IsEmptyEventFormat1() bool {
	return len(strings.TrimSpace(c.EventName1)) == 0
}

// ComputeEventsCount returns the count of events to generate.
func (c *Config) ComputeEventsCount() int {
	count := 0
	if !c.IsEmptyEventFormat0() && !c.IsVersionFormatEmpty() {
		count += c.GenerateCount0
	}
	if !c.IsEmptyEventFormat1() && !c.IsVersionFormatEmpty() {
		count += c.GenerateCount1
	}
	return count
}

// ComputeTotalEventsPerSecond returns the total events per second.
func (c *Config) ComputeTotalEventsPerSecond() int {
	count := 0
	for i, eps := 0, c.EpsStart0; i < c.GenerateCount0; i, eps = i+1, c.EpsStart0+(c.EpsIncrement0*(i+1)) {
		count += eps
	}
	for i, eps := 0, c.EpsStart1; i < c.GenerateCount1; i, eps = i+1, c.EpsStart1+(c.EpsIncrement1*(i+1)) {
		count += eps
	}
	return count
}

func (c *Config) PrintTotalEventsPerSecond() {
	log.Printf("Total EPS: %d", c.ComputeTotalEventsPerSecond())
}

func (c *Config) String() string {
	return fmt.Sprintf(
		"ServerAddress: %v PublishEndpoint: %v MaxInflightMessages0: %v MaxInflightMessages1: %v EventFormat0: %v EventFormat1: %v GenerateCount0: %v GenerateCount1: %v EpsStart0: %v EpsStart1: %v EpsIncrement0: %v EpsIncrement1: %v EpsLimit: %v Workers: %v MaxIdleConns: %v MaxConnsPerHost : %v MaxIdleConnsPerHost: %v IdleConnTimeout: %v",
		c.ServerAddress, c.PublishEndpoint, c.MaxInflightMessages0, c.MaxInflightMessages1, c.EventName0, c.EventName1, c.GenerateCount0, c.GenerateCount1, c.EpsStart0, c.EpsStart1, c.EpsIncrement0, c.EpsIncrement1, c.EpsLimit, c.Workers, c.MaxIdleConns, c.MaxConnsPerHost, c.MaxIdleConnsPerHost, c.IdleConnTimeout,
	)
}
