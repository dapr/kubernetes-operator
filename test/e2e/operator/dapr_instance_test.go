package operator

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/lburgazzoli/gomega-matchers/pkg/matchers/jq"

	"github.com/dapr/kubernetes-operator/pkg/conditions"
	"github.com/dapr/kubernetes-operator/pkg/controller"
	"github.com/dapr/kubernetes-operator/test/support/dapr"

	"github.com/rs/xid"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	. "github.com/dapr/kubernetes-operator/test/support"
	. "github.com/onsi/gomega"

	daprAc "github.com/dapr/kubernetes-operator/pkg/client/applyconfiguration/operator/v1alpha1"
	daprTC "github.com/dapr/kubernetes-operator/test/e2e/common"
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
		WithTransform(json.Marshal, And(
			jq.Match(`.status.chart.name == "dapr"`),
			jq.Match(`.status.chart.repo == "embedded"`),
			jq.Match(`.status.chart.version == "%s"`, os.Getenv("DAPR_HELM_CHART_VERSION")),
		)),
	)
}

func TestDaprInstanceGC(t *testing.T) {
	test := With(t)

	{
		instance := dapr.DeployInstance(
			test,
			daprAc.DaprInstanceSpec().
				WithValues(nil),
		)

		test.Eventually(Deployment(test, "dapr-operator", instance.Namespace), TestTimeoutLong).Should(
			WithTransform(ConditionStatus(appsv1.DeploymentAvailable), Equal(corev1.ConditionTrue)))
		test.Eventually(Deployment(test, "dapr-sentry", instance.Namespace), TestTimeoutLong).Should(
			WithTransform(ConditionStatus(appsv1.DeploymentAvailable), Equal(corev1.ConditionTrue)))
		test.Eventually(Deployment(test, "dapr-sidecar-injector", instance.Namespace), TestTimeoutLong).Should(
			WithTransform(ConditionStatus(appsv1.DeploymentAvailable), Equal(corev1.ConditionTrue)))
	}

	{
		instance := dapr.DeployInstance(
			test,
			daprAc.DaprInstanceSpec().
				WithValues(dapr.Values(test, map[string]any{
					"dapr_sidecar_injector": map[string]any{
						"enabled": false,
					},
				})),
		)

		test.Eventually(Deployment(test, "dapr-operator", instance.Namespace), TestTimeoutLong).Should(
			WithTransform(ConditionStatus(appsv1.DeploymentAvailable), Equal(corev1.ConditionTrue)))
		test.Eventually(Deployment(test, "dapr-sentry", instance.Namespace), TestTimeoutLong).Should(
			WithTransform(ConditionStatus(appsv1.DeploymentAvailable), Equal(corev1.ConditionTrue)))
		test.Eventually(Deployment(test, "dapr-sidecar-injector", instance.Namespace), TestTimeoutLong).Should(
			BeNil())
	}
}

func TestDaprInstanceDeployWithCustomChart(t *testing.T) {
	test := With(t)

	instance := dapr.DeployInstance(
		test,
		daprAc.DaprInstanceSpec().
			WithChart(daprAc.ChartSpec().
				WithVersion("1.14.0")).
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
		WithTransform(json.Marshal, And(
			jq.Match(`.status.chart.name == "dapr"`),
			jq.Match(`.status.chart.repo == "https://dapr.github.io/helm-charts"`),
			jq.Match(`.status.chart.version == "1.14.0"`),
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
		WithTransform(json.Marshal, And(
			jq.Match(`.status.chart.name == "dapr"`),
			jq.Match(`.status.chart.repo == "embedded"`),
			jq.Match(`.status.chart.version == "%s"`, os.Getenv("DAPR_HELM_CHART_VERSION")),
		)),
	)

	test.Eventually(PodList(test, "app=dapr-sidecar-injector", instance.Namespace), TestTimeoutLong).Should(
		WithTransform(json.Marshal, And(
			jq.Match(`.items[0].spec.containers[0].env[] | select(.name == "SIDECAR_IMAGE") | .value == "docker.io/daprio/daprd:%s"`, test.ID()),
			jq.Match(`.items[0].spec.containers[0].env[] | select(.name == "SIDECAR_IMAGE_PULL_POLICY") | .value == "%s"`, corev1.PullAlways),
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
		WithTransform(json.Marshal, And(
			jq.Match(`.status.chart.name == "dapr"`),
			jq.Match(`.status.chart.repo == "embedded"`),
			jq.Match(`.status.chart.version == "%s"`, os.Getenv("DAPR_HELM_CHART_VERSION")),
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
