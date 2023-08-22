package controller

type ClusterType string

const (
	ClusterTypeVanilla   ClusterType = "Vanilla"
	ClusterTypeOpenShift ClusterType = "OpenShift"
)

type Options struct {
	MetricsAddr                   string
	ProbeAddr                     string
	PprofAddr                     string
	LeaderElectionID              string
	LeaderElectionNamespace       string
	EnableLeaderElection          bool
	ReleaseLeaderElectionOnCancel bool
}
