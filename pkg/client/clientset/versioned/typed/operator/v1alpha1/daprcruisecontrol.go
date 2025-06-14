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
// Code generated by client-gen. DO NOT EDIT.

package v1alpha1

import (
	context "context"

	operatorv1alpha1 "github.com/dapr/kubernetes-operator/api/operator/v1alpha1"
	applyconfigurationoperatorv1alpha1 "github.com/dapr/kubernetes-operator/pkg/client/applyconfiguration/operator/v1alpha1"
	scheme "github.com/dapr/kubernetes-operator/pkg/client/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	gentype "k8s.io/client-go/gentype"
)

// DaprCruiseControlsGetter has a method to return a DaprCruiseControlInterface.
// A group's client should implement this interface.
type DaprCruiseControlsGetter interface {
	DaprCruiseControls(namespace string) DaprCruiseControlInterface
}

// DaprCruiseControlInterface has methods to work with DaprCruiseControl resources.
type DaprCruiseControlInterface interface {
	Create(ctx context.Context, daprCruiseControl *operatorv1alpha1.DaprCruiseControl, opts v1.CreateOptions) (*operatorv1alpha1.DaprCruiseControl, error)
	Update(ctx context.Context, daprCruiseControl *operatorv1alpha1.DaprCruiseControl, opts v1.UpdateOptions) (*operatorv1alpha1.DaprCruiseControl, error)
	// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
	UpdateStatus(ctx context.Context, daprCruiseControl *operatorv1alpha1.DaprCruiseControl, opts v1.UpdateOptions) (*operatorv1alpha1.DaprCruiseControl, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*operatorv1alpha1.DaprCruiseControl, error)
	List(ctx context.Context, opts v1.ListOptions) (*operatorv1alpha1.DaprCruiseControlList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *operatorv1alpha1.DaprCruiseControl, err error)
	Apply(ctx context.Context, daprCruiseControl *applyconfigurationoperatorv1alpha1.DaprCruiseControlApplyConfiguration, opts v1.ApplyOptions) (result *operatorv1alpha1.DaprCruiseControl, err error)
	// Add a +genclient:noStatus comment above the type to avoid generating ApplyStatus().
	ApplyStatus(ctx context.Context, daprCruiseControl *applyconfigurationoperatorv1alpha1.DaprCruiseControlApplyConfiguration, opts v1.ApplyOptions) (result *operatorv1alpha1.DaprCruiseControl, err error)
	DaprCruiseControlExpansion
}

// daprCruiseControls implements DaprCruiseControlInterface
type daprCruiseControls struct {
	*gentype.ClientWithListAndApply[*operatorv1alpha1.DaprCruiseControl, *operatorv1alpha1.DaprCruiseControlList, *applyconfigurationoperatorv1alpha1.DaprCruiseControlApplyConfiguration]
}

// newDaprCruiseControls returns a DaprCruiseControls
func newDaprCruiseControls(c *OperatorV1alpha1Client, namespace string) *daprCruiseControls {
	return &daprCruiseControls{
		gentype.NewClientWithListAndApply[*operatorv1alpha1.DaprCruiseControl, *operatorv1alpha1.DaprCruiseControlList, *applyconfigurationoperatorv1alpha1.DaprCruiseControlApplyConfiguration](
			"daprcruisecontrols",
			c.RESTClient(),
			scheme.ParameterCodec,
			namespace,
			func() *operatorv1alpha1.DaprCruiseControl { return &operatorv1alpha1.DaprCruiseControl{} },
			func() *operatorv1alpha1.DaprCruiseControlList { return &operatorv1alpha1.DaprCruiseControlList{} },
		),
	}
}
