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

package integrationtest

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	avinodes "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"

	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func SetupDomain() {
	mcache := cache.SharedAviObjCache()
	cloud, _ := mcache.CloudKeyCache.AviCacheGet("Default-Cloud")
	cloudProperty, _ := cloud.(*cache.AviCloudPropertyCache)
	subdomains := []string{"avi.internal", ".com"}
	cloudProperty.NSIpamDNS = subdomains
}

func SetUpIngressForCacheSyncCheck(t *testing.T, modelName string, tlsIngress, withSecret bool) {
	SetupDomain()
	SetUpTestForIngress(t, modelName)
	PollForCompletion(t, modelName, 5)
	ingressObject := FakeIngress{
		Name:        "foo-with-targets",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/foo"},
		ServiceName: "avisvc",
	}
	if withSecret {
		AddSecret("my-secret", "default", "tlsCert", "tlsKey")
	}
	if tlsIngress {
		ingressObject.TlsSecretDNS = map[string][]string{
			"my-secret": {"foo.com"},
		}
	}
	ingrFake := ingressObject.Ingress()
	if _, err := KubeClient.NetworkingV1beta1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	PollForCompletion(t, modelName, 5)
}

func TearDownIngressForCacheSyncCheck(t *testing.T, modelName string, g *gomega.GomegaWithT) {
	if err := KubeClient.NetworkingV1beta1().Ingresses("default").Delete(context.TODO(), "foo-with-targets", metav1.DeleteOptions{}); err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	g.Eventually(func() error {
		_, err := KubeClient.NetworkingV1beta1().Ingresses("default").Get(context.TODO(), "foo-with-targets", metav1.GetOptions{})
		return err
	}, 30*time.Second).Should(gomega.Not(gomega.BeNil()))
	TearDownTestForIngress(t, modelName)
}

func TestCreateIngressCacheSync(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	var found bool

	modelName := "admin/cluster--Shared-L7-6"
	SetUpIngressForCacheSyncCheck(t, modelName, false, false)

	g.Eventually(func() bool {
		found, _ = objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 10*time.Second).Should(gomega.Equal(true))

	mcache := cache.SharedAviObjCache()
	vsKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--Shared-L7-6"}
	vsCache, found := mcache.VsCacheMeta.AviCacheGet(vsKey)
	if !found {
		t.Fatalf("Cache not found for VS: %v", vsKey)
	}
	vsCacheObj, ok := vsCache.(*cache.AviVsCache)
	if !ok {
		t.Fatalf("Invalid VS object. Cannot cast.")
	}
	g.Expect(vsCacheObj.Name).To(gomega.Equal("cluster--Shared-L7-6"))
	g.Expect(vsCacheObj.PGKeyCollection).To(gomega.HaveLen(1))
	g.Expect(vsCacheObj.PoolKeyCollection).To(gomega.HaveLen(1))
	g.Expect(vsCacheObj.PoolKeyCollection[0].Name).To(gomega.ContainSubstring("foo-with-targets"))
	g.Expect(vsCacheObj.DSKeyCollection).To(gomega.HaveLen(1))
	g.Expect(vsCacheObj.SSLKeyCertCollection).To(gomega.BeNil())

	TearDownIngressForCacheSyncCheck(t, modelName, g)
}
func TestCreateIngressWithFaultCacheSync(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	var found bool

	injectFault := true
	AddMiddleware(func(w http.ResponseWriter, r *http.Request) {
		var resp map[string]interface{}
		var finalResponse []byte
		url := r.URL.EscapedPath()

		if strings.Contains(url, "macro") && r.Method == "POST" {
			data, _ := ioutil.ReadAll(r.Body)
			json.Unmarshal(data, &resp)
			rData, rModelName := resp["data"].(map[string]interface{}), strings.ToLower(resp["model_name"].(string))
			if rModelName == "virtualservice" && injectFault {
				injectFault = false
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprintln(w, `{"error": "bad request"}`)
			} else {
				rName := rData["name"].(string)
				objURL := fmt.Sprintf("https://localhost/api/%s/%s-%s#%s", rModelName, rModelName, RANDOMUUID, rName)

				// adding additional 'uuid' and 'url' (read-only) fields in the response
				rData["url"] = objURL
				rData["uuid"] = fmt.Sprintf("%s-%s-%s", rModelName, rName, RANDOMUUID)
				finalResponse, _ = json.Marshal([]interface{}{resp["data"]})
				w.WriteHeader(http.StatusOK)
				fmt.Fprintln(w, string(finalResponse))
			}
		} else if r.Method == "PUT" {
			data, _ := ioutil.ReadAll(r.Body)
			json.Unmarshal(data, &resp)
			resp["uuid"] = strings.Split(strings.Trim(url, "/"), "/")[2]
			finalResponse, _ = json.Marshal(resp)
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, string(finalResponse))
		} else if r.Method == "DELETE" {
			w.WriteHeader(http.StatusNoContent)
			fmt.Fprintln(w, string(finalResponse))
		} else if strings.Contains(url, "login") {
			// This is used for /login --> first request to controller
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, `{"success": "true"}`)
		}
	})
	defer ResetMiddleware()

	modelName := "admin/cluster--Shared-L7-6"
	SetUpIngressForCacheSyncCheck(t, modelName, false, false)

	g.Eventually(func() int {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return len(nodes[0].PoolRefs)
	}, 60*time.Second).Should(gomega.Equal(1))

	mcache := cache.SharedAviObjCache()
	vsKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--Shared-L7-6"}
	g.Eventually(func() int {
		vsCache, _ := mcache.VsCacheMeta.AviCacheGet(vsKey)
		vsCacheObj, _ := vsCache.(*cache.AviVsCache)
		return len(vsCacheObj.PoolKeyCollection)
	}, 5*time.Second).Should(gomega.Equal(1))

	vsCache, found := mcache.VsCacheMeta.AviCacheGet(vsKey)
	if !found {
		t.Fatalf("Cache not found for VS: %v", vsKey)
	}
	vsCacheObj, ok := vsCache.(*cache.AviVsCache)
	if !ok {
		t.Fatalf("Invalid VS object. Cannot cast.")
	}
	g.Expect(vsCacheObj.Name).To(gomega.Equal("cluster--Shared-L7-6"))
	g.Expect(vsCacheObj.PGKeyCollection).To(gomega.HaveLen(1))
	g.Expect(vsCacheObj.PoolKeyCollection).To(gomega.HaveLen(1))
	g.Expect(vsCacheObj.PoolKeyCollection[0].Name).To(gomega.ContainSubstring("foo-with-targets"))
	g.Expect(vsCacheObj.DSKeyCollection).To(gomega.HaveLen(1))
	g.Expect(vsCacheObj.SSLKeyCertCollection).To(gomega.BeNil())

	TearDownIngressForCacheSyncCheck(t, modelName, g)
}

