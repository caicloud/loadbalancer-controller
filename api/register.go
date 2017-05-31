package api

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8scheme "k8s.io/client-go/kubernetes/scheme"
)

const (
	// GroupName is the group name use in this package
	GroupName = "alpha.caicloud.io"

	// Version is the Version use in this package
	Version = "v1alpha1"

	// LoadBalancerTPRName for third party resource
	LoadBalancerTPRName = "load-balancer"

	// LoadBalancerName is the name of loadbalancer
	LoadBalancerName = "loadbalancer"

	// LoadBalancerPlural is plural of loadbalancer
	LoadBalancerPlural = "loadbalancers"

	// LoadBalancerKind for TypeMeta
	LoadBalancerKind = "LoadBalancer"
)

var (
	// SchemeBuilder ...
	SchemeBuilder = runtime.NewSchemeBuilder()

	// SchemeGroupVersion is group version used to register these objects
	SchemeGroupVersion = schema.GroupVersion{Group: GroupName, Version: Version}
)

// Resource takes an unqualified resource and returns a Group qualified GroupResource
func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

func init() {
	// We only register manually written functions here. The registration of the
	// generated functions takes place in the generated files. The separation
	// makes the code compile even when the generated files are missing.
	SchemeBuilder.Register(addKnownTypes)

	SchemeBuilder.AddToScheme(k8scheme.Scheme)
}

// Adds the list of known types to api.Scheme.
func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&LoadBalancer{},
		&LoadBalancerList{},
	)
	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}
