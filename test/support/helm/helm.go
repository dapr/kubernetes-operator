package helm

import (
	"os"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/registry"
)

type ReleaseOptions[T any] struct {
	Client *T
	Values map[string]interface{}
}

type ConfigurationOption func(*action.Configuration)

func WithLog(value func(string, ...interface{})) ConfigurationOption {
	return func(opt *action.Configuration) {
		opt.Log = value
	}
}

func New(options ...ConfigurationOption) (*Helm, error) {

	settings := cli.New()
	config := action.Configuration{}

	for _, option := range options {
		option(&config)
	}

	registryClient, err := registry.NewClient(
		registry.ClientOptDebug(settings.Debug),
		registry.ClientOptEnableCache(true),
		registry.ClientOptWriter(os.Stdout),
		registry.ClientOptCredentialsFile(settings.RegistryConfig),
	)
	if err != nil {
		return nil, err
	}

	config.RegistryClient = registryClient

	err = config.Init(settings.RESTClientGetter(), settings.Namespace(), "memory", config.Log)
	if err != nil {
		return nil, err
	}

	h := Helm{
		settings: settings,
		config:   &config,
	}

	return &h, nil
}

type Helm struct {
	settings *cli.EnvSettings
	config   *action.Configuration
}
