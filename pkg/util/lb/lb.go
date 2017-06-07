/*
Copyright 2017 Caicloud authors. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package lb

import (
	"fmt"
	"math/rand"
	"time"

	netv1alpha1 "github.com/caicloud/loadbalancer-controller/pkg/apis/networking/v1alpha1"
	"k8s.io/client-go/kubernetes/scheme"
	extensions "k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

// CalculateReplicas helps you to calculate replicas of lb
// determines if you need to add node affinity
func CalculateReplicas(lb *netv1alpha1.LoadBalancer) (int32, bool) {
	var replicas int32
	var needNodeAffinity bool

	if lb.Spec.Type == netv1alpha1.LoadBalancerTypeInternal && lb.Spec.Nodes.Replicas != nil {
		replicas = *lb.Spec.Nodes.Replicas
	}

	if len(lb.Spec.Nodes.Names) != 0 {
		// use nodes length override replicas
		replicas = int32(len(lb.Spec.Nodes.Names))
		needNodeAffinity = true
	}

	return replicas, needNodeAffinity
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
