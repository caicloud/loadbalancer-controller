package v1

import (
	"reflect"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
)

var (
	DefaultObjectMetaLab = &objectMetaImpl{}
)

// ObjectMetaLab contains some utils for ObjectMeta
type ObjectMetaLab interface {
	// Merge only merges Labels and Annotations
	Merge(dst, src *metav1.ObjectMeta)
	// IsEqual only checks Labels and Annotations
	IsEqual(a, b *metav1.ObjectMeta) bool
}

type objectMetaImpl struct{}

func (l *objectMetaImpl) Merge(dst, src *metav1.ObjectMeta) {
	if len(src.Labels) > 0 && dst.Labels == nil {
		dst.Labels = make(map[string]string)
	}
	for k, v := range src.Labels {
		dst.Labels[k] = v
	}
	if len(src.Annotations) > 0 && dst.Labels == nil {
		dst.Annotations = make(map[string]string)
	}
	for k, v := range src.Annotations {
		dst.Annotations[k] = v
	}
	dst.OwnerReferences = src.OwnerReferences
}

func (l *objectMetaImpl) IsEqual(a, b *metav1.ObjectMeta) bool {
	if !reflect.DeepEqual(a.Labels, b.Labels) {
		klog.V(5).Infof("%v/%v labels changed, a.Labels: %v, b.Labels: %v", a.Namespace, b.Name, a.Labels, b.Labels)
		return false
	}
	if !reflect.DeepEqual(a.Annotations, b.Annotations) {
		klog.V(5).Infof("%v/%v Annotations changed, a.Annotations: %v, b.Annotations: %v", a.Namespace, b.Name, a.Annotations, b.Annotations)
		return false
	}
	return true
}
