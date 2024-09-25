package client

import (
	"context"
	"fmt"
	"time"

	"github.com/dapr/kubernetes-operator/pkg/resources"

	"golang.org/x/time/rate"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/scale"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"

	daprClient "github.com/dapr/kubernetes-operator/pkg/client/clientset/versioned"
)

const (
	DiscoveryLimiterBurst = 30
)

var scaleConverter = scale.NewScaleConverter()
var codecs = serializer.NewCodecFactory(scaleConverter.Scheme())

type Client struct {
	ctrl.Client
	kubernetes.Interface
	apiextv1.ApiextensionsV1Interface

	Dapr      daprClient.Interface
	Discovery discovery.DiscoveryInterface

	dynamic          *dynamic.DynamicClient
	scheme           *runtime.Scheme
	config           *rest.Config
	rest             rest.Interface
	mapper           meta.RESTMapper
	discoveryCache   discovery.CachedDiscoveryInterface
	discoveryLimiter *rate.Limiter
}

func NewClient(cfg *rest.Config, scheme *runtime.Scheme, cc ctrl.Client) (*Client, error) {
	discoveryCl, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("unable to construct a Discovery client: %w", err)
	}

	kubeCl, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("unable to construct a Kubernetes client: %w", err)
	}

	restCl, err := newRESTClientForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("unable to construct a REST client: %w", err)
	}

	dynCl, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("unable to construct a Dynamic client: %w", err)
	}

	daprCl, err := daprClient.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("unable to construct a Dapr client: %w", err)
	}

	apiextCl, err := apiextv1.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("unable to construct an API Extension client: %w", err)
	}

	c := Client{
		Client:                   cc,
		Interface:                kubeCl,
		ApiextensionsV1Interface: apiextCl,
		Discovery:                discoveryCl,
		Dapr:                     daprCl,
		dynamic:                  dynCl,
		scheme:                   scheme,
		config:                   cfg,
		rest:                     restCl,
	}

	c.discoveryLimiter = rate.NewLimiter(rate.Every(time.Second), DiscoveryLimiterBurst)
	c.discoveryCache = memory.NewMemCacheClient(discoveryCl)
	c.mapper = restmapper.NewDeferredDiscoveryRESTMapper(c.discoveryCache)

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

	rc, err := rest.RESTClientFor(cfg)
	if err != nil {
		return nil, fmt.Errorf("unable to construct a REST client: %w", err)
	}

	return rc, nil
}

func (c *Client) Dynamic(namespace string, obj *unstructured.Unstructured) (dynamic.ResourceInterface, error) {
	if c.discoveryLimiter.Allow() {
		c.discoveryCache.Invalidate()
	}

	c.discoveryCache.Fresh()

	mapping, err := c.mapper.RESTMapping(obj.GroupVersionKind().GroupKind(), obj.GroupVersionKind().Version)
	if err != nil {
		return nil, fmt.Errorf(
			"unable to identify preferred resource mapping for %s/%s: %w",
			obj.GroupVersionKind().GroupKind(),
			obj.GroupVersionKind().Version,
			err)
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

func (c *Client) Invalidate() {
	if c.discoveryCache != nil {
		c.discoveryCache.Invalidate()
	}
}

func (c *Client) ApplyStatus(ctx context.Context, obj ctrl.Object, opts ...ctrl.SubResourcePatchOption) error {
	u, err := resources.ToUnstructured(c.Scheme(), obj)
	if err != nil {
		return fmt.Errorf("unable to convert object %s to unstructured: %w", obj, err)
	}

	// Reset field not meaningful for patch
	delete(u.Object, "spec")

	u.SetResourceVersion("")
	u.SetManagedFields(nil)

	err = c.Client.Status().Patch(ctx, u, ctrl.Apply, opts...)
	if err != nil {
		return fmt.Errorf("unable to pactch object %s: %w", obj, err)
	}

	return nil
}
