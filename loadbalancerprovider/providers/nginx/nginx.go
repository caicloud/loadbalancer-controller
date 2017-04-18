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

package nginx

import (
	"encoding/json"
	"os"

	tprapi "github.com/caicloud/loadbalancer-controller/api"
	"github.com/caicloud/loadbalancer-controller/controller"
	"github.com/caicloud/loadbalancer-controller/loadbalancerprovider"
	"github.com/golang/glog"

	"k8s.io/client-go/1.5/dynamic"
	"k8s.io/client-go/1.5/kubernetes"
	"k8s.io/client-go/1.5/pkg/api"
	"k8s.io/client-go/1.5/pkg/api/errors"
	"k8s.io/client-go/1.5/pkg/api/resource"
	"k8s.io/client-go/1.5/pkg/api/unversioned"
	"k8s.io/client-go/1.5/pkg/api/v1"
	"k8s.io/client-go/1.5/pkg/util/intstr"
)

var keepalibedImage, nginxIngressImage string

const (
	KeepalivedVIPKey  = "keepalived.k8s.io/vip"
	KeepalivedVRIDKey = "keepalived.k8s.io/vrid"
)

// read keepalivedImage and nginxIngressImage from environment variable; or
// fallover to default images.
func init() {
	keepalibedImage = os.Getenv("INGRESS_KEEPALIVED_IMAGE")
	if keepalibedImage == "" {
		keepalibedImage = "cargo.caicloud.io/caicloud/keepalived-sidecar:v0.1.0"
	}
	nginxIngressImage = os.Getenv("INGRESS_NGINX_IMAGE")
	if nginxIngressImage == "" {
		nginxIngressImage = "cargo.caicloud.io/caicloud/ingress-nginx:v0.1.0"
	}
}

// ProbeLoadBalancerPlugin returns nginx loadbalancer plugin.
func ProbeLoadBalancerPlugin() loadbalancerprovider.LoadBalancerPlugin {
	return &nginxLoadBalancerPlugin{}
}

const (
	nginxLoadBalancerPluginName = "ingress.alpha.k8s.io/ingress-nginx"
	ingressRoleLabelKey         = "ingress.alpha.k8s.io/role"
)

var (
	lbresource = &unversioned.APIResource{Name: "loadbalancers", Kind: "loadbalancer", Namespaced: true}
)

var _ loadbalancerprovider.LoadBalancerPlugin = &nginxLoadBalancerPlugin{}

type nginxLoadBalancerPlugin struct{}

// GetPluginName implements LoadBalancerPlugin interface.
func (plugin *nginxLoadBalancerPlugin) GetPluginName() string {
	return nginxLoadBalancerPluginName
}

// CanSupport implement LoadBalancerPlugin interface.
func (plugin *nginxLoadBalancerPlugin) CanSupport(claim *tprapi.LoadBalancerClaim) bool {
	if claim == nil || claim.Annotations == nil {
		return false
	}
	return claim.Annotations[controller.IngressProvisioningClassKey] == nginxLoadBalancerPluginName
}

// NewProvisioner implements LoadBalancerPlugin interface.
// It returns a nginx loadbalancer provisioner.
func (plugin *nginxLoadBalancerPlugin) NewProvisioner(options loadbalancerprovider.LoadBalancerOptions) loadbalancerprovider.Provisioner {
	return &nginxLoadbalancerProvisioner{
		options: options,
	}
}

type nginxLoadbalancerProvisioner struct {
	options loadbalancerprovider.LoadBalancerOptions
}

var _ loadbalancerprovider.Provisioner = &nginxLoadbalancerProvisioner{}

