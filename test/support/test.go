package support

import (
	"context"
	"sync"
	"testing"

	"github.com/dapr-sandbox/dapr-kubernetes-operator/api/operator/v1alpha1"
	daprCP "github.com/dapr-sandbox/dapr-kubernetes-operator/internal/controller/operator"
	daprAc "github.com/dapr-sandbox/dapr-kubernetes-operator/pkg/client/operator/applyconfiguration/operator/v1alpha1"
	"github.com/dapr-sandbox/dapr-kubernetes-operator/pkg/pointer"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	corev1 "k8s.io/api/core/v1"

	"github.com/onsi/gomega"
)

type Test interface {
	T() *testing.T
	Ctx() context.Context
	Client() *Client

	NewTestNamespace(...Option[*corev1.Namespace]) *corev1.Namespace
	NewDaprControlPlane(spec *daprAc.DaprControlPlaneSpecApplyConfiguration) *v1alpha1.DaprControlPlane
	NewNamespacedNameDaprControlPlane(types.NamespacedName, *daprAc.DaprControlPlaneSpecApplyConfiguration) *v1alpha1.DaprControlPlane

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
	ctx := context.Background()
	if deadline, ok := t.Deadline(); ok {
		withDeadline, cancel := context.WithDeadline(ctx, deadline)
		t.Cleanup(cancel)
		ctx = withDeadline
	}

	return &T{
		WithT: gomega.NewWithT(t),
		t:     t,
		ctx:   ctx,
	}
}

type T struct {
	*gomega.WithT

	t      *testing.T
	client *Client
	once   sync.Once

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
		c, err := newClient()
		if err != nil {
			t.T().Fatalf("Error creating client: %v", err)
		}
		t.client = c
	})
	return t.client
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

	cp := t.Client().DaprCP

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
