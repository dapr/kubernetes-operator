package conditions

import (
	daprApi "github.com/dapr-sandbox/dapr-kubernetes-operator/api/operator/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// ConditionType is a valid value for Condition.Type.
type ConditionType string

// Getter interface defines methods that an object should implement in order to
// use the conditions package for getting conditions.
type Getter interface {
	metav1.Object
	runtime.Object

	// GetConditions returns the list of conditions for an object.
	GetConditions() Conditions
}

type Conditions []metav1.Condition

// Get returns the condition with the given type, if the condition does not exist,
// it returns nil.
func Get(from Getter, t ConditionType) *metav1.Condition {
	conditions := from.GetConditions()
	if conditions == nil {
		return nil
	}

	for _, condition := range conditions {
		if condition.Type == string(t) {
			return &condition
		}
	}
	return nil
}

type GenericConditionType interface {
	~string
}

func ConditionStatus[T GenericConditionType](object any, conditionType T) corev1.ConditionStatus {
	switch o := object.(type) {
	case Getter:
		if c := Get(o, ConditionType(conditionType)); c != nil {
			return corev1.ConditionStatus(c.Status)
		}
	case *appsv1.Deployment:
		if o != nil {
			for i := range o.Status.Conditions {
				if string(o.Status.Conditions[i].Type) == string(conditionType) {
					return o.Status.Conditions[i].Status
				}
			}
		}
	case appsv1.Deployment:
		for i := range o.Status.Conditions {
			if string(o.Status.Conditions[i].Type) == string(conditionType) {
				return o.Status.Conditions[i].Status
			}
		}
	case *corev1.Pod:
		if o != nil {
			for i := range o.Status.Conditions {
				if string(o.Status.Conditions[i].Type) == string(conditionType) {
					return o.Status.Conditions[i].Status
				}
			}
		}
	case *daprApi.DaprControlPlane:
		if o != nil {
			for i := range o.Status.Conditions {
				if o.Status.Conditions[i].Type == string(conditionType) {
					return corev1.ConditionStatus(o.Status.Conditions[i].Status)
				}
			}
		}
	}

	return corev1.ConditionUnknown
}

func ConditionReason[T GenericConditionType](object any, conditionType T) string {
	switch o := object.(type) {
	case Getter:
		if c := Get(o, ConditionType(conditionType)); c != nil {
			return c.Reason
		}
	case *appsv1.Deployment:
		if o != nil {
			for i := range o.Status.Conditions {
				if string(o.Status.Conditions[i].Type) == string(conditionType) {
					return o.Status.Conditions[i].Reason
				}
			}
		}
	case appsv1.Deployment:
		for i := range o.Status.Conditions {
			if string(o.Status.Conditions[i].Type) == string(conditionType) {
				return o.Status.Conditions[i].Reason
			}
		}
	case *corev1.Pod:
		if o != nil {
			for i := range o.Status.Conditions {
				if string(o.Status.Conditions[i].Type) == string(conditionType) {
					return o.Status.Conditions[i].Reason
				}
			}
		}
	case *daprApi.DaprControlPlane:
		if o != nil {
			for i := range o.Status.Conditions {
				if o.Status.Conditions[i].Type == string(conditionType) {
					return o.Status.Conditions[i].Reason
				}
			}
		}
	}

	return ""
}
