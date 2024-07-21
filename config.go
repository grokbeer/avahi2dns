package main

import (
	"github.com/alexflint/go-arg"
	"github.com/sirupsen/logrus"
)

var defaultDomains = []string{
	"home",
	"internal",
	"intranet",
	"lan",
	"local",
	"private",
	"test",
}

type config struct {
	Domains     []string `arg:"-d" help:"comma-separated list of domains to resolve"`
	BindAddr    string   `arg:"-a,--addr,env:BIND" default:"localhost" help:"address to bind on"`
	Port        uint16   `arg:"-p,env" default:"53" help:"port to bind on"`
	IPv4Only    bool     `arg:"-4,env:IPV4_ONLY" default:"false" help:"only support IPv4 (A) queries"`
	IPv6Only    bool     `arg:"-6,env:IPV6_ONLY" default:"false" help:"only support IPv6 (AAAA) queries"`
	TimeoutSecs uint8    `arg:"-t,--timeout,env:TIMEOUT" default:"0" help:"timeout for avahi requests (in seconds, 0 for default avahi timeout)"`
	Debug       bool     `arg:"-v" default:"false" help:"also include debug information"`
}

func parseArgs(logger *logrus.Logger) (*config, error) {
	cfg := &config{}
	arg.MustParse(cfg)
	if len(cfg.Domains) == 0 {
		cfg.Domains = defaultDomains
	}
	// Logic to handle conflicting arguments, if necessary
	if cfg.IPv4Only && cfg.IPv6Only {
		logger.Warn("Both IPv4 and IPv6 support enabled. This is the default behavior.")
		cfg.IPv4Only = false
		cfg.IPv6Only = false
	}
	configureLogger(logger, cfg)
	logger.WithField("config", cfg).Debug("config parsed")
	return cfg, nil
}

func configureLogger(logger *logrus.Logger, cfg *config) {
	if cfg.Debug {
		logger.SetReportCaller(true)
		logger.SetLevel(logrus.DebugLevel)
	}
}
