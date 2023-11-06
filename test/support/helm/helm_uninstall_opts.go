package helm

import (
	"time"

	"helm.sh/helm/v3/pkg/action"
)

func WithUninstallTimeout(value time.Duration) UninstallOption {
	return func(install *ReleaseOptions[action.Uninstall]) {
		install.Client.Timeout = value
	}
}

func WithUninstallWait(value bool) UninstallOption {
	return func(install *ReleaseOptions[action.Uninstall]) {
		install.Client.Wait = value
	}
}
