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
	"reflect"
	"strings"
	"time"

	"github.com/caicloud/clientset/kubernetes"
	lbapi "github.com/caicloud/clientset/pkg/apis/loadbalance/v1alpha2"
	"github.com/caicloud/clientset/util/status"
	"github.com/caicloud/loadbalancer-controller/pkg/api"
	stringsutil "github.com/caicloud/loadbalancer-controller/pkg/util/strings"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	// NodeUnreachablePodReason is the reason and message set on a pod
	// when its state cannot be confirmed as kubelet is unresponsive
	// on the node it is (was) running.
	// copy from k8s.io/kubernetes/pkg/util/node
	NodeUnreachablePodReason = "NodeLost"
)

// DefaultRetry is the recommended retry for a conflict where multiple clients
// are making changes to the same resource.
var DefaultRetry = wait.Backoff{
	Steps:    5,
	Duration: 10 * time.Millisecond,
	Factor:   1.0,
	Jitter:   0.1,
}

// SortPodStatusByName ...
type SortPodStatusByName []lbapi.PodStatus

func (s SortPodStatusByName) Len() int {
	return len(s)
}

func (s SortPodStatusByName) Less(i, j int) bool {
	return s[i].Name < s[j].Name
}

func (s SortPodStatusByName) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// SplitNamespaceAndNameByDot returns the namespace and name that
// encoded into the label or value by dot
func SplitNamespaceAndNameByDot(value string) (namespace, name string, err error) {
	parts := strings.Split(value, ".")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("unexpected format: %q", value)
	}

	return parts[0], parts[1], nil
}

// ProxyStatusEqual check whether the given two PorxyStatuses are equal
func ProxyStatusEqual(a, b lbapi.ProxyStatus) bool {

	if !PodStatusesEqual(a.PodStatuses, b.PodStatuses) {
		return false
	}
	a.PodStatuses = lbapi.PodStatuses{}
	b.PodStatuses = lbapi.PodStatuses{}
	return reflect.DeepEqual(a, b)
}

// IpvsdrProviderStatusEqual check whether the given two Statuses are equal
func IpvsdrProviderStatusEqual(a, b lbapi.IpvsdrProviderStatus) bool {
	if !PodStatusesEqual(a.PodStatuses, b.PodStatuses) {
		return false
	}
	a.PodStatuses = lbapi.PodStatuses{}
	b.PodStatuses = lbapi.PodStatuses{}
	return reflect.DeepEqual(a, b)
}

// ExternalProviderStatusEqual check whether the given two ExpternalProviderStatus are equal
func ExternalProviderStatusEqual(a, b lbapi.ExpternalProviderStatus) bool {
	return reflect.DeepEqual(a, b)
}

// PodStatusesEqual check whether the given two PodStatuses are equal
func PodStatusesEqual(a, b lbapi.PodStatuses) bool {
	aStatus := a.Statuses
	bStatus := b.Statuses

	if len(aStatus) != len(bStatus) {
		return false
	}

	a.Statuses = nil
	b.Statuses = nil

	if !reflect.DeepEqual(a, b) {
		return false
	}

	for _, as := range aStatus {
		equal := false
		for _, bs := range bStatus {
			if as.Name == bs.Name {
				equal = reflect.DeepEqual(as, bs)
				break
			}
		}
		if !equal {
			return false
		}
	}

	return true
}

// CalculateReplicas helps you to calculate replicas of lb
// determines if you need to add node affinity
func CalculateReplicas(lb *lbapi.LoadBalancer) (int32, bool) {
	var replicas int32
	var hostnetwork bool

	if lb.Spec.Nodes.Replicas != nil {
		replicas = *lb.Spec.Nodes.Replicas
	}

	if len(lb.Spec.Nodes.Names) != 0 {
		// use nodes length override replicas
		replicas = int32(len(lb.Spec.Nodes.Names))
		hostnetwork = true
	}

	return replicas, hostnetwork
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

// ComputePodStatus computes the pod's current status
func ComputePodStatus(pod *v1.Pod) lbapi.PodStatus {
	s := status.JudgePodStatus(pod, nil)
	return lbapi.PodStatus{
		Name:            pod.Name,
		Ready:           s.Ready,
		NodeName:        s.NodeName,
		ReadyContainers: s.ReadyContainers,
		TotalContainers: s.TotalContainers,
		Phase:           string(s.Phase),
		Reason:          s.Reason,
		Message:         s.Message,
	}
}

// IsStatic checks if lb is a static loadbalancer
func IsStatic(lb *lbapi.LoadBalancer) bool {
	if lb == nil {
		return false
	}
	anno := lb.Annotations
	if anno == nil {
		return false
	}
	_, ok := anno[api.KeyStatic]
	return ok
}

// EvictPod deletes the pod scheduled to the wrong node
func EvictPod(client kubernetes.Interface, lb *lbapi.LoadBalancer, pod *v1.Pod) {
	if len(lb.Spec.Nodes.Names) == 0 {
		return
	}

	// fix: avoid evicting unscheduled pod
	if pod.Spec.NodeName == "" {
		return
	}

	evict := func() {
		client.CoreV1().Pods(pod.Namespace).Delete(pod.Name, &metav1.DeleteOptions{})
	}

	// FIXME: when RequiredDuringSchedulingRequiredDuringExecution finished
	// This is a special issue.
	// There is bug when the nodes.Names change。
	// According to nodeAffinity RequiredDuringSchedulingIgnoredDuringExecution,
	// the system may or may not try to eventually evict the pod from its node.
	// the pod may still running on the wrong node, so we evict it manually
	if !stringsutil.StringInSlice(pod.Spec.NodeName, lb.Spec.Nodes.Names) &&
		pod.DeletionTimestamp == nil {
		evict()
		return
	}

	// evict pod MatchNodeSelector Failed
	if IsPodMatchNodeSelectorFailed(pod) && pod.DeletionTimestamp == nil {
		evict()
		return
	}
}
