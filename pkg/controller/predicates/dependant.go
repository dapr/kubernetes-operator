package predicates

import (
	"encoding/json"
	"reflect"

	"github.com/wI2L/jsondiff"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

var _ predicate.Predicate = DependentPredicate{}

type DependentPredicate struct {
	WatchDelete bool
	WatchUpdate bool

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
		log.Error(nil, "unexpected object type", "gvk", e.Object.GetObjectKind().GroupVersionKind().String())
		return false
	}

	log.Info("Reconciling due to dependent resource deletion",
		"name", o.GetName(),
		"namespace", o.GetNamespace(),
		"apiVersion", o.GroupVersionKind().GroupVersion(),
		"kind", o.GroupVersionKind().Kind)

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
		log.Error(nil, "unexpected old object type", "gvk", e.ObjectOld.GetObjectKind().GroupVersionKind().String())
		return false
	}

	newObj, ok := e.ObjectNew.(*unstructured.Unstructured)
	if !ok {
		log.Error(nil, "unexpected new object type", "gvk", e.ObjectOld.GetObjectKind().GroupVersionKind().String())
		return false
	}

	oldObj = oldObj.DeepCopy()
	newObj = newObj.DeepCopy()

	// Update filters out events that change only the dependent resource
	// status. It is not typical for the controller of a primary
	// resource to write to the status of one its dependent resources.
	delete(oldObj.Object, "status")
	delete(newObj.Object, "status")

	// Reset field not meaningful for comparison
	oldObj.SetResourceVersion("")
	newObj.SetResourceVersion("")
	oldObj.SetManagedFields(nil)
	newObj.SetManagedFields(nil)

	if reflect.DeepEqual(oldObj.Object, newObj.Object) {
		return false
	}

	patch, err := jsondiff.Compare(oldObj, newObj)
	if err != nil {
		log.Error(err, "failed to generate diff")
		return true
	}
	d, err := json.Marshal(patch)
	if err != nil {
		log.Error(err, "failed to generate diff")
		return true
	}

	log.Info("Reconciling due to dependent resource update",
		"name", newObj.GetName(),
		"namespace", newObj.GetNamespace(),
		"apiVersion", newObj.GroupVersionKind().GroupVersion(),
		"kind", newObj.GroupVersionKind().Kind,
		"diff", string(d))

	return true
}
