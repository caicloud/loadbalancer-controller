package lb

import (
	"github.com/caicloud/loadbalancer-controller/pkg/kubelab"
	appsv1 "k8s.io/api/apps/v1"
)

// MergeDeployment merges fields in src into dst, and igornes some fields
// by default for debugging
func MergeDeployment(dst, src *appsv1.Deployment) (*appsv1.Deployment, bool) {
	dstcopy := dst.DeepCopy()
	helper := kubelab.New().Apps().V1().Deployments()

	_ = helper.Merge(dstcopy, src)

	equal := helper.IsEqual(dst, dstcopy, func(in *appsv1.Deployment) {
		// ignore some fields for debug
		for i := range in.Spec.Template.Spec.Containers {
			in.Spec.Template.Spec.Containers[i].Args = nil
			in.Spec.Template.Spec.Containers[i].LivenessProbe = nil
			in.Spec.Template.Spec.Containers[i].ReadinessProbe = nil
		}
	})
	return dstcopy, !equal
}

func completeHelmAnnotation(ann map[string]string, namespace, name string) map[string]string {
	if ann == nil {
		ann = make(map[string]string)
	}
	ann["helm.sh/namespace"] = namespace
	ann["helm.sh/release"] = name
	ann["helm.sh/path"] = name // The key "helm.sh/path" is needed, the value is meaningless
	return ann
}

// InsertHelmAnnotation inserts helm.sh field into annotation
func InsertHelmAnnotation(dp *appsv1.Deployment, namespace, name string) {
	dp.Annotations = completeHelmAnnotation(dp.Annotations, namespace, name)
	dp.Spec.Template.Annotations = completeHelmAnnotation(dp.Spec.Template.Annotations, namespace, name)
}
