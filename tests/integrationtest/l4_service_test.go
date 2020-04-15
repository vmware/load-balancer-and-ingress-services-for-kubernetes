/*
* [2013] - [2019] Avi Networks Incorporated
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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"ako/pkg/cache"
	"ako/pkg/k8s"
	avinodes "ako/pkg/nodes"
	"ako/pkg/objects"

	"github.com/avinetworks/container-lib/utils"
	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

func SetUpTestForSvcLB(t *testing.T) {
	objects.SharedAviGraphLister().Delete(SINGLEPORTMODEL)
	CreateSVC(t, NAMESPACE, SINGLEPORTSVC, corev1.ServiceTypeLoadBalancer, false)
	CreateEP(t, NAMESPACE, SINGLEPORTSVC, false, false)
	PollForCompletion(t, SINGLEPORTMODEL, 5)
}

func TearDownTestForSvcLB(t *testing.T, g *gomega.GomegaWithT) {
	objects.SharedAviGraphLister().Delete(SINGLEPORTMODEL)
	DelSVC(t, NAMESPACE, SINGLEPORTSVC)
	DelEP(t, NAMESPACE, SINGLEPORTSVC)
	mcache := cache.SharedAviObjCache()
	vsKey := cache.NamespaceName{Namespace: AVINAMESPACE, Name: fmt.Sprintf("global--%s--%s", SINGLEPORTSVC, NAMESPACE)}
	g.Eventually(func() bool {
		_, found := mcache.VsCache.AviCacheGet(vsKey)
		return found
	}, 5*time.Second).Should(gomega.Equal(false))
}

func SetUpTestForSvcLBMultiport(t *testing.T) {
	objects.SharedAviGraphLister().Delete(MULTIPORTMODEL)
	CreateSVC(t, NAMESPACE, MULTIPORTSVC, corev1.ServiceTypeLoadBalancer, true)
	CreateEP(t, NAMESPACE, MULTIPORTSVC, true, true)
	PollForCompletion(t, MULTIPORTMODEL, 10)
}

func TearDownTestForSvcLBMultiport(t *testing.T, g *gomega.GomegaWithT) {
	objects.SharedAviGraphLister().Delete(MULTIPORTMODEL)
	DelSVC(t, NAMESPACE, MULTIPORTSVC)
	DelEP(t, NAMESPACE, MULTIPORTSVC)
	mcache := cache.SharedAviObjCache()
	vsKey := cache.NamespaceName{Namespace: AVINAMESPACE, Name: fmt.Sprintf("global--%s--%s", MULTIPORTSVC, NAMESPACE)}
	g.Eventually(func() bool {
		_, found := mcache.VsCache.AviCacheGet(vsKey)
		return found
	}, 5*time.Second).Should(gomega.Equal(false))
}

func TestMain(m *testing.M) {
	KubeClient = k8sfake.NewSimpleClientset()
	registeredInformers := []string{
		utils.ServiceInformer,
		utils.EndpointInformer,
		utils.ExtV1IngressInformer,
		utils.SecretInformer,
		utils.NSInformer,
		utils.NodeInformer,
		utils.ConfigMapInformer,
	}
	utils.NewInformers(utils.KubeClientIntf{KubeClient}, registeredInformers)
	informers := k8s.K8sinformers{Cs: KubeClient}

	NewAviFakeClientInstance()
	defer AviFakeClientInstance.Close()

	ctrl = k8s.SharedAviController()
	stopCh := utils.SetupSignalHandler()
	ctrlCh := make(chan struct{})
	ctrl.HandleConfigMap(informers, ctrlCh, stopCh)
	go ctrl.InitController(informers, ctrlCh, stopCh)
	AddConfigMap()
	os.Exit(m.Run())
}

func TestAviNodeCreationSinglePort(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetUpTestForSvcLB(t)

	found, aviModel := objects.SharedAviGraphLister().Get(SINGLEPORTMODEL)
	if !found {
		t.Fatalf("Couldn't find model %v", SINGLEPORTMODEL)
	} else {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(nodes).To(gomega.HaveLen(1))
		g.Expect(nodes[0].Name).To(gomega.Equal(fmt.Sprintf("global--%s--%s", NAMESPACE, SINGLEPORTSVC)))
		g.Expect(nodes[0].Tenant).To(gomega.Equal(AVINAMESPACE))
		g.Expect(nodes[0].EastWest).To(gomega.Equal(false))
		g.Expect(nodes[0].PortProto[0].Port).To(gomega.Equal(int32(8080)))

		// Check for the pools
		g.Expect(nodes[0].PoolRefs).To(gomega.HaveLen(1))
		address := "1.1.1.1"
		g.Expect(nodes[0].PoolRefs[0].Servers[0].Ip.Addr).To(gomega.Equal(&address))
		g.Expect(nodes[0].TCPPoolGroupRefs).To(gomega.HaveLen(1))
		g.Expect(nodes[0].PoolGroupRefs).To(gomega.HaveLen(1))
	}

	TearDownTestForSvcLB(t, g)
}

func TestAviNodeCreationMultiPort(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	modelName := fmt.Sprintf("%s/global--%s--%s", AVINAMESPACE, NAMESPACE, MULTIPORTSVC)

	SetUpTestForSvcLBMultiport(t)

	found, aviModel := objects.SharedAviGraphLister().Get(modelName)
	if !found {
		t.Fatalf("Couldn't find model %v", modelName)
	} else {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(nodes).To(gomega.HaveLen(1))
		g.Expect(nodes[0].Name).To(gomega.Equal(fmt.Sprintf("global--%s--%s", NAMESPACE, MULTIPORTSVC)))
		g.Expect(nodes[0].Tenant).To(gomega.Equal(AVINAMESPACE))
		g.Expect(nodes[0].EastWest).To(gomega.Equal(false))
		g.Expect(nodes[0].PortProto[0].Port).To(gomega.Equal(int32(8080)))

		// Check for the pools
		g.Expect(nodes[0].PoolRefs).To(gomega.HaveLen(3))
		for _, node := range nodes[0].PoolRefs {
			if node.Port == 8080 {
				address := "1.1.1.1"
				g.Expect(node.Servers).To(gomega.HaveLen(3))
				g.Expect(node.Servers[0].Ip.Addr).To(gomega.Equal(&address))
			} else if node.Port == 8081 {
				address := "1.1.1.4"
				g.Expect(node.Servers).To(gomega.HaveLen(2))
				g.Expect(node.Servers[0].Ip.Addr).To(gomega.Equal(&address))
			} else {
				address := "1.1.1.6"
				g.Expect(node.Servers).To(gomega.HaveLen(1))
				g.Expect(node.Servers[0].Ip.Addr).To(gomega.Equal(&address))
			}
		}
		g.Expect(nodes[0].TCPPoolGroupRefs).To(gomega.HaveLen(3))
		g.Expect(nodes[0].PoolGroupRefs).To(gomega.HaveLen(3))
		g.Expect(nodes[0].ApplicationProfile).To(gomega.Equal(utils.DEFAULT_L4_APP_PROFILE))
		g.Expect(nodes[0].NetworkProfile).To(gomega.Equal(utils.DEFAULT_TCP_NW_PROFILE))
	}

	TearDownTestForSvcLBMultiport(t, g)
}

func TestAviNodeMultiPortApplicationProf(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	modelName := fmt.Sprintf("%s/global--%s--%s", AVINAMESPACE, NAMESPACE, MULTIPORTSVC)

	SetUpTestForSvcLBMultiport(t)

	found, aviModel := objects.SharedAviGraphLister().Get(modelName)
	if !found {
		t.Fatalf("Couldn't find model %v", modelName)
	} else {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(nodes).To(gomega.HaveLen(1))
		g.Expect(nodes[0].Name).To(gomega.Equal(fmt.Sprintf("global--%s--%s", NAMESPACE, MULTIPORTSVC)))
		g.Expect(nodes[0].Tenant).To(gomega.Equal(AVINAMESPACE))
		g.Expect(nodes[0].EastWest).To(gomega.Equal(false))
		g.Expect(nodes[0].PortProto[0].Port).To(gomega.Equal(int32(8080)))

		// Check for the pools
		g.Expect(nodes[0].PoolRefs).To(gomega.HaveLen(3))
		for _, node := range nodes[0].PoolRefs {
			if node.Port == 8080 {
				address := "1.1.1.1"
				g.Expect(node.Servers).To(gomega.HaveLen(3))
				g.Expect(node.Servers[0].Ip.Addr).To(gomega.Equal(&address))
			} else if node.Port == 8081 {
				address := "1.1.1.4"
				g.Expect(node.Servers).To(gomega.HaveLen(2))
				g.Expect(node.Servers[0].Ip.Addr).To(gomega.Equal(&address))
			} else if node.Port == 8082 {
				address := "1.1.1.6"
				g.Expect(node.Servers).To(gomega.HaveLen(1))
				g.Expect(node.Servers[0].Ip.Addr).To(gomega.Equal(&address))
			}
		}
		g.Expect(nodes[0].TCPPoolGroupRefs).To(gomega.HaveLen(3))
		g.Expect(nodes[0].PoolGroupRefs).To(gomega.HaveLen(3))
		g.Expect(nodes[0].SharedVS).To(gomega.Equal(false))
		g.Expect(nodes[0].ApplicationProfile).To(gomega.Equal(utils.DEFAULT_L4_APP_PROFILE))
		g.Expect(nodes[0].NetworkProfile).To(gomega.Equal(utils.DEFAULT_TCP_NW_PROFILE))
	}

	TearDownTestForSvcLBMultiport(t, g)
}

func TestAviNodeUpdateEndpoint(t *testing.T) {
	var err error
	g := gomega.NewGomegaWithT(t)
	modelName := fmt.Sprintf("%s/global--%s--%s", AVINAMESPACE, NAMESPACE, SINGLEPORTSVC)

	SetUpTestForSvcLB(t)

	epExample := &corev1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{Namespace: NAMESPACE, Name: SINGLEPORTSVC},
		Subsets: []corev1.EndpointSubset{{
			Addresses: []corev1.EndpointAddress{{IP: "1.2.3.14"}, {IP: "1.2.3.24"}},
			Ports:     []corev1.EndpointPort{{Name: "foo", Port: 8080, Protocol: "TCP"}},
		}},
	}
	if _, err = KubeClient.CoreV1().Endpoints(NAMESPACE).Update(epExample); err != nil {
		t.Fatalf("Error in updating the Endpoint: %v", err)
	}

	var aviModel interface{}
	g.Eventually(func() []avinodes.AviPoolMetaServer {
		_, aviModel = objects.SharedAviGraphLister().Get(modelName)
		pools := aviModel.(*avinodes.AviObjectGraph).GetAviPoolNodes()
		return pools[0].Servers
	}, 5*time.Second).Should(gomega.HaveLen(2))

	found, aviModel := objects.SharedAviGraphLister().Get(modelName)
	if !found {
		t.Fatalf("Couldn't find model %v", modelName)
	} else {
		pools := aviModel.(*avinodes.AviObjectGraph).GetAviPoolNodes()
		for _, pool := range pools {
			if pool.Port == 8080 {
				address := "1.2.3.24"
				g.Expect(pool.Servers).To(gomega.HaveLen(2))
				g.Expect(pool.Servers[1].Ip.Addr).To(gomega.Equal(&address))
			} else {
				g.Expect(pool.Servers).To(gomega.HaveLen(0))
			}
		}
	}

	TearDownTestForSvcLB(t, g)
}

// Rest Cache sync tests

func TestCreateServiceLBCacheSync(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	SetUpTestForSvcLB(t)

	mcache := cache.SharedAviObjCache()
	vsKey := cache.NamespaceName{Namespace: AVINAMESPACE, Name: fmt.Sprintf("global--%s--%s", NAMESPACE, SINGLEPORTSVC)}
	vsCache, found := mcache.VsCache.AviCacheGet(vsKey)
	if !found {
		t.Fatalf("Cache not found for VS: %v", vsKey)
	} else {
		vsCacheObj, ok := vsCache.(*cache.AviVsCache)
		if !ok {
			t.Fatalf("Invalid VS object. Cannot cast.")
		}
		g.Expect(vsCacheObj.Name).To(gomega.Equal(fmt.Sprintf("global--%s--%s", NAMESPACE, SINGLEPORTSVC)))
		g.Expect(vsCacheObj.Tenant).To(gomega.Equal(AVINAMESPACE))
		g.Expect(vsCacheObj.PoolKeyCollection).To(gomega.HaveLen(1))
		g.Expect(vsCacheObj.PoolKeyCollection[0].Name).To(gomega.MatchRegexp("global--red-ns--testsvc--8080"))
		g.Expect(vsCacheObj.PGKeyCollection).To(gomega.HaveLen(1))
		g.Expect(vsCacheObj.PGKeyCollection[0].Name).To(gomega.MatchRegexp("global--red-ns--testsvc--8080"))
	}

	TearDownTestForSvcLB(t, g)
	g.Eventually(func() bool {
		_, found := mcache.VsCache.AviCacheGet(vsKey)
		return found
	}, 10*time.Second).Should(gomega.Equal(false))
}

func TestCreateServiceLBWithFaultCacheSync(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

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
		} else if strings.Contains(url, "login") {
			// This is used for /login --> first request to controller
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, `{"success": "true"}`)
		}
	})
	defer ResetMiddleware()

	SetUpTestForSvcLB(t)

	mcache := cache.SharedAviObjCache()
	vsKey := cache.NamespaceName{Namespace: AVINAMESPACE, Name: fmt.Sprintf("global--%s--%s", NAMESPACE, SINGLEPORTSVC)}
	g.Eventually(func() bool {
		_, found := mcache.VsCache.AviCacheGet(vsKey)
		return found
	}, 10*time.Second).Should(gomega.Equal(true))
	vsCache, found := mcache.VsCache.AviCacheGet(vsKey)
	if !found {
		t.Fatalf("Cache not found for VS: %v", vsKey)
	} else {
		vsCacheObj, ok := vsCache.(*cache.AviVsCache)
		if !ok {
			t.Fatalf("Invalid VS object. Cannot cast.")
		}
		g.Expect(vsCacheObj.Name).To(gomega.Equal(fmt.Sprintf("global--%s--%s", NAMESPACE, SINGLEPORTSVC)))
		g.Expect(vsCacheObj.Tenant).To(gomega.Equal(AVINAMESPACE))
		g.Expect(vsCacheObj.PoolKeyCollection).To(gomega.HaveLen(1))
		g.Expect(vsCacheObj.PoolKeyCollection[0].Name).To(gomega.MatchRegexp("global--red-ns--testsvc--8080"))
		g.Expect(vsCacheObj.PGKeyCollection).To(gomega.HaveLen(1))
		g.Expect(vsCacheObj.PGKeyCollection[0].Name).To(gomega.MatchRegexp("global--red-ns--testsvc--8080"))
	}

	TearDownTestForSvcLB(t, g)
}

func TestCreateMultiportServiceLBCacheSync(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	MULTIPORTSVC, NAMESPACE, AVINAMESPACE := "testsvcmulti", "red-ns", "admin"

	SetUpTestForSvcLBMultiport(t)

	mcache := cache.SharedAviObjCache()
	vsKey := cache.NamespaceName{Namespace: AVINAMESPACE, Name: fmt.Sprintf("global--%s--%s", NAMESPACE, MULTIPORTSVC)}
	vsCache, found := mcache.VsCache.AviCacheGet(vsKey)
	if !found {
		t.Fatalf("Cache not found for VS: %v", vsKey)
	}
	vsCacheObj, ok := vsCache.(*cache.AviVsCache)
	if !ok {
		t.Fatalf("Invalid VS object. Cannot cast.")
	}
	g.Expect(vsCacheObj.Name).To(gomega.Equal(fmt.Sprintf("global--%s--%s", NAMESPACE, MULTIPORTSVC)))
	g.Expect(vsCacheObj.Tenant).To(gomega.Equal(AVINAMESPACE))
	g.Expect(vsCacheObj.PoolKeyCollection).To(gomega.HaveLen(3))
	g.Expect(vsCacheObj.PoolKeyCollection[0].Name).To(gomega.MatchRegexp(`^(global--[a-zA-Z0-9-]+-808(0|1|2))$`))
	g.Expect(vsCacheObj.PGKeyCollection).To(gomega.HaveLen(3))
	g.Expect(vsCacheObj.PGKeyCollection[0].Name).To(gomega.MatchRegexp(`^(global--[a-zA-Z0-9-]+-808(0|1|2))$`))

	TearDownTestForSvcLBMultiport(t, g)
}

func TestUpdateAndDeleteServiceLBCacheSync(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	var err error

	SetUpTestForSvcLB(t)

	// Get hold of the pool checksum on CREATE
	poolName := "global--red-ns--testsvc--8080"
	mcache := cache.SharedAviObjCache()
	poolKey := cache.NamespaceName{Namespace: AVINAMESPACE, Name: poolName}
	poolCacheBefore, _ := mcache.PoolCache.AviCacheGet(poolKey)
	poolCacheBeforeObj, _ := poolCacheBefore.(*cache.AviPoolCache)
	oldPoolCksum := poolCacheBeforeObj.CloudConfigCksum

	// UPDATE Test: After Endpoint update, Cache checksums must change
	epExample := &corev1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{Namespace: NAMESPACE, Name: SINGLEPORTSVC},
		Subsets: []corev1.EndpointSubset{{
			Addresses: []corev1.EndpointAddress{{IP: "1.2.3.14"}, {IP: "1.2.3.24"}},
			Ports:     []corev1.EndpointPort{{Name: "foo", Port: 8080, Protocol: "TCP"}},
		}},
	}
	if _, err = KubeClient.CoreV1().Endpoints(NAMESPACE).Update(epExample); err != nil {
		t.Fatalf("Error in updating the Endpoint: %v", err)
	}

	var poolCacheObj *cache.AviPoolCache
	var poolCache interface{}
	var found, ok bool
	g.Eventually(func() string {
		if poolCache, found = mcache.PoolCache.AviCacheGet(poolKey); found {
			if poolCacheObj, ok = poolCache.(*cache.AviPoolCache); ok {
				return poolCacheObj.CloudConfigCksum
			}
		}
		return oldPoolCksum
	}, 5*time.Second).Should(gomega.Not(gomega.Equal(oldPoolCksum)))
	if poolCache, found = mcache.PoolCache.AviCacheGet(poolKey); !found {
		t.Fatalf("Cache not updated for Pool: %v", poolKey)
	}
	if poolCacheObj, ok = poolCache.(*cache.AviPoolCache); !ok {
		t.Fatalf("Invalid Pool object. Cannot cast.")
	}
	g.Expect(poolCacheObj.Name).To(gomega.Equal(poolName))
	g.Expect(poolCacheObj.Tenant).To(gomega.Equal(AVINAMESPACE))

	// DELETE Test: Cache corresponding to the pool MUST NOT be found
	TearDownTestForSvcLB(t, g)
	g.Eventually(func() bool {
		_, found = mcache.PoolCache.AviCacheGet(poolKey)
		return found
	}, 5*time.Second).Should(gomega.Equal(false))
}

//TestScaleUpAndDownServiceLBCacheSync tests the avi node graph and rest layer functionality when the
//multiport serviceLB is increased from 1 to 30 and then decreased back to 1
func TestScaleUpAndDownServiceLBCacheSync(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	var model, service string

	// Simulate a delay of 200ms in the Avi API
	AddMiddleware(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		NormalControllerServer(w, r)
	})
	defer ResetMiddleware()

	SetUpTestForSvcLB(t)

	// create 30 more multiport service of type loadbalancer
	numScale := 30
	for i := 0; i < numScale; i++ {
		service = fmt.Sprintf("%s%d", MULTIPORTSVC, i)
		model = strings.Replace(MULTIPORTMODEL, MULTIPORTSVC, service, 1)

		objects.SharedAviGraphLister().Delete(model)
		CreateSVC(t, NAMESPACE, service, corev1.ServiceTypeLoadBalancer, true)
		CreateEP(t, NAMESPACE, service, true, true)
	}

	// verify that 30 services are created on the graph and corresponding cache objects
	var found bool
	var vsKey cache.NamespaceName
	var aviModel interface{}

	mcache := cache.SharedAviObjCache()
	for i := 0; i < numScale; i++ {
		service = fmt.Sprintf("%s%d", MULTIPORTSVC, i)
		model = strings.Replace(MULTIPORTMODEL, MULTIPORTSVC, service, 1)

		PollForCompletion(t, model, 5)
		found, aviModel = objects.SharedAviGraphLister().Get(model)
		g.Expect(found).To(gomega.Equal(true))
		g.Expect(aviModel).To(gomega.Not(gomega.BeNil()))

		vsKey = cache.NamespaceName{Namespace: AVINAMESPACE, Name: strings.TrimPrefix(model, AVINAMESPACE+"/")}
		g.Eventually(func() bool {
			_, found = mcache.VsCache.AviCacheGet(vsKey)
			return found
		}, 15*time.Second).Should(gomega.Equal(true))
	}

	// delete the 30 services
	for i := 0; i < numScale; i++ {
		service = fmt.Sprintf("%s%d", MULTIPORTSVC, i)
		model = strings.Replace(MULTIPORTMODEL, MULTIPORTSVC, service, 1)
		objects.SharedAviGraphLister().Delete(model)
		DelSVC(t, NAMESPACE, service)
		DelEP(t, NAMESPACE, service)
	}

	// verify that the graph nodes and corresponding cache are deleted for the 30 services
	for i := 0; i < numScale; i++ {
		service = fmt.Sprintf("%s%d", MULTIPORTSVC, i)
		model = strings.Replace(MULTIPORTMODEL, MULTIPORTSVC, service, 1)
		g.Eventually(func() interface{} {
			found, aviModel = objects.SharedAviGraphLister().Get(model)
			return aviModel
		}, 25*time.Second).Should(gomega.BeNil())

		vsKey = cache.NamespaceName{Namespace: AVINAMESPACE, Name: strings.TrimPrefix(model, AVINAMESPACE+"/")}
		g.Eventually(func() bool {
			_, found = mcache.VsCache.AviCacheGet(vsKey)
			return found
		}, 25*time.Second).Should(gomega.Equal(false))
	}

	// verifying whether the first service created still has the corresponding cache entry
	vsKey = cache.NamespaceName{Namespace: AVINAMESPACE, Name: fmt.Sprintf("global--%s--%s", NAMESPACE, SINGLEPORTSVC)}
	g.Eventually(func() bool {
		_, found = mcache.VsCache.AviCacheGet(vsKey)
		return found
	}, 10*time.Second).Should(gomega.Equal(true))
	TearDownTestForSvcLB(t, g)
}
