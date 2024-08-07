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
// Code generated by lister-gen. DO NOT EDIT.

package v1alpha1

import (
	v1alpha1 "github.com/dapr/kubernetes-operator/api/operator/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// DaprCruiseControlLister helps list DaprCruiseControls.
// All objects returned here must be treated as read-only.
type DaprCruiseControlLister interface {
	// List lists all DaprCruiseControls in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.DaprCruiseControl, err error)
	// DaprCruiseControls returns an object that can list and get DaprCruiseControls.
	DaprCruiseControls(namespace string) DaprCruiseControlNamespaceLister
	DaprCruiseControlListerExpansion
}

// daprCruiseControlLister implements the DaprCruiseControlLister interface.
type daprCruiseControlLister struct {
	indexer cache.Indexer
}

// NewDaprCruiseControlLister returns a new DaprCruiseControlLister.
func NewDaprCruiseControlLister(indexer cache.Indexer) DaprCruiseControlLister {
	return &daprCruiseControlLister{indexer: indexer}
}

// List lists all DaprCruiseControls in the indexer.
func (s *daprCruiseControlLister) List(selector labels.Selector) (ret []*v1alpha1.DaprCruiseControl, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.DaprCruiseControl))
	})
	return ret, err
}

// DaprCruiseControls returns an object that can list and get DaprCruiseControls.
func (s *daprCruiseControlLister) DaprCruiseControls(namespace string) DaprCruiseControlNamespaceLister {
	return daprCruiseControlNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// DaprCruiseControlNamespaceLister helps list and get DaprCruiseControls.
// All objects returned here must be treated as read-only.
type DaprCruiseControlNamespaceLister interface {
	// List lists all DaprCruiseControls in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.DaprCruiseControl, err error)
	// Get retrieves the DaprCruiseControl from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1alpha1.DaprCruiseControl, error)
	DaprCruiseControlNamespaceListerExpansion
}

// daprCruiseControlNamespaceLister implements the DaprCruiseControlNamespaceLister
// interface.
type daprCruiseControlNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all DaprCruiseControls in the indexer for a given namespace.
func (s daprCruiseControlNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.DaprCruiseControl, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.DaprCruiseControl))
	})
	return ret, err
}

// Get retrieves the DaprCruiseControl from the indexer for a given namespace and name.
func (s daprCruiseControlNamespaceLister) Get(name string) (*v1alpha1.DaprCruiseControl, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("daprcruisecontrol"), name)
	}
	return obj.(*v1alpha1.DaprCruiseControl), nil
}
