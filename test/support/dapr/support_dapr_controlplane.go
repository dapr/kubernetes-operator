package dapr

import (
	daprApi "github.com/dapr-sandbox/dapr-kubernetes-operator/api/operator/v1alpha1"
	"github.com/dapr-sandbox/dapr-kubernetes-operator/internal/controller/operator/controlplane"
	daprAc "github.com/dapr-sandbox/dapr-kubernetes-operator/pkg/client/applyconfiguration/operator/v1alpha1"
	"github.com/dapr-sandbox/dapr-kubernetes-operator/pkg/controller"
	"github.com/dapr-sandbox/dapr-kubernetes-operator/test/support"
	"github.com/onsi/gomega"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

func ControlPlane(t support.Test, dapr *daprApi.DaprControlPlane) func(g gomega.Gomega) (*daprApi.DaprControlPlane, error) {
	return func(g gomega.Gomega) (*daprApi.DaprControlPlane, error) {
		answer, err := t.Client().Dapr().OperatorV1alpha1().DaprControlPlanes(dapr.Namespace).Get(
			t.Ctx(),
			dapr.Name,
			metav1.GetOptions{},
		)

		if k8serrors.IsNotFound(err) {
			return nil, nil
		}

		return answer, err
	}
}

type ControlPlaneOptions struct {
	types.NamespacedName
}

type ControlPlaneOption func(*ControlPlaneOptions)

func WithControlPlaneName(value string) ControlPlaneOption {
	return func(options *ControlPlaneOptions) {
		options.Name = value
	}
}

func WithControlPlaneNamespace(value string) ControlPlaneOption {
	return func(options *ControlPlaneOptions) {
		options.Namespace = value
	}
}

func DeployControlPlane(
	t support.Test,
	spec *daprAc.DaprControlPlaneSpecApplyConfiguration,
	opts ...ControlPlaneOption,
) *daprApi.DaprControlPlane {
	t.T().Helper()

	cpo := ControlPlaneOptions{}
	cpo.Name = controlplane.DaprControlPlaneResourceName
	cpo.Namespace = controller.NamespaceDefault

	for _, o := range opts {
		o(&cpo)
	}

	t.T().Logf("Setting up DaprControlPlane %s in namespace %s", cpo.Name, cpo.Namespace)

	cp := t.Client().Dapr().OperatorV1alpha1().DaprControlPlanes(cpo.Namespace)

	res, err := cp.Apply(
		t.Ctx(),
		daprAc.DaprControlPlane(cpo.Name, cpo.Namespace).
			WithSpec(spec),
		metav1.ApplyOptions{
			FieldManager: "dapr-e2e-" + t.ID(),
			Force:        true,
		})

	t.Expect(err).
		ToNot(gomega.HaveOccurred())

	t.Cleanup(func() []runtime.Object {
		return []runtime.Object{res}
	})

	return res
}
