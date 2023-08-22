package support

import (
	"github.com/onsi/gomega"
	olmV1Alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func CatalogSource(t Test, name string, namespace string) func(g gomega.Gomega) (*olmV1Alpha1.CatalogSource, error) {
	return func(g gomega.Gomega) (*olmV1Alpha1.CatalogSource, error) {
		answer, err := t.Client().OLM.OperatorsV1alpha1().CatalogSources(namespace).Get(
			t.Ctx(),
			name,
			metav1.GetOptions{},
		)

		if k8serrors.IsNotFound(err) {
			return nil, nil
		}

		return answer, err
	}
}

func ExtractCatalogState() func(*olmV1Alpha1.CatalogSource) string {
	return func(in *olmV1Alpha1.CatalogSource) string {
		if in.Status.GRPCConnectionState == nil {
			return ""
		}

		return in.Status.GRPCConnectionState.LastObservedState
	}
}

func Subscription(t Test, name string, namespace string) func(g gomega.Gomega) (*olmV1Alpha1.Subscription, error) {
	return func(g gomega.Gomega) (*olmV1Alpha1.Subscription, error) {
		answer, err := t.Client().OLM.OperatorsV1alpha1().Subscriptions(namespace).Get(
			t.Ctx(),
			name,
			metav1.GetOptions{},
		)

		if k8serrors.IsNotFound(err) {
			return nil, nil
		}

		return answer, err
	}
}

func ExtractSubscriptionInstallPlan() func(*olmV1Alpha1.Subscription) string {
	return func(in *olmV1Alpha1.Subscription) string {
		if in.Status.InstallPlanRef == nil {
			return ""
		}

		return in.Status.InstallPlanRef.Name
	}
}
