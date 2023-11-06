package operator

import (
	"os"
	"testing"

	"github.com/dapr-sandbox/dapr-kubernetes-operator/internal/controller/operator/instance"
	"github.com/dapr-sandbox/dapr-kubernetes-operator/test/support/dapr"
	"github.com/dapr-sandbox/dapr-kubernetes-operator/test/support/olm"

	daprAc "github.com/dapr-sandbox/dapr-kubernetes-operator/pkg/client/operator/applyconfiguration/operator/v1alpha1"
	daprTC "github.com/dapr-sandbox/dapr-kubernetes-operator/test/e2e/common"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	. "github.com/dapr-sandbox/dapr-kubernetes-operator/test/support"
	. "github.com/onsi/gomega"
)

func TestDaprDeployWithInstance(t *testing.T) {
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

	res := dapr.DeployInstance(
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

	test.Eventually(Deployment(test, "dapr-operator", res.Namespace), TestTimeoutLong).Should(
		WithTransform(ConditionStatus(appsv1.DeploymentAvailable), Equal(corev1.ConditionTrue)))
	test.Eventually(Deployment(test, "dapr-sentry", res.Namespace), TestTimeoutLong).Should(
		WithTransform(ConditionStatus(appsv1.DeploymentAvailable), Equal(corev1.ConditionTrue)))
	test.Eventually(Deployment(test, "dapr-sidecar-injector", res.Namespace), TestTimeoutLong).Should(
		WithTransform(ConditionStatus(appsv1.DeploymentAvailable), Equal(corev1.ConditionTrue)))

	//
	// Dapr Application
	//

	daprTC.ValidateDaprApp(test, res.Namespace)

}
