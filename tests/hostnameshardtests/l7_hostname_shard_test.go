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

package hostnameshardtests

import (
	"os"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/avinetworks/ako/internal/cache"
	crdfake "github.com/avinetworks/ako/internal/client/clientset/versioned/fake"
	"github.com/avinetworks/ako/internal/k8s"
	"github.com/avinetworks/ako/internal/lib"
	avinodes "github.com/avinetworks/ako/internal/nodes"
	"github.com/avinetworks/ako/internal/objects"
	"github.com/avinetworks/ako/tests/integrationtest"

	utils "github.com/avinetworks/ako/pkg/utils"

	"github.com/avinetworks/sdk/go/models"
	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

func TestMain(m *testing.M) {
	os.Setenv("INGRESS_API", "extensionv1")
	os.Setenv("NETWORK_NAME", "net123")
	os.Setenv("CLUSTER_NAME", "cluster")
	os.Setenv("CLOUD_NAME", "Default-Cloud")
	os.Setenv("SEG_NAME", "Default-Group")
	os.Setenv("NODE_NETWORK_LIST", `[{"networkName":"net123","cidrs":["10.79.168.0/22"]}]`)

	KubeClient = k8sfake.NewSimpleClientset()
	CRDClient = crdfake.NewSimpleClientset()
	lib.SetCRDClientset(CRDClient)

	registeredInformers := []string{
		utils.ServiceInformer,
		utils.EndpointInformer,
		utils.IngressInformer,
		utils.SecretInformer,
		utils.NSInformer,
		utils.NodeInformer,
		utils.ConfigMapInformer,
	}
	utils.NewInformers(utils.KubeClientIntf{KubeClient}, registeredInformers)
	informers := k8s.K8sinformers{Cs: KubeClient}
	k8s.NewCRDInformers(CRDClient)

	mcache := cache.SharedAviObjCache()
	cloudObj := &cache.AviCloudPropertyCache{Name: "Default-Cloud", VType: "mock"}
	subdomains := []string{"avi.internal", ".com"}
	cloudObj.NSIpamDNS = subdomains
	mcache.CloudKeyCache.AviCacheAdd("Default-Cloud", cloudObj)

	integrationtest.InitializeFakeAKOAPIServer()

	integrationtest.NewAviFakeClientInstance()
	defer integrationtest.AviFakeClientInstance.Close()

	ctrl = k8s.SharedAviController()
	stopCh := utils.SetupSignalHandler()
	ctrlCh := make(chan struct{})
	quickSyncCh := make(chan struct{})
	waitGroupMap := make(map[string]*sync.WaitGroup)
	wgIngestion := &sync.WaitGroup{}
	waitGroupMap["ingestion"] = wgIngestion
	wgFastRetry := &sync.WaitGroup{}
	waitGroupMap["fastretry"] = wgFastRetry
	wgGraph := &sync.WaitGroup{}
	waitGroupMap["graph"] = wgGraph
	ctrl.HandleConfigMap(informers, ctrlCh, stopCh, quickSyncCh)
	go ctrl.InitController(informers, registeredInformers, ctrlCh, stopCh, quickSyncCh, waitGroupMap)
	AddConfigMap()
	integrationtest.KubeClient = KubeClient
	os.Exit(m.Run())
}

var KubeClient *k8sfake.Clientset
var CRDClient *crdfake.Clientset
var ctrl *k8s.AviController

func AddConfigMap() {
	aviCM := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "avi-system",
			Name:      "avi-k8s-config",
		},
	}
	KubeClient.CoreV1().ConfigMaps("avi-system").Create(aviCM)

	integrationtest.PollForSyncStart(ctrl, 10)
}

func SetUpTestForIngress(t *testing.T, modelName string) {
	os.Setenv("SHARD_VS_SIZE", "LARGE")
	os.Setenv("L7_SHARD_SCHEME", "hostname")

	objects.SharedAviGraphLister().Delete(modelName)
	integrationtest.CreateSVC(t, "default", "avisvc", corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEP(t, "default", "avisvc", false, false, "1.1.1")
}

func TearDownTestForIngress(t *testing.T, modelName string) {
	os.Setenv("SHARD_VS_SIZE", "")
	os.Setenv("CLOUD_NAME", "")

	objects.SharedAviGraphLister().Delete(modelName)
	integrationtest.DelSVC(t, "default", "avisvc")
	integrationtest.DelEP(t, "default", "avisvc")
}

func VerifyIngressDeletion(t *testing.T, g *gomega.WithT, aviModel interface{}, poolCount int) {
	var nodes []*avinodes.AviVsNode
	g.Eventually(func() []*avinodes.AviPoolNode {
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return nodes[0].PoolRefs
	}, 10*time.Second).Should(gomega.HaveLen(poolCount))

	g.Eventually(func() []*models.PoolGroupMember {
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return nodes[0].PoolGroupRefs[0].Members
	}, 10*time.Second).Should(gomega.HaveLen(poolCount))
}

func VerifySNIIngressDeletion(t *testing.T, g *gomega.WithT, aviModel interface{}, sniCount int) {
	var nodes []*avinodes.AviVsNode
	g.Eventually(func() []*avinodes.AviVsNode {
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return nodes[0].SniNodes
	}, 10*time.Second).Should(gomega.HaveLen(sniCount))

	g.Expect(len(nodes[0].PoolGroupRefs[0].Members)).To(gomega.Equal(sniCount))
}

func TestL7Model(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName := "admin/cluster--Shared-L7-0"
	SetUpTestForIngress(t, modelName)

	integrationtest.PollForCompletion(t, modelName, 5)
	found, _ := objects.SharedAviGraphLister().Get(modelName)
	if found {
		// We shouldn't get an update for this update since it neither belongs to an ingress nor a L4 LB service
		t.Fatalf("Couldn't find Model for DELETE event %v", modelName)
	}
	ingrFake := (integrationtest.FakeIngress{
		Name:        "foo-with-targets",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		ServiceName: "avisvc",
	}).Ingress()

	_, err := KubeClient.ExtensionsV1beta1().Ingresses("default").Create(ingrFake)
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	integrationtest.PollForCompletion(t, modelName, 5)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 5*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(len(nodes)).To(gomega.Equal(1))
	g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
	g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))

	err = KubeClient.ExtensionsV1beta1().Ingresses("default").Delete("foo-with-targets", nil)
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngress(t, modelName)
}

func TestMultiIngressToSameSvc(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	os.Setenv("SHARD_VS_SIZE", "LARGE")

	modelName := "admin/cluster--Shared-L7-0"
	objects.SharedAviGraphLister().Delete(modelName)
	svcExample := (integrationtest.FakeService{
		Name:         "avisvc",
		Namespace:    "default",
		Type:         corev1.ServiceTypeClusterIP,
		ServicePorts: []integrationtest.Serviceport{{PortName: "foo", Protocol: "TCP", PortNumber: 8080, TargetPort: 8080}},
	}).Service()

	_, err := KubeClient.CoreV1().Services("default").Create(svcExample)
	if err != nil {
		t.Fatalf("error in adding Service: %v", err)
	}
	epExample := &corev1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "avisvc",
		},
		Subsets: []corev1.EndpointSubset{{
			Addresses: []corev1.EndpointAddress{{IP: "1.2.3.4"}},
			Ports:     []corev1.EndpointPort{{Name: "foo", Port: 8080, Protocol: "TCP"}},
		}},
	}
	_, err = KubeClient.CoreV1().Endpoints("default").Create(epExample)
	if err != nil {
		t.Fatalf("error in creating Endpoint: %v", err)
	}
	ingrFake1 := (integrationtest.FakeIngress{
		Name:        "foo-with-targets1",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		ServiceName: "avisvc",
	}).Ingress()

	_, err = KubeClient.ExtensionsV1beta1().Ingresses("default").Create(ingrFake1)
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	ingrFake2 := (integrationtest.FakeIngress{
		Name:        "foo-with-targets2",
		Namespace:   "default",
		DnsNames:    []string{"bar.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		ServiceName: "avisvc",
	}).Ingress()

	_, err = KubeClient.ExtensionsV1beta1().Ingresses("default").Create(ingrFake2)
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	integrationtest.PollForCompletion(t, modelName, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(modelName)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Expect(nodes[0].SharedVS).To(gomega.Equal(true))
		dsNodes := aviModel.(*avinodes.AviObjectGraph).GetAviHTTPDSNode()
		g.Expect(len(dsNodes)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(1))
		for _, pool := range nodes[0].PoolRefs {
			if pool.Name == "cluster--foo.com_foo-default-foo-with-targets1" {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("foo.com/foo"))
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
			} else {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("bar.com/foo"))
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
			}
		}
		// Delete the model.
		objects.SharedAviGraphLister().Delete(modelName)
	} else {
		t.Fatalf("Could not find model on ingress delete: %v", err)
	}
	//====== VERIFICATION OF SERVICE DELETE
	// Now we have cleared the layer 2 queue for both the models. Let's delete the service.
	err = KubeClient.CoreV1().Endpoints("default").Delete("avisvc", nil)
	if err != nil {
		t.Fatalf("Couldn't DELETE the Endpoint %v", err)
	}
	err = KubeClient.CoreV1().Services("default").Delete("avisvc", nil)
	if err != nil {
		t.Fatalf("Couldn't DELETE the Service %v", err)
	}
	// We should be able to get one model now in the queue
	integrationtest.PollForCompletion(t, modelName, 5)
	found, aviModel = objects.SharedAviGraphLister().Get(modelName)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		dsNodes := aviModel.(*avinodes.AviObjectGraph).GetAviHTTPDSNode()
		g.Expect(len(dsNodes)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(1))

	} else {
		t.Fatalf("Could not find model on service delete: %v", err)
	}
	_, err = KubeClient.CoreV1().Endpoints("default").Create(epExample)
	if err != nil {
		t.Fatalf("error in creating Endpoint: %v", err)
	}
	//====== VERIFICATION OF ONE INGRESS DELETE
	// Now let's delete one ingress and expect the update for that.
	err = KubeClient.ExtensionsV1beta1().Ingresses("default").Delete("foo-with-targets1", nil)
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	integrationtest.DetectModelChecksumChange(t, modelName, 5)
	found, aviModel = objects.SharedAviGraphLister().Get(modelName)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(0))

		// Delete the model.
		objects.SharedAviGraphLister().Delete(modelName)
	} else {
		t.Fatalf("Could not find model on ingress delete: %v", err)
	}
	//====== VERIFICATION OF SERVICE ADD
	// Let's add the service back now - the ingress's associated with this service should be returned
	_, err = KubeClient.CoreV1().Services("default").Create(svcExample)
	if err != nil {
		t.Fatalf("error in adding Service: %v", err)
	}
	modelName = "admin/cluster--Shared-L7-1"
	integrationtest.PollForCompletion(t, modelName, 5)
	// We should be able to get one model now in the queue
	found, aviModel = objects.SharedAviGraphLister().Get(modelName)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(1))

		objects.SharedAviGraphLister().Delete(modelName)
	} else {
		t.Fatalf("Could not find model on service ADD: %v", err)
	}
	//====== VERIFICATION OF ONE ENDPOINT DELETE
	err = KubeClient.CoreV1().Endpoints("default").Delete("avisvc", nil)
	if err != nil {
		t.Fatalf("Couldn't DELETE the Endpoint %v", err)
	}
	integrationtest.PollForCompletion(t, modelName, 5)
	// Deletion should also give us the affected ingress objects
	found, aviModel = objects.SharedAviGraphLister().Get(modelName)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		// Delete the model.
		objects.SharedAviGraphLister().Delete(modelName)
	} else {
		t.Fatalf("Could not find model on service ADD: %v", err)
	}
	err = KubeClient.CoreV1().Services("default").Delete("avisvc", nil)
	if err != nil {
		t.Fatalf("Couldn't DELETE the Service %v", err)
	}
	err = KubeClient.ExtensionsV1beta1().Ingresses("default").Delete("foo-with-targets2", nil)
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
}

