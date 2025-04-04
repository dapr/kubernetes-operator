package run

import (
	"fmt"

	"github.com/spf13/cobra"

	admregv1 "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	ctrlRt "sigs.k8s.io/controller-runtime"
	rtcache "sigs.k8s.io/controller-runtime/pkg/cache"
	rtclient "sigs.k8s.io/controller-runtime/pkg/client"

	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	daprApi "github.com/dapr/kubernetes-operator/api/operator/v1beta1"
	"github.com/dapr/kubernetes-operator/internal/controller/operator/instance"
	"github.com/dapr/kubernetes-operator/pkg/controller"
	"github.com/dapr/kubernetes-operator/pkg/helm"
)

const (
	cmdName = "run"
)

func NewCmd() *cobra.Command {
	co := controller.Options{
		MetricsAddr:                   ":8080",
		ProbeAddr:                     ":8081",
		PprofAddr:                     "",
		LeaderElectionID:              "9aa9f118.dapr.io",
		EnableLeaderElection:          true,
		ReleaseLeaderElectionOnCancel: true,
		LeaderElectionNamespace:       "",
	}

	helmOpts := helm.Options{
		ChartsDir: helm.ChartsDir,
	}

	cmd := cobra.Command{
		Use:   cmdName,
		Short: cmdName,
		RunE: func(cmd *cobra.Command, args []string) error {
			selector, err := computeListWatch()
			if err != nil {
				return fmt.Errorf("unable to compute cache's ListWatchr: %w", err)
			}

			co.WatchSelectors = selector
			co.Logger = ctrlRt.Log.WithName("controller")

			return controller.Start(co, func(manager manager.Manager, opts controller.Options) error {
				if _, err := instance.NewReconciler(cmd.Context(), manager, helmOpts); err != nil {
					return fmt.Errorf("unable to set-up DaprInstance reconciler: %w", err)
				}

				return nil
			})
		},
	}

	cmd.Flags().StringVar(
		&co.LeaderElectionID, "leader-election-id", co.LeaderElectionID, "The leader election ID of the operator.")
	cmd.Flags().StringVar(
		&co.LeaderElectionNamespace, "leader-election-namespace", co.LeaderElectionNamespace, "The leader election namespace.")
	cmd.Flags().BoolVar(
		&co.EnableLeaderElection, "leader-election", co.EnableLeaderElection, "Enable leader election for controller manager.")
	cmd.Flags().BoolVar(
		&co.ReleaseLeaderElectionOnCancel, "leader-election-release", co.ReleaseLeaderElectionOnCancel, "If the leader should step down voluntarily.")

	cmd.Flags().StringVar(
		&co.MetricsAddr, "metrics-bind-address", co.MetricsAddr, "The address the metric endpoint binds to.")
	cmd.Flags().StringVar(
		&co.ProbeAddr, "health-probe-bind-address", co.ProbeAddr, "The address the probe endpoint binds to.")
	cmd.Flags().StringVar(
		&co.PprofAddr, "pprof-bind-address", co.PprofAddr, "The address the pprof endpoint binds to.")

	cmd.Flags().StringVar(
		&helmOpts.ChartsDir, "helm-charts-dir", helmOpts.ChartsDir, "Helm charts dir.")

	return &cmd
}

//nolint:gochecknoinits
func init() {
	utilruntime.Must(daprApi.AddToScheme(controller.Scheme))
	utilruntime.Must(apiextensions.AddToScheme(controller.Scheme))
}

// computeListWatch computes the cache's ListWatch by object.
func computeListWatch() (map[rtclient.Object]rtcache.ByObject, error) {
	selector, err := helm.ReleaseSelector()
	if err != nil {
		return nil, fmt.Errorf("unable to compute cache's watch selector: %w", err)
	}

	selectors := map[rtclient.Object]rtcache.ByObject{
		&admregv1.MutatingWebhookConfiguration{}: {Label: selector},
		&appsv1.StatefulSet{}:                    {Label: selector},
		&appsv1.Deployment{}:                     {Label: selector},
		&corev1.Secret{}:                         {Label: selector},
		&corev1.ConfigMap{}:                      {Label: selector},
		&corev1.Service{}:                        {Label: selector},
		&corev1.ServiceAccount{}:                 {Label: selector},
		&policyv1.PodDisruptionBudget{}:          {Label: selector},
		&rbacv1.ClusterRole{}:                    {Label: selector},
		&rbacv1.ClusterRoleBinding{}:             {Label: selector},
		&rbacv1.Role{}:                           {Label: selector},
		&rbacv1.RoleBinding{}:                    {Label: selector},
	}

	return selectors, nil
}
