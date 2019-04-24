package azure

import "github.com/caicloud/loadbalancer-controller/pkg/plugin"

func AddToRegistry(registry *plugin.Registry) error {
	registry.Register(providerName, New())
	return nil
}
