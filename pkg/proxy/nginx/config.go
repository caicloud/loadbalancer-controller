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
	"encoding/json"
	"fmt"
	"reflect"

	lbapi "github.com/caicloud/clientset/pkg/apis/loadbalance/v1alpha2"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	log "k8s.io/klog"
)

var (
	defaultConfig = map[string]string{
		"enable-sticky-sessions": "true",
		"ssl-redirect":           "false",
		"force-ssl-redirect":     "false",
		"enable-vts-status":      "true",
		"proxy-body-size":        "5G",
		"server-tokens":          "false",
		"skip-access-log-urls":   "/nginx_status/format/json",
	}

	// managedConfig is fully controlled by CPS. we should delete these from configmap if they are not specified.
	managedConfig = map[string]string{
		"proxy-buffer-size":          "",
		"proxy-buffers-number":       "",
		"limit-conn-zone-variable":   "",
		"proxy-read-timeout":         "",
		"proxy-send-timeout":         "",
		"upstream-keepalive-timeout": "",
	}

	annotationExternalConfigMaps = "external-configs"
	labelExternalConfig          = "loadbalance.caicloud.io/external-config"
)

func mapDel(base map[string]string, dels ...map[string]string) map[string]string {
	ret := make(map[string]string)
	for k, v := range base {
		has := false
		for _, del := range dels {
			if _, has = del[k]; has {
				break
			}
		}
		if !has {
			ret[k] = v
		}
	}

	return ret
}
func mapAdd(base map[string]string, adds ...map[string]string) map[string]string {
	ret := make(map[string]string)
	for k, v := range base {
		ret[k] = v
	}

	for _, add := range adds {
		for k, v := range add {
			ret[k] = v
		}
	}

	return ret
}

func (f *nginx) ensureConfigMaps(lb *lbapi.LoadBalancer) error {
	labels := f.selector(lb)

	// For ingress-nginx configuration
	cmName := fmt.Sprintf(configMapName, lb.Name)
	cm, err := f.ensureConfigMap(cmName, lb.Namespace, labels)
	if err != nil {
		return err
	}

	err = f.updateConfig(lb, cm)
	if err != nil {
		return err
	}

	// For L4 TCP rules
	tcpcmName := fmt.Sprintf(tcpConfigMapName, lb.Name)
	_, err = f.ensureConfigMap(tcpcmName, lb.Namespace, labels)
	if err != nil {
		return err
	}

	// For L4 UDP rules
	udpcmName := fmt.Sprintf(udpConfigMapName, lb.Name)
	_, err = f.ensureConfigMap(udpcmName, lb.Namespace, labels)

	return err
}

func (f *nginx) ensureConfigMap(name, namespace string, labels map[string]string) (*v1.ConfigMap, error) {
	cm, err := f.client.CoreV1().ConfigMaps(namespace).Get(name, metav1.GetOptions{})

	if err == nil {
		return cm, nil
	}

	if !errors.IsNotFound(err) {
		return nil, err
	}
	cm = &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
	}
	log.Infof("About to craete ConfigMap %v/%v for proxy", namespace, cm.Name)
	return f.client.CoreV1().ConfigMaps(namespace).Create(cm)
}

