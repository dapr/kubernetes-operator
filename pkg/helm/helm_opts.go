package helm

import "helm.sh/helm/v3/pkg/action"

type ChartOptions struct {
	action.ChartPathOptions
	Name string
}

type ChartOption func(*ChartOptions)

func WithName(value string) ChartOption {
	return func(opts *ChartOptions) {
		opts.Name = value
	}
}

func WithVersion(value string) ChartOption {
	return func(opts *ChartOptions) {
		opts.Version = value
	}
}

func WithRepoURL(value string) ChartOption {
	return func(opts *ChartOptions) {
		opts.RepoURL = value
	}
}

func WithUsername(value string) ChartOption {
	return func(opts *ChartOptions) {
		opts.Username = value
	}
}

func WithPassword(value string) ChartOption {
	return func(opts *ChartOptions) {
		opts.Password = value
	}
}