func TestMultiVSIngress(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName := "admin/cluster--Shared-L7-0"
	SetUpTestForIngress(t, modelName)

	ingrFake := (integrationtest.FakeIngress{
		Name:        "foo-with-targets",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		ServiceName: "avisvc",
	}).Ingress()

	_, err := KubeClient.ExtensionsV1beta1().Ingresses("default").Create(ingrFake)
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	integrationtest.PollForCompletion(t, modelName, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(modelName)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		dsNodes := aviModel.(*avinodes.AviObjectGraph).GetAviHTTPDSNode()
		g.Expect(len(dsNodes)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(1))
		for _, pool := range nodes[0].PoolRefs {
			if pool.Name == "cluster--foo.com_foo" {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("foo.com/foo"))
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
			}
		}
	} else {
		t.Fatalf("Could not find model: %v", err)
	}
	randoming := (integrationtest.FakeIngress{
		Name:        "randomNamespacethatyeildsdiff",
		Namespace:   "randomNamespacethatyeildsdiff",
		DnsNames:    []string{"foo.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		ServiceName: "avisvc",
	}).Ingress()
	_, err = KubeClient.ExtensionsV1beta1().Ingresses("randomNamespacethatyeildsdiff").Create(randoming)
	integrationtest.PollForCompletion(t, modelName, 10)
	found, aviModel = objects.SharedAviGraphLister().Get(modelName)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		dsNodes := aviModel.(*avinodes.AviObjectGraph).GetAviHTTPDSNode()
		g.Expect(len(dsNodes)).To(gomega.Equal(1))
		g.Eventually(func() int {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
			return len(nodes[0].PoolRefs)
		}, 10*time.Second).Should(gomega.Equal(2))

	} else {
		t.Fatalf("Could not find model: %v", err)
	}
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	err = KubeClient.ExtensionsV1beta1().Ingresses("default").Delete("foo-with-targets", nil)
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	err = KubeClient.ExtensionsV1beta1().Ingresses("randomNamespacethatyeildsdiff").Delete("randomNamespacethatyeildsdiff", nil)
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngress(t, modelName)
}

func TestMultiPathIngress(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	var err error

	modelName := "admin/cluster--Shared-L7-0"
	SetUpTestForIngress(t, modelName)

	ingrFake := (integrationtest.FakeIngress{
		Name:        "ingress-multipath",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/foo", "/bar"},
		ServiceName: "avisvc",
	}).IngressMultiPath()

	_, err = KubeClient.ExtensionsV1beta1().Ingresses("default").Create(ingrFake)
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	integrationtest.PollForCompletion(t, modelName, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(modelName)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(2))
		for _, pool := range nodes[0].PoolRefs {
			if pool.Name == "cluster--foo.com_foo-default-ingress-multipath" {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("foo.com/foo"))
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
			} else if pool.Name == "cluster--foo.com_bar-default-ingress-multipath" {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("foo.com/bar"))
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
			} else {
				t.Fatalf("unexpected pool: %s", pool.Name)
			}
		}
		g.Expect(len(nodes[0].PoolGroupRefs)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].PoolGroupRefs[0].Members)).To(gomega.Equal(2))
		for _, pool := range nodes[0].PoolGroupRefs[0].Members {
			if *pool.PoolRef == "/api/pool?name=cluster--foo.com_foo-default-ingress-multipath" {
				g.Expect(*pool.PriorityLabel).To(gomega.Equal("foo.com/foo"))
			} else if *pool.PoolRef == "/api/pool?name=cluster--foo.com_bar-default-ingress-multipath" {
				g.Expect(*pool.PriorityLabel).To(gomega.Equal("foo.com/bar"))
			} else {
				t.Fatalf("unexpected pool: %s", *pool.PoolRef)
			}
		}
	} else {
		t.Fatalf("Could not find model: %s", modelName)
	}
	err = KubeClient.ExtensionsV1beta1().Ingresses("default").Delete("ingress-multipath", nil)
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngress(t, modelName)
}

func TestMultiPortServiceIngress(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	var err error

	modelName := "admin/cluster--Shared-L7-0"
	objects.SharedAviGraphLister().Delete(modelName)
	integrationtest.CreateSVC(t, "default", "avisvc", corev1.ServiceTypeClusterIP, true)
	integrationtest.CreateEP(t, "default", "avisvc", true, true, "1.1.1")
	ingrFake := (integrationtest.FakeIngress{
		Name:        "ingress-multipath",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/foo"},
		ServiceName: "avisvc",
	}).Ingress(true)

	_, err = KubeClient.ExtensionsV1beta1().Ingresses("default").Create(ingrFake)
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	integrationtest.PollForCompletion(t, modelName, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(modelName)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(2))
		for _, pool := range nodes[0].PoolRefs {
			if pool.Name == "cluster--foo.com_foo-default-ingress-multipath" {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("foo.com/foo"))
				g.Expect(pool.Port).To(gomega.Equal(int32(8080)))
				g.Expect(len(pool.Servers)).To(gomega.Equal(3))
			} else if pool.Name == "cluster--foo.com_bar-default-ingress-multipath" {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("foo.com/bar"))
				g.Expect(pool.Port).To(gomega.Equal(int32(8081)))
				g.Expect(len(pool.Servers)).To(gomega.Equal(2))
			} else {
				t.Fatalf("unexpected pool: %s", pool.Name)
			}
		}
		g.Expect(len(nodes[0].PoolGroupRefs)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].PoolGroupRefs[0].Members)).To(gomega.Equal(2))
		for _, pool := range nodes[0].PoolGroupRefs[0].Members {
			if *pool.PoolRef == "/api/pool?name=cluster--foo.com_foo-default-ingress-multipath" {
				g.Expect(*pool.PriorityLabel).To(gomega.Equal("foo.com/foo"))
			} else if *pool.PoolRef == "/api/pool?name=cluster--foo.com_bar-default-ingress-multipath" {
				g.Expect(*pool.PriorityLabel).To(gomega.Equal("foo.com/bar"))
			} else {
				t.Fatalf("unexpected pool: %s", *pool.PoolRef)
			}
		}
	} else {
		t.Fatalf("Could not find model: %s", modelName)
	}
	err = KubeClient.ExtensionsV1beta1().Ingresses("default").Delete("ingress-multipath", nil)
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngress(t, modelName)
}

func TestMultiIngressSameHost(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName := "admin/cluster--Shared-L7-0"
	SetUpTestForIngress(t, modelName)

	ingrFake1 := (integrationtest.FakeIngress{
		Name:        "ingress-multi1",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/foo"},
		ServiceName: "avisvc",
	}).Ingress()

	_, err := KubeClient.ExtensionsV1beta1().Ingresses("default").Create(ingrFake1)
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	ingrFake2 := (integrationtest.FakeIngress{
		Name:        "ingress-multi2",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/bar"},
		ServiceName: "avisvc",
	}).Ingress()

	_, err = KubeClient.ExtensionsV1beta1().Ingresses("default").Create(ingrFake2)
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	integrationtest.PollForCompletion(t, modelName, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(modelName)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(2))
		for _, pool := range nodes[0].PoolRefs {
			if pool.Name == "cluster--foo.com_foo-default-ingress-multi1" {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("foo.com/foo"))
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
			} else if pool.Name == "cluster--foo.com_bar-default-ingress-multi2" {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("foo.com/bar"))
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
			} else {
				t.Fatalf("unexpected pool: %s", pool.Name)
			}
		}
		g.Expect(len(nodes[0].PoolGroupRefs)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].PoolGroupRefs[0].Members)).To(gomega.Equal(2))
		for _, pool := range nodes[0].PoolGroupRefs[0].Members {
			if *pool.PoolRef == "/api/pool?name=cluster--foo.com_foo-default-ingress-multi1" {
				g.Expect(*pool.PriorityLabel).To(gomega.Equal("foo.com/foo"))
			} else if *pool.PoolRef == "/api/pool?name=cluster--foo.com_bar-default-ingress-multi2" {
				g.Expect(*pool.PriorityLabel).To(gomega.Equal("foo.com/bar"))
			} else {
				t.Fatalf("unexpected pool: %s", *pool.PoolRef)
			}
		}
	} else {
		t.Fatalf("Could not find model: %s", modelName)
	}
	err = KubeClient.ExtensionsV1beta1().Ingresses("default").Delete("ingress-multi1", nil)
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 1)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(len(nodes)).To(gomega.Equal(1))
	g.Expect(nodes[0].PoolRefs[0].Name).To(gomega.Equal("cluster--foo.com_bar-default-ingress-multi2"))

	err = KubeClient.ExtensionsV1beta1().Ingresses("default").Delete("ingress-multi2", nil)
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngress(t, modelName)
}

