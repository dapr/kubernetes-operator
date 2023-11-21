package dapr

import (
	daprApi "github.com/dapr-sandbox/dapr-kubernetes-operator/api/operator/v1alpha1"
	"github.com/dapr-sandbox/dapr-kubernetes-operator/internal/controller/operator/instance"
	daprAc "github.com/dapr-sandbox/dapr-kubernetes-operator/pkg/client/operator/applyconfiguration/operator/v1alpha1"
	"github.com/dapr-sandbox/dapr-kubernetes-operator/pkg/controller"
	"github.com/dapr-sandbox/dapr-kubernetes-operator/test/support"
	"github.com/onsi/gomega"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

func Instance(t support.Test, dapr *daprApi.DaprInstance) func(g gomega.Gomega) (*daprApi.DaprInstance, error) {
	return func(g gomega.Gomega) (*daprApi.DaprInstance, error) {
		answer, err := t.Client().Dapr().OperatorV1alpha1().DaprInstances(dapr.Namespace).Get(
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

type InstanceOptions struct {
	types.NamespacedName
}

type InstanceOption func(*InstanceOptions)

func WithInstanceName(value string) InstanceOption {
	return func(options *InstanceOptions) {
		options.Name = value
	}
}

func WithInstanceNamespace(value string) InstanceOption {
	return func(options *InstanceOptions) {
		options.Namespace = value
	}
}

func DeployInstance(
	t support.Test,
	spec *daprAc.DaprInstanceSpecApplyConfiguration,
	opts ...InstanceOption,
) *daprApi.DaprInstance {
	t.T().Helper()

	io := InstanceOptions{}
	io.Name = instance.DaprInstanceResourceName
	io.Namespace = controller.NamespaceDefault

	for _, o := range opts {
		o(&io)
	}

	t.T().Logf("Setting up DaprInstance %s in namespace %s", io.Name, io.Namespace)

	cp := t.Client().Dapr().OperatorV1alpha1().DaprInstances(io.Namespace)

	res, err := cp.Apply(
		t.Ctx(),
		daprAc.DaprInstance(io.Name, io.Namespace).
			WithSpec(spec),
		metav1.ApplyOptions{
			FieldManager: "dapr-e2e-" + t.ID(),
			Force:        true,
		})

	t.Expect(err).
		ToNot(gomega.HaveOccurred())

	t.T().Logf("DaprInstance %s in namespace %s created", io.Name, io.Namespace)

	t.Cleanup(func() []runtime.Object {
		return []runtime.Object{res}
	})

	return res
}
