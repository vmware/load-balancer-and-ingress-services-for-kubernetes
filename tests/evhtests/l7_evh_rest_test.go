/*
 * Copyright 2019-2020 VMware, Inc.
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

package evhtests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	avinodes "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	utils "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/testlib"

	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestMultiHostIngressStatusCheckForEvh(t *testing.T) {
	// Skip this test for VIP per Namespace model, as the validations have to be changed.
	// Separate tests would be added to check these conditions.
	if lib.VIPPerNamespace() {
		t.Skip()
	}
	SetupDomain()
	SetUpTestForIngress(t, testlib.AllModels...)
	modelName, _ := GetModelName("foo.com", "default")
	testlib.PollForCompletion(t, modelName, 5)
	g := gomega.NewGomegaWithT(t)

	ingressObject := testlib.FakeIngress{
		Name:        "foo-with-targets",
		Namespace:   "default",
		DnsNames:    []string{"foo.com", "bar.com", "xyz.com"},
		Paths:       []string{"/foo", "/bar", "/xyz"},
		ServiceName: "avisvc",
		TlsSecretDNS: map[string][]string{
			"my-secret":    {"foo.com"},
			"my-secret-v2": {"bar.com"},
		},
	}

	ingressObject2 := testlib.FakeIngress{
		Name:        "foo-with-targets-2",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/doo"},
		ServiceName: "avisvc",
		TlsSecretDNS: map[string][]string{
			"my-secret": {"foo.com"},
		},
	}
	ingrFake := ingressObject.Ingress()
	if _, err := utils.GetInformers().ClientSet.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	ingrFake_2 := ingressObject2.Ingress()
	if _, err := utils.GetInformers().ClientSet.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake_2, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	testlib.PollForCompletion(t, modelName, 5)

	testlib.AddSecret("my-secret-v2", "default", "tlsCert", "tlsKey")
	testlib.AddSecret("my-secret", "default", "tlsCert", "tlsKey")

	// Shard scheme: cluster--Shared-L7-EVH-0 -> foo.com
	// Shard scheme: cluster--Shared-L7-EVH-3 -> xyz.com
	// Shard scheme: cluster--Shared-L7-EVH-1 -> bar.com

	g.Eventually(func() int {
		ingress, _ := utils.GetInformers().ClientSet.NetworkingV1().Ingresses("default").Get(context.TODO(), "foo-with-targets", metav1.GetOptions{})
		return len(ingress.Status.LoadBalancer.Ingress)
	}, 45*time.Second).Should(gomega.Equal(3))
	ingress, _ := utils.GetInformers().ClientSet.NetworkingV1().Ingresses("default").Get(context.TODO(), "foo-with-targets", metav1.GetOptions{})
	// fake avi controller server returns IP in the form: 10.250.250.1<Shared-L7-NUM>
	g.Expect(ingress.Status.LoadBalancer.Ingress[0].IP).To(gomega.MatchRegexp(`^(10.250.250.1(0|1|3))`))
	g.Expect(ingress.Status.LoadBalancer.Ingress[0].Hostname).To(gomega.MatchRegexp(`^((foo|bar|xyz).com)$`))
	g.Expect(ingress.Status.LoadBalancer.Ingress[1].IP).To(gomega.MatchRegexp(`^(10.250.250.1(0|1|3))`))
	g.Expect(ingress.Status.LoadBalancer.Ingress[1].Hostname).To(gomega.MatchRegexp(`^((foo|bar|xyz).com)$`))
	g.Expect(ingress.Status.LoadBalancer.Ingress[2].IP).To(gomega.MatchRegexp(`^(10.250.250.1(0|1|3))`))
	g.Expect(ingress.Status.LoadBalancer.Ingress[2].Hostname).To(gomega.MatchRegexp(`^((foo|bar|xyz).com)$`))

	testlib.DeleteObject(t, lib.Ingress, "foo-with-targets", "default")
	testlib.DeleteObject(t, lib.Ingress, "foo-with-targets-2", "default")
	TearDownTestForIngress(t, modelName)
}

func TestMultiHostUpdateIngressStatusCheckForEvh(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	var err error
	modelName, _ := GetModelName("foo.com", "default")
	ingressId := "thmhuisc"
	ingressName := fmt.Sprintf("ing-%s", ingressId)
	pathSuffix := "-" + ingressName + ".com"

	SetupDomain()
	SetUpTestForIngress(t, testlib.AllModels...)
	testlib.PollForCompletion(t, modelName, 5)

	ingressObject := testlib.FakeIngress{
		Name:        ingressName,
		Namespace:   "default",
		DnsNames:    []string{"foo" + pathSuffix, "xyz" + pathSuffix},
		Paths:       []string{"/foo", "/xyz"},
		ServiceName: "avisvc",
		TlsSecretDNS: map[string][]string{
			"my-secret": {"foo" + pathSuffix},
		},
	}
	testlib.AddSecret("my-secret", "default", "tlsCert", "tlsKey")
	ingrFake := ingressObject.Ingress()
	if _, err = utils.GetInformers().ClientSet.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	testlib.PollForCompletion(t, modelName, 5)

	g.Eventually(func() int {
		ingress, _ := utils.GetInformers().ClientSet.NetworkingV1().Ingresses("default").Get(context.TODO(), ingressName, metav1.GetOptions{})
		return len(ingress.Status.LoadBalancer.Ingress)
	}, 60*time.Second).Should(gomega.Equal(2))
	ingress, _ := utils.GetInformers().ClientSet.NetworkingV1().Ingresses("default").Get(context.TODO(), ingressName, metav1.GetOptions{})

	// donot update status
	var ingressStatusIPs, ingressStatusNames []string
	for _, i := range ingress.Status.LoadBalancer.Ingress {
		ingressStatusIPs = append(ingressStatusIPs, i.IP)
		ingressStatusNames = append(ingressStatusNames, i.Hostname)
	}

	// remove one hostname
	ingressUpdate := (testlib.FakeIngress{
		Name:        ingressName,
		Namespace:   "default",
		DnsNames:    []string{"foo" + pathSuffix},
		Paths:       []string{"/foo"},
		Ips:         ingressStatusIPs,
		HostNames:   ingressStatusNames,
		ServiceName: "avisvc",
		TlsSecretDNS: map[string][]string{
			"my-secret": {"foo" + pathSuffix},
		},
	}).Ingress()
	ingressUpdate.ResourceVersion = "2"
	testlib.UpdateObjectOrFail(t, lib.Ingress, ingressUpdate)

	// statuses go down from 3 to 2
	g.Eventually(func() int {
		ingress, _ := utils.GetInformers().ClientSet.NetworkingV1().Ingresses("default").Get(context.TODO(), ingressName, metav1.GetOptions{})
		return len(ingress.Status.LoadBalancer.Ingress)
	}, 10*time.Second).Should(gomega.Equal(2))

	testlib.DeleteObject(t, lib.Ingress, ingressName, "default")
	TearDownTestForIngress(t, modelName)
}

func TestCreateIngressCacheSyncForEvh(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	var found bool

	modelName, vsName := GetModelName("foo.com", "default")
	SetUpIngressForCacheSyncCheck(t, false, false, modelName)

	g.Eventually(func() bool {
		found, _ = objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 5*time.Second).Should(gomega.Equal(true))

	mcache := cache.SharedAviObjCache()
	vsKey := cache.NamespaceName{Namespace: "admin", Name: vsName}
	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: lib.Encode("cluster--foo.com", lib.EVHVS)}
	vsCache, found := mcache.VsCacheMeta.AviCacheGet(vsKey)
	if !found {
		t.Fatalf("Cache not found for VS: %v", vsKey)
	}
	vsCacheObj, ok := vsCache.(*cache.AviVsCache)
	if !ok {
		t.Fatalf("Invalid VS object. Cannot cast.")
	}
	g.Expect(vsCacheObj.Name).To(gomega.Equal(vsName))
	g.Expect(vsCacheObj.SNIChildCollection).To(gomega.HaveLen(1))

	g.Expect(vsCacheObj.PGKeyCollection).To(gomega.HaveLen(0))
	g.Eventually(func() int {
		vsCache, _ := mcache.VsCacheMeta.AviCacheGet(vsKey)
		vsCacheObj, _ := vsCache.(*cache.AviVsCache)
		return len(vsCacheObj.PoolKeyCollection)
	}, 20*time.Second).Should(gomega.Equal(0))

	g.Expect(vsCacheObj.PoolKeyCollection).To(gomega.HaveLen(0))
	g.Expect(vsCacheObj.DSKeyCollection).To(gomega.HaveLen(0))

	sniCache, _ := mcache.VsCacheMeta.AviCacheGet(sniVSKey)
	sniCacheObj, _ := sniCache.(*cache.AviVsCache)
	g.Expect(sniCacheObj.SSLKeyCertCollection).To(gomega.BeNil())
	g.Expect(sniCacheObj.ParentVSRef).To(gomega.Equal(vsKey))
	g.Expect(sniCacheObj.PoolKeyCollection).To(gomega.HaveLen(1))
	g.Expect(sniCacheObj.PGKeyCollection).To(gomega.HaveLen(1))
	g.Expect(sniCacheObj.HTTPKeyCollection).To(gomega.HaveLen(1))

	testlib.DeleteObject(t, lib.Ingress, "foo-with-targets", "default")
	testlib.DeleteObject(t, lib.Secret, "my-secret", "default")
	TearDownTestForIngress(t, modelName)
}

func TestIngressStatusCheckForEvh(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName, vsName := GetModelName("foo.com", "default")
	SetUpIngressForCacheSyncCheck(t, false, false, modelName)

	mcache := cache.SharedAviObjCache()
	vsKey := cache.NamespaceName{Namespace: "admin", Name: vsName}
	g.Eventually(func() bool {
		_, found := mcache.VsCacheMeta.AviCacheGet(vsKey)
		return found
	}, 5*time.Second).Should(gomega.Equal(true))

	g.Eventually(func() int {
		ingress, _ := utils.GetInformers().ClientSet.NetworkingV1().Ingresses("default").Get(context.TODO(), "foo-with-targets", metav1.GetOptions{})
		return len(ingress.Status.LoadBalancer.Ingress)
	}, 5*time.Second).Should(gomega.Equal(1))
	ingress, _ := utils.GetInformers().ClientSet.NetworkingV1().Ingresses("default").Get(context.TODO(), "foo-with-targets", metav1.GetOptions{})
	g.Expect(ingress.Status.LoadBalancer.Ingress).To(gomega.HaveLen(1))

	TearDownIngressForCacheSyncCheck(t, modelName)
}

func TestUpdatePoolCacheSyncForEvh(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName, _ := GetModelName("foo.com", "default")
	SetUpIngressForCacheSyncCheck(t, false, false, modelName)

	// Get hold of the pool checksum on CREATE
	poolName := "cluster--default-foo.com_foo-foo-with-targets-avisvc"
	mcache := cache.SharedAviObjCache()
	poolKey := cache.NamespaceName{Namespace: testlib.AVINAMESPACE, Name: lib.Encode(poolName, lib.Pool)}
	poolCacheBefore, _ := mcache.PoolCache.AviCacheGet(poolKey)
	poolCacheBeforeObj, _ := poolCacheBefore.(*cache.AviPoolCache)
	oldPoolCksum := poolCacheBeforeObj.CloudConfigCksum

	epExample := &corev1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "avisvc"},
		Subsets: []corev1.EndpointSubset{{
			Addresses: []corev1.EndpointAddress{{IP: "1.2.3.4"}, {IP: "1.2.3.5"}},
			Ports:     []corev1.EndpointPort{{Name: "foo", Port: 8080, Protocol: "TCP"}},
		}},
	}
	epExample.ResourceVersion = "2"
	testlib.UpdateObjectOrFail(t, lib.Endpoint, epExample)

	g.Eventually(func() []avinodes.AviPoolMetaServer {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		vs := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return vs[0].EvhNodes[0].PoolRefs[0].Servers
	}, 5*time.Second).Should(gomega.HaveLen(2))

	g.Eventually(func() string {
		if poolCache, found := mcache.PoolCache.AviCacheGet(poolKey); found {
			if poolCacheObj, ok := poolCache.(*cache.AviPoolCache); ok {
				return poolCacheObj.CloudConfigCksum
			}
		}
		return ""
	}, 10*time.Second).Should(gomega.Not(gomega.Equal(oldPoolCksum)))
	// If we transition the service from clusterIP to Loadbalancer - pools' servers should get deleted.
	svcExample := (testlib.FakeService{
		Name:         "avisvc",
		Namespace:    "default",
		Type:         corev1.ServiceTypeLoadBalancer,
		ServicePorts: []testlib.Serviceport{{PortName: "foo0", Protocol: "TCP", PortNumber: 8080, TargetPort: 8080}},
	}).Service()
	svcExample.ResourceVersion = "3"
	testlib.UpdateObjectOrFail(t, lib.Service, svcExample)

	g.Eventually(func() []avinodes.AviPoolMetaServer {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		vs := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return vs[0].EvhNodes[0].PoolRefs[0].Servers
	}, 60*time.Second).Should(gomega.HaveLen(0))
	// If we transition the service from Loadbalancer to clusterIP - pools' servers should get populated.
	svcExample = (testlib.FakeService{
		Name:         "avisvc",
		Namespace:    "default",
		Type:         corev1.ServiceTypeClusterIP,
		ServicePorts: []testlib.Serviceport{{PortName: "foo0", Protocol: "TCP", PortNumber: 8080, TargetPort: 8080}},
	}).Service()
	svcExample.ResourceVersion = "4"
	testlib.UpdateObjectOrFail(t, lib.Service, svcExample)

	g.Eventually(func() []avinodes.AviPoolMetaServer {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		vs := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return vs[0].EvhNodes[0].PoolRefs[0].Servers
	}, 15*time.Second).Should(gomega.HaveLen(2))
	TearDownIngressForCacheSyncCheck(t, modelName)
}

func TestCreateCacheSyncForEvh(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName, vsName := GetModelName("foo.com", "default")
	SetUpIngressForCacheSyncCheck(t, true, true, modelName)

	mcache := cache.SharedAviObjCache()
	parentVSKey := cache.NamespaceName{Namespace: "admin", Name: vsName}
	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: lib.Encode("cluster--foo.com", lib.EVHVS)}

	g.Eventually(func() bool {
		_, found := mcache.VsCacheMeta.AviCacheGet(sniVSKey)
		return found
	}, 10*time.Second).Should(gomega.Equal(true))
	parentCache, _ := mcache.VsCacheMeta.AviCacheGet(parentVSKey)
	parentCacheObj, _ := parentCache.(*cache.AviVsCache)
	g.Expect(parentCacheObj.Name).To(gomega.Equal(vsName))
	g.Expect(parentCacheObj.SNIChildCollection).To(gomega.HaveLen(1))
	g.Expect(parentCacheObj.PGKeyCollection).To(gomega.HaveLen(0))
	g.Expect(parentCacheObj.PoolKeyCollection).To(gomega.HaveLen(0))
	g.Expect(parentCacheObj.PoolKeyCollection).To(gomega.HaveLen(0))
	g.Expect(parentCacheObj.DSKeyCollection).To(gomega.HaveLen(0))
	g.Expect(parentCacheObj.SSLKeyCertCollection).To(gomega.HaveLen(1))

	sniCache, _ := mcache.VsCacheMeta.AviCacheGet(sniVSKey)
	sniCacheObj, _ := sniCache.(*cache.AviVsCache)
	g.Expect(sniCacheObj.SSLKeyCertCollection).To(gomega.BeNil())
	g.Expect(sniCacheObj.ParentVSRef).To(gomega.Equal(parentVSKey))
	g.Expect(sniCacheObj.PoolKeyCollection).To(gomega.HaveLen(1))
	g.Expect(sniCacheObj.PGKeyCollection).To(gomega.HaveLen(1))
	g.Expect(sniCacheObj.HTTPKeyCollection).To(gomega.HaveLen(2))

	TearDownIngressForCacheSyncCheck(t, modelName)

}

func TestUpdateCacheSyncForEvh(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName, _ := GetModelName("foo.com", "default")
	SetUpIngressForCacheSyncCheck(t, true, true, modelName)

	mcache := cache.SharedAviObjCache()
	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: lib.Encode("cluster--foo.com", lib.EVHVS)}
	g.Eventually(func() bool {
		_, found := mcache.VsCacheMeta.AviCacheGet(sniVSKey)
		return found
	}, 15*time.Second).Should(gomega.Equal(true))
	oldSniCache, _ := mcache.VsCacheMeta.AviCacheGet(sniVSKey)
	oldSniCacheObj, _ := oldSniCache.(*cache.AviVsCache)

	ingressUpdate := (testlib.FakeIngress{
		Name:        "foo-with-targets",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/bar-updated"},
		ServiceName: "avisvc",
		TlsSecretDNS: map[string][]string{
			"my-secret": {"foo.com"},
		},
	}).Ingress()
	ingressUpdate.ResourceVersion = "2"
	testlib.UpdateObjectOrFail(t, lib.Ingress, ingressUpdate)

	// verify that a NEW httppolicy set object is created
	oldHttpPolKey := cache.NamespaceName{Namespace: "admin", Name: lib.Encode("cluster--default-foo.com", lib.HTTPPS)}
	newHttpPolKey := cache.NamespaceName{Namespace: "admin", Name: lib.Encode("cluster--default-foo.com", lib.HTTPPS)}
	g.Eventually(func() bool {
		_, found := mcache.HTTPPolicyCache.AviCacheGet(newHttpPolKey)
		return found
	}, 10*time.Second).Should(gomega.Equal(true))

	g.Eventually(func() bool {
		_, found := mcache.HTTPPolicyCache.AviCacheGet(oldHttpPolKey)
		return found
	}, 10*time.Second).Should(gomega.Equal(true))

	// verify same vs cksum
	g.Eventually(func() string {
		sniVSCache, found := mcache.VsCacheMeta.AviCacheGet(sniVSKey)
		sniVSCacheObj, ok := sniVSCache.(*cache.AviVsCache)
		if found && ok {
			return sniVSCacheObj.CloudConfigCksum
		}
		return "456def"
	}, 15*time.Second).Should(gomega.Equal(oldSniCacheObj.CloudConfigCksum))
	sniVSCache, _ := mcache.VsCacheMeta.AviCacheGet(sniVSKey)
	sniVSCacheObj, _ := sniVSCache.(*cache.AviVsCache)
	g.Expect(sniVSCacheObj.HTTPKeyCollection).To(gomega.HaveLen(2))
	g.Expect(sniVSCacheObj.SSLKeyCertCollection).To(gomega.HaveLen(0))

	TearDownIngressForCacheSyncCheck(t, modelName)
}

func TestMultiHostMultiSecretCacheSyncForEvh(t *testing.T) {
	// Skip this test for VIP per Namespace model, as the validations have to be changed
	if lib.VIPPerNamespace() {
		t.Skip()
	}
	g := gomega.NewGomegaWithT(t)

	modelName, vsName := GetModelName("foo.com", "default")
	SetUpIngressForCacheSyncCheck(t, true, true, modelName)
	mcache := cache.SharedAviObjCache()
	testlib.AddSecret("my-secret", "default", "tlsCert", "tlsKey")
	// update ingress
	ingressObject := testlib.FakeIngress{
		Name:        "foo-with-targets",
		Namespace:   "default",
		DnsNames:    []string{"foo.com", "bar.com"},
		Paths:       []string{"/foo", "/bar"},
		ServiceName: "avisvc",
		TlsSecretDNS: map[string][]string{
			"my-secret":    {"foo.com"},
			"my-secret-v2": {"bar.com"},
		},
	}
	testlib.AddSecret("my-secret-v2", "default", "tlsCert", "tlsKey")
	ingrFake := ingressObject.Ingress()
	ingrFake.ResourceVersion = "2"
	testlib.UpdateObjectOrFail(t, lib.Ingress, ingrFake)

	sniVSKey1 := cache.NamespaceName{Namespace: "admin", Name: lib.Encode("cluster--foo.com", lib.EVHVS)}
	sniVSKey2 := cache.NamespaceName{Namespace: "admin", Name: lib.Encode("cluster--bar.com", lib.EVHVS)}
	g.Eventually(func() bool {
		sniCache1, found1 := mcache.VsCacheMeta.AviCacheGet(sniVSKey1)
		sniCache2, found2 := mcache.VsCacheMeta.AviCacheGet(sniVSKey2)
		sniCacheObj1, _ := sniCache1.(*cache.AviVsCache)
		sniCacheObj2, _ := sniCache2.(*cache.AviVsCache)
		if found1 && found2 &&
			len(sniCacheObj1.PGKeyCollection) == 1 &&
			len(sniCacheObj2.PGKeyCollection) == 1 {
			return true
		}
		return false
	}, 20*time.Second).Should(gomega.Equal(true))

	parentVSKey := cache.NamespaceName{Namespace: "admin", Name: vsName}
	parentCache, _ := mcache.VsCacheMeta.AviCacheGet(parentVSKey)
	//cluster--foo.com is the cert bound to this VS
	g.Eventually(func() string {
		parentCacheObj, _ := parentCache.(*cache.AviVsCache)
		if parentCacheObj != nil && len(parentCacheObj.SSLKeyCertCollection) > 0 {
			return parentCacheObj.SSLKeyCertCollection[0].Name
		}
		return ""
	}, 40*time.Second).Should(gomega.Equal(lib.Encode("cluster--foo.com", lib.SSLKeyCert)))

	parentVSKey1 := cache.NamespaceName{Namespace: "admin", Name: "cluster--Shared-L7-EVH-1"}
	parentCache1, _ := mcache.VsCacheMeta.AviCacheGet(parentVSKey1)
	//cluster--bar.com is the cert bound to this VS
	g.Eventually(func() string {
		parentCacheObj1, _ := parentCache1.(*cache.AviVsCache)
		if parentCacheObj1 != nil && len(parentCacheObj1.SSLKeyCertCollection) > 0 {
			return parentCacheObj1.SSLKeyCertCollection[0].Name
		}
		return ""
	}, 40*time.Second).Should(gomega.Equal(lib.Encode("cluster--bar.com", lib.SSLKeyCert)))

	g.Eventually(func() string {
		sniCache1, _ := mcache.VsCacheMeta.AviCacheGet(sniVSKey1)
		sniCacheObj1, _ := sniCache1.(*cache.AviVsCache)
		return sniCacheObj1.ParentVSRef.Name
	}, 15*time.Second).Should(gomega.Not(gomega.Equal("")))

	// create the ingress
	ingressObject = testlib.FakeIngress{
		Name:        "foo-with-targets-2",
		Namespace:   "red",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/doo"},
		ServiceName: "avisvc",
		TlsSecretDNS: map[string][]string{
			"my-secret": {"foo.com"},
		},
	}
	testlib.AddSecret("my-secret", "red", "tlsCert", "tlsKey")
	ingrFake = ingressObject.Ingress()
	if _, err := utils.GetInformers().ClientSet.NetworkingV1().Ingresses("red").Create(context.TODO(), ingrFake, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in updating Ingress: %v", err)
	}
	sniVSKey3 := cache.NamespaceName{Namespace: "admin", Name: lib.Encode("cluster--foo.com", lib.EVHVS)}
	g.Eventually(func() bool {
		_, found1 := mcache.VsCacheMeta.AviCacheGet(sniVSKey3)
		return found1
	}, 20*time.Second).Should(gomega.Equal(true))
	testlib.DeleteObject(t, lib.Ingress, "foo-with-targets-2", "red")
	TearDownIngressForCacheSyncCheck(t, modelName)
}

func TestMultiHostMultiSecretUpdateCacheSyncForEvh(t *testing.T) {
	// Skip this test for VIP per Namespace model, as the validations have to be changed
	if lib.VIPPerNamespace() {
		t.Skip()
	}
	g := gomega.NewGomegaWithT(t)
	modelName, vsName := GetModelName("foo.com", "default")

	SetupDomain()
	SetUpTestForIngress(t, modelName, modelName)
	testlib.PollForCompletion(t, modelName, 5)

	ingressObject := testlib.FakeIngress{
		Name:        "foo-with-targets",
		Namespace:   "default",
		DnsNames:    []string{"foo.com", "bar.com", "xyz.com"},
		Paths:       []string{"/foo", "/bar", "/xyz"},
		ServiceName: "avisvc",
		TlsSecretDNS: map[string][]string{
			"my-secret":    {"foo.com"},
			"my-secret-v2": {"bar.com"},
		},
	}
	testlib.AddSecret("my-secret-v2", "default", "tlsCert", "tlsKey")
	testlib.AddSecret("my-secret", "default", "tlsCert", "tlsKey")

	ingrFake := ingressObject.Ingress()
	if _, err := utils.GetInformers().ClientSet.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	testlib.PollForCompletion(t, modelName, 5)

	mcache := cache.SharedAviObjCache()
	parentVSKey := cache.NamespaceName{Namespace: "admin", Name: vsName}
	sniVSKey1 := cache.NamespaceName{Namespace: "admin", Name: lib.Encode("cluster--foo.com", lib.EVHVS)}
	sniVSKey2 := cache.NamespaceName{Namespace: "admin", Name: lib.Encode("cluster--bar.com", lib.EVHVS)}
	xyzParentKey := cache.NamespaceName{Namespace: "admin", Name: vsName}
	_, barVSName := GetModelName("bar.com", "default")
	barParentKey := cache.NamespaceName{Namespace: "admin", Name: barVSName}

	// Shard scheme: cluster--Shared-L7-EVH-0 -> foo.com
	// Shard scheme: cluster--Shared-L7-3 -> xyz.com

	parentCache, _ := mcache.VsCacheMeta.AviCacheGet(parentVSKey)
	parentCacheObj, _ := parentCache.(*cache.AviVsCache)
	g.Expect(parentCacheObj.SNIChildCollection).To(gomega.HaveLen(1))

	barCache, _ := mcache.VsCacheMeta.AviCacheGet(barParentKey)
	barCacheObj, _ := barCache.(*cache.AviVsCache)
	g.Expect(barCacheObj.SNIChildCollection).To(gomega.HaveLen(1))

	xyzCache, _ := mcache.VsCacheMeta.AviCacheGet(xyzParentKey)
	xyzCacheObj, _ := xyzCache.(*cache.AviVsCache)
	g.Expect(xyzCacheObj.SNIChildCollection).To(gomega.HaveLen(1))

	g.Eventually(func() int {
		sniCache, _ := mcache.VsCacheMeta.AviCacheGet(xyzParentKey)
		sniCacheObj, _ := sniCache.(*cache.AviVsCache)
		return len(sniCacheObj.SNIChildCollection)
	}, 10*time.Second).Should(gomega.Equal(1))
	sniCache, _ := mcache.VsCacheMeta.AviCacheGet(xyzParentKey)
	sniCacheObj, _ := sniCache.(*cache.AviVsCache)
	g.Expect(sniCacheObj.SNIChildCollection).To(gomega.HaveLen(1))

	g.Eventually(func() bool {
		sniCache1, found1 := mcache.VsCacheMeta.AviCacheGet(sniVSKey1)
		sniCache2, found2 := mcache.VsCacheMeta.AviCacheGet(sniVSKey2)
		sniCacheObj1, _ := sniCache1.(*cache.AviVsCache)
		sniCacheObj2, _ := sniCache2.(*cache.AviVsCache)
		if found1 && found2 &&
			len(sniCacheObj1.HTTPKeyCollection) == 2 && len(sniCacheObj2.HTTPKeyCollection) == 2 {
			return true
		}
		return false
	}, 20*time.Second).Should(gomega.Equal(true))

	g.Eventually(func() int {
		sniCache, found := mcache.VsCacheMeta.AviCacheGet(sniVSKey1)
		sniCacheObj, ok := sniCache.(*cache.AviVsCache)
		if found && ok {
			return len(sniCacheObj.PoolKeyCollection)
		}
		return 0
	}, 10*time.Second).Should(gomega.Equal(1))
	sniCache, _ = mcache.VsCacheMeta.AviCacheGet(sniVSKey1)
	sniCacheObj, _ = sniCache.(*cache.AviVsCache)
	g.Expect(sniCacheObj.SSLKeyCertCollection).To(gomega.HaveLen(0))
	g.Expect(parentCacheObj.SSLKeyCertCollection).To(gomega.HaveLen(1))
	g.Expect(parentCacheObj.SSLKeyCertCollection[0].Name).To(gomega.Equal(lib.Encode("cluster--foo.com", lib.SSLKeyCert)))

	g.Expect(barCacheObj.SSLKeyCertCollection).To(gomega.HaveLen(1))
	g.Expect(barCacheObj.SSLKeyCertCollection[0].Name).To(gomega.Equal(lib.Encode("cluster--bar.com", lib.SSLKeyCert)))

	// delete one secret
	testlib.DeleteObject(t, lib.Secret, "my-secret-v2", "default")
	ingressUpdateObject := testlib.FakeIngress{
		Name:        "foo-with-targets",
		Namespace:   "default",
		DnsNames:    []string{"foo.com", "bar.com", "xyz.com"},
		Paths:       []string{"/foo", "/bar", "/xyz"},
		ServiceName: "avisvc",
		TlsSecretDNS: map[string][]string{
			"my-secret": {"foo.com"},
		},
	}

	ingrUpdate := ingressUpdateObject.Ingress()
	ingrUpdate.ResourceVersion = "2"
	testlib.UpdateObjectOrFail(t, lib.Ingress, ingrUpdate)

	// Shard scheme: cluster--Shared-L7-1 -> bar.com
	g.Eventually(func() int {
		sniCache, _ := mcache.VsCacheMeta.AviCacheGet(barParentKey)
		sniCacheObj, _ := sniCache.(*cache.AviVsCache)
		return len(sniCacheObj.SNIChildCollection)
	}, 10*time.Second).Should(gomega.Equal(1))

	// bar should not be deleted.
	g.Eventually(func() bool {
		_, found := mcache.VsCacheMeta.AviCacheGet(sniVSKey2)
		return found
	}, 10*time.Second).Should(gomega.Equal(true))

	g.Eventually(func() bool {
		_, found := mcache.VsCacheMeta.AviCacheGet(sniVSKey1)
		return found
	}, 10*time.Second).Should(gomega.Equal(true))
	sniCache, _ = mcache.VsCacheMeta.AviCacheGet(sniVSKey1)
	sniCacheObj, _ = sniCache.(*cache.AviVsCache)
	g.Expect(sniCacheObj.PoolKeyCollection).To(gomega.HaveLen(1))
	g.Expect(sniCacheObj.SSLKeyCertCollection).To(gomega.HaveLen(0))

	testlib.DeleteObject(t, lib.Ingress, "foo-with-targets", "default")
	testlib.DeleteObject(t, lib.Secret, "my-secret", "default")
	g.Eventually(func() bool {
		_, found := mcache.VsCacheMeta.AviCacheGet(sniVSKey1)
		return found
	}, 15*time.Second).Should(gomega.Equal(false))
	TearDownTestForIngress(t, modelName)
}

// Secure ingress to insecure ingress transition
func TestDeleteCacheSyncForEvh(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName, vsName := GetModelName("foo.com", "default")
	SetUpIngressForCacheSyncCheck(t, true, true, modelName)

	mcache := cache.SharedAviObjCache()
	parentVSKey := cache.NamespaceName{Namespace: "admin", Name: vsName}
	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: lib.Encode("cluster--foo.com", lib.EVHVS)}

	ingressUpdate := (testlib.FakeIngress{
		Name:        "foo-with-targets",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		Paths:       []string{"/foo"},
		ServiceName: "avisvc",
	}).Ingress()
	ingressUpdate.ResourceVersion = "2"
	testlib.UpdateObjectOrFail(t, lib.Ingress, ingressUpdate)

	// verify that evh vs is not deleted, and a redirect rule is added to the parent shard vs
	g.Eventually(func() bool {
		_, found := mcache.VsCacheMeta.AviCacheGet(sniVSKey)
		parentSniCache, _ := mcache.VsCacheMeta.AviCacheGet(parentVSKey)
		parentSniCacheObj, _ := parentSniCache.(*cache.AviVsCache)

		if found && len(parentSniCacheObj.SNIChildCollection) == 1 && len(parentSniCacheObj.HTTPKeyCollection) == 0 {
			return true
		}
		return false
	}, 20*time.Second).Should(gomega.Equal(true))

	TearDownIngressForCacheSyncCheck(t, modelName)
}

func TestCUDSecretCacheSyncForEvh(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName, vsName := GetModelName("foo.com", "default")
	SetUpIngressForCacheSyncCheck(t, true, false, modelName)

	mcache := cache.SharedAviObjCache()
	parentVSKey := cache.NamespaceName{Namespace: "admin", Name: vsName}
	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: lib.Encode("cluster--foo.com", lib.EVHVS)}
	sslKey := cache.NamespaceName{Namespace: "admin", Name: lib.Encode("cluster--foo.com", lib.SSLKeyCert)}

	// no ssl key cache would be found since the secret is not yet added
	g.Eventually(func() bool {
		_, found := mcache.SSLKeyCache.AviCacheGet(sslKey)
		return found
	}, 5*time.Second).Should(gomega.Equal(false))

	// add Secret
	testlib.AddSecret("my-secret", "default", "tlsCert", "tlsKey")

	// ssl key should be created now and must be attached to the sni vs cache
	g.Eventually(func() bool {
		_, found := mcache.SSLKeyCache.AviCacheGet(sslKey)
		return found
	}, 50*time.Second).Should(gomega.Equal(true))
	parentVSCache, _ := mcache.VsCacheMeta.AviCacheGet(parentVSKey)
	parentVSCacheObj, _ := parentVSCache.(*cache.AviVsCache)
	g.Expect(parentVSCacheObj.HTTPKeyCollection).To(gomega.HaveLen(0))
	g.Expect(parentVSCacheObj.SSLKeyCertCollection).To(gomega.HaveLen(1))

	// update Secret
	secretUpdate := (testlib.FakeSecret{
		Namespace: "default",
		Name:      "my-secret",
		Cert:      "tlsCert_Updated",
		Key:       "tlsKey_Updated",
	}).Secret()
	secretUpdate.ResourceVersion = "2"
	testlib.UpdateObjectOrFail(t, lib.Secret, secretUpdate)

	// can't check update rn, ssl cache object doesnot have checksum,
	// but PUTs happen, everytime though

	// delete Secret
	testlib.DeleteObject(t, lib.Secret, "my-secret", "default")

	// ssl key must be deleted again and evh vs should  be deleted
	g.Eventually(func() bool {
		_, sslfound := mcache.SSLKeyCache.AviCacheGet(sslKey)
		_, snifound := mcache.VsCacheMeta.AviCacheGet(sniVSKey)
		return sslfound && snifound
	}, 15*time.Second).Should(gomega.Equal(false))

	g.Eventually(func() bool {
		parentVSCache, found := mcache.VsCacheMeta.AviCacheGet(parentVSKey)
		parentVSCacheObj, ok := parentVSCache.(*cache.AviVsCache)
		if found && ok && len(parentVSCacheObj.HTTPKeyCollection) == 0 {
			return true
		}
		return false
	}, 30*time.Second).Should(gomega.Equal(true))

	TearDownIngressForCacheSyncCheck(t, modelName)
}

func TestDeleteSecretSecureIngressStatusCheckForEvh(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	modelName, _ := GetModelName("foo.com", "default")
	SetUpIngressForCacheSyncCheck(t, true, true, modelName)

	g.Eventually(func() int {
		ingress, _ := utils.GetInformers().ClientSet.NetworkingV1().Ingresses("default").Get(context.TODO(), "foo-with-targets", metav1.GetOptions{})
		return len(ingress.Status.LoadBalancer.Ingress)
	}, 30*time.Second).Should(gomega.Equal(1))

	// post this EVH VS should get deleted, and ingress status must be updated accordingly
	testlib.DeleteObject(t, lib.Secret, "my-secret", "default")

	g.Eventually(func() int {
		ingress, _ := utils.GetInformers().ClientSet.NetworkingV1().Ingresses("default").Get(context.TODO(), "foo-with-targets", metav1.GetOptions{})
		return len(ingress.Status.LoadBalancer.Ingress)
	}, 50*time.Second).Should(gomega.Equal(0))

	TearDownIngressForCacheSyncCheck(t, modelName)
}
