/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package v1beta1

import (
	"encoding/json"
	"errors"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var ErrUnmarshalOnNil = errors.New("UnmarshalJSON on nil pointer")

// RawMessage is a raw encoded JSON value.
// It implements Marshaler and Unmarshaler and can
// be used to delay JSON decoding or precompute a JSON encoding.
// +kubebuilder:validation:Type=""
// +kubebuilder:validation:Format=""
// +kubebuilder:pruning:PreserveUnknownFields
type RawMessage []byte

// +kubebuilder:validation:Type=""
// JSON represents any valid JSON value.
// These types are supported: bool, int64, float64, string, []interface{}, map[string]interface{} and nil.
type JSON struct {
	RawMessage `json:",inline"`
}

// MarshalJSON returns m as the JSON encoding of m.
func (m RawMessage) MarshalJSON() ([]byte, error) {
	if m == nil {
		return []byte("null"), nil
	}

	return m, nil
}

// UnmarshalJSON sets *m to a copy of data.
func (m *RawMessage) UnmarshalJSON(data []byte) error {
	if m == nil {
		return fmt.Errorf("json.RawMessage: %w", ErrUnmarshalOnNil)
	}

	*m = append((*m)[0:0], data...)

	return nil
}

// String returns a string representation of RawMessage.
func (m *RawMessage) String() string {
	if m == nil {
		return ""
	}

	b, err := m.MarshalJSON()
	if err != nil {
		return ""
	}

	return string(b)
}

var _ json.Marshaler = (*RawMessage)(nil)
var _ json.Unmarshaler = (*RawMessage)(nil)

type ChartSpec struct {
	// +kubebuilder:default:="https://dapr.github.io/helm-charts"
	Repo string `json:"repo,omitempty"`

	// +kubebuilder:default:="dapr"
	Name string `json:"name,omitempty"`

	// +kubebuilder:validation:Optional
	Version string `json:"version,omitempty"`

	// +kubebuilder:validation:Optional
	Secret string `json:"secret,omitempty"`
}

type ChartMeta struct {
	Repo    string `json:"repo,omitempty"`
	Name    string `json:"name,omitempty"`
	Version string `json:"version,omitempty"`
}

type Status struct {
	Phase              string             `json:"phase"`
	Conditions         []metav1.Condition `json:"conditions,omitempty"`
	ObservedGeneration int64              `json:"observedGeneration,omitempty"`
}
