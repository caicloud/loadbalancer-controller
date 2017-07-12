package nginx

import (
	"fmt"
	"reflect"

	netv1alpha1 "github.com/caicloud/loadbalancer-controller/pkg/apis/networking/v1alpha1"
	log "github.com/zoumo/logdog"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/pkg/api/v1"
)

var (
	defaultConfig = map[string]string{
		"enable-sticky-sessions": "true",
		"ssl-redirect":           "false",
		"force-ssl-redirect":     "false",
		"enable-vts-status":      "true",
	}
)

func merge(dst, src map[string]string) map[string]string {
	ret := make(map[string]string)

	for k, v := range dst {
		ret[k] = v
	}
	for k, v := range src {
		ret[k] = v
	}

	return ret
}

func (f *nginx) ensureConfigMaps(lb *netv1alpha1.LoadBalancer) error {
	labels := map[string]string{
		netv1alpha1.LabelKeyCreatedBy: fmt.Sprintf(netv1alpha1.LabelValueFormatCreateby, lb.Namespace, lb.Name),
		netv1alpha1.LabelKeyProxy:     "nginx",
	}
	cmName := fmt.Sprintf(configMapName, lb.Name)
	config := merge(defaultConfig, lb.Spec.Proxy.Config)
	err := f.ensureConfigMap(cmName, lb.Namespace, labels, config)
	if err != nil {
		return err
	}
	tcpcmName := fmt.Sprintf(tcpConfigMapName, lb.Name)
	err = f.ensureConfigMap(tcpcmName, lb.Namespace, labels, nil)
	if err != nil {
		return err
	}
	udpcmName := fmt.Sprintf(udpConfigMapName, lb.Name)
	err = f.ensureConfigMap(udpcmName, lb.Namespace, labels, nil)
	if err != nil {
		return err
	}

	return nil
}

func (f *nginx) ensureConfigMap(name, namespace string, labels, data map[string]string) error {
	cm, err := f.client.CoreV1().ConfigMaps(namespace).Get(name, metav1.GetOptions{})

	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	if errors.IsNotFound(err) {
		cm = &v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:   name,
				Labels: labels,
			},
			Data: data,
		}
		log.Info("About to craete ConfigMap for proxy", log.Fields{"cm.ns": namespace, "cm.name": cm.Name})
		_, nerr := f.client.CoreV1().ConfigMaps(namespace).Create(cm)
		if nerr != nil {
			return nerr
		}
	}

	if data == nil {
		// do not update data if data == nil
		// tcp and udp config map will be changed by other app
		// controller create it only
		return nil
	}

	if reflect.DeepEqual(cm.Data, data) {
		return nil
	}

	cm.Data = data
	log.Info("About to update ConfigMap data", log.Fields{"cm.ns": namespace, "cm.name": cm.Name})
	_, err = f.client.CoreV1().ConfigMaps(namespace).Update(cm)

	return err
}
