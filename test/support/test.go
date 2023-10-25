package support

import (
	"bytes"
	"context"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/hashicorp/go-cleanhttp"

	"github.com/rs/xid"
	netv1 "k8s.io/api/networking/v1"

	"github.com/dapr-sandbox/dapr-kubernetes-operator/test/support/helm"

	"github.com/dapr-sandbox/dapr-kubernetes-operator/api/operator/v1alpha1"
	daprCP "github.com/dapr-sandbox/dapr-kubernetes-operator/internal/controller/operator"
	daprAc "github.com/dapr-sandbox/dapr-kubernetes-operator/pkg/client/operator/applyconfiguration/operator/v1alpha1"
	"github.com/dapr-sandbox/dapr-kubernetes-operator/pkg/pointer"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	corev1 "k8s.io/api/core/v1"

	"github.com/onsi/gomega"
)

//nolint:interfacebloat
type Test interface {
	T() *testing.T
	Ctx() context.Context
	Client() *Client
	HTTPClient() *http.Client

	NewTestNamespace(...Option[*corev1.Namespace]) *corev1.Namespace
	NewDaprControlPlane(*daprAc.DaprControlPlaneSpecApplyConfiguration) *v1alpha1.DaprControlPlane
	NewNamespacedNameDaprControlPlane(types.NamespacedName, *daprAc.DaprControlPlaneSpecApplyConfiguration) *v1alpha1.DaprControlPlane
	InstallChart(string, ...helm.InstallOption)
	SetUpIngress(string, netv1.HTTPIngressPath) *netv1.Ingress

	GET(string) func(g gomega.Gomega) (*http.Response, error)
	POST(string, string, []byte) func(g gomega.Gomega) (*http.Response, error)

	gomega.Gomega
}

type Option[T any] interface {
	applyTo(to T) error
}

type errorOption[T any] func(to T) error

//nolint:unused
func (o errorOption[T]) applyTo(to T) error {
	return o(to)
}

var _ Option[any] = errorOption[any](nil)

func With(t *testing.T) Test {
	t.Helper()

	t.Log()
	ctx := context.Background()
	if deadline, ok := t.Deadline(); ok {
		withDeadline, cancel := context.WithDeadline(ctx, deadline)
		t.Cleanup(cancel)
		ctx = withDeadline
	}

	answer := &T{
		WithT: gomega.NewWithT(t),
		t:     t,
		ctx:   ctx,
		http:  cleanhttp.DefaultClient(),
	}

	answer.SetDefaultEventuallyPollingInterval(500 * time.Millisecond)
	answer.SetDefaultEventuallyTimeout(TestTimeoutLong)
	answer.SetDefaultConsistentlyDuration(500 * time.Millisecond)
	answer.SetDefaultConsistentlyDuration(TestTimeoutLong)

	return answer
}

type T struct {
	*gomega.WithT

	t      *testing.T
	client *Client
	once   sync.Once
	http   *http.Client

	//nolint:containedctx
	ctx context.Context
}

func (t *T) T() *testing.T {
	return t.t
}

func (t *T) Ctx() context.Context {
	return t.ctx
}

func (t *T) Client() *Client {
	t.once.Do(func() {
		c, err := newClient(t.t.Logf)
		if err != nil {
			t.T().Fatalf("Error creating client: %v", err)
		}
		t.client = c
	})
	return t.client
}

func (t *T) HTTPClient() *http.Client {
	t.once.Do(func() {
		t.http = cleanhttp.DefaultClient()
	})

	return t.http
}

func (t *T) NewTestNamespace(options ...Option[*corev1.Namespace]) *corev1.Namespace {
	t.T().Helper()

	namespace := createTestNamespace(t, options...)

	t.T().Cleanup(func() {
		deleteTestNamespace(t, namespace)
	})

	return namespace
}

func (t *T) NewDaprControlPlane(
	spec *daprAc.DaprControlPlaneSpecApplyConfiguration,
) *v1alpha1.DaprControlPlane {

	return t.NewNamespacedNameDaprControlPlane(
		types.NamespacedName{
			Name:      daprCP.DaprControlPlaneName,
			Namespace: daprCP.DaprControlPlaneNamespaceDefault,
		},
		spec,
	)
}

