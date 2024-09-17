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

package ingresstests

import (
	"context"
	"net/http"
	"os"
	"sort"
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
	v1beta1crdfake "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/client/v1beta1/clientset/versioned/fake"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/integrationtest"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/api"
	utils "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	"github.com/onsi/gomega"
	"github.com/vmware/alb-sdk/go/models"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

var (
	KubeClient           *k8sfake.Clientset
	CRDClient            *crdfake.Clientset
	v1beta1CRDClient     *v1beta1crdfake.Clientset
	ctrl                 *k8s.AviController
	akoApiServer         *api.FakeApiServer
	keyChan              chan string
	endpointSliceEnabled bool
	objNameMap           integrationtest.ObjectNameMap
)

const (
	MODEL_NAME_PREFIX = "admin/cluster--Shared-L7-"
)

func TestMain(m *testing.M) {
	os.Setenv("INGRESS_API", "extensionv1")
	os.Setenv("VIP_NETWORK_LIST", `[{"networkName":"net123"}]`)
	os.Setenv("CLUSTER_NAME", "cluster")
	os.Setenv("CLOUD_NAME", "CLOUD_VCENTER")
	os.Setenv("SEG_NAME", "Default-Group")
	os.Setenv("NODE_NETWORK_LIST", `[{"networkName":"net123","cidrs":["10.79.168.0/22"]}]`)
	os.Setenv("POD_NAMESPACE", utils.AKO_DEFAULT_NS)
	os.Setenv("SHARD_VS_SIZE", "LARGE")
	os.Setenv("AUTO_L4_FQDN", "default")
	os.Setenv("POD_NAME", "ako-0")

	akoControlConfig := lib.AKOControlConfig()
	endpointSliceEnabled = lib.GetEndpointSliceEnabled()
	akoControlConfig.SetEndpointSlicesEnabled(endpointSliceEnabled)

	KubeClient = k8sfake.NewSimpleClientset()
	CRDClient = crdfake.NewSimpleClientset()
	v1beta1CRDClient = v1beta1crdfake.NewSimpleClientset()
	akoControlConfig.SetCRDClientset(CRDClient)
	akoControlConfig.Setv1beta1CRDClientset(v1beta1CRDClient)
	akoControlConfig.SetAKOInstanceFlag(true)
	akoControlConfig.SetEventRecorder(lib.AKOEventComponent, KubeClient, true)
	akoControlConfig.SetDefaultLBController(true)
	data := map[string][]byte{
		"username": []byte("admin"),
		"password": []byte("admin"),
	}
	object := metav1.ObjectMeta{Name: "avi-secret", Namespace: utils.GetAKONamespace()}
	secret := &corev1.Secret{Data: data, ObjectMeta: object}
	KubeClient.CoreV1().Secrets(utils.GetAKONamespace()).Create(context.TODO(), secret, metav1.CreateOptions{})

	registeredInformers := []string{
		utils.ServiceInformer,
		utils.IngressInformer,
		utils.IngressClassInformer,
		utils.SecretInformer,
		utils.NSInformer,
		utils.NodeInformer,
		utils.ConfigMapInformer,
	}
	if akoControlConfig.GetEndpointSlicesEnabled() {
		registeredInformers = append(registeredInformers, utils.EndpointSlicesInformer)
	} else {
		registeredInformers = append(registeredInformers, utils.EndpointInformer)
	}
	utils.NewInformers(utils.KubeClientIntf{ClientSet: KubeClient}, registeredInformers)
	informers := k8s.K8sinformers{Cs: KubeClient}
	k8s.NewCRDInformers()

	mcache := cache.SharedAviObjCache()
	cloudObj := &cache.AviCloudPropertyCache{Name: "Default-Cloud", VType: "mock"}
	subdomains := []string{"avi.internal", ".com"}
	cloudObj.NSIpamDNS = subdomains
	mcache.CloudKeyCache.AviCacheAdd("Default-Cloud", cloudObj)

	akoApiServer = integrationtest.InitializeFakeAKOAPIServer()

	integrationtest.NewAviFakeClientInstance(KubeClient)
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
	wgSlowRetry := &sync.WaitGroup{}
	waitGroupMap["slowretry"] = wgSlowRetry
	wgGraph := &sync.WaitGroup{}
	waitGroupMap["graph"] = wgGraph
	wgStatus := &sync.WaitGroup{}
	waitGroupMap["status"] = wgStatus
	wgLeaderElection := &sync.WaitGroup{}
	waitGroupMap["leaderElection"] = wgLeaderElection

	integrationtest.AddConfigMap(KubeClient)
	integrationtest.PollForSyncStart(ctrl, 10)
	keyChan = make(chan string)
	ctrl.HandleConfigMap(informers, ctrlCh, stopCh, quickSyncCh)
	integrationtest.KubeClient = KubeClient
	integrationtest.AddDefaultIngressClass()
	ctrl.SetSEGroupCloudNameFromNSAnnotations()
	integrationtest.AddDefaultNamespace()
	integrationtest.AddDefaultNamespace("red")

	go ctrl.InitController(informers, registeredInformers, ctrlCh, stopCh, quickSyncCh, waitGroupMap)
	objNameMap.InitMap()
	os.Exit(m.Run())
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

	modelName := MODEL_NAME_PREFIX + "0"
	svcName := objNameMap.GenerateName("avisvc")
	ingName := objNameMap.GenerateName("foo-with-targets")
	SetUpTestForIngress(t, svcName, modelName)

	integrationtest.PollForCompletion(t, modelName, 5)
	found, _ := objects.SharedAviGraphLister().Get(modelName)
	if found {
		// We shouldn't get an update for this update since it neither belongs to an ingress nor a L4 LB service
		t.Fatalf("Couldn't find Model for DELETE event %v", modelName)
	}
	ingrFake := (integrationtest.FakeIngress{
		Name:        ingName,
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		ServiceName: svcName,
	}).Ingress()

	_, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	integrationtest.PollForCompletion(t, modelName, 5)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 150*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(len(nodes)).To(gomega.Equal(1))
	g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
	g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))

	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), ingName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngress(t, svcName, modelName)
}

