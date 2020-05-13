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
	"context"
	"os"

	"k8s.io/klog"

	"github.com/caicloud/loadbalancer-controller/cmd/controller/app"
	"github.com/caicloud/loadbalancer-controller/cmd/controller/app/options"

	"github.com/caicloud/go-common/kubernetes/leaderelection"
	"github.com/caicloud/go-common/signal"
)

func main() {
	s := options.NewOptions()
	_ = s.Flags().Set("logtostderr", "true")
	if err := s.Flags().Parse(os.Args); err != nil {
		klog.Exitln(err)
	}

	c, err := s.Config()
	if err != nil {
		klog.Exitln(err)
	}

	stopCh := signal.SetupStopSignalHandler()
	leaderelection.RunOrDie(leaderelection.Option{
		LeaseLockName:      "loadbalancer-controller",
		LeaseLockNamespace: "default",
		KubeClient:         c.Cfg.Client,
		Run: func(_ context.Context) {
			_ = app.Run(c, stopCh)
		},
		Port:   s.HealthzPort,
		StopCh: stopCh,
	})
}
