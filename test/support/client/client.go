package client

import (
	"fmt"
	"testing"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"

	"sigs.k8s.io/controller-runtime/pkg/client/config"

	daprClient "github.com/dapr-sandbox/dapr-kubernetes-operator/pkg/client/operator/clientset/versioned"
	olmAC "github.com/operator-framework/operator-lifecycle-manager/pkg/api/client/clientset/versioned"
)

type Client struct {
	kubernetes.Interface

	dapr      daprClient.Interface
	discovery discovery.DiscoveryInterface
	dynamic   dynamic.Interface
	olm       olmAC.Interface
	scheme    *runtime.Scheme
	config    *rest.Config
}

func New(t *testing.T) (*Client, error) {
	t.Helper()

	cfg, err := config.GetConfig()
	if err != nil {
		return nil, err
	}

	discoveryClient, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		return nil, err
	}
	dynamicClient, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	dpClient, err := daprClient.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	oClient, err := olmAC.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	c := Client{
		Interface: kubeClient,
		discovery: discoveryClient,
		dynamic:   dynamicClient,
		dapr:      dpClient,
		olm:       oClient,
		config:    cfg,
		scheme:    scheme.Scheme,
	}

	return &c, nil
}

func (c *Client) Dapr() daprClient.Interface {
	return c.dapr
}
func (c *Client) Discovery() discovery.DiscoveryInterface {
	return c.discovery
}
func (c *Client) Dynamic() dynamic.Interface {
	return c.dynamic
}
func (c *Client) OLM() olmAC.Interface {
	return c.olm
}
func (c *Client) Scheme() *runtime.Scheme {
	return c.scheme
}

func (c *Client) RESTMapper() (meta.RESTMapper, error) {
	gr, err := restmapper.GetAPIGroupResources(c.Discovery())
	if err != nil {
		return nil, err
	}

	return restmapper.NewDiscoveryRESTMapper(gr), nil

}
func (c *Client) ForResource(in *unstructured.Unstructured) (dynamic.ResourceInterface, error) {
	gvk := in.GetObjectKind().GroupVersionKind()
	gk := schema.GroupKind{Group: gvk.Group, Kind: gvk.Kind}

	rm, err := c.RESTMapper()
	if err != nil {
		return nil, fmt.Errorf("error computing RESTMapper, %w", err)
	}

	mapping, err := rm.RESTMapping(gk, gvk.Version)
	if err != nil {
		return nil, fmt.Errorf("error computing resource mapping, %w", err)
	}

	var rc dynamic.ResourceInterface

	if in.GetNamespace() != "" {
		rc = c.Dynamic().Resource(mapping.Resource).Namespace(in.GetNamespace())
	} else {
		rc = c.Dynamic().Resource(mapping.Resource)
	}

	return rc, nil
}
