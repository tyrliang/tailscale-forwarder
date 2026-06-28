package config

import (
	"fmt"
	"os"
	"strconv"

	"main/internal/logger"
	"main/internal/util"

	"github.com/caarlos0/env/v11"
)

var Cfg = config{}

func Load() {
	var errs []error

	if err := env.Parse(&Cfg); err != nil {
		if e, ok := err.(env.AggregateError); ok {
			errs = append(errs, e.Errors...)
		} else {
			errs = append(errs, err)
		}
	}

	connectionMappings, err := parseConnectionMappings("CONNECTION_MAPPING_", os.Environ())
	if err != nil {
		errs = append(errs, err)
	}

	egressConnectionMappings, egressErr := parseConnectionMappings("EGRESS_CONNECTION_MAPPING_", os.Environ())
	if egressErr != nil {
		errs = append(errs, egressErr)
	}

	if len(connectionMappings) == 0 && len(egressConnectionMappings) == 0 && err == nil && egressErr == nil {
		errs = append(errs, fmt.Errorf("at least one CONNECTION_MAPPING_[n] or EGRESS_CONNECTION_MAPPING_[n] is required"))
	}

	Cfg.ConnectionMappings = connectionMappings
	Cfg.EgressConnectionMappings = egressConnectionMappings

	sanitizedHostname := util.SanitizeString(Cfg.TSHostname)

	if sanitizedHostname == "" {
		errs = append(errs, fmt.Errorf("TS_HOSTNAME must be a valid hostname, before sanitization: \"%s\", after sanitization: \"%s\"", Cfg.TSHostname, sanitizedHostname))
	}

	Cfg.TSHostname = sanitizedHostname

	if len(errs) > 0 {
		logger.StderrWithSource.Error("configuration error(s) found", logger.ErrorsAttr(errs...))
		os.Exit(1)
	}

	// Tailscale reads TS_DEBUG_MTU directly from the environment when bringing
	// the network up, so write the resolved value back for it to pick up.
	os.Setenv("TS_DEBUG_MTU", strconv.Itoa(Cfg.TSDebugMTU))
}
