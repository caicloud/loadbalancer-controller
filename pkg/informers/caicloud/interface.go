package caicloud

import (
	"github.com/caicloud/loadbalancer-controller/pkg/informers/caicloud/v1beta1"
	"github.com/caicloud/loadbalancer-controller/pkg/informers/internalinterfaces"
)

// Interface provides access to each of this group's versions.
type Interface interface {
	// V1beta1 provides access to shared informers for resources in V1beta1.
	V1beta1() v1beta1.Interface
}

type group struct {
	internalinterfaces.SharedInformerFactory
}

// New returns a new Interface.
func New(f internalinterfaces.SharedInformerFactory) Interface {
	return &group{f}
}

// V1beta1 returns a new v1beta1.Interface.
func (g *group) V1beta1() v1beta1.Interface {
	return v1beta1.New(g.SharedInformerFactory)
}
