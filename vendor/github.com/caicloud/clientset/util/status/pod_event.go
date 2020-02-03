package status

import (
	"sort"
	"time"

	"github.com/caicloud/clientset/util/event"

	"k8s.io/api/core/v1"
)

var (
	// error event case will change the pod's status no matter what state it is in now
	errorEventCases = []event.EventCase{
		{
			// Liveness probe failed
			EventType: v1.EventTypeWarning,
			Reason:    event.ContainerUnhealthy,
			MsgKeys:   []string{"Liveness probe failed"},
		},
		{
			// failed to mount volume
			EventType: v1.EventTypeWarning,
			Reason:    event.FailedMountVolume,
		},
	}

	// warning event case change the pod's status only if it is not in PodRunning, PodSucceeded
	warningEventCase = []event.EventCase{
		{
			// Readiness probe failed
			EventType: v1.EventTypeWarning,
			Reason:    event.ContainerUnhealthy,
			MsgKeys:   []string{"Readiness probe failed"},
		},
	}
)

// JudgePodStatus judges the current status of pod from Pod.Status
// and correct it with events.
func JudgePodStatus(pod *v1.Pod, events []*v1.Event) PodStatus {
	if pod == nil {
		return PodStatus{}
	}

	status := judgePod(pod)
	// only the latest event is useful
	e := getLatestEventForPod(pod, events)

	// If the liveness probe fails, count and lastTimestamp of event will be updated (it performed again).
	// But success doesn't create a new event or update the event.
	// So if the duration from event's lastTimestamp to now is longer than threshold timeout, we suppose that the
	// container is healthy.
	// We detect pod's phase based on the latest event, but
	// if failure occurs to any of the pod's containers and then they become healthy,
	// the latest event still shows unhealthy, but the pod's phase should be healthy.
	// So we ignore the abnormal event in this condition
	if e != nil && e.Reason == event.ContainerUnhealthy {
		var threshold int32
		for _, c := range pod.Spec.Containers {
			if c.LivenessProbe != nil {
				lp := c.LivenessProbe
				t := (lp.TimeoutSeconds + lp.PeriodSeconds) * (lp.FailureThreshold)
				if t > threshold {
					threshold = t
				}
			}
		}
		if threshold != 0 && time.Since(e.LastTimestamp.Time) > time.Duration(threshold)*time.Second {
			e = nil
		}
	}

	// error event case will change the pod's status no matter what state it is in now
	for _, c := range errorEventCases {
		if c.Match(e) {
			status.Phase = PodError
			status.Reason = e.Reason
			status.Message = e.Message
			break
		}
	}

	if status.Phase != PodRunning && status.Phase != PodSucceeded {
		// warning event case change the pod's status only if it is not PodRunning, PodSucceeded
		for _, c := range warningEventCase {
			if c.Match(e) {
				status.Phase = PodError
				status.Reason = e.Reason
				status.Message = e.Message
				break
			}
		}
	}

	switch status.Phase {
	case PodRunning, PodSucceeded:
		status.State = PodNormal
		// when phase == Succeeded, the pod is not ready
		// status.Ready = true
	case PodFailed, PodError, PodUnknown:
		status.State = PodAbnormal
		status.Ready = false
	default:
		status.State = PodUncertain
	}

	return status
}

func getLatestEventForPod(pod *v1.Pod, events []*v1.Event) *v1.Event {
	if len(events) == 0 {
		return nil
	}
	ret := make([]*v1.Event, 0)

	for _, e := range events {
		if e.InvolvedObject.Kind == "Pod" &&
			e.InvolvedObject.Name == pod.Name &&
			e.InvolvedObject.Namespace == pod.Namespace &&
			e.InvolvedObject.UID == pod.UID {
			ret = append(ret, e)
		}
	}

	if len(ret) == 0 {
		return nil
	}

	sort.Sort(event.EventByLastTimestamp(ret))
	return ret[0]
}
