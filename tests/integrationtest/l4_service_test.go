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
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/k8s"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	avinodes "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	crdfake "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/client/v1alpha1/clientset/versioned/fake"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

func SetUpTestForSvcLB(t *testing.T) {
	objects.SharedAviGraphLister().Delete(SINGLEPORTMODEL)
	CreateSVC(t, NAMESPACE, SINGLEPORTSVC, corev1.ServiceTypeLoadBalancer, false)
	CreateEP(t, NAMESPACE, SINGLEPORTSVC, false, false, "1.1.1")
	PollForCompletion(t, SINGLEPORTMODEL, 5)
}

func TearDownTestForSvcLB(t *testing.T, g *gomega.GomegaWithT) {
	objects.SharedAviGraphLister().Delete(SINGLEPORTMODEL)
	DelSVC(t, NAMESPACE, SINGLEPORTSVC)
	DelEP(t, NAMESPACE, SINGLEPORTSVC)
	mcache := cache.SharedAviObjCache()
	vsKey := cache.NamespaceName{Namespace: AVINAMESPACE, Name: fmt.Sprintf("cluster--%s-%s", SINGLEPORTSVC, NAMESPACE)}
	g.Eventually(func() bool {
		_, found := mcache.VsCacheMeta.AviCacheGet(vsKey)
		return found
	}, 5*time.Second).Should(gomega.Equal(false))
}

func SetUpTestForSvcLBMultiport(t *testing.T) {
	objects.SharedAviGraphLister().Delete(MULTIPORTMODEL)
	CreateSVC(t, NAMESPACE, MULTIPORTSVC, corev1.ServiceTypeLoadBalancer, true)
	CreateEP(t, NAMESPACE, MULTIPORTSVC, true, true, "1.1.1")
	PollForCompletion(t, MULTIPORTMODEL, 10)
}

func TearDownTestForSvcLBMultiport(t *testing.T, g *gomega.GomegaWithT) {
	objects.SharedAviGraphLister().Delete(MULTIPORTMODEL)
	DelSVC(t, NAMESPACE, MULTIPORTSVC)
	DelEP(t, NAMESPACE, MULTIPORTSVC)
	mcache := cache.SharedAviObjCache()
	vsKey := cache.NamespaceName{Namespace: AVINAMESPACE, Name: fmt.Sprintf("cluster--%s-%s", MULTIPORTSVC, NAMESPACE)}
	g.Eventually(func() bool {
		_, found := mcache.VsCacheMeta.AviCacheGet(vsKey)
		return found
	}, 5*time.Second).Should(gomega.Equal(false))
}

func TestMain(m *testing.M) {
	os.Setenv("VIP_NETWORK_LIST", `[{"networkName":"net123"}]`)
	os.Setenv("CLUSTER_NAME", "cluster")
	os.Setenv("CLOUD_NAME", "CLOUD_VCENTER")
	os.Setenv("SEG_NAME", "Default-Group")
	os.Setenv("NODE_NETWORK_LIST", `[{"networkName":"net123","cidrs":["10.79.168.0/22"]}]`)
	os.Setenv("SERVICE_TYPE", "ClusterIP")
	os.Setenv("AUTO_L4_FQDN", "disable")
	os.Setenv("POD_NAMESPACE", utils.AKO_DEFAULT_NS)
	os.Setenv("SHARD_VS_SIZE", "LARGE")

	akoControlConfig := lib.AKOControlConfig()
	KubeClient = k8sfake.NewSimpleClientset()
	CRDClient = crdfake.NewSimpleClientset()
	akoControlConfig.SetCRDClientset(CRDClient)
	akoControlConfig.SetAKOInstanceFlag(true)
	akoControlConfig.SetEventRecorder(lib.AKOEventComponent, KubeClient, true)
	data := map[string][]byte{
		"username": []byte("admin"),
		"password": []byte("admin"),
	}
	object := metav1.ObjectMeta{Name: "avi-secret", Namespace: utils.GetAKONamespace()}
	secret := &corev1.Secret{Data: data, ObjectMeta: object}
	KubeClient.CoreV1().Secrets(utils.GetAKONamespace()).Create(context.TODO(), secret, metav1.CreateOptions{})

	registeredInformers := []string{
		utils.ServiceInformer,
		utils.EndpointInformer,
		utils.IngressInformer,
		utils.IngressClassInformer,
		utils.SecretInformer,
		utils.NSInformer,
		utils.NodeInformer,
		utils.ConfigMapInformer,
	}
	utils.NewInformers(utils.KubeClientIntf{ClientSet: KubeClient}, registeredInformers)
	informers := k8s.K8sinformers{Cs: KubeClient}
	k8s.NewCRDInformers(CRDClient)

	InitializeFakeAKOAPIServer()

	NewAviFakeClientInstance(KubeClient)
	defer AviFakeClientInstance.Close()

	ctrl = k8s.SharedAviController()
	stopCh := utils.SetupSignalHandler()
	ctrlCh := make(chan struct{})
	quickSyncCh := make(chan struct{})
	waitGroupMap := make(map[string]*sync.WaitGroup)
	wgIngestion := &sync.WaitGroup{}
	waitGroupMap["ingestion"] = wgIngestion
	wgFastRetry := &sync.WaitGroup{}
	waitGroupMap["fastretry"] = wgFastRetry
	wgSlowRetry := &sync.WaitGroup{}
	waitGroupMap["slowretry"] = wgSlowRetry
	wgGraph := &sync.WaitGroup{}
	waitGroupMap["graph"] = wgGraph
	wgStatus := &sync.WaitGroup{}
	waitGroupMap["status"] = wgStatus
	wgLeaderElection := &sync.WaitGroup{}
	waitGroupMap["leaderElection"] = wgLeaderElection

	AddConfigMap(KubeClient)
	ctrl.SetSEGroupCloudNameFromNSAnnotations()
	PollForSyncStart(ctrl, 10)

	ctrl.HandleConfigMap(informers, ctrlCh, stopCh, quickSyncCh)
	AddDefaultIngressClass()

	go ctrl.InitController(informers, registeredInformers, ctrlCh, stopCh, quickSyncCh, waitGroupMap)
	os.Exit(m.Run())
}

