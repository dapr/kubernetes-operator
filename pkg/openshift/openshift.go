package openshift

import (
	"fmt"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/discovery"
)

// IsOpenShift returns true if we are connected to a OpenShift cluster.
func IsOpenShift(client discovery.DiscoveryInterface) (bool, error) {
	if client == nil {
		return false, nil
	}

	_, err := client.ServerResourcesForGroupVersion("route.openshift.io/v1")
	if err != nil && k8serrors.IsNotFound(err) {
		return false, nil
	}

	if err != nil {
		return false, fmt.Errorf("unable to determine if the Kubernetes distro is OpenShift: %w", err)
	}

	return true, nil
}
