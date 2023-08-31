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

// Code generated by informer-gen. DO NOT EDIT.

package v1beta1

import (
	"context"
	time "time"

	akov1beta1 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/apis/ako/v1beta1"
	versioned "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/client/v1beta1/clientset/versioned"
	internalinterfaces "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/client/v1beta1/informers/externalversions/internalinterfaces"
	v1beta1 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/client/v1beta1/listers/ako/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// AviInfraSettingInformer provides access to a shared informer and lister for
// AviInfraSettings.
type AviInfraSettingInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1beta1.AviInfraSettingLister
}

type aviInfraSettingInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
}

// NewAviInfraSettingInformer constructs a new informer for AviInfraSetting type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewAviInfraSettingInformer(client versioned.Interface, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredAviInfraSettingInformer(client, resyncPeriod, indexers, nil)
}

// NewFilteredAviInfraSettingInformer constructs a new informer for AviInfraSetting type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredAviInfraSettingInformer(client versioned.Interface, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.AkoV1beta1().AviInfraSettings().List(context.TODO(), options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.AkoV1beta1().AviInfraSettings().Watch(context.TODO(), options)
			},
		},
		&akov1beta1.AviInfraSetting{},
		resyncPeriod,
		indexers,
	)
}

func (f *aviInfraSettingInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredAviInfraSettingInformer(client, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *aviInfraSettingInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&akov1beta1.AviInfraSetting{}, f.defaultInformer)
}

func (f *aviInfraSettingInformer) Lister() v1beta1.AviInfraSettingLister {
	return v1beta1.NewAviInfraSettingLister(f.Informer().GetIndexer())
}
