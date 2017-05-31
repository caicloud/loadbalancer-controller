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
	log "github.com/zoumo/logdog"
	"gopkg.in/urfave/cli.v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// NeverStop may be passed to Until to make it never stop.
var NeverStop <-chan struct{} = make(chan struct{})

// RunController start lb controller
func RunController(opts *Options, stopCh <-chan struct{}) error {

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

	controller.Run(5, NeverStop)

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
		if err := RunController(opts, NeverStop); err != nil {
			msg := fmt.Sprintf("running loadbalancer controller failed, with err: %v\n", err)
			return cli.NewExitError(msg, 1)
		}
		return nil
	}

	// sort flags by name
	sort.Sort(cli.FlagsByName(app.Flags))

	app.Run(os.Args)

}
