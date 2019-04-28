package v1

// Interface provides access to all the informers in this group version.
type Interface interface {
	ObjectMeta() ObjectMetaLab
}

type version struct {
}

// New returns a new Interface.
func New() Interface {
	return &version{}
}

// ObjectMeta returns a ObjectMetaLab.
func (g *version) ObjectMeta() ObjectMetaLab {
	return &objectMetaImpl{}
}
