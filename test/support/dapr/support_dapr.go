package dapr

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/dapr/kubernetes-operator/pkg/pointer"

	"k8s.io/apimachinery/pkg/runtime"

	netv1 "k8s.io/api/networking/v1"
	netv1ac "k8s.io/client-go/applyconfigurations/networking/v1"

	daprAc "github.com/dapr/kubernetes-operator/pkg/client/applyconfiguration/operator/v1alpha1"
	"github.com/dapr/kubernetes-operator/pkg/resources"
	"github.com/dapr/kubernetes-operator/test/support"
	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
	appsv1ac "k8s.io/client-go/applyconfigurations/apps/v1"
	corev1ac "k8s.io/client-go/applyconfigurations/core/v1"
	metav1ac "k8s.io/client-go/applyconfigurations/meta/v1"
)

const (
	TestAppPort        = 8080
	TestAppServicePort = 80
)

func DeployTestApp(t support.Test, name string, namespace string) {
	t.T().Helper()
	t.T().Logf("Setting up Dapr Application %s in namespace %s", name, namespace)

	componentRes := schema.GroupVersionResource{Group: "dapr.io", Version: "v1alpha1", Resource: "components"}

	//
	// Component
	//

	t.T().Logf("Setting up Dapr Component %s in namespace %s", name, namespace)

	c, err := t.Client().Dynamic().Resource(componentRes).Namespace(namespace).Apply(
		t.Ctx(),
		name,
		&unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": componentRes.GroupVersion().String(),
				"kind":       "Component",
				"metadata": map[string]interface{}{
					"name":      name,
					"namespace": namespace,
				},
				"spec": map[string]interface{}{
					"type":     "state.in-memory",
					"version":  "v1",
					"metadata": []interface{}{},
				},
			},
		},
		metav1.ApplyOptions{
			FieldManager: "dapr-e2e-" + t.ID(),
			Force:        true,
		},
	)

	t.Expect(err).
		ToNot(gomega.HaveOccurred())

	t.Cleanup(func() []runtime.Object {
		err := t.Client().Dynamic().Resource(componentRes).Namespace(namespace).Delete(
			t.Ctx(),
			name,
			metav1.DeleteOptions{
				PropagationPolicy: pointer.Any(metav1.DeletePropagationForeground),
			},
		)

		t.Expect(err).
			ToNot(gomega.HaveOccurred())

		return []runtime.Object{}
	})

	t.T().Logf("Dapr Component %s in namespace %s has been created", c.GetName(), c.GetNamespace())

	//
	// Deployment
	//

	t.T().Logf("Setting up App Deployment %s in namespace %s", name, namespace)

	d, err := t.Client().AppsV1().Deployments(namespace).Apply(
		t.Ctx(),
		appsv1ac.Deployment(name, namespace).
			WithLabels(map[string]string{
				"app": name,
			}).
			WithSpec(appsv1ac.DeploymentSpec().
				WithSelector(metav1ac.LabelSelector().WithMatchLabels(map[string]string{
					"app": name,
				})).
				WithTemplate(corev1ac.PodTemplateSpec().
					WithLabels(map[string]string{
						"app": name,
					}).
					WithAnnotations(map[string]string{
						"dapr.io/app-id":             name,
						"dapr.io/app-port":           strconv.Itoa(TestAppPort),
						"dapr.io/enabled":            "true",
						"dapr.io/enable-api-logging": "true",
					}).
					WithSpec(corev1ac.PodSpec().
						WithContainers(corev1ac.Container().
							WithImage("kind.local/dapr-test-app:latest").
							WithImagePullPolicy(corev1.PullNever).
							WithName("app").
							WithPorts(resources.WithPort("http", TestAppPort)).
							WithReadinessProbe(resources.WithHTTPProbe("/health/readiness", TestAppPort)).
							WithLivenessProbe(resources.WithHTTPProbe("/health/liveness", TestAppPort)).
							WithEnv(resources.WithEnv("STATESTORE_NAME", name)),
						),
					),
				),
			),
		metav1.ApplyOptions{
			FieldManager: "dapr-e2e-" + t.ID(),
			Force:        true,
		})

	t.Expect(err).
		ToNot(gomega.HaveOccurred())

	t.Cleanup(func() []runtime.Object {
		return []runtime.Object{d}
	})

	//
	// Service
	//

	t.T().Logf("Setting up App Service %s in namespace %s", name, namespace)

	s, err := t.Client().CoreV1().Services(namespace).Apply(
		t.Ctx(),
		corev1ac.Service(name, namespace).
			WithLabels(map[string]string{
				"app": name,
			}).
			WithSpec(corev1ac.ServiceSpec().
				WithPorts(corev1ac.ServicePort().
					WithName("http").
					WithProtocol(corev1.ProtocolTCP).
					WithPort(TestAppServicePort).
					WithTargetPort(intstr.FromInt32(TestAppPort))).
				WithSelector(map[string]string{
					"app": name,
				}).
				WithSessionAffinity(corev1.ServiceAffinityNone).
				WithPublishNotReadyAddresses(false)),
		metav1.ApplyOptions{
			FieldManager: "dapr-e2e-" + t.ID(),
			Force:        true,
		})

	t.Expect(err).
		ToNot(gomega.HaveOccurred())

	t.Cleanup(func() []runtime.Object {
		return []runtime.Object{s}
	})

	//
	// Ingress
	//

	t.T().Logf("Setting up ingress %s in namespace %s", name, namespace)

	path := name
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	if !strings.HasSuffix(path, "(/|$)(.*)") {
		path += "(/|$)(.*)"
	}

	ing, err := t.Client().NetworkingV1().Ingresses(namespace).Apply(
		t.Ctx(),
		netv1ac.Ingress(name, namespace).
			WithAnnotations(map[string]string{
				"nginx.ingress.kubernetes.io/use-regex":      "true",
				"nginx.ingress.kubernetes.io/rewrite-target": "/$2",
			}).
			WithSpec(netv1ac.IngressSpec().
				WithRules(netv1ac.IngressRule().
					WithHTTP(netv1ac.HTTPIngressRuleValue().
						WithPaths(netv1ac.HTTPIngressPath().
							WithPathType(netv1.PathTypeImplementationSpecific).
							WithPath(path).
							WithBackend(netv1ac.IngressBackend().
								WithService(netv1ac.IngressServiceBackend().
									WithName(name).
									WithPort(netv1ac.ServiceBackendPort().
										WithNumber(TestAppServicePort),
									),
								),
							),
						),
					),
				),
			),
		metav1.ApplyOptions{
			FieldManager: "dapr-e2e-" + t.ID(),
			Force:        true,
		},
	)

	t.Expect(err).
		ToNot(gomega.HaveOccurred())
	t.Expect(ing).
		ToNot(gomega.BeNil())

	t.Cleanup(func() []runtime.Object {
		return []runtime.Object{ing}
	})
}

func Values(t support.Test, in map[string]any) *daprAc.JSONApplyConfiguration {
	t.T().Helper()

	if len(in) == 0 {
		return nil
	}

	data, err := json.Marshal(in)

	t.Expect(err).
		ToNot(gomega.HaveOccurred())

	cfg := daprAc.JSON()
	cfg.RawMessage = data

	return cfg
}
