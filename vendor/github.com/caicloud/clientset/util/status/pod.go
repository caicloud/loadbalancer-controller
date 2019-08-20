package status

import (
	"fmt"
	"strings"

	"k8s.io/api/core/v1"
)

// judgePod judges the current status of pod from Pod.Status without events
func judgePod(pod *v1.Pod) PodStatus {
	if pod == nil {
		return PodStatus{}
	}
	ready := false
	restarts := 0
	readyContainers := 0
	initContainers := len(pod.Spec.InitContainers)
	totalContainers := len(pod.Spec.Containers)
	phase := pod.Status.Phase
	reason := choseReason(string(pod.Status.Phase), pod.Status.Reason)
	message := ""

	if phase == v1.PodPending {
		// detect pending error
		for i := range pod.Status.Conditions {
			condition := pod.Status.Conditions[i]
			// unschedulable error
			if condition.Type == v1.PodScheduled &&
				condition.Status == v1.ConditionFalse &&
				condition.Reason == v1.PodReasonUnschedulable {
				phase = PodError
				reason = condition.Reason
				message = condition.Message
			}
		}
		// use v1.PodScheduled error first
		if phase != PodError {
			// detect pending error from ContainerStatuses
			for i := range pod.Status.ContainerStatuses {
				cs := pod.Status.ContainerStatuses[i]
				if cs.State.Waiting != nil {
					w := cs.State.Waiting
					// CreateContainerConfigError error
					if w.Reason == "CreateContainerConfigError" {
						phase = PodError
						reason = w.Reason
						message = w.Message
					}
				}
			}
		}
	}

	initializing := false
	for i := range pod.Status.InitContainerStatuses {
		container := pod.Status.InitContainerStatuses[i]
		restarts += int(container.RestartCount)
		switch {
		case container.State.Terminated != nil && container.State.Terminated.ExitCode == 0:
			// initialized success
			continue
		case container.State.Terminated != nil:
			// initialization is failed
			reason = fmt.Sprintf("Init:ExitCode:%d", container.State.Terminated.ExitCode)
			if container.State.Terminated.Signal != 0 {
				reason = fmt.Sprintf("Init:Signal:%d", container.State.Terminated.Signal)
			}
			if len(container.State.Terminated.Reason) > 0 {
				reason = "Init:" + container.State.Terminated.Reason
			}
			message = container.State.Terminated.Message
			phase = PodError
			initializing = true
		case container.State.Waiting != nil && len(container.State.Waiting.Reason) > 0 && container.State.Waiting.Reason != "PodInitializing":
			reason = "Init:" + container.State.Waiting.Reason
			message = container.State.Waiting.Message
			phase = PodInitializing
			initializing = true
		default:
			reason = fmt.Sprintf("Init:%d/%d", i, len(pod.Spec.InitContainers))
			message = string(PodInitializing)
			phase = PodInitializing
			initializing = true
		}
		break
	}

	if !initializing {
		for i := len(pod.Status.ContainerStatuses) - 1; i >= 0; i-- {
			container := pod.Status.ContainerStatuses[i]
			restarts += int(container.RestartCount)

			state, stateReason, stateMessage := judgeContainerState(container.State)
			reason = choseReason(reason, stateReason)
			message = chose(message, stateMessage)
			switch state {
			case containerWaiting:
				// There are some backoff state of container located in
				// containerWaiting, we should treat the pod as Error pahse.
				// And when pod is in CrashLoopBackOff, the uesful information
				// is stored in lastTerminationState.
				// please check the code in test case
				if strings.HasSuffix(reason, "BackOff") {
					phase = PodError
					lastState, lastReason, lastMessage := judgeContainerState(container.LastTerminationState)
					if lastState == containerTerminated {
						reason = choseReason(reason, lastReason)
						message = chose(message, lastMessage)
					}
				}
			case containerTerminated:
				if container.State.Terminated.ExitCode != 0 {
					// if container's exit code != 0, we think that pod is in error phase
					phase = PodError
				}
			case containerRunning:
				if container.Ready {
					readyContainers++
				}
			}
		}
	}

	// all containers are ready and container number > 0
	// we think pod is ready
	if readyContainers == totalContainers && readyContainers > 0 {
		ready = true
	}

	if pod.DeletionTimestamp == nil {
		// kubernetes tells us the pod is running, but we recognize
		// that pod is not ready, and no other errors are found above
		// we think the pod is Initializing
		if phase == v1.PodRunning && !ready {
			phase = PodInitializing
		}
	} else {
		// DeletionTimestamp != nil means the pod may be in Terminating

		// In this phase, pod is not ready
		ready = false
		if pod.Status.Reason == NodeUnreachablePodReason {
			phase = v1.PodUnknown
		} else {
			if phase == v1.PodRunning || phase == v1.PodPending {
				// only if phase is Running, change phase to terminating
				phase = PodTerminating
			}
			reason = "Terminating"
		}
	}
	return PodStatus{
		Ready:           ready,
		RestartCount:    int32(restarts),
		ReadyContainers: int32(readyContainers),
		InitContainers:  int32(initContainers),
		TotalContainers: int32(totalContainers),
		NodeName:        pod.Spec.NodeName,
		Phase:           phase,
		Reason:          reason,
		Message:         message,
	}
}

func judgeContainerState(conaitnerState v1.ContainerState) (state containerState, reason, message string) {
	if conaitnerState.Waiting != nil {
		state = containerWaiting
		reason = conaitnerState.Waiting.Reason
		message = conaitnerState.Waiting.Message
	} else if conaitnerState.Terminated != nil {
		state = containerTerminated
		reason = fmt.Sprintf("ExitCode:%d", conaitnerState.Terminated.ExitCode)
		message = conaitnerState.Terminated.Message
		if conaitnerState.Terminated.Signal != 0 {
			reason = fmt.Sprintf("Signal:%d", conaitnerState.Terminated.Signal)
		}
		if conaitnerState.Terminated.Reason != "" {
			reason = conaitnerState.Terminated.Reason
		}
	} else if conaitnerState.Running != nil {
		state = containerRunning
	}
	return
}

func chose(origin, newOne string) string {
	if newOne != "" {
		return newOne
	}
	return origin
}

func choseReason(origin, newOne string) string {
	if origin == "" {
		return newOne
	}

	if newOne != "" && newOne != "Error" {
		return newOne
	}

	return origin
}
