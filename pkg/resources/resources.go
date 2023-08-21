package resources

import (
	"fmt"

	"github.com/dapr-sandbox/dapr-kubernetes-operator/pkg/pointer"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func OwnerReference(owner client.Object) metav1.OwnerReference {
	return metav1.OwnerReference{
		APIVersion:         owner.GetObjectKind().GroupVersionKind().GroupVersion().String(),
		Kind:               owner.GetObjectKind().GroupVersionKind().Kind,
		Name:               owner.GetName(),
		UID:                owner.GetUID(),
		BlockOwnerDeletion: pointer.Any(true),
		Controller:         pointer.Any(true),
	}
}

func OwnerReferences(owner client.Object) []metav1.OwnerReference {
	return []metav1.OwnerReference{
		OwnerReference(owner),
	}
}

func Annotations(target *unstructured.Unstructured, annotations map[string]string) {
	m := target.GetAnnotations()

	if m == nil {
		m = make(map[string]string)
	}

	for k, v := range annotations {
		m[k] = v
	}

	target.SetAnnotations(m)
}

func Labels(target *unstructured.Unstructured, labels map[string]string) {
	m := target.GetLabels()

	if m == nil {
		m = make(map[string]string)
	}

	for k, v := range labels {
		m[k] = v
	}

	target.SetLabels(m)
}

func Ref(obj *unstructured.Unstructured) string {
	name := obj.GetName()
	if obj.GetNamespace() == "" {
		name = obj.GetNamespace() + ":" + obj.GetName()
	}

	return fmt.Sprintf(
		"%s:%s:%s",
		obj.GroupVersionKind().Kind,
		obj.GroupVersionKind().GroupVersion().String(),
		name,
	)
}
