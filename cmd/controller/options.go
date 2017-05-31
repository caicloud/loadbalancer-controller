package main

import (
	"gopkg.in/urfave/cli.v1"
)

// Options contains controller options
type Options struct {
	// ProxyOptions
	// ProviderOptions
	// ControllerOptions
	Kubeconfig string
	Debug      bool
}

// NewOptions reutrns a new Options
func NewOptions() *Options {
	return &Options{}
}

// AddFlags add flags to app
func (opts *Options) AddFlags(app *cli.App) {

	flags := []cli.Flag{
		cli.StringFlag{
			Name:        "kubeconfig",
			Usage:       "Path to a kube config. Only required if out-of-cluster.",
			Destination: &opts.Kubeconfig,
		},
		cli.BoolFlag{
			Name:        "debug",
			Usage:       "run with debug mode",
			Destination: &opts.Debug,
		},
	}

	app.Flags = append(app.Flags, flags...)

}
