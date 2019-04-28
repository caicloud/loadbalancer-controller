package kubelab

import "github.com/caicloud/loadbalancer-controller/pkg/kubelab/apps"

// Interface provides useful utils for resources in all known API group versions
type Interface interface {
	Apps() apps.Interface
}

type kubelab struct{}

// New constructs a new instance of a Kubelib
func New() Interface {
	return &kubelab{}
}

func (l *kubelab) Apps() apps.Interface {
	return apps.New()
}