func TestDeleteBackendService(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName := "admin/cluster--Shared-L7-0"
	SetUpTestForIngress(t, modelName)

	ingrFake1 := (integrationtest.FakeIngress{
		Name:        "ingress-multi1",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/foo"},
		ServiceName: "avisvc",
	}).Ingress()

	_, err := KubeClient.ExtensionsV1beta1().Ingresses("default").Create(ingrFake1)
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	ingrFake2 := (integrationtest.FakeIngress{
		Name:        "ingress-multi2",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/bar"},
		ServiceName: "avisvc",
	}).Ingress()

	_, err = KubeClient.ExtensionsV1beta1().Ingresses("default").Create(ingrFake2)
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	integrationtest.PollForCompletion(t, modelName, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(modelName)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(2))
		for _, pool := range nodes[0].PoolRefs {
			if pool.Name == "cluster--foo.com_foo-default-ingress-multi1" {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("foo.com/foo"))
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
			} else if pool.Name == "cluster--foo.com_bar-default-ingress-multi2" {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("foo.com/bar"))
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
			} else {
				t.Fatalf("unexpected pool: %s", pool.Name)
			}
		}
		g.Expect(len(nodes[0].PoolGroupRefs)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].PoolGroupRefs[0].Members)).To(gomega.Equal(2))
		for _, pool := range nodes[0].PoolGroupRefs[0].Members {
			if *pool.PoolRef == "/api/pool?name=cluster--foo.com_foo-default-ingress-multi1" {
				g.Expect(*pool.PriorityLabel).To(gomega.Equal("foo.com/foo"))
			} else if *pool.PoolRef == "/api/pool?name=cluster--foo.com_bar-default-ingress-multi2" {
				g.Expect(*pool.PriorityLabel).To(gomega.Equal("foo.com/bar"))
			} else {
				t.Fatalf("unexpected pool: %s", *pool.PoolRef)
			}
		}
	} else {
		t.Fatalf("Could not find model: %s", modelName)
	}
	// Delete the service
	integrationtest.DelSVC(t, "default", "avisvc")
	integrationtest.DelEP(t, "default", "avisvc")
	found, aviModel = objects.SharedAviGraphLister().Get(modelName)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(2))

		g.Eventually(func() int {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
			return len(nodes[0].PoolRefs[0].Servers)
		}, 10*time.Second).Should(gomega.Equal(0))

		g.Eventually(func() int {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
			if len(nodes[0].PoolRefs) > 1 {
				return len(nodes[0].PoolRefs[1].Servers)
			}
			return 1
		}, 10*time.Second).Should(gomega.Equal(0))

		g.Expect(len(nodes[0].PoolGroupRefs)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].PoolGroupRefs[0].Members)).To(gomega.Equal(2))
	} else {
		t.Fatalf("Could not find model: %s", modelName)
	}
	err = KubeClient.ExtensionsV1beta1().Ingresses("default").Delete("ingress-multi1", nil)
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 1)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(len(nodes)).To(gomega.Equal(1))
	g.Expect(nodes[0].PoolRefs[0].Name).To(gomega.Equal("cluster--foo.com_bar-default-ingress-multi2"))

	err = KubeClient.ExtensionsV1beta1().Ingresses("default").Delete("ingress-multi2", nil)
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 0)

}

func TestUpdateBackendService(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	modelName := "admin/cluster--Shared-L7-0"
	SetUpTestForIngress(t, modelName)
	ingrFake1 := (integrationtest.FakeIngress{
		Name:        "ingress-backend-svc",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/foo"},
		ServiceName: "avisvc",
	}).Ingress()
	_, err := KubeClient.ExtensionsV1beta1().Ingresses("default").Create(ingrFake1)
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	integrationtest.PollForCompletion(t, modelName, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(modelName)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(*nodes[0].PoolRefs[0].Servers[0].Ip.Addr).To(gomega.Equal("1.1.1.1"))

	} else {
		t.Fatalf("Could not find model: %s", modelName)
	}
	// Update the service

	integrationtest.CreateSVC(t, "default", "avisvc2", corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEP(t, "default", "avisvc2", false, false, "2.2.2")

	ingrFake1, err = (integrationtest.FakeIngress{
		Name:        "ingress-backend-svc",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/foo"},
		ServiceName: "avisvc2",
	}).UpdateIngress()
	if err != nil {
		t.Fatalf("error in updating ingress %s", err)
	}

	found, aviModel = objects.SharedAviGraphLister().Get(modelName)
	if found {
		g.Eventually(func() string {
			_, aviModel := objects.SharedAviGraphLister().Get(modelName)
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
			return *nodes[0].PoolRefs[0].Servers[0].Ip.Addr
		}, 10*time.Second).Should(gomega.Equal("2.2.2.1"))
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(1))
	} else {
		t.Fatalf("Could not find model: %s", modelName)
	}
	err = KubeClient.ExtensionsV1beta1().Ingresses("default").Delete("ingress-backend-svc", nil)
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	integrationtest.DelSVC(t, "default", "avisvc2")
	integrationtest.DelEP(t, "default", "avisvc2")
	VerifyIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngress(t, modelName)
}

func TestL2ChecksumsUpdate(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	integrationtest.AddSecret("my-secret", "default", "tlsCert", "tlsKey")
	modelName := "admin/cluster--Shared-L7-0"
	SetUpTestForIngress(t, modelName)
	//create ingress with tls secret
	ingrFake1 := (integrationtest.FakeIngress{
		Name:      "ingress-chksum",
		Namespace: "default",
		DnsNames:  []string{"foo.com"},
		Ips:       []string{"8.8.8.8"},
		Paths:     []string{"/foo"},
		HostNames: []string{"v1"},
		TlsSecretDNS: map[string][]string{
			"my-secret": []string{"foo.com"},
		},
		ServiceName: "avisvc",
	}).Ingress()
	_, err := KubeClient.ExtensionsV1beta1().Ingresses("default").Create(ingrFake1)
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	initCheckSums := make(map[string]uint32)
	integrationtest.PollForCompletion(t, modelName, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(modelName)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		initCheckSums["nodes[0]"] = nodes[0].CloudConfigCksum

		g.Expect(len(nodes[0].SniNodes)).To(gomega.Equal(1))
		initCheckSums["nodes[0].SniNodes[0]"] = nodes[0].SniNodes[0].CloudConfigCksum

		g.Expect(len(nodes[0].SniNodes[0].PoolRefs)).To(gomega.Equal(1))
		initCheckSums["nodes[0].SniNodes[0].PoolRefs[0]"] = nodes[0].SniNodes[0].PoolRefs[0].CloudConfigCksum

		g.Expect(len(nodes[0].SniNodes[0].SSLKeyCertRefs)).To(gomega.Equal(1))
		initCheckSums["nodes[0].SniNodes[0].SSLKeyCertRefs[0]"] = nodes[0].SniNodes[0].SSLKeyCertRefs[0].CloudConfigCksum

		g.Expect(len(nodes[0].SniNodes[0].HttpPolicyRefs)).To(gomega.Equal(1))
		initCheckSums["nodes[0].SniNodes[0].HttpPolicyRefs[0]"] = nodes[0].SniNodes[0].HttpPolicyRefs[0].CloudConfigCksum

	} else {
		t.Fatalf("Could not find model: %s", modelName)
	}

	integrationtest.CreateSVC(t, "default", "avisvc2", corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEP(t, "default", "avisvc2", false, false, "2.2.2")
	integrationtest.AddSecret("my-secret-new", "default", "tlsCert-new", "tlsKey")

	ingrFake1, err = (integrationtest.FakeIngress{
		Name:      "ingress-chksum",
		Namespace: "default",
		DnsNames:  []string{"foo.com"},
		Ips:       []string{"8.8.8.8"},
		//to update httppolicyref checksum
		Paths:     []string{"/bar"},
		HostNames: []string{"v1"},
		TlsSecretDNS: map[string][]string{
			//to update tls secret checksum
			"my-secret-new": []string{"foo.com"},
		},
		//to update poolref checksum
		ServiceName: "avisvc2",
	}).UpdateIngress()
	if err != nil {
		t.Fatalf("error in updating ingress %s", err)
	}
	integrationtest.PollForCompletion(t, modelName, 5)
	found, aviModel = objects.SharedAviGraphLister().Get(modelName)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Eventually(len(nodes), 5*time.Second).Should(gomega.Equal(1))

		g.Expect(len(nodes[0].SniNodes)).To(gomega.Equal(1))
		g.Eventually(func() uint32 {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
			return nodes[0].SniNodes[0].CloudConfigCksum
		}, 5*time.Second).ShouldNot(gomega.Equal(initCheckSums["nodes[0].SniNodes[0]"]))

		g.Expect(len(nodes[0].SniNodes[0].PoolRefs)).To(gomega.Equal(1))
		g.Eventually(func() uint32 {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
			return nodes[0].SniNodes[0].PoolRefs[0].CloudConfigCksum
		}, 5*time.Second).ShouldNot(gomega.Equal(initCheckSums["nodes[0].SniNodes[0].PoolRefs[0]"]))

		g.Expect(len(nodes[0].SniNodes[0].SSLKeyCertRefs)).To(gomega.Equal(1))
		g.Eventually(func() uint32 {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
			return nodes[0].SniNodes[0].SSLKeyCertRefs[0].CloudConfigCksum
		}, 5*time.Second).ShouldNot(gomega.Equal(initCheckSums["nodes[0].SniNodes[0].SSLKeyCertRefs[0]"]))

		g.Expect(len(nodes[0].SniNodes[0].HttpPolicyRefs)).To(gomega.Equal(1))
		g.Eventually(func() uint32 {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
			return nodes[0].SniNodes[0].HttpPolicyRefs[0].CloudConfigCksum
		}, 5*time.Second).ShouldNot(gomega.Equal(initCheckSums["nodes[0].SniNodes[0].HttpPolicyRefs[0]"]))

	} else {
		t.Fatalf("Could not find model: %s", modelName)
	}
	err = KubeClient.ExtensionsV1beta1().Ingresses("default").Delete("ingress-chksum", nil)
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	integrationtest.DelSVC(t, "default", "avisvc2")
	integrationtest.DelEP(t, "default", "avisvc2")
	KubeClient.CoreV1().Secrets("default").Delete("my-secret", nil)
	KubeClient.CoreV1().Secrets("default").Delete("my-secret-new", nil)
	VerifySNIIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngress(t, modelName)
}

