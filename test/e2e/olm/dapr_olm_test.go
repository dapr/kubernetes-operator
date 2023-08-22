package operator

import (
	"os"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	olmV1 "github.com/operator-framework/api/pkg/operators/v1"
	olmV1Alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/rs/xid"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/dapr-sandbox/dapr-kubernetes-operator/test/support"
	. "github.com/onsi/gomega"
)

func TestDaprDeploy(t *testing.T) {
	test := With(t)

	ns := test.NewTestNamespace()
	id := xid.New().String()

	image := os.Getenv("CATALOG_CONTAINER_IMAGE")

	test.Expect(image).
		ToNot(BeEmpty())

	//
	// Install OperatorGroups
	//

	_, err := test.Client().OLM.OperatorsV1().OperatorGroups(ns.Name).Create(
		test.Ctx(),
		&olmV1.OperatorGroup{
			ObjectMeta: metav1.ObjectMeta{
				Name:      id,
				Namespace: ns.Name,
			},
		},
		metav1.CreateOptions{},
	)

	test.Expect(err).
		ToNot(HaveOccurred())
	test.Expect(image).
		ToNot(BeEmpty())

	//
	// Install CatalogSource
	//

	catalog, err := test.Client().OLM.OperatorsV1alpha1().CatalogSources(ns.Name).Create(
		test.Ctx(),
		&olmV1Alpha1.CatalogSource{
			ObjectMeta: metav1.ObjectMeta{
				Name:      id,
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
							Duration: 10 * time.Minute,
						},
					},
				},
			},
		},
		metav1.CreateOptions{},
	)

	test.Expect(err).
		ToNot(HaveOccurred())
	test.Expect(catalog).
		ToNot(BeNil())

	test.Eventually(CatalogSource(test, catalog.Name, catalog.Namespace), TestTimeoutLong).Should(
		WithTransform(ExtractCatalogState(), Equal("READY")),
	)

	//
	// Install Subscription
	//

	sub, err := test.Client().OLM.OperatorsV1alpha1().Subscriptions(ns.Name).Create(
		test.Ctx(),
		&olmV1Alpha1.Subscription{
			ObjectMeta: metav1.ObjectMeta{
				Name:      id,
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
		ToNot(HaveOccurred())
	test.Expect(sub).
		ToNot(BeNil())

	test.Eventually(Subscription(test, sub.Name, sub.Namespace), TestTimeoutLong).Should(
		WithTransform(ExtractSubscriptionInstallPlan(), Not(BeEmpty())),
	)

	//
	// Control plane
	//

	test.Eventually(Deployment(test, "dapr-control-plane", sub.Namespace), TestTimeoutLong).Should(
		WithTransform(ConditionStatus(appsv1.DeploymentAvailable), Equal(corev1.ConditionTrue)))

}