func (t *T) NewNamespacedNameDaprControlPlane(
	nn types.NamespacedName,
	spec *daprAc.DaprControlPlaneSpecApplyConfiguration,
) *v1alpha1.DaprControlPlane {

	t.T().Logf("Setting up Dapr ControlPlane %s in namespace %s", nn.Name, nn.Namespace)

	cp := t.Client().Dapr.OperatorV1alpha1().DaprControlPlanes(nn.Namespace)

	instance, err := cp.Apply(
		t.Ctx(),
		daprAc.DaprControlPlane(nn.Name, nn.Namespace).
			WithSpec(spec),
		metav1.ApplyOptions{
			FieldManager: "dapr-e2e-" + t.T().Name(),
		})

	t.Expect(err).
		ToNot(gomega.HaveOccurred())

	t.T().Cleanup(func() {
		t.Expect(
			cp.Delete(t.Ctx(), instance.Name, metav1.DeleteOptions{
				PropagationPolicy: pointer.Any(metav1.DeletePropagationForeground),
			}),
		).ToNot(gomega.HaveOccurred())
	})

	return instance
}

func (t *T) InstallChart(
	chart string,
	options ...helm.InstallOption,
) {
	allopt := make([]helm.InstallOption, 0)
	allopt = append(allopt, options...)
	allopt = append(allopt, helm.WithInstallTimeout(TestTimeoutLong))

	release, err := t.Client().Helm.Install(
		t.Ctx(),
		chart,
		allopt...)

	t.Expect(err).
		ToNot(gomega.HaveOccurred())

	t.T().Cleanup(func() {
		err := t.Client().Helm.Uninstall(
			t.Ctx(),
			release.Name,
			helm.WithUninstallTimeout(TestTimeoutLong))

		t.Expect(err).
			ToNot(gomega.HaveOccurred())
	})
}

func (t *T) SetUpIngress(
	namespace string,
	path netv1.HTTPIngressPath,
) *netv1.Ingress {
	name := xid.New().String()

	t.T().Logf("Setting up ingress %s in namespace %s", name, namespace)

	ing, err := t.Client().NetworkingV1().Ingresses(namespace).Create(
		t.Ctx(),
		&netv1.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Spec: netv1.IngressSpec{
				Rules: []netv1.IngressRule{{
					IngressRuleValue: netv1.IngressRuleValue{
						HTTP: &netv1.HTTPIngressRuleValue{
							Paths: []netv1.HTTPIngressPath{path},
						},
					},
				}},
			},
		},
		metav1.CreateOptions{},
	)

	t.Expect(err).
		ToNot(gomega.HaveOccurred())
	t.Expect(ing).
		ToNot(gomega.BeNil())

	t.T().Cleanup(func() {
		t.Expect(
			t.Client().NetworkingV1().Ingresses(namespace).Delete(
				t.Ctx(),
				ing.Name,
				metav1.DeleteOptions{
					PropagationPolicy: pointer.Any(metav1.DeletePropagationForeground),
				},
			),
		).ToNot(gomega.HaveOccurred())
	})

	return ing
}

func (t *T) GET(url string) func(g gomega.Gomega) (*http.Response, error) {
	return func(g gomega.Gomega) (*http.Response, error) {
		req, err := http.NewRequestWithContext(t.Ctx(), http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}

		return t.HTTPClient().Do(req)
	}
}

func (t *T) POST(url string, contentType string, content []byte) func(g gomega.Gomega) (*http.Response, error) {
	return func(g gomega.Gomega) (*http.Response, error) {
		data := content
		if data == nil {
			data = []byte{}
		}

		req, err := http.NewRequestWithContext(t.Ctx(), http.MethodPost, url, bytes.NewReader(data))
		if err != nil {
			return nil, err
		}

		if contentType != "" {
			req.Header.Add("Content-Type", contentType)
		}

		return t.HTTPClient().Do(req)
	}
}
