package operator

import (
	"testing"

	"github.com/dapr-sandbox/dapr-kubernetes-operator/pkg/conditions"
	"github.com/dapr-sandbox/dapr-kubernetes-operator/pkg/controller"
	"github.com/dapr-sandbox/dapr-kubernetes-operator/test/support/dapr"

	"github.com/rs/xid"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	. "github.com/dapr-sandbox/dapr-kubernetes-operator/test/support"
	. "github.com/onsi/gomega"

	daprAc "github.com/dapr-sandbox/dapr-kubernetes-operator/pkg/client/applyconfiguration/operator/v1alpha1"
	daprTC "github.com/dapr-sandbox/dapr-kubernetes-operator/test/e2e/common"
)

func TestDaprControlPlaneDeployWithApp(t *testing.T) {
	test := With(t)

	instance := dapr.DeployControlPlane(
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
	)

	test.Eventually(CustomResourceDefinition(test, "components.dapr.io"), TestTimeoutLong).Should(Not(BeNil()))
	test.Eventually(CustomResourceDefinition(test, "configurations.dapr.io"), TestTimeoutLong).Should(Not(BeNil()))
	test.Eventually(CustomResourceDefinition(test, "httpendpoints.dapr.io"), TestTimeoutLong).Should(Not(BeNil()))
	test.Eventually(CustomResourceDefinition(test, "resiliencies.dapr.io"), TestTimeoutLong).Should(Not(BeNil()))
	test.Eventually(CustomResourceDefinition(test, "subscriptions.dapr.io"), TestTimeoutLong).Should(Not(BeNil()))

	test.Eventually(Deployment(test, "dapr-operator", instance.Namespace), TestTimeoutLong).Should(
		WithTransform(ConditionStatus(appsv1.DeploymentAvailable), Equal(corev1.ConditionTrue)))
	test.Eventually(Deployment(test, "dapr-sentry", instance.Namespace), TestTimeoutLong).Should(
		WithTransform(ConditionStatus(appsv1.DeploymentAvailable), Equal(corev1.ConditionTrue)))
	test.Eventually(Deployment(test, "dapr-sidecar-injector", instance.Namespace), TestTimeoutLong).Should(
		WithTransform(ConditionStatus(appsv1.DeploymentAvailable), Equal(corev1.ConditionTrue)))

	//
	// Dapr Application
	//

	daprTC.ValidateDaprApp(test, instance.Namespace)
}

func TestDaprControlPlaneDeployWrongCR(t *testing.T) {
	test := With(t)

	instance := dapr.DeployControlPlane(
		test,
		daprAc.DaprControlPlaneSpec().
			WithValues(nil),
		dapr.WithControlPlaneName(xid.New().String()),
		dapr.WithControlPlaneNamespace(controller.NamespaceDefault),
	)

	test.Eventually(dapr.ControlPlane(test, instance), TestTimeoutLong).Should(
		WithTransform(ConditionStatus(conditions.TypeReconciled), Equal(corev1.ConditionFalse)))
	test.Eventually(dapr.ControlPlane(test, instance), TestTimeoutLong).Should(
		WithTransform(ConditionReason(conditions.TypeReconciled), Equal(conditions.ReasonUnsupportedConfiguration)))
}
