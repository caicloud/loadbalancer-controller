package v1

// Interface provides access to all the informers in this group version.
type Interface interface {
	Pods() PodLab
}

type version struct {
}

// New returns a new Interface.
func New() Interface {
	return &version{}
}

// Pods returns a PodLab.
func (g *version) Pods() PodLab {
	return &podImpl{}
}
