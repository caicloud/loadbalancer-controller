package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const GroupName = "devops.caicloud.io"

// GroupVersion is the identifier for the API which includes
// the name of the group and the version of the API
var SchemeGroupVersion = schema.GroupVersion{
	Group:   GroupName,
	Version: "v1",
}

// create a SchemeBuilder which uses functions to add types to the scheme
var (
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)
	AddToScheme   = SchemeBuilder.AddToScheme
)

func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

// addKnownTypes adds our types to the API scheme by registering
// Cargo and CargoList
func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(
		SchemeGroupVersion,
		&Cargo{},
		&CargoList{},
	)

	// register the type in the scheme
	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}