func TestShardNamingConvention(t *testing.T) {
	// checks naming convention of all generated nodes
	g := gomega.NewGomegaWithT(t)

	modelName := MODEL_NAME_PREFIX + "0"
	svcName := objNameMap.GenerateName("avisvc")
	secretName := objNameMap.GenerateName("my-secret")
	ingName := objNameMap.GenerateName("foo-with-targets")
	SetUpTestForIngress(t, svcName, modelName)
	integrationtest.AddSecret(secretName, "default", "tlsCert", "tlsKey")

	// foo.com and noo.com compute the same hashed shard vs num
	ingrFake := (integrationtest.FakeIngress{
		Name:      ingName,
		Namespace: "default",
		DnsNames:  []string{"foo.com", "noo.com"},
		Ips:       []string{"8.8.8.8"},
		Paths:     []string{"/foo/bar"},
		HostNames: []string{"v1"},
		TlsSecretDNS: map[string][]string{
			secretName: {"foo.com"},
		},
		ServiceName: svcName,
	}).IngressMultiPath()

	_, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	integrationtest.PollForCompletion(t, modelName, 5)

	verifyIng, _ := KubeClient.NetworkingV1().Ingresses("default").Get(context.TODO(), ingName, metav1.GetOptions{})
	for i, host := range []string{"foo.com", "noo.com"} {
		if verifyIng.Spec.Rules[i].Host == host {
			g.Expect(verifyIng.Spec.Rules[i].Host).To(gomega.Equal(host))
			g.Expect(verifyIng.Spec.Rules[i].HTTP.Paths[0].Path).To(gomega.Equal("/foo/bar"))
		}
	}

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 15*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes[0].Name).To(gomega.Equal("cluster--Shared-L7-0"))
	g.Expect(nodes[0].PoolGroupRefs[0].Name).To(gomega.Equal("cluster--Shared-L7-0"))
	g.Expect(nodes[0].PoolRefs[0].Name).To(gomega.Equal("cluster--noo.com_foo_bar-default-" + ingName))
	g.Expect(nodes[0].HTTPDSrefs[0].Name).To(gomega.Equal("cluster--Shared-L7-0"))
	g.Expect(nodes[0].VSVIPRefs[0].Name).To(gomega.Equal("cluster--Shared-L7-0"))
	g.Expect(nodes[0].SniNodes[0].Name).To(gomega.Equal("cluster--foo.com"))
	g.Expect(nodes[0].SniNodes[0].PoolGroupRefs[0].Name).To(gomega.Equal("cluster--default-foo.com_foo_bar-" + ingName))
	g.Expect(nodes[0].SniNodes[0].PoolRefs[0].Name).To(gomega.Equal("cluster--default-foo.com_foo_bar-" + ingName))
	g.Expect(nodes[0].SniNodes[0].SSLKeyCertRefs[0].Name).To(gomega.Equal("cluster--foo.com"))
	g.Expect(nodes[0].SniNodes[0].HttpPolicyRefs[0].HppMap[0].Name).To(gomega.Equal("cluster--default-foo.com_foo_bar-" + ingName))

	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), ingName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	TearDownTestForIngress(t, svcName, modelName)
}

func TestNoBackendL7Model(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName := MODEL_NAME_PREFIX + "0"
	svcName := objNameMap.GenerateName("avisvc")
	ingName := objNameMap.GenerateName("foo-with-targets")
	SetUpTestForIngress(t, svcName, modelName)

	integrationtest.PollForCompletion(t, modelName, 5)
	found, _ := objects.SharedAviGraphLister().Get(modelName)
	if found {
		// We shouldn't get an update for this update since it neither belongs to an ingress nor a L4 LB service
		t.Fatalf("Couldn't find Model for DELETE event %v", modelName)
	}
	ingrFake := (integrationtest.FakeIngress{
		Name:      ingName,
		Namespace: "default",
		DnsNames:  []string{"foo.com"},
		Paths:     []string{"/"},
	}).IngressOnlyHostNoBackend()

	_, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	integrationtest.PollForCompletion(t, modelName, 5)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 5*time.Second).Should(gomega.Equal(false))

	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), ingName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}

	TearDownTestForIngress(t, svcName, modelName)
}

func TestMultiIngressToSameSvc(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	modelName := MODEL_NAME_PREFIX + "0"
	svcName := objNameMap.GenerateName("avisvc")
	ingName := objNameMap.GenerateName("foo-with-targets")
	ingName2 := objNameMap.GenerateName("foo-with-targets")
	objects.SharedAviGraphLister().Delete(modelName)
	svcExample := (integrationtest.FakeService{
		Name:         svcName,
		Namespace:    "default",
		Type:         corev1.ServiceTypeClusterIP,
		ServicePorts: []integrationtest.Serviceport{{PortName: "foo", Protocol: "TCP", PortNumber: 8080, TargetPort: intstr.FromInt(8080)}},
	}).Service()

	_, err := KubeClient.CoreV1().Services("default").Create(context.TODO(), svcExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Service: %v", err)
	}

	integrationtest.CreateEPorEPS(t, "default", svcName, false, false, "1.1.1")
	ingrFake1 := (integrationtest.FakeIngress{
		Name:        ingName,
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		ServiceName: svcName,
	}).Ingress()

	_, err = KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake1, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	ingrFake2 := (integrationtest.FakeIngress{
		Name:        ingName2,
		Namespace:   "default",
		DnsNames:    []string{"bar.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		ServiceName: svcName,
	}).Ingress()

	_, err = KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake2, metav1.CreateOptions{})
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
			if pool.Name == "cluster--foo.com_foo-default-"+ingName {
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

	integrationtest.DelEPorEPS(t, "default", svcName)
	err = KubeClient.CoreV1().Services("default").Delete(context.TODO(), svcName, metav1.DeleteOptions{})
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

	integrationtest.CreateEPorEPS(t, "default", svcName, false, false, "1.1.1")

	//====== VERIFICATION OF ONE INGRESS DELETE
	// Now let's delete one ingress and expect the update for that.
	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), ingName, metav1.DeleteOptions{})
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
	_, err = KubeClient.CoreV1().Services("default").Create(context.TODO(), svcExample, metav1.CreateOptions{})
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

	integrationtest.DelEPorEPS(t, "default", svcName)
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
	err = KubeClient.CoreV1().Services("default").Delete(context.TODO(), svcName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Service %v", err)
	}
	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), ingName2, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
}

func TestMultiVSIngress(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	integrationtest.AddDefaultNamespace("randomNamespacethatyeildsdiff")
	modelName := MODEL_NAME_PREFIX + "0"
	svcName := objNameMap.GenerateName("avisvc")
	ingName := objNameMap.GenerateName("foo-with-targets")
	SetUpTestForIngress(t, svcName, modelName)

	ingrFake := (integrationtest.FakeIngress{
		Name:        ingName,
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		ServiceName: svcName,
	}).Ingress()

	_, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{})
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
		ServiceName: svcName,
	}).Ingress()
	_, err = KubeClient.NetworkingV1().Ingresses("randomNamespacethatyeildsdiff").Create(context.TODO(), randoming, metav1.CreateOptions{})
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
			if nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS(); len(nodes) > 0 {
				return len(nodes[0].PoolRefs)
			}
			return 0
		}, 10*time.Second).Should(gomega.Equal(2))

	} else {
		t.Fatalf("Could not find model: %v", err)
	}
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), ingName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	err = KubeClient.NetworkingV1().Ingresses("randomNamespacethatyeildsdiff").Delete(context.TODO(), "randomNamespacethatyeildsdiff", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngress(t, svcName, modelName)
}

func TestMultiPathIngress(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	var err error

	modelName := MODEL_NAME_PREFIX + "0"
	svcName := objNameMap.GenerateName("avisvc")
	SetUpTestForIngress(t, svcName, modelName)

	ingrFake := (integrationtest.FakeIngress{
		Name:        "ingress-multipath",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/foo", "/bar"},
		ServiceName: svcName,
	}).IngressMultiPath()

	_, err = KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{})
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
	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), "ingress-multipath", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngress(t, svcName, modelName)
}

func TestMultiPortServiceIngress(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	var err error

	modelName := MODEL_NAME_PREFIX + "0"
	svcName := objNameMap.GenerateName("avisvc")
	objects.SharedAviGraphLister().Delete(modelName)
	integrationtest.CreateSVC(t, "default", svcName, corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, true)
	integrationtest.CreateEPorEPS(t, "default", svcName, true, true, "1.1.1")
	ingrFake := (integrationtest.FakeIngress{
		Name:        "ingress-multipath",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/foo"},
		ServiceName: svcName,
	}).IngressMultiPort()

	_, err = KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{})
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
	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), "ingress-multipath", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngress(t, svcName, modelName)
}

