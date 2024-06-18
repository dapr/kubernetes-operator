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

type DaprControlPlaneSpec struct {
	// +kubebuilder:validation:Optional
	Values *JSON `json:"values"`
}

type DaprControlPlaneStatus struct {
	Status `json:",inline"`
	Chart  *ChartMeta `json:"chart,omitempty"`
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
// +kubebuilder:resource:path=daprcontrolplanes,scope=Namespaced,shortName=dcp,categories=dapr
// +kubebuilder:deprecatedversion:warning="v1alpha1.DaprControlPlane is deprecated, please, use v1alpha1.DaprInstance instead"

type DaprControlPlane struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DaprControlPlaneSpec   `json:"spec,omitempty"`
	Status DaprControlPlaneStatus `json:"status,omitempty"`
}

func (in *DaprControlPlane) GetStatus() *Status {
	return &in.Status.Status
}

func (in *DaprControlPlane) GetConditions() conditions.Conditions {
	return in.Status.Conditions
}

// +kubebuilder:object:root=true

type DaprControlPlaneList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DaprControlPlane `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DaprControlPlane{}, &DaprControlPlaneList{})
}
