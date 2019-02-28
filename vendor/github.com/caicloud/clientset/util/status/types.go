package status

import (
	"k8s.io/api/core/v1"
)

const (
	// NodeUnreachablePodReason is the reason and message set on a pod
	// when its state cannot be confirmed as kubelet is unresponsive
	// on the node it is (was) running.
	// copy from k8s.io/kubernetes/pkg/util/node
	NodeUnreachablePodReason = "NodeLost"
)

// These are the valid phase of pod.
const (
	// PodPending means the pod has been accepted by the system, but one or more of the containers
	// has not been started. This includes time before being bound to a node, as well as time spent
	// pulling images onto the host.
	PodPending v1.PodPhase = v1.PodPending
	// PodInitializing means:
	//   - some of pod's initContainers are not finished
	//   - some of pod's containers are not running
	PodInitializing v1.PodPhase = "Initializing"
	// PodTerminating means that pod is in terminating
	PodTerminating v1.PodPhase = "Terminating"
	// PodRunning means the pod has been bound to a node and all of the containers have been started.
	// At least one container is still running or is in the process of being restarted.
	PodRunning v1.PodPhase = v1.PodRunning
	// PodSucceeded means that all containers in the pod have voluntarily terminated
	// with a container exit code of 0, and the system is not going to restart any of these containers.
	PodSucceeded v1.PodPhase = v1.PodSucceeded
	// PodFailed means that all containers in the pod have terminated, and at least one container has
	// terminated in a failure (exited with a non-zero exit code or was stopped by the system).
	PodFailed v1.PodPhase = v1.PodFailed
	// PodUnknown means that for some reason the state of the pod could not be obtained, typically due
	// to an error in communicating with the host of the pod.
	PodUnknown v1.PodPhase = v1.PodUnknown
	// PodError means that:
	//   - When pod is initializing, at least one init container is terminated without code 0.
	//   - When pod is terminating, at least one container is terminated without code 0.
	PodError v1.PodPhase = "Error"
)

// PodState indicates the pod is normal or abnormal
type PodState string

const (
	// PodNormal describes that the pod phase is Running or Succeeded.
	PodNormal PodState = "Normal"
	// PodAbnormal describes that the pod phase is Error or Unknown.
	PodAbnormal PodState = "Abnormal"
	// PodUncertain describes that we can not identify what state the pod is in now.
	PodUncertain PodState = "Uncertain"
)

// PodStatus represents the current status of a pod
type PodStatus struct {
	Ready           bool        `json:"ready,omitempty"`
	RestartCount    int32       `json:"restartCount,omitempty"`
	InitContainers  int32       `json:"initContainers,omitempty"`
	ReadyContainers int32       `json:"readyContainers,omitempty"`
	TotalContainers int32       `json:"totalContainers,omitempty"`
	NodeName        string      `json:"nodeName,omitempty"`
	State           PodState    `json:"state,omitempty"`
	Phase           v1.PodPhase `json:"phase,omitempty"`
	Reason          string      `json:"reason,omitempty"`
	Message         string      `json:"message,omitempty"`
}

var (
	// EmptyPodStatus helps you to judge if the status is empty
	EmptyPodStatus = PodStatus{}
)

type containerState string

const (
	containerWaiting    containerState = "waiting"
	containerTerminated containerState = "terminated"
	containerRunning    containerState = "running"
)
