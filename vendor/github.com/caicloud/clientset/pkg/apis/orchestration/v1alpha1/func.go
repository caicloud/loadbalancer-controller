package v1alpha1

// ID implements gonum.org/v1/gonum/graph.Node interface
func (n Vertex) ID() int64 {
	return n.Index
}
