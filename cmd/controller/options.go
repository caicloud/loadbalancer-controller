/*
Copyright 2017 Caicloud authors. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"github.com/caicloud/loadbalancer-controller/config"
	log "github.com/zoumo/logdog"
	"gopkg.in/urfave/cli.v1"
)

// Options contains controller options
type Options struct {
	Kubeconfig string
	Debug      bool
	Cfg        config.Configuration
}

// NewOptions reutrns a new Options
func NewOptions() *Options {
	return &Options{}
}

// AddFlags add flags to app
func (opts *Options) AddFlags(app *cli.App) {
	opts.Cfg.AddFlags(app)

	flags := []cli.Flag{
		cli.StringFlag{
			Name:        "kubeconfig",
			Usage:       "Path to a kube config. Only required if out-of-cluster.",
			Destination: &opts.Kubeconfig,
		},
		cli.BoolFlag{
			Name:        "debug",
			Usage:       "Run with debug mode",
			Destination: &opts.Debug,
		},
		cli.BoolFlag{
			Name:        "log-force-color",
			Usage:       "Force log to output with colore",
			Destination: &log.ForceColor,
		},
	}

	app.Flags = append(app.Flags, flags...)

}
