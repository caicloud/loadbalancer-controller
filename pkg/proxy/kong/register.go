package kong

import "github.com/caicloud/loadbalancer-controller/pkg/plugin"

func AddToRegistry(registry *plugin.Registry) error {
	registry.Register(proxyName, New())
	return nil
}
