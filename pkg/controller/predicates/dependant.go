package predicates

import (
	"reflect"

	"github.com/dapr/kubernetes-operator/pkg/resources"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

var _ predicate.Predicate = DependentPredicate{}

type DependentPredicateOption func(*DependentPredicate) *DependentPredicate

func WithWatchDeleted(val bool) DependentPredicateOption {
	return func(in *DependentPredicate) *DependentPredicate {
		in.WatchDelete = val
		return in
	}
}

func WithWatchUpdate(val bool) DependentPredicateOption {
	return func(in *DependentPredicate) *DependentPredicate {
		in.WatchUpdate = val
		return in
	}
}

func WithWatchStatus(val bool) DependentPredicateOption {
	return func(in *DependentPredicate) *DependentPredicate {
		in.WatchStatus = val
		return in
	}
}

type DependentPredicate struct {
	WatchDelete bool
	WatchUpdate bool
	WatchStatus bool

	predicate.Funcs
}

func (p DependentPredicate) Create(event.CreateEvent) bool {
	return false
}

func (p DependentPredicate) Generic(event.GenericEvent) bool {
	return false
}

func (p DependentPredicate) Delete(e event.DeleteEvent) bool {
	if !p.WatchDelete {
		return false
	}

	o, ok := e.Object.(*unstructured.Unstructured)
	if !ok {
		log.Error(nil, "unexpected object type", "gvks", e.Object.GetObjectKind().GroupVersionKind().String())
		return false
	}

	log.Info(
		"Reconciling due to dependent resource delete",
		"ref", resources.Ref(o),
		"partial", false)

	return true
}

func (p DependentPredicate) Update(e event.UpdateEvent) bool {
	if !p.WatchUpdate {
		return false
	}

	if e.ObjectOld.GetResourceVersion() == e.ObjectNew.GetResourceVersion() {
		return false
	}

	oldObj, ok := e.ObjectOld.(*unstructured.Unstructured)
	if !ok {
		log.Error(nil, "unexpected old object type", "gvks", e.ObjectOld.GetObjectKind().GroupVersionKind().String())
		return false
	}

	newObj, ok := e.ObjectNew.(*unstructured.Unstructured)
	if !ok {
		log.Error(nil, "unexpected new object type", "gvks", e.ObjectOld.GetObjectKind().GroupVersionKind().String())
		return false
	}

	oldObj = oldObj.DeepCopy()
	newObj = newObj.DeepCopy()

	if !p.WatchStatus {
		// Update filters out events that change only the dependent resource
		// status. It is not typical for the controller of a primary
		// resource to write to the status of one its dependent resources.
		delete(oldObj.Object, "status")
		delete(newObj.Object, "status")
	}

	// Reset field not meaningful for comparison
	oldObj.SetResourceVersion("")
	newObj.SetResourceVersion("")
	oldObj.SetManagedFields(nil)
	newObj.SetManagedFields(nil)

	if reflect.DeepEqual(oldObj.Object, newObj.Object) {
		return false
	}

	log.Info(
		"Reconciling due to dependent resource update",
		"ref", resources.Ref(newObj),
		"partial", false)

	return true
}

var _ predicate.Predicate = PartialDependentPredicate{}

type PartialDependentPredicateOption func(*PartialDependentPredicate) *PartialDependentPredicate

func PartialWatchDeleted(val bool) PartialDependentPredicateOption {
	return func(in *PartialDependentPredicate) *PartialDependentPredicate {
		in.WatchDelete = val
		return in
	}
}

func PartialWatchUpdate(val bool) PartialDependentPredicateOption {
	return func(in *PartialDependentPredicate) *PartialDependentPredicate {
		in.WatchUpdate = val
		return in
	}
}

type PartialDependentPredicate struct {
	WatchDelete bool
	WatchUpdate bool

	predicate.Funcs
}

func (p PartialDependentPredicate) Create(event.CreateEvent) bool {
	return false
}

func (p PartialDependentPredicate) Generic(event.GenericEvent) bool {
	return false
}

func (p PartialDependentPredicate) Delete(e event.DeleteEvent) bool {
	if !p.WatchDelete {
		return false
	}

	o, ok := e.Object.(*metav1.PartialObjectMetadata)
	if !ok {
		log.Error(nil, "unexpected object type", "gvks", e.Object.GetObjectKind().GroupVersionKind().String())
		return false
	}

	log.Info(
		"Reconciling due to dependent resource delete",
		"ref", resources.Ref(o),
		"partial", true)

	return true
}

func (p PartialDependentPredicate) Update(e event.UpdateEvent) bool {
	if !p.WatchUpdate {
		return false
	}

	if e.ObjectOld.GetResourceVersion() == e.ObjectNew.GetResourceVersion() {
		return false
	}

	newObj, ok := e.ObjectNew.(*metav1.PartialObjectMetadata)
	if !ok {
		log.Error(nil, "unexpected new object type", "gvks", e.ObjectNew.GetObjectKind().GroupVersionKind().String())
		return false
	}

	log.Info(
		"Reconciling due to dependent resource update",
		"ref", resources.Ref(newObj),
		"partial", true)

	return true
}
