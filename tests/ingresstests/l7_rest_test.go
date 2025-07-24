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

package ingresstests

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	avinodes "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/integrationtest"

	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type IngressTestObject struct {
	ingressName string
	namespace   string
	dnsNames    []string
	ipAddrs     []string
	hostnames   []string
	paths       []string
	isTLS       bool
	withSecret  bool
	secretName  string
	serviceName string
	modelNames  []string
}

func (ing *IngressTestObject) FillParams() {
	if ing.namespace == "" {
		ing.namespace = "default"
	}
	if len(ing.dnsNames) == 0 {
		ing.dnsNames = append(ing.dnsNames, "foo.com")
	}
	if len(ing.ipAddrs) == 0 {
		ing.ipAddrs = append(ing.ipAddrs, "8.8.8.8")
	}
	if len(ing.hostnames) == 0 {
		ing.hostnames = append(ing.hostnames, "v1")
	}
	if len(ing.paths) == 0 {
		ing.paths = append(ing.paths, "/foo")
	}
}

func SetupDomain() {
	mcache := cache.SharedAviObjCache()
	cloudObj := &cache.AviCloudPropertyCache{Name: "Default-Cloud", VType: "mock"}
	subdomains := []string{"avi.internal", ".com"}
	cloudObj.NSIpamDNS = subdomains
	mcache.CloudKeyCache.AviCacheAdd("Default-Cloud", cloudObj)
}

func SetUpIngressForCacheSyncCheck(t *testing.T, ingTestObj IngressTestObject) {
	SetupDomain()
	SetUpTestForIngress(t, ingTestObj.serviceName, ingTestObj.modelNames...)
	ingressObject := integrationtest.FakeIngress{
		Name:        ingTestObj.ingressName,
		Namespace:   ingTestObj.namespace,
		DnsNames:    ingTestObj.dnsNames,
		Ips:         ingTestObj.ipAddrs,
		HostNames:   ingTestObj.hostnames,
		Paths:       ingTestObj.paths,
		ServiceName: ingTestObj.serviceName,
	}
	if len(ingTestObj.paths) == 0 {
		ingressObject.NoPath = true
	}
	if ingTestObj.withSecret {
		integrationtest.AddSecret(ingTestObj.secretName, ingTestObj.namespace, "tlsCert", "tlsKey")
	}
	if ingTestObj.isTLS {
		ingressObject.TlsSecretDNS = map[string][]string{
			ingTestObj.secretName: {ingTestObj.dnsNames[0]},
		}
	}
	ingrFake := ingressObject.Ingress()
	if _, err := KubeClient.NetworkingV1().Ingresses(ingTestObj.namespace).Create(context.TODO(), ingrFake, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	integrationtest.PollForCompletion(t, ingTestObj.modelNames[0], 5)
}

func CreateIngress(t *testing.T, ingTestObj IngressTestObject) {
	ingressObject := integrationtest.FakeIngress{
		Name:        ingTestObj.ingressName,
		Namespace:   ingTestObj.namespace,
		DnsNames:    ingTestObj.dnsNames,
		Ips:         ingTestObj.ipAddrs,
		HostNames:   ingTestObj.hostnames,
		Paths:       ingTestObj.paths,
		ServiceName: ingTestObj.serviceName,
	}
	if len(ingTestObj.paths) == 0 {
		ingressObject.NoPath = true
	}
	if ingTestObj.withSecret {
		integrationtest.AddSecret(ingTestObj.secretName, ingTestObj.namespace, "tlsCert", "tlsKey")
	}
	if ingTestObj.isTLS {
		ingressObject.TlsSecretDNS = map[string][]string{
			ingTestObj.secretName: {"foo.com"},
		}
	}
	ingrFake := ingressObject.Ingress()
	if _, err := KubeClient.NetworkingV1().Ingresses(ingTestObj.namespace).Create(context.TODO(), ingrFake, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	integrationtest.PollForCompletion(t, ingTestObj.modelNames[0], 5)
}

func TearDownIngressForCacheSyncCheck(t *testing.T, ingName, svcName, secretName, modelName string) {
	if err := KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), ingName, metav1.DeleteOptions{}); err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	if secretName != "" {
		KubeClient.CoreV1().Secrets("default").Delete(context.TODO(), secretName, metav1.DeleteOptions{})
	}
	TearDownTestForIngress(t, svcName, modelName)
}

func CleanupCache(vsName string) {
	mcache := cache.SharedAviObjCache()
	vsKey := cache.NamespaceName{Namespace: "admin", Name: vsName}
	mcache.VsCacheMeta.AviCacheDelete(vsKey)
}

func TestCreateIngressCacheSync(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	var found bool

	CleanupCache("cluster--Shared-L7-0")

	modelName := MODEL_NAME_PREFIX + "0"
	svcName := objNameMap.GenerateName("avisvc")
	ingName := objNameMap.GenerateName("foo-with-targets")
	ingTestObj := IngressTestObject{
		ingressName: ingName,
		isTLS:       false,
		withSecret:  false,
		serviceName: svcName,
		modelNames:  []string{modelName},
	}
	ingTestObj.FillParams()
	SetUpIngressForCacheSyncCheck(t, ingTestObj)

	g.Eventually(func() bool {
		found, _ = objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 5*time.Second).Should(gomega.Equal(true))

	mcache := cache.SharedAviObjCache()
	vsKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--Shared-L7-0"}
	vsCache, found := mcache.VsCacheMeta.AviCacheGet(vsKey)
	if !found {
		t.Fatalf("Cache not found for VS: %v", vsKey)
	}
	vsCacheObj, ok := vsCache.(*cache.AviVsCache)
	if !ok {
		t.Fatalf("Invalid VS object. Cannot cast.")
	}
	g.Expect(vsCacheObj.Name).To(gomega.Equal("cluster--Shared-L7-0"))
	g.Expect(vsCacheObj.PGKeyCollection).To(gomega.HaveLen(1))
	g.Eventually(func() int {
		vsCache, _ := mcache.VsCacheMeta.AviCacheGet(vsKey)
		vsCacheObj, _ := vsCache.(*cache.AviVsCache)
		return len(vsCacheObj.PoolKeyCollection)
	}, 20*time.Second).Should(gomega.Equal(1))

	g.Expect(vsCacheObj.PoolKeyCollection).To(gomega.HaveLen(1))
	g.Expect(vsCacheObj.PoolKeyCollection[0].Name).To(gomega.ContainSubstring(ingName))
	g.Expect(vsCacheObj.DSKeyCollection).To(gomega.HaveLen(1))
	g.Expect(vsCacheObj.SSLKeyCertCollection).To(gomega.BeNil())

	if err := KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), ingName, metav1.DeleteOptions{}); err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	// make sure that ingress deletion is synced in cache, then delete the model
	g.Eventually(func() int {
		vsCache, _ := mcache.VsCacheMeta.AviCacheGet(vsKey)
		vsCacheObj, _ := vsCache.(*cache.AviVsCache)
		return len(vsCacheObj.PoolKeyCollection)
	}, 10*time.Second).Should(gomega.Equal(0))
	TearDownTestForIngress(t, svcName, modelName)
}

