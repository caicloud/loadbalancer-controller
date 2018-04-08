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
	"flag"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/caicloud/clientset/kubernetes"
	lbcontroller "github.com/caicloud/loadbalancer-controller/pkg/controller"
	_ "github.com/caicloud/loadbalancer-controller/pkg/provider/providers"
	_ "github.com/caicloud/loadbalancer-controller/pkg/proxy/proxies"
	"github.com/caicloud/loadbalancer-controller/pkg/version"
	log "github.com/zoumo/logdog"
	"gopkg.in/urfave/cli.v1"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/clientcmd"
)

// RunController start lb controller
func RunController(opts *Options, stopCh <-chan struct{}) error {

	log.Infof("Controller Build Information, %v", version.Get().Pretty())

	log.Info("Controller Running with", log.Fields{
		"debug":                 opts.Debug,
		"kubconfig":             opts.Kubeconfig,
		"additionalTolerations": opts.Cfg.AdditionalTolerations,
	})

	if opts.Debug {
		log.ApplyOptions(log.DebugLevel)
	} else {
		log.ApplyOptions(log.InfoLevel)
	}

	// build config
	log.Infof("load kubeconfig from %s", opts.Kubeconfig)
	config, err := clientcmd.BuildConfigFromFlags("", opts.Kubeconfig)
	if err != nil {
		log.Fatal("Create kubeconfig error", log.Fields{"err": err})
		return err
	}

	// create clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal("Create kubernetes client error", log.Fields{"err": err})
		return err
	}

	opts.Cfg.Client = clientset
	// start a controller on instances of lb
	controller := lbcontroller.NewLoadBalancerController(opts.Cfg)

	controller.Run(5, stopCh)

	return nil
}

func main() {
	// fix for avoiding glog Noisy logs
	flag.CommandLine.Parse([]string{})

	app := cli.NewApp()
	app.Name = "loadbalancer-controller"
	app.Version = version.Get().Version
	app.Compiled = time.Now()
	app.Usage = "k8s loadbalancer resource controller"

	// add flags to app
	opts := NewOptions()
	opts.AddFlags(app)

	app.Action = func(c *cli.Context) error {
		if err := RunController(opts, wait.NeverStop); err != nil {
			msg := fmt.Sprintf("running loadbalancer controller failed, with err: %v\n", err)
			return cli.NewExitError(msg, 1)
		}
		return nil
	}

	// sort flags by name
	sort.Sort(cli.FlagsByName(app.Flags))

	app.Run(os.Args)

}
