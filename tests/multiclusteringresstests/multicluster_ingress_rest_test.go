/*
 * Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
 * All Rights Reserved.
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
*   http://www.apache.org/licenses/LICENSE-2.0
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
*/

package multiclusteringresstests

import (
	"context"
	"testing"
	"time"

	"github.com/onsi/gomega"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	avinodes "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/integrationtest"

	utils "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestMultiClusterIngressStatusAndAnnotations(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName, _ := GetModelName("foo.com", utils.GetAKONamespace())

	paths := []string{"foo"}
	SetUpTest(t, true, paths, modelName)
	integrationtest.PollForCompletion(t, modelName, 5)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 5*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(len(nodes)).To(gomega.Equal(1))
	g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
	g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
	g.Expect(len(nodes[0].EvhNodes)).To(gomega.Equal(1))

	g.Eventually(func() int {
		mci, _ := CRDClient.AkoV1alpha1().MultiClusterIngresses(utils.GetAKONamespace()).Get(context.TODO(), getMultiClusterIngressName(paths[0]), metav1.GetOptions{})
		return len(mci.Status.LoadBalancer.Ingress)
	}, 10*time.Second).Should(gomega.Equal(1))
	mci, _ := CRDClient.AkoV1alpha1().MultiClusterIngresses(utils.GetAKONamespace()).Get(context.TODO(), getMultiClusterIngressName(paths[0]), metav1.GetOptions{})
	g.Expect(mci.Status.LoadBalancer.Ingress[0].IP).To(gomega.Equal("10.250.250.10"))
	g.Expect(mci.Status.LoadBalancer.Ingress[0].Hostname).To(gomega.ContainSubstring("foo.com"))
	g.Expect(mci.Status.Status.Accepted).To(gomega.BeTrue())

	evhVSName := lib.Encode("cluster--foo.com", lib.EVHVS)
	g.Expect(mci.ObjectMeta.Annotations["ako.vmware.com/host-fqdn-vs-uuid-map"]).To(gomega.ContainSubstring(evhVSName))
	g.Expect(mci.Status.LoadBalancer.Ingress[0].Hostname).To(gomega.ContainSubstring("foo.com"))
	g.Expect(mci.Status.Status.Accepted).To(gomega.BeTrue())

	TearDownTest(t, paths, modelName)
}

func TestMultiClusterIngressCacheSync(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName, vsName := GetModelName("foo.com", utils.GetAKONamespace())

	paths := []string{"foo"}
	SetUpTest(t, true, paths, modelName)
	integrationtest.PollForCompletion(t, modelName, 5)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 5*time.Second).Should(gomega.Equal(true))

	mcache := cache.SharedAviObjCache()
	vsKey := cache.NamespaceName{Namespace: "admin", Name: vsName}
	evhVSKey := cache.NamespaceName{Namespace: "admin", Name: lib.Encode("cluster--foo.com", lib.EVHVS)}
	g.Eventually(func() int {
		if vsCache, found := mcache.VsCacheMeta.AviCacheGet(vsKey); found {
			if vsCacheObj, ok := vsCache.(*cache.AviVsCache); ok {
				return len(vsCacheObj.SNIChildCollection)
			}
		}
		return -1
	}, 15*time.Second).Should(gomega.Equal(1))
	vsCache, _ := mcache.VsCacheMeta.AviCacheGet(vsKey)
	vsCacheObj, _ := vsCache.(*cache.AviVsCache)
	g.Expect(vsCacheObj.Name).To(gomega.Equal(vsName))
	g.Expect(vsCacheObj.PGKeyCollection).To(gomega.HaveLen(0))
	g.Expect(vsCacheObj.PoolKeyCollection).To(gomega.HaveLen(0))
	g.Expect(vsCacheObj.SNIChildCollection).To(gomega.HaveLen(1))

	evhCache, _ := mcache.VsCacheMeta.AviCacheGet(evhVSKey)
	evhCacheObj, _ := evhCache.(*cache.AviVsCache)
	g.Expect(evhCacheObj.ParentVSRef).To(gomega.Equal(vsKey))
	g.Expect(evhCacheObj.PGKeyCollection).To(gomega.HaveLen(1))
	g.Expect(evhCacheObj.PoolKeyCollection).To(gomega.HaveLen(1))

	TearDownTest(t, paths, modelName)
}