func TestIngressStatusCheck(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	mcache := cache.SharedAviObjCache()
	integrationtest.ClearAllCache(mcache)

	modelName := MODEL_NAME_PREFIX + "0"
	svcName := objNameMap.GenerateName("avisvc")
	ingName := objNameMap.GenerateName("foo-with-targets")
	ingTestObj := IngressTestObject{
		ingressName: ingName,
		isTLS:       false,
		withSecret:  false,
		serviceName: svcName,
		modelNames:  []string{modelName},
	}
	ingTestObj.FillParams()
	SetUpIngressForCacheSyncCheck(t, ingTestObj)

	vsKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--Shared-L7-0"}
	g.Eventually(func() bool {
		_, found := mcache.VsCacheMeta.AviCacheGet(vsKey)
		return found
	}, 5*time.Second).Should(gomega.Equal(true))

	g.Eventually(func() int {
		ingress, _ := KubeClient.NetworkingV1().Ingresses("default").Get(context.TODO(), ingName, metav1.GetOptions{})
		return len(ingress.Status.LoadBalancer.Ingress)
	}, 20*time.Second).Should(gomega.Equal(1))
	ingress, _ := KubeClient.NetworkingV1().Ingresses("default").Get(context.TODO(), ingName, metav1.GetOptions{})
	g.Expect(ingress.Status.LoadBalancer.Ingress).To(gomega.HaveLen(1))
	g.Expect(ingress.Status.LoadBalancer.Ingress[0].IP).To(gomega.Equal("10.250.250.10"))
	g.Expect(ingress.Status.LoadBalancer.Ingress[0].Hostname).To(gomega.ContainSubstring("foo.com"))

	TearDownIngressForCacheSyncCheck(t, ingName, svcName, "", modelName)
}