func TestSniHttpPolicy(t *testing.T) {
	/*
		-> Create Ingress with TLS key/secret and 2 paths
		-> Verify removing path works by updating Ingress with single path
		-> Verify adding path works by updating Ingress with 2 new paths
	*/

	g := gomega.NewGomegaWithT(t)
	integrationtest.AddSecret("my-secret", "default", "tlsCert", "tlsKey")
	modelName := "admin/cluster--Shared-L7-0"
	SetUpTestForIngress(t, modelName)
	ingrFake1 := (integrationtest.FakeIngress{
		Name:      "ingress-shp",
		Namespace: "default",
		DnsNames:  []string{"foo.com"},
		Ips:       []string{"8.8.8.8"},
		Paths:     []string{"/foo", "/bar"},
		HostNames: []string{"v1"},
		TlsSecretDNS: map[string][]string{
			"my-secret": []string{"foo.com"},
		},
		ServiceName: "avisvc",
	}).IngressMultiPath()
	_, err := KubeClient.ExtensionsV1beta1().Ingresses("default").Create(ingrFake1)
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	integrationtest.PollForCompletion(t, modelName, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(modelName)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Eventually(len(nodes), 30*time.Second).Should(gomega.Equal(1))
		g.Expect(len(nodes[0].SniNodes), gomega.Equal(1))
		g.Expect(len(nodes[0].SniNodes[0].HttpPolicyRefs), gomega.Equal(2))
		g.Expect(len(nodes[0].SniNodes[0].HttpPolicyRefs[0].HppMap[0].Path), gomega.Equal(1))
		g.Expect(len(nodes[0].SniNodes[0].HttpPolicyRefs[0].HppMap[0].Path), gomega.Equal(1))
		g.Expect(len(nodes[0].SniNodes[0].SSLKeyCertRefs), gomega.Equal(1))
		g.Expect(func() []string {
			p := []string{
				nodes[0].SniNodes[0].HttpPolicyRefs[0].HppMap[0].Path[0],
				nodes[0].SniNodes[0].HttpPolicyRefs[1].HppMap[0].Path[0]}
			sort.Strings(p)
			return p
		}, gomega.Equal([]string{"/bar", "/foo"}))
		g.Expect(func() []string {
			p := []string{
				nodes[0].SniNodes[0].HttpPolicyRefs[0].Name,
				nodes[0].SniNodes[0].HttpPolicyRefs[1].Name}
			sort.Strings(p)
			return p
		}, gomega.Equal([]string{"cluster--default-foo.com_bar-ingress-shp",
			"cluster--default-foo.com_foo-ingress-shp"}))

	} else {
		t.Fatalf("Could not find model: %s", modelName)
	}

	ingrFake1, err = (integrationtest.FakeIngress{
		Name:      "ingress-shp",
		Namespace: "default",
		DnsNames:  []string{"foo.com"},
		Ips:       []string{"8.8.8.8"},
		Paths:     []string{"/foo"},
		HostNames: []string{"v1"},
		TlsSecretDNS: map[string][]string{
			"my-secret": []string{"foo.com"},
		},
		ServiceName: "avisvc",
	}).UpdateIngress()
	if err != nil {
		t.Fatalf("error in updating ingress %s", err)
	}
	integrationtest.PollForCompletion(t, modelName, 5)

	g.Eventually(func() int {
		found, aviModel = objects.SharedAviGraphLister().Get(modelName)
		if found {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
			return len(nodes[0].SniNodes[0].HttpPolicyRefs)
		} else {
			return 0
		}
	}, 30*time.Second).Should(gomega.Equal(1))
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(len(nodes[0].SniNodes[0].HttpPolicyRefs[0].HppMap[0].Path), gomega.Equal(1))
	g.Expect(nodes[0].SniNodes[0].HttpPolicyRefs[0].HppMap[0].Path[0], gomega.Equal("/foo"))
	g.Expect(nodes[0].SniNodes[0].HttpPolicyRefs[0].Name, gomega.Equal("cluster--default-foo.com_foo-ingress-shp"))
	g.Expect(len(nodes[0].SniNodes[0].SSLKeyCertRefs), gomega.Equal(1))

	ingrFake1, err = (integrationtest.FakeIngress{
		Name:      "ingress-shp",
		Namespace: "default",
		DnsNames:  []string{"foo.com"},
		Ips:       []string{"8.8.8.8"},
		Paths:     []string{"/foo", "/bar", "/baz"},
		HostNames: []string{"v1"},
		TlsSecretDNS: map[string][]string{
			"my-secret": []string{"foo.com"},
		},
		ServiceName: "avisvc",
	}).UpdateIngress()
	if err != nil {
		t.Fatalf("error in updating ingress %s", err)
	}
	integrationtest.PollForCompletion(t, modelName, 5)
	g.Eventually(func() int {
		found, aviModel = objects.SharedAviGraphLister().Get(modelName)
		if found {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
			return len(nodes[0].SniNodes[0].HttpPolicyRefs)
		} else {
			return 0
		}
	}, 30*time.Second).Should(gomega.Equal(3))
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(len(nodes[0].SniNodes[0].HttpPolicyRefs[0].HppMap[0].Path), gomega.Equal(1))
	g.Expect(len(nodes[0].SniNodes[0].HttpPolicyRefs[0].HppMap[0].Path), gomega.Equal(1))
	g.Expect(func() []string {
		p := []string{
			nodes[0].SniNodes[0].HttpPolicyRefs[0].HppMap[0].Path[0],
			nodes[0].SniNodes[0].HttpPolicyRefs[1].HppMap[0].Path[0],
			nodes[0].SniNodes[0].HttpPolicyRefs[2].HppMap[0].Path[0]}
		sort.Strings(p)
		return p
	}, gomega.Equal([]string{"/bar", "/baz", "/foo"}))
	g.Expect(func() []string {
		p := []string{
			nodes[0].SniNodes[0].HttpPolicyRefs[0].Name,
			nodes[0].SniNodes[0].HttpPolicyRefs[1].Name,
			nodes[0].SniNodes[0].HttpPolicyRefs[2].Name}
		sort.Strings(p)
		return p
	}, gomega.Equal([]string{"cluster--default-foo.com_bar-ingress-shp",
		"cluster--default-foo.com_baz-ingress-shp",
		"cluster--default-foo.com_foo-ingress-shp"}))
	g.Expect(len(nodes[0].SniNodes[0].SSLKeyCertRefs), gomega.Equal(1))

	err = KubeClient.ExtensionsV1beta1().Ingresses("default").Delete("ingress-shp", nil)
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	KubeClient.CoreV1().Secrets("default").Delete("my-secret", nil)
	VerifySNIIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngress(t, modelName)
}

func TestFullSyncCacheNoOp(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	integrationtest.AddSecret("my-secret", "default", "tlsCert", "tlsKey")
	modelName := "admin/cluster--Shared-L7-0"
	SetUpTestForIngress(t, modelName)
	//create multipath ingress with tls secret
	ingrFake1 := (integrationtest.FakeIngress{
		Name:      "ingress-fsno",
		Namespace: "default",
		DnsNames:  []string{"foo.com"},
		Ips:       []string{"8.8.8.8"},
		Paths:     []string{"/foo", "/bar"},
		HostNames: []string{"v1"},
		TlsSecretDNS: map[string][]string{
			"my-secret": []string{"foo.com"},
		},
		ServiceName: "avisvc",
	}).IngressMultiPath()
	_, err := KubeClient.ExtensionsV1beta1().Ingresses("default").Create(ingrFake1)
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	integrationtest.PollForCompletion(t, modelName, 5)
	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--foo.com"}

	//store old chksum
	mcache := cache.SharedAviObjCache()
	oldSniCache, _ := mcache.VsCacheMeta.AviCacheGet(sniVSKey)
	oldSniCacheObj, _ := oldSniCache.(*cache.AviVsCache)
	oldChksum := oldSniCacheObj.CloudConfigCksum

	//call fullsync
	ctrl.FullSync()
	ctrl.FullSyncK8s()

	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	integrationtest.PollForCompletion(t, modelName, 5)

	//compare with new chksum
	g.Eventually(func() string {
		mcache := cache.SharedAviObjCache()
		newSniCache, _ := mcache.VsCacheMeta.AviCacheGet(sniVSKey)
		newSniCacheObj, _ := newSniCache.(*cache.AviVsCache)
		if newSniCacheObj != nil {
			return newSniCacheObj.CloudConfigCksum
		}
		return ""
	}, 30*time.Second).Should(gomega.Equal(oldChksum))

	err = KubeClient.ExtensionsV1beta1().Ingresses("default").Delete("ingress-fsno", nil)
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	KubeClient.CoreV1().Secrets("default").Delete("my-secret", nil)
	VerifySNIIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngress(t, modelName)
}

func TestMultiHostIngress(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	modelName := "admin/cluster--Shared-L7-0"
	SetUpTestForIngress(t, modelName)

	ingrFake := (integrationtest.FakeIngress{
		Name:        "ingress-multihost",
		Namespace:   "default",
		DnsNames:    []string{"foo.com", "bar.com"},
		Paths:       []string{"/foo", "/bar"},
		ServiceName: "avisvc",
	}).Ingress()

	_, err := KubeClient.ExtensionsV1beta1().Ingresses("default").Create(ingrFake)
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	integrationtest.PollForCompletion(t, modelName, 10)
	found, aviModel := objects.SharedAviGraphLister().Get(modelName)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(1))
		for _, pool := range nodes[0].PoolRefs {
			if pool.Name == "cluster--foo.com_foo-default-ingress-multihost" {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("foo.com/foo"))
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
			} else {
				t.Fatalf("unexpected pool: %s", pool.Name)
			}
		}
		g.Expect(len(nodes[0].PoolGroupRefs)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].PoolGroupRefs[0].Members)).To(gomega.Equal(1))
		for _, pool := range nodes[0].PoolGroupRefs[0].Members {
			if *pool.PoolRef == "/api/pool?name=cluster--foo.com_foo-default-ingress-multihost" {
				g.Expect(*pool.PriorityLabel).To(gomega.Equal("foo.com/foo"))
			} else {
				t.Fatalf("unexpected pool: %s", *pool.PoolRef)
			}
		}
	} else {
		t.Fatalf("Could not find model: %s", modelName)
	}

	modelName = "admin/cluster--Shared-L7-1"

	integrationtest.PollForCompletion(t, modelName, 5)
	found, aviModel = objects.SharedAviGraphLister().Get(modelName)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(1))
		for _, pool := range nodes[0].PoolRefs {
			if pool.Name == "cluster--bar.com_bar-default-ingress-multihost" {

				g.Expect(pool.PriorityLabel).To(gomega.Equal("bar.com/bar"))
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
			} else {
				t.Fatalf("unexpected pool: %s", pool.Name)
			}
		}
		g.Expect(len(nodes[0].PoolGroupRefs)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].PoolGroupRefs[0].Members)).To(gomega.Equal(1))
		for _, pool := range nodes[0].PoolGroupRefs[0].Members {
			if *pool.PoolRef == "/api/pool?name=cluster--bar.com_bar-default-ingress-multihost" {
				g.Expect(*pool.PriorityLabel).To(gomega.Equal("bar.com/bar"))
			} else {
				t.Fatalf("unexpected pool: %s", *pool.PoolRef)
			}
		}
	} else {
		t.Fatalf("Could not find model: %s", modelName)
	}
	err = KubeClient.ExtensionsV1beta1().Ingresses("default").Delete("ingress-multihost", nil)
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngress(t, modelName)
}

