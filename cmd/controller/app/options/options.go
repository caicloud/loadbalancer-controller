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

package options

import (
	goflag "flag"
	"os"
	"strconv"

	"github.com/caicloud/clientset/kubernetes"
	gocommonclient "github.com/caicloud/go-common/kubernetes/client"
	"github.com/caicloud/loadbalancer-controller/cmd/controller/app/config"
	lbconfig "github.com/caicloud/loadbalancer-controller/pkg/config"
	"github.com/spf13/pflag"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
)

const (
	// UserAgent is the userAgent name when starting loadbalancer controller.
	UserAgent = "loadbalancer-controller"
)

// Options is the main context object for the admission controller.
type Options struct {
	Master      string
	Kubeconfig  string
	HealthzPort int
	Cfg         lbconfig.Configuration
}

// NewOptions creates a new AddmissionOptions with a default config.
func NewOptions() *Options {
	return &Options{
		HealthzPort: 8081,
	}
}

// Flags returns flags for admission controller
func (s *Options) Flags() *pflag.FlagSet {
	fs := pflag.NewFlagSet("options", pflag.ExitOnError)

	s.Cfg.AddFlags(fs)

	fs.StringVar(&s.Kubeconfig, "kubeconfig", s.Kubeconfig, "Path to kubeconfig file with authorization and master location information.")
	fs.StringVar(&s.Master, "master", s.Master, "The address of the Kubernetes API server (overrides any value in kubeconfig).")
	fs.IntVar(&s.HealthzPort, "healthzPort", s.HealthzPort, "Port for health check")

	// init log
	gofs := goflag.NewFlagSet("klog", goflag.ExitOnError)
	klog.InitFlags(gofs)

	fs.AddGoFlagSet(gofs)

	return fs
}

// Config return a controller config objective
func (s *Options) Config() (*config.Config, error) {
	kubeconfig, err := clientcmd.BuildConfigFromFlags(s.Master, s.Kubeconfig)
	if err != nil {
		return nil, err
	}
	kubeconfig = restclient.AddUserAgent(kubeconfig, UserAgent)
	kubeconfig = setupConfigQPS(kubeconfig)

	client, err := kubernetes.NewForConfig(restclient.AddUserAgent(kubeconfig, UserAgent))
	if err != nil {
		return nil, err
	}

	s.Cfg.Client = client
	c := &config.Config{
		Cfg: s.Cfg,
	}

	return c, nil
}

func setupConfigQPS(c *restclient.Config) *restclient.Config {
	const EnvKubeClientQPS = "ENV_KUBE_CLIENT_QPS"
	const EnvKubeClientBurst = "ENV_KUBE_CLIENT_BURST"
	getQPSEnvInt := func(key string, min int) int {
		v := 0
		s := os.Getenv(key)
		if s != "" {
			v, _ = strconv.Atoi(s)
		}
		if v < min {
			v = min
		}
		return v
	}

	qps := getQPSEnvInt(EnvKubeClientQPS, gocommonclient.DefaultQPS)
	burst := getQPSEnvInt(EnvKubeClientBurst, gocommonclient.DefaultBurst)
	c.QPS = float32(qps)
	c.Burst = burst
	return c
}
