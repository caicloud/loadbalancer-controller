package lb

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/caicloud/loadbalancer-controller/api"
	"k8s.io/client-go/kubernetes/scheme"
	extensions "k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

// CalculateReplicas helps you to calculate replicas of lb
func CalculateReplicas(lb *api.LoadBalancer) int32 {
	var replicas int32

	if lb.Replicas != nil {
		replicas = *lb.Replicas
	}

	if lb.Nodes != nil {
		// use nodes length override replicas
		replicas = int32(len(lb.Nodes.Names))
	}

	return replicas
}

// DeploymentDeepCopy returns a deepcopy for given deployment
func DeploymentDeepCopy(deployment *extensions.Deployment) (*extensions.Deployment, error) {
	objCopy, err := scheme.Scheme.DeepCopy(deployment)
	if err != nil {
		return nil, err
	}
	copied, ok := objCopy.(*extensions.Deployment)
	if !ok {
		return nil, fmt.Errorf("expected Deployment, got %#v", objCopy)
	}
	return copied, nil
}

// RandStringBytesRmndr returns a randome string.
func RandStringBytesRmndr(n int) string {
	rand.Seed(int64(time.Now().Nanosecond()))
	var letterBytes = "abcdefghijklmnopqrstuvwxyz1234567890"
	b := make([]byte, n)
	b[0] = letterBytes[rand.Int63()%26]
	for i := 1; i < n; i++ {
		b[i] = letterBytes[rand.Int63()%int64(len(letterBytes))]
	}
	return string(b)
}
