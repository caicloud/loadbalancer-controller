package v1

import (
	"k8s.io/api/core/v1"
	"k8s.io/kubernetes/pkg/api/legacyscheme"
	k8spodutil "k8s.io/kubernetes/pkg/api/pod"
	k8score "k8s.io/kubernetes/pkg/apis/core"
)

var (
	DefaultPodLab = &podImpl{}
)

type PodLab interface {
	DropDisabledAlphaFields(in *v1.PodSpec)
}

type podImpl struct{}

// DropDisabledAlphaFields removes disabled fields from the pod spec.
// This should be called from PrepareForCreate/PrepareForUpdate for all resources containing a pod spec.
//
// TODO: the feature in utilfeature.DefaultFeatureGate must be the same as apiserver
func (l *podImpl) DropDisabledAlphaFields(in *v1.PodSpec) {
	out := &k8score.PodSpec{}
	legacyscheme.Scheme.Convert(in, out, nil)
	// drop disabled alpha fields in podSpec
	k8spodutil.DropDisabledAlphaFields(out)
	legacyscheme.Scheme.Convert(out, in, nil)
}