func TestMultiIngressSameHost(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName := MODEL_NAME_PREFIX + "0"
	svcName := objNameMap.GenerateName("avisvc")
	SetUpTestForIngress(t, svcName, modelName)

	ingrFake1 := (integrationtest.FakeIngress{
		Name:        "ingress-multi1",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/foo"},
		ServiceName: svcName,
	}).Ingress()

	_, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake1, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	ingrFake2 := (integrationtest.FakeIngress{
		Name:        "ingress-multi2",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/bar"},
		ServiceName: svcName,
	}).Ingress()

	_, err = KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake2, metav1.CreateOptions{})
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
	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), "ingress-multi1", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 1)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(len(nodes)).To(gomega.Equal(1))
	g.Expect(nodes[0].PoolRefs[0].Name).To(gomega.Equal("cluster--foo.com_bar-default-ingress-multi2"))

	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), "ingress-multi2", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngress(t, svcName, modelName)
}

func TestDeleteBackendService(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName := MODEL_NAME_PREFIX + "0"
	svcName := objNameMap.GenerateName("avisvc")
	SetUpTestForIngress(t, svcName, modelName)

	ingrFake1 := (integrationtest.FakeIngress{
		Name:        "ingress-multi1",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/foo"},
		ServiceName: svcName,
	}).Ingress()

	_, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake1, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	ingrFake2 := (integrationtest.FakeIngress{
		Name:        "ingress-multi2",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/bar"},
		ServiceName: svcName,
	}).Ingress()

	_, err = KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake2, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	integrationtest.PollForCompletion(t, modelName, 5)
	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 15*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
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

	// Delete the service
	integrationtest.DelSVC(t, "default", svcName)
	integrationtest.DelEPorEPS(t, "default", svcName)
	found, aviModel := objects.SharedAviGraphLister().Get(modelName)
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
	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), "ingress-multi1", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 1)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(len(nodes)).To(gomega.Equal(1))
	g.Expect(nodes[0].PoolRefs[0].Name).To(gomega.Equal("cluster--foo.com_bar-default-ingress-multi2"))

	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), "ingress-multi2", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 0)

}

func TestUpdateBackendService(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	modelName := MODEL_NAME_PREFIX + "0"
	svcName := objNameMap.GenerateName("avisvc")
	svcName2 := objNameMap.GenerateName("avisvc")
	SetUpTestForIngress(t, svcName, modelName)
	ingrFake1 := (integrationtest.FakeIngress{
		Name:        "ingress-backend-svc",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/foo"},
		ServiceName: svcName,
	}).Ingress()
	_, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake1, metav1.CreateOptions{})
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

	integrationtest.CreateSVC(t, "default", svcName2, corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPorEPS(t, "default", svcName2, false, false, "2.2.2")

	_, err = (integrationtest.FakeIngress{
		Name:        "ingress-backend-svc",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/foo"},
		ServiceName: svcName2,
	}).UpdateIngress()
	if err != nil {
		t.Fatalf("error in updating ingress %s", err)
	}

	found, aviModel = objects.SharedAviGraphLister().Get(modelName)
	if found {
		g.Eventually(func() string {
			_, aviModel := objects.SharedAviGraphLister().Get(modelName)
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
			if len(nodes) > 0 && len(nodes[0].PoolRefs) > 0 && len(nodes[0].PoolRefs[0].Servers) > 0 {
				return *nodes[0].PoolRefs[0].Servers[0].Ip.Addr
			}
			return ""
		}, 10*time.Second).Should(gomega.Equal("2.2.2.1"))
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(len(nodes[0].PoolRefs)).To(gomega.Equal(1))
	} else {
		t.Fatalf("Could not find model: %s", modelName)
	}
	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), "ingress-backend-svc", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	integrationtest.DelSVC(t, "default", svcName)
	integrationtest.DelEPorEPS(t, "default", svcName)
	VerifyIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngress(t, svcName, modelName)
}

