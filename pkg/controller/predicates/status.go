package predicates

import (
	"reflect"

	"k8s.io/apimachinery/pkg/api/equality"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

var _ predicate.Predicate = StatusChanged{}

// StatusChanged implements a generic update predicate function on status change.
type StatusChanged struct {
	predicate.Funcs
}

func (p StatusChanged) Create(event.CreateEvent) bool {
	return false
}

func (p StatusChanged) Generic(event.GenericEvent) bool {
	return false
}

func (p StatusChanged) Delete(event.DeleteEvent) bool {
	return false
}

// Update implements default UpdateEvent filter for validating status change.
func (p StatusChanged) Update(e event.UpdateEvent) bool {
	if e.ObjectOld == nil {
		log.Error(nil, "Update event has no old object to update", "event", e)
		return false
	}
	if e.ObjectNew == nil {
		log.Error(nil, "Update event has no new object to update", "event", e)
		return false
	}

	s1 := reflect.ValueOf(e.ObjectOld).Elem().FieldByName("Status")
	if !s1.IsValid() {
		log.Error(nil, "Update event old object has no Status field", "event", e)
		return false
	}

	s2 := reflect.ValueOf(e.ObjectNew).Elem().FieldByName("Status")
	if !s2.IsValid() {
		log.Error(nil, "Update event new object has no Status field", "event", e)
		return false
	}

	return !equality.Semantic.DeepEqual(s1.Interface(), s2.Interface())
}
