package v1

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/InVisionApp/conjungo"
	libcorev1 "github.com/caicloud/loadbalancer-controller/pkg/kubelab/core/v1"
	libmetav1 "github.com/caicloud/loadbalancer-controller/pkg/kubelab/meta/v1"
	"github.com/mattbaird/jsonpatch"

	v1 "k8s.io/api/apps/v1"
	"k8s.io/klog"
	k8sappsv1 "k8s.io/kubernetes/pkg/apis/apps/v1"
)

// DeploymentLab contains some utils for Deployments
type DeploymentLab interface {
	// Merge merges the following fields from src to dst
	// - ObjectMeta.Labels
	// - ObjectMeta.Annotations
	// - Spec
	Merge(dst, src *v1.Deployment) error
	// IsEqual checks if the given two deployments are equal
	//
	// If ignoreFields is provided, the function will be call on each
	// deployment's deepcopy(be free to mutate it) before comparing to
	// ignore specified fields.
	// You can mutate object in the function like:
	// func (in *v1.Deployment) {
	//    in.Spec.Replicas = nil
	// }
	IsEqual(a, b *v1.Deployment, ignoreFields func(*v1.Deployment)) bool
}

func newOptions() *conjungo.Options {
	o := conjungo.NewOptions()
	o.SetKindMergeFunc(reflect.Slice, func(t, s reflect.Value, o *conjungo.Options) (reflect.Value, error) {
		// Merges two slices of the same type by appending source to target.
		if t.Type() != s.Type() {
			return reflect.Value{}, fmt.Errorf("slices must have same type: T: %v S: %v", t.Type(), s.Type())
		}
		if !t.CanSet() {
			return reflect.Value{}, fmt.Errorf("the target value can not be set")
		}
		if o.Overwrite {
			// overwrite it no matter what it is
			t.Set(s)
		} else if t.Len() == 0 && s.Len() > 0 {
			// without overwrite
			// only change target when it is empty but src is not
			t.Set(s)
		}
		return t, nil
	})

	return o
}

type deploymentImpl struct{}

func (l *deploymentImpl) Merge(dst, src *v1.Deployment) error {
	setObjectDefaultsDeployments(src)
	libmetav1.DefaultObjectMetaLab.Merge(&dst.ObjectMeta, &src.ObjectMeta)
	// merge spec
	err := conjungo.Merge(&dst.Spec, src.Spec, newOptions())
	if err != nil {
		return err
	}
	return nil
}

func (l *deploymentImpl) IsEqual(a, b *v1.Deployment, ignoreFields func(*v1.Deployment)) bool {
	acopy := a.DeepCopy()
	bcopy := b.DeepCopy()

	setObjectDefaultsDeployments(acopy)
	setObjectDefaultsDeployments(bcopy)

	if ignoreFields != nil {
		ignoreFields(acopy)
		ignoreFields(bcopy)
	}

	if !libmetav1.DefaultObjectMetaLab.IsEqual(&acopy.ObjectMeta, &bcopy.ObjectMeta) {
		klog.V(2).Infof("deployment %v/%v metadata changed", a.Namespace, a.Name)
		return false
	}

	aSpecBytes, _ := json.Marshal(acopy.Spec)
	bSpecBytes, _ := json.Marshal(bcopy.Spec)

	if !reflect.DeepEqual(aSpecBytes, bSpecBytes) {
		if klog.V(2) {
			if patch, err := jsonpatch.CreatePatch(aSpecBytes, bSpecBytes); err == nil {
				klog.Infof("deployment %v/%v spec changed, the patch is: %v", a.Namespace, a.Name, patch)
			}
		}
		klog.V(5).Infof("deployment %v/%v spec changed\na => %v\nb => %v", a.Namespace, a.Name, string(aSpecBytes), string(bSpecBytes))
		return false
	}
	return true
}

func setObjectDefaultsDeployments(in *v1.Deployment) {
	k8sappsv1.SetObjectDefaults_Deployment(in)
	libcorev1.DefaultPodLab.DropDisabledAlphaFields(&in.Spec.Template.Spec)
}
