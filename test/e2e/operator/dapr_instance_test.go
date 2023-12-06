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
	. "github.com/dapr-sandbox/dapr-kubernetes-operator/test/support/matchers"
	. "github.com/onsi/gomega"

	daprAc "github.com/dapr-sandbox/dapr-kubernetes-operator/pkg/client/operator/applyconfiguration/operator/v1alpha1"
	daprTC "github.com/dapr-sandbox/dapr-kubernetes-operator/test/e2e/common"
)

func TestDaprInstanceDeployWithDefaults(t *testing.T) {
	test := With(t)

	instance := dapr.DeployInstance(
		test,
		daprAc.DaprInstanceSpec().
			WithValues(nil),
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

	test.Eventually(dapr.Instance(test, instance), TestTimeoutLong).Should(
		WithTransform(AsJSON(), And(
			MatchJQ(`.status.chart.name == "dapr"`),
			MatchJQ(`.status.chart.repo == "embedded"`),
			MatchJQ(`.status.chart.version == "1.12.0"`),
		)),
	)
}

func TestDaprInstanceDeployWithCustomChart(t *testing.T) {
	test := With(t)

	instance := dapr.DeployInstance(
		test,
		daprAc.DaprInstanceSpec().
			WithChart(daprAc.ChartSpec().
				WithVersion("1.11.3")).
			WithValues(nil),
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

	test.Eventually(dapr.Instance(test, instance), TestTimeoutLong).Should(
		WithTransform(AsJSON(), And(
			MatchJQ(`.status.chart.name == "dapr"`),
			MatchJQ(`.status.chart.repo == "https://dapr.github.io/helm-charts"`),
			MatchJQ(`.status.chart.version == "1.11.3"`),
		)),
	)
}

func TestDaprInstanceDeployWithCustomSidecarImage(t *testing.T) {
	test := With(t)

	instance := dapr.DeployInstance(
		test,
		daprAc.DaprInstanceSpec().
			WithValues(dapr.Values(test, map[string]any{
				"dapr_sidecar_injector": map[string]any{
					"image": map[string]any{
						"name": "docker.io/daprio/daprd:" + test.ID(),
					},
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

	test.Eventually(dapr.Instance(test, instance), TestTimeoutLong).Should(
		WithTransform(AsJSON(), And(
			MatchJQ(`.status.chart.name == "dapr"`),
			MatchJQ(`.status.chart.repo == "embedded"`),
			MatchJQ(`.status.chart.version == "1.12.0"`),
		)),
	)

	test.Eventually(PodList(test, "app=dapr-sidecar-injector", instance.Namespace), TestTimeoutLong).Should(
		WithTransform(AsJSON(), And(
			MatchJQf(`.items[0].spec.containers[0].env[] | select(.name == "SIDECAR_IMAGE") | .value == "docker.io/daprio/daprd:%s"`, test.ID()),
			MatchJQf(`.items[0].spec.containers[0].env[] | select(.name == "SIDECAR_IMAGE_PULL_POLICY") | .value == "%s"`, corev1.PullAlways),
		)),
	)

}

func TestDaprInstanceDeployWithApp(t *testing.T) {
	test := With(t)

	instance := dapr.DeployInstance(
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

	test.Eventually(dapr.Instance(test, instance), TestTimeoutLong).Should(
		WithTransform(AsJSON(), And(
			MatchJQ(`.status.chart.name == "dapr"`),
			MatchJQ(`.status.chart.repo == "embedded"`),
			MatchJQ(`.status.chart.version == "1.12.0"`),
		)),
	)

	//
	// Dapr Application
	//

	daprTC.ValidateDaprApp(test, instance.Namespace)
}

func TestDaprInstanceDeployWrongCR(t *testing.T) {
	test := With(t)

	instance := dapr.DeployInstance(
		test,
		daprAc.DaprInstanceSpec().
			WithValues(nil),
		dapr.WithInstanceName(xid.New().String()),
		dapr.WithInstanceNamespace(controller.NamespaceDefault),
	)

	test.Eventually(dapr.Instance(test, instance), TestTimeoutLong).Should(
		WithTransform(ConditionStatus(conditions.TypeReconciled), Equal(corev1.ConditionFalse)))
	test.Eventually(dapr.Instance(test, instance), TestTimeoutLong).Should(
		WithTransform(ConditionReason(conditions.TypeReconciled), Equal(conditions.ReasonUnsupportedConfiguration)))
}
