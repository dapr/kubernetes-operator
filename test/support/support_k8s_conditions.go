package support

import (
	"github.com/dapr/kubernetes-operator/pkg/conditions"
	corev1 "k8s.io/api/core/v1"
)

func ConditionStatus[T conditions.GenericConditionType](conditionType T) func(any) corev1.ConditionStatus {
	return func(object any) corev1.ConditionStatus {
		return conditions.ConditionStatus(object, conditionType)
	}
}

func ConditionReason[T conditions.GenericConditionType](conditionType T) func(any) string {
	return func(object any) string {
		return conditions.ConditionReason(object, conditionType)
	}
}
