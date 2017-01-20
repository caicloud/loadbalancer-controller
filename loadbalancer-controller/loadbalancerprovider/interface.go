package loadbalancerprovider

import (
	"fmt"
	"strings"
	"sync"

	"github.com/caicloud/ingress-admin/loadbalancer-controller/api"

	"k8s.io/client-go/1.5/dynamic"
	"k8s.io/client-go/1.5/kubernetes"
	"k8s.io/client-go/1.5/pkg/api/v1"
	"k8s.io/client-go/1.5/pkg/util/validation"
)

type LoadBalancerPlugin interface {
	GetPluginName() string
	CanSupport(spec *api.LoadBalancerClaim) bool
	NewProvisioner(options LoadBalancerOptions) Provisioner
}

type LoadBalancerOptions struct {
	Resources        v1.ResourceRequirements
	LoadBalancerName string
	LoadBalancerVIP  string
}

type Provisioner interface {
	Provision(clientset *kubernetes.Clientset, dynamicClient *dynamic.Client) (string, error)
}

// LoadBalancerPluginMgr tracks registered plugins.
type LoadBalancerPluginMgr struct {
	mutex   sync.Mutex
	plugins map[string]LoadBalancerPlugin
}

var PluginMgr LoadBalancerPluginMgr = LoadBalancerPluginMgr{
	mutex:   sync.Mutex{},
	plugins: map[string]LoadBalancerPlugin{},
}

func RegisterPlugin(plugin LoadBalancerPlugin) error {
	PluginMgr.mutex.Lock()
	defer PluginMgr.mutex.Unlock()

	name := plugin.GetPluginName()
	if errs := validation.IsQualifiedName(name); len(errs) != 0 {
		return fmt.Errorf("volume plugin has invalid name: %q: %s", name, strings.Join(errs, ";"))
	}
	if _, found := PluginMgr.plugins[name]; found {
		return fmt.Errorf("volume plugin %q was registered more than once", name)
	}

	PluginMgr.plugins[name] = plugin

	return nil
}

// FindPluginBySpec find a matching loadbalancer provider by LoadBalancerClaimSpec
func (pm *LoadBalancerPluginMgr) FindPluginBySpec(claim *api.LoadBalancerClaim) (LoadBalancerPlugin, error) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	matches := []string{}
	for k, v := range pm.plugins {
		if v.CanSupport(claim) {
			matches = append(matches, k)
		}
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("no ingress service plugin matched")
	}
	if len(matches) > 1 {
		return nil, fmt.Errorf("multiple ingress service plugins matched: %s", strings.Join(matches, ","))
	}
	return pm.plugins[matches[0]], nil
}
