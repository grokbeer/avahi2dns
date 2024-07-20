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
	Domains  []string `arg:"-d" help:"comma-separated list of domains to resolve"`
	BindAddr string   `arg:"-a,--addr" default:"localhost" help:"address to bind on" env:"BIND"`
	Port     uint16   `arg:"-p" default:"53" help:"port to bind on" env:"PORT"`
	Debug    bool     `arg:"-v" default:"false" help:"also include debug information"`
    v4Only   bool     `arg:"-4" help:"only support IPv4 (A) queries"`
    v6Only   bool     `arg:"-6" help:"only support IPv6 (AAAA) queries"`
}

func parseArgs(logger *logrus.Logger) (*config, error) {
	cfg := &config{}
	arg.MustParse(cfg)
	if len(cfg.Domains) == 0 {
		cfg.Domains = defaultDomains
	}
	// Logic to handle conflicting arguments, if necessary
	if cfg.v4Only && cfg.v6Only {
		logger.Warn("Both IPv4 and IPv6 support enabled. This is the default behavior.")
		cfg.v4Only = false
		cfg.v6Only = false
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