func TestL2ChecksumsUpdate(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	secretName := objNameMap.GenerateName("my-secret")
	integrationtest.AddSecret(secretName, "default", "tlsCert", "tlsKey")
	modelName := MODEL_NAME_PREFIX + "0"
	svcName := objNameMap.GenerateName("avisvc")
	svcName2 := objNameMap.GenerateName("avisvc")
	SetUpTestForIngress(t, svcName, modelName)
	//create ingress with tls secret
	ingrFake1 := (integrationtest.FakeIngress{
		Name:      "ingress-chksum",
		Namespace: "default",
		DnsNames:  []string{"foo.com"},
		Ips:       []string{"8.8.8.8"},
		Paths:     []string{"/foo"},
		HostNames: []string{"v1"},
		TlsSecretDNS: map[string][]string{
			secretName: {"foo.com"},
		},
		ServiceName: svcName,
	}).Ingress()
	_, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake1, metav1.CreateOptions{})
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

	integrationtest.CreateSVC(t, "default", svcName2, corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPorEPS(t, "default", svcName2, false, false, "2.2.2")
	secretName2 := secretName + "-new"
	integrationtest.AddSecret(secretName2, "default", "tlsCert-new", "tlsKey")

	_, err = (integrationtest.FakeIngress{
		Name:      "ingress-chksum",
		Namespace: "default",
		DnsNames:  []string{"foo.com"},
		Ips:       []string{"8.8.8.8"},
		//to update httppolicyref checksum
		Paths:     []string{"/bar"},
		HostNames: []string{"v1"},
		TlsSecretDNS: map[string][]string{
			//to update tls secret checksum
			secretName2: {"foo.com"},
		},
		//to update poolref checksum
		ServiceName: svcName2,
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
	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), "ingress-chksum", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	integrationtest.DelSVC(t, "default", svcName2)
	integrationtest.DelEPorEPS(t, "default", svcName2)
	KubeClient.CoreV1().Secrets("default").Delete(context.TODO(), secretName, metav1.DeleteOptions{})
	KubeClient.CoreV1().Secrets("default").Delete(context.TODO(), secretName2, metav1.DeleteOptions{})
	VerifySNIIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngress(t, svcName, modelName)
}

func TestSniHttpPolicy(t *testing.T) {
	/*
		-> Create Ingress with TLS key/secret and 2 paths
		-> Verify removing path works by updating Ingress with single path
		-> Verify adding path works by updating Ingress with 2 new paths
	*/

	g := gomega.NewGomegaWithT(t)
	secretName := objNameMap.GenerateName("my-secret")
	integrationtest.AddSecret(secretName, "default", "tlsCert", "tlsKey")
	modelName := MODEL_NAME_PREFIX + "0"
	svcName := objNameMap.GenerateName("avisvc")
	SetUpTestForIngress(t, svcName, modelName)
	ingrFake1 := (integrationtest.FakeIngress{
		Name:      "ingress-shp",
		Namespace: "default",
		DnsNames:  []string{"foo.com"},
		Ips:       []string{"8.8.8.8"},
		Paths:     []string{"/foo", "/bar"},
		HostNames: []string{"v1"},
		TlsSecretDNS: map[string][]string{
			secretName: {"foo.com"},
		},
		ServiceName: svcName,
	}).IngressMultiPath()
	_, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake1, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	integrationtest.PollForCompletion(t, modelName, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(modelName)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Eventually(len(nodes), 30*time.Second).Should(gomega.Equal(1))
		g.Expect(len(nodes[0].SniNodes), gomega.Equal(1))
		g.Expect(len(nodes[0].SniNodes[0].HttpPolicyRefs), gomega.Equal(1))
		g.Expect(len(nodes[0].SniNodes[0].HttpPolicyRefs[0].HppMap), gomega.Equal(2))
		g.Expect(len(nodes[0].SniNodes[0].HttpPolicyRefs[0].HppMap[0].Path)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].SniNodes[0].HttpPolicyRefs[0].HppMap[1].Path)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].SniNodes[0].SSLKeyCertRefs)).To(gomega.Equal(1))
		g.Expect(func() []string {
			p := []string{
				nodes[0].SniNodes[0].HttpPolicyRefs[0].HppMap[0].Path[0],
				nodes[0].SniNodes[0].HttpPolicyRefs[0].HppMap[1].Path[0]}
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

	_, err = (integrationtest.FakeIngress{
		Name:      "ingress-shp",
		Namespace: "default",
		DnsNames:  []string{"foo.com"},
		Ips:       []string{"8.8.8.8"},
		Paths:     []string{"/foo"},
		HostNames: []string{"v1"},
		TlsSecretDNS: map[string][]string{
			secretName: {"foo.com"},
		},
		ServiceName: svcName,
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
	g.Expect(len(nodes[0].SniNodes[0].HttpPolicyRefs[0].HppMap[0].Path)).To(gomega.Equal(1))
	g.Expect(nodes[0].SniNodes[0].HttpPolicyRefs[0].HppMap[0].Path[0]).To(gomega.Equal("/foo"))
	g.Expect(nodes[0].SniNodes[0].HttpPolicyRefs[0].Name).To(gomega.Equal("cluster--default-foo.com"))
	g.Expect(nodes[0].SniNodes[0].HttpPolicyRefs[0].HppMap[0].Name).To(gomega.Equal("cluster--default-foo.com_foo-ingress-shp"))
	g.Expect(len(nodes[0].SniNodes[0].SSLKeyCertRefs)).To(gomega.Equal(1))

	_, err = (integrationtest.FakeIngress{
		Name:      "ingress-shp",
		Namespace: "default",
		DnsNames:  []string{"foo.com"},
		Ips:       []string{"8.8.8.8"},
		Paths:     []string{"/foo", "/bar", "/baz"},
		HostNames: []string{"v1"},
		TlsSecretDNS: map[string][]string{
			secretName: {"foo.com"},
		},
		ServiceName: svcName,
	}).UpdateIngress()
	if err != nil {
		t.Fatalf("error in updating ingress %s", err)
	}
	integrationtest.PollForCompletion(t, modelName, 5)
	g.Eventually(func() int {
		found, aviModel = objects.SharedAviGraphLister().Get(modelName)
		if found {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
			return len(nodes[0].SniNodes[0].HttpPolicyRefs[0].HppMap)
		} else {
			return 0
		}
	}, 30*time.Second).Should(gomega.Equal(3))
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(len(nodes[0].SniNodes[0].HttpPolicyRefs[0].HppMap[0].Path)).To(gomega.Equal(1))
	g.Expect(len(nodes[0].SniNodes[0].HttpPolicyRefs[0].HppMap[1].Path)).To(gomega.Equal(1))
	g.Expect(len(nodes[0].SniNodes[0].HttpPolicyRefs[0].HppMap[2].Path)).To(gomega.Equal(1))
	g.Expect(func() []string {
		p := []string{
			nodes[0].SniNodes[0].HttpPolicyRefs[0].HppMap[0].Path[0],
			nodes[0].SniNodes[0].HttpPolicyRefs[0].HppMap[1].Path[0],
			nodes[0].SniNodes[0].HttpPolicyRefs[0].HppMap[2].Path[0]}
		sort.Strings(p)
		return p
	}, gomega.Equal([]string{"/bar", "/baz", "/foo"}))
	g.Expect(func() []string {
		p := []string{
			nodes[0].SniNodes[0].HttpPolicyRefs[0].HppMap[0].Name,
			nodes[0].SniNodes[0].HttpPolicyRefs[0].HppMap[1].Name,
			nodes[0].SniNodes[0].HttpPolicyRefs[0].HppMap[2].Name}
		sort.Strings(p)
		return p
	}, gomega.Equal([]string{"cluster--default-foo.com_bar-ingress-shp",
		"cluster--default-foo.com_baz-ingress-shp",
		"cluster--default-foo.com_foo-ingress-shp"}))
	g.Expect(len(nodes[0].SniNodes[0].SSLKeyCertRefs), gomega.Equal(1))

	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), "ingress-shp", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	KubeClient.CoreV1().Secrets("default").Delete(context.TODO(), secretName, metav1.DeleteOptions{})
	VerifySNIIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngress(t, svcName, modelName)
}

func TestFullSyncCacheNoOp(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	secretName := objNameMap.GenerateName("my-secret")
	integrationtest.AddSecret(secretName, "default", "tlsCert", "tlsKey")
	modelName := MODEL_NAME_PREFIX + "0"
	svcName := objNameMap.GenerateName("avisvc")
	SetUpTestForIngress(t, svcName, modelName)
	//create multipath ingress with tls secret
	ingrFake1 := (integrationtest.FakeIngress{
		Name:      "ingress-fsno",
		Namespace: "default",
		DnsNames:  []string{"foo.com"},
		Ips:       []string{"8.8.8.8"},
		Paths:     []string{"/foo", "/bar"},
		HostNames: []string{"v1"},
		TlsSecretDNS: map[string][]string{
			secretName: {"foo.com"},
		},
		ServiceName: svcName,
	}).IngressMultiPath()
	_, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake1, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	integrationtest.PollForCompletion(t, modelName, 5)
	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--foo.com"}

	mcache := cache.SharedAviObjCache()

	//store old chksum
	g.Eventually(func() bool {
		_, ok := mcache.VsCacheMeta.AviCacheGet(sniVSKey)
		return ok
	}, 30*time.Second).Should(gomega.Equal(true))

	oldSniCache, _ := mcache.VsCacheMeta.AviCacheGet(sniVSKey)
	oldSniCacheObj, _ := oldSniCache.(*cache.AviVsCache)
	oldChksum := oldSniCacheObj.CloudConfigCksum

	//call fullsync
	ctrl.FullSync()
	ctrl.FullSyncK8s(true)

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

	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), "ingress-fsno", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	KubeClient.CoreV1().Secrets("default").Delete(context.TODO(), secretName, metav1.DeleteOptions{})
	VerifySNIIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngress(t, svcName, modelName)
}

func TestMultiHostIngress(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	modelName := MODEL_NAME_PREFIX + "0"
	svcName := objNameMap.GenerateName("avisvc")
	SetUpTestForIngress(t, svcName, integrationtest.AllModels...)

	ingrFake := (integrationtest.FakeIngress{
		Name:        "ingress-multihost",
		Namespace:   "default",
		DnsNames:    []string{"foo.com", "bar.com"},
		Paths:       []string{"/foo", "/bar"},
		ServiceName: svcName,
	}).Ingress()

	_, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{})
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
	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), "ingress-multihost", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngress(t, svcName, integrationtest.AllModels...)
}

