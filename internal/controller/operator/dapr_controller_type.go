package operator

import (
	"context"

	daprApi "github.com/dapr-sandbox/dapr-kubernetes-operator/api/operator/v1alpha1"
	"github.com/dapr-sandbox/dapr-kubernetes-operator/pkg/controller"
	"github.com/dapr-sandbox/dapr-kubernetes-operator/pkg/controller/client"
	"helm.sh/helm/v3/pkg/chart"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/builder"
)

const (
	DaprReleaseGeneration = "controlplane.operator.dapr.io/release.generation"
	DaprReleaseName       = "controlplane.operator.dapr.io/release.name"
	DaprReleaseNamespace  = "controlplane.operator.dapr.io/release.namespace"

	DaprFinalizerName = "controlplane.operator.dapr.io/finalizer"
	DaprFieldManager  = "dapr-controlplane"

	HelmChartsDir = "helm-charts/dapr"

	DaprControlPlaneName             = "dapr-control-plane"
	DaprControlPlaneNamespaceDefault = "dapr-system"
	DaprControlPlaneNamespaceEnv     = "DAPR_CONTROL_PLANE_NAMESPACE"

	DaprConditionReconciled                     = "Reconcile"
	DaprConditionReady                          = "Ready"
	DaprPhaseError                              = "Error"
	DaprPhaseReady                              = "Ready"
	DaprConditionReasonUnsupportedConfiguration = "UnsupportedConfiguration"
)

type HelmOptions struct {
	ChartsDir string
}

type ReconciliationRequest struct {
	*client.Client
	types.NamespacedName

	Reconciler  *Reconciler
	ClusterType controller.ClusterType
	Resource    *daprApi.DaprControlPlane
	Chart       *chart.Chart
	Overrides   map[string]interface{}
}

type Action interface {
	Configure(context.Context, *client.Client, *builder.Builder) (*builder.Builder, error)
	Run(context.Context, *ReconciliationRequest) error
	Cleanup(context.Context, *ReconciliationRequest) error
}
