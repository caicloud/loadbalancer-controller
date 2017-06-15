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

	lbcontroller "github.com/caicloud/loadbalancer-controller/controller"
	"github.com/caicloud/loadbalancer-controller/pkg/tprclient"
	"github.com/caicloud/loadbalancer-controller/provider"
	_ "github.com/caicloud/loadbalancer-controller/provider/providers"
	"github.com/caicloud/loadbalancer-controller/proxy"
	_ "github.com/caicloud/loadbalancer-controller/proxy/proxies"
	"github.com/caicloud/loadbalancer-controller/version"
	log "github.com/zoumo/logdog"
	"gopkg.in/urfave/cli.v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// RunController start lb controller
func RunController(opts *Options, stopCh <-chan struct{}) error {

	log.Notice("Controller Build Information", log.Fields{
		"release": version.RELEASE,
		"commit":  version.COMMIT,
		"repo":    version.REPO,
	})

	log.Info("Controller Running with", log.Fields{
		"debug":     opts.Debug,
		"kubconfig": opts.Kubeconfig,
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

	// create tpr clientset
	tprclientset, err := tprclient.NewForConfig(config)
	if err != nil {
		log.Fatal("Create tpr client error", log.Fields{"err": err})
		return err
	}

	// start a controller on instances of lb
	controller := lbcontroller.NewLoadBalancerController(clientset, tprclientset)

	controller.Run(5, stopCh)

	return nil
}

func main() {
	// fix for avoiding glog Noisy logs
	flag.CommandLine.Parse([]string{})

	app := cli.NewApp()
	app.Name = "loadbalancer-controller"
	app.Version = "v0.1.0"
	app.Compiled = time.Now()
	app.Usage = "k8s loadbalancer resource controller"

	// add flags to app
	opts := NewOptions()
	opts.AddFlags(app)
	proxy.AddFlags(app)
	provider.AddFlags(app)

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
