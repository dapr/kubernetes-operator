package support

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/onsi/gomega"

	"github.com/dapr-sandbox/dapr-kubernetes-operator/pkg/pointer"
	"github.com/dapr-sandbox/dapr-kubernetes-operator/pkg/resources"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	TestTimeoutMini   = 5 * time.Second
	TestTimeoutShort  = 1 * time.Minute
	TestTimeoutMedium = 2 * time.Minute
	TestTimeoutLong   = 5 * time.Minute
)

func runCleanup(t Test, in runtime.Object) error {
	un, err := resources.ToUnstructured(t.Client().Scheme(), in)
	if err != nil {
		return fmt.Errorf("failed to run ToUnstructured, %w", err)
	}

	rc, err := t.Client().ForResource(un)
	if err != nil {
		return fmt.Errorf("failed to compute ResourceInterface, %w", err)
	}

	if t.T().Failed() {
		raw, err := rc.Get(t.Ctx(), un.GetName(), metav1.GetOptions{})
		if err != nil && !k8serrors.IsNotFound(err) {
			return fmt.Errorf("failed to get current object, %w", err)
		}
		if err != nil && k8serrors.IsNotFound(err) {
			return nil
		}

		data, err := json.MarshalIndent(raw, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to unmarshal curret object, %w", err)
		}

		t.T().Log(string(data))
	}

	t.T().Logf("Deleting resource %s, with name %s in namespace %s",
		un.GetObjectKind().GroupVersionKind().String(),
		un.GetName(),
		un.GetNamespace())

	err = rc.Delete(
		t.Ctx(),
		un.GetName(),
		metav1.DeleteOptions{
			PropagationPolicy: pointer.Any(metav1.DeletePropagationForeground),
		})

	if err != nil && !k8serrors.IsNotFound(err) {
		return fmt.Errorf("failed to delete object, %w", err)
	}

	t.Eventually(Resource(t, rc, un), TestTimeoutShort).Should(
		gomega.BeNil(),
		fmt.Sprintf("Failure deleting resource %s, with name %s in namespace %s",
			un.GetObjectKind().GroupVersionKind().String(),
			un.GetName(),
			un.GetNamespace(),
		),
	)

	t.T().Logf("Resource %s, with name %s in namespace %s deleted",
		un.GetObjectKind().GroupVersionKind().String(),
		un.GetName(),
		un.GetNamespace())

	return nil
}
