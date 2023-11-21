package controller

import (
	rtcache "sigs.k8s.io/controller-runtime/pkg/cache"
	rtclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type ClusterType string

const (
	ClusterTypeVanilla   ClusterType = "Vanilla"
	ClusterTypeOpenShift ClusterType = "OpenShift"

	FieldManager     = "dapr-kubernetes-controller"
	NamespaceDefault = "dapr-system"
	NamespaceEnv     = "DAPR_KUBERNETES_OPERATOR_NAMESPACE"
)

type Options struct {
	MetricsAddr                   string
	ProbeAddr                     string
	PprofAddr                     string
	LeaderElectionID              string
	LeaderElectionNamespace       string
	EnableLeaderElection          bool
	ReleaseLeaderElectionOnCancel bool
	WatchSelectors                map[rtclient.Object]rtcache.ByObject
}
