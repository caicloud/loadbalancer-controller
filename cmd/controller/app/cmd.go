package app

import (
	"github.com/caicloud/loadbalancer-controller/cmd/controller/app/config"
	"github.com/caicloud/loadbalancer-controller/cmd/controller/app/options"
	lbcontroller "github.com/caicloud/loadbalancer-controller/pkg/controller"
	"github.com/caicloud/loadbalancer-controller/pkg/version"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog"
)

// NewCommand creates a *cobra.Command object with default parameters
func NewCommand() *cobra.Command {
	s := options.NewOptions()

	cmd := &cobra.Command{
		Use:  "loadbalancer-controller",
		Long: `k8s loadbalancer resource controller`,
		Run: func(cmd *cobra.Command, args []string) {
			c, err := s.Config()
			if err != nil {
				klog.Exitln(err)
			}
			if err := Run(c, wait.NeverStop); err != nil {
				klog.Exitln(err)
			}
		},
	}

	fs := cmd.Flags()
	fs.AddFlagSet(s.Flags())

	fs.Set("logtostderr", "true")

	return cmd
}

// Run runs the Config.  This should never exit.
func Run(c *config.Config, stopCh <-chan struct{}) error {
	klog.Info(version.Get().Pretty())
	klog.Infof("Controller Running with additionalTolerations %v", c.Cfg.AdditionalTolerations)

	// start a controller on instances of lb
	controller := lbcontroller.NewLoadBalancerController(c.Cfg)
	controller.Run(5, stopCh)

	return nil
}