func TestCreateIngressWithFaultCacheSync(t *testing.T) {
	t.Skip("skipping test, but needs UT fix")
	g := gomega.NewGomegaWithT(t)
	var found bool

	modelName := MODEL_NAME_PREFIX + "0"
	svcName := objNameMap.GenerateName("avisvc")
	ingName := objNameMap.GenerateName("foo-with-targets")

	injectFault := true
	integrationtest.AddMiddleware(func(w http.ResponseWriter, r *http.Request) {
		var resp map[string]interface{}
		var finalResponse []byte
		url := r.URL.EscapedPath()

		if strings.Contains(url, "macro") && r.Method == "POST" {
			data, _ := io.ReadAll(r.Body)
			json.Unmarshal(data, &resp)
			rData, rModelName := resp["data"].(map[string]interface{}), strings.ToLower(resp["model_name"].(string))
			if rModelName == "virtualservice" && injectFault {
				injectFault = false
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprintln(w, `{"error": "bad request"}`)
			} else {
				rName := rData["name"].(string)
				objURL := fmt.Sprintf("https://localhost/api/%s/%s-%s#%s", rModelName, rModelName, integrationtest.RANDOMUUID, rName)

				// adding additional 'uuid' and 'url' (read-only) fields in the response
				rData["url"] = objURL
				rData["uuid"] = fmt.Sprintf("%s-%s-%s", rModelName, rName, integrationtest.RANDOMUUID)
				finalResponse, _ = json.Marshal([]interface{}{resp["data"]})
				w.WriteHeader(http.StatusOK)
				fmt.Fprintln(w, string(finalResponse))
			}
		} else if r.Method == "PUT" {
			data, _ := io.ReadAll(r.Body)
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
	defer integrationtest.ResetMiddleware()

	ingTestObj := IngressTestObject{
		ingressName: ingName,
		isTLS:       false,
		withSecret:  false,
		serviceName: svcName,
		modelNames:  []string{modelName},
	}
	ingTestObj.FillParams()
	SetUpIngressForCacheSyncCheck(t, ingTestObj)

	g.Eventually(func() int {
		if ok, aviModel := objects.SharedAviGraphLister().Get(modelName); ok {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
			if len(nodes) > 0 {
				return len(nodes[0].PoolRefs)
			}
		}
		return -1
	}, 5*time.Second).Should(gomega.Equal(1))

	t.Fatalf("arif fail")

	mcache := cache.SharedAviObjCache()
	vsKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--Shared-L7-0"}
	g.Eventually(func() int {
		if vsCache, ok := mcache.VsCacheMeta.AviCacheGet(vsKey); ok {
			if vsCacheObj, ok := vsCache.(*cache.AviVsCache); ok {
				return len(vsCacheObj.PoolKeyCollection)
			}
		}
		return -1
	}, 5*time.Second).Should(gomega.Equal(0))

	vsCache, found := mcache.VsCacheMeta.AviCacheGet(vsKey)
	if !found {
		t.Fatalf("Cache not found for VS: %v", vsKey)
	}
	vsCacheObj, ok := vsCache.(*cache.AviVsCache)
	if !ok {
		t.Fatalf("Invalid VS object. Cannot cast.")
	}
	g.Expect(vsCacheObj.Name).To(gomega.Equal("cluster--Shared-L7-0"))
	g.Expect(vsCacheObj.PGKeyCollection).To(gomega.HaveLen(1))
	g.Expect(vsCacheObj.PoolKeyCollection).To(gomega.HaveLen(1))
	g.Expect(vsCacheObj.PoolKeyCollection[0].Name).To(gomega.ContainSubstring(ingName))
	g.Expect(vsCacheObj.DSKeyCollection).To(gomega.HaveLen(1))
	g.Expect(vsCacheObj.SSLKeyCertCollection).To(gomega.BeNil())

	TearDownIngressForCacheSyncCheck(t, ingName, svcName, "", modelName)
}

func TestUpdatePoolCacheSync(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	var err error

	modelName := MODEL_NAME_PREFIX + "0"
	svcName := objNameMap.GenerateName("avisvc")
	ingName := objNameMap.GenerateName("foo-with-targets")
	ingTestObj := IngressTestObject{
		ingressName: ingName,
		isTLS:       false,
		withSecret:  false,
		serviceName: svcName,
		modelNames:  []string{modelName},
	}
	ingTestObj.FillParams()
	SetUpIngressForCacheSyncCheck(t, ingTestObj)

	// Get hold of the pool checksum on CREATE
	poolName := "cluster--foo.com_foo-default-" + ingName
	mcache := cache.SharedAviObjCache()
	poolKey := cache.NamespaceName{Namespace: integrationtest.AVINAMESPACE, Name: poolName}
	poolCacheBefore, _ := mcache.PoolCache.AviCacheGet(poolKey)
	poolCacheBeforeObj, _ := poolCacheBefore.(*cache.AviPoolCache)
	oldPoolCksum := poolCacheBeforeObj.CloudConfigCksum

	integrationtest.ScaleCreateEPS(t, "default", svcName)

	g.Eventually(func() []avinodes.AviPoolMetaServer {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		vs := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return vs[0].PoolRefs[0].Servers
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
	svcExample := (integrationtest.FakeService{
		Name:         svcName,
		Namespace:    "default",
		Type:         corev1.ServiceTypeLoadBalancer,
		ServicePorts: []integrationtest.Serviceport{{PortName: "foo0", Protocol: "TCP", PortNumber: 8080, TargetPort: intstr.FromInt(8080)}},
	}).Service()
	svcExample.ResourceVersion = "3"
	_, err = KubeClient.CoreV1().Services("default").Update(context.TODO(), svcExample, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in adding Service: %v", err)
	}
	g.Eventually(func() []avinodes.AviPoolMetaServer {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		vs := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return vs[0].PoolRefs[0].Servers
	}, 60*time.Second).Should(gomega.HaveLen(0))
	// If we transition the service from Loadbalancer to clusterIP - pools' servers should get populated.
	svcExample = (integrationtest.FakeService{
		Name:         svcName,
		Namespace:    "default",
		Type:         corev1.ServiceTypeClusterIP,
		ServicePorts: []integrationtest.Serviceport{{PortName: "foo0", Protocol: "TCP", PortNumber: 8080, TargetPort: intstr.FromInt(8080)}},
	}).Service()
	svcExample.ResourceVersion = "4"
	_, err = KubeClient.CoreV1().Services("default").Update(context.TODO(), svcExample, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in adding Service: %v", err)
	}
	g.Eventually(func() []avinodes.AviPoolMetaServer {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		vs := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return vs[0].PoolRefs[0].Servers
	}, 15*time.Second).Should(gomega.HaveLen(2))
	TearDownIngressForCacheSyncCheck(t, ingName, svcName, "", modelName)
}

func TestDeletePoolCacheSync(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	var err error

	modelName := MODEL_NAME_PREFIX + "0"
	modelName1 := MODEL_NAME_PREFIX + "1"
	svcName := objNameMap.GenerateName("avisvc")
	ingName := objNameMap.GenerateName("foo-with-targets")
	ingTestObj := IngressTestObject{
		ingressName: ingName,
		isTLS:       false,
		withSecret:  false,
		serviceName: svcName,
		modelNames:  []string{modelName},
	}
	ingTestObj.FillParams()
	SetUpIngressForCacheSyncCheck(t, ingTestObj)

	ingressUpdate := (integrationtest.FakeIngress{
		Name:        ingName,
		Namespace:   "default",
		DnsNames:    []string{"bar.com"},
		Ips:         []string{"8.8.8.8"},
		ServiceName: svcName,
	}).Ingress()
	ingressUpdate.ResourceVersion = "2"
	if _, err = KubeClient.NetworkingV1().Ingresses("default").Update(context.TODO(), ingressUpdate, metav1.UpdateOptions{}); err != nil {
		t.Fatalf("error in updating Ingress: %v", err)
	}
	integrationtest.PollForCompletion(t, modelName1, 5)

	g.Eventually(func() string {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName1)
		if vs := aviModel.(*avinodes.AviObjectGraph).GetAviVS(); len(vs) > 0 && len(vs[0].PoolRefs) > 0 {
			return vs[0].PoolRefs[0].Name
		}
		return ""
	}, 5*time.Second).Should(gomega.ContainSubstring("bar.com"))

	// check that old pool is deleted and new one is created, will have different names
	oldPoolKey := cache.NamespaceName{Namespace: integrationtest.AVINAMESPACE, Name: "cluster--foo.com_foo-default-" + ingName}
	newPoolKey := cache.NamespaceName{Namespace: integrationtest.AVINAMESPACE, Name: "cluster--bar.com_foo-default-" + ingName}
	mcache := cache.SharedAviObjCache()
	g.Eventually(func() bool {
		_, found := mcache.PoolCache.AviCacheGet(oldPoolKey)
		return found
	}, 5*time.Second).Should(gomega.Equal(false))
	g.Eventually(func() bool {
		_, found := mcache.PoolCache.AviCacheGet(newPoolKey)
		return found
	}, 10*time.Second).Should(gomega.Equal(true))
	newPoolCache, _ := mcache.PoolCache.AviCacheGet(newPoolKey)
	newPoolCacheObj, _ := newPoolCache.(*cache.AviPoolCache)
	g.Expect(newPoolCacheObj.Name).To(gomega.Not(gomega.ContainSubstring("foo.com")))
	g.Expect(newPoolCacheObj.Name).To(gomega.ContainSubstring("bar.com"))

	TearDownIngressForCacheSyncCheck(t, ingName, svcName, "", modelName)
}

func TestCreateSNICacheSync(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName := MODEL_NAME_PREFIX + "0"
	svcName := objNameMap.GenerateName("avisvc")
	secretName := objNameMap.GenerateName("my-secret")
	ingName := objNameMap.GenerateName("foo-with-targets")
	ingTestObj := IngressTestObject{
		ingressName: ingName,
		isTLS:       true,
		withSecret:  true,
		secretName:  secretName,
		serviceName: svcName,
		modelNames:  []string{modelName},
	}
	ingTestObj.FillParams()
	SetUpIngressForCacheSyncCheck(t, ingTestObj)

	mcache := cache.SharedAviObjCache()
	parentVSKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--Shared-L7-0"}
	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--foo.com"}

	g.Eventually(func() bool {
		_, found := mcache.VsCacheMeta.AviCacheGet(sniVSKey)
		return found
	}, 20*time.Second).Should(gomega.Equal(true))
	parentCache, _ := mcache.VsCacheMeta.AviCacheGet(parentVSKey)
	parentCacheObj, _ := parentCache.(*cache.AviVsCache)
	g.Expect(parentCacheObj.SNIChildCollection).To(gomega.HaveLen(1))
	g.Expect(parentCacheObj.SNIChildCollection[0]).To(gomega.ContainSubstring("cluster--foo.com"))
	g.Expect(parentCacheObj.HTTPKeyCollection).To(gomega.HaveLen(1))

	sniCache, _ := mcache.VsCacheMeta.AviCacheGet(sniVSKey)
	sniCacheObj, _ := sniCache.(*cache.AviVsCache)
	g.Expect(sniCacheObj.SSLKeyCertCollection).To(gomega.HaveLen(1))
	g.Expect(sniCacheObj.SSLKeyCertCollection[0].Name).To(gomega.ContainSubstring("cluster--foo.com"))
	g.Expect(sniCacheObj.HTTPKeyCollection).To(gomega.HaveLen(1))
	g.Expect(sniCacheObj.HTTPKeyCollection[0].Name).To(gomega.ContainSubstring("cluster--default-foo.com"))
	g.Expect(sniCacheObj.ParentVSRef).To(gomega.Equal(parentVSKey))

	TearDownIngressForCacheSyncCheck(t, ingName, svcName, secretName, modelName)
}

func TestUpdateSNICacheSync(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	var err error

	modelName := MODEL_NAME_PREFIX + "0"
	svcName := objNameMap.GenerateName("avisvc")
	secretName := objNameMap.GenerateName("my-secret")
	ingName := objNameMap.GenerateName("foo-with-targets")
	ingTestObj := IngressTestObject{
		ingressName: ingName,
		isTLS:       true,
		withSecret:  true,
		secretName:  secretName,
		serviceName: svcName,
		modelNames:  []string{modelName},
	}
	ingTestObj.FillParams()
	SetUpIngressForCacheSyncCheck(t, ingTestObj)

	mcache := cache.SharedAviObjCache()
	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--foo.com"}
	g.Eventually(func() bool {
		_, found := mcache.VsCacheMeta.AviCacheGet(sniVSKey)
		return found
	}, 15*time.Second).Should(gomega.Equal(true))
	oldSniCache, _ := mcache.VsCacheMeta.AviCacheGet(sniVSKey)
	oldSniCacheObj, _ := oldSniCache.(*cache.AviVsCache)

	ingressUpdate := (integrationtest.FakeIngress{
		Name:        ingName,
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/bar-updated"},
		ServiceName: svcName,
		TlsSecretDNS: map[string][]string{
			secretName: {"foo.com"},
		},
	}).Ingress()
	ingressUpdate.ResourceVersion = "2"
	_, err = KubeClient.NetworkingV1().Ingresses("default").Update(context.TODO(), ingressUpdate, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating Ingress: %v", err)
	}

	// verify that a NEW httppolicy set object is created
	oldHttpPolKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--default-foo.com"}
	newHttpPolKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--default-foo.com"}
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
	g.Expect(sniVSCacheObj.HTTPKeyCollection).To(gomega.HaveLen(1))
	g.Expect(sniVSCacheObj.SSLKeyCertCollection).To(gomega.HaveLen(1))

	TearDownIngressForCacheSyncCheck(t, ingName, svcName, secretName, modelName)
}

func TestMultiHostMultiSecretSNICacheSync(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName := MODEL_NAME_PREFIX + "0"
	svcName := objNameMap.GenerateName("avisvc")
	secretName := objNameMap.GenerateName("my-secret")
	secretName2 := objNameMap.GenerateName("my-secret")
	ingName := objNameMap.GenerateName("foo-with-targets")
	ingName2 := objNameMap.GenerateName("foo-with-targets")
	ingTestObj := IngressTestObject{
		ingressName: ingName,
		isTLS:       true,
		withSecret:  true,
		secretName:  secretName,
		serviceName: svcName,
		modelNames:  []string{modelName},
	}
	ingTestObj.FillParams()
	SetUpIngressForCacheSyncCheck(t, ingTestObj)
	mcache := cache.SharedAviObjCache()
	integrationtest.AddSecret(secretName, "default", "tlsCert", "tlsKey")
	// update ingress
	ingressObject := integrationtest.FakeIngress{
		Name:        ingName,
		Namespace:   "default",
		DnsNames:    []string{"foo.com", "bar.com"},
		Paths:       []string{"/foo", "/bar"},
		ServiceName: svcName,
		TlsSecretDNS: map[string][]string{
			secretName:  {"foo.com"},
			secretName2: {"bar.com"},
		},
	}
	integrationtest.AddSecret(secretName2, "default", "tlsCert", "tlsKey")
	ingrFake := ingressObject.Ingress()
	ingrFake.ResourceVersion = "2"
	if _, err := KubeClient.NetworkingV1().Ingresses("default").Update(context.TODO(), ingrFake, metav1.UpdateOptions{}); err != nil {
		t.Fatalf("error in updating Ingress: %v", err)
	}

	sniVSKey1 := cache.NamespaceName{Namespace: "admin", Name: "cluster--foo.com"}
	sniVSKey2 := cache.NamespaceName{Namespace: "admin", Name: "cluster--bar.com"}
	g.Eventually(func() bool {
		sniCache1, found1 := mcache.VsCacheMeta.AviCacheGet(sniVSKey1)
		sniCache2, found2 := mcache.VsCacheMeta.AviCacheGet(sniVSKey2)
		sniCacheObj1, _ := sniCache1.(*cache.AviVsCache)
		sniCacheObj2, _ := sniCache2.(*cache.AviVsCache)
		if found1 && found2 &&
			len(sniCacheObj1.SSLKeyCertCollection) == 1 &&
			len(sniCacheObj2.SSLKeyCertCollection) == 1 && len(sniCacheObj1.HTTPKeyCollection) == 1 {
			return true
		}
		return false
	}, 20*time.Second).Should(gomega.Equal(true))

	sniCache1, _ := mcache.VsCacheMeta.AviCacheGet(sniVSKey1)
	sniCacheObj1, _ := sniCache1.(*cache.AviVsCache)
	sniCache2, _ := mcache.VsCacheMeta.AviCacheGet(sniVSKey2)
	sniCacheObj2, _ := sniCache2.(*cache.AviVsCache)
	g.Expect(sniCacheObj1.SSLKeyCertCollection[0].Name).To(gomega.Equal("cluster--foo.com"))
	g.Expect(sniCacheObj2.SSLKeyCertCollection[0].Name).To(gomega.Equal("cluster--bar.com"))

	g.Eventually(func() string {
		sniCache1, _ := mcache.VsCacheMeta.AviCacheGet(sniVSKey1)
		sniCacheObj1, _ := sniCache1.(*cache.AviVsCache)
		return sniCacheObj1.ParentVSRef.Name
	}, 15*time.Second).Should(gomega.Not(gomega.Equal("")))
	ingressObject = integrationtest.FakeIngress{
		Name:        ingName2,
		Namespace:   "red",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/doo"},
		ServiceName: svcName,
		TlsSecretDNS: map[string][]string{
			secretName: {"foo.com"},
		},
	}
	integrationtest.AddSecret(secretName, "red", "tlsCert", "tlsKey")
	ingrFake = ingressObject.Ingress()
	if _, err := KubeClient.NetworkingV1().Ingresses("red").Create(context.TODO(), ingrFake, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in updating Ingress: %v", err)
	}
	g.Eventually(func() bool {
		sniCache1, found1 := mcache.VsCacheMeta.AviCacheGet(sniVSKey1)
		sniCacheObj1, _ := sniCache1.(*cache.AviVsCache)
		if found1 &&
			len(sniCacheObj1.SSLKeyCertCollection) == 1 && len(sniCacheObj1.HTTPKeyCollection) == 2 {
			return true
		}
		return false
	}, 20*time.Second).Should(gomega.Equal(true))
	if err := KubeClient.NetworkingV1().Ingresses("red").Delete(context.TODO(), ingName2, metav1.DeleteOptions{}); err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	TearDownIngressForCacheSyncCheck(t, ingName, svcName, secretName, modelName)
}

func TestMultiHostMultiSecretUpdateSNICacheSync(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	modelName := MODEL_NAME_PREFIX + "0"
	svcName := objNameMap.GenerateName("avisvc")
	secretName := objNameMap.GenerateName("my-secret")
	secretName2 := objNameMap.GenerateName("my-secret")
	ingName := objNameMap.GenerateName("foo-with-targets")

	SetupDomain()
	SetUpTestForIngress(t, svcName, integrationtest.AllModels...)
	integrationtest.PollForCompletion(t, modelName, 5)
	ingressObject := integrationtest.FakeIngress{
		Name:        ingName,
		Namespace:   "default",
		DnsNames:    []string{"foo.com", "bar.com", "xyz.com"},
		Paths:       []string{"/foo", "/bar", "/xyz"},
		ServiceName: svcName,
		TlsSecretDNS: map[string][]string{
			secretName:  {"foo.com"},
			secretName2: {"bar.com"},
		},
	}
	integrationtest.AddSecret(secretName2, "default", "tlsCert", "tlsKey")
	integrationtest.AddSecret(secretName, "default", "tlsCert", "tlsKey")

	ingrFake := ingressObject.Ingress()
	if _, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	integrationtest.PollForCompletion(t, modelName, 5)

	mcache := cache.SharedAviObjCache()
	sniVSKey1 := cache.NamespaceName{Namespace: "admin", Name: "cluster--foo.com"}
	sniVSKey2 := cache.NamespaceName{Namespace: "admin", Name: "cluster--bar.com"}

	// Shard scheme: cluster--Shared-L7-0 -> foo.com
	// Shard scheme: cluster--Shared-L7-3 -> xyz.com
	xyzParentKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--Shared-L7-3"}
	g.Eventually(func() int {
		sniCache, _ := mcache.VsCacheMeta.AviCacheGet(xyzParentKey)
		sniCacheObj, _ := sniCache.(*cache.AviVsCache)
		return len(sniCacheObj.PoolKeyCollection)
	}, 10*time.Second).Should(gomega.Equal(1))
	sniCache, _ := mcache.VsCacheMeta.AviCacheGet(xyzParentKey)
	sniCacheObj, _ := sniCache.(*cache.AviVsCache)
	g.Expect(sniCacheObj.PoolKeyCollection[0].Name).To(gomega.ContainSubstring("xyz.com"))

	g.Eventually(func() bool {
		_, found1 := mcache.VsCacheMeta.AviCacheGet(sniVSKey1)
		_, found2 := mcache.VsCacheMeta.AviCacheGet(sniVSKey2)
		if found1 && found2 {
			return true
		}
		return false
	}, 5*time.Second).Should(gomega.Equal(true))

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
	g.Expect(sniCacheObj.PoolKeyCollection[0].Name).To(gomega.ContainSubstring("foo.com"))
	g.Expect(sniCacheObj.SSLKeyCertCollection).To(gomega.HaveLen(1))
	g.Expect(sniCacheObj.SSLKeyCertCollection[0].Name).To(gomega.Equal("cluster--foo.com"))

	g.Eventually(func() int {
		sniCache, found := mcache.VsCacheMeta.AviCacheGet(sniVSKey2)
		sniCacheObj, ok := sniCache.(*cache.AviVsCache)
		if found && ok {
			return len(sniCacheObj.PoolKeyCollection)
		}
		return 0
	}, 10*time.Second).Should(gomega.Equal(1))
	sniCache, _ = mcache.VsCacheMeta.AviCacheGet(sniVSKey2)
	sniCacheObj, _ = sniCache.(*cache.AviVsCache)
	g.Expect(sniCacheObj.PoolKeyCollection[0].Name).To(gomega.ContainSubstring("bar.com"))
	g.Expect(sniCacheObj.SSLKeyCertCollection).To(gomega.HaveLen(1))
	g.Expect(sniCacheObj.SSLKeyCertCollection[0].Name).To(gomega.Equal("cluster--bar.com"))

	// delete one secret
	KubeClient.CoreV1().Secrets("default").Delete(context.TODO(), secretName2, metav1.DeleteOptions{})
	ingressUpdateObject := integrationtest.FakeIngress{
		Name:        ingName,
		Namespace:   "default",
		DnsNames:    []string{"foo.com", "bar.com", "xyz.com"},
		Paths:       []string{"/foo", "/bar", "/xyz"},
		ServiceName: svcName,
		TlsSecretDNS: map[string][]string{
			secretName: {"foo.com"},
		},
	}

	ingrUpdate := ingressUpdateObject.Ingress()
	ingrUpdate.ResourceVersion = "2"
	if _, err := KubeClient.NetworkingV1().Ingresses("default").Update(context.TODO(), ingrUpdate, metav1.UpdateOptions{}); err != nil {
		t.Fatalf("error in updating Ingress: %v", err)
	}

	// Shard scheme: cluster--Shared-L7-1 -> bar.com
	barParentKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--Shared-L7-1"}
	g.Eventually(func() int {
		sniCache, _ := mcache.VsCacheMeta.AviCacheGet(barParentKey)
		sniCacheObj, _ := sniCache.(*cache.AviVsCache)
		return len(sniCacheObj.PoolKeyCollection)
	}, 10*time.Second).Should(gomega.Equal(1))

	// should not be found
	g.Eventually(func() bool {
		_, found := mcache.VsCacheMeta.AviCacheGet(sniVSKey2)
		return found
	}, 10*time.Second).Should(gomega.Equal(false))

	g.Eventually(func() bool {
		_, found := mcache.VsCacheMeta.AviCacheGet(sniVSKey1)
		return found
	}, 10*time.Second).Should(gomega.Equal(true))
	sniCache, _ = mcache.VsCacheMeta.AviCacheGet(sniVSKey1)
	sniCacheObj, _ = sniCache.(*cache.AviVsCache)
	g.Expect(sniCacheObj.PoolKeyCollection).To(gomega.HaveLen(1))
	g.Expect(sniCacheObj.PoolKeyCollection[0].Name).To(gomega.ContainSubstring("foo.com"))
	g.Expect(sniCacheObj.SSLKeyCertCollection).To(gomega.HaveLen(1))
	g.Expect(sniCacheObj.SSLKeyCertCollection[0].Name).To(gomega.Equal("cluster--foo.com"))

	KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), ingName, metav1.DeleteOptions{})
	KubeClient.CoreV1().Secrets("default").Delete(context.TODO(), secretName, metav1.DeleteOptions{})
	g.Eventually(func() bool {
		_, found := mcache.VsCacheMeta.AviCacheGet(sniVSKey1)
		return found
	}, 15*time.Second).Should(gomega.Equal(false))
	TearDownTestForIngress(t, svcName, modelName)
}

func TestDeleteSNICacheSync(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	var err error

	modelName := MODEL_NAME_PREFIX + "0"
	svcName := objNameMap.GenerateName("avisvc")
	secretName := objNameMap.GenerateName("my-secret")
	ingName := objNameMap.GenerateName("foo-with-targets")
	ingTestObj := IngressTestObject{
		ingressName: ingName,
		isTLS:       true,
		withSecret:  true,
		secretName:  secretName,
		serviceName: svcName,
		modelNames:  []string{modelName},
	}
	ingTestObj.FillParams()
	SetUpIngressForCacheSyncCheck(t, ingTestObj)

	mcache := cache.SharedAviObjCache()
	parentVSKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--Shared-L7-0"}
	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--foo.com"}

	ingressUpdate := (integrationtest.FakeIngress{
		Name:        ingName,
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		Paths:       []string{"/foo"},
		ServiceName: svcName,
	}).Ingress()
	ingressUpdate.ResourceVersion = "2"
	_, err = KubeClient.NetworkingV1().Ingresses("default").Update(context.TODO(), ingressUpdate, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating Ingress: %v", err)
	}

	// verify that sni vs is deleted, but the parent vs is not
	// deleted snivs key should be deleted from parent vs snichildcollection
	g.Eventually(func() bool {
		_, found := mcache.VsCacheMeta.AviCacheGet(sniVSKey)
		parentSniCache, _ := mcache.VsCacheMeta.AviCacheGet(parentVSKey)
		parentSniCacheObj, _ := parentSniCache.(*cache.AviVsCache)

		if !found && len(parentSniCacheObj.SNIChildCollection) == 0 && len(parentSniCacheObj.HTTPKeyCollection) == 0 {
			return true
		}
		return false
	}, 20*time.Second).Should(gomega.Equal(true))

	TearDownIngressForCacheSyncCheck(t, ingName, svcName, secretName, modelName)
}

func TestCUDSecretCacheSync(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName := MODEL_NAME_PREFIX + "0"
	svcName := objNameMap.GenerateName("avisvc")
	secretName := objNameMap.GenerateName("my-secret")
	ingName := objNameMap.GenerateName("foo-with-targets")
	ingTestObj := IngressTestObject{
		ingressName: ingName,
		isTLS:       true,
		withSecret:  false,
		secretName:  secretName,
		serviceName: svcName,
		modelNames:  []string{modelName},
	}
	ingTestObj.FillParams()
	SetUpIngressForCacheSyncCheck(t, ingTestObj)

	mcache := cache.SharedAviObjCache()
	parentVSKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--Shared-L7-0"}
	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--foo.com"}
	sslKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--foo.com"}

	// no ssl key cache would be found since the secret is not yet added
	g.Eventually(func() bool {
		_, found := mcache.SSLKeyCache.AviCacheGet(sslKey)
		return found
	}, 5*time.Second).Should(gomega.Equal(false))

	// add Secret

	integrationtest.AddSecret(secretName, "default", "tlsCert", "tlsKey")

	// ssl key should be created now and must be attached to the sni vs cache
	g.Eventually(func() bool {
		_, found := mcache.SSLKeyCache.AviCacheGet(sslKey)
		return found
	}, 30*time.Second).Should(gomega.Equal(true))
	sniVSCache, _ := mcache.VsCacheMeta.AviCacheGet(sniVSKey)
	sniVSCacheObj, _ := sniVSCache.(*cache.AviVsCache)
	g.Expect(sniVSCacheObj.SSLKeyCertCollection).To(gomega.HaveLen(1))
	parentVSCache, _ := mcache.VsCacheMeta.AviCacheGet(parentVSKey)
	parentVSCacheObj, _ := parentVSCache.(*cache.AviVsCache)
	g.Expect(parentVSCacheObj.HTTPKeyCollection).To(gomega.HaveLen(1))

	// update Secret
	secretUpdate := (integrationtest.FakeSecret{
		Namespace: "default",
		Name:      secretName,
		Cert:      "tlsCert_Updated",
		Key:       "tlsKey_Updated",
	}).Secret()
	secretUpdate.ResourceVersion = "2"
	KubeClient.CoreV1().Secrets("default").Update(context.TODO(), secretUpdate, metav1.UpdateOptions{})

	// can't check update rn, ssl cache object doesnot have checksum,
	// but PUTs happen, everytime though

	// delete Secret
	KubeClient.CoreV1().Secrets("default").Delete(context.TODO(), secretName, metav1.DeleteOptions{})

	// ssl key must be deleted again and sni vs as well
	g.Eventually(func() bool {
		_, found := mcache.SSLKeyCache.AviCacheGet(sslKey)
		return found
	}, 5*time.Second).Should(gomega.Equal(false))
	_, found := mcache.VsCacheMeta.AviCacheGet(sniVSKey)
	g.Expect(found).To(gomega.Equal(false))

	g.Eventually(func() bool {
		parentVSCache, found := mcache.VsCacheMeta.AviCacheGet(parentVSKey)
		parentVSCacheObj, ok := parentVSCache.(*cache.AviVsCache)
		if found && ok && len(parentVSCacheObj.HTTPKeyCollection) == 0 {
			return true
		}
		return false
	}, 10*time.Second).Should(gomega.Equal(true))

	TearDownIngressForCacheSyncCheck(t, ingName, svcName, secretName, modelName)
}

func TestDeleteSecretSecureIngressStatusCheck(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	modelName := MODEL_NAME_PREFIX + "0"
	svcName := objNameMap.GenerateName("avisvc")
	secretName := objNameMap.GenerateName("my-secret")
	ingName := objNameMap.GenerateName("foo-with-targets")
	ingTestObj := IngressTestObject{
		ingressName: ingName,
		isTLS:       true,
		withSecret:  true,
		secretName:  secretName,
		serviceName: svcName,
		modelNames:  []string{modelName},
	}
	ingTestObj.FillParams()
	SetUpIngressForCacheSyncCheck(t, ingTestObj)

	g.Eventually(func() int {
		ingress, _ := KubeClient.NetworkingV1().Ingresses("default").Get(context.TODO(), ingName, metav1.GetOptions{})
		return len(ingress.Status.LoadBalancer.Ingress)
	}, 30*time.Second).Should(gomega.Equal(1))

	// post this SNI VS should get deleted, and ingress status must be updated accordingly
	KubeClient.CoreV1().Secrets("default").Delete(context.TODO(), secretName, metav1.DeleteOptions{})

	g.Eventually(func() int {
		ingress, _ := KubeClient.NetworkingV1().Ingresses("default").Get(context.TODO(), ingName, metav1.GetOptions{})
		return len(ingress.Status.LoadBalancer.Ingress)
	}, 50*time.Second).Should(gomega.Equal(0))

	TearDownIngressForCacheSyncCheck(t, ingName, svcName, secretName, modelName)
}

func TestMultiHostIngressStatusCheck(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	modelName := MODEL_NAME_PREFIX + "0"
	svcName := objNameMap.GenerateName("avisvc")
	secretName := objNameMap.GenerateName("my-secret")
	secretName2 := objNameMap.GenerateName("my-secret")
	ingName := objNameMap.GenerateName("foo-with-targets")
	ingName2 := objNameMap.GenerateName("foo-with-targets")

	SetupDomain()
	SetUpTestForIngress(t, svcName, integrationtest.AllModels...)
	integrationtest.PollForCompletion(t, modelName, 5)
	ingressObject := integrationtest.FakeIngress{
		Name:        ingName,
		Namespace:   "default",
		DnsNames:    []string{"foo.com", "bar.com", "xyz.com"},
		Paths:       []string{"/foo", "/bar", "/xyz"},
		ServiceName: svcName,
		TlsSecretDNS: map[string][]string{
			secretName:  {"foo.com"},
			secretName2: {"bar.com"},
		},
	}
	integrationtest.AddSecret(secretName, "default", "tlsCert", "tlsKey")
	integrationtest.AddSecret(secretName2, "default", "tlsCert", "tlsKey")
	time.Sleep(10 * time.Second) //secrets should be created before creating ingress
	ingressObject2 := integrationtest.FakeIngress{
		Name:        ingName2,
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/doo"},
		ServiceName: svcName,
		TlsSecretDNS: map[string][]string{
			secretName: {"foo.com"},
		},
	}

	ingrFake := ingressObject.Ingress()
	if _, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	time.Sleep(2 * time.Second)
	ingrFake_2 := ingressObject2.Ingress()
	if _, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake_2, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	integrationtest.PollForCompletion(t, modelName, 5)

	// Shard scheme: cluster--Shared-L7-0 -> foo.com
	// Shard scheme: cluster--Shared-L7-3 -> xyz.com
	// Shard scheme: cluster--Shared-L7-1 -> bar.com

	g.Eventually(func() int {
		ingress, _ := KubeClient.NetworkingV1().Ingresses("default").Get(context.TODO(), ingName, metav1.GetOptions{})
		return len(ingress.Status.LoadBalancer.Ingress)
	}, 50*time.Second).Should(gomega.Equal(3))
	ingress, _ := KubeClient.NetworkingV1().Ingresses("default").Get(context.TODO(), ingName, metav1.GetOptions{})
	// fake avi controller server returns IP in the form: 10.250.250.1<Shared-L7-NUM>
	g.Expect(ingress.Status.LoadBalancer.Ingress[0].IP).To(gomega.MatchRegexp(`^(10.250.250.1(0|1|3))`))
	g.Expect(ingress.Status.LoadBalancer.Ingress[0].Hostname).To(gomega.MatchRegexp(`^((foo|bar|xyz).com)$`))
	g.Expect(ingress.Status.LoadBalancer.Ingress[1].IP).To(gomega.MatchRegexp(`^(10.250.250.1(0|1|3))`))
	g.Expect(ingress.Status.LoadBalancer.Ingress[1].Hostname).To(gomega.MatchRegexp(`^((foo|bar|xyz).com)$`))
	g.Expect(ingress.Status.LoadBalancer.Ingress[2].IP).To(gomega.MatchRegexp(`^(10.250.250.1(0|1|3))`))
	g.Expect(ingress.Status.LoadBalancer.Ingress[2].Hostname).To(gomega.MatchRegexp(`^((foo|bar|xyz).com)$`))

	if err := KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), ingName, metav1.DeleteOptions{}); err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}

	// Check if the other ingress also updated or not.
	g.Eventually(func() int {
		ingress, _ := KubeClient.NetworkingV1().Ingresses("default").Get(context.TODO(), ingName2, metav1.GetOptions{})
		return len(ingress.Status.LoadBalancer.Ingress)
	}, 50*time.Second).Should(gomega.Equal(1))
	ingress, _ = KubeClient.NetworkingV1().Ingresses("default").Get(context.TODO(), ingName2, metav1.GetOptions{})
	g.Expect(ingress.Status.LoadBalancer.Ingress[0].IP).To(gomega.MatchRegexp(`^(10.250.250.1(0|1|3))`))
	g.Expect(ingress.Status.LoadBalancer.Ingress[0].Hostname).To(gomega.MatchRegexp(`^((foo).com)$`))
	if err := KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), ingName2, metav1.DeleteOptions{}); err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}

	TearDownTestForIngress(t, svcName, modelName)
}