func TestMultiHostSameHostNameIngress(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	modelName := MODEL_NAME_PREFIX + "0"
	svcName := objNameMap.GenerateName("avisvc")
	SetUpTestForIngress(t, svcName, integrationtest.AllModels...)

	ingrFake := (integrationtest.FakeIngress{
		Name:        "ingress-multihost",
		Namespace:   "default",
		DnsNames:    []string{"foo.com", "foo.com"},
		Paths:       []string{"/foo", "/bar"},
		ServiceName: svcName,
	}).Ingress()

	_, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{})
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

	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), "ingress-multihost", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngress(t, svcName, modelName)
}

func TestEditPathIngress(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	modelName := MODEL_NAME_PREFIX + "0"
	svcName := objNameMap.GenerateName("avisvc")
	SetUpTestForIngress(t, svcName, modelName)

	ingrFake := (integrationtest.FakeIngress{
		Name:        "ingress-edit",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/foo"},
		ServiceName: svcName,
	}).Ingress()
	ingrFake.ResourceVersion = "1"
	_, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	integrationtest.PollForCompletion(t, modelName, 5)
	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 10*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
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

	ingrFake = (integrationtest.FakeIngress{
		Name:        "ingress-edit",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/bar"},
		ServiceName: svcName,
	}).Ingress()
	ingrFake.ResourceVersion = "2"
	_, err = KubeClient.NetworkingV1().Ingresses("default").Update(context.TODO(), ingrFake, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating Ingress: %v", err)
	}

	found, aviModel := objects.SharedAviGraphLister().Get(modelName)
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

	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), "ingress-edit", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngress(t, svcName, modelName)
}

func TestEditMultiPathIngress(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	modelName := MODEL_NAME_PREFIX + "0"
	svcName := objNameMap.GenerateName("avisvc")
	SetUpTestForIngress(t, svcName, modelName)

	ingrFake := (integrationtest.FakeIngress{
		Name:        "ingress-multipath-edit",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/foo"},
		ServiceName: svcName,
	}).Ingress()
	ingrFake.ResourceVersion = "1"

	_, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	integrationtest.PollForCompletion(t, modelName, 5)
	ingrFake = (integrationtest.FakeIngress{
		Name:        "ingress-multipath-edit",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/foo", "/bar"},
		ServiceName: svcName,
	}).IngressMultiPath()
	ingrFake.ResourceVersion = "2"
	_, err = KubeClient.NetworkingV1().Ingresses("default").Update(context.TODO(), ingrFake, metav1.UpdateOptions{})
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
		ServiceName: svcName,
	}).IngressMultiPath()
	ingrFake.ResourceVersion = "3"
	objects.SharedAviGraphLister().Delete(modelName)
	_, err = KubeClient.NetworkingV1().Ingresses("default").Update(context.TODO(), ingrFake, metav1.UpdateOptions{})
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

	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), "ingress-multipath-edit", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngress(t, svcName, modelName)
}

func TestEditMultiIngressSameHost(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName := MODEL_NAME_PREFIX + "0"
	svcName := objNameMap.GenerateName("avisvc")
	SetUpTestForIngress(t, svcName, modelName)

	ingrFake1 := (integrationtest.FakeIngress{
		Name:        "ingress-multi1",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/foo"},
		ServiceName: svcName,
	}).Ingress()

	_, err := integrationtest.KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake1, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	ingrFake2 := (integrationtest.FakeIngress{
		Name:        "ingress-multi2",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/bar"},
		ServiceName: svcName,
	}).Ingress()

	_, err = integrationtest.KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake2, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	ingrFake2 = (integrationtest.FakeIngress{
		Name:        "ingress-multi2",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/foobar"},
		ServiceName: svcName,
	}).Ingress()

	_, err = integrationtest.KubeClient.NetworkingV1().Ingresses("default").Update(context.TODO(), ingrFake2, metav1.UpdateOptions{})
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
		t.Fatalf("Could not find model: %s", modelName)
	}
	err = integrationtest.KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), "ingress-multi1", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 1)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(len(nodes)).To(gomega.Equal(1))
	g.Expect(nodes[0].PoolRefs[0].Name).To(gomega.Equal("cluster--foo.com_foobar-default-ingress-multi2"))

	err = integrationtest.KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), "ingress-multi2", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngress(t, svcName, modelName)
}

func TestEditMultiHostIngress(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	modelName := MODEL_NAME_PREFIX + "0"
	svcName := objNameMap.GenerateName("avisvc")
	SetUpTestForIngress(t, svcName, integrationtest.AllModels...)

	ingrFake := (integrationtest.FakeIngress{
		Name:        "ingress-multihost",
		Namespace:   "default",
		DnsNames:    []string{"foo.com", "bar.com"},
		Paths:       []string{"/foo", "/bar"},
		ServiceName: svcName,
	}).Ingress()

	_, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	ingrFake = (integrationtest.FakeIngress{
		Name:        "ingress-multihost",
		Namespace:   "default",
		DnsNames:    []string{"foo.com", "foobar.com"},
		Paths:       []string{"/foo", "/bar"},
		ServiceName: svcName,
	}).Ingress()

	_, err = KubeClient.NetworkingV1().Ingresses("default").Update(context.TODO(), ingrFake, metav1.UpdateOptions{})
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

	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), "ingress-multihost", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngress(t, svcName, modelName)
}

func TestNoHostIngress(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	modelName := MODEL_NAME_PREFIX + "2"
	svcName := objNameMap.GenerateName("avisvc")
	SetUpTestForIngress(t, svcName, modelName)

	ingrFake := (integrationtest.FakeIngress{
		Name:        "ingress-nohost",
		Namespace:   "default",
		Paths:       []string{"/foo"},
		ServiceName: svcName,
	}).IngressNoHost()

	_, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{})
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

	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), "ingress-nohost", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngress(t, svcName, modelName)
}

func TestEditNoHostIngress(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	modelName := MODEL_NAME_PREFIX + "2"
	svcName := objNameMap.GenerateName("avisvc")
	SetUpTestForIngress(t, svcName, modelName)

	ingrFake := (integrationtest.FakeIngress{
		Name:        "ingress-nohost",
		Namespace:   "default",
		Paths:       []string{"/foo"},
		ServiceName: svcName,
	}).IngressNoHost()
	_, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	ingrFake = (integrationtest.FakeIngress{
		Name:        "ingress-nohost",
		Namespace:   "default",
		Paths:       []string{"/bar"},
		ServiceName: svcName,
	}).IngressNoHost()
	ingrFake.ResourceVersion = "2"
	_, err = KubeClient.NetworkingV1().Ingresses("default").Update(context.TODO(), ingrFake, metav1.UpdateOptions{})
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

	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), "ingress-nohost", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngress(t, svcName, modelName)
}

func TestEditNoHostToHostIngress(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	modelName := MODEL_NAME_PREFIX + "2"
	svcName := objNameMap.GenerateName("avisvc")
	SetUpTestForIngress(t, svcName, modelName)

	ingrFake := (integrationtest.FakeIngress{
		Name:        "ingress-nohost",
		Namespace:   "default",
		Paths:       []string{"/foo"},
		ServiceName: svcName,
	}).IngressNoHost()

	_, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{})
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
		ServiceName: svcName,
	}).Ingress()

	ingrFake.ResourceVersion = "2"
	_, err = KubeClient.NetworkingV1().Ingresses("default").Update(context.TODO(), ingrFake, metav1.UpdateOptions{})
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
	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), "ingress-nohost", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngress(t, svcName, modelName)
}