func (f *nginx) updateConfig(lb *lbapi.LoadBalancer, cm *v1.ConfigMap) error {
	// 1. if cm has unmanaged config, we should generate cm.Data with old method
	done, err := f.updateIfHasUnmanagedConfig(lb, cm)
	if err != nil || done {
		return err
	}

	// 2. otherwise, use new method that generate cm.Data by external configs
	//externalConfigMaps, preset, override, err := f.getExternalConfig(lb)
	externalConfigMaps, preset, override, err := f.getExternalConfig(lb)
	if err != nil {
		return err
	}
	newConfig := mapAdd(preset, defaultConfig, lb.Spec.Proxy.Config, override)
	bs, _ := json.Marshal(externalConfigMaps)
	externalConfigMapsStr := string(bs)

	if reflect.DeepEqual(cm.Data, newConfig) && externalConfigMapsStr == cm.Annotations[annotationExternalConfigMaps] {
		return nil
	}

	if cm.Annotations == nil {
		cm.Annotations = make(map[string]string)
	}
	cm.Annotations[annotationExternalConfigMaps] = externalConfigMapsStr
	cm.Data = newConfig
	log.Infof("About to update ConfigMap %v/%v data, with exnternal configs: %v", cm.Namespace, cm.Name, externalConfigMapsStr)
	_, err = f.client.CoreV1().ConfigMaps(cm.Namespace).Update(cm)
	return err
}

func (f *nginx) updateIfHasUnmanagedConfig(lb *lbapi.LoadBalancer, cm *v1.ConfigMap) (bool, error) {
	var err error

	var oldExternalConfigMaps []string
	if value, ok := cm.Annotations[annotationExternalConfigMaps]; ok {
		err = json.Unmarshal([]byte(value), &oldExternalConfigMaps)
		if err != nil {
			return false, err
		}
	}

	var unmanagedConifgs map[string]string
	// we consider that configs may contains unmanaged config only if cm is not marked (oldExternalConfigMaps is empty).
	if len(oldExternalConfigMaps) == 0 {
		unmanagedConifgs = mapDel(cm.Data, defaultConfig, managedConfig)
	}

	// if unmanagedConifgs is empty, we can update configs with new method safetly.
	// otherwise, we should keep using old method to not lose user's unmanaged conifgs.
	if len(unmanagedConifgs) == 0 {
		return false, nil
	}

	log.Warningf("Found unmanaged configs in %s/%s: %v", lb.Namespace, lb.Name, unmanagedConifgs)
	newConfig := mapAdd(unmanagedConifgs, defaultConfig, lb.Spec.Proxy.Config)

	if !reflect.DeepEqual(cm.Data, newConfig) {
		cm.Data = newConfig
		log.Warningf("About to update ConfigMap %v/%v data with old method", cm.Namespace, cm.Name)
		_, err = f.client.CoreV1().ConfigMaps(cm.Namespace).Update(cm)
	}
	return true, err
}

func (f *nginx) getExternalConfig(lb *lbapi.LoadBalancer) ([]string, map[string]string, map[string]string, error) {

	presetConfigMaps := []string{}
	overrideConfigMaps := []string{}
	presetConfig := make(map[string]string)
	overrideConfig := make(map[string]string)

	for _, scope := range []string{"platform", "cluster", "instance-%s"} {

		for _, postfix := range []string{"", "-override"} {
			name := fmt.Sprintf("cfg-lb-nginx-%s%s", scope, postfix)
			if scope == "instance-%s" {
				name = fmt.Sprintf(name, lb.Name)
			}
			cm, err := f.client.CoreV1().ConfigMaps(lb.Namespace).Get(name, metav1.GetOptions{})
			if errors.IsNotFound(err) {
				continue
			}
			if err != nil {
				return nil, nil, nil, err
			}

			if _, ok := cm.Labels[labelExternalConfig]; !ok {
				log.Infof("external configmap: %s doesn't has a '%s' label", cm.Name, labelExternalConfig)
				continue
			}

			if postfix == "" {
				presetConfig = mapAdd(presetConfig, cm.Data)
				presetConfigMaps = append(presetConfigMaps, cm.Name)
				continue
			}
			overrideConfig = mapAdd(overrideConfig, cm.Data)
			overrideConfigMaps = append(overrideConfigMaps, cm.Name)
		}
	}

	presetConfigMaps = append(presetConfigMaps, "") // empty string delimits 'preset' and 'override'
	presetConfigMaps = append(presetConfigMaps, overrideConfigMaps...)
	return presetConfigMaps, presetConfig, overrideConfig, nil
}