func TestMultiHostUpdateIngressStatusCheck(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	var err error
	modelName := MODEL_NAME_PREFIX + "0"
	svcName := objNameMap.GenerateName("avisvc")
	secretName := objNameMap.GenerateName("my-secret")
	ingressId := "thmhuisc"
	ingressName := fmt.Sprintf("ing-%s", ingressId)
	pathSuffix := "-" + ingressName + ".com"

	SetupDomain()
	SetUpTestForIngress(t, svcName, integrationtest.AllModels...)
	integrationtest.PollForCompletion(t, modelName, 5)
	ingressObject := integrationtest.FakeIngress{
		Name:        ingressName,
		Namespace:   "default",
		DnsNames:    []string{"foo" + pathSuffix, "xyz" + pathSuffix},
		Paths:       []string{"/foo", "/xyz"},
		ServiceName: svcName,
		TlsSecretDNS: map[string][]string{
			secretName: {"foo" + pathSuffix},
		},
	}
	integrationtest.AddSecret(secretName, "default", "tlsCert", "tlsKey")
	ingrFake := ingressObject.Ingress()
	if _, err = KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	integrationtest.PollForCompletion(t, modelName, 5)

	g.Eventually(func() int {
		ingress, _ := KubeClient.NetworkingV1().Ingresses("default").Get(context.TODO(), ingressName, metav1.GetOptions{})
		return len(ingress.Status.LoadBalancer.Ingress)
	}, 60*time.Second).Should(gomega.Equal(2))
	ingress, _ := KubeClient.NetworkingV1().Ingresses("default").Get(context.TODO(), ingressName, metav1.GetOptions{})

	// donot update status
	var ingressStatusIPs, ingressStatusNames []string
	for _, i := range ingress.Status.LoadBalancer.Ingress {
		ingressStatusIPs = append(ingressStatusIPs, i.IP)
		ingressStatusNames = append(ingressStatusNames, i.Hostname)
	}

	// remove one hostname
	ingressUpdate := (integrationtest.FakeIngress{
		Name:        ingressName,
		Namespace:   "default",
		DnsNames:    []string{"foo" + pathSuffix},
		Paths:       []string{"/foo"},
		Ips:         ingressStatusIPs,
		HostNames:   ingressStatusNames,
		ServiceName: svcName,
		TlsSecretDNS: map[string][]string{
			secretName: {"foo" + pathSuffix},
		},
	}).Ingress()
	ingressUpdate.ResourceVersion = "2"
	_, err = KubeClient.NetworkingV1().Ingresses("default").Update(context.TODO(), ingressUpdate, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating Ingress: %v", err)
	}

	// statuses go down from 3 to 2
	g.Eventually(func() int {
		ingress, _ := KubeClient.NetworkingV1().Ingresses("default").Get(context.TODO(), ingressName, metav1.GetOptions{})
		return len(ingress.Status.LoadBalancer.Ingress)
	}, 10*time.Second).Should(gomega.Equal(2))

	if err := KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), ingressName, metav1.DeleteOptions{}); err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	TearDownTestForIngress(t, svcName, modelName)
}