func TestMultiHostSameHostNameIngress(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	modelName := "admin/cluster--Shared-L7-0"
	SetUpTestForIngress(t, modelName)

	ingrFake := (integrationtest.FakeIngress{
		Name:        "ingress-multihost",
		Namespace:   "default",
		DnsNames:    []string{"foo.com", "foo.com"},
		Paths:       []string{"/foo", "/bar"},
		ServiceName: "avisvc",
	}).Ingress()

	_, err := KubeClient.ExtensionsV1beta1().Ingresses("default").Create(ingrFake)
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	integrationtest.PollForCompletion(t, modelName, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(modelName)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(2))
		for _, pool := range nodes[0].PoolRefs {
			if pool.Name == "cluster--foo.com_foo-default-ingress-multihost" {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("foo.com/foo"))
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
			} else if pool.Name == "cluster--foo.com_bar-default-ingress-multihost" {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("foo.com/bar"))
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
			} else {
				t.Fatalf("unexpected pool: %s", pool.Name)
			}
		}
		g.Expect(len(nodes[0].PoolGroupRefs)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].PoolGroupRefs[0].Members)).To(gomega.Equal(2))
		for _, pool := range nodes[0].PoolGroupRefs[0].Members {
			if *pool.PoolRef == "/api/pool?name=cluster--foo.com_foo-default-ingress-multihost" {
				g.Expect(*pool.PriorityLabel).To(gomega.Equal("foo.com/foo"))
			} else if *pool.PoolRef == "/api/pool?name=cluster--foo.com_bar-default-ingress-multihost" {
				g.Expect(*pool.PriorityLabel).To(gomega.Equal("foo.com/bar"))
			} else {
				t.Fatalf("unexpected pool: %s", *pool.PoolRef)
			}
		}
	} else {
		t.Fatalf("Could not find model: %s", modelName)
	}

	err = KubeClient.ExtensionsV1beta1().Ingresses("default").Delete("ingress-multihost", nil)
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngress(t, modelName)
}

func TestEditPathIngress(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	modelName := "admin/cluster--Shared-L7-0"
	SetUpTestForIngress(t, modelName)

	ingrFake := (integrationtest.FakeIngress{
		Name:        "ingress-edit",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/foo"},
		ServiceName: "avisvc",
	}).Ingress()
	ingrFake.ResourceVersion = "1"
	_, err := KubeClient.ExtensionsV1beta1().Ingresses("default").Create(ingrFake)
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	integrationtest.PollForCompletion(t, modelName, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(modelName)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Eventually(len(nodes), 10*time.Second).Should(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Eventually(func() []*avinodes.AviPoolNode {
			return nodes[0].PoolRefs
		}, 10*time.Second).Should(gomega.HaveLen(1))

		g.Expect(nodes[0].PoolRefs[0].Name).To(gomega.Equal("cluster--foo.com_foo-default-ingress-edit"))
		g.Expect(nodes[0].PoolRefs[0].PriorityLabel).To(gomega.Equal("foo.com/foo"))
		g.Expect(len(nodes[0].PoolRefs[0].Servers)).To(gomega.Equal(1))

		g.Expect(len(nodes[0].PoolGroupRefs)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].PoolGroupRefs[0].Members)).To(gomega.Equal(1))

		pool := nodes[0].PoolGroupRefs[0].Members[0]
		g.Expect(*pool.PoolRef).To(gomega.Equal("/api/pool?name=cluster--foo.com_foo-default-ingress-edit"))
		g.Expect(*pool.PriorityLabel).To(gomega.Equal("foo.com/foo"))

	} else {
		t.Fatalf("Could not find model: %s", modelName)
	}

	ingrFake = (integrationtest.FakeIngress{
		Name:        "ingress-edit",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/bar"},
		ServiceName: "avisvc",
	}).Ingress()
	ingrFake.ResourceVersion = "2"
	_, err = KubeClient.ExtensionsV1beta1().Ingresses("default").Update(ingrFake)
	if err != nil {
		t.Fatalf("error in updating Ingress: %v", err)
	}

	found, aviModel = objects.SharedAviGraphLister().Get(modelName)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Eventually(len(nodes), 10*time.Second).Should(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Eventually(func() []*avinodes.AviPoolNode {
			return nodes[0].PoolRefs
		}, 10*time.Second).Should(gomega.HaveLen(1))
		g.Eventually(func() string {
			return nodes[0].PoolRefs[0].Name
		}, 10*time.Second).Should(gomega.Equal("cluster--foo.com_bar-default-ingress-edit"))
		g.Expect(nodes[0].PoolRefs[0].PriorityLabel).To(gomega.Equal("foo.com/bar"))
		g.Expect(len(nodes[0].PoolRefs[0].Servers)).To(gomega.Equal(1))

		pool := nodes[0].PoolGroupRefs[0].Members[0]
		g.Expect(*pool.PoolRef).To(gomega.Equal("/api/pool?name=cluster--foo.com_bar-default-ingress-edit"))
		g.Expect(*pool.PriorityLabel).To(gomega.Equal("foo.com/bar"))
	} else {
		t.Fatalf("Could not find model: %s", modelName)
	}

	err = KubeClient.ExtensionsV1beta1().Ingresses("default").Delete("ingress-edit", nil)
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngress(t, modelName)
}

func TestEditMultiPathIngress(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	modelName := "admin/cluster--Shared-L7-0"
	SetUpTestForIngress(t, modelName)

	ingrFake := (integrationtest.FakeIngress{
		Name:        "ingress-multipath-edit",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/foo"},
		ServiceName: "avisvc",
	}).Ingress()
	ingrFake.ResourceVersion = "1"

	_, err := KubeClient.ExtensionsV1beta1().Ingresses("default").Create(ingrFake)
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	integrationtest.PollForCompletion(t, modelName, 5)
	ingrFake = (integrationtest.FakeIngress{
		Name:        "ingress-multipath-edit",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/foo", "/bar"},
		ServiceName: "avisvc",
	}).IngressMultiPath()
	ingrFake.ResourceVersion = "2"
	_, err = KubeClient.ExtensionsV1beta1().Ingresses("default").Update(ingrFake)
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	found, aviModel := objects.SharedAviGraphLister().Get(modelName)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Eventually(len(nodes), 10*time.Second).Should(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		//g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(2))
		g.Eventually(func() []*avinodes.AviPoolNode {
			return nodes[0].PoolRefs
		}, 10*time.Second).Should(gomega.HaveLen(2))
		for _, pool := range nodes[0].PoolRefs {
			if pool.Name == "cluster--foo.com_foo-default-ingress-multipath-edit" {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("foo.com/foo"))
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
			} else if pool.Name == "cluster--foo.com_bar-default-ingress-multipath-edit" {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("foo.com/bar"))
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
			} else {
				t.Fatalf("unexpected pool: %s", pool.Name)
			}
		}

		g.Expect(len(nodes[0].PoolGroupRefs)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].PoolGroupRefs[0].Members)).To(gomega.Equal(2))
		for _, pool := range nodes[0].PoolGroupRefs[0].Members {
			if *pool.PoolRef == "/api/pool?name=cluster--foo.com_foo-default-ingress-multipath-edit" {
				g.Expect(*pool.PriorityLabel).To(gomega.Equal("foo.com/foo"))
			} else if *pool.PoolRef == "/api/pool?name=cluster--foo.com_bar-default-ingress-multipath-edit" {
				g.Expect(*pool.PriorityLabel).To(gomega.Equal("foo.com/bar"))
			} else {
				t.Fatalf("unexpected pool: %s", *pool.PoolRef)
			}
		}
	} else {
		t.Fatalf("Could not find model: %s", modelName)
	}
	ingrFake = (integrationtest.FakeIngress{
		Name:        "ingress-multipath-edit",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/foo", "/foobar"},
		ServiceName: "avisvc",
	}).IngressMultiPath()
	ingrFake.ResourceVersion = "3"
	objects.SharedAviGraphLister().Delete(modelName)
	_, err = KubeClient.ExtensionsV1beta1().Ingresses("default").Update(ingrFake)
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	integrationtest.PollForCompletion(t, modelName, 5)
	found, aviModel = objects.SharedAviGraphLister().Get(modelName)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Eventually(len(nodes), 10*time.Second).Should(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Eventually(func() []*avinodes.AviPoolNode {
			return nodes[0].PoolRefs
		}, 10*time.Second).Should(gomega.HaveLen(2))
		for _, pool := range nodes[0].PoolRefs {
			if pool.Name == "cluster--foo.com_foo-default-ingress-multipath-edit" {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("foo.com/foo"))
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
			} else if pool.Name == "cluster--foo.com_foobar-default-ingress-multipath-edit" {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("foo.com/foobar"))
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
			} else {
				t.Fatalf("unexpected pool: %s", pool.Name)
			}
		}

		g.Expect(len(nodes[0].PoolGroupRefs)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].PoolGroupRefs[0].Members)).To(gomega.Equal(2))
		for _, pool := range nodes[0].PoolGroupRefs[0].Members {
			if *pool.PoolRef == "/api/pool?name=cluster--foo.com_foo-default-ingress-multipath-edit" {
				g.Expect(*pool.PriorityLabel).To(gomega.Equal("foo.com/foo"))
			} else if *pool.PoolRef == "/api/pool?name=cluster--foo.com_foobar-default-ingress-multipath-edit" {
				g.Expect(*pool.PriorityLabel).To(gomega.Equal("foo.com/foobar"))
			} else {
				t.Fatalf("unexpected pool: %s", *pool.PoolRef)
			}
		}
	} else {
		t.Fatalf("Could not find model: %s", modelName)
	}

	err = KubeClient.ExtensionsV1beta1().Ingresses("default").Delete("ingress-multipath-edit", nil)
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngress(t, modelName)
}

