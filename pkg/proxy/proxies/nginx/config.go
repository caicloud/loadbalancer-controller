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
	"reflect"

	lbapi "github.com/caicloud/clientset/pkg/apis/loadbalance/v1alpha2"
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

func (f *nginx) ensureConfigMaps(lb *lbapi.LoadBalancer) error {
	labels := f.selector(lb)

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

	return err
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
