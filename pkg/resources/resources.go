package resources

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"

	"k8s.io/apimachinery/pkg/runtime/schema"

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

func UnstructuredFor(group string, version string, kind string) *unstructured.Unstructured {
	u := unstructured.Unstructured{}
	u.SetGroupVersionKind(schema.GroupVersionKind{
		Kind:    kind,
		Group:   group,
		Version: version,
	})

	return &u
}

func ToUnstructured(s *runtime.Scheme, obj runtime.Object) (*unstructured.Unstructured, error) {
	switch ot := obj.(type) {
	case *unstructured.Unstructured:
		return ot, nil
	default:
		var err error
		var u unstructured.Unstructured

		u.Object, err = runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
		if err != nil {
			return nil, fmt.Errorf("failed to convert to unstructured: %w", err)
		}

		gvk := u.GroupVersionKind()
		if gvk.Group == "" || gvk.Kind == "" {
			gvks, _, err := s.ObjectKinds(obj)
			if err != nil {
				return nil, fmt.Errorf("failed to convert to unstructured - unable to get GVK %w", err)
			}
			apiv, k := gvks[0].ToAPIVersionAndKind()

			u.SetAPIVersion(apiv)
			u.SetKind(k)
		}

		return &u, nil
	}
}
