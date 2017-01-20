package controller

const (
	// these pair of constants are used by the provisioner.
	// The key is a kube namespaced key that denotes a ingress service requires provisioning.
	// The value is set only when provisioning is completed.  Any other value will tell the provisioner
	// that provisioning has not yet occurred.
	ingressProvisioningRequiredAnnotationKey    = "ingress.alpha.k8s.io/provisioning-required"
	ingressProvisioningCompletedAnnotationValue = "ingress.alpha.k8s.io/provisioning-completed"
	ingressProvisioningFailedAnnotationValue 	= "ingress.alpha.k8s.io/provisioning-failed"

	IngressProvisioningClassKey   = "ingress.alpha.k8s.io/ingress-class"

	ingressParameterCPUKey = "ingress.alpha.k8s.io/ingress-cpu"
	ingressParameterMEMKey = "ingress.alpha.k8s.io/ingress-mem"
	IngressParameterVIPKey = "ingress.alpha.k8s.io/ingress-vip"
)

