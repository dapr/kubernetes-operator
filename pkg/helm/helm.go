package helm

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/dapr-sandbox/dapr-kubernetes-operator/pkg/resources"

	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"

	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/engine"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	k8syaml "k8s.io/apimachinery/pkg/runtime/serializer/yaml"

	daprApi "github.com/dapr-sandbox/dapr-kubernetes-operator/api/operator/v1alpha1"
	"github.com/dapr-sandbox/dapr-kubernetes-operator/pkg/utils/maputils"
)

const (
	ReleaseGeneration = "helm.operator.dapr.io/release.generation"
	ReleaseName       = "helm.operator.dapr.io/release.name"
	ReleaseNamespace  = "helm.operator.dapr.io/release.namespace"

	ChartsDir = "helm-charts/dapr"
)

type ValuesCustomizer func(map[string]any) (map[string]any, error)

type Options struct {
	ChartsDir string
}

func NewEngine() *Engine {
	return &Engine{
		e:                 engine.Engine{},
		env:               cli.New(),
		decoder:           k8syaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme),
		valuesCustomizers: make([]ValuesCustomizer, 0),
	}
}

type Engine struct {
	e                 engine.Engine
	env               *cli.EnvSettings
	decoder           runtime.Serializer
	valuesCustomizers []ValuesCustomizer
}

func (e *Engine) Customizer(customizer ValuesCustomizer, customizers ...ValuesCustomizer) {
	e.valuesCustomizers = append(e.valuesCustomizers, customizer)
	e.valuesCustomizers = append(e.valuesCustomizers, customizers...)
}

func (e *Engine) Load(options ChartOptions) (*chart.Chart, error) {
	path, err := options.LocateChart(options.Name, e.env)
	if err != nil {
		return nil, fmt.Errorf("unable to load chart (repo: %s, name: %s, version: %s), reson: %w", options.RepoURL, options.Name, options.Version, err)
	}

	return loader.Load(path)
}

func (e *Engine) Render(c *chart.Chart, dapr *daprApi.DaprInstance, overrides map[string]interface{}) ([]unstructured.Unstructured, error) {
	rv, err := e.renderValues(c, dapr, overrides)
	if err != nil {
		return nil, fmt.Errorf("cannot render values: %w", err)
	}

	files, err := e.e.Render(c, rv)
	if err != nil {
		return nil, fmt.Errorf("cannot render a chart: %w", err)
	}

	keys := make([]string, 0, len(files))

	for k := range files {
		if !strings.HasSuffix(k, ".yaml") && !strings.HasSuffix(k, ".yml") {
			continue
		}

		keys = append(keys, k)
	}

	sort.Strings(keys)

	result := make([]unstructured.Unstructured, 0)

	for _, k := range keys {
		v := files[k]

		ul, err := resources.Decode(e.decoder, []byte(v))
		if err != nil {
			return nil, fmt.Errorf("cannot decode %s: %w", k, err)
		}

		if ul == nil {
			continue
		}

		result = append(result, ul...)
	}

	return result, nil
}

func (e *Engine) renderValues(
	c *chart.Chart,
	dapr *daprApi.DaprInstance,
	overrides map[string]interface{},
) (chartutil.Values, error) {
	values := make(map[string]interface{})

	if dapr.Spec.Values != nil {
		if err := json.Unmarshal(dapr.Spec.Values.RawMessage, &values); err != nil {
			return chartutil.Values{}, fmt.Errorf("unable to decode values: %w", err)
		}
	}

	for i := range e.valuesCustomizers {
		nv, err := e.valuesCustomizers[i](values)
		if err != nil {
			return chartutil.Values{}, fmt.Errorf("unable to cusomize values: %w", err)
		}

		values = nv
	}

	values = maputils.Merge(values, overrides)

	err := chartutil.ProcessDependencies(c, values)
	if err != nil {
		return chartutil.Values{}, fmt.Errorf("cannot process dependencies: %w", err)
	}

	rv, err := chartutil.ToRenderValues(
		c,
		values,
		chartutil.ReleaseOptions{
			Name:      dapr.Name,
			Namespace: dapr.Namespace,
			Revision:  int(dapr.Generation),
			IsInstall: false,
			IsUpgrade: true,
		},
		nil)

	if err != nil {
		return chartutil.Values{}, fmt.Errorf("cannot render values: %w", err)
	}

	return rv, nil
}