func TestEditNoHostMultiPathIngress(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	modelName := MODEL_NAME_PREFIX + "3"
	svcName := objNameMap.GenerateName("avisvc")
	SetUpTestForIngress(t, svcName, modelName)

	ingrFake := (integrationtest.FakeIngress{
		Name:        "nohost-multipath",
		Namespace:   "default",
		Paths:       []string{"/foo", "/bar"},
		ServiceName: svcName,
	}).IngressNoHost()

	_, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	ingrFake = (integrationtest.FakeIngress{
		Name:        "nohost-multipath",
		Namespace:   "default",
		Paths:       []string{"/foo", "/foobar"},
		ServiceName: svcName,
	}).IngressNoHost()
	ingrFake.ResourceVersion = "2"
	_, err = KubeClient.NetworkingV1().Ingresses("default").Update(context.TODO(), ingrFake, metav1.UpdateOptions{})
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

	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), "nohost-multipath", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngress(t, svcName, modelName)
}

func TestScaleEndpoints(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName := MODEL_NAME_PREFIX + "0"
	svcName := objNameMap.GenerateName("avisvc")
	SetUpTestForIngress(t, svcName, modelName)

	ingrFake1 := (integrationtest.FakeIngress{
		Name:        "ingress-multi1",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/foo"},
		ServiceName: svcName,
	}).Ingress()

	_, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake1, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	ingrFake2 := (integrationtest.FakeIngress{
		Name:        "ingress-multi2",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/bar"},
		ServiceName: svcName,
	}).Ingress()

	_, err = KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake2, metav1.CreateOptions{})
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
	integrationtest.ScaleCreateEPorEPS(t, "default", svcName)
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

		g.Eventually(func() bool {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
			if len(nodes[0].PoolRefs) == 2 &&
				len(nodes[0].PoolRefs[1].Servers) == 2 {
				return true
			}
			return false
		}, 10*time.Second).Should(gomega.Equal(true))

		g.Expect(len(nodes[0].PoolGroupRefs)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].PoolGroupRefs[0].Members)).To(gomega.Equal(2))
	} else {
		t.Fatalf("Could not find model: %s", modelName)
	}
	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), "ingress-multi1", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 1)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(len(nodes)).To(gomega.Equal(1))
	g.Expect(nodes[0].PoolRefs[0].Name).To(gomega.Equal("cluster--foo.com_bar-default-ingress-multi2"))

	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), "ingress-multi2", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 0)
	TearDownTestForIngress(t, svcName, modelName)

}

// All SNI test cases follow:

func TestL7ModelSNI(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	secretName := objNameMap.GenerateName("my-secret")
	integrationtest.AddSecret(secretName, "default", "tlsCert", "tlsKey")
	modelName := MODEL_NAME_PREFIX + "0"
	svcName := objNameMap.GenerateName("avisvc")
	ingName := objNameMap.GenerateName("foo-with-targets")
	SetUpTestForIngress(t, svcName, modelName)

	integrationtest.PollForCompletion(t, modelName, 5)

	// foo.com and noo.com compute the same hashed shard vs num
	ingrFake := (integrationtest.FakeIngress{
		Name:      ingName,
		Namespace: "default",
		DnsNames:  []string{"foo.com", "noo.com"},
		Ips:       []string{"8.8.8.8"},
		HostNames: []string{"v1"},
		TlsSecretDNS: map[string][]string{
			secretName: {"foo.com"},
		},
		ServiceName: svcName,
	}).Ingress()

	_, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{})
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
	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), ingName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	KubeClient.CoreV1().Secrets("default").Delete(context.TODO(), secretName, metav1.DeleteOptions{})
	VerifySNIIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngress(t, svcName, modelName)
}

func TestL7ModelNoSecretToSecret(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	modelName := MODEL_NAME_PREFIX + "0"
	svcName := objNameMap.GenerateName("avisvc")
	secretName := objNameMap.GenerateName("my-secret")
	SetUpTestForIngress(t, svcName, modelName)

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
		ServiceName: svcName,
		TlsSecretDNS: map[string][]string{
			secretName: {"foo.com"},
		},
	}).Ingress()
	_, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{})
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
	integrationtest.AddSecret(secretName, "default", "tlsCert", "tlsKey")
	found, aviModel = objects.SharedAviGraphLister().Get(modelName)
	if found {
		g.Eventually(func() int {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
			return len(nodes[0].SniNodes)
		}, 10*time.Second).Should(gomega.Equal(1))

	} else {
		t.Fatalf("Could not find model: %s", modelName)
	}

	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), "foo-no-secret", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	KubeClient.CoreV1().Secrets("default").Delete(context.TODO(), secretName, metav1.DeleteOptions{})
	VerifySNIIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngress(t, svcName, modelName)
}

func TestL7ModelOneSecretToMultiIng(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	modelName := MODEL_NAME_PREFIX + "0"
	svcName := objNameMap.GenerateName("avisvc")
	secretName := objNameMap.GenerateName("my-secret")
	SetUpTestForIngress(t, svcName, modelName)

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
		ServiceName: svcName,
		TlsSecretDNS: map[string][]string{
			secretName: {"foo.com"},
		},
	}).Ingress()
	_, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake1, metav1.CreateOptions{})
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
			secretName: {"foo.com"},
		},
		ServiceName: svcName,
	}).Ingress()
	_, err = KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake2, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	integrationtest.PollForCompletion(t, modelName, 5)
	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 10*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(len(nodes)).To(gomega.Equal(1))
	g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
	g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
	g.Expect(len(nodes[0].SniNodes)).To(gomega.Equal(0))

	// Now create the secret and verify the models.
	integrationtest.AddSecret(secretName, "default", "tlsCert", "tlsKey")
	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 10*time.Second).Should(gomega.Equal(true))
	// Check if the secret affected both the models.
	g.Eventually(func() int {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return len(nodes[0].SniNodes)
	}, 10*time.Second).Should(gomega.Equal(1))
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes[0].SniNodes[0].VHDomainNames[0]).To(gomega.Equal("foo.com"))

	KubeClient.CoreV1().Secrets("default").Delete(context.TODO(), secretName, metav1.DeleteOptions{})
	VerifySNIIngressDeletion(t, g, aviModel, 0)
	// Since we deleted the secret, both SNIs should get removed.
	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 10*time.Second).Should(gomega.Equal(true))
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	// Check if the secret affected both the models.
	g.Eventually(func() int {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return len(nodes[0].SniNodes)
	}, 10*time.Second).Should(gomega.Equal(0))

	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), "foo-no-secret1", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), "foo-no-secret2", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	TearDownTestForIngress(t, svcName, modelName)
}

func TestL7ModelMultiSNI(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	secretName := objNameMap.GenerateName("my-secret")
	integrationtest.AddSecret(secretName, "default", "tlsCert", "tlsKey")
	modelName := MODEL_NAME_PREFIX + "0"
	svcName := objNameMap.GenerateName("avisvc")
	ingName := objNameMap.GenerateName("foo-with-targets")
	SetUpTestForIngress(t, svcName, modelName)

	ingrFake := (integrationtest.FakeIngress{
		Name:        ingName,
		Namespace:   "default",
		DnsNames:    []string{"foo.com", "bar.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		ServiceName: svcName,
		TlsSecretDNS: map[string][]string{
			secretName: {"foo.com", "bar.com"},
		},
	}).Ingress()
	_, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{})
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

	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), ingName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	KubeClient.CoreV1().Secrets("default").Delete(context.TODO(), secretName, metav1.DeleteOptions{})
	VerifySNIIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngress(t, svcName, modelName)
}

