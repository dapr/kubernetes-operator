package predicates

import (
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

var _ predicate.Predicate = HasLabel{}

type HasLabel struct {
	predicate.Funcs
	Name string
}

func (p HasLabel) Create(event.CreateEvent) bool {
	return false
}

func (p HasLabel) Generic(event.GenericEvent) bool {
	return false
}

func (p HasLabel) Delete(e event.DeleteEvent) bool {
	return p.test(e.Object)
}

func (p HasLabel) Update(e event.UpdateEvent) bool {
	return p.test(e.ObjectNew)
}

func (p HasLabel) test(obj client.Object) bool {
	if obj == nil {
		return false
	}
	if obj.GetLabels() == nil {
		return false
	}

	_, ok := obj.GetLabels()[p.Name]

	return ok
}
