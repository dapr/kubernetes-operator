package predicates

import (
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

var _ predicate.Predicate = AnnotationChanged{}

type AnnotationChanged struct {
	predicate.Funcs
	Name string
}

func (p AnnotationChanged) Create(event.CreateEvent) bool {
	return false
}

func (p AnnotationChanged) Generic(event.GenericEvent) bool {
	return false
}

func (p AnnotationChanged) Delete(event.DeleteEvent) bool {
	return false
}

// Update implements default UpdateEvent filter for validating annotation change.
func (p AnnotationChanged) Update(e event.UpdateEvent) bool {
	if e.ObjectOld == nil {
		log.Error(nil, "Update event has no old object to update", "event", e)
		return false
	}
	if e.ObjectOld.GetAnnotations() == nil {
		log.Error(nil, "Update event has no old object annotations to update", "event", e)
		return false
	}

	if e.ObjectNew == nil {
		log.Error(nil, "Update event has no new object for update", "event", e)
		return false
	}

	if e.ObjectNew.GetAnnotations() == nil {
		log.Error(nil, "Update event has no new object annotations for update", "event", e)
		return false
	}

	oldAnnotations := e.ObjectOld.GetAnnotations()
	newAnnotations := e.ObjectNew.GetAnnotations()

	return oldAnnotations[p.Name] != newAnnotations[p.Name]
}