func TestAviSvcCreationSinglePort(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetUpTestForSvcLB(t)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(SINGLEPORTMODEL)
		return found
	}, 10*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(SINGLEPORTMODEL)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].Name).To(gomega.Equal(fmt.Sprintf("cluster--%s-%s", NAMESPACE, SINGLEPORTSVC)))
	g.Expect(nodes[0].Tenant).To(gomega.Equal(AVINAMESPACE))
	g.Expect(nodes[0].PortProto[0].Port).To(gomega.Equal(int32(8080)))

	// Check for the pools
	g.Expect(nodes[0].PoolRefs).To(gomega.HaveLen(1))
	address := "1.1.1.1"
	g.Expect(nodes[0].PoolRefs[0].Servers[0].Ip.Addr).To(gomega.Equal(&address))

	// If we transition the service from Loadbalancer to ClusterIP - it should get deleted.
	svcExample := (FakeService{
		Name:         SINGLEPORTSVC,
		Namespace:    NAMESPACE,
		Type:         corev1.ServiceTypeClusterIP,
		ServicePorts: []Serviceport{{PortName: "foo1", Protocol: "TCP", PortNumber: 8080, TargetPort: intstr.FromInt(8080)}},
	}).Service()
	svcExample.ResourceVersion = "2"
	_, err := KubeClient.CoreV1().Services(NAMESPACE).Update(context.TODO(), svcExample, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in adding Service: %v", err)
	}
	mcache := cache.SharedAviObjCache()
	vsKey := cache.NamespaceName{Namespace: AVINAMESPACE, Name: fmt.Sprintf("cluster--%s-%s", NAMESPACE, SINGLEPORTSVC)}
	g.Eventually(func() bool {
		_, found := mcache.VsCacheMeta.AviCacheGet(vsKey)
		return found
	}, 15*time.Second).Should(gomega.Equal(false))
	// If we transition the service from clusterIP to Loadbalancer - vs should get ceated
	svcExample = (FakeService{
		Name:         SINGLEPORTSVC,
		Namespace:    NAMESPACE,
		Type:         corev1.ServiceTypeLoadBalancer,
		ServicePorts: []Serviceport{{PortName: "foo1", Protocol: "TCP", PortNumber: 8080, TargetPort: intstr.FromInt(8080)}},
	}).Service()
	svcExample.ResourceVersion = "3"
	_, err = KubeClient.CoreV1().Services(NAMESPACE).Update(context.TODO(), svcExample, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in adding Service: %v", err)
	}

	g.Eventually(func() bool {
		_, found := mcache.VsCacheMeta.AviCacheGet(vsKey)
		return found
	}, 15*time.Second).Should(gomega.Equal(true))
	TearDownTestForSvcLB(t, g)
}

func TestAviSvcCreationSinglePortMultiTenantEnabled(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetAkoTenant()
	defer ResetAkoTenant()
	modelName := fmt.Sprintf("%s/cluster--red-ns-testsvc", AKOTENANT)
	objects.SharedAviGraphLister().Delete(modelName)
	CreateSVC(t, NAMESPACE, SINGLEPORTSVC, corev1.ServiceTypeLoadBalancer, false)
	CreateEP(t, NAMESPACE, SINGLEPORTSVC, false, false, "1.1.1")

	var aviModel interface{}
	var found bool
	g.Eventually(func() bool {
		found, aviModel = objects.SharedAviGraphLister().Get(modelName)
		if found && aviModel != nil {
			return true
		}
		return false
	}, 40*time.Second).Should(gomega.Equal(true))

	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].Name).To(gomega.Equal(fmt.Sprintf("cluster--%s-%s", NAMESPACE, SINGLEPORTSVC)))
	// Tenant should be akotenant instead of admin
	g.Expect(nodes[0].Tenant).To(gomega.Equal(AKOTENANT))
	g.Expect(nodes[0].PortProto[0].Port).To(gomega.Equal(int32(8080)))

	// Check for the pools
	g.Expect(nodes[0].PoolRefs).To(gomega.HaveLen(1))
	address := "1.1.1.1"
	g.Expect(nodes[0].PoolRefs[0].Servers[0].Ip.Addr).To(gomega.Equal(&address))

	objects.SharedAviGraphLister().Delete(modelName)
	DelSVC(t, NAMESPACE, SINGLEPORTSVC)
	DelEP(t, NAMESPACE, SINGLEPORTSVC)
	mcache := cache.SharedAviObjCache()
	vsKey := cache.NamespaceName{Namespace: AVINAMESPACE, Name: fmt.Sprintf("cluster--%s-%s", SINGLEPORTSVC, NAMESPACE)}
	g.Eventually(func() bool {
		_, found := mcache.VsCacheMeta.AviCacheGet(vsKey)
		return found
	}, 5*time.Second).Should(gomega.Equal(false))
}

