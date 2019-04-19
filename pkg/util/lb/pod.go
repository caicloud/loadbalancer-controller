package lb

import "k8s.io/api/core/v1"

// IsPodMatchNodeSelectorFailed returns true if pod is in MatchNodeSelector failed
func IsPodMatchNodeSelectorFailed(pod *v1.Pod) bool {
	if pod.Status.Phase == v1.PodFailed && pod.Status.Reason == "MatchNodeSelector" {
		return true
	}
	return false
}
