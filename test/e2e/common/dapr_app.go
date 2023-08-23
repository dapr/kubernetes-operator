package common

import (
	"fmt"
	"net/http"

	"github.com/dapr-sandbox/dapr-kubernetes-operator/pkg/pointer"
	netv1 "k8s.io/api/networking/v1"

	appsv1 "k8s.io/api/apps/v1"

	corev1 "k8s.io/api/core/v1"

	"github.com/dapr-sandbox/dapr-kubernetes-operator/test/support"
	"github.com/dapr-sandbox/dapr-kubernetes-operator/test/support/helm"
	"github.com/onsi/gomega"
)

func ValidateDaprApp(test support.Test, namespace string) {
	test.T().Helper()

	//
	// Install Dapr Test App
	//

	test.InstallChart(
		"oci://docker.io/salaboy/testing-app",
		helm.WithInstallName("testing-app"),
		helm.WithInstallNamespace(namespace),
		helm.WithInstallVersion("v0.1.0"),
	)

	test.Eventually(support.Deployment(test, "testing-app-deployment", namespace), support.TestTimeoutShort).Should(
		gomega.WithTransform(support.ConditionStatus(appsv1.DeploymentAvailable), gomega.Equal(corev1.ConditionTrue)),
	)
	test.Eventually(support.Service(test, "testing-app-service", namespace), support.TestTimeoutShort).Should(
		gomega.Not(gomega.BeNil()),
	)

	//
	// Expose app
	//

	ing := test.SetUpIngress(namespace, netv1.HTTPIngressPath{
		PathType: pointer.Any(netv1.PathTypePrefix),
		Path:     "/",
		Backend: netv1.IngressBackend{
			Service: &netv1.IngressServiceBackend{
				Name: "testing-app-service",
				Port: netv1.ServiceBackendPort{
					Number: 80,
				},
			},
		},
	})

	test.Eventually(support.Ingress(test, ing.Name, ing.Namespace), support.TestTimeoutLong).Should(
		gomega.WithTransform(
			support.ExtractFirstLoadBalancerIngressHostname(),
			gomega.Equal("localhost")),
		"Failure to set-up ingress")

	//
	// Test the app
	//

	test.T().Log("test app")

	base := fmt.Sprintf("http://localhost:%d", 8081)

	//nolint:bodyclose
	test.Eventually(test.GET(base+"/read"), support.TestTimeoutLong).Should(
		gomega.And(
			gomega.HaveHTTPStatus(http.StatusOK),
			gomega.HaveHTTPBody(gomega.MatchJSON(`{ "Values": null }`)),
		),
		"Failure to invoke initial read",
	)

	//nolint:bodyclose
	test.Eventually(test.POST(base+"/write?message=hello", "text/plain", nil), support.TestTimeoutLong).Should(
		gomega.HaveHTTPStatus(http.StatusOK),
		"Failure to invoke post",
	)

	//nolint:bodyclose
	test.Eventually(test.GET(base+"/read"), support.TestTimeoutLong).Should(
		gomega.And(
			gomega.HaveHTTPStatus(http.StatusOK),
			gomega.HaveHTTPBody(gomega.MatchJSON(`{ "Values":["hello"] }`)),
		),
		"Failure to invoke read",
	)
}
