package kong

import (
	"encoding/json"
	"reflect"

	lbapi "github.com/caicloud/clientset/pkg/apis/loadbalance/v1alpha2"
	"github.com/caicloud/loadbalancer-controller/pkg/api"
	libmetav1 "github.com/caicloud/loadbalancer-controller/pkg/kubelab/meta/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	corev1 "k8s.io/kubernetes/pkg/apis/core/v1"
)

func (k *kong) ensureService(lb *lbapi.LoadBalancer) error {
	if err := k.ensureProxySvc(lb); err != nil {
		return err
	}

	return k.ensureValidationWebhookSvc(lb)
}

func (k *kong) ensureProxySvc(lb *lbapi.LoadBalancer) error {
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

	labels := k.selector(lb)
	name := "kong-proxy-" + lb.Name
	t := true
	proxySvc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: metav1.NamespaceSystem,
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
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Name:     "proxy",
					Port:     int32(httpPort),
					Protocol: v1.ProtocolTCP,
				},
				{
					Name:     "proxy-ssl",
					Port:     int32(httpsPort),
					Protocol: v1.ProtocolTCP,
				},
			},
			Selector: labels,
			Type:     v1.ServiceTypeLoadBalancer,
		},
	}

	svc, err := k.client.CoreV1().Services(metav1.NamespaceSystem).Get(name, metav1.GetOptions{})
	if err != nil && errors.IsNotFound(err) {
		if _, err := k.client.CoreV1().Services(metav1.NamespaceSystem).Create(proxySvc); err != nil {
			return err
		}
	}

	if err == nil && !isEqual(svc, proxySvc) {
		newSvc := svc.DeepCopy()
		newSvc.Labels = proxySvc.Labels
		newSvc.OwnerReferences = proxySvc.OwnerReferences
		newSvc.Spec.Ports = proxySvc.Spec.Ports
		newSvc.Spec.Selector = proxySvc.Spec.Selector
		if _, err := k.client.CoreV1().Services(metav1.NamespaceSystem).Update(newSvc); err != nil {
			return err
		}
	}
	return err
}

func (k *kong) ensureValidationWebhookSvc(lb *lbapi.LoadBalancer) error {
	name := "kong-validation-webhook-" + lb.Name
	t := true
	labels := k.selector(lb)
	webhookSvc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: metav1.NamespaceSystem,
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
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Name:     "webhook",
					Port:     443,
					Protocol: v1.ProtocolTCP,
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 8080,
					},
				},
			},
			Selector: labels,
			Type:     v1.ServiceTypeClusterIP,
		},
	}

	svc, err := k.client.CoreV1().Services(metav1.NamespaceSystem).Get(name, metav1.GetOptions{})
	if err != nil && errors.IsNotFound(err) {
		if _, err := k.client.CoreV1().Services(metav1.NamespaceSystem).Create(webhookSvc); err != nil {
			return err
		}
	}

	if err == nil && !isEqual(svc, webhookSvc) {
		newSvc := svc.DeepCopy()
		newSvc.Labels = webhookSvc.Labels
		newSvc.OwnerReferences = webhookSvc.OwnerReferences
		newSvc.Spec.Ports = webhookSvc.Spec.Ports
		newSvc.Spec.Selector = webhookSvc.Spec.Selector
		if _, err := k.client.CoreV1().Services(metav1.NamespaceSystem).Update(newSvc); err != nil {
			return err
		}
	}
	return err
}

func isEqual(dst, src *v1.Service) bool {
	dstCopy := dst.DeepCopy()
	srcCopy := src.DeepCopy()

	corev1.SetObjectDefaults_Service(dstCopy)
	corev1.SetObjectDefaults_Service(srcCopy)

	if !libmetav1.DefaultObjectMetaLab.IsEqual(&dstCopy.ObjectMeta, &srcCopy.ObjectMeta) {
		return false
	}
	dstCopy.Spec.ClusterIP = ""
	srcCopy.Spec.ClusterIP = ""

	dstByte, _ := json.Marshal(dstCopy)
	srcByte, _ := json.Marshal(srcCopy)

	return reflect.DeepEqual(dstByte, srcByte)
}