func TestUpdatePoolCacheSync(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	var err error

	modelName := "admin/cluster--Shared-L7-6"
	SetUpIngressForCacheSyncCheck(t, modelName, false, false)

	// Get hold of the pool checksum on CREATE
	poolName := "cluster--foo.com_foo-default-foo-with-targets"
	mcache := cache.SharedAviObjCache()
	poolKey := cache.NamespaceName{Namespace: AVINAMESPACE, Name: poolName}
	g.Eventually(func() bool {
		_, found := mcache.PoolCache.AviCacheGet(poolKey)
		return found
	}, 60*time.Second).Should(gomega.Equal(true))
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
	if _, err = KubeClient.CoreV1().Endpoints("default").Update(context.TODO(), epExample, metav1.UpdateOptions{}); err != nil {
		t.Fatalf("error in creating Endpoint: %v", err)
	}

	g.Eventually(func() int {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		vs := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		if len(vs) < 1 {
			return 0
		}
		if len(vs[0].PoolRefs) < 1 {
			return 0
		}
		return len(vs[0].PoolRefs[0].Servers)
	}, 60*time.Second).Should(gomega.Equal(2))

	g.Eventually(func() string {
		if poolCache, found := mcache.PoolCache.AviCacheGet(poolKey); found {
			if poolCacheObj, ok := poolCache.(*cache.AviPoolCache); ok {
				return poolCacheObj.CloudConfigCksum
			}
		}
		return ""
	}, 20*time.Second).Should(gomega.Not(gomega.Equal(oldPoolCksum)))

	TearDownIngressForCacheSyncCheck(t, modelName, g)
}