func TestProfilesAttachedToVS(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	// middleware verifies the application and network profiles attached to the VS
	integrationtest.AddMiddleware(func(w http.ResponseWriter, r *http.Request) {
		url := r.URL.EscapedPath()
		if r.Method == http.MethodPost &&
			strings.Contains(url, "/api/virtualservice") {
			var resp map[string]interface{}
			data, _ := io.ReadAll(r.Body)
			json.Unmarshal(data, &resp)
			if strings.Contains(resp["name"].(string), "Shared") {
				g.Expect(resp["application_profile_ref"]).Should(gomega.HaveSuffix("System-HTTP"))
				g.Expect(resp["network_profile_ref"]).Should(gomega.HaveSuffix("System-TCP-Proxy"))
			} else {
				g.Expect(resp["application_profile_ref"]).Should(gomega.HaveSuffix("System-Secure-HTTP"))
				g.Expect(resp["network_profile_ref"]).Should(gomega.HaveSuffix("System-TCP-Proxy"))
			}
			resp["uuid"] = "virtualservice--name-RANDOMUUID"
			finalResponse, _ := json.Marshal(resp)
			w.WriteHeader(http.StatusOK)
			w.Write(finalResponse)
			return
		}
		integrationtest.NormalControllerServer(w, r)
	})

	modelName := MODEL_NAME_PREFIX + "0"
	svcName := objNameMap.GenerateName("avisvc")
	secretName := objNameMap.GenerateName("my-secret")
	ingName := objNameMap.GenerateName("foo-with-targets")
	ingTestObj := IngressTestObject{
		ingressName: ingName,
		isTLS:       true,
		withSecret:  true,
		secretName:  secretName,
		serviceName: svcName,
		modelNames:  []string{modelName},
	}
	ingTestObj.FillParams()
	SetUpIngressForCacheSyncCheck(t, ingTestObj)
	g.Eventually(func() int {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes, ok := aviModel.(*avinodes.AviObjectGraph)
		if !ok {
			return 0
		}
		return len(nodes.GetAviVS())
	}, 30*time.Second).Should(gomega.Equal(1))

	mcache := cache.SharedAviObjCache()
	parentVSKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--Shared-L7-0"}
	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--foo.com"}

	g.Eventually(func() bool {
		_, found := mcache.VsCacheMeta.AviCacheGet(parentVSKey)
		return found
	}, 60*time.Second).Should(gomega.Equal(true))

	g.Eventually(func() bool {
		_, found := mcache.VsCacheMeta.AviCacheGet(sniVSKey)
		return found
	}, 60*time.Second).Should(gomega.Equal(true))

	TearDownIngressForCacheSyncCheck(t, ingName, svcName, secretName, modelName)

	integrationtest.ResetMiddleware()
}