func TestL7ModelMultiSNIMultiCreateEditSecret(t *testing.T) {
	// This test covers creating multiple SNI nodes via multiple secrets.
	g := gomega.NewGomegaWithT(t)
	secretName := objNameMap.GenerateName("my-secret")
	secretName2 := objNameMap.GenerateName("my-secret")
	integrationtest.AddSecret(secretName, "default", "tlsCert", "tlsKey")
	integrationtest.AddSecret(secretName2, "default", "tlsCert", "tlsKey")
	// Clean up any earlier models.
	svcName := objNameMap.GenerateName("avisvc")
	ingName := objNameMap.GenerateName("foo-with-targets")
	modelName := MODEL_NAME_PREFIX + "1"
	objects.SharedAviGraphLister().Delete(modelName)
	modelName = "admin/cluster--Shared-L7-0"
	objects.SharedAviGraphLister().Delete(modelName)
	SetUpTestForIngress(t, svcName, modelName)

	ingrFake := (integrationtest.FakeIngress{
		Name:        ingName,
		Namespace:   "default",
		DnsNames:    []string{"foo.com", "FOO.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		ServiceName: svcName,
		TlsSecretDNS: map[string][]string{
			secretName:  {"foo.com"},
			secretName2: {"FOO.com"},
		},
	}).Ingress()

	_, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{})
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
		Name:        ingName,
		Namespace:   "default",
		DnsNames:    []string{"foo.com", "bar.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		ServiceName: svcName,
		TlsSecretDNS: map[string][]string{
			secretName:  {"foo.com"},
			secretName2: {"bar.com"},
		},
	}).Ingress()
	ingrFake.ResourceVersion = "2"
	_, err = KubeClient.NetworkingV1().Ingresses("default").Update(context.TODO(), ingrFake, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating Ingress: %v", err)
	}

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
	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), ingName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}

	KubeClient.CoreV1().Secrets("default").Delete(context.TODO(), secretName, metav1.DeleteOptions{})
	KubeClient.CoreV1().Secrets("default").Delete(context.TODO(), secretName2, metav1.DeleteOptions{})
	VerifySNIIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngress(t, svcName, modelName)
}

func TestL7WrongSubDomainMultiSNI(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	secretName := objNameMap.GenerateName("my-secret")
	secretName2 := objNameMap.GenerateName("my-secret")
	integrationtest.AddSecret(secretName, "default", "tlsCert", "tlsKey")
	integrationtest.AddSecret(secretName2, "default", "tlsCert", "tlsKey")
	modelName := MODEL_NAME_PREFIX + "1"
	svcName := objNameMap.GenerateName("avisvc")
	ingName := objNameMap.GenerateName("foo-with-targets")
	SetUpTestForIngress(t, svcName, modelName)

	ingrFake := (integrationtest.FakeIngress{
		Name:        ingName,
		Namespace:   "default",
		DnsNames:    []string{"foo.org"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		ServiceName: svcName,
		TlsSecretDNS: map[string][]string{
			secretName: {"foo.org"},
		},
	}).Ingress()
	_, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	integrationtest.PollForCompletion(t, modelName, 5)
	found, _ := objects.SharedAviGraphLister().Get(modelName)
	if found {
		// This will not generate a model.
		t.Fatalf("Could not find Model: %v", err)
	}
	ingrFake = (integrationtest.FakeIngress{
		Name:        ingName,
		Namespace:   "default",
		DnsNames:    []string{"foo.org", "bar.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		ServiceName: svcName,
		TlsSecretDNS: map[string][]string{
			secretName:  {"foo.org"},
			secretName2: {"bar.com"},
		},
	}).Ingress()
	ingrFake.ResourceVersion = "2"
	_, err = KubeClient.NetworkingV1().Ingresses("default").Update(context.TODO(), ingrFake, metav1.UpdateOptions{})
	integrationtest.PollForCompletion(t, modelName, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(modelName)
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
	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), ingName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	KubeClient.CoreV1().Secrets("default").Delete(context.TODO(), secretName, metav1.DeleteOptions{})
	VerifySNIIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngress(t, svcName, modelName)
}

func TestClusterRuntimeUpSinceChange(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	onBootup := true

	// Injecting middleware for cluster runtime up_since time changes for api shutdown
	integrationtest.AddMiddleware(func(w http.ResponseWriter, r *http.Request) {
		url := r.URL.EscapedPath()
		if r.Method == "GET" && strings.Contains(url, "/api/cluster/runtime") {
			if onBootup {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"node_states": [{"name": "10.79.169.60","role": "CLUSTER_LEADER","up_since": "2020-10-28 04:58:48"}]}`))
				onBootup = false
			} else {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"node_states": [{"name": "10.79.169.60","role": "CLUSTER_LEADER","up_since": "2020-10-28 05:58:48"}]}`))
			}
			return
		}
		integrationtest.NormalControllerServer(w, r)
	})

	ctrl.FullSync()
	ctrl.FullSync()
	g.Eventually(func() bool {
		return akoApiServer.Shutdown
	}, 60*time.Second).Should(gomega.Equal(true))
	integrationtest.ResetMiddleware()
}

func TestFQDNCountInL7Model(t *testing.T) {
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
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if aviModel == nil {
			return 0
		}
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return len(nodes)
	}, 10*time.Second).Should(gomega.Equal(1))

	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	node := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0]

	g.Expect(node.VSVIPRefs).To(gomega.HaveLen(1))
	g.Expect(node.VSVIPRefs[0].FQDNs).To(gomega.HaveLen(2))
	for _, fqdn := range node.VSVIPRefs[0].FQDNs {
		if fqdn == "foo.com" {
			continue
		}
		g.Expect(fqdn).Should(gomega.ContainSubstring("Shared-L7"))
	}

	TearDownIngressForCacheSyncCheck(t, ingName, svcName, secretName, modelName)
}

func TestPortsForInsecureAndSecureSNI(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	modelName := MODEL_NAME_PREFIX + "0"
	svcName := objNameMap.GenerateName("avisvc")
	secretName := objNameMap.GenerateName("my-secret")
	SetUpTestForIngress(t, svcName, modelName)

	// Insecure
	integrationtest.PollForCompletion(t, modelName, 5)
	found, _ := objects.SharedAviGraphLister().Get(modelName)
	if found {
		t.Fatalf("Couldn't find Model for DELETE event %v", modelName)
	}
	ingrFake := (integrationtest.FakeIngress{
		Name:        "foo-no-secret",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		ServiceName: svcName,
		TlsSecretDNS: map[string][]string{
			secretName: {"foo.com"},
		},
	}).Ingress()
	_, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	integrationtest.PollForCompletion(t, modelName, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(modelName)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(nodes[0].PortProto).To(gomega.HaveLen(2))
		var ports []int
		for _, port := range nodes[0].PortProto {
			ports = append(ports, int(port.Port))
			if port.EnableSSL {
				g.Expect(int(port.Port)).To(gomega.Equal(443))
			}
		}
		sort.Ints(ports)
		g.Expect(ports[0]).To(gomega.Equal(80))
		g.Expect(ports[1]).To(gomega.Equal(443))
	} else {
		t.Fatalf("Could not find Model: %v", err)
	}

	// Secure
	integrationtest.AddSecret(secretName, "default", "tlsCert", "tlsKey")
	found, aviModel = objects.SharedAviGraphLister().Get(modelName)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		var ports []int
		for _, port := range nodes[0].PortProto {
			ports = append(ports, int(port.Port))
			if port.EnableSSL {
				g.Expect(int(port.Port)).To(gomega.Equal(443))
			}
		}
		sort.Ints(ports)
		g.Expect(ports[0]).To(gomega.Equal(80))
		g.Expect(ports[1]).To(gomega.Equal(443))

	} else {
		t.Fatalf("Could not find model: %s", modelName)
	}

	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), "foo-no-secret", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	KubeClient.CoreV1().Secrets("default").Delete(context.TODO(), secretName, metav1.DeleteOptions{})

	TearDownTestForIngress(t, svcName, modelName)
}

