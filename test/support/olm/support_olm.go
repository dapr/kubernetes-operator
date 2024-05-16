package olm

import (
	"fmt"
	"time"

	"github.com/dapr-sandbox/dapr-kubernetes-operator/pkg/pointer"

	"k8s.io/apimachinery/pkg/runtime"

	"github.com/dapr-sandbox/dapr-kubernetes-operator/test/support"
	"github.com/onsi/gomega"
	olmV1 "github.com/operator-framework/api/pkg/operators/v1"
	olmV1Alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	CatalogRegistryPollInterval = 10 * time.Minute
)

func DeployOperator(test support.Test, ns *corev1.Namespace, image string) {
	//
	// Install OperatorGroups
	//
	og, err := test.Client().OLM().OperatorsV1().OperatorGroups(ns.Name).Create(
		test.Ctx(),
		&olmV1.OperatorGroup{
			ObjectMeta: metav1.ObjectMeta{
				Name:      test.ID(),
				Namespace: ns.Name,
			},
		},
		metav1.CreateOptions{},
	)

	test.Expect(err).
		ToNot(gomega.HaveOccurred())
	test.Expect(image).
		ToNot(gomega.BeEmpty())

	test.Cleanup(func() []runtime.Object {
		//
		// For some reason, panic happen if the OperatorGroup is translated to
		// an unstructured object, hence the need of a dedicated cleanup logic
		//
		// see:
		//   https://github.com/operator-framework/api/issues/269
		//
		err := test.Client().OLM().OperatorsV1().OperatorGroups(ns.Name).Delete(
			test.Ctx(),
			og.Name,
			metav1.DeleteOptions{
				PropagationPolicy: pointer.Any(metav1.DeletePropagationForeground),
			},
		)

		test.Expect(err).
			ToNot(gomega.HaveOccurred())

		test.Eventually(OperatorGroup(test, og.Name, og.Namespace), support.TestTimeoutShort).
			Should(
				gomega.BeNil(),
				fmt.Sprintf("Failure deleting OperatorGroups with name %s in namespace %s",
					og.GetName(),
					og.GetNamespace(),
				),
			)

		return []runtime.Object{}
	})

	//
	// Install CatalogSource
	//

	catalog, err := test.Client().OLM().OperatorsV1alpha1().CatalogSources(ns.Name).Create(
		test.Ctx(),
		&olmV1Alpha1.CatalogSource{
			ObjectMeta: metav1.ObjectMeta{
				Name:      test.ID(),
				Namespace: ns.Name,
			},
			Spec: olmV1Alpha1.CatalogSourceSpec{
				SourceType:  "grpc",
				Image:       image,
				DisplayName: "Dapr.io Catalog",
				Publisher:   "dapr.io",
				GrpcPodConfig: &olmV1Alpha1.GrpcPodConfig{
					SecurityContextConfig: "restricted",
				},
				UpdateStrategy: &olmV1Alpha1.UpdateStrategy{
					RegistryPoll: &olmV1Alpha1.RegistryPoll{
						Interval: &metav1.Duration{
							Duration: CatalogRegistryPollInterval,
						},
					},
				},
			},
		},
		metav1.CreateOptions{},
	)

	test.Expect(err).
		ToNot(gomega.HaveOccurred())
	test.Expect(catalog).
		ToNot(gomega.BeNil())

	test.Cleanup(func() []runtime.Object {
		return []runtime.Object{catalog}
	})

	test.Eventually(CatalogSource(test, catalog.Name, catalog.Namespace), support.TestTimeoutLong).Should(
		gomega.WithTransform(ExtractCatalogState(), gomega.Equal("READY")),
	)

	//
	// Install Subscription
	//

	sub, err := test.Client().OLM().OperatorsV1alpha1().Subscriptions(ns.Name).Create(
		test.Ctx(),
		&olmV1Alpha1.Subscription{
			ObjectMeta: metav1.ObjectMeta{
				Name:      test.ID(),
				Namespace: ns.Name,
			},
			Spec: &olmV1Alpha1.SubscriptionSpec{
				Channel:                "alpha",
				Package:                "dapr-kubernetes-operator",
				InstallPlanApproval:    olmV1Alpha1.ApprovalAutomatic,
				CatalogSource:          catalog.Name,
				CatalogSourceNamespace: catalog.Namespace,
			},
		},
		metav1.CreateOptions{},
	)

	test.Expect(err).
		ToNot(gomega.HaveOccurred())
	test.Expect(sub).
		ToNot(gomega.BeNil())

	test.Cleanup(func() []runtime.Object {
		return []runtime.Object{sub}
	})

	test.Eventually(Subscription(test, sub.Name, sub.Namespace), support.TestTimeoutLong).Should(
		gomega.WithTransform(ExtractSubscriptionInstallPlan(), gomega.Not(gomega.BeEmpty())),
	)
}

func CatalogSource(t support.Test, name string, namespace string) func(g gomega.Gomega) (*olmV1Alpha1.CatalogSource, error) {
	return func(g gomega.Gomega) (*olmV1Alpha1.CatalogSource, error) {
		answer, err := t.Client().OLM().OperatorsV1alpha1().CatalogSources(namespace).Get(
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

func OperatorGroup(t support.Test, name string, namespace string) func(g gomega.Gomega) (*olmV1.OperatorGroup, error) {
	return func(g gomega.Gomega) (*olmV1.OperatorGroup, error) {
		answer, err := t.Client().OLM().OperatorsV1().OperatorGroups(namespace).Get(
			t.Ctx(),
			name,
			metav1.GetOptions{},
		)

		if err != nil && !k8serrors.IsNotFound(err) {
			return nil, err
		}

		if err != nil && k8serrors.IsNotFound(err) {
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

func Subscription(t support.Test, name string, namespace string) func(g gomega.Gomega) (*olmV1Alpha1.Subscription, error) {
	return func(g gomega.Gomega) (*olmV1Alpha1.Subscription, error) {
		answer, err := t.Client().OLM().OperatorsV1alpha1().Subscriptions(namespace).Get(
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