func TestDeletePoolCacheSync(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	var err error

	modelName := "admin/cluster--Shared-L7-6"
	SetUpIngressForCacheSyncCheck(t, modelName, false, false)

	ingressUpdate := (FakeIngress{
		Name:        "foo-with-targets",
		Namespace:   "default",
		DnsNames:    []string{"bar.com"},
		Ips:         []string{"8.8.8.8"},
		ServiceName: "avisvc",
	}).Ingress()
	ingressUpdate.ResourceVersion = "2"
	if _, err = KubeClient.NetworkingV1beta1().Ingresses("default").Update(context.TODO(), ingressUpdate, metav1.UpdateOptions{}); err != nil {
		t.Fatalf("error in updating Ingress: %v", err)
	}

	// check that old pool is deleted and new one is created, will have different names
	oldPoolKey := cache.NamespaceName{Namespace: AVINAMESPACE, Name: "cluster--foo.com_foo-default-foo-with-targets"}
	newPoolKey := cache.NamespaceName{Namespace: AVINAMESPACE, Name: "cluster--bar.com_foo-default-foo-with-targets"}
	mcache := cache.SharedAviObjCache()
	g.Eventually(func() bool {
		_, found := mcache.PoolCache.AviCacheGet(oldPoolKey)
		return found
	}, 5*time.Second).Should(gomega.Equal(false))
	g.Eventually(func() bool {
		_, found := mcache.PoolCache.AviCacheGet(newPoolKey)
		return found
	}, 60*time.Second).Should(gomega.Equal(true))
	newPoolCache, _ := mcache.PoolCache.AviCacheGet(newPoolKey)
	newPoolCacheObj, _ := newPoolCache.(*cache.AviPoolCache)
	g.Expect(newPoolCacheObj.Name).To(gomega.Not(gomega.ContainSubstring("foo.com")))
	g.Expect(newPoolCacheObj.Name).To(gomega.ContainSubstring("bar.com"))

	TearDownIngressForCacheSyncCheck(t, modelName, g)
}

func TestCreateSNICacheSync(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName := "admin/cluster--Shared-L7-6"
	SetUpIngressForCacheSyncCheck(t, modelName, true, true)

	mcache := cache.SharedAviObjCache()
	parentVSKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--Shared-L7-6"}
	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--foo-with-targets-default-my-secret"}

	g.Eventually(func() bool {
		_, found := mcache.VsCacheMeta.AviCacheGet(sniVSKey)
		return found
	}, 30*time.Second).Should(gomega.Equal(true))

	g.Eventually(func() bool {
		parentCache, found := mcache.VsCacheMeta.AviCacheGet(parentVSKey)
		parentCacheObj, ok := parentCache.(*cache.AviVsCache)
		if found && ok && len(parentCacheObj.SNIChildCollection) == 1 {
			return true
		}
		return false
	}, 30*time.Second).Should(gomega.Equal(true))
	parentCache, _ := mcache.VsCacheMeta.AviCacheGet(parentVSKey)
	parentCacheObj, _ := parentCache.(*cache.AviVsCache)
	g.Expect(parentCacheObj.SNIChildCollection[0]).To(gomega.ContainSubstring("cluster--foo-with-targets-default-my-secret"))
	g.Expect(parentCacheObj.HTTPKeyCollection).To(gomega.HaveLen(1))

	sniCache, _ := mcache.VsCacheMeta.AviCacheGet(sniVSKey)
	sniCacheObj, _ := sniCache.(*cache.AviVsCache)
	g.Expect(sniCacheObj.SSLKeyCertCollection).To(gomega.HaveLen(1))
	g.Expect(sniCacheObj.SSLKeyCertCollection[0].Name).To(gomega.ContainSubstring("cluster--default-my-secret"))
	g.Expect(sniCacheObj.HTTPKeyCollection).To(gomega.HaveLen(1))
	g.Expect(sniCacheObj.HTTPKeyCollection[0].Name).To(gomega.ContainSubstring("cluster--default-foo.com"))
	g.Expect(sniCacheObj.ParentVSRef).To(gomega.Equal(parentVSKey))

	TearDownIngressForCacheSyncCheck(t, modelName, g)
}

