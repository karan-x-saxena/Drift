package config

import (
	"drift/pkg"
	"errors"
	"fmt"

	"github.com/valyala/fasthttp"
)

const (
	defaultMaxConn                   uint = 512
	defaultMaxConnTimeout            uint = 30
	defaultMaxConnDuration           uint = 10
	defaultMaxIdleConnDuration       uint = 10
	defaultMaxIdemponentCallAttempts uint = 5
	defaultHealtCheckerTime          uint = 30
	defaultMaxIdleWorkerDuration     uint = 10

	Http1 = "http1.1"
	Http2 = "http2"
)

var ValidAlgo []string = []string{"round-robin", "w-round-robin", "ip-hash", "random", "least-connection", "least-response-time"}

type Drift struct {
	HttpVersion                   string `yaml:"http_version" json:"http_version"`
	CertFile                      string `yaml:"cert_file" json:"cert_file"`
	KeyFile                       string `yaml:"key_file" json:"key_file"`
	MaxIdleWorkerDuration         uint   `yaml:"max_idle_worker_duration" json:"max_idle_worker_duration"`
	TCPKeepAlivePeriod            uint   `yaml:"tcp_keepalive_period" json:"tcp_keepalive_period"`
	Concurrency                   uint   `yaml:"concurrency" json:"concurrency"`
	ReadTimeout                   uint   `yaml:"read_timeout" json:"read_timeout"`
	WriteTimeout                  uint   `yaml:"write_timeout" json:"write_timeout"`
	IdleTimeout                   uint   `yaml:"idle_timeout" json:"idle_timeout"`
	DisableKeepAlive              bool   `yaml:"disable_keepalive" json:"disable_keepalive"`
	DisableHeaderNamesNormalizing bool   `yaml:"disable_header_names_normalizing" json:"disable_header_names_normalizing"`
}

func (d *Drift) addDriftDefault() {
	if d.HttpVersion == "" || d.HttpVersion != Http2 {
		d.HttpVersion = Http1
	}

	if d.MaxIdleWorkerDuration == 0 {
		d.MaxIdleWorkerDuration = defaultMaxIdleWorkerDuration
	}

	if d.Concurrency == 0 {
		d.Concurrency = fasthttp.DefaultConcurrency
	}

}

func (d *Drift) validateDrift() error {
	if d.HttpVersion == Http2 && (d.CertFile == "" || d.KeyFile == "") {
		return errors.New("the HTTP/2 connection can be only established if the server is using TLS. Please provide cert and key file")
	}

	if err := pkg.IsFileExist(d.CertFile); err != nil && d.CertFile != "" {
		return err
	}

	if err := pkg.IsFileExist(d.KeyFile); err != nil && d.KeyFile != "" {
		return err
	}
	return nil

}

type Server struct {
	Url                       string `yaml:"url" json:"url"`
	HealthCheckPath           string `yaml:"health_check_path" json:"health_check_path"`
	MaxConnection             uint   `yaml:"max_connection" json:"max_connectoin"`
	MaxConnectionTimeout      uint   `yaml:"max_connection_timeout" json:"max_connection_timeout"`
	MaxConnectionDuration     uint   `yaml:"max_connection_duration" json:"max_connection_duration"`
	MaxIdleConnectionDuration uint   `yaml:"max_idle_connection_duration" json:"max_idle_connection_duration"`
	MaxIdemponentCallAttempts uint   `yaml:"max_idemponent_call_attempts" json:"max_idemponent_call_attempts"`
}

func (s *Server) addServerDefaults() {
	if s.HealthCheckPath == "" {
		s.HealthCheckPath = "/"

	}

	if s.MaxConnection == 0 {
		s.MaxConnection = defaultMaxConn
	}

	if s.MaxConnectionTimeout == 0 {
		s.MaxConnectionTimeout = defaultMaxConnTimeout
	}

	if s.MaxConnectionDuration == 0 {
		s.MaxConnectionDuration = defaultMaxConnDuration
	}

	if s.MaxIdleConnectionDuration == 0 {
		s.MaxIdleConnectionDuration = defaultMaxIdleConnDuration
	}

	if s.MaxIdemponentCallAttempts == 0 {
		s.MaxIdemponentCallAttempts = defaultMaxIdemponentCallAttempts
	}

}

func (s *Server) validateServer() error {
	if s.Url == "" {
		return errors.New("undefined server url")
	}

	return nil
}

func (s *Server) GetHealthCheckURL() string {
	return "http://" + s.Url + s.HealthCheckPath
}

type Config struct {
	Algo            string   `yaml:"algo" json:"algo"`
	Port            uint     `yaml:"port" json:"port"`
	Host            string   `yaml:"host" json:"host"`
	HeathCheckTimer uint     `yaml:"heatlt_check_timer" json:"heatlt_check_timer"`
	Servers         []Server `yaml:"servers" json:"servers"`
	Drift           Drift    `yaml:"drift" json:"drift"`
}

func (c *Config) addDefaults() {
	if c.HeathCheckTimer == 0 {
		c.HeathCheckTimer = defaultHealtCheckerTime
	}

	for index, server := range c.Servers {
		server.addServerDefaults()
		c.Servers[index] = server
	}

	c.Drift.addDriftDefault()

}

func (c *Config) validate() error {
	if len(c.Servers) == 0 {
		return errors.New("at least one backend must be set")
	}

	if !pkg.Contains(ValidAlgo, c.Algo) {
		return fmt.Errorf("select one algo in conf from %s", ValidAlgo)
	}

	if 1024 > c.Port || c.Port > 49151 {
		return errors.New("please choose valid port between 1024 and 49151")
	}

	if c.Host == "" {
		return errors.New("undifined host")
	}

	for _, server := range c.Servers {
		err := server.validateServer()
		if err != nil {
			return err
		}
	}

	err := c.Drift.validateDrift()
	if err != nil {
		return err
	}

	return nil

}