func TestEditMultiIngressSameHost(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	model_name := "admin/cluster--Shared-L7-0"
	SetUpTestForIngress(t, model_name)

	ingrFake1 := (integrationtest.FakeIngress{
		Name:        "ingress-multi1",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/foo"},
		ServiceName: "avisvc",
	}).Ingress()

	_, err := integrationtest.KubeClient.ExtensionsV1beta1().Ingresses("default").Create(ingrFake1)
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	ingrFake2 := (integrationtest.FakeIngress{
		Name:        "ingress-multi2",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/bar"},
		ServiceName: "avisvc",
	}).Ingress()

	_, err = integrationtest.KubeClient.ExtensionsV1beta1().Ingresses("default").Create(ingrFake2)
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	ingrFake2 = (integrationtest.FakeIngress{
		Name:        "ingress-multi2",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/foobar"},
		ServiceName: "avisvc",
	}).Ingress()

	_, err = integrationtest.KubeClient.ExtensionsV1beta1().Ingresses("default").Update(ingrFake2)
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	integrationtest.PollForCompletion(t, model_name, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(model_name)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(2))
		for _, pool := range nodes[0].PoolRefs {
			if pool.Name == "cluster--foo.com_foo-default-ingress-multi1" {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("foo.com/foo"))
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
			} else if pool.Name == "cluster--foo.com_foobar-default-ingress-multi2" {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("foo.com/foobar"))
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
			} else {
				t.Fatalf("unexpected pool: %s", pool.Name)
			}
		}
		g.Expect(len(nodes[0].PoolGroupRefs)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].PoolGroupRefs[0].Members)).To(gomega.Equal(2))
		for _, pool := range nodes[0].PoolGroupRefs[0].Members {
			if *pool.PoolRef == "/api/pool?name=cluster--foo.com_foo-default-ingress-multi1" {
				g.Expect(*pool.PriorityLabel).To(gomega.Equal("foo.com/foo"))
			} else if *pool.PoolRef == "/api/pool?name=cluster--foo.com_foobar-default-ingress-multi2" {
				g.Expect(*pool.PriorityLabel).To(gomega.Equal("foo.com/foobar"))
			} else {
				t.Fatalf("unexpected pool: %s", *pool.PoolRef)
			}
		}
	} else {
		t.Fatalf("Could not find model: %s", model_name)
	}
	err = integrationtest.KubeClient.ExtensionsV1beta1().Ingresses("default").Delete("ingress-multi1", nil)
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 1)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(len(nodes)).To(gomega.Equal(1))
	g.Expect(nodes[0].PoolRefs[0].Name).To(gomega.Equal("cluster--foo.com_foobar-default-ingress-multi2"))

	err = integrationtest.KubeClient.ExtensionsV1beta1().Ingresses("default").Delete("ingress-multi2", nil)
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngress(t, model_name)
}

func TestEditMultiHostIngress(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	modelName := "admin/cluster--Shared-L7-0"
	SetUpTestForIngress(t, modelName)

	ingrFake := (integrationtest.FakeIngress{
		Name:        "ingress-multihost",
		Namespace:   "default",
		DnsNames:    []string{"foo.com", "bar.com"},
		Paths:       []string{"/foo", "/bar"},
		ServiceName: "avisvc",
	}).Ingress()

	_, err := KubeClient.ExtensionsV1beta1().Ingresses("default").Create(ingrFake)
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	ingrFake = (integrationtest.FakeIngress{
		Name:        "ingress-multihost",
		Namespace:   "default",
		DnsNames:    []string{"foo.com", "foobar.com"},
		Paths:       []string{"/foo", "/bar"},
		ServiceName: "avisvc",
	}).Ingress()

	_, err = KubeClient.ExtensionsV1beta1().Ingresses("default").Update(ingrFake)

	integrationtest.PollForCompletion(t, modelName, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(modelName)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(1))
		for _, pool := range nodes[0].PoolRefs {
			if pool.Name == "cluster--foo.com_foo-default-ingress-multihost" {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("foo.com/foo"))
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
			} else {
				t.Fatalf("unexpected pool: %s", pool.Name)
			}
		}

		g.Expect(len(nodes[0].PoolGroupRefs)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].PoolGroupRefs[0].Members)).To(gomega.Equal(1))
		for _, pool := range nodes[0].PoolGroupRefs[0].Members {
			if *pool.PoolRef == "/api/pool?name=cluster--foo.com_foo-default-ingress-multihost" {
				g.Expect(*pool.PriorityLabel).To(gomega.Equal("foo.com/foo"))
			} else {
				t.Fatalf("unexpected pool: %s", *pool.PoolRef)
			}
		}
	} else {
		t.Fatalf("Could not find model: %s", modelName)
	}

	modelName = "admin/cluster--Shared-L7-3"
	found, aviModel = objects.SharedAviGraphLister().Get(modelName)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(1))
		for _, pool := range nodes[0].PoolRefs {
			if pool.Name == "cluster--foobar.com_bar-default-ingress-multihost" {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("foobar.com/bar"))
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
			} else {
				t.Fatalf("unexpected pool: %s", pool.Name)
			}
		}

		g.Expect(len(nodes[0].PoolGroupRefs)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].PoolGroupRefs[0].Members)).To(gomega.Equal(1))
		for _, pool := range nodes[0].PoolGroupRefs[0].Members {
			if *pool.PoolRef == "/api/pool?name=cluster--foobar.com_bar-default-ingress-multihost" {
				g.Expect(*pool.PriorityLabel).To(gomega.Equal("foobar.com/bar"))
			} else {
				t.Fatalf("unexpected pool: %s", *pool.PoolRef)
			}
		}
	} else {
		t.Fatalf("Could not find model: %s", modelName)
	}

	err = KubeClient.ExtensionsV1beta1().Ingresses("default").Delete("ingress-multihost", nil)
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngress(t, modelName)
}

func TestNoHostIngress(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	modelName := "admin/cluster--Shared-L7-2"
	SetUpTestForIngress(t, modelName)

	ingrFake := (integrationtest.FakeIngress{
		Name:        "ingress-nohost",
		Namespace:   "default",
		Paths:       []string{"/foo"},
		ServiceName: "avisvc",
	}).IngressNoHost()

	_, err := KubeClient.ExtensionsV1beta1().Ingresses("default").Create(ingrFake)
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	integrationtest.PollForCompletion(t, modelName, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(modelName)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(1))
		g.Expect(nodes[0].PoolRefs[0].Name).To(gomega.Equal("cluster--ingress-nohost.default.com_foo-default-ingress-nohost"))
		g.Expect(nodes[0].PoolRefs[0].PriorityLabel).To(gomega.Equal("ingress-nohost.default.com/foo"))

		g.Expect(len(nodes[0].PoolGroupRefs)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].PoolGroupRefs[0].Members)).To(gomega.Equal(1))

		pool := nodes[0].PoolGroupRefs[0].Members[0]
		g.Expect(*pool.PoolRef).To(gomega.Equal("/api/pool?name=cluster--ingress-nohost.default.com_foo-default-ingress-nohost"))
		g.Expect(*pool.PriorityLabel).To(gomega.Equal("ingress-nohost.default.com/foo"))
	} else {
		t.Fatalf("Could not find model: %s", modelName)
	}

	err = KubeClient.ExtensionsV1beta1().Ingresses("default").Delete("ingress-nohost", nil)
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngress(t, modelName)
}

func TestEditNoHostIngress(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	modelName := "admin/cluster--Shared-L7-2"
	SetUpTestForIngress(t, modelName)

	ingrFake := (integrationtest.FakeIngress{
		Name:        "ingress-nohost",
		Namespace:   "default",
		Paths:       []string{"/foo"},
		ServiceName: "avisvc",
	}).IngressNoHost()
	_, err := KubeClient.ExtensionsV1beta1().Ingresses("default").Create(ingrFake)
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	ingrFake = (integrationtest.FakeIngress{
		Name:        "ingress-nohost",
		Namespace:   "default",
		Paths:       []string{"/bar"},
		ServiceName: "avisvc",
	}).IngressNoHost()
	ingrFake.ResourceVersion = "2"
	_, err = KubeClient.ExtensionsV1beta1().Ingresses("default").Update(ingrFake)
	if err != nil {
		t.Fatalf("error in Updating Ingress: %v", err)
	}

	integrationtest.PollForCompletion(t, modelName, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(modelName)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(1))
		g.Eventually(func() string {
			return nodes[0].PoolRefs[0].Name
		}, 10*time.Second).Should(gomega.Equal("cluster--ingress-nohost.default.com_bar-default-ingress-nohost"))
		g.Expect(nodes[0].PoolRefs[0].PriorityLabel).To(gomega.Equal("ingress-nohost.default.com/bar"))

		g.Expect(len(nodes[0].PoolGroupRefs)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].PoolGroupRefs[0].Members)).To(gomega.Equal(1))

		pool := nodes[0].PoolGroupRefs[0].Members[0]
		g.Expect(*pool.PoolRef).To(gomega.Equal("/api/pool?name=cluster--ingress-nohost.default.com_bar-default-ingress-nohost"))
		g.Expect(*pool.PriorityLabel).To(gomega.Equal("ingress-nohost.default.com/bar"))
	} else {
		t.Fatalf("Could not find model: %s", modelName)
	}

	err = KubeClient.ExtensionsV1beta1().Ingresses("default").Delete("ingress-nohost", nil)
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngress(t, modelName)
}

func TestEditNoHostToHostIngress(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	modelName := "admin/cluster--Shared-L7-2"
	SetUpTestForIngress(t, modelName)

	ingrFake := (integrationtest.FakeIngress{
		Name:        "ingress-nohost",
		Namespace:   "default",
		Paths:       []string{"/foo"},
		ServiceName: "avisvc",
	}).IngressNoHost()

	_, err := KubeClient.ExtensionsV1beta1().Ingresses("default").Create(ingrFake)
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	integrationtest.PollForCompletion(t, modelName, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(modelName)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(1))
		g.Expect(nodes[0].PoolRefs[0].Name).To(gomega.Equal("cluster--ingress-nohost.default.com_foo-default-ingress-nohost"))
		g.Expect(nodes[0].PoolRefs[0].PriorityLabel).To(gomega.Equal("ingress-nohost.default.com/foo"))

		g.Expect(len(nodes[0].PoolGroupRefs)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].PoolGroupRefs[0].Members)).To(gomega.Equal(1))

		pool := nodes[0].PoolGroupRefs[0].Members[0]
		g.Expect(*pool.PoolRef).To(gomega.Equal("/api/pool?name=cluster--ingress-nohost.default.com_foo-default-ingress-nohost"))
		g.Expect(*pool.PriorityLabel).To(gomega.Equal("ingress-nohost.default.com/foo"))
	} else {
		t.Fatalf("Could not find model: %s", modelName)
	}

	ingrFake = (integrationtest.FakeIngress{
		Name:        "ingress-nohost",
		Namespace:   "default",
		DnsNames:    []string{"bar.com"},
		Paths:       []string{"/bar"},
		ServiceName: "avisvc",
	}).Ingress()

	ingrFake.ResourceVersion = "2"
	_, err = KubeClient.ExtensionsV1beta1().Ingresses("default").Update(ingrFake)
	if err != nil {
		t.Fatalf("error in Updating Ingress: %v", err)
	}
	found, aviModel = objects.SharedAviGraphLister().Get(modelName)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Eventually(func() int {
			return len(nodes[0].PoolRefs)
		}, 10*time.Second).Should(gomega.Equal(0))

	} else {
		t.Fatalf("Could not find model: %s", modelName)
	}
	VerifyIngressDeletion(t, g, aviModel, 0)
	modelName = "admin/cluster--Shared-L7-1"
	found, aviModel = objects.SharedAviGraphLister().Get(modelName)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Eventually(func() int {
			return len(nodes[0].PoolRefs)
		}, 10*time.Second).Should(gomega.Equal(1))
		g.Expect(nodes[0].PoolRefs[0].PriorityLabel).To(gomega.Equal("bar.com/bar"))
	} else {
		t.Fatalf("Could not find model: %s", modelName)
	}
	err = KubeClient.ExtensionsV1beta1().Ingresses("default").Delete("ingress-nohost", nil)
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngress(t, modelName)
}

