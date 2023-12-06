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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DaprCruiseControlSpec defines the desired state of DaprCruiseControl.
type DaprCruiseControlSpec struct {
}

// DaprCruiseControlStatus defines the observed state of DaprCruiseControl.
type DaprCruiseControlStatus struct {
	Phase              string             `json:"phase"`
	Conditions         []metav1.Condition `json:"conditions,omitempty"`
	ObservedGeneration int64              `json:"observedGeneration,omitempty"`
	Chart              *ChartMeta         `json:"chart,omitempty"`
}

// +genclient
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
// +kubebuilder:printcolumn:name="Ready",type=string,JSONPath=`.status.conditions[?(@.type=="Ready")].status`,description="Ready"
// +kubebuilder:printcolumn:name="Reason",type=string,JSONPath=`.status.conditions[?(@.type=="Ready")].reason`,description="Reason"
// +kubebuilder:printcolumn:name="Chart Name",type=string,JSONPath=`.status.chart.name`,description="Chart Name"
// +kubebuilder:printcolumn:name="Chart Repo",type=string,JSONPath=`.status.chart.repo`,description="Chart Repo"
// +kubebuilder:printcolumn:name="Chart Version",type=string,JSONPath=`.status.chart.version`,description="Chart Version"
// +kubebuilder:resource:path=daprcruiscontrols,scope=Namespaced,shortName=dcc,categories=dapr

// DaprCruiseControl is the Schema for the daprcruisecontrols API.
type DaprCruiseControl struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DaprCruiseControlSpec   `json:"spec,omitempty"`
	Status DaprCruiseControlStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// DaprCruiseControlList contains a list of DaprCruiseControl.
type DaprCruiseControlList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DaprCruiseControl `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DaprCruiseControl{}, &DaprCruiseControlList{})
}