func TestAviSvcCreationMultiPort(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	modelName := fmt.Sprintf("%s/cluster--%s-%s", AVINAMESPACE, NAMESPACE, MULTIPORTSVC)

	SetUpTestForSvcLBMultiport(t)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 10*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].Name).To(gomega.Equal(fmt.Sprintf("cluster--%s-%s", NAMESPACE, MULTIPORTSVC)))
	g.Expect(nodes[0].Tenant).To(gomega.Equal(AVINAMESPACE))
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
	g.Expect(nodes[0].L4PolicyRefs).To(gomega.HaveLen(1))
	g.Expect(nodes[0].ApplicationProfile).To(gomega.Equal(utils.DEFAULT_L4_APP_PROFILE))
	g.Expect(nodes[0].NetworkProfile).To(gomega.Equal(utils.TCP_NW_FAST_PATH))

	TearDownTestForSvcLBMultiport(t, g)
}

func TestL4NamingConvention(t *testing.T) {
	// checks naming convention of all generated nodes
	g := gomega.NewGomegaWithT(t)
	modelName := fmt.Sprintf("%s/cluster--%s-%s", AVINAMESPACE, NAMESPACE, MULTIPORTSVC)

	SetUpTestForSvcLBMultiport(t)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 15*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes[0].Name).To(gomega.Equal("cluster--red-ns-testsvcmulti"))
	g.Expect(nodes[0].PoolRefs[0].Name).To(gomega.ContainSubstring("cluster--red-ns-testsvcmulti-TCP-808"))
	g.Expect(nodes[0].VSVIPRefs[0].Name).To(gomega.Equal("cluster--red-ns-testsvcmulti"))
	g.Expect(nodes[0].L4PolicyRefs[0].Name).To(gomega.Equal("cluster--red-ns-testsvcmulti"))

	TearDownTestForSvcLBMultiport(t, g)
}

func TestAviSvcMultiPortApplicationProf(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	modelName := fmt.Sprintf("%s/cluster--%s-%s", AVINAMESPACE, NAMESPACE, MULTIPORTSVC)

	SetUpTestForSvcLBMultiport(t)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 10*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].Name).To(gomega.Equal(fmt.Sprintf("cluster--%s-%s", NAMESPACE, MULTIPORTSVC)))
	g.Expect(nodes[0].Tenant).To(gomega.Equal(AVINAMESPACE))
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
	g.Expect(nodes[0].L4PolicyRefs).To(gomega.HaveLen(1))
	g.Expect(nodes[0].SharedVS).To(gomega.Equal(false))
	g.Expect(nodes[0].ApplicationProfile).To(gomega.Equal(utils.DEFAULT_L4_APP_PROFILE))
	g.Expect(nodes[0].NetworkProfile).To(gomega.Equal(utils.TCP_NW_FAST_PATH))

	TearDownTestForSvcLBMultiport(t, g)
}

func TestAviSvcUpdateEndpoint(t *testing.T) {
	var err error
	g := gomega.NewGomegaWithT(t)
	modelName := fmt.Sprintf("%s/cluster--%s-%s", AVINAMESPACE, NAMESPACE, SINGLEPORTSVC)

	SetUpTestForSvcLB(t)

	epExample := &corev1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{Namespace: NAMESPACE, Name: SINGLEPORTSVC},
		Subsets: []corev1.EndpointSubset{{
			Addresses: []corev1.EndpointAddress{{IP: "1.2.3.14"}, {IP: "1.2.3.24"}},
			Ports:     []corev1.EndpointPort{{Name: "foo", Port: 8080, Protocol: "TCP"}},
		}},
	}
	if _, err = KubeClient.CoreV1().Endpoints(NAMESPACE).Update(context.TODO(), epExample, metav1.UpdateOptions{}); err != nil {
		t.Fatalf("Error in updating the Endpoint: %v", err)
	}

	var aviModel interface{}
	g.Eventually(func() []avinodes.AviPoolMetaServer {
		_, aviModel = objects.SharedAviGraphLister().Get(modelName)
		node := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0]
		return node.PoolRefs[0].Servers
	}, 5*time.Second).Should(gomega.HaveLen(2))

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 10*time.Second).Should(gomega.Equal(true))
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	for _, pool := range nodes[0].PoolRefs {
		if pool.Port == 8080 {
			address := "1.2.3.24"
			g.Expect(pool.Servers).To(gomega.HaveLen(2))
			g.Expect(pool.Servers[1].Ip.Addr).To(gomega.Equal(&address))
		} else {
			g.Expect(pool.Servers).To(gomega.HaveLen(0))
		}
	}

	TearDownTestForSvcLB(t, g)
}