func TestEditNoHostMultiPathIngress(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	modelName := "admin/cluster--Shared-L7-3"
	SetUpTestForIngress(t, modelName)

	ingrFake := (integrationtest.FakeIngress{
		Name:        "nohost-multipath",
		Namespace:   "default",
		Paths:       []string{"/foo", "/bar"},
		ServiceName: "avisvc",
	}).IngressNoHost()

	_, err := KubeClient.ExtensionsV1beta1().Ingresses("default").Create(ingrFake)
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	ingrFake = (integrationtest.FakeIngress{
		Name:        "nohost-multipath",
		Namespace:   "default",
		Paths:       []string{"/foo", "/foobar"},
		ServiceName: "avisvc",
	}).IngressNoHost()
	ingrFake.ResourceVersion = "2"
	_, err = KubeClient.ExtensionsV1beta1().Ingresses("default").Update(ingrFake)
	if err != nil {
		t.Fatalf("error in updating Ingress: %v", err)
	}

	integrationtest.PollForCompletion(t, modelName, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(modelName)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))

		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(2))
		for _, pool := range nodes[0].PoolRefs {
			if pool.Name == "cluster--nohost-multipath.default.com_foo-default-nohost-multipath" {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("nohost-multipath.default.com/foo"))
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
			} else if pool.Name == "cluster--nohost-multipath.default.com_foobar-default-nohost-multipath" {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("nohost-multipath.default.com/foobar"))
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
			} else {
				t.Fatalf("unexpected pool: %s", pool.Name)
			}
		}

		g.Expect(len(nodes[0].PoolGroupRefs)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].PoolGroupRefs[0].Members)).To(gomega.Equal(2))
		for _, pool := range nodes[0].PoolGroupRefs[0].Members {
			if *pool.PoolRef == "/api/pool?name=cluster--nohost-multipath.default.com_foo-default-nohost-multipath" {
				g.Expect(*pool.PriorityLabel).To(gomega.Equal("nohost-multipath.default.com/foo"))
			} else if *pool.PoolRef == "/api/pool?name=cluster--nohost-multipath.default.com_foobar-default-nohost-multipath" {
				g.Expect(*pool.PriorityLabel).To(gomega.Equal("nohost-multipath.default.com/foobar"))
			} else {
				t.Fatalf("unexpected pool: %s", *pool.PoolRef)
			}
		}

	} else {
		t.Fatalf("Could not find model: %s", modelName)
	}

	err = KubeClient.ExtensionsV1beta1().Ingresses("default").Delete("nohost-multipath", nil)
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngress(t, modelName)
}

func TestScaleEndpoints(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName := "admin/cluster--Shared-L7-0"
	SetUpTestForIngress(t, modelName)

	ingrFake1 := (integrationtest.FakeIngress{
		Name:        "ingress-multi1",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/foo"},
		ServiceName: "avisvc",
	}).Ingress()

	_, err := KubeClient.ExtensionsV1beta1().Ingresses("default").Create(ingrFake1)
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	ingrFake2 := (integrationtest.FakeIngress{
		Name:        "ingress-multi2",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/bar"},
		ServiceName: "avisvc",
	}).Ingress()

	_, err = KubeClient.ExtensionsV1beta1().Ingresses("default").Create(ingrFake2)
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	integrationtest.PollForCompletion(t, modelName, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(modelName)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(2))
		for _, pool := range nodes[0].PoolRefs {
			if pool.Name == "cluster--foo.com_foo-default-ingress-multi1" {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("foo.com/foo"))
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
			} else if pool.Name == "cluster--foo.com_bar-default-ingress-multi2" {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("foo.com/bar"))
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
			} else {
				t.Fatalf("unexpected pool: %s", pool.Name)
			}
		}
		g.Expect(len(nodes[0].PoolGroupRefs)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].PoolGroupRefs[0].Members)).To(gomega.Equal(2))
		for _, pool := range nodes[0].PoolGroupRefs[0].Members {
			if *pool.PoolRef == "/api/pool?name=cluster--foo.com_foo-default-ingress-multi1" {
				g.Expect(*pool.PriorityLabel).To(gomega.Equal("foo.com/foo"))
			} else if *pool.PoolRef == "/api/pool?name=cluster--foo.com_bar-default-ingress-multi2" {
				g.Expect(*pool.PriorityLabel).To(gomega.Equal("foo.com/bar"))
			} else {
				t.Fatalf("unexpected pool: %s", *pool.PoolRef)
			}
		}
	} else {
		t.Fatalf("Could not find model: %s", modelName)
	}
	integrationtest.ScaleCreateEP(t, "default", "avisvc")
	found, aviModel = objects.SharedAviGraphLister().Get(modelName)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(2))

		g.Eventually(func() int {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
			return len(nodes[0].PoolRefs[0].Servers)
		}, 10*time.Second).Should(gomega.Equal(2))

		g.Eventually(func() int {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
			return len(nodes[0].PoolRefs[1].Servers)
		}, 10*time.Second).Should(gomega.Equal(2))

		g.Expect(len(nodes[0].PoolGroupRefs)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].PoolGroupRefs[0].Members)).To(gomega.Equal(2))
	} else {
		t.Fatalf("Could not find model: %s", modelName)
	}
	err = KubeClient.ExtensionsV1beta1().Ingresses("default").Delete("ingress-multi1", nil)
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 1)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(len(nodes)).To(gomega.Equal(1))
	g.Expect(nodes[0].PoolRefs[0].Name).To(gomega.Equal("cluster--foo.com_bar-default-ingress-multi2"))

	err = KubeClient.ExtensionsV1beta1().Ingresses("default").Delete("ingress-multi2", nil)
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 0)
	TearDownTestForIngress(t, modelName)

}

// All SNI test cases follow:

func TestL7ModelSNI(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	integrationtest.AddSecret("my-secret", "default", "tlsCert", "tlsKey")
	modelName := "admin/cluster--Shared-L7-0"
	SetUpTestForIngress(t, modelName)

	integrationtest.PollForCompletion(t, modelName, 5)

	// foo.com and noo.com compute the same hashed shard vs num
	ingrFake := (integrationtest.FakeIngress{
		Name:      "foo-with-targets",
		Namespace: "default",
		DnsNames:  []string{"foo.com", "noo.com"},
		Ips:       []string{"8.8.8.8"},
		HostNames: []string{"v1"},
		TlsSecretDNS: map[string][]string{
			"my-secret": []string{"foo.com"},
		},
		ServiceName: "avisvc",
	}).Ingress()

	_, err := KubeClient.ExtensionsV1beta1().Ingresses("default").Create(ingrFake)
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	integrationtest.PollForCompletion(t, modelName, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(modelName)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Expect(nodes[0].PoolRefs).To(gomega.HaveLen(1))
		g.Expect(nodes[0].PoolRefs[0].Name).To(gomega.ContainSubstring("noo.com"))
		g.Expect(nodes[0].HttpPolicyRefs).To(gomega.HaveLen(1)) // redirect http->https policy
		g.Expect(nodes[0].HttpPolicyRefs[0].RedirectPorts[0].Hosts[0]).To(gomega.Equal("foo.com"))

		g.Expect(nodes[0].SniNodes[0].VHDomainNames[0]).To(gomega.Equal("foo.com"))
		g.Expect(len(nodes[0].SniNodes)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].SniNodes[0].PoolGroupRefs)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].SniNodes[0].HttpPolicyRefs)).To(gomega.Equal(1))
		g.Expect(nodes[0].SniNodes[0].PoolRefs[0].Name).To(gomega.ContainSubstring("foo.com"))
		g.Expect(len(nodes[0].SniNodes[0].PoolRefs)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].SniNodes[0].PoolRefs[0].Servers)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].SniNodes[0].SSLKeyCertRefs)).To(gomega.Equal(1))
	} else {
		t.Fatalf("Could not find Model: %v", err)
	}
	err = KubeClient.ExtensionsV1beta1().Ingresses("default").Delete("foo-with-targets", nil)
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	KubeClient.CoreV1().Secrets("default").Delete("my-secret", nil)
	VerifySNIIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngress(t, modelName)
}

func TestL7ModelNoSecretToSecret(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	modelName := "admin/cluster--Shared-L7-0"
	SetUpTestForIngress(t, modelName)

	integrationtest.PollForCompletion(t, modelName, 5)
	found, _ := objects.SharedAviGraphLister().Get(modelName)
	if found {
		// We shouldn't get an update for this update since it neither belongs to an ingress nor a L4 LB service
		t.Fatalf("Couldn't find Model for DELETE event %v", modelName)
	}
	ingrFake := (integrationtest.FakeIngress{
		Name:        "foo-no-secret",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		ServiceName: "avisvc",
		TlsSecretDNS: map[string][]string{
			"my-secret": []string{"foo.com"},
		},
	}).Ingress()
	_, err := KubeClient.ExtensionsV1beta1().Ingresses("default").Create(ingrFake)
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	integrationtest.PollForCompletion(t, modelName, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(modelName)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Expect(len(nodes[0].SniNodes)).To(gomega.Equal(0))
		g.Expect(nodes[0].VHDomainNames).To(gomega.HaveLen(0))
		g.Expect(nodes[0].HttpPolicyRefs).To(gomega.HaveLen(0))
	} else {
		t.Fatalf("Could not find Model: %v", err)
	}

	// Now create the secret and verify the models.
	integrationtest.AddSecret("my-secret", "default", "tlsCert", "tlsKey")
	found, aviModel = objects.SharedAviGraphLister().Get(modelName)
	if found {
		g.Eventually(func() int {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
			return len(nodes[0].SniNodes)
		}, 10*time.Second).Should(gomega.Equal(1))

	} else {
		t.Fatalf("Could not find model: %s", modelName)
	}

	err = KubeClient.ExtensionsV1beta1().Ingresses("default").Delete("foo-no-secret", nil)
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	KubeClient.CoreV1().Secrets("default").Delete("my-secret", nil)
	VerifySNIIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngress(t, modelName)
}

