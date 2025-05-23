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

package v1beta1

import (
	labels "k8s.io/apimachinery/pkg/labels"
	listers "k8s.io/client-go/listers"
	cache "k8s.io/client-go/tools/cache"
	apisv1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"
)

// ReferenceGrantLister helps list ReferenceGrants.
// All objects returned here must be treated as read-only.
type ReferenceGrantLister interface {
	// List lists all ReferenceGrants in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*apisv1beta1.ReferenceGrant, err error)
	// ReferenceGrants returns an object that can list and get ReferenceGrants.
	ReferenceGrants(namespace string) ReferenceGrantNamespaceLister
	ReferenceGrantListerExpansion
}

// referenceGrantLister implements the ReferenceGrantLister interface.
type referenceGrantLister struct {
	listers.ResourceIndexer[*apisv1beta1.ReferenceGrant]
}

// NewReferenceGrantLister returns a new ReferenceGrantLister.
func NewReferenceGrantLister(indexer cache.Indexer) ReferenceGrantLister {
	return &referenceGrantLister{listers.New[*apisv1beta1.ReferenceGrant](indexer, apisv1beta1.Resource("referencegrant"))}
}

// ReferenceGrants returns an object that can list and get ReferenceGrants.
func (s *referenceGrantLister) ReferenceGrants(namespace string) ReferenceGrantNamespaceLister {
	return referenceGrantNamespaceLister{listers.NewNamespaced[*apisv1beta1.ReferenceGrant](s.ResourceIndexer, namespace)}
}

// ReferenceGrantNamespaceLister helps list and get ReferenceGrants.
// All objects returned here must be treated as read-only.
type ReferenceGrantNamespaceLister interface {
	// List lists all ReferenceGrants in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*apisv1beta1.ReferenceGrant, err error)
	// Get retrieves the ReferenceGrant from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*apisv1beta1.ReferenceGrant, error)
	ReferenceGrantNamespaceListerExpansion
}

// referenceGrantNamespaceLister implements the ReferenceGrantNamespaceLister
// interface.
type referenceGrantNamespaceLister struct {
	listers.ResourceIndexer[*apisv1beta1.ReferenceGrant]
}
