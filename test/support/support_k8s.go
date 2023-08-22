package support

import (
	"github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Deployment(t Test, name string, namespace string) func(g gomega.Gomega) (*appsv1.Deployment, error) {
	return func(g gomega.Gomega) (*appsv1.Deployment, error) {
		answer, err := t.Client().AppsV1().Deployments(namespace).Get(
			t.Ctx(),
			name,
			metav1.GetOptions{},
		)

		if k8serrors.IsNotFound(err) {
			return nil, nil
		}

		return answer, err
	}
}
