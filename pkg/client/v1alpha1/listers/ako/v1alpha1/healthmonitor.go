/*
Copyright The Kubernetes Authors.

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
	v1alpha1 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/apis/ako/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// HealthMonitorLister helps list HealthMonitors.
// All objects returned here must be treated as read-only.
type HealthMonitorLister interface {
	// List lists all HealthMonitors in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.HealthMonitor, err error)
	// HealthMonitors returns an object that can list and get HealthMonitors.
	HealthMonitors(namespace string) HealthMonitorNamespaceLister
	HealthMonitorListerExpansion
}

// healthMonitorLister implements the HealthMonitorLister interface.
type healthMonitorLister struct {
	indexer cache.Indexer
}

// NewHealthMonitorLister returns a new HealthMonitorLister.
func NewHealthMonitorLister(indexer cache.Indexer) HealthMonitorLister {
	return &healthMonitorLister{indexer: indexer}
}

// List lists all HealthMonitors in the indexer.
func (s *healthMonitorLister) List(selector labels.Selector) (ret []*v1alpha1.HealthMonitor, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.HealthMonitor))
	})
	return ret, err
}

// HealthMonitors returns an object that can list and get HealthMonitors.
func (s *healthMonitorLister) HealthMonitors(namespace string) HealthMonitorNamespaceLister {
	return healthMonitorNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// HealthMonitorNamespaceLister helps list and get HealthMonitors.
// All objects returned here must be treated as read-only.
type HealthMonitorNamespaceLister interface {
	// List lists all HealthMonitors in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.HealthMonitor, err error)
	// Get retrieves the HealthMonitor from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1alpha1.HealthMonitor, error)
	HealthMonitorNamespaceListerExpansion
}

// healthMonitorNamespaceLister implements the HealthMonitorNamespaceLister
// interface.
type healthMonitorNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all HealthMonitors in the indexer for a given namespace.
func (s healthMonitorNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.HealthMonitor, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.HealthMonitor))
	})
	return ret, err
}

// Get retrieves the HealthMonitor from the indexer for a given namespace and name.
func (s healthMonitorNamespaceLister) Get(name string) (*v1alpha1.HealthMonitor, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("healthmonitor"), name)
	}
	return obj.(*v1alpha1.HealthMonitor), nil
}
