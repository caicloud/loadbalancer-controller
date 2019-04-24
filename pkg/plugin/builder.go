package plugin

// RegistryBuilder collects functions that add things to a registry. It's to allow
// code to compile without explicitly referencing generated types. You should
// declare one in each package that will have generated deep copy or conversion
// functions.
type RegistryBuilder []func(*Registry) error

// AddToRegistry applies all the stored functions to the registry. A non-nil error
// indicates that one function failed and the attempt was abandoned.
func (rb *RegistryBuilder) AddToRegistry(r *Registry) error {
	for _, f := range *rb {
		if err := f(r); err != nil {
			return err
		}
	}
	return nil
}

// Register adds a registry setup function to the list.
func (rb *RegistryBuilder) Register(funcs ...func(*Registry) error) {
	for _, f := range funcs {
		*rb = append(*rb, f)
	}
}

// NewRegistryBuilder calls Register for you.
func NewRegistryBuilder(funcs ...func(*Registry) error) RegistryBuilder {
	var sb RegistryBuilder
	sb.Register(funcs...)
	return sb
}
