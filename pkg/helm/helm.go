package helm

import (
	"bytes"
	"encoding/json"
	"io"
	"sort"
	"strings"

	"github.com/dapr-sandbox/dapr-kubernetes-operator/pkg/utils/mergemap"

	daprApi "github.com/dapr-sandbox/dapr-kubernetes-operator/api/operator/v1alpha1"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/engine"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	k8syaml "k8s.io/apimachinery/pkg/runtime/serializer/yaml"
)

func NewEngine() *Engine {
	return &Engine{
		e:       engine.Engine{},
		decoder: k8syaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme),
	}
}

type Engine struct {
	e       engine.Engine
	decoder runtime.Serializer
}

func (e *Engine) Render(c *chart.Chart, dapr *daprApi.DaprControlPlane, overrides map[string]interface{}) ([]unstructured.Unstructured, error) {
	rv, err := e.renderValues(c, dapr, overrides)
	if err != nil {
		return nil, errors.Wrap(err, "cannot render values")
	}

	files, err := engine.Engine{}.Render(c, rv)
	if err != nil {
		return nil, errors.Wrap(err, "cannot render a chart")
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
		ul, err := e.decode([]byte(v))
		if err != nil {
			return nil, errors.Wrapf(err, "cannot decode %s", k)
		}
		if ul == nil {
			continue
		}

		result = append(result, ul...)
	}

	return result, nil
}

func (e *Engine) decode(content []byte) ([]unstructured.Unstructured, error) {
	results := make([]unstructured.Unstructured, 0)

	r := bytes.NewReader(content)
	decoder := yaml.NewDecoder(r)

	for {
		var out map[string]interface{}

		err := decoder.Decode(&out)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return nil, err
		}

		if len(out) == 0 {
			continue
		}
		if out["Kind"] == "" {
			continue
		}

		encoded, err := yaml.Marshal(out)
		if err != nil {
			return nil, err
		}

		var obj unstructured.Unstructured

		if _, _, err = e.decoder.Decode(encoded, nil, &obj); err != nil {
			if runtime.IsMissingKind(err) {
				continue
			}

			return nil, err
		}

		results = append(results, obj)

	}

	return results, nil
}

func (e *Engine) renderValues(c *chart.Chart, dapr *daprApi.DaprControlPlane, overrides map[string]interface{}) (chartutil.Values, error) {
	values := make(map[string]interface{})

	if dapr.Spec.Values != nil {
		if err := json.Unmarshal(dapr.Spec.Values.RawMessage, &values); err != nil {
			return chartutil.Values{}, errors.Wrap(err, "unable to decode values")
		}
	}

	values = mergemap.Merge(values, overrides)

	err := chartutil.ProcessDependencies(c, values)
	if err != nil {
		return chartutil.Values{}, errors.Wrap(err, "cannot process dependencies")
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
		return chartutil.Values{}, errors.Wrap(err, "cannot render values")
	}

	return rv, nil

}
