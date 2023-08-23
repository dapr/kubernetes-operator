package support

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/dapr-sandbox/dapr-kubernetes-operator/test/support/helm"

	olmAC "github.com/operator-framework/operator-lifecycle-manager/pkg/api/client/clientset/versioned"

	daprClient "github.com/dapr-sandbox/dapr-kubernetes-operator/pkg/client/operator/clientset/versioned"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type Client struct {
	kubernetes.Interface

	Dapr      daprClient.Interface
	Discovery discovery.DiscoveryInterface
	OLM       olmAC.Interface
	Helm      *helm.Helm

	//nolint:unused
	scheme *runtime.Scheme
	config *rest.Config
}

func newClient(logFn func(string, ...interface{})) (*Client, error) {
	kc := os.Getenv("KUBECONFIG")
	if kc == "" {
		home := homedir.HomeDir()
		if home != "" {
			kc = filepath.Join(home, ".kube", "config")
		}
	}

	if kc == "" {
		return nil, errors.New("unable to determine KUBECONFIG")
	}

	cfg, err := clientcmd.BuildConfigFromFlags("", kc)
	if err != nil {
		return nil, err
	}

	discoveryClient, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		return nil, err
	}
	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	dClient, err := daprClient.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	oClient, err := olmAC.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	hClient, err := helm.New(helm.WithLog(logFn))
	if err != nil {
		return nil, err
	}

	c := Client{
		Interface: kubeClient,
		Discovery: discoveryClient,
		Dapr:      dClient,
		OLM:       oClient,
		Helm:      hClient,
		config:    cfg,
	}

	return &c, nil
}