func TestUpdateSNICacheSync(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	var err error

	modelName := "admin/cluster--Shared-L7-6"
	SetUpIngressForCacheSyncCheck(t, modelName, true, true)

	mcache := cache.SharedAviObjCache()
	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--foo-with-targets-default-my-secret"}
	g.Eventually(func() bool {
		_, found := mcache.VsCacheMeta.AviCacheGet(sniVSKey)
		return found
	}, 15*time.Second).Should(gomega.Equal(true))
	oldSniCache, _ := mcache.VsCacheMeta.AviCacheGet(sniVSKey)
	oldSniCacheObj, _ := oldSniCache.(*cache.AviVsCache)

	ingressUpdate := (FakeIngress{
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
	_, err = KubeClient.NetworkingV1beta1().Ingresses("default").Update(context.TODO(), ingressUpdate, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating Ingress: %v", err)
	}

	// verify that a NEW httppolicy set object is created
	oldHttpPolKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--default-foo.com_foo-foo-with-targets"}
	newHttpPolKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--default-foo.com_bar-updated-foo-with-targets"}
	g.Eventually(func() bool {
		_, found := mcache.HTTPPolicyCache.AviCacheGet(newHttpPolKey)
		return found
	}, 60*time.Second).Should(gomega.Equal(true))

	g.Eventually(func() bool {
		_, found := mcache.HTTPPolicyCache.AviCacheGet(oldHttpPolKey)
		return found
	}, 60*time.Second).Should(gomega.Equal(false))

	// verify same vs cksum
	g.Eventually(func() string {
		sniVSCache, found := mcache.VsCacheMeta.AviCacheGet(sniVSKey)
		sniVSCacheObj, ok := sniVSCache.(*cache.AviVsCache)
		if found && ok {
			return sniVSCacheObj.CloudConfigCksum
		}
		return "456def"
	}, 60*time.Second).Should(gomega.Equal(oldSniCacheObj.CloudConfigCksum))
	sniVSCache, _ := mcache.VsCacheMeta.AviCacheGet(sniVSKey)
	sniVSCacheObj, _ := sniVSCache.(*cache.AviVsCache)
	g.Expect(sniVSCacheObj.HTTPKeyCollection).To(gomega.HaveLen(1))
	g.Expect(sniVSCacheObj.SSLKeyCertCollection).To(gomega.HaveLen(1))

	TearDownIngressForCacheSyncCheck(t, modelName, g)
}

func TestMultiHostMultiSecretSNICacheSync(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName := "admin/cluster--Shared-L7-6"
	SetUpIngressForCacheSyncCheck(t, modelName, true, true)
	mcache := cache.SharedAviObjCache()

	// update ingress
	ingressObject := FakeIngress{
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
	AddSecret("my-secret-v2", "default", "tlsCert", "tlsKey")
	ingrFake := ingressObject.Ingress()
	ingrFake.ResourceVersion = "2"
	if _, err := KubeClient.NetworkingV1beta1().Ingresses("default").Update(context.TODO(), ingrFake, metav1.UpdateOptions{}); err != nil {
		t.Fatalf("error in updating Ingress: %v", err)
	}

	sniVSKey1 := cache.NamespaceName{Namespace: "admin", Name: "cluster--foo-with-targets-default-my-secret"}
	sniVSKey2 := cache.NamespaceName{Namespace: "admin", Name: "cluster--foo-with-targets-default-my-secret-v2"}
	g.Eventually(func() bool {
		_, found1 := mcache.VsCacheMeta.AviCacheGet(sniVSKey1)
		_, found2 := mcache.VsCacheMeta.AviCacheGet(sniVSKey2)
		if found1 && found2 {
			return true
		}
		return false
	}, 15*time.Second).Should(gomega.Equal(true))

	g.Eventually(func() int {
		sniCache1, _ := mcache.VsCacheMeta.AviCacheGet(sniVSKey1)
		sniCacheObj1, _ := sniCache1.(*cache.AviVsCache)
		return len(sniCacheObj1.SSLKeyCertCollection)
	}, 30*time.Second).Should(gomega.Equal(1))

	g.Eventually(func() string {
		sniCache2, _ := mcache.VsCacheMeta.AviCacheGet(sniVSKey2)
		sniCacheObj2, _ := sniCache2.(*cache.AviVsCache)
		if len(sniCacheObj2.SSLKeyCertCollection) > 0 {
			return sniCacheObj2.SSLKeyCertCollection[0].Name
		}
		return ""
	}, 30*time.Second).Should(gomega.Equal("cluster--default-my-secret-v2"))

	TearDownIngressForCacheSyncCheck(t, modelName, g)
}

func TestMultiHostMultiSecretUpdateSNICacheSync(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	modelName := "admin/cluster--Shared-L7-6"

	SetupDomain()
	SetUpTestForIngress(t, modelName)
	PollForCompletion(t, modelName, 5)
	ingressObject := FakeIngress{
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
	AddSecret("my-secret-v2", "default", "tlsCert", "tlsKey")
	AddSecret("my-secret", "default", "tlsCert", "tlsKey")

	ingrFake := ingressObject.Ingress()
	if _, err := KubeClient.NetworkingV1beta1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	PollForCompletion(t, modelName, 5)

	mcache := cache.SharedAviObjCache()
	parentVSKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--Shared-L7-6"}
	sniVSKey1 := cache.NamespaceName{Namespace: "admin", Name: "cluster--foo-with-targets-default-my-secret"}
	sniVSKey2 := cache.NamespaceName{Namespace: "admin", Name: "cluster--foo-with-targets-default-my-secret-v2"}

	g.Eventually(func() int {
		sniCache, _ := mcache.VsCacheMeta.AviCacheGet(parentVSKey)
		sniCacheObj, _ := sniCache.(*cache.AviVsCache)
		return len(sniCacheObj.PoolKeyCollection)
	}, 30*time.Second).Should(gomega.Equal(1))
	sniCache, _ := mcache.VsCacheMeta.AviCacheGet(parentVSKey)
	sniCacheObj, _ := sniCache.(*cache.AviVsCache)
	g.Expect(sniCacheObj.PoolKeyCollection[0].Name).To(gomega.ContainSubstring("xyz.com"))

	g.Eventually(func() bool {
		sniCache, found := mcache.VsCacheMeta.AviCacheGet(sniVSKey1)
		sniCacheObj, ok := sniCache.(*cache.AviVsCache)
		if found && ok &&
			len(sniCacheObj.PoolKeyCollection) == 1 &&
			len(sniCacheObj.SSLKeyCertCollection) == 1 {
			return true
		}
		return false
	}, 15*time.Second).Should(gomega.Equal(true))
	sniCache, _ = mcache.VsCacheMeta.AviCacheGet(sniVSKey1)
	sniCacheObj, _ = sniCache.(*cache.AviVsCache)
	g.Expect(sniCacheObj.PoolKeyCollection[0].Name).To(gomega.ContainSubstring("foo.com"))
	g.Expect(sniCacheObj.SSLKeyCertCollection[0].Name).To(gomega.Equal("cluster--default-my-secret"))

	g.Eventually(func() bool {
		sniCache, found := mcache.VsCacheMeta.AviCacheGet(sniVSKey2)
		sniCacheObj, ok := sniCache.(*cache.AviVsCache)
		if found && ok {
			if len(sniCacheObj.PoolKeyCollection) == 1 && len(sniCacheObj.SSLKeyCertCollection) == 1 {
				return true
			}
		}
		return false
	}, 25*time.Second).Should(gomega.Equal(true))
	sniCache, _ = mcache.VsCacheMeta.AviCacheGet(sniVSKey2)
	sniCacheObj, _ = sniCache.(*cache.AviVsCache)
	g.Expect(sniCacheObj.PoolKeyCollection[0].Name).To(gomega.ContainSubstring("bar.com"))
	g.Expect(sniCacheObj.SSLKeyCertCollection).To(gomega.HaveLen(1))
	g.Expect(sniCacheObj.SSLKeyCertCollection[0].Name).To(gomega.Equal("cluster--default-my-secret-v2"))

	// delete cert
	KubeClient.CoreV1().Secrets("default").Delete(context.TODO(), "my-secret-v2", metav1.DeleteOptions{})
	ingressUpdateObject := FakeIngress{
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
	utils.AviLog.Infof("Updated ingress after removing one secret: %s", utils.Stringify(ingrUpdate))
	ingrUpdate.ResourceVersion = "2"
	if _, err := KubeClient.NetworkingV1beta1().Ingresses("default").Update(context.TODO(), ingrUpdate, metav1.UpdateOptions{}); err != nil {
		t.Fatalf("error in updating Ingress: %v", err)
	}

	g.Eventually(func() int {
		sniCache, _ := mcache.VsCacheMeta.AviCacheGet(parentVSKey)
		sniCacheObj, _ := sniCache.(*cache.AviVsCache)
		return len(sniCacheObj.PoolKeyCollection)
	}, 60*time.Second).Should(gomega.Equal(2))

	// should not be found
	g.Eventually(func() bool {
		_, found := mcache.VsCacheMeta.AviCacheGet(sniVSKey2)
		return found
	}, 30*time.Second).Should(gomega.Equal(false))

	sniCache, _ = mcache.VsCacheMeta.AviCacheGet(sniVSKey1)
	sniCacheObj, _ = sniCache.(*cache.AviVsCache)
	g.Expect(sniCacheObj.PoolKeyCollection).To(gomega.HaveLen(1))
	g.Expect(sniCacheObj.PoolKeyCollection[0].Name).To(gomega.ContainSubstring("foo.com"))
	g.Expect(sniCacheObj.SSLKeyCertCollection).To(gomega.HaveLen(1))
	g.Expect(sniCacheObj.SSLKeyCertCollection[0].Name).To(gomega.Equal("cluster--default-my-secret"))

	KubeClient.NetworkingV1beta1().Ingresses("default").Delete(context.TODO(), "foo-with-targets", metav1.DeleteOptions{})
	KubeClient.CoreV1().Secrets("default").Delete(context.TODO(), "my-secret", metav1.DeleteOptions{})
	g.Eventually(func() bool {
		_, found := mcache.VsCacheMeta.AviCacheGet(sniVSKey1)
		return found
	}, 60*time.Second).Should(gomega.Equal(false))
	TearDownTestForIngress(t, modelName)
}

func TestDeleteSNICacheSync(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	var err error

	modelName := "admin/cluster--Shared-L7-6"
	SetUpIngressForCacheSyncCheck(t, modelName, true, true)

	mcache := cache.SharedAviObjCache()
	parentVSKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--Shared-L7-6"}
	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--foo-with-targets-default-my-secret"}

	ingressUpdate := (FakeIngress{
		Name:        "foo-with-targets",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/foo"},
		ServiceName: "avisvc",
	}).Ingress()
	ingressUpdate.ResourceVersion = "2"
	_, err = KubeClient.NetworkingV1beta1().Ingresses("default").Update(context.TODO(), ingressUpdate, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating Ingress: %v", err)
	}

	// verify that sni vs is deleted, but the parent vs is not
	// deleted snivs key should be deleted from parent vs snichildcollection
	g.Eventually(func() bool {
		_, found := mcache.VsCacheMeta.AviCacheGet(sniVSKey)
		return found
	}, 15*time.Second).Should(gomega.Equal(false))

	g.Eventually(func() bool {
		parentSniCache, _ := mcache.VsCacheMeta.AviCacheGet(parentVSKey)
		parentSniCacheObj, _ := parentSniCache.(*cache.AviVsCache)
		if len(parentSniCacheObj.SNIChildCollection) != 0 {
			return false
		}
		if len(parentSniCacheObj.HTTPKeyCollection) != 0 {
			return false
		}
		return true
	}, 30*time.Second).Should(gomega.Equal(true))

	TearDownIngressForCacheSyncCheck(t, modelName, g)
}

func TestCUDSecretCacheSync(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName := "admin/cluster--Shared-L7-6"
	SetUpIngressForCacheSyncCheck(t, modelName, true, false)

	mcache := cache.SharedAviObjCache()
	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--foo-with-targets-default-my-secret"}
	sslKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--default-my-secret"}

	// no ssl key cache would be found since the secret is not yet added
	g.Eventually(func() bool {
		_, found := mcache.SSLKeyCache.AviCacheGet(sslKey)
		return found
	}, 5*time.Second).Should(gomega.Equal(false))

	// add Secret
	AddSecret("my-secret", "default", "tlsCert", "tlsKey")

	// ssl key should be created now and must be attached to the sni vs cache
	g.Eventually(func() bool {
		_, found := mcache.SSLKeyCache.AviCacheGet(sslKey)
		return found
	}, 5*time.Second).Should(gomega.Equal(true))
	sniVSCache, _ := mcache.VsCacheMeta.AviCacheGet(sniVSKey)
	sniVSCacheObj, _ := sniVSCache.(*cache.AviVsCache)
	g.Expect(sniVSCacheObj.SSLKeyCertCollection).To(gomega.HaveLen(1))

	// update Secret
	secretUpdate := (FakeSecret{
		Namespace: "default",
		Name:      "my-secret",
		Cert:      "tlsCert_Updated",
		Key:       "tlsKey_Updated",
	}).Secret()
	secretUpdate.ResourceVersion = "2"
	KubeClient.CoreV1().Secrets("default").Update(context.TODO(), secretUpdate, metav1.UpdateOptions{})

	// can't check update rn, ssl cache object doesnot have checksum,
	// but PUTs happen, everytime though

	// delete Secret
	KubeClient.CoreV1().Secrets("default").Delete(context.TODO(), "my-secret", metav1.DeleteOptions{})

	// ssl key must be deleted again and sni vs as well
	g.Eventually(func() bool {
		_, found := mcache.SSLKeyCache.AviCacheGet(sslKey)
		return found
	}, 60*time.Second).Should(gomega.Equal(false))
	_, found := mcache.VsCacheMeta.AviCacheGet(sniVSKey)
	g.Expect(found).To(gomega.Equal(false))

	TearDownIngressForCacheSyncCheck(t, modelName, g)
}

func TestIngressStatusCheck(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName := "admin/cluster--Shared-L7-6"
	SetUpIngressForCacheSyncCheck(t, modelName, false, false)

	mcache := cache.SharedAviObjCache()
	vsKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--Shared-L7-6"}
	g.Eventually(func() bool {
		_, found := mcache.VsCacheMeta.AviCacheGet(vsKey)
		return found
	}, 5*time.Second).Should(gomega.Equal(true))

	g.Eventually(func() int {
		ingress, _ := KubeClient.NetworkingV1beta1().Ingresses("default").Get(context.TODO(), "foo-with-targets", metav1.GetOptions{})
		return len(ingress.Status.LoadBalancer.Ingress)
	}, 60*time.Second).Should(gomega.Equal(1))
	ingress, _ := KubeClient.NetworkingV1beta1().Ingresses("default").Get(context.TODO(), "foo-with-targets", metav1.GetOptions{})
	g.Expect(ingress.Status.LoadBalancer.Ingress[0].IP).To(gomega.Equal("10.250.250.16"))
	g.Expect(ingress.Status.LoadBalancer.Ingress[0].Hostname).To(gomega.ContainSubstring("foo.com"))

	TearDownIngressForCacheSyncCheck(t, modelName, g)
}

func TestMultiHostIngressStatusCheck(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	modelName := "admin/cluster--Shared-L7-6"
	ingressName := "foo-with-targets-2"

	SetupDomain()
	SetUpTestForIngress(t, modelName)
	PollForCompletion(t, modelName, 5)
	ingressObject := FakeIngress{
		Name:        ingressName,
		Namespace:   "default",
		DnsNames:    []string{"foo2.com", "bar2.com", "xyz2.com"},
		Paths:       []string{"/foo", "/bar", "/xyz"},
		ServiceName: "avisvc",
		TlsSecretDNS: map[string][]string{
			"my-secret":    {"foo2.com"},
			"my-secret-v2": {"xyz2.com"},
		},
	}
	AddSecret("my-secret-v2", "default", "tlsCert", "tlsKey")
	AddSecret("my-secret", "default", "tlsCert", "tlsKey")

	ingrFake := ingressObject.Ingress()
	if _, err := KubeClient.NetworkingV1beta1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	PollForCompletion(t, modelName, 5)

	g.Eventually(func() int {
		ingress, _ := KubeClient.NetworkingV1beta1().Ingresses("default").Get(context.TODO(), ingressName, metav1.GetOptions{})
		return len(ingress.Status.LoadBalancer.Ingress)
	}, 60*time.Second).Should(gomega.Equal(3))
	ingress, _ := KubeClient.NetworkingV1beta1().Ingresses("default").Get(context.TODO(), ingressName, metav1.GetOptions{})

	// fake avi controller server returns IP in the form: 10.250.250.1<Shard-VS-NUM>
	g.Expect(ingress.Status.LoadBalancer.Ingress[0].IP).To(gomega.Equal("10.250.250.16"))
	g.Expect(ingress.Status.LoadBalancer.Ingress[0].Hostname).To(gomega.MatchRegexp(`^((foo|bar|xyz)2.com)$`))
	g.Expect(ingress.Status.LoadBalancer.Ingress[1].IP).To(gomega.Equal("10.250.250.16"))
	g.Expect(ingress.Status.LoadBalancer.Ingress[1].Hostname).To(gomega.MatchRegexp(`^((foo|bar|xyz)2.com)$`))
	g.Expect(ingress.Status.LoadBalancer.Ingress[2].IP).To(gomega.Equal("10.250.250.16"))
	g.Expect(ingress.Status.LoadBalancer.Ingress[2].Hostname).To(gomega.MatchRegexp(`^((foo|bar|xyz)2.com)$`))

	KubeClient.NetworkingV1beta1().Ingresses("default").Delete(context.TODO(), "foo-with-targets-2", metav1.DeleteOptions{})
	TearDownTestForIngress(t, modelName)
}

func TestMultiHostUpdateIngressStatusCheck(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	var err error
	modelName := "admin/cluster--Shared-L7-6"
	ingressName := "foo-with-targets-3"

	SetupDomain()
	SetUpTestForIngress(t, modelName)
	PollForCompletion(t, modelName, 5)
	ingressObject := FakeIngress{
		Name:        ingressName,
		Namespace:   "default",
		DnsNames:    []string{"foo3.com", "bar3.com", "xyz3.com"},
		Paths:       []string{"/foo", "/bar", "/xyz"},
		ServiceName: "avisvc",
		TlsSecretDNS: map[string][]string{
			"my-secret": {"foo3.com", "bar3.com"},
		},
	}
	AddSecret("my-secret", "default", "tlsCert", "tlsKey")
	ingrFake := ingressObject.Ingress()
	if _, err = KubeClient.NetworkingV1beta1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	PollForCompletion(t, modelName, 5)

	g.Eventually(func() int {
		ingress, _ := KubeClient.NetworkingV1beta1().Ingresses("default").Get(context.TODO(), ingressName, metav1.GetOptions{})
		return len(ingress.Status.LoadBalancer.Ingress)
	}, 15*time.Second).Should(gomega.Equal(3))
	ingress, _ := KubeClient.NetworkingV1beta1().Ingresses("default").Get(context.TODO(), ingressName, metav1.GetOptions{})
	var ingressStatusIPs, ingressStatusNames []string
	for _, i := range ingress.Status.LoadBalancer.Ingress {
		ingressStatusIPs = append(ingressStatusIPs, i.IP)
		ingressStatusNames = append(ingressStatusNames, i.Hostname)
	}

	// remove one hostname
	ingressUpdate := (FakeIngress{
		Name:        ingressName,
		Namespace:   "default",
		DnsNames:    []string{"foo3.com", "xyz3.com"},
		Ips:         ingressStatusIPs,
		HostNames:   ingressStatusNames,
		Paths:       []string{"/foo", "/xyz"},
		ServiceName: "avisvc",
		TlsSecretDNS: map[string][]string{
			"my-secret": {"foo3.com"},
		},
	}).Ingress()
	ingressUpdate.ResourceVersion = "2"
	_, err = KubeClient.NetworkingV1beta1().Ingresses("default").Update(context.TODO(), ingressUpdate, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating Ingress: %v", err)
	}

	// statuses go down from 3 to 2
	g.Eventually(func() int {
		ingress, _ := KubeClient.NetworkingV1beta1().Ingresses("default").Get(context.TODO(), ingressName, metav1.GetOptions{})
		return len(ingress.Status.LoadBalancer.Ingress)
	}, 60*time.Second).Should(gomega.Equal(2))

	KubeClient.NetworkingV1beta1().Ingresses("default").Delete(context.TODO(), ingressName, metav1.DeleteOptions{})
	TearDownTestForIngress(t, modelName)
}

func TestDeleteSecretSecureIngressStatusCheck(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	modelName := "admin/cluster--Shared-L7-6"
	SetUpIngressForCacheSyncCheck(t, modelName, true, true)

	g.Eventually(func() int {
		ingress, _ := KubeClient.NetworkingV1beta1().Ingresses("default").Get(context.TODO(), "foo-with-targets", metav1.GetOptions{})
		return len(ingress.Status.LoadBalancer.Ingress)
	}, 30*time.Second).Should(gomega.Equal(1))

	// post this SNI VS should get deleted, and ingress status must be updated accordingly
	KubeClient.CoreV1().Secrets("default").Delete(context.TODO(), "my-secret", metav1.DeleteOptions{})

	g.Eventually(func() int {
		ingress, _ := KubeClient.NetworkingV1beta1().Ingresses("default").Get(context.TODO(), "foo-with-targets", metav1.GetOptions{})
		return len(ingress.Status.LoadBalancer.Ingress)
	}, 60*time.Second).Should(gomega.Equal(0))

	TearDownIngressForCacheSyncCheck(t, modelName, g)
}

func TestCreateIngressCacheSyncWithMultiTenant(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetAkoTenant()
	defer ResetAkoTenant()
	var found bool
	modelName := fmt.Sprintf("%s/cluster--Shared-L7-6", AKOTENANT)
	SetUpIngressForCacheSyncCheck(t, modelName, false, false)

	g.Eventually(func() bool {
		found, _ = objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 10*time.Second).Should(gomega.Equal(true))

	mcache := cache.SharedAviObjCache()
	vsKey := cache.NamespaceName{Namespace: AKOTENANT, Name: "cluster--Shared-L7-6"}
	vsCache, found := mcache.VsCacheMeta.AviCacheGet(vsKey)
	if !found {
		t.Fatalf("Cache not found for VS: %v", vsKey)
	}
	vsCacheObj, ok := vsCache.(*cache.AviVsCache)
	if !ok {
		t.Fatalf("Invalid VS object. Cannot cast.")
	}
	g.Expect(vsCacheObj.Name).To(gomega.Equal("cluster--Shared-L7-6"))
	TearDownIngressForCacheSyncCheck(t, modelName, g)
}
