package client

import (
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/discovery"
	memory "k8s.io/client-go/discovery/cached"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/scale"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"

	daprClient "github.com/dapr-sandbox/dapr-kubernetes-operator/pkg/client/operator/clientset/versioned"
)

var scaleConverter = scale.NewScaleConverter()
var codecs = serializer.NewCodecFactory(scaleConverter.Scheme())

type Client struct {
	ctrl.Client
	kubernetes.Interface

	Dapr      daprClient.Interface
	Discovery discovery.DiscoveryInterface

	dynamic *dynamic.DynamicClient
	scheme  *runtime.Scheme
	config  *rest.Config
	rest    rest.Interface
	mapper  *restmapper.DeferredDiscoveryRESTMapper
}

func NewClient(cfg *rest.Config, scheme *runtime.Scheme, cc ctrl.Client) (*Client, error) {

	discoveryCl, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		return nil, err
	}
	kubeCl, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	restCl, err := newRESTClientForConfig(cfg)
	if err != nil {
		return nil, err
	}
	dynCl, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	daprCl, err := daprClient.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	c := Client{
		Client:    cc,
		Interface: kubeCl,
		Discovery: discoveryCl,
		Dapr:      daprCl,
		dynamic:   dynCl,
		mapper:    restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(discoveryCl)),
		scheme:    scheme,
		config:    cfg,
		rest:      restCl,
	}

	return &c, nil
}

func newRESTClientForConfig(config *rest.Config) (*rest.RESTClient, error) {
	cfg := rest.CopyConfig(config)
	// so that the RESTClientFor doesn't complain
	cfg.GroupVersion = &schema.GroupVersion{}
	cfg.NegotiatedSerializer = codecs.WithoutConversion()
	if len(cfg.UserAgent) == 0 {
		cfg.UserAgent = rest.DefaultKubernetesUserAgent()
	}

	return rest.RESTClientFor(cfg)
}

// IsOpenShift returns true if we are connected to a OpenShift cluster.
func (c *Client) IsOpenShift() (bool, error) {
	if c.Discovery == nil {
		return false, nil
	}

	return IsOpenShift(c.Discovery)
}

func (c *Client) Dynamic(namespace string, obj *unstructured.Unstructured) (dynamic.ResourceInterface, error) {
	mapping, err := c.mapper.RESTMapping(obj.GroupVersionKind().GroupKind(), obj.GroupVersionKind().Version)
	if err != nil {
		return nil, err
	}

	var dr dynamic.ResourceInterface

	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		dr = &NamespacedResource{
			ResourceInterface: c.dynamic.Resource(mapping.Resource).Namespace(namespace),
		}
	} else {
		dr = &ClusteredResource{
			ResourceInterface: c.dynamic.Resource(mapping.Resource),
		}
	}

	return dr, nil
}

func IsOpenShift(d discovery.DiscoveryInterface) (bool, error) {
	_, err := d.ServerResourcesForGroupVersion("route.openshift.io/v1")
	if err != nil && k8serrors.IsNotFound(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return true, nil
}
