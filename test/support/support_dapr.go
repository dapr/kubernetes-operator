package support

import (
	daprApi "github.com/dapr-sandbox/dapr-kubernetes-operator/api/operator/v1alpha1"
	"github.com/onsi/gomega"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ControlPlane(t Test, dapr *daprApi.DaprControlPlane) func(g gomega.Gomega) (*daprApi.DaprControlPlane, error) {
	return func(g gomega.Gomega) (*daprApi.DaprControlPlane, error) {
		answer, err := t.Client().Dapr.OperatorV1alpha1().DaprControlPlanes(dapr.Namespace).Get(
			t.Ctx(),
			dapr.Name,
			metav1.GetOptions{},
		)

		if k8serrors.IsNotFound(err) {
			return nil, nil
		}

		return answer, err
	}
}
