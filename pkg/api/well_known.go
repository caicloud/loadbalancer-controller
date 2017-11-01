package api

import (
	lbapi "github.com/caicloud/clientset/pkg/apis/loadbalance/v1alpha2"
)

const (
	// LabelHostname ...
	LabelHostname = "kubernetes.io/hostname"

	// LoadBalancerKind ...
	LoadBalancerKind = "LoadBalancer"
)

var (
	// ControllerKind contains the schema.GroupVersionKind for this controller type.
	ControllerKind = lbapi.SchemeGroupVersion.WithKind(LoadBalancerKind)
)
