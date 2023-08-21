package helm

import (
	"time"

	"github.com/dapr-sandbox/dapr-kubernetes-operator/pkg/utils/mergemap"
	"helm.sh/helm/v3/pkg/action"
)

func WithInstallName(value string) InstallOption {
	return func(install *ReleaseOptions[action.Install]) {
		install.Client.ReleaseName = value
	}
}

func WithInstallNamespace(value string) InstallOption {
	return func(install *ReleaseOptions[action.Install]) {
		install.Client.Namespace = value
	}
}

func WithInstallValue(name string, value interface{}) InstallOption {
	return func(install *ReleaseOptions[action.Install]) {
		install.Values[name] = value
	}
}

func WithInstallVersion(value string) InstallOption {
	return func(install *ReleaseOptions[action.Install]) {
		install.Client.Version = value
	}
}

func WithInstallValues(values map[string]interface{}) InstallOption {
	return func(install *ReleaseOptions[action.Install]) {
		install.Values = mergemap.Merge(install.Values, values)
	}
}

func WithInstallTimeout(value time.Duration) InstallOption {
	return func(install *ReleaseOptions[action.Install]) {
		install.Client.Timeout = value
	}
}
