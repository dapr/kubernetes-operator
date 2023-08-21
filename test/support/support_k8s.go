package support

import (
	"github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
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

func Pod(t Test, name string, namespace string) func(g gomega.Gomega) (*corev1.Pod, error) {
	return func(g gomega.Gomega) (*corev1.Pod, error) {
		answer, err := t.Client().CoreV1().Pods(namespace).Get(
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

func Service(t Test, name string, namespace string) func(g gomega.Gomega) (*corev1.Service, error) {
	return func(g gomega.Gomega) (*corev1.Service, error) {
		answer, err := t.Client().CoreV1().Services(namespace).Get(
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

func Ingress(t Test, name string, namespace string) func(g gomega.Gomega) (*netv1.Ingress, error) {
	return func(g gomega.Gomega) (*netv1.Ingress, error) {
		answer, err := t.Client().NetworkingV1().Ingresses(namespace).Get(
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

func ExtractFirstLoadBalancerIngressHostname() func(*netv1.Ingress) string {
	return func(in *netv1.Ingress) string {
		if len(in.Status.LoadBalancer.Ingress) != 1 {
			return ""
		}

		return in.Status.LoadBalancer.Ingress[0].Hostname
	}
}
