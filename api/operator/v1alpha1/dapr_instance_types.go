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
	"github.com/dapr/kubernetes-operator/pkg/conditions"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DaprInstanceSpec defines the desired state of DaprInstance.
type DaprInstanceSpec struct {
	// +kubebuilder:validation:Optional
	Chart *ChartSpec `json:"chart,omitempty"`

	// +kubebuilder:validation:Optional
	Values *JSON `json:"values"`
}

// DaprInstanceStatus defines the observed state of DaprInstance.
type DaprInstanceStatus struct {
	Status `json:",inline"`

	Chart *ChartMeta `json:"chart,omitempty"`
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
// +kubebuilder:resource:path=daprinstances,scope=Namespaced,shortName=di,categories=dapr

// DaprInstance is the Schema for the daprinstances API.
type DaprInstance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DaprInstanceSpec   `json:"spec,omitempty"`
	Status DaprInstanceStatus `json:"status,omitempty"`
}

func (in *DaprInstance) GetStatus() *Status {
	return &in.Status.Status
}

func (in *DaprInstance) GetConditions() conditions.Conditions {
	return in.Status.Conditions
}

// +kubebuilder:object:root=true

// DaprInstanceList contains a list of DaprInstance.
type DaprInstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []DaprInstance `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DaprInstance{}, &DaprInstanceList{})
}
