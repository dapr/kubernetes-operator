package client

import "k8s.io/client-go/dynamic"

type ResourceScope string

const (
	ResourceScopeNamespace ResourceScope = "namespace"
	ResourceScopeCluster   ResourceScope = "cluster"
)

type ResourceInterface interface {
	dynamic.ResourceInterface

	Scope() ResourceScope
}

type NamespacedResource struct {
	dynamic.ResourceInterface
}

func (r *NamespacedResource) Scope() ResourceScope {
	return ResourceScopeNamespace
}

type ClusteredResource struct {
	dynamic.ResourceInterface
}

func (r *ClusteredResource) Scope() ResourceScope {
	return ResourceScopeCluster
}
