package gc

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"

	"golang.org/x/exp/maps"

	"github.com/go-logr/logr"
	"golang.org/x/time/rate"
	authorization "k8s.io/api/authorization/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlCli "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/dapr/kubernetes-operator/pkg/controller/client"
	"github.com/dapr/kubernetes-operator/pkg/resources"
)

func New() *GC {
	return &GC{
		l:               ctrl.Log.WithName("gc"),
		limiter:         rate.NewLimiter(rate.Every(time.Minute), 1),
		collectableGVKs: make([]schema.GroupVersionKind, 0),
	}
}

type GC struct {
	l               logr.Logger
	lock            sync.Mutex
	limiter         *rate.Limiter
	collectableGVKs []schema.GroupVersionKind
}

func (gc *GC) Run(
	ctx context.Context,
	c *client.Client,
	ns string,
	selector labels.Selector,
	predicate func(context.Context, unstructured.Unstructured) (bool, error),
) (int, error) {
	gc.lock.Lock()
	defer gc.lock.Unlock()

	err := gc.computeDeletableTypes(ctx, c, ns)
	if err != nil {
		return 0, fmt.Errorf("cannot discover GVK types: %w", err)
	}

	return gc.deleteEachOf(ctx, c, selector, predicate)
}

func (gc *GC) deleteEachOf(
	ctx context.Context,
	c *client.Client,
	selector labels.Selector,
	predicate func(context.Context, unstructured.Unstructured) (bool, error),
) (int, error) {
	deleted := 0

	for _, GVK := range gc.collectableGVKs {
		items := unstructured.UnstructuredList{
			Object: map[string]interface{}{
				"apiVersion": GVK.GroupVersion().String(),
				"kind":       GVK.Kind,
			},
		}
		options := []ctrlCli.ListOption{
			ctrlCli.MatchingLabelsSelector{Selector: selector},
		}

		if err := c.List(ctx, &items, options...); err != nil {
			if k8serrors.IsForbidden(err) {
				gc.l.Info("cannot gc, forbidden", "gvks", GVK.String())

				continue
			}

			if !k8serrors.IsNotFound(err) {
				return 0, fmt.Errorf("cannot list child resources %s: %w", GVK.String(), err)
			}

			continue
		}

		for i := range items.Items {
			ok, err := gc.delete(ctx, c, items.Items[i], predicate)
			if err != nil {
				return 0, err
			}

			if ok {
				deleted++
			}
		}
	}

	return deleted, nil
}

func (gc *GC) delete(
	ctx context.Context,
	c *client.Client,
	resource unstructured.Unstructured,
	predicate func(context.Context, unstructured.Unstructured) (bool, error),
) (bool, error) {
	if !gc.canBeDeleted(ctx, resource.GroupVersionKind()) {
		return false, nil
	}

	canBeDeleted, err := predicate(ctx, resource)
	if err != nil {
		return false, err
	}

	if !canBeDeleted {
		return false, err
	}

	gc.l.Info("deleting", "ref", resources.Ref(&resource))

	err = c.Delete(ctx, &resource, ctrlCli.PropagationPolicy(metav1.DeletePropagationForeground))
	if err != nil {
		// The resource may have already been deleted
		if k8serrors.IsNotFound(err) {
			return true, nil
		}

		return false, fmt.Errorf(
			"cannot delete resources gvks:%s, namespace: %s, name: %s, err: %w",
			resource.GroupVersionKind().String(),
			resource.GetNamespace(),
			resource.GetName(),
			err,
		)
	}

	gc.l.Info("deleted", "ref", resources.Ref(&resource))

	return true, nil
}

func (gc *GC) canBeDeleted(_ context.Context, gvk schema.GroupVersionKind) bool {
	if gvk.Group == "coordination.k8s.io" && gvk.Kind == "Lease" {
		return false
	}

	return true
}

// computeDeletableTypes is a complex function, leave it as it is
//
//nolint:cyclop,gocognit,funlen
func (gc *GC) computeDeletableTypes(ctx context.Context, c *client.Client, ns string) error {
	// Rate limit to avoid Discovery and SelfSubjectRulesReview requests at every reconciliation.
	if !gc.limiter.Allow() {
		// Return the cached set of garbage collectable GVKs.
		return nil
	}

	// We rely on the discovery API to retrieve all the resources GVK,
	// that results in an unbounded set that can impact garbage collection latency when scaling up.
	items, err := c.Discovery.ServerPreferredNamespacedResources()

	// Swallow group discovery errors, e.g., Knative serving exposes
	// an aggregated API for custom.metrics.k8s.io that requires special
	// authentication scheme while discovering preferred resources.
	if err != nil && !discovery.IsGroupDiscoveryFailedError(err) {
		return fmt.Errorf("failure retireving supported namespaced resources: %w", err)
	}

	// We only take types that support the "delete" verb,
	// to prevents from performing queries that we know are going to return "MethodNotAllowed".
	apiResourceLists := discovery.FilteredBy(discovery.SupportsAllVerbs{Verbs: []string{"delete"}}, items)

	// Retrieve the permissions granted to the operator service account.
	// We assume the operator has only to garbage collect the resources it has created.
	ssrr := &authorization.SelfSubjectRulesReview{
		Spec: authorization.SelfSubjectRulesReviewSpec{
			Namespace: ns,
		},
	}

	ssrr, err = c.AuthorizationV1().SelfSubjectRulesReviews().Create(ctx, ssrr, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("unable to create SelfSubjectRulesReviews: %w", err)
	}

	GVKs := make(map[schema.GroupVersionKind]struct{})

	for _, res := range apiResourceLists {
		for i := range res.APIResources {
			resourceGroup := res.APIResources[i].Group

			if resourceGroup == "" {
				// Empty implies the group of the containing resource list should be used
				gv, err := schema.ParseGroupVersion(res.GroupVersion)
				if err != nil {
					return fmt.Errorf("failure creating SelfSubjectRulesReview: %w", err)
				}

				resourceGroup = gv.Group
			}

		rule:
			for _, rule := range ssrr.Status.ResourceRules {
				if !slices.Contains(rule.Verbs, "delete") && !slices.Contains(rule.Verbs, "*") {
					continue
				}

				for _, ruleGroup := range rule.APIGroups {
					for _, ruleResource := range rule.Resources {
						if (resourceGroup == ruleGroup || ruleGroup == "*") && (res.APIResources[i].Name == ruleResource || ruleResource == "*") {
							GVK := schema.FromAPIVersionAndKind(res.GroupVersion, res.APIResources[i].Kind)
							if gc.canBeDeleted(ctx, GVK) {
								GVKs[GVK] = struct{}{}
							}

							break rule
						}
					}
				}
			}
		}
	}

	gc.collectableGVKs = maps.Keys(GVKs)
	slices.SortFunc(gc.collectableGVKs, func(a, b schema.GroupVersionKind) int {
		return strings.Compare(a.String(), b.String())
	})

	return nil
}