// Provision provisions a nginx LoadBalancer, including ingress controllers, configurations,
// services, etc; where ingress controller listens to ingress rules and does the actual forwarding.
func (p *nginxLoadbalancerProvisioner) Provision(clientset *kubernetes.Clientset, dynamicClient *dynamic.Client) (string, error) {
	service, rc, configmap, loadbalancer := p.getService(), p.getReplicationController(), p.getConfigMap(), p.getLoadBalancer()
	tcpCM, udpCM := p.getTCPConfigMap(), p.getUDPConfigMap()

	lbUnstructed, err := loadbalancer.ToUnstructured()
	if err != nil {
		return "", err
	}

	err = func() error {
		if _, err := clientset.Core().Services("kube-system").Create(service); err != nil {
			return err
		}
		if _, err := clientset.Core().ReplicationControllers("kube-system").Create(rc); err != nil {
			return err
		}
		if _, err := clientset.Core().ConfigMaps("kube-system").Create(configmap); err != nil {
			return err
		}
		if _, err := clientset.Core().ConfigMaps("kube-system").Create(tcpCM); err != nil {
			return err
		}
		if _, err := clientset.Core().ConfigMaps("kube-system").Create(udpCM); err != nil {
			return err
		}
		if _, err := dynamicClient.Resource(lbresource, "kube-system").Create(lbUnstructed); err != nil {
			return err
		}
		return nil
	}()

	if err != nil {
		if err = clientset.Core().Services("kube-system").Delete(service.Name, nil); err != nil && !errors.IsNotFound(err) {
			glog.Errorf("failed to delete service due to: %v", err)
		}
		if err = clientset.Core().ReplicationControllers("kube-system").Delete(rc.Name, nil); err != nil && !errors.IsNotFound(err) {
			glog.Errorf("failed to delete rc due to: %v", err)
		}
		if err = clientset.Core().ConfigMaps("kube-system").Delete(configmap.Name, nil); err != nil && !errors.IsNotFound(err) {
			glog.Errorf("failed to delete configmap due to: %v", err)
		}
		if err = clientset.Core().ConfigMaps("kube-system").Delete(tcpCM.Name, nil); err != nil && !errors.IsNotFound(err) {
			glog.Errorf("failed to delete tcp configmap due to: %v", err)
		}
		if err = clientset.Core().ConfigMaps("kube-system").Delete(udpCM.Name, nil); err != nil && !errors.IsNotFound(err) {
			glog.Errorf("failed to delete udp configmap due to: %v", err)
		}
		if err = dynamicClient.Resource(lbresource, "kube-system").Delete(lbUnstructed.GetName(), nil); err != nil && !errors.IsNotFound(err) {
			glog.Errorf("failed to delete lb due to: %v", err)
		}

		return "", err
	}

	return p.options.LoadBalancerName, nil
}

// getLoadBalancer returns a LoadBalancer API object to represent the provisioned LoadBalancer.
func (p *nginxLoadbalancerProvisioner) getLoadBalancer() *tprapi.LoadBalancer {
	return &tprapi.LoadBalancer{
		TypeMeta: unversioned.TypeMeta{
			Kind: "Loadbalancer",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: p.options.LoadBalancerName,
			Labels: map[string]string{
				"kubernetes.io/createdby": "loadbalancer-nginx-dynamic-provisioner",
			},
			Annotations: map[string]string{
				controller.IngressParameterVIPKey: p.options.LoadBalancerVIP,
			},
		},
		Spec: tprapi.LoadBalancerSpec{
			NginxLoadBalancer: &tprapi.NginxLoadBalancer{
				Service: v1.ObjectReference{
					Kind:      "Service",
					Namespace: "kube-system",
					Name:      p.options.LoadBalancerName,
				},
			},
		},
	}
}

// getService returns a service to be created along with nginx LoadBalancer; the service
// is used to reference and dereference ingress <-> loadbalancer.
func (p *nginxLoadbalancerProvisioner) getService() *v1.Service {
	return &v1.Service{
		ObjectMeta: v1.ObjectMeta{
			Name: p.options.LoadBalancerName,
			Labels: map[string]string{
				"kubernetes.io/createdby": "loadbalancer-nginx-dynamic-provisioner",
			},
			Annotations: map[string]string{
				KeepalivedVIPKey:  p.options.LoadBalancerVIP,
				KeepalivedVRIDKey: p.options.LoadBalancerVRID,
			},
		},
		Spec: v1.ServiceSpec{
			Selector: map[string]string{
				"k8s-app": p.options.LoadBalancerName,
			},
			Ports: []v1.ServicePort{
				{
					Port:       80,
					TargetPort: intstr.FromInt(80),
				},
			},
		},
	}
}

