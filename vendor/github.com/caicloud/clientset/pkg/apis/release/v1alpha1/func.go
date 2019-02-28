package v1alpha1

// ResourceStatusFrom returns a new ResourceStatus with phase
func ResourceStatusFrom(phase ResourcePhase) ResourceStatus {
	return ResourceStatus{
		Phase: phase,
	}
}
