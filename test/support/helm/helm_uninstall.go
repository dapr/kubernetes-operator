package helm

import (
	"context"
	"fmt"

	"helm.sh/helm/v3/pkg/action"
)

type UninstallOption func(*ReleaseOptions[action.Uninstall])

func (h *Helm) Uninstall(_ context.Context, name string, options ...UninstallOption) error {
	client := action.NewUninstall(h.config)
	client.DeletionPropagation = "foreground"
	client.KeepHistory = false

	io := ReleaseOptions[action.Uninstall]{
		Client: client,
		Values: make(map[string]interface{}),
	}

	for _, option := range options {
		option(&io)
	}

	_, err := client.Run(name)
	if err != nil {
		return fmt.Errorf("unabele to uninstall release %s: %w", name, err)
	}

	return nil
}
