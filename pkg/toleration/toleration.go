package toleration

import (
	netv1alpha1 "github.com/caicloud/loadbalancer-controller/pkg/apis/networking/v1alpha1"
	"k8s.io/client-go/pkg/api/v1"
)

var (
	defaultTolerationKey = []string{netv1alpha1.TaintKey}
)
var (
	// AdditionalTolerationKeys contains additional toleration keys which
	// loadbalancer should tolerate
	AdditionalTolerationKeys = []string{}
)

// GenerateTolerations generates the tolerations
func GenerateTolerations() []v1.Toleration {
	keys := append(defaultTolerationKey, AdditionalTolerationKeys...)
	tolerations := make([]v1.Toleration, 0)

	for _, key := range keys {
		tolerations = append(tolerations, v1.Toleration{
			Key:      key,
			Operator: v1.TolerationOpExists,
		})
	}

	return tolerations
}
