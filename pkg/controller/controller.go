package controller

import (
	"fmt"
	"net/http"
	"net/http/pprof"
	"time"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	"github.com/dapr-sandbox/dapr-kubernetes-operator/pkg/logger"
)

var (
	Scheme = runtime.NewScheme()
	Log    = ctrl.Log.WithName("controller")
)

const (
	PprofReadTimeout  = 10 * time.Second
	PprofWriteTimeout = 10 * time.Second
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(Scheme))
}

func Start(options Options, setup func(manager.Manager, Options) error) error {
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&logger.Options)))

	ctx := ctrl.SetupSignalHandler()

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                        Scheme,
		HealthProbeBindAddress:        options.ProbeAddr,
		LeaderElection:                options.EnableLeaderElection,
		LeaderElectionID:              options.LeaderElectionID,
		LeaderElectionReleaseOnCancel: options.ReleaseLeaderElectionOnCancel,
		LeaderElectionNamespace:       options.LeaderElectionNamespace,

		Metrics: metricsserver.Options{
			BindAddress: options.MetricsAddr,
		},
		Cache: cache.Options{
			ByObject: options.WatchSelectors,
		},
	})

	if err != nil {
		return fmt.Errorf("unable to create manager: %w", err)
	}

	if err := setup(mgr, options); err != nil {
		return fmt.Errorf("unable to set up controllers: %w", err)
	}

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		return fmt.Errorf("unable to set up health check: %w", err)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		return fmt.Errorf("unable to set up readiness check: %w", err)
	}

	if options.PprofAddr != "" {
		mux := http.NewServeMux()
		mux.HandleFunc("/debug/pprof/", pprof.Index)
		mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

		server := &http.Server{
			Addr:         options.PprofAddr,
			ReadTimeout:  PprofReadTimeout,
			WriteTimeout: PprofWriteTimeout,
			Handler:      mux,
		}

		Log.Info("starting pprof")

		go func() {
			err := server.ListenAndServe()
			Log.Error(err, "pprof")
		}()
	}

	Log.Info("starting manager")

	if err := mgr.Start(ctx); err != nil {
		return fmt.Errorf("problem running manager: %w", err)
	}

	return nil
}
