package dapr

import (
	daprApi "github.com/dapr/kubernetes-operator/api/operator/v1beta1"
	"github.com/dapr/kubernetes-operator/internal/controller/operator/instance"
	daprAc "github.com/dapr/kubernetes-operator/pkg/client/applyconfiguration/operator/v1beta1"
	"github.com/dapr/kubernetes-operator/test/support"
	"github.com/onsi/gomega"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func Instance(t support.Test, dapr *daprApi.DaprInstance) func(g gomega.Gomega) (*daprApi.DaprInstance, error) {
	return func(g gomega.Gomega) (*daprApi.DaprInstance, error) {
		answer, err := t.Client().Dapr().OperatorV1beta1().DaprInstances().Get(
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

func DeployInstance(
	t support.Test,
	spec *daprAc.DaprInstanceSpecApplyConfiguration,
) *daprApi.DaprInstance {
	t.T().Helper()

	cp := t.Client().Dapr().OperatorV1beta1().DaprInstances()

	res, err := cp.Apply(
		t.Ctx(),
		daprAc.DaprInstance(instance.DaprInstanceResourceName).
			WithSpec(spec),
		metav1.ApplyOptions{
			FieldManager: "dapr-e2e-" + t.ID(),
			Force:        true,
		})

	t.Expect(err).
		ToNot(gomega.HaveOccurred())

	t.T().Logf("DaprInstance %s in namespace %s created", res.Name, res.Spec.Deployment.Namespace)

	t.Cleanup(func() []runtime.Object {
		return []runtime.Object{res}
	})

	return res
}
