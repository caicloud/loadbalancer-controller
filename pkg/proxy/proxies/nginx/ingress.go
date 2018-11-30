package nginx

import (
	"fmt"
	"strconv"

	lbapi "github.com/caicloud/clientset/pkg/apis/loadbalance/v1alpha2"
	"github.com/caicloud/loadbalancer-controller/pkg/api"
	"github.com/caicloud/loadbalancer-controller/pkg/toleration"
	lbutil "github.com/caicloud/loadbalancer-controller/pkg/util/lb"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
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
)

func (f *nginx) generateDeployment(lb *lbapi.LoadBalancer) *appsv1.Deployment {
	terminationGracePeriodSeconds := int64(30)
	hostNetwork := true
	dnsPolicy := v1.DNSClusterFirstWithHostNet
	replicas, needNodeAffinity := lbutil.CalculateReplicas(lb)
	maxSurge := intstr.FromInt(0)
	t := true
	labels := f.selector(lb)

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
	podAffinity := &v1.PodAntiAffinity{
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
					// host network ?
					HostNetwork: hostNetwork,
					DNSPolicy:   dnsPolicy,
					// TODO
					TerminationGracePeriodSeconds: &terminationGracePeriodSeconds,
					Affinity: &v1.Affinity{
						// don't co-locate pods of this deployment in same node
						PodAntiAffinity: podAffinity,
					},
					Tolerations: toleration.GenerateTolerations(),
					Containers: []v1.Container{
						{
							Name:            "proxy",
							Image:           f.image,
							ImagePullPolicy: v1.PullAlways,
							Resources:       lb.Spec.Proxy.Resources,
							Ports: []v1.ContainerPort{
								{
									ContainerPort: 80,
								},
								{
									ContainerPort: 443,
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
							},
							ReadinessProbe: &v1.Probe{
								Handler: v1.Handler{
									HTTPGet: &v1.HTTPGetAction{
										Path:   healthCheckPath,
										Port:   intstr.FromInt(80),
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
										Port:   intstr.FromInt(80),
										Scheme: v1.URISchemeHTTP,
									},
								},
							},
						},
						{
							Name:            "sidecar",
							Image:           f.sidecar,
							ImagePullPolicy: v1.PullAlways,
							Resources: v1.ResourceRequirements{
								Limits: v1.ResourceList{
									v1.ResourceCPU:    resource.MustParse("100m"),
									v1.ResourceMemory: resource.MustParse("50Mi"),
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
						},
					},
				},
			},
		},
	}

	if needNodeAffinity {
		// decide running on which node
		deploy.Spec.Template.Spec.Affinity.NodeAffinity = nodeAffinity
	}

	if f.defaultSSLCertificate != "" {
		deploy.Spec.Template.Spec.Containers[0].Args = append(
			deploy.Spec.Template.Spec.Containers[0].Args,
			"--default-ssl-certificate="+f.defaultSSLCertificate,
		)
	}

	return deploy
}
