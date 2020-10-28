package kong

import (
	"fmt"

	lbapi "github.com/caicloud/clientset/pkg/apis/loadbalance/v1alpha2"
	"github.com/caicloud/loadbalancer-controller/pkg/api"
	"github.com/caicloud/loadbalancer-controller/pkg/toleration"
	lbutil "github.com/caicloud/loadbalancer-controller/pkg/util/lb"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func (k *kong) generateDeployment(lb *lbapi.LoadBalancer) *appsv1.Deployment {
	replicas, _ := lbutil.CalculateReplicas(lb)
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

	t := true
	proxyContainer := v1.Container{
		Name:  "proxy",
		Image: k.proxyImage,
		Env:   makeProxyEnvs(lb, httpPort, httpsPort),
		Ports: []v1.ContainerPort{
			{
				ContainerPort: int32(httpPort),
				Name:          "proxy",
				Protocol:      v1.ProtocolTCP,
			},
			{
				ContainerPort: int32(httpsPort),
				Name:          "proxy-ssl",
				Protocol:      v1.ProtocolTCP,
			},
			{
				ContainerPort: 8100,
				Name:          "metrics",
				Protocol:      v1.ProtocolTCP,
			},
		},
		LivenessProbe:  makeProbe("/status", 8100),
		ReadinessProbe: makeProbe("/status", 8100),
		Lifecycle: &v1.Lifecycle{
			PreStop: &v1.Handler{
				Exec: &v1.ExecAction{Command: []string{"/bin/sh", "-c", "kong quit"}},
			},
		},
		ImagePullPolicy: v1.PullAlways,
		SecurityContext: &v1.SecurityContext{
			RunAsUser: func(i int64) *int64 { return &i }(0),
		},
	}

	ingressContainer := v1.Container{
		Name:  "ingress-controller",
		Image: k.ingressImage,
		Ports: []v1.ContainerPort{
			{
				ContainerPort: 8080,
				Name:          "webhook",
				Protocol:      v1.ProtocolTCP,
			},
		},
		Env:             makeIngressEnvs(lb.Name),
		Resources:       lb.Spec.Proxy.Resources,
		LivenessProbe:   makeProbe("/healthz", 10254),
		ReadinessProbe:  makeProbe("/healthz", 10254),
		ImagePullPolicy: v1.PullAlways,
	}

	labels := k.selector(lb)
	terminationGracePeriodSeconds := int64(30)
	maxSurge := intstr.FromInt(0)
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      lb.Name + proxyNameSuffix + "-" + lbutil.RandStringBytesRmndr(5),
			Namespace: "kube-system",
			Labels:    labels,
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
						"prometheus.io/port":                           "8100",
						"prometheus.io/scrape":                         "true",
						"traffic.sidecar.istio.io/includeInboundPorts": "",
					},
				},
				Spec: v1.PodSpec{
					Containers:                    []v1.Container{proxyContainer, ingressContainer},
					TerminationGracePeriodSeconds: &terminationGracePeriodSeconds,
					DNSPolicy:                     v1.DNSClusterFirstWithHostNet,
					HostNetwork:                   true,
					Affinity: &v1.Affinity{
						NodeAffinity:    nodeAffinity,
						PodAntiAffinity: podAntiAffinity,
					},
					Tolerations:       toleration.GenerateTolerations(),
					PriorityClassName: "system-node-critical",
				},
			},
		},
	}
}

func makeProbe(path string, port int32) *v1.Probe {
	return &v1.Probe{
		Handler: v1.Handler{
			HTTPGet: &v1.HTTPGetAction{
				Path: path,
				Port: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: port,
				},
				Scheme: "HTTP",
			},
		},
		InitialDelaySeconds: 5,
		TimeoutSeconds:      1,
		PeriodSeconds:       10,
		SuccessThreshold:    1,
		FailureThreshold:    3,
	}
}
func makeProxyEnvs(lb *lbapi.LoadBalancer, httpPort, httpsPort int) []v1.EnvVar {
	envs := []v1.EnvVar{
		{
			Name:  "KONG_DATABASE",
			Value: "off",
		},
		{
			Name:  "KONG_PROXY_LISTEN",
			Value: fmt.Sprintf("0.0.0.0:%v, 0.0.0.0:%v ssl", httpPort, httpsPort),
		},
		{
			Name:  "KONG_ADMIN_LISTEN",
			Value: "127.0.0.1:8444 ssl",
		},
		{
			Name:  "KONG_STATUS_LISTEN",
			Value: "0.0.0.0:8100",
		},
		{
			Name:  "KONG_NGINX_WORKER_PROCESSES",
			Value: "1",
		},
		{
			Name:  "KONG_ADMIN_ACCESS_LOG",
			Value: "/dev/stdout",
		},
		{
			Name:  "KONG_ADMIN_ERROR_LOG",
			Value: "/dev/stderr",
		},
		{
			Name:  "KONG_PROXY_ERROR_LOG",
			Value: "/dev/stderr",
		},
	}

	// 128k
	if bufferSize, found := lb.Spec.Proxy.Config["proxy-buffers-size"]; found {
		envs = append(envs, v1.EnvVar{
			Name:  "KONG_NGINX_HTTP_PROXY_BUFFER_SIZE",
			Value: bufferSize,
		})
	}

	// 32 32k
	if buffer, found := lb.Spec.Proxy.Config["proxy-buffers"]; found {
		envs = append(envs, v1.EnvVar{
			Name:  "KONG_NGINX_HTTP_PROXY_BUFFERS",
			Value: buffer,
		})
	}

	// 128k
	if busyBufferSize, found := lb.Spec.Proxy.Config["proxy-busy-buffers-size"]; found {
		envs = append(envs, v1.EnvVar{
			Name:  "KONG_NGINX_HTTP_PROXY_BUSY_BUFFERS_SIZE",
			Value: busyBufferSize,
		})
	}

	if readTimeout, found := lb.Spec.Proxy.Config["proxy-read-timeout"]; found {
		envs = append(envs, v1.EnvVar{
			Name:  "KONG_NGINX_HTTP_PROXY_READ_TIMEOUT",
			Value: readTimeout,
		})
	}
	return envs
}

func makeIngressEnvs(ingressName string) []v1.EnvVar {
	return []v1.EnvVar{
		{
			Name:  "CONTROLLER_KONG_ADMIN_URL",
			Value: "https://127.0.0.1:8444",
		},
		{
			Name:  "CONTROLLER_KONG_ADMIN_TLS_SKIP_VERIFY",
			Value: "true",
		},
		{
			Name:  "CONTROLLER_PUBLISH_SERVICE",
			Value: fmt.Sprintf("kube-system/kong-proxy-%s", ingressName),
		},
		{
			Name:  "CONTROLLER_INGRESS_CLASS",
			Value: fmt.Sprintf("kube-system.%s", ingressName),
		},
		{
			Name: "POD_NAME",
			ValueFrom: &v1.EnvVarSource{
				FieldRef: &v1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "metadata.name",
				},
			},
		},
		{
			Name: "POD_NAMESPACE",
			ValueFrom: &v1.EnvVarSource{
				FieldRef: &v1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "metadata.namespace",
				},
			},
		},
	}
}
