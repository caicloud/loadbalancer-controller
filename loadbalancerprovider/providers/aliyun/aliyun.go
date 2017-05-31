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

package aliyun

import (
	"os"

	"k8s.io/client-go/1.5/dynamic"
	"k8s.io/client-go/1.5/kubernetes"
	"k8s.io/client-go/1.5/pkg/api"
	"k8s.io/client-go/1.5/pkg/api/errors"
	"k8s.io/client-go/1.5/pkg/api/unversioned"
	"k8s.io/client-go/1.5/pkg/api/v1"

	tprapi "github.com/caicloud/loadbalancer-controller/api"
	"github.com/caicloud/loadbalancer-controller/controller"
	"github.com/caicloud/loadbalancer-controller/loadbalancerprovider"
	"github.com/golang/glog"
)

var aliyunIngressImage string

// Aliyun custom variables.
var aliyunCustomSlbEndpoint string
var aliyunCustomRegion string
var aliyunCustomZone string

func init() {
	aliyunIngressImage = os.Getenv("INGRESS_ALIYUN_IMAGE")
	if aliyunIngressImage == "" {
		aliyunIngressImage = "cargo.caicloud.io/caicloud/ingress-aliyun:v0.1.0"
	}
	aliyunCustomSlbEndpoint = os.Getenv("ALIYUN_CUSTOM_SLB_ENDPOINT")
	aliyunCustomRegion = os.Getenv("ALIYUN_CUSTOM_REGION_ID")
	aliyunCustomZone = os.Getenv("ALIYUN_CUSTOM_ZONE_ID")
}

func ProbeLoadBalancerPlugin() loadbalancerprovider.LoadBalancerPlugin {
	return &aliyunLoadBalancerPlugin{}
}

var _ loadbalancerprovider.LoadBalancerPlugin = &aliyunLoadBalancerPlugin{}

const (
	aliyunLoadBalancerPluginName = "ingress.alpha.k8s.io/ingress-aliyun"
)

var (
	lbresource = &unversioned.APIResource{Name: "loadbalancers", Kind: "loadbalancer", Namespaced: true}
)

type aliyunLoadBalancerPlugin struct{}

func (plugin *aliyunLoadBalancerPlugin) GetPluginName() string {
	return aliyunLoadBalancerPluginName
}

func (plugin *aliyunLoadBalancerPlugin) CanSupport(claim *tprapi.LoadBalancerClaim) bool {
	if claim == nil || claim.Annotations == nil {
		return false
	}
	return claim.Annotations[controller.IngressProvisioningClassKey] == aliyunLoadBalancerPluginName
}

func (plugin *aliyunLoadBalancerPlugin) NewProvisioner(options loadbalancerprovider.LoadBalancerOptions) loadbalancerprovider.Provisioner {
	if aliyunCustomRegion != "" {
		options.AliyunRegionID = aliyunCustomRegion
	}
	if aliyunCustomZone != "" {
		options.AliyunZoneID = aliyunCustomZone
	}
	return &aliyunLoadBalancerProvisioner{
		options: options,
	}
}

type aliyunLoadBalancerProvisioner struct {
	options loadbalancerprovider.LoadBalancerOptions
}

var _ loadbalancerprovider.Provisioner = &aliyunLoadBalancerProvisioner{}

func (p *aliyunLoadBalancerProvisioner) Provision(clientset *kubernetes.Clientset, dynamicClient *dynamic.Client) (string, error) {
	rc, loadbalancer := p.getReplicationController(), p.getLoadBalancer()

	lbUnstructed, err := loadbalancer.ToUnstructured()
	if err != nil {
		return "", err
	}

	err = func() error {
		if _, err := clientset.Core().ReplicationControllers("kube-system").Create(rc); err != nil {
			return err
		}
		if _, err := dynamicClient.Resource(lbresource, "kube-system").Create(lbUnstructed); err != nil {
			return err
		}
		return nil
	}()

	if err != nil {
		if err = clientset.Core().ReplicationControllers("kube-system").Delete(rc.Name, &api.DeleteOptions{}); err != nil && !errors.IsNotFound(err) {
			glog.Errorf("failed to delete rc due to: %v", err)
		}
		if err = dynamicClient.Resource(lbresource, "kube-system").Delete(lbUnstructed.GetName(), &v1.DeleteOptions{}); err != nil && !errors.IsNotFound(err) {
			glog.Errorf("failed to delete lb due to: %v", err)
		}

		return "", err
	}

	return p.options.LoadBalancerName, nil
}

func (p *aliyunLoadBalancerProvisioner) getLoadBalancer() *tprapi.LoadBalancer {
	return &tprapi.LoadBalancer{
		TypeMeta: unversioned.TypeMeta{
			Kind: "Loadbalancer",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: p.options.LoadBalancerName,
			Annotations: map[string]string{
				"kubernetes.io/createdby": "loadbalancer-aliyun-dynamic-provisioner",
			},
		},
		Spec: tprapi.LoadBalancerSpec{
			AliyunLoadBalancer: &tprapi.AliyunLoadBalancer{
				LoadBalancerName: p.options.LoadBalancerName,
			},
		},
	}
}

func (p *aliyunLoadBalancerProvisioner) getReplicationController() *v1.ReplicationController {
	lbReplicas, terminationGracePeriodSeconds := int32(1), int64(60)
	return &v1.ReplicationController{
		ObjectMeta: v1.ObjectMeta{
			Name: p.options.LoadBalancerName,
			Labels: map[string]string{
				"k8s-app": p.options.LoadBalancerName,
			},
			Annotations: map[string]string{
				"kubernetes.io/createdby": "loadbalancer-aliyun-dynamic-provisioner",
			},
		},
		Spec: v1.ReplicationControllerSpec{
			Replicas: &lbReplicas,
			Template: &v1.PodTemplateSpec{
				ObjectMeta: v1.ObjectMeta{
					Labels: map[string]string{
						"k8s-app": p.options.LoadBalancerName,
					},
				},
				Spec: v1.PodSpec{
					HostNetwork:                   true,
					TerminationGracePeriodSeconds: &terminationGracePeriodSeconds,
					Containers: []v1.Container{
						{
							Name:            "ingress-aliyun",
							Image:           aliyunIngressImage,
							ImagePullPolicy: v1.PullAlways,
							Resources:       p.options.Resources,
							Env: []v1.EnvVar{
								{
									Name:  "CLUSTER_NAME",
									Value: p.options.ClusterName,
								},
								{
									Name:  "LOADBALANCER_NAME",
									Value: p.options.LoadBalancerName,
								},
								{
									Name:  "ALIYUN_ACCESS_KEY_ID",
									Value: p.options.AliyunAccessKeyID,
								},
								{
									Name:  "ALIYUN_ACCESS_KEY_SECRET",
									Value: p.options.AliyunAccessKeySecret,
								},
								{
									Name:  "ALIYUN_REGION_ID",
									Value: p.options.AliyunRegionID,
								},
								{
									Name:  "ALIYUN_ZONE_ID",
									Value: p.options.AliyunZoneID,
								},
								{
									Name:  "SLB_ENDPOINT",
									Value: aliyunCustomSlbEndpoint,
								},
							},
						},
					},
				},
			},
		},
	}
}