// Rest Cache sync tests

func TestCreateServiceLBCacheSync(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	SetUpTestForSvcLB(t)

	mcache := cache.SharedAviObjCache()
	vsKey := cache.NamespaceName{Namespace: AVINAMESPACE, Name: fmt.Sprintf("cluster--%s-%s", NAMESPACE, SINGLEPORTSVC)}
	vsCache, found := mcache.VsCacheMeta.AviCacheGet(vsKey)
	if !found {
		t.Fatalf("Cache not found for VS: %v", vsKey)
	} else {
		vsCacheObj, ok := vsCache.(*cache.AviVsCache)
		if !ok {
			t.Fatalf("Invalid VS object. Cannot cast.")
		}
		g.Expect(vsCacheObj.Name).To(gomega.Equal(fmt.Sprintf("cluster--%s-%s", NAMESPACE, SINGLEPORTSVC)))
		g.Expect(vsCacheObj.Tenant).To(gomega.Equal(AVINAMESPACE))
		g.Expect(vsCacheObj.PoolKeyCollection).To(gomega.HaveLen(1))
		g.Expect(vsCacheObj.PoolKeyCollection[0].Name).To(gomega.MatchRegexp("cluster--red-ns-testsvc-TCP-8080"))
		g.Expect(vsCacheObj.L4PolicyCollection).To(gomega.HaveLen(1))
		g.Expect(vsCacheObj.L4PolicyCollection[0].Name).To(gomega.MatchRegexp("cluster--red-ns-testsvc"))
	}

	TearDownTestForSvcLB(t, g)
	g.Eventually(func() bool {
		_, found := mcache.VsCacheMeta.AviCacheGet(vsKey)
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

		rModelName := ""
		if r.Method == "POST" && !strings.Contains(url, "login") {
			data, _ := io.ReadAll(r.Body)
			json.Unmarshal(data, &resp)
			if strings.Contains(url, "virtualservice") && injectFault {
				injectFault = false
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprintln(w, `{"error": "bad request"}`)
			} else {
				if strings.Contains(url, "virtualservice") {
					rModelName = "virtualservice"
				} else if strings.Contains(url, "vsvip") {
					rModelName = "vsvip"
				} else if strings.Contains(url, "l4policy") {
					rModelName = "l4policyset"
				}
				rName := resp["name"].(string)
				objURL := fmt.Sprintf("https://localhost/api/%s/%s-%s#%s", rModelName, rModelName, RANDOMUUID, rName)

				// adding additional 'uuid' and 'url' (read-only) fields in the response
				resp["url"] = objURL
				resp["uuid"] = fmt.Sprintf("%s-%s-%s", rModelName, rName, RANDOMUUID)
				finalResponse, _ = json.Marshal(resp)
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
	vsKey := cache.NamespaceName{Namespace: AVINAMESPACE, Name: fmt.Sprintf("cluster--%s-%s", NAMESPACE, SINGLEPORTSVC)}
	g.Eventually(func() bool {
		_, found := mcache.VsCacheMeta.AviCacheGet(vsKey)
		return found
	}, 10*time.Second).Should(gomega.Equal(true))
	vsCache, found := mcache.VsCacheMeta.AviCacheGet(vsKey)
	if !found {
		t.Fatalf("Cache not found for VS: %v", vsKey)
	} else {
		vsCacheObj, ok := vsCache.(*cache.AviVsCache)
		if !ok {
			t.Fatalf("Invalid VS object. Cannot cast.")
		}
		g.Expect(vsCacheObj.Name).To(gomega.Equal(fmt.Sprintf("cluster--%s-%s", NAMESPACE, SINGLEPORTSVC)))
		g.Expect(vsCacheObj.Tenant).To(gomega.Equal(AVINAMESPACE))
		g.Expect(vsCacheObj.PoolKeyCollection).To(gomega.HaveLen(1))
		g.Expect(vsCacheObj.PoolKeyCollection[0].Name).To(gomega.MatchRegexp("cluster--red-ns-testsvc-TCP-8080"))
		g.Expect(vsCacheObj.L4PolicyCollection).To(gomega.HaveLen(1))
		g.Expect(vsCacheObj.L4PolicyCollection[0].Name).To(gomega.MatchRegexp("cluster--red-ns-testsvc"))
	}

	TearDownTestForSvcLB(t, g)
}

func TestCreateMultiportServiceLBCacheSync(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	MULTIPORTSVC, NAMESPACE, AVINAMESPACE := "testsvcmulti", "red-ns", "admin"

	SetUpTestForSvcLBMultiport(t)

	mcache := cache.SharedAviObjCache()
	vsKey := cache.NamespaceName{Namespace: AVINAMESPACE, Name: fmt.Sprintf("cluster--%s-%s", NAMESPACE, MULTIPORTSVC)}
	vsCache, found := mcache.VsCacheMeta.AviCacheGet(vsKey)
	if !found {
		t.Fatalf("Cache not found for VS: %v", vsKey)
	}
	vsCacheObj, ok := vsCache.(*cache.AviVsCache)
	if !ok {
		t.Fatalf("Invalid VS object. Cannot cast.")
	}
	g.Expect(vsCacheObj.Name).To(gomega.Equal(fmt.Sprintf("cluster--%s-%s", NAMESPACE, MULTIPORTSVC)))
	g.Expect(vsCacheObj.Tenant).To(gomega.Equal(AVINAMESPACE))
	g.Expect(vsCacheObj.PoolKeyCollection).To(gomega.HaveLen(3))
	g.Expect(vsCacheObj.PoolKeyCollection[0].Name).To(gomega.MatchRegexp(`^(cluster--[a-zA-Z0-9-]+-808(0|1|2))$`))
	g.Expect(vsCacheObj.L4PolicyCollection).To(gomega.HaveLen(1))
	g.Expect(vsCacheObj.L4PolicyCollection[0].Name).To(gomega.MatchRegexp(`^(cluster--[a-zA-Z0-9-]+)$`))

	TearDownTestForSvcLBMultiport(t, g)
}

func TestUpdateAndDeleteServiceLBCacheSync(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	var err error

	SetUpTestForSvcLB(t)

	// Get hold of the pool checksum on CREATE
	poolName := "cluster--red-ns-testsvc-TCP-8080"
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
	if _, err = KubeClient.CoreV1().Endpoints(NAMESPACE).Update(context.TODO(), epExample, metav1.UpdateOptions{}); err != nil {
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

// TestScaleUpAndDownServiceLBCacheSync tests the avi node graph and rest layer functionality when the
// multiport serviceLB is increased from 1 to 5 and then decreased back to 1
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

	// create numScale more multiport service of type loadbalancer
	numScale := 5
	for i := 0; i < numScale; i++ {
		service = fmt.Sprintf("%s%d", MULTIPORTSVC, i)
		model = strings.Replace(MULTIPORTMODEL, MULTIPORTSVC, service, 1)

		objects.SharedAviGraphLister().Delete(model)
		CreateSVC(t, NAMESPACE, service, corev1.ServiceTypeLoadBalancer, true)
		CreateEP(t, NAMESPACE, service, true, true, "1.1.1")
	}

	// verify that numScale services are created on the graph and corresponding cache objects
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
			_, found = mcache.VsCacheMeta.AviCacheGet(vsKey)
			return found
		}, 15*time.Second).Should(gomega.Equal(true))
	}

	// delete the numScale services
	for i := 0; i < numScale; i++ {
		service = fmt.Sprintf("%s%d", MULTIPORTSVC, i)
		model = strings.Replace(MULTIPORTMODEL, MULTIPORTSVC, service, 1)
		objects.SharedAviGraphLister().Delete(model)
		DelSVC(t, NAMESPACE, service)
		DelEP(t, NAMESPACE, service)
	}

	// verify that the graph nodes and corresponding cache are deleted for the numScale services
	for i := 0; i < numScale; i++ {
		service = fmt.Sprintf("%s%d", MULTIPORTSVC, i)
		model = strings.Replace(MULTIPORTMODEL, MULTIPORTSVC, service, 1)
		g.Eventually(func() interface{} {
			found, aviModel = objects.SharedAviGraphLister().Get(model)
			return aviModel
		}, 40*time.Second).Should(gomega.BeNil())

		vsKey = cache.NamespaceName{Namespace: AVINAMESPACE, Name: strings.TrimPrefix(model, AVINAMESPACE+"/")}
		g.Eventually(func() bool {
			_, found = mcache.VsCacheMeta.AviCacheGet(vsKey)
			return found
		}, 60*time.Second).Should(gomega.Equal(false))
	}

	// verifying whether the first service created still has the corresponding cache entry
	vsKey = cache.NamespaceName{Namespace: AVINAMESPACE, Name: fmt.Sprintf("cluster--%s-%s", NAMESPACE, SINGLEPORTSVC)}
	g.Eventually(func() bool {
		_, found = mcache.VsCacheMeta.AviCacheGet(vsKey)
		return found
	}, 40*time.Second).Should(gomega.Equal(true))
	TearDownTestForSvcLB(t, g)
}

func TestAviSvcCreationWithStaticIP(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	staticIP := "80.80.80.80"
	objects.SharedAviGraphLister().Delete(SINGLEPORTMODEL)
	svcExample := (FakeService{
		Name:           SINGLEPORTSVC,
		Namespace:      NAMESPACE,
		Type:           corev1.ServiceTypeLoadBalancer,
		LoadBalancerIP: staticIP,
		ServicePorts:   []Serviceport{{PortName: "foo1", Protocol: "TCP", PortNumber: 8080, TargetPort: intstr.FromInt(8080)}},
	}).Service()
	_, err := KubeClient.CoreV1().Services(NAMESPACE).Create(context.TODO(), svcExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in creating Service: %v", err)
	}
	CreateEP(t, NAMESPACE, SINGLEPORTSVC, false, false, "1.1.1")
	PollForCompletion(t, SINGLEPORTMODEL, 5)

	g.Eventually(func() string {
		if found, aviModel := objects.SharedAviGraphLister().Get(SINGLEPORTMODEL); found && aviModel != nil {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
			if len(nodes) > 0 && len(nodes[0].VSVIPRefs) > 0 {
				return nodes[0].VSVIPRefs[0].IPAddress
			}
		}
		return ""
	}, 20*time.Second).Should(gomega.Equal(staticIP))
	TearDownTestForSvcLB(t, g)
}

// Infra CRD tests via service annotation

func TestWithInfraSettingStatusUpdates(t *testing.T) {
	// create infraSetting, svcLB with bad seGroup/networkName
	// check for Rejected status, check layer 2 for defaults
	// change to good seGroup/networkName, check for Accepted status
	// check layer 2 model

	g := gomega.NewGomegaWithT(t)
	settingName := "infra-setting"

	objects.SharedAviGraphLister().Delete(SINGLEPORTMODEL)
	svcExample := (FakeService{
		Name:         SINGLEPORTSVC,
		Namespace:    NAMESPACE,
		Type:         corev1.ServiceTypeLoadBalancer,
		ServicePorts: []Serviceport{{PortName: "foo1", Protocol: "TCP", PortNumber: 8080, TargetPort: intstr.FromInt(8080)}},
	}).Service()
	svcExample.Annotations = map[string]string{lib.InfraSettingNameAnnotation: settingName}
	_, err := KubeClient.CoreV1().Services(NAMESPACE).Create(context.TODO(), svcExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in creating Service: %v", err)
	}
	CreateEP(t, NAMESPACE, SINGLEPORTSVC, false, false, "1.1.1")
	PollForCompletion(t, SINGLEPORTMODEL, 5)

	// Create with bad seGroup ref.
	settingCreate := (FakeAviInfraSetting{
		Name:        settingName,
		SeGroupName: "thisisBADaviref-seGroup",
		Networks:    []string{"thisisaviref-networkName"},
		EnableRhi:   true,
	}).AviInfraSetting()
	if _, err := lib.AKOControlConfig().CRDClientset().AkoV1alpha1().AviInfraSettings().Create(context.TODO(), settingCreate, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding AviInfraSetting: %v", err)
	}

	g.Eventually(func() string {
		setting, _ := CRDClient.AkoV1alpha1().AviInfraSettings().Get(context.TODO(), settingName, metav1.GetOptions{})
		return setting.Status.Status
	}, 15*time.Second).Should(gomega.Equal("Rejected"))

	// defaults to global seGroup and networkName.
	netList := lib.GetVipNetworkList()
	g.Eventually(func() bool {
		if found, aviModel := objects.SharedAviGraphLister().Get(SINGLEPORTMODEL); found && aviModel != nil {
			if nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS(); len(nodes) > 0 {
				return nodes[0].ServiceEngineGroup == lib.GetSEGName() &&
					len(nodes[0].VSVIPRefs[0].VipNetworks) > 0 &&
					nodes[0].VSVIPRefs[0].VipNetworks[0].NetworkName == netList[0].NetworkName &&
					!*nodes[0].EnableRhi
			}
		}
		return false
	}, 40*time.Second).Should(gomega.Equal(true))
	settingUpdate := (FakeAviInfraSetting{
		Name:        settingName,
		SeGroupName: "thisisaviref-seGroup",
		Networks:    []string{"thisisaviref-networkName"},
		EnableRhi:   true,
	}).AviInfraSetting()
	settingUpdate.ResourceVersion = "2"
	if _, err := lib.AKOControlConfig().CRDClientset().AkoV1alpha1().AviInfraSettings().Update(context.TODO(), settingUpdate, metav1.UpdateOptions{}); err != nil {
		t.Fatalf("error in updating AviInfraSetting: %v", err)
	}

	g.Eventually(func() string {
		setting, _ := lib.AKOControlConfig().CRDClientset().AkoV1alpha1().AviInfraSettings().Get(context.TODO(), settingName, metav1.GetOptions{})
		return setting.Status.Status
	}, 15*time.Second).Should(gomega.Equal("Accepted"))

	g.Eventually(func() bool {
		if found, aviModel := objects.SharedAviGraphLister().Get(SINGLEPORTMODEL); found && aviModel != nil {
			if nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS(); len(nodes) > 0 {
				return nodes[0].ServiceEngineGroup == "thisisaviref-seGroup" &&
					len(nodes[0].VSVIPRefs[0].VipNetworks) > 0 &&
					nodes[0].VSVIPRefs[0].VipNetworks[0].NetworkName == "thisisaviref-networkName" &&
					*nodes[0].EnableRhi
			}
		}
		return false
	}, 45*time.Second).Should(gomega.Equal(true))

	settingUpdate = (FakeAviInfraSetting{
		Name:        settingName,
		SeGroupName: "thisisaviref-seGroup",
		Networks:    []string{"thisisBADaviref-networkName"},
		EnableRhi:   true,
	}).AviInfraSetting()
	settingUpdate.ResourceVersion = "3"
	if _, err := lib.AKOControlConfig().CRDClientset().AkoV1alpha1().AviInfraSettings().Update(context.TODO(), settingUpdate, metav1.UpdateOptions{}); err != nil {
		t.Fatalf("error in updating AviInfraSetting: %v", err)
	}

	g.Eventually(func() string {
		setting, _ := lib.AKOControlConfig().CRDClientset().AkoV1alpha1().AviInfraSettings().Get(context.TODO(), settingName, metav1.GetOptions{})
		return setting.Status.Status
	}, 15*time.Second).Should(gomega.Equal("Rejected"))
	g.Eventually(func() bool {
		if found, aviModel := objects.SharedAviGraphLister().Get(SINGLEPORTMODEL); found && aviModel != nil {
			if nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS(); len(nodes) > 0 {
				return nodes[0].ServiceEngineGroup == lib.GetSEGName() &&
					len(nodes[0].VSVIPRefs[0].VipNetworks) > 0 &&
					nodes[0].VSVIPRefs[0].VipNetworks[0].NetworkName == netList[0].NetworkName &&
					!*nodes[0].EnableRhi
			}
		}
		return false
	}, 40*time.Second).Should(gomega.Equal(true))

	settingUpdate = (FakeAviInfraSetting{
		Name:        settingName,
		SeGroupName: "thisisaviref-seGroup",
		Networks:    []string{"multivip-network1", "multivip-network2", "multivip-network3"},
		EnableRhi:   true,
	}).AviInfraSetting()
	settingUpdate.ResourceVersion = "4"
	if _, err := lib.AKOControlConfig().CRDClientset().AkoV1alpha1().AviInfraSettings().Update(context.TODO(), settingUpdate, metav1.UpdateOptions{}); err != nil {
		t.Fatalf("error in updating AviInfraSetting: %v", err)
	}

	g.Eventually(func() string {
		setting, _ := lib.AKOControlConfig().CRDClientset().AkoV1alpha1().AviInfraSettings().Get(context.TODO(), settingName, metav1.GetOptions{})
		return setting.Status.Status
	}, 15*time.Second).Should(gomega.Equal("Accepted"))

	g.Eventually(func() string {
		if found, aviModel := objects.SharedAviGraphLister().Get(SINGLEPORTMODEL); found && aviModel != nil {
			if nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS(); len(nodes) > 0 {
				return nodes[0].ServiceEngineGroup
			}
		}
		return ""
	}, 35*time.Second).Should(gomega.Equal("thisisaviref-seGroup"))
	_, aviModel := objects.SharedAviGraphLister().Get(SINGLEPORTMODEL)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(len(nodes[0].VSVIPRefs[0].VipNetworks)).Should(gomega.Equal(3))
	g.Expect(nodes[0].VSVIPRefs[0].VipNetworks[0].NetworkName).Should(gomega.Equal("multivip-network1"))
	g.Expect(nodes[0].VSVIPRefs[0].VipNetworks[1].NetworkName).Should(gomega.Equal("multivip-network2"))
	g.Expect(nodes[0].VSVIPRefs[0].VipNetworks[2].NetworkName).Should(gomega.Equal("multivip-network3"))
	g.Expect(*nodes[0].EnableRhi).Should(gomega.Equal(true))

	TeardownAviInfraSetting(t, settingName)
	TearDownTestForSvcLB(t, g)
}

func TestInfraSettingDelete(t *testing.T) {
	// create infraSetting, svcLB
	// delete infraSetting, fallback to defaults

	g := gomega.NewGomegaWithT(t)
	settingName := "infra-setting"

	objects.SharedAviGraphLister().Delete(SINGLEPORTMODEL)
	svcExample := (FakeService{
		Name:         SINGLEPORTSVC,
		Namespace:    NAMESPACE,
		Type:         corev1.ServiceTypeLoadBalancer,
		ServicePorts: []Serviceport{{PortName: "foo1", Protocol: "TCP", PortNumber: 8080, TargetPort: intstr.FromInt(8080)}},
	}).Service()
	svcExample.Annotations = map[string]string{lib.InfraSettingNameAnnotation: settingName}
	_, err := KubeClient.CoreV1().Services(NAMESPACE).Create(context.TODO(), svcExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in creating Service: %v", err)
	}
	CreateEP(t, NAMESPACE, SINGLEPORTSVC, false, false, "1.1.1")
	PollForCompletion(t, SINGLEPORTMODEL, 5)

	SetupAviInfraSetting(t, settingName, "")

	g.Eventually(func() string {
		if found, aviModel := objects.SharedAviGraphLister().Get(SINGLEPORTMODEL); found && aviModel != nil {
			if nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS(); len(nodes) > 0 {
				return nodes[0].ServiceEngineGroup
			}
		}
		return ""
	}, 35*time.Second).Should(gomega.Equal("thisisaviref-" + settingName + "-seGroup"))
	_, aviModel := objects.SharedAviGraphLister().Get(SINGLEPORTMODEL)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes[0].VSVIPRefs[0].VipNetworks[0].NetworkName).Should(gomega.Equal("thisisaviref-" + settingName + "-networkName"))
	g.Expect(*nodes[0].EnableRhi).Should(gomega.Equal(true))

	TeardownAviInfraSetting(t, settingName)

	// defaults to global seGroup and networkName.
	netList := lib.GetVipNetworkList()
	g.Eventually(func() bool {
		if found, aviModel := objects.SharedAviGraphLister().Get(SINGLEPORTMODEL); found && aviModel != nil {
			if nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS(); len(nodes) > 0 {
				return nodes[0].ServiceEngineGroup == lib.GetSEGName() &&
					len(nodes[0].VSVIPRefs[0].VipNetworks) > 0 &&
					nodes[0].VSVIPRefs[0].VipNetworks[0].NetworkName == netList[0].NetworkName &&
					!*nodes[0].EnableRhi
			}
		}
		return true
	}, 20*time.Second).Should(gomega.Equal(true))

	TearDownTestForSvcLB(t, g)
}

func TestInfraSettingChangeMapping(t *testing.T) {
	// create 2 infraSettings, svcLB
	// update infraSetting from one to another in service annotation
	// check changed model

	g := gomega.NewGomegaWithT(t)

	settingName1, settingName2 := "infra-setting1", "infra-setting2"

	objects.SharedAviGraphLister().Delete(SINGLEPORTMODEL)
	svcExample := (FakeService{
		Name:         SINGLEPORTSVC,
		Namespace:    NAMESPACE,
		Type:         corev1.ServiceTypeLoadBalancer,
		ServicePorts: []Serviceport{{PortName: "foo1", Protocol: "TCP", PortNumber: 8080, TargetPort: intstr.FromInt(8080)}},
	}).Service()
	svcExample.Annotations = map[string]string{lib.InfraSettingNameAnnotation: settingName1}
	_, err := KubeClient.CoreV1().Services(NAMESPACE).Create(context.TODO(), svcExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in creating Service: %v", err)
	}
	CreateEP(t, NAMESPACE, SINGLEPORTSVC, false, false, "1.1.1")
	PollForCompletion(t, SINGLEPORTMODEL, 5)

	SetupAviInfraSetting(t, settingName1, "")
	SetupAviInfraSetting(t, settingName2, "")

	g.Eventually(func() string {
		if found, aviModel := objects.SharedAviGraphLister().Get(SINGLEPORTMODEL); found && aviModel != nil {
			if nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS(); len(nodes) > 0 {
				return nodes[0].ServiceEngineGroup
			}
		}
		return ""
	}, 35*time.Second).Should(gomega.Equal("thisisaviref-" + settingName1 + "-seGroup"))
	_, aviModel := objects.SharedAviGraphLister().Get(SINGLEPORTMODEL)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes[0].VSVIPRefs[0].VipNetworks[0].NetworkName).Should(gomega.Equal("thisisaviref-" + settingName1 + "-networkName"))

	// TODO: Change service annotation to have infraSettting2
	svcUpdate := (FakeService{
		Name:         SINGLEPORTSVC,
		Namespace:    NAMESPACE,
		Type:         corev1.ServiceTypeLoadBalancer,
		ServicePorts: []Serviceport{{PortName: "foo1", Protocol: "TCP", PortNumber: 8080, TargetPort: intstr.FromInt(8080)}},
	}).Service()
	svcUpdate.Annotations = map[string]string{lib.InfraSettingNameAnnotation: settingName2}
	svcUpdate.ResourceVersion = "2"
	_, err = KubeClient.CoreV1().Services(NAMESPACE).Update(context.TODO(), svcUpdate, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating Service: %v", err)
	}

	g.Eventually(func() bool {
		if found, aviModel := objects.SharedAviGraphLister().Get(SINGLEPORTMODEL); found && aviModel != nil {
			if nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS(); len(nodes) > 0 {
				return nodes[0].ServiceEngineGroup == "thisisaviref-"+settingName2+"-seGroup" &&
					len(nodes[0].VSVIPRefs[0].VipNetworks) > 0 &&
					nodes[0].VSVIPRefs[0].VipNetworks[0].NetworkName == "thisisaviref-"+settingName2+"-networkName"
			}
		}
		return false
	}, 35*time.Second).Should(gomega.Equal(true))

	TeardownAviInfraSetting(t, settingName1)
	TeardownAviInfraSetting(t, settingName2)
	TearDownTestForSvcLB(t, g)
}
