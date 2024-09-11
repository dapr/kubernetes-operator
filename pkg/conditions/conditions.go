package conditions

import (
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
		if c, ok := FindDeploymentStatusCondition(o, string(conditionType)); ok {
			return c.Status
		}
	case appsv1.Deployment:
		if c, ok := FindDeploymentStatusCondition(&o, string(conditionType)); ok {
			return c.Status
		}
	case *corev1.Pod:
		if c, ok := FindPodStatusCondition(o, string(conditionType)); ok {
			return c.Status
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
		if c, ok := FindDeploymentStatusCondition(o, string(conditionType)); ok {
			return c.Reason
		}
	case appsv1.Deployment:
		if c, ok := FindDeploymentStatusCondition(&o, string(conditionType)); ok {
			return c.Reason
		}
	case *corev1.Pod:
		if c, ok := FindPodStatusCondition(o, string(conditionType)); ok {
			return c.Reason
		}
	}

	return ""
}

func FindDeploymentStatusCondition(in *appsv1.Deployment, conditionType string) (appsv1.DeploymentCondition, bool) {
	if in == nil {
		return appsv1.DeploymentCondition{}, false
	}

	for i := range in.Status.Conditions {
		if string(in.Status.Conditions[i].Type) == conditionType {
			return in.Status.Conditions[i], true
		}
	}

	return appsv1.DeploymentCondition{}, false
}

func FindPodStatusCondition(in *corev1.Pod, conditionType string) (corev1.PodCondition, bool) {
	if in == nil {
		return corev1.PodCondition{}, false
	}

	for i := range in.Status.Conditions {
		if string(in.Status.Conditions[i].Type) == conditionType {
			return in.Status.Conditions[i], true
		}
	}

	return corev1.PodCondition{}, false
}
