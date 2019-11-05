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

package nginx

import (
	"fmt"
	"strconv"

	lbapi "github.com/caicloud/clientset/pkg/apis/loadbalance/v1alpha2"
	"github.com/caicloud/loadbalancer-controller/pkg/api"
	"github.com/caicloud/loadbalancer-controller/pkg/toleration"
	lbutil "github.com/caicloud/loadbalancer-controller/pkg/util/lb"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// ingress controller flags
const (
	configMapName    = "%s-proxy-nginx-config"
	tcpConfigMapName = "%s-proxy-nginx-tcp"
	udpConfigMapName = "%s-proxy-nginx-udp"
	healthCheckPath  = "/healthz"
	// ingress controller use this port to export metrics and pprof information
	ingressControllerPort = 450
	// ingress controller use this port to export nginx status page
	ingressStatusPort = 451
	// ingress controller use this priority class to create pod
	ingressPriorityClass = "system-node-critical"
)

func (f *nginx) generateDeployment(lb *lbapi.LoadBalancer) *appsv1.Deployment {
	terminationGracePeriodSeconds := int64(30)
	dnsPolicy := v1.DNSClusterFirst
	replicas, hostNetwork := lbutil.CalculateReplicas(lb)
	maxSurge := intstr.FromInt(0)
	t := true
	labels := f.selector(lb)
	affinity := v1.Affinity{}

	// run on this node
	nodeAffinity := &v1.NodeAffinity{
		RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{
			NodeSelectorTerms: []v1.NodeSelectorTerm{
				{
					MatchExpressions: []v1.NodeSelectorRequirement{
						{
							Key:      fmt.Sprintf(lbapi.UniqueLabelKeyFormat, lb.Namespace, lb.Name),
							Operator: v1.NodeSelectorOpIn,
							Values:   []string{"true"},
						},
					},
				},
			},
		},
	}

	// do not run with this pod
	podAntiAffinity := &v1.PodAntiAffinity{
		RequiredDuringSchedulingIgnoredDuringExecution: []v1.PodAffinityTerm{
			{
				LabelSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						lbapi.LabelKeyProxy: proxyName,
					},
				},
				TopologyKey: api.LabelHostname,
			},
		},
	}

	httpPort := lb.Spec.Proxy.HTTPPort
	httpsPort := lb.Spec.Proxy.HTTPSPort
	if httpPort <= 0 {
		// default http port is 80
		httpPort = 80
	}
	if httpsPort <= 0 {
		// default https port is 443
		httpsPort = 443
	}

	ingressContainer := v1.Container{
		Name:            "proxy",
		Image:           f.image,
		ImagePullPolicy: v1.PullAlways,
		Resources:       lb.Spec.Proxy.Resources,
		Ports: []v1.ContainerPort{
			{
				ContainerPort: int32(httpPort),
			},
			{
				ContainerPort: int32(httpsPort),
			},
			{
				ContainerPort: ingressControllerPort,
			},
			{
				ContainerPort: ingressStatusPort,
			},
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
		// TODO
		Args: []string{
			"/nginx-ingress-controller",
			"--default-backend-service=" + fmt.Sprintf("%s/%s", defaultHTTPBackendNamespace, defaultHTTPBackendName),
			"--ingress-class=" + fmt.Sprintf(lbapi.LabelValueFormatCreateby, lb.Namespace, lb.Name),
			"--configmap=" + fmt.Sprintf("%s/"+configMapName, lb.Namespace, lb.Name),
			"--tcp-services-configmap=" + fmt.Sprintf("%s/"+tcpConfigMapName, lb.Namespace, lb.Name),
			"--udp-services-configmap=" + fmt.Sprintf("%s/"+udpConfigMapName, lb.Namespace, lb.Name),
			"--health-check-path=" + healthCheckPath,
			"--healthz-port=" + strconv.Itoa(ingressControllerPort),
			"--status-port=" + strconv.Itoa(ingressStatusPort),
			"--annotations-prefix=" + f.annotationPrefix,
			"--enable-ssl-passthrough",
			"--enable-ssl-chain-completion=false",
			"--http-port=" + strconv.Itoa(httpPort),
			"--https-port=" + strconv.Itoa(httpsPort),
		},
		ReadinessProbe: &v1.Probe{
			Handler: v1.Handler{
				HTTPGet: &v1.HTTPGetAction{
					Path:   healthCheckPath,
					Port:   intstr.FromInt(httpPort),
					Scheme: v1.URISchemeHTTP,
				},
			},
		},
		LivenessProbe: &v1.Probe{
			// wait 120s before liveness probe is initiated
			InitialDelaySeconds: 120,
			Handler: v1.Handler{
				HTTPGet: &v1.HTTPGetAction{
					Path:   healthCheckPath,
					Port:   intstr.FromInt(httpPort),
					Scheme: v1.URISchemeHTTP,
				},
			},
		},
	}

	sidecarContainer := v1.Container{
		Name:            "sidecar",
		Image:           f.sidecar,
		ImagePullPolicy: v1.PullAlways,
		Resources: v1.ResourceRequirements{
			Limits: v1.ResourceList{
				v1.ResourceCPU:    resource.MustParse("20m"),
				v1.ResourceMemory: resource.MustParse("100Mi"),
			},
			Requests: v1.ResourceList{
				v1.ResourceCPU:    resource.MustParse("20m"),
				v1.ResourceMemory: resource.MustParse("100Mi"),
			},
		},
		SecurityContext: &v1.SecurityContext{
			// ingress controller sidecar need provileged to change sysctl settings
			Privileged: &t,
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
				Name:  "LOADBALANCER_NAMESPACE",
				Value: lb.Namespace,
			},
			{
				Name:  "LOADBALANCER_NAME",
				Value: lb.Name,
			},
		},
	}

	if f.defaultSSLCertificate != "" {
		ingressContainer.Args = append(ingressContainer.Args, "--default-ssl-certificate="+f.defaultSSLCertificate)
	}

	containers := []v1.Container{
		ingressContainer,
	}

	if hostNetwork {
		// decide running on which node
		affinity.NodeAffinity = nodeAffinity
		// don't co-locate pods of this deployment in same node
		affinity.PodAntiAffinity = podAntiAffinity
		// add sidecar container
		containers = append(containers, sidecarContainer)
		// change dns policy in hostnetwork
		dnsPolicy = v1.DNSClusterFirstWithHostNet
	}

	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:   lb.Name + proxyNameSuffix + "-" + lbutil.RandStringBytesRmndr(5),
			Labels: labels,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion:         api.ControllerKind.GroupVersion().String(),
					Kind:               api.ControllerKind.Kind,
					Name:               lb.Name,
					UID:                lb.UID,
					Controller:         &t,
					BlockOwnerDeletion: &t,
				},
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Strategy: appsv1.DeploymentStrategy{
				RollingUpdate: &appsv1.RollingUpdateDeployment{
					MaxSurge: &maxSurge,
				},
			},
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
					Annotations: map[string]string{
						"prometheus.io/port":   strconv.Itoa(ingressControllerPort),
						"prometheus.io/scrape": "true",
					},
				},
				Spec: v1.PodSpec{
					HostNetwork: hostNetwork,
					DNSPolicy:   dnsPolicy,
					// TODO
					TerminationGracePeriodSeconds: &terminationGracePeriodSeconds,
					Affinity:                      &affinity,
					Tolerations:                   toleration.GenerateTolerations(),
					Containers:                    containers,
					PriorityClassName:             ingressPriorityClass,
				},
			},
		},
	}

	return deploy
}
