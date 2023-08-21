package run

import (
	"github.com/spf13/cobra"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/dapr-sandbox/dapr-kubernetes-operator/pkg/controller"

	daprApi "github.com/dapr-sandbox/dapr-kubernetes-operator/api/operator/v1alpha1"
	daprCtl "github.com/dapr-sandbox/dapr-kubernetes-operator/internal/controller/operator"
	routev1 "github.com/openshift/api/route/v1"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
)

func init() {
	utilruntime.Must(daprApi.AddToScheme(controller.Scheme))
	utilruntime.Must(routev1.Install(controller.Scheme))
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
			return controller.Start(controllerOpts, func(manager manager.Manager, opts controller.Options) error {
				_, err := daprCtl.NewReconciler(cmd.Context(), manager, helmOpts)
				if err != nil {
					return err
				}

				return err
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
