package run

import (
	"fmt"

	"github.com/spf13/cobra"
	admregv1 "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	rtcache "sigs.k8s.io/controller-runtime/pkg/cache"
	rtclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	daprApi "github.com/dapr-sandbox/dapr-kubernetes-operator/api/operator/v1alpha1"
	daprCtl "github.com/dapr-sandbox/dapr-kubernetes-operator/internal/controller/operator"
	"github.com/dapr-sandbox/dapr-kubernetes-operator/pkg/controller"
	"github.com/dapr-sandbox/dapr-kubernetes-operator/pkg/resources"
)

func init() {
	utilruntime.Must(daprApi.AddToScheme(controller.Scheme))
}

func NewRunCmd() *cobra.Command {

	controllerOpts := controller.Options{
		MetricsAddr:                   ":8080",
		ProbeAddr:                     ":8081",
		PprofAddr:                     "",
		LeaderElectionID:              "9aa9f118.dapr.io",
		EnableLeaderElection:          true,
		ReleaseLeaderElectionOnCancel: true,
		LeaderElectionNamespace:       "",
	}

	helmOpts := daprCtl.HelmOptions{
		ChartsDir: daprCtl.HelmChartsDir,
	}

	cmd := cobra.Command{
		Use:   "run",
		Short: "run",
		RunE: func(cmd *cobra.Command, args []string) error {
			selector, err := daprCtl.ReleaseSelector()
			if err != nil {
				return fmt.Errorf("unable to compute cache's watch selector: %w", err)
			}

			controllerOpts.WatchSelectors = map[rtclient.Object]rtcache.ByObject{
				// k8s
				&rbacv1.ClusterRole{}:                    {Label: selector},
				&rbacv1.ClusterRoleBinding{}:             {Label: selector},
				&rbacv1.Role{}:                           {Label: selector},
				&rbacv1.RoleBinding{}:                    {Label: selector},
				&admregv1.MutatingWebhookConfiguration{}: {Label: selector},
				&corev1.Secret{}:                         {Label: selector},
				&corev1.Service{}:                        {Label: selector},
				&corev1.ServiceAccount{}:                 {Label: selector},
				&appsv1.StatefulSet{}:                    {Label: selector},
				&appsv1.Deployment{}:                     {Label: selector},
				// dapr
				resources.UnstructuredFor("dapr.io", "v1alpha1", "Configuration"): {Label: selector},
			}

			return controller.Start(controllerOpts, func(manager manager.Manager, opts controller.Options) error {
				_, err := daprCtl.NewReconciler(cmd.Context(), manager, helmOpts)
				if err != nil {
					return fmt.Errorf("unable to set-up DaprControlPlane reconciler: %w", err)
				}

				return nil
			})
		},
	}

	cmd.Flags().StringVar(&controllerOpts.LeaderElectionID, "leader-election-id", controllerOpts.LeaderElectionID, "The leader election ID of the operator.")
	cmd.Flags().StringVar(&controllerOpts.LeaderElectionNamespace, "leader-election-namespace", controllerOpts.LeaderElectionNamespace, "The leader election namespace.")
	cmd.Flags().BoolVar(&controllerOpts.EnableLeaderElection, "leader-election", controllerOpts.EnableLeaderElection, "Enable leader election for controller manager.")
	cmd.Flags().BoolVar(&controllerOpts.ReleaseLeaderElectionOnCancel, "leader-election-release", controllerOpts.ReleaseLeaderElectionOnCancel, "If the leader should step down voluntarily.")

	cmd.Flags().StringVar(&controllerOpts.MetricsAddr, "metrics-bind-address", controllerOpts.MetricsAddr, "The address the metric endpoint binds to.")
	cmd.Flags().StringVar(&controllerOpts.ProbeAddr, "health-probe-bind-address", controllerOpts.ProbeAddr, "The address the probe endpoint binds to.")
	cmd.Flags().StringVar(&controllerOpts.PprofAddr, "pprof-bind-address", controllerOpts.PprofAddr, "The address the pprof endpoint binds to.")

	cmd.Flags().StringVar(&helmOpts.ChartsDir, "helm-charts-dir", helmOpts.ChartsDir, "Helm charts dir.")

	return &cmd
}
