package operator

import (
	"os"
	"testing"

	"github.com/dapr/kubernetes-operator/test/support/dapr"
	"github.com/dapr/kubernetes-operator/test/support/olm"

	daprAc "github.com/dapr/kubernetes-operator/pkg/client/applyconfiguration/operator/v1beta1"
	daprTC "github.com/dapr/kubernetes-operator/test/e2e/common"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	. "github.com/dapr/kubernetes-operator/test/support"
	. "github.com/onsi/gomega"
)

func TestDaprDeploy(t *testing.T) {
	test := With(t)

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

	_ = dapr.DeployInstance(
		test,
		daprAc.DaprInstanceSpec().
			WithDeployment(
				daprAc.DeploymentSpec().
					WithNamespace(ns.Name),
			).
			WithValues(dapr.Values(test, map[string]interface{}{
				// enable pod watchdog as sometimes the sidecar for some
				// (yet) unknown reason is not injected when the pod is
				// created, hence the dapr app won't properly start up
				"dapr_operator": map[string]interface{}{
					"watchInterval": "1s",
				},
			})),
	)

	test.Eventually(Deployment(test, "dapr-operator", ns.Name), TestTimeoutLong).Should(
		WithTransform(ConditionStatus(appsv1.DeploymentAvailable), Equal(corev1.ConditionTrue)))
	test.Eventually(Deployment(test, "dapr-sentry", ns.Name), TestTimeoutLong).Should(
		WithTransform(ConditionStatus(appsv1.DeploymentAvailable), Equal(corev1.ConditionTrue)))
	test.Eventually(Deployment(test, "dapr-sidecar-injector", ns.Name), TestTimeoutLong).Should(
		WithTransform(ConditionStatus(appsv1.DeploymentAvailable), Equal(corev1.ConditionTrue)))

	//
	// Dapr Application
	//

	daprTC.ValidateDaprApp(test, ns.Name)
}
