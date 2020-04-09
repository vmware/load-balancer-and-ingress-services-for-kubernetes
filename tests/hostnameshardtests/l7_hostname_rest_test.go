package hostnameshardtests

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"ako/pkg/cache"
	avinodes "ako/pkg/nodes"
	"ako/pkg/objects"
	"ako/tests/integrationtest"

	"github.com/avinetworks/container-lib/utils"
	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestHostnameCacheGETOKStatus(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	// Verify the cache.
	cacheobj := cache.SharedAviObjCache()
	vsKey := cache.NamespaceName{Namespace: "admin", Name: "Shard-VS-5"}
	vs_cache, found := cacheobj.VsCache.AviCacheGet(vsKey)

	if !found {
		t.Fatalf("Cache not found for VS: %v", vsKey)
	} else {
		vs_cache_obj, ok := vs_cache.(*cache.AviVsCache)
		if !ok {
			t.Fatalf("Invalid VS object. Cannot cast.")
		}
		g.Expect(len(vs_cache_obj.PoolKeyCollection)).To(gomega.Equal(3))
		g.Expect(len(vs_cache_obj.PGKeyCollection)).To(gomega.Equal(1))
		g.Expect(len(vs_cache_obj.DSKeyCollection)).To(gomega.Equal(1))
	}
}

func SetUpIngressForCacheSyncCheck(t *testing.T, modelName string, tlsIngress, withSecret bool) {
	SetUpTestForIngress(t, modelName)
	integrationtest.PollForCompletion(t, modelName, 5)
	ingressObject := integrationtest.FakeIngress{
		Name:        "foo-with-targets",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		Paths:       []string{"/foo"},
		ServiceName: "avisvc",
	}
	if withSecret {
		integrationtest.AddSecret("my-secret", "default")
	}
	if tlsIngress {
		ingressObject.TlsSecretDNS = map[string][]string{
			"my-secret": []string{"foo.com"},
		}
	}
	ingrFake := ingressObject.Ingress()
	if _, err := KubeClient.ExtensionsV1beta1().Ingresses("default").Create(ingrFake); err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	integrationtest.PollForCompletion(t, modelName, 5)
}

func VerifyCacheRemoval(modelName string, g *gomega.GomegaWithT) {
	mcache := cache.SharedAviObjCache()
	vsKey := cache.NamespaceName{Namespace: "admin", Name: strings.Split(modelName, "/")[1]}
	g.Eventually(func() int {
		vsCache, found := mcache.VsCache.AviCacheGet(vsKey)
		vsCacheObj, ok := vsCache.(*cache.AviVsCache)
		if found && ok {
			return len(vsCacheObj.PoolKeyCollection)
		}
		return -1
	}, 5*time.Second).Should(gomega.Equal(0))
}

func TearDownIngressForCacheSyncCheck(t *testing.T, modelName string) {
	if err := KubeClient.ExtensionsV1beta1().Ingresses("default").Delete("foo-with-targets", nil); err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	KubeClient.CoreV1().Secrets("default").Delete("my-secret", nil)
	TearDownTestForIngress(t, modelName)
}

