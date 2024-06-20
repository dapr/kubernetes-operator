package operator

import (
	"os"
	"testing"

	"github.com/dapr/kubernetes-operator/internal/controller/operator/controlplane"
	"github.com/dapr/kubernetes-operator/internal/controller/operator/instance"
	daprAc "github.com/dapr/kubernetes-operator/pkg/client/applyconfiguration/operator/v1alpha1"
	"github.com/dapr/kubernetes-operator/test/support/dapr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/dapr/kubernetes-operator/test/support/olm"

	daprTC "github.com/dapr/kubernetes-operator/test/e2e/common"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	. "github.com/dapr/kubernetes-operator/test/support"
	. "github.com/onsi/gomega"
)

func TestDaprDeploy(t *testing.T) {
	t.Run("With ControlPlane", func(t *testing.T) {
		testDaprDeploy(
			With(t),
			func(test Test, ns *corev1.Namespace) client.Object {
				return dapr.DeployControlPlane(
					test,
					daprAc.DaprControlPlaneSpec().
						WithValues(dapr.Values(test, map[string]interface{}{
							// enable pod watchdog as sometimes the sidecar for some
							// (yet) unknown reason is not injected when the pod is
							// created, hence the dapr app won't properly start up
							"dapr_operator": map[string]interface{}{
								"watchInterval": "1s",
							},
						})),
					dapr.WithControlPlaneName(controlplane.DaprControlPlaneResourceName),
					dapr.WithControlPlaneNamespace(ns.Name),
				)
			},
		)
	})

	t.Run("With Instance", func(t *testing.T) {
		testDaprDeploy(
			With(t),
			func(test Test, ns *corev1.Namespace) client.Object {
				return dapr.DeployInstance(
					test,
					daprAc.DaprInstanceSpec().
						WithValues(dapr.Values(test, map[string]interface{}{
							// enable pod watchdog as sometimes the sidecar for some
							// (yet) unknown reason is not injected when the pod is
							// created, hence the dapr app won't properly start up
							"dapr_operator": map[string]interface{}{
								"watchInterval": "1s",
							},
						})),
					dapr.WithInstanceName(instance.DaprInstanceResourceName),
					dapr.WithInstanceNamespace(ns.Name),
				)
			},
		)
	})
}

func testDaprDeploy(test Test, f func(t Test, ns *corev1.Namespace) client.Object) {
	test.T().Helper()

	ns := test.NewTestNamespace()
	image := os.Getenv("CATALOG_CONTAINER_IMAGE")

	test.Expect(image).
		ToNot(BeEmpty())

	//
	// Install Operator
	//

	olm.DeployOperator(test, ns, image)

	//
	// Control plane
	//

	test.Eventually(Deployment(test, "dapr-control-plane", ns.Name), TestTimeoutLong).Should(
		WithTransform(ConditionStatus(appsv1.DeploymentAvailable), Equal(corev1.ConditionTrue)))

	//
	// Dapr
	//

	res := f(test, ns)

	test.Eventually(Deployment(test, "dapr-operator", res.GetNamespace()), TestTimeoutLong).Should(
		WithTransform(ConditionStatus(appsv1.DeploymentAvailable), Equal(corev1.ConditionTrue)))
	test.Eventually(Deployment(test, "dapr-sentry", res.GetNamespace()), TestTimeoutLong).Should(
		WithTransform(ConditionStatus(appsv1.DeploymentAvailable), Equal(corev1.ConditionTrue)))
	test.Eventually(Deployment(test, "dapr-sidecar-injector", res.GetNamespace()), TestTimeoutLong).Should(
		WithTransform(ConditionStatus(appsv1.DeploymentAvailable), Equal(corev1.ConditionTrue)))

	//
	// Dapr Application
	//

	daprTC.ValidateDaprApp(test, res.GetNamespace())
}
