package app

import (
	"github.com/caicloud/loadbalancer-controller/cmd/controller/app/config"
	lbcontroller "github.com/caicloud/loadbalancer-controller/pkg/controller"
	"github.com/caicloud/loadbalancer-controller/pkg/version"

	"k8s.io/klog"
)

// Run runs the Config.  This should never exit.
func Run(c *config.Config, stopCh <-chan struct{}) error {
	klog.Info(version.Get().Pretty())
	klog.Infof("Controller Running with additionalTolerations %v", c.Cfg.AdditionalTolerations)

	// start a controller on instances of lb
	controller := lbcontroller.NewLoadBalancerController(c.Cfg)
	controller.Run(5, stopCh)

	return nil
}