func TestV6BackendService(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	modelName := MODEL_NAME_PREFIX + "0"
	svcName := objNameMap.GenerateName("avisvc")
	objects.SharedAviGraphLister().Delete(modelName)

	v6Svc := integrationtest.ConstructService("default", svcName, corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false, make(map[string]string), "")
	ipFamilyPolicy := corev1.IPFamilyPolicy("SingleStack")
	v6Svc.Spec.IPFamilies = []corev1.IPFamily{"IPv6"}
	v6Svc.Spec.IPFamilyPolicy = &ipFamilyPolicy
	_, err := KubeClient.CoreV1().Services("default").Create(context.TODO(), v6Svc, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Service: %v", err)
	}
	integrationtest.CreateEPorEPS(t, "default", svcName, false, false, "ff06::")

	ingrFake1 := (integrationtest.FakeIngress{
		Name:        "ingress-v6-backend-svc",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/foo"},
		ServiceName: svcName,
	}).Ingress()
	_, err = KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake1, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	integrationtest.PollForCompletion(t, modelName, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(modelName)
	if found {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(*nodes[0].PoolRefs[0].Servers[0].Ip.Addr).To(gomega.Equal("ff06::1"))

	} else {
		t.Fatalf("Could not find model: %s", modelName)
	}

	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), "ingress-v6-backend-svc", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	integrationtest.DelSVC(t, "default", svcName)
	integrationtest.DelEPorEPS(t, "default", svcName)
	VerifyIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngress(t, svcName, modelName)
}

func TestDualStackBackendService(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	modelName := MODEL_NAME_PREFIX + "0"
	svcName := objNameMap.GenerateName("avisvc")
	objects.SharedAviGraphLister().Delete(modelName)

	dsSvc := integrationtest.ConstructService("default", svcName, corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false, make(map[string]string), "")
	ipFamilyPolicy := corev1.IPFamilyPolicy("RequireDualStack")
	dsSvc.Spec.IPFamilies = []corev1.IPFamily{"IPv4", "IPv6"}
	dsSvc.Spec.IPFamilyPolicy = &ipFamilyPolicy
	_, err := KubeClient.CoreV1().Services("default").Create(context.TODO(), dsSvc, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Service: %v", err)
	}
	integrationtest.CreateEPorEPS(t, "default", svcName, false, false, "1.1.1")

	ingrFake1 := (integrationtest.FakeIngress{
		Name:        "ingress-ds-backend-svc",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/foo"},
		ServiceName: svcName,
	}).Ingress()
	_, err = KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake1, metav1.CreateOptions{})
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

	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), "ingress-ds-backend-svc", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	integrationtest.DelSVC(t, "default", svcName)
	integrationtest.DelEPorEPS(t, "default", svcName)
	VerifyIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngress(t, svcName, modelName)
}

func TestDualStackMultipleBackendService(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	modelName := MODEL_NAME_PREFIX + "0"
	objects.SharedAviGraphLister().Delete(modelName)
	svcName := objNameMap.GenerateName("avisvc")
	svcName2 := objNameMap.GenerateName("avisvc")

	v4Svc := integrationtest.ConstructService("default", svcName, corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false, make(map[string]string), "")
	ipFamilyPolicy := corev1.IPFamilyPolicy("SingleStack")
	v4Svc.Spec.IPFamilies = []corev1.IPFamily{"IPv4"}
	v4Svc.Spec.IPFamilyPolicy = &ipFamilyPolicy
	_, err := KubeClient.CoreV1().Services("default").Create(context.TODO(), v4Svc, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Service: %v", err)
	}
	integrationtest.CreateEPorEPS(t, "default", svcName, false, false, "1.1.1")

	v6Svc := integrationtest.ConstructService("default", svcName2, corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false, make(map[string]string), "")
	v6Svc.Spec.IPFamilies = []corev1.IPFamily{"IPv6"}
	v6Svc.Spec.IPFamilyPolicy = &ipFamilyPolicy
	_, err = KubeClient.CoreV1().Services("default").Create(context.TODO(), v6Svc, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Service: %v", err)
	}
	integrationtest.CreateEPorEPS(t, "default", svcName2, false, false, "ff06::")

	ingrFake1 := (integrationtest.FakeIngress{
		Name:        "ingress-ds-multipath",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/foo", "/bar"},
		ServiceName: svcName,
	}).IngressMultiPath()

	ingrFake1.Spec.Rules[0].IngressRuleValue.HTTP.Paths[1].Backend.Service.Name = svcName2
	_, err = KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake1, metav1.CreateOptions{})
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
			if pool.Name == "cluster--foo.com_foo-default-ingress-ds-multipath" {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("foo.com/foo"))
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
				g.Expect(*pool.Servers[0].Ip.Addr).To(gomega.Equal("1.1.1.1"))
			} else if pool.Name == "cluster--foo.com_bar-default-ingress-ds-multipath" {
				g.Expect(pool.PriorityLabel).To(gomega.Equal("foo.com/bar"))
				g.Expect(len(pool.Servers)).To(gomega.Equal(1))
				g.Expect(*pool.Servers[0].Ip.Addr).To(gomega.Equal("ff06::1"))
			} else {
				t.Fatalf("unexpected pool: %s", pool.Name)
			}
		}
		g.Expect(len(nodes[0].PoolGroupRefs)).To(gomega.Equal(1))
		g.Expect(len(nodes[0].PoolGroupRefs[0].Members)).To(gomega.Equal(2))
		for _, pool := range nodes[0].PoolGroupRefs[0].Members {
			if *pool.PoolRef == "/api/pool?name=cluster--foo.com_foo-default-ingress-ds-multipath" {
				g.Expect(*pool.PriorityLabel).To(gomega.Equal("foo.com/foo"))
			} else if *pool.PoolRef == "/api/pool?name=cluster--foo.com_bar-default-ingress-ds-multipath" {
				g.Expect(*pool.PriorityLabel).To(gomega.Equal("foo.com/bar"))
			} else {
				t.Fatalf("unexpected pool: %s", *pool.PoolRef)
			}
		}

	} else {
		t.Fatalf("Could not find model: %s", modelName)
	}

	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), "ingress-ds-multipath", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	integrationtest.DelSVC(t, "default", svcName)
	integrationtest.DelEPorEPS(t, "default", svcName)
	integrationtest.DelSVC(t, "default", svcName2)
	integrationtest.DelEPorEPS(t, "default", svcName2)
	VerifyIngressDeletion(t, g, aviModel, 0)

	TearDownTestForIngress(t, svcName, modelName)
}