func TestHostnameCreateIngressCacheSync(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	var found bool

	modelName := "admin/Shard-VS---global-0"
	SetUpIngressForCacheSyncCheck(t, modelName, false, false)

	g.Eventually(func() bool {
		found, _ = objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 5*time.Second).Should(gomega.Equal(true))

	mcache := cache.SharedAviObjCache()
	vsKey := cache.NamespaceName{Namespace: "admin", Name: "Shard-VS---global-0"}
	vsCache, found := mcache.VsCache.AviCacheGet(vsKey)
	if !found {
		t.Fatalf("Cache not found for VS: %v", vsKey)
	}
	vsCacheObj, ok := vsCache.(*cache.AviVsCache)
	if !ok {
		t.Fatalf("Invalid VS object. Cannot cast.")
	}
	g.Expect(vsCacheObj.Name).To(gomega.Equal("Shard-VS---global-0"))
	g.Expect(vsCacheObj.PGKeyCollection).To(gomega.HaveLen(1))
	g.Expect(vsCacheObj.PoolKeyCollection).To(gomega.HaveLen(1))
	g.Expect(vsCacheObj.PoolKeyCollection[0].Name).To(gomega.ContainSubstring("foo-with-targets"))
	g.Expect(vsCacheObj.DSKeyCollection).To(gomega.HaveLen(1))
	g.Expect(vsCacheObj.SSLKeyCertCollection).To(gomega.BeNil())

	TearDownIngressForCacheSyncCheck(t, modelName)
}

func TestHostnameCreateIngressWithFaultCacheSync(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	var found bool

	injectFault := true
	integrationtest.AddMiddleware(func(w http.ResponseWriter, r *http.Request) {
		var resp map[string]interface{}
		var finalResponse []byte
		url := r.URL.EscapedPath()

		if strings.Contains(url, "macro") && r.Method == "POST" {
			data, _ := ioutil.ReadAll(r.Body)
			json.Unmarshal(data, &resp)
			rData, rModelName := resp["data"].(map[string]interface{}), strings.ToLower(resp["model_name"].(string))
			fmt.Printf("%s %+v\n", rModelName, resp)
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
	defer integrationtest.ResetMiddleware()

	modelName := "admin/Shard-VS---global-0"
	SetUpIngressForCacheSyncCheck(t, modelName, false, false)

	g.Eventually(func() int {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return len(nodes[0].PoolRefs)
	}, 5*time.Second).Should(gomega.Equal(1))

	mcache := cache.SharedAviObjCache()
	vsKey := cache.NamespaceName{Namespace: "admin", Name: "Shard-VS---global-0"}
	g.Eventually(func() int {
		vsCache, _ := mcache.VsCache.AviCacheGet(vsKey)
		vsCacheObj, _ := vsCache.(*cache.AviVsCache)
		return len(vsCacheObj.PoolKeyCollection)
	}, 5*time.Second).Should(gomega.Equal(1))

	vsCache, found := mcache.VsCache.AviCacheGet(vsKey)
	if !found {
		t.Fatalf("Cache not found for VS: %v", vsKey)
	}
	vsCacheObj, ok := vsCache.(*cache.AviVsCache)
	if !ok {
		t.Fatalf("Invalid VS object. Cannot cast.")
	}
	g.Expect(vsCacheObj.Name).To(gomega.Equal("Shard-VS---global-0"))
	g.Expect(vsCacheObj.PGKeyCollection).To(gomega.HaveLen(1))
	g.Expect(vsCacheObj.PoolKeyCollection).To(gomega.HaveLen(1))
	g.Expect(vsCacheObj.PoolKeyCollection[0].Name).To(gomega.ContainSubstring("foo-with-targets"))
	g.Expect(vsCacheObj.DSKeyCollection).To(gomega.HaveLen(1))
	g.Expect(vsCacheObj.SSLKeyCertCollection).To(gomega.BeNil())

	TearDownIngressForCacheSyncCheck(t, modelName)
}

func TestHostnameUpdatePoolCacheSync(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	var err error

	modelName := "admin/Shard-VS---global-0"
	SetUpIngressForCacheSyncCheck(t, modelName, false, false)

	// Get hold of the pool checksum on CREATE
	poolName := "global--foo.com/foo--default--foo-with-targets"
	mcache := cache.SharedAviObjCache()
	poolKey := cache.NamespaceName{Namespace: integrationtest.AVINAMESPACE, Name: poolName}
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
	if _, err = KubeClient.CoreV1().Endpoints("default").Update(epExample); err != nil {
		t.Fatalf("error in creating Endpoint: %v", err)
	}

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

	TearDownIngressForCacheSyncCheck(t, modelName)
}

func TestHostnameDeletePoolCacheSync(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	var err error

	modelName := "admin/Shard-VS---global-0"
	modelName1 := "admin/Shard-VS---global-1"
	SetUpIngressForCacheSyncCheck(t, modelName, false, false)

	ingressUpdate := (integrationtest.FakeIngress{
		Name:        "foo-with-targets",
		Namespace:   "default",
		DnsNames:    []string{"bar.com"},
		Ips:         []string{"8.8.8.8"},
		ServiceName: "avisvc",
	}).Ingress()
	ingressUpdate.ResourceVersion = "2"
	if _, err = KubeClient.ExtensionsV1beta1().Ingresses("default").Update(ingressUpdate); err != nil {
		t.Fatalf("error in updating Ingress: %v", err)
	}
	integrationtest.PollForCompletion(t, modelName1, 5)

	g.Eventually(func() string {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName1)
		vs := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return vs[0].PoolRefs[0].Name
	}, 5*time.Second).Should(gomega.ContainSubstring("bar.com"))

	// check that old pool is deleted and new one is created, will have different names
	oldPoolKey := cache.NamespaceName{Namespace: integrationtest.AVINAMESPACE, Name: "global--foo.com/foo--default--foo-with-targets"}
	newPoolKey := cache.NamespaceName{Namespace: integrationtest.AVINAMESPACE, Name: "global--bar.com/foo--default--foo-with-targets"}
	mcache := cache.SharedAviObjCache()
	g.Eventually(func() bool {
		_, found := mcache.PoolCache.AviCacheGet(oldPoolKey)
		return found
	}, 5*time.Second).Should(gomega.Equal(false))
	newPoolCache, found := mcache.PoolCache.AviCacheGet(newPoolKey)
	newPoolCacheObj, _ := newPoolCache.(*cache.AviPoolCache)
	g.Expect(found).To(gomega.Equal(true))
	g.Expect(newPoolCacheObj.Name).To(gomega.Not(gomega.ContainSubstring("foo.com")))
	g.Expect(newPoolCacheObj.Name).To(gomega.ContainSubstring("bar.com"))

	TearDownIngressForCacheSyncCheck(t, modelName)
}

func TestHostnameCreateSNICacheSync(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName := "admin/Shard-VS---global-0"
	SetUpIngressForCacheSyncCheck(t, modelName, true, true)

	mcache := cache.SharedAviObjCache()
	parentVSKey := cache.NamespaceName{Namespace: "admin", Name: "Shard-VS---global-0"}
	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: "global--foo-with-targets--default--foo.com"}

	g.Eventually(func() bool {
		_, found := mcache.VsCache.AviCacheGet(sniVSKey)
		return found
	}, 10*time.Second).Should(gomega.Equal(true))
	parentCache, _ := mcache.VsCache.AviCacheGet(parentVSKey)
	parentCacheObj, _ := parentCache.(*cache.AviVsCache)
	g.Expect(parentCacheObj.SNIChildCollection).To(gomega.HaveLen(1))
	g.Expect(parentCacheObj.SNIChildCollection[0]).To(gomega.ContainSubstring("global--foo-with-targets--default--foo.com"))

	sniCache, _ := mcache.VsCache.AviCacheGet(sniVSKey)
	sniCacheObj, _ := sniCache.(*cache.AviVsCache)
	g.Expect(sniCacheObj.SSLKeyCertCollection).To(gomega.HaveLen(1))
	g.Expect(sniCacheObj.SSLKeyCertCollection[0].Name).To(gomega.ContainSubstring("global--default--my-secret"))
	g.Expect(sniCacheObj.HTTPKeyCollection).To(gomega.HaveLen(1))
	g.Expect(sniCacheObj.HTTPKeyCollection[0].Name).To(gomega.ContainSubstring("global--default--foo.com"))
	g.Expect(sniCacheObj.ParentVSRef).To(gomega.Equal(parentVSKey))

	TearDownIngressForCacheSyncCheck(t, modelName)
}

func TestHostnameUpdateSNICacheSync(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	var err error

	modelName := "admin/Shard-VS---global-0"
	SetUpIngressForCacheSyncCheck(t, modelName, true, true)

	mcache := cache.SharedAviObjCache()
	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: "global--foo-with-targets--default--foo.com"}
	g.Eventually(func() bool {
		_, found := mcache.VsCache.AviCacheGet(sniVSKey)
		return found
	}, 15*time.Second).Should(gomega.Equal(true))
	oldSniCache, _ := mcache.VsCache.AviCacheGet(sniVSKey)
	oldSniCacheObj, _ := oldSniCache.(*cache.AviVsCache)

	ingressUpdate := (integrationtest.FakeIngress{
		Name:        "foo-with-targets",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		Paths:       []string{"/bar-updated"},
		ServiceName: "avisvc",
		TlsSecretDNS: map[string][]string{
			"my-secret": []string{"foo.com"},
		},
	}).Ingress()
	ingressUpdate.ResourceVersion = "2"
	_, err = KubeClient.ExtensionsV1beta1().Ingresses("default").Update(ingressUpdate)
	if err != nil {
		t.Fatalf("error in updating Ingress: %v", err)
	}

	// verify that a NEW httppolicy set object is created
	oldHttpPolKey := cache.NamespaceName{Namespace: "admin", Name: "global--default--foo.com/foo--foo-with-targets"}
	newHttpPolKey := cache.NamespaceName{Namespace: "admin", Name: "global--default--foo.com/bar-updated--foo-with-targets"}
	g.Eventually(func() bool {
		_, found := mcache.HTTPPolicyCache.AviCacheGet(newHttpPolKey)
		return found
	}, 10*time.Second).Should(gomega.Equal(true))

	g.Eventually(func() bool {
		_, found := mcache.HTTPPolicyCache.AviCacheGet(oldHttpPolKey)
		return found
	}, 10*time.Second).Should(gomega.Equal(false))

	// verify same vs cksum
	g.Eventually(func() string {
		sniVSCache, found := mcache.VsCache.AviCacheGet(sniVSKey)
		sniVSCacheObj, ok := sniVSCache.(*cache.AviVsCache)
		if found && ok {
			return sniVSCacheObj.CloudConfigCksum
		}
		return "456def"
	}, 15*time.Second).Should(gomega.Equal(oldSniCacheObj.CloudConfigCksum))
	sniVSCache, _ := mcache.VsCache.AviCacheGet(sniVSKey)
	sniVSCacheObj, _ := sniVSCache.(*cache.AviVsCache)
	g.Expect(sniVSCacheObj.HTTPKeyCollection).To(gomega.HaveLen(1))
	g.Expect(sniVSCacheObj.SSLKeyCertCollection).To(gomega.HaveLen(1))

	TearDownIngressForCacheSyncCheck(t, modelName)
}

func TestHostnameMultiHostMultiSecretSNICacheSync(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName := "admin/Shard-VS---global-0"
	SetUpIngressForCacheSyncCheck(t, modelName, true, true)
	mcache := cache.SharedAviObjCache()

	// update ingress
	ingressObject := integrationtest.FakeIngress{
		Name:        "foo-with-targets",
		Namespace:   "default",
		DnsNames:    []string{"foo.com", "bar.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		Paths:       []string{"/foo", "/bar"},
		ServiceName: "avisvc",
		TlsSecretDNS: map[string][]string{
			"my-secret":    []string{"foo.com"},
			"my-secret-v2": []string{"bar.com"},
		},
	}
	integrationtest.AddSecret("my-secret-v2", "default")
	ingrFake := ingressObject.Ingress()
	ingrFake.ResourceVersion = "2"
	if _, err := KubeClient.ExtensionsV1beta1().Ingresses("default").Update(ingrFake); err != nil {
		t.Fatalf("error in updating Ingress: %v", err)
	}

	sniVSKey1 := cache.NamespaceName{Namespace: "admin", Name: "global--foo-with-targets--default--foo.com"}
	sniVSKey2 := cache.NamespaceName{Namespace: "admin", Name: "global--foo-with-targets--default--bar.com"}
	g.Eventually(func() bool {
		_, found1 := mcache.VsCache.AviCacheGet(sniVSKey1)
		_, found2 := mcache.VsCache.AviCacheGet(sniVSKey2)
		if found1 && found2 {
			return true
		}
		return false
	}, 15*time.Second).Should(gomega.Equal(true))

	sniCache1, _ := mcache.VsCache.AviCacheGet(sniVSKey1)
	sniCacheObj1, _ := sniCache1.(*cache.AviVsCache)
	sniCache2, _ := mcache.VsCache.AviCacheGet(sniVSKey2)
	sniCacheObj2, _ := sniCache2.(*cache.AviVsCache)
	g.Expect(sniCacheObj1.SSLKeyCertCollection).To(gomega.HaveLen(1))
	g.Expect(sniCacheObj1.SSLKeyCertCollection[0].Name).To(gomega.Equal("global--default--my-secret"))
	g.Expect(sniCacheObj2.SSLKeyCertCollection).To(gomega.HaveLen(1))
	g.Expect(sniCacheObj2.SSLKeyCertCollection[0].Name).To(gomega.Equal("global--default--my-secret-v2"))

	g.Eventually(func() string {
		sniCache1, _ := mcache.VsCache.AviCacheGet(sniVSKey1)
		sniCacheObj1, _ := sniCache1.(*cache.AviVsCache)
		return sniCacheObj1.ParentVSRef.Name
	}, 15*time.Second).Should(gomega.Not(gomega.Equal("")))

	time.Sleep(5 * time.Second)
	fmt.Printf("X: %+v", utils.Stringify(sniCacheObj1))
	fmt.Printf("Y: %+v", utils.Stringify(sniCacheObj2))

	TearDownIngressForCacheSyncCheck(t, modelName)
}

func TestHostnameMultiHostMultiSecretUpdateSNICacheSync(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	modelName := "admin/Shard-VS---global-0"

	SetUpTestForIngress(t, modelName)
	integrationtest.PollForCompletion(t, modelName, 5)
	ingressObject := integrationtest.FakeIngress{
		Name:        "foo-with-targets",
		Namespace:   "default",
		DnsNames:    []string{"foo.com", "bar.com", "xyz.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		Paths:       []string{"/foo", "/bar", "/xyz"},
		ServiceName: "avisvc",
		TlsSecretDNS: map[string][]string{
			"my-secret":    []string{"foo.com"},
			"my-secret-v2": []string{"bar.com"},
		},
	}
	integrationtest.AddSecret("my-secret-v2", "default")
	integrationtest.AddSecret("my-secret", "default")

	ingrFake := ingressObject.Ingress()
	if _, err := KubeClient.ExtensionsV1beta1().Ingresses("default").Create(ingrFake); err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	integrationtest.PollForCompletion(t, modelName, 5)

	mcache := cache.SharedAviObjCache()
	sniVSKey1 := cache.NamespaceName{Namespace: "admin", Name: "global--foo-with-targets--default--foo.com"}
	sniVSKey2 := cache.NamespaceName{Namespace: "admin", Name: "global--foo-with-targets--default--bar.com"}

	// Shard scheme: Shard-VS---global-0 -> foo.com
	// Shard scheme: Shard-VS---global-3 -> xyz.com
	xyzParentKey := cache.NamespaceName{Namespace: "admin", Name: "Shard-VS---global-3"}
	g.Eventually(func() int {
		sniCache, _ := mcache.VsCache.AviCacheGet(xyzParentKey)
		sniCacheObj, _ := sniCache.(*cache.AviVsCache)
		return len(sniCacheObj.PoolKeyCollection)
	}, 10*time.Second).Should(gomega.Equal(1))
	sniCache, _ := mcache.VsCache.AviCacheGet(xyzParentKey)
	sniCacheObj, _ := sniCache.(*cache.AviVsCache)
	g.Expect(sniCacheObj.PoolKeyCollection[0].Name).To(gomega.ContainSubstring("xyz.com"))

	g.Eventually(func() bool {
		_, found1 := mcache.VsCache.AviCacheGet(sniVSKey1)
		_, found2 := mcache.VsCache.AviCacheGet(sniVSKey2)
		if found1 && found2 {
			return true
		}
		return false
	}, 5*time.Second).Should(gomega.Equal(true))

	g.Eventually(func() int {
		sniCache, found := mcache.VsCache.AviCacheGet(sniVSKey1)
		sniCacheObj, ok := sniCache.(*cache.AviVsCache)
		if found && ok {
			return len(sniCacheObj.PoolKeyCollection)
		}
		return 0
	}, 10*time.Second).Should(gomega.Equal(1))
	sniCache, _ = mcache.VsCache.AviCacheGet(sniVSKey1)
	sniCacheObj, _ = sniCache.(*cache.AviVsCache)
	g.Expect(sniCacheObj.PoolKeyCollection[0].Name).To(gomega.ContainSubstring("foo.com"))
	g.Expect(sniCacheObj.SSLKeyCertCollection).To(gomega.HaveLen(1))
	g.Expect(sniCacheObj.SSLKeyCertCollection[0].Name).To(gomega.Equal("global--default--my-secret"))

	g.Eventually(func() int {
		sniCache, found := mcache.VsCache.AviCacheGet(sniVSKey2)
		sniCacheObj, ok := sniCache.(*cache.AviVsCache)
		if found && ok {
			return len(sniCacheObj.PoolKeyCollection)
		}
		return 0
	}, 10*time.Second).Should(gomega.Equal(1))
	sniCache, _ = mcache.VsCache.AviCacheGet(sniVSKey2)
	sniCacheObj, _ = sniCache.(*cache.AviVsCache)
	g.Expect(sniCacheObj.PoolKeyCollection[0].Name).To(gomega.ContainSubstring("bar.com"))
	g.Expect(sniCacheObj.SSLKeyCertCollection).To(gomega.HaveLen(1))
	g.Expect(sniCacheObj.SSLKeyCertCollection[0].Name).To(gomega.Equal("global--default--my-secret-v2"))

	// delete one secret
	KubeClient.CoreV1().Secrets("default").Delete("my-secret-v2", nil)
	ingressUpdateObject := integrationtest.FakeIngress{
		Name:        "foo-with-targets",
		Namespace:   "default",
		DnsNames:    []string{"foo.com", "bar.com", "xyz.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		Paths:       []string{"/foo", "/bar", "/xyz"},
		ServiceName: "avisvc",
		TlsSecretDNS: map[string][]string{
			"my-secret": []string{"foo.com"},
		},
	}

	ingrUpdate := ingressUpdateObject.Ingress()
	ingrUpdate.ResourceVersion = "2"
	if _, err := KubeClient.ExtensionsV1beta1().Ingresses("default").Update(ingrUpdate); err != nil {
		t.Fatalf("error in updating Ingress: %v", err)
	}

	// Shard scheme: Shard-VS---global-1 -> bar.com
	barParentKey := cache.NamespaceName{Namespace: "admin", Name: "Shard-VS---global-1"}
	g.Eventually(func() int {
		sniCache, _ := mcache.VsCache.AviCacheGet(barParentKey)
		sniCacheObj, _ := sniCache.(*cache.AviVsCache)
		return len(sniCacheObj.PoolKeyCollection)
	}, 10*time.Second).Should(gomega.Equal(1))

	// should not be found
	g.Eventually(func() bool {
		_, found := mcache.VsCache.AviCacheGet(sniVSKey2)
		return found
	}, 10*time.Second).Should(gomega.Equal(false))

	sniCache, _ = mcache.VsCache.AviCacheGet(sniVSKey1)
	sniCacheObj, _ = sniCache.(*cache.AviVsCache)
	g.Expect(sniCacheObj.PoolKeyCollection).To(gomega.HaveLen(1))
	g.Expect(sniCacheObj.PoolKeyCollection[0].Name).To(gomega.ContainSubstring("foo.com"))
	g.Expect(sniCacheObj.SSLKeyCertCollection).To(gomega.HaveLen(1))
	g.Expect(sniCacheObj.SSLKeyCertCollection[0].Name).To(gomega.Equal("global--default--my-secret"))

	KubeClient.ExtensionsV1beta1().Ingresses("default").Delete("foo-with-targets", nil)
	KubeClient.CoreV1().Secrets("default").Delete("my-secret", nil)
	g.Eventually(func() bool {
		_, found := mcache.VsCache.AviCacheGet(sniVSKey1)
		return found
	}, 15*time.Second).Should(gomega.Equal(false))
	TearDownTestForIngress(t, modelName)
}

func TestHostnameDeleteSNICacheSync(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	var err error

	modelName := "admin/Shard-VS---global-0"
	SetUpIngressForCacheSyncCheck(t, modelName, true, true)

	mcache := cache.SharedAviObjCache()
	parentVSKey := cache.NamespaceName{Namespace: "admin", Name: "Shard-VS---global-0"}
	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: "global--foo-with-targets--default--foo.com"}

	ingressUpdate := (integrationtest.FakeIngress{
		Name:        "foo-with-targets",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		Paths:       []string{"/foo"},
		ServiceName: "avisvc",
	}).Ingress()
	ingressUpdate.ResourceVersion = "2"
	_, err = KubeClient.ExtensionsV1beta1().Ingresses("default").Update(ingressUpdate)
	if err != nil {
		t.Fatalf("error in updating Ingress: %v", err)
	}

	// verify that sni vs is deleted, but the parent vs is not
	// deleted snivs key should be deleted from parent vs snichildcollection
	g.Eventually(func() bool {
		_, found := mcache.VsCache.AviCacheGet(sniVSKey)
		return found
	}, 15*time.Second).Should(gomega.Equal(false))

	oldSniCache, _ := mcache.VsCache.AviCacheGet(parentVSKey)
	oldSniCacheObj, _ := oldSniCache.(*cache.AviVsCache)
	g.Expect(oldSniCacheObj.SNIChildCollection).To(gomega.HaveLen(0))

	TearDownIngressForCacheSyncCheck(t, modelName)
}

func TestHostnameCUDSecretCacheSync(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName := "admin/Shard-VS---global-0"
	SetUpIngressForCacheSyncCheck(t, modelName, true, false)

	mcache := cache.SharedAviObjCache()
	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: "global--foo-with-targets--default--foo.com"}
	sslKey := cache.NamespaceName{Namespace: "admin", Name: "global--default--my-secret"}

	// no ssl key cache would be found since the secret is not yet added
	g.Eventually(func() bool {
		_, found := mcache.SSLKeyCache.AviCacheGet(sslKey)
		return found
	}, 5*time.Second).Should(gomega.Equal(false))

	// add Secret
	integrationtest.AddSecret("my-secret", "default")

	// ssl key should be created now and must be attached to the sni vs cache
	g.Eventually(func() bool {
		_, found := mcache.SSLKeyCache.AviCacheGet(sslKey)
		return found
	}, 5*time.Second).Should(gomega.Equal(true))
	sniVSCache, _ := mcache.VsCache.AviCacheGet(sniVSKey)
	sniVSCacheObj, _ := sniVSCache.(*cache.AviVsCache)
	g.Expect(sniVSCacheObj.SSLKeyCertCollection).To(gomega.HaveLen(1))

	// update Secret
	secretUpdate := (integrationtest.FakeSecret{
		Namespace: "default",
		Name:      "my-secret",
		Cert:      "tlsCert_Updated",
		Key:       "tlsKey_Updated",
	}).Secret()
	secretUpdate.ResourceVersion = "2"
	KubeClient.CoreV1().Secrets("default").Update(secretUpdate)

	// can't check update rn, ssl cache object doesnot have checksum,
	// but PUTs happen, everytime though

	// delete Secret
	KubeClient.CoreV1().Secrets("default").Delete("my-secret", nil)

	// ssl key must be deleted again and sni vs as well
	g.Eventually(func() bool {
		_, found := mcache.SSLKeyCache.AviCacheGet(sslKey)
		return found
	}, 5*time.Second).Should(gomega.Equal(false))
	_, found := mcache.VsCache.AviCacheGet(sniVSKey)
	g.Expect(found).To(gomega.Equal(false))

	TearDownIngressForCacheSyncCheck(t, modelName)
}
