/*
Copyright 2017 The Kubernetes Authors.
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
	"os"
	"time"

	"github.com/golang/glog"
	"github.com/spf13/pflag"

	"k8s.io/client-go/1.5/dynamic"
	"k8s.io/client-go/1.5/kubernetes"
	"k8s.io/client-go/1.5/pkg/api/errors"
	"k8s.io/client-go/1.5/pkg/api/resource"
	"k8s.io/client-go/1.5/pkg/api/unversioned"
	"k8s.io/client-go/1.5/pkg/api/v1"
	"k8s.io/client-go/1.5/pkg/util/intstr"
	"k8s.io/client-go/1.5/pkg/util/wait"
	"k8s.io/client-go/1.5/rest"

	"github.com/caicloud/loadbalancer-controller/controller"
	"github.com/caicloud/loadbalancer-controller/loadbalancerprovider"
	"github.com/caicloud/loadbalancer-controller/loadbalancerprovider/providers"
)

const (
	LoadbalancerClaimName string = "loadbalancerclaims"
	LoadbalancerClaimKind string = "loadbalancerclaim"

	LoadbalancerName string = "loadbalancers"
	LoadbalancerKind string = "loadbalancer"
)

var (
	flags = pflag.NewFlagSet("", pflag.ExitOnError)
)

func init() {
	flag.Set("logtostderr", "true")
	flag.Parse()
	go wait.Until(glog.Flush, 10*time.Second, wait.NeverStop)
}

var defaultBackendName string
var defaultBackendImage string
var defaultBackendLabelSelector map[string]string

func init() {
	defaultBackendImage = os.Getenv("INGRESS_DEFAULT_BACKEND_IMAGE")
	if defaultBackendImage == "" {
		defaultBackendImage = "index.caicloud.io/caicloud/default-http-backend:v0.0.1"
	}
	defaultBackendName = "default-http-backend"
	defaultBackendLabelSelector = map[string]string{"app": "default-http-backend"}
}

// register loadbalancer providers
func init() {
	loadbalancerprovider.RegisterPlugin(nginx.ProbeLoadBalancerPlugin())
}

func main() {
	flags.AddGoFlagSet(flag.CommandLine)
	flags.Parse(os.Args)

	// workaround of noisy log, see https://github.com/kubernetes/kubernetes/issues/17162
	flag.CommandLine.Parse([]string{})

	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err)
	}

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	if err := ensureDefaultBackendService(clientset); err != nil {
		panic(err)
	}

	// create dynamic client
	resources := []*unversioned.APIResourceList{
		{
			GroupVersion: "k8s.io/v1",
			APIResources: []unversioned.APIResource{
				{
					Name:       LoadbalancerClaimName,
					Kind:       LoadbalancerClaimKind,
					Namespaced: true,
				},
				{
					Name:       LoadbalancerName,
					Kind:       LoadbalancerKind,
					Namespaced: true,
				},
			},
		},
	}
	mapper, err := dynamic.NewDiscoveryRESTMapper(resources, dynamic.VersionInterfaces)
	if err != nil {
		panic(err.Error())
	}
	dynamicClient, err := dynamic.NewClientPool(config, mapper, dynamic.LegacyAPIPathResolverFunc).
		ClientForGroupVersionKind(unversioned.GroupVersionKind{Group: "k8s.io", Version: "v1"})
	if err != nil {
		panic(err.Error())
	}

	pc := controller.NewProvisionController(clientset, dynamicClient, loadbalancerprovider.PluginMgr)
	pc.Run(5, wait.NeverStop)
}

// ensureDefaultBackendService ensure a default backend service exists. When nginx
// loadbalancer receives a request which doesn't match any ingress rules, it forwards
// the request to default backend service. default-http-backend will always respond
// with "404".
func ensureDefaultBackendService(clientset *kubernetes.Clientset) error {
	pod := v1.Pod{
		ObjectMeta: v1.ObjectMeta{
			Namespace: "default",
			Name:      defaultBackendName,
			Labels:    defaultBackendLabelSelector,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:            defaultBackendName,
					Image:           defaultBackendImage,
					ImagePullPolicy: v1.PullAlways,
					Resources: v1.ResourceRequirements{
						Requests: v1.ResourceList{
							v1.ResourceCPU:    resource.MustParse("50m"),
							v1.ResourceMemory: resource.MustParse("50Mi"),
						},
						Limits: v1.ResourceList{
							v1.ResourceCPU:    resource.MustParse("50m"),
							v1.ResourceMemory: resource.MustParse("50Mi"),
						},
					},
				},
			},
		},
	}

	svc := v1.Service{
		ObjectMeta: v1.ObjectMeta{
			Namespace: "default",
			Name:      defaultBackendName,
			Labels:    defaultBackendLabelSelector,
		},
		Spec: v1.ServiceSpec{
			Type:            v1.ServiceTypeClusterIP,
			SessionAffinity: v1.ServiceAffinityNone,
			Selector:        defaultBackendLabelSelector,
			Ports: []v1.ServicePort{
				{
					Port:       int32(80),
					TargetPort: intstr.FromInt(80),
					Protocol:   v1.ProtocolTCP,
				},
			},
		},
	}

	if _, err := clientset.Core().Pods("default").Create(&pod); err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	if _, err := clientset.Core().Services("default").Create(&svc); err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	return nil
}
