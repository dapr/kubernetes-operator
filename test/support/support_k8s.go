package support

import (
	"github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/dynamic"
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

func Deployments(t Test, namespace string, selector labels.Selector) func(g gomega.Gomega) ([]appsv1.Deployment, error) {
	return func(g gomega.Gomega) ([]appsv1.Deployment, error) {
		answer, err := t.Client().AppsV1().Deployments(namespace).List(
			t.Ctx(),
			metav1.ListOptions{
				LabelSelector: selector.String(),
			},
		)

		if k8serrors.IsNotFound(err) {
			return nil, nil
		}

		return answer.Items, err
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

func PodList(t Test, selector string, namespace string) func(g gomega.Gomega) (*corev1.PodList, error) {
	return func(g gomega.Gomega) (*corev1.PodList, error) {
		answer, err := t.Client().CoreV1().Pods(namespace).List(
			t.Ctx(),
			metav1.ListOptions{
				LabelSelector: selector,
			},
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

func Resource(t Test, ri dynamic.ResourceInterface, un *unstructured.Unstructured) func(g gomega.Gomega) (*unstructured.Unstructured, error) {
	return func(g gomega.Gomega) (*unstructured.Unstructured, error) {
		raw, err := ri.Get(t.Ctx(), un.GetName(), metav1.GetOptions{})
		if err != nil && !k8serrors.IsNotFound(err) {
			return nil, err
		}

		if err != nil && k8serrors.IsNotFound(err) {
			return nil, nil
		}

		return raw, err
	}
}

func CustomResourceDefinition(t Test, name string) func(g gomega.Gomega) (*apiextv1.CustomResourceDefinition, error) {
	return func(g gomega.Gomega) (*apiextv1.CustomResourceDefinition, error) {
		answer, err := t.Client().CustomResourceDefinitions().Get(
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