// getConfigMap returns a configmap used to configure nginx ingress controller, for
// details, see https://github.com/kubernetes/ingress/blob/master/controllers/nginx/configuration.md
func (p *nginxLoadbalancerProvisioner) getConfigMap() *v1.ConfigMap {
	return &v1.ConfigMap{
		ObjectMeta: v1.ObjectMeta{
			Name: p.options.LoadBalancerName,
			Labels: map[string]string{
				"k8s-app":                 p.options.LoadBalancerName,
				"kubernetes.io/createdby": "loadbalancer-nginx-dynamic-provisioner",
			},
		},
		Data: map[string]string{
			"enable-sticky-sessions": "true",
		},
	}
}

// getTCPConfigMap returns a configmap used to configure nginx ingress controller, for
// details, see https://github.com/kubernetes/ingress/blob/master/controllers/nginx/configuration.md
func (p *nginxLoadbalancerProvisioner) getTCPConfigMap() *v1.ConfigMap {
	return &v1.ConfigMap{
		ObjectMeta: v1.ObjectMeta{
			Name: p.options.LoadBalancerName + "-tcp",
			Labels: map[string]string{
				"k8s-app":                 p.options.LoadBalancerName,
				"kubernetes.io/createdby": "loadbalancer-nginx-dynamic-provisioner",
			},
		},
		Data: map[string]string{},
	}
}

// getUDPConfigMap returns a configmap used to configure nginx ingress controller, for
// details, see https://github.com/kubernetes/ingress/blob/master/controllers/nginx/configuration.md
func (p *nginxLoadbalancerProvisioner) getUDPConfigMap() *v1.ConfigMap {
	return &v1.ConfigMap{
		ObjectMeta: v1.ObjectMeta{
			Name: p.options.LoadBalancerName + "-udp",
			Labels: map[string]string{
				"k8s-app":                 p.options.LoadBalancerName,
				"kubernetes.io/createdby": "loadbalancer-nginx-dynamic-provisioner",
			},
		},
		Data: map[string]string{},
	}
}

