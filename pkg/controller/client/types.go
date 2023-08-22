package client

import "k8s.io/client-go/dynamic"

type NamespacedResource struct {
	dynamic.ResourceInterface
}

type ClusteredResource struct {
	dynamic.ResourceInterface
}
