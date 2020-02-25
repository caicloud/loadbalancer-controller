package v1

// Interface provides access to all the informers in this group version.
type Interface interface {
	Deployments() DeploymentLab
}

type version struct {
}

// New returns a new Interface.
func New() Interface {
	return &version{}
}

// Deployments returns a DeploymentLab.
func (g *version) Deployments() DeploymentLab {
	return &deploymentImpl{}
}
