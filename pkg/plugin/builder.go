/*
Copyright 2017 Caicloud authors. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