func TestL7ModelOneSecretToMultiIng(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	modelName := "admin/cluster--Shared-L7-0"
	SetUpTestForIngress(t, modelName)

	integrationtest.PollForCompletion(t, modelName, 5)
	found, _ := objects.SharedAviGraphLister().Get(modelName)
	if found {
		// We shouldn't get an update for this update since it neither belongs to an ingress nor a L4 LB service
		t.Fatalf("Couldn't find Model for DELETE event %v", modelName)
	}
	ingrFake1 := (integrationtest.FakeIngress{
		Name:        "foo-no-secret1",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		ServiceName: "avisvc",
		TlsSecretDNS: map[string][]string{
			"my-secret": []string{"foo.com"},
		},
	}).Ingress()
	_, err := KubeClient.ExtensionsV1beta1().Ingresses("default").Create(ingrFake1)
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	ingrFake2 := (integrationtest.FakeIngress{
		Name:      "foo-no-secret2",
		Namespace: "default",
		DnsNames:  []string{"foo.com"},
		Ips:       []string{"8.8.8.8"},
		HostNames: []string{"v1"},
		TlsSecretDNS: map[string][]string{
			"my-secret": []string{"foo.com"},
		},
		ServiceName: "avisvc",
	}).Ingress()
	_, err = KubeClient.ExtensionsV1beta1().Ingresses("default").Create(ingrFake2)
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	integrationtest.PollForCompletion(t, modelName, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(modelName)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Expect(len(nodes[0].SniNodes)).To(gomega.Equal(0))
	} else {
		t.Fatalf("Could not find Model: %v", err)
	}

	// Now create the secret and verify the models.
	integrationtest.AddSecret("my-secret", "default", "tlsCert", "tlsKey")
	found, aviModel = objects.SharedAviGraphLister().Get(modelName)
	if found {
		// Check if the secret affected both the models.
		g.Eventually(func() int {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
			return len(nodes[0].SniNodes)
		}, 10*time.Second).Should(gomega.Equal(1))
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(nodes[0].SniNodes[0].VHDomainNames[0]).To(gomega.Equal("foo.com"))
	} else {
		t.Fatalf("Could not find model: %s", modelName)
	}
	KubeClient.CoreV1().Secrets("default").Delete("my-secret", nil)
	VerifySNIIngressDeletion(t, g, aviModel, 0)
	// Since we deleted the secret, both SNIs should get removed.
	found, aviModel = objects.SharedAviGraphLister().Get(modelName)
	if found {
		// Check if the secret affected both the models.
		g.Eventually(func() int {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
			return len(nodes[0].SniNodes)
		}, 10*time.Second).Should(gomega.Equal(0))

	} else {
		t.Fatalf("Could not find model: %s", modelName)
	}
	err = KubeClient.ExtensionsV1beta1().Ingresses("default").Delete("foo-no-secret1", nil)
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	err = KubeClient.ExtensionsV1beta1().Ingresses("default").Delete("foo-no-secret2", nil)
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	TearDownTestForIngress(t, modelName)
}

func TestL7ModelMultiSNI(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	integrationtest.AddSecret("my-secret", "default", "tlsCert", "tlsKey")
	modelName := "admin/cluster--Shared-L7-0"
	SetUpTestForIngress(t, modelName)

	ingrFake := (integrationtest.FakeIngress{
		Name:        "foo-with-targets",
		Namespace:   "default",
		DnsNames:    []string{"foo.com", "bar.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		ServiceName: "avisvc",
		TlsSecretDNS: map[string][]string{
			"my-secret": []string{"foo.com", "bar.com"},
		},
	}).Ingress()
	_, err := KubeClient.ExtensionsV1beta1().Ingresses("default").Create(ingrFake)
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	integrationtest.PollForCompletion(t, modelName, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(modelName)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Expect(nodes[0].HttpPolicyRefs).To(gomega.HaveLen(1))
		g.Expect(len(nodes[0].SniNodes)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].SniNodes[0].PoolGroupRefs)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].SniNodes[0].HttpPolicyRefs)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].SniNodes[0].PoolRefs)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].SniNodes[0].PoolRefs[0].Servers)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].SniNodes[0].SSLKeyCertRefs)).To(gomega.Equal(1))
		g.Expect(nodes[0].SniNodes[0].VHDomainNames).To(gomega.HaveLen(1))
	} else {
		t.Fatalf("Could not find Model: %v", err)
	}

	err = KubeClient.ExtensionsV1beta1().Ingresses("default").Delete("foo-with-targets", nil)
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	KubeClient.CoreV1().Secrets("default").Delete("my-secret", nil)
	VerifySNIIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngress(t, modelName)
}

func TestL7ModelMultiSNIMultiCreateEditSecret(t *testing.T) {
	// This test covers creating multiple SNI nodes via multiple secrets.
	g := gomega.NewGomegaWithT(t)
	integrationtest.AddSecret("my-secret", "default", "tlsCert", "tlsKey")
	integrationtest.AddSecret("my-secret2", "default", "tlsCert", "tlsKey")
	// Clean up any earlier models.
	modelName := "admin/cluster--Shared-L7-1"
	objects.SharedAviGraphLister().Delete(modelName)
	modelName = "admin/cluster--Shared-L7-0"
	objects.SharedAviGraphLister().Delete(modelName)
	SetUpTestForIngress(t, modelName)

	ingrFake := (integrationtest.FakeIngress{
		Name:        "foo-with-targets",
		Namespace:   "default",
		DnsNames:    []string{"foo.com", "FOO.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		ServiceName: "avisvc",
		TlsSecretDNS: map[string][]string{
			"my-secret":  []string{"foo.com"},
			"my-secret2": []string{"FOO.com"},
		},
	}).Ingress()

	_, err := KubeClient.ExtensionsV1beta1().Ingresses("default").Create(ingrFake)

	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	integrationtest.PollForCompletion(t, modelName, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(modelName)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Expect(nodes[0].HttpPolicyRefs).To(gomega.HaveLen(1))
		g.Expect(nodes[0].HttpPolicyRefs[0].RedirectPorts[0].Hosts).To(gomega.HaveLen(2))
		g.Expect(len(nodes[0].SniNodes)).To(gomega.Equal(2))
		g.Expect(len(nodes[0].SniNodes[0].PoolGroupRefs)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].SniNodes[0].HttpPolicyRefs)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].SniNodes[0].PoolRefs)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].SniNodes[0].PoolRefs[0].Servers)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].SniNodes[0].SSLKeyCertRefs)).To(gomega.Equal(1))
		g.Expect(nodes[0].SniNodes[0].VHDomainNames).To(gomega.HaveLen(1))
	} else {
		t.Fatalf("Could not find Model: %v", err)
	}

	ingrFake = (integrationtest.FakeIngress{
		Name:        "foo-with-targets",
		Namespace:   "default",
		DnsNames:    []string{"foo.com", "bar.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		ServiceName: "avisvc",
		TlsSecretDNS: map[string][]string{
			"my-secret":  []string{"foo.com"},
			"my-secret2": []string{"bar.com"},
		},
	}).Ingress()
	ingrFake.ResourceVersion = "2"
	_, err = KubeClient.ExtensionsV1beta1().Ingresses("default").Update(ingrFake)

	// Because of change of the hostnames, the SNI nodes should now get distributed to two shared VSes.
	found, aviModel = objects.SharedAviGraphLister().Get(modelName)
	if found {
		// Check if the secret affected both the models.
		g.Eventually(func() int {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
			return len(nodes[0].SniNodes)
		}, 10*time.Second).Should(gomega.Equal(1))
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(nodes[0].SniNodes[0].VHDomainNames).To(gomega.HaveLen(1))
	} else {
		t.Fatalf("Could not find model: %s", modelName)
	}
	modelName = "admin/cluster--Shared-L7-1"
	found, aviModel = objects.SharedAviGraphLister().Get(modelName)
	if found {
		g.Eventually(func() int {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
			return len(nodes[0].SniNodes)
		}, 10*time.Second).Should(gomega.Equal(1))
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(nodes[0].SniNodes[0].VHDomainNames).To(gomega.HaveLen(1))
	} else {
		t.Fatalf("Could not find model: %s", modelName)
	}
	err = KubeClient.ExtensionsV1beta1().Ingresses("default").Delete("foo-with-targets", nil)
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}

	KubeClient.CoreV1().Secrets("default").Delete("my-secret", nil)
	KubeClient.CoreV1().Secrets("default").Delete("my-secret2", nil)
	VerifySNIIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngress(t, modelName)
}

func TestL7WrongSubDomainMultiSNI(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	integrationtest.AddSecret("my-secret", "default", "tlsCert", "tlsKey")
	integrationtest.AddSecret("my-secret2", "default", "tlsCert", "tlsKey")
	modelName := "admin/cluster--Shared-L7-1"
	SetUpTestForIngress(t, modelName)

	ingrFake := (integrationtest.FakeIngress{
		Name:        "foo-with-targets",
		Namespace:   "default",
		DnsNames:    []string{"foo.org"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		ServiceName: "avisvc",
		TlsSecretDNS: map[string][]string{
			"my-secret": []string{"foo.org"},
		},
	}).Ingress()
	_, err := KubeClient.ExtensionsV1beta1().Ingresses("default").Create(ingrFake)
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	integrationtest.PollForCompletion(t, modelName, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(modelName)
	if found {
		// This will not generate a model.
		t.Fatalf("Could not find Model: %v", err)
	}
	ingrFake = (integrationtest.FakeIngress{
		Name:        "foo-with-targets",
		Namespace:   "default",
		DnsNames:    []string{"foo.org", "bar.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		ServiceName: "avisvc",
		TlsSecretDNS: map[string][]string{
			"my-secret":  []string{"foo.org"},
			"my-secret2": []string{"bar.com"},
		},
	}).Ingress()
	ingrFake.ResourceVersion = "2"
	_, err = KubeClient.ExtensionsV1beta1().Ingresses("default").Update(ingrFake)
	integrationtest.PollForCompletion(t, modelName, 5)
	found, aviModel = objects.SharedAviGraphLister().Get(modelName)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes)).To(gomega.Equal(1))
		g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
		g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
		g.Expect(len(nodes[0].SniNodes)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].SniNodes[0].PoolGroupRefs)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].SniNodes[0].HttpPolicyRefs)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].SniNodes[0].PoolRefs)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].SniNodes[0].PoolRefs[0].Servers)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].SniNodes[0].SSLKeyCertRefs)).To(gomega.Equal(1))
		g.Expect(nodes[0].SniNodes[0].VHDomainNames).To(gomega.HaveLen(1))
	} else {
		t.Fatalf("Could not find Model: %v", err)
	}
	err = KubeClient.ExtensionsV1beta1().Ingresses("default").Delete("foo-with-targets", nil)
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	KubeClient.CoreV1().Secrets("default").Delete("my-secret", nil)
	VerifySNIIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngress(t, modelName)
}
