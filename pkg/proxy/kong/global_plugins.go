package kong

import (
	"fmt"
	"time"

	cfgv1 "github.com/caicloud/loadbalancer-controller/pkg/kong/apis/configuration/v1"
	kongclient "github.com/caicloud/loadbalancer-controller/pkg/kong/client/clientset/versioned"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
)

func installGlobalPlugins(ingressClass string) error {
	kubeconfig, err := clientcmd.BuildConfigFromFlags("", "")
	if err != nil {
		return err
	}

	client, err := kongclient.NewForConfig(kubeconfig)
	if err != nil {
		return err
	}

	var prometheusPlugin = &cfgv1.KongClusterPlugin{
		ObjectMeta: metav1.ObjectMeta{
			Name: "global-prometheus-" + ingressClass,
			Annotations: map[string]string{
				"kubernetes.io/ingress.class": fmt.Sprintf("kube-system.%s", ingressClass),
			},
			Labels: map[string]string{
				"global": "true",
			},
		},
		PluginName: "prometheus",
		ConfigFrom: cfgv1.NamespacedConfigSource{
			SecretValue: cfgv1.NamespacedSecretValueFromSource{
				Namespace: "",
				Secret:    "",
				Key:       "",
			},
		},
	}

	return wait.Poll(100*time.Millisecond, 3*time.Second, func() (bool, error) {
		_, err := client.ConfigurationV1().KongClusterPlugins().Get(prometheusPlugin.Name, metav1.GetOptions{})
		if err != nil && errors.IsNotFound(err) {
			if _, err := client.ConfigurationV1().KongClusterPlugins().Create(prometheusPlugin); err != nil {
				klog.Errorf("Create global plugin error: %v", err)
				return false, nil
			}
		}

		if err == nil {
			return true, nil
		}
		return false, nil
	})
}
