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
// Code generated by applyconfiguration-gen. DO NOT EDIT.

package v1alpha1

import (
	v1 "k8s.io/client-go/applyconfigurations/meta/v1"
)

// DaprCruiseControlStatusApplyConfiguration represents a declarative configuration of the DaprCruiseControlStatus type for use
// with apply.
type DaprCruiseControlStatusApplyConfiguration struct {
	StatusApplyConfiguration `json:",inline"`
	Chart                    *ChartMetaApplyConfiguration `json:"chart,omitempty"`
}

// DaprCruiseControlStatusApplyConfiguration constructs a declarative configuration of the DaprCruiseControlStatus type for use with
// apply.
func DaprCruiseControlStatus() *DaprCruiseControlStatusApplyConfiguration {
	return &DaprCruiseControlStatusApplyConfiguration{}
}

// WithPhase sets the Phase field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Phase field is set to the value of the last call.
func (b *DaprCruiseControlStatusApplyConfiguration) WithPhase(value string) *DaprCruiseControlStatusApplyConfiguration {
	b.StatusApplyConfiguration.Phase = &value
	return b
}

// WithConditions adds the given value to the Conditions field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the Conditions field.
func (b *DaprCruiseControlStatusApplyConfiguration) WithConditions(values ...*v1.ConditionApplyConfiguration) *DaprCruiseControlStatusApplyConfiguration {
	for i := range values {
		if values[i] == nil {
			panic("nil value passed to WithConditions")
		}
		b.StatusApplyConfiguration.Conditions = append(b.StatusApplyConfiguration.Conditions, *values[i])
	}
	return b
}

// WithObservedGeneration sets the ObservedGeneration field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the ObservedGeneration field is set to the value of the last call.
func (b *DaprCruiseControlStatusApplyConfiguration) WithObservedGeneration(value int64) *DaprCruiseControlStatusApplyConfiguration {
	b.StatusApplyConfiguration.ObservedGeneration = &value
	return b
}

// WithChart sets the Chart field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Chart field is set to the value of the last call.
func (b *DaprCruiseControlStatusApplyConfiguration) WithChart(value *ChartMetaApplyConfiguration) *DaprCruiseControlStatusApplyConfiguration {
	b.Chart = value
	return b
}
