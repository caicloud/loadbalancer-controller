package config

import (
	lbconfig "github.com/caicloud/loadbalancer-controller/pkg/config"
)

// Config is the main context object for the controller manager.
type Config struct {
	Cfg lbconfig.Configuration
}