// getReplicationController returns RC for nginx controller.
func (p *nginxLoadbalancerProvisioner) getReplicationController() *v1.ReplicationController {
	nginxlbReplicas, terminationGracePeriodSeconds, nginxlbPrivileged := int32(2), int64(60), true

	lbTolerations, _ := json.Marshal([]api.Toleration{{Key: "dedicated", Value: "loadbalancer", Effect: api.TaintEffectPreferNoSchedule}})

	nodeAffinity := &v1.NodeAffinity{
		RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{
			NodeSelectorTerms: []v1.NodeSelectorTerm{{
				MatchExpressions: []v1.NodeSelectorRequirement{{
					Key: ingressRoleLabelKey, Operator: v1.NodeSelectorOpIn, Values: []string{"loadbalancer"},
				}},
			}},
		},
	}

	podAntiAffinity := &v1.PodAntiAffinity{
		RequiredDuringSchedulingIgnoredDuringExecution: []v1.PodAffinityTerm{
			{
				LabelSelector: &unversioned.LabelSelector{
					MatchLabels: map[string]string{
						"k8s-app": p.options.LoadBalancerName,
					},
				},
				TopologyKey: unversioned.LabelHostname,
			},
		},
	}
	affinityAnnotation, _ := json.Marshal(v1.Affinity{NodeAffinity: nodeAffinity, PodAntiAffinity: podAntiAffinity})

	return &v1.ReplicationController{
		ObjectMeta: v1.ObjectMeta{
			Name: p.options.LoadBalancerName,
			Labels: map[string]string{
				"k8s-app":                 p.options.LoadBalancerName,
				"kubernetes.io/createdby": "loadbalancer-nginx-dynamic-provisioner",
			},
		},
		Spec: v1.ReplicationControllerSpec{
			Replicas: &nginxlbReplicas,
			Template: &v1.PodTemplateSpec{
				ObjectMeta: v1.ObjectMeta{
					Labels: map[string]string{
						"k8s-app": p.options.LoadBalancerName,
					},
					Annotations: map[string]string{
						api.TolerationsAnnotationKey: string(lbTolerations),
						api.AffinityAnnotationKey:    string(affinityAnnotation),
					},
				},
				Spec: v1.PodSpec{
					HostNetwork:                   true,
					TerminationGracePeriodSeconds: &terminationGracePeriodSeconds,
					Containers: []v1.Container{
						{
							Name:            "keepalived",
							Image:           keepalibedImage,
							ImagePullPolicy: v1.PullAlways,
							Resources: v1.ResourceRequirements{
								Requests: v1.ResourceList{
									v1.ResourceCPU:    resource.MustParse("100m"),
									v1.ResourceMemory: resource.MustParse("100Mi"),
								},
								Limits: v1.ResourceList{
									v1.ResourceCPU:    resource.MustParse("100m"),
									v1.ResourceMemory: resource.MustParse("100Mi"),
								},
							},
							SecurityContext: &v1.SecurityContext{
								Privileged: &nginxlbPrivileged,
							},
							Env: []v1.EnvVar{
								{
									Name: "POD_NAME",
									ValueFrom: &v1.EnvVarSource{
										FieldRef: &v1.ObjectFieldSelector{
											FieldPath: "metadata.name",
										},
									},
								},
								{
									Name: "POD_NAMESPACE",
									ValueFrom: &v1.EnvVarSource{
										FieldRef: &v1.ObjectFieldSelector{
											FieldPath: "metadata.namespace",
										},
									},
								},
								{
									Name:  "SERVICE_NAME",
									Value: p.options.LoadBalancerName,
								},
							},
						},
						{
							Name:            "ingress-nginx",
							Image:           nginxIngressImage,
							ImagePullPolicy: v1.PullAlways,
							Resources:       p.options.Resources,
							ReadinessProbe: &v1.Probe{
								Handler: v1.Handler{
									HTTPGet: &v1.HTTPGetAction{
										Path:   "/healthz",
										Port:   intstr.FromInt(10254),
										Scheme: v1.URISchemeHTTP,
									},
								},
							},
							LivenessProbe: &v1.Probe{
								Handler: v1.Handler{
									HTTPGet: &v1.HTTPGetAction{
										Path:   "/healthz",
										Port:   intstr.FromInt(10254),
										Scheme: v1.URISchemeHTTP,
									},
								},
								InitialDelaySeconds: 10,
							},
							Env: []v1.EnvVar{
								{
									Name: "POD_NAME",
									ValueFrom: &v1.EnvVarSource{
										FieldRef: &v1.ObjectFieldSelector{
											FieldPath: "metadata.name",
										},
									},
								},
								{
									Name: "POD_NAMESPACE",
									ValueFrom: &v1.EnvVarSource{
										FieldRef: &v1.ObjectFieldSelector{
											FieldPath: "metadata.namespace",
										},
									},
								},
							},
							Ports: []v1.ContainerPort{
								{
									ContainerPort: 80,
									//HostPort:      80,
									HostIP: p.options.LoadBalancerVIP,
								},
								{
									ContainerPort: 443,
									//HostPort:      443,
									HostIP: p.options.LoadBalancerVIP,
								},
							},
							Args: []string{
								"/nginx-ingress-controller",
								"--default-backend-service=default/default-http-backend",
								"--configmap=kube-system/" + p.options.LoadBalancerName,
								"--ingress-class=" + p.options.LoadBalancerName,
								"--tcp-services-configmap=kube-system/" + p.options.LoadBalancerName + "-tcp",
								"--udp-services-configmap=kube-system/" + p.options.LoadBalancerName + "-udp",
								"--healthz-port=" + "10254",
								"--health-check-path=" + "/healthz",
							},
						},
					},
				},
			},
		},
	}
}
