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

package dedicatedvstests

import (
	"context"
	"os"
	"sort"
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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

var (
	KubeClient           *k8sfake.Clientset
	CRDClient            *crdfake.Clientset
	V1beta1Client        *v1beta1crdfake.Clientset
	endpointSliceEnabled bool
	ctrl                 *k8s.AviController
	akoApiServer         *api.FakeApiServer
	objNameMap           integrationtest.ObjectNameMap
)

func TestMain(m *testing.M) {
	os.Setenv("INGRESS_API", "extensionv1")
	os.Setenv("VIP_NETWORK_LIST", `[{"networkName":"net123"}]`)
	os.Setenv("CLUSTER_NAME", "cluster")
	os.Setenv("CLOUD_NAME", "CLOUD_VCENTER")
	os.Setenv("SEG_NAME", "Default-Group")
	os.Setenv("NODE_NETWORK_LIST", `[{"networkName":"net123","cidrs":["10.79.168.0/22"]}]`)
	os.Setenv("POD_NAMESPACE", utils.AKO_DEFAULT_NS)
	os.Setenv("SHARD_VS_SIZE", "DEDICATED")
	os.Setenv("AUTO_L4_FQDN", "default")
	os.Setenv("POD_NAME", "ako-0")

	akoControlConfig := lib.AKOControlConfig()
	endpointSliceEnabled = lib.GetEndpointSliceEnabled()
	akoControlConfig.SetEndpointSlicesEnabled(endpointSliceEnabled)
	akoControlConfig.SetAKOInstanceFlag(true)
	KubeClient = k8sfake.NewSimpleClientset()
	CRDClient = crdfake.NewSimpleClientset()
	V1beta1Client = v1beta1crdfake.NewSimpleClientset()
	akoControlConfig.SetCRDClientset(CRDClient)
	akoControlConfig.Setv1beta1CRDClientset(V1beta1Client)
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

	ctrl.HandleConfigMap(informers, ctrlCh, stopCh, quickSyncCh)
	integrationtest.KubeClient = KubeClient
	integrationtest.AddDefaultIngressClass()
	ctrl.SetSEGroupCloudNameFromNSAnnotations()
	integrationtest.AddDefaultNamespace()

	go ctrl.InitController(informers, registeredInformers, ctrlCh, stopCh, quickSyncCh, waitGroupMap)
	objNameMap.InitMap()
	os.Exit(m.Run())
}

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

func SetUpTestForIngress(t *testing.T, svcName string, modelNames ...string) {
	for _, model := range modelNames {
		objects.SharedAviGraphLister().Delete(model)
	}
	integrationtest.CreateSVC(t, "default", svcName, corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPorEPS(t, "default", svcName, false, false, "1.1.1")
}

func TearDownTestForIngress(t *testing.T, svcName string, modelNames ...string) {
	for _, model := range modelNames {
		objects.SharedAviGraphLister().Delete(model)
	}
	integrationtest.DelSVC(t, "default", svcName)
	integrationtest.DelEPorEPS(t, "default", svcName)
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
			ingTestObj.secretName: {"foo.com"},
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

func SetUpTestForIngressInNodePortMode(t *testing.T, svcName, model_Name, externalTrafficPolicy string) {
	objects.SharedAviGraphLister().Delete(model_Name)
	if externalTrafficPolicy == "" {
		integrationtest.CreateSVC(t, "default", svcName, corev1.ProtocolTCP, corev1.ServiceTypeNodePort, false)
	} else {
		integrationtest.CreateSvcWithExternalTrafficPolicy(t, "default", svcName, corev1.ProtocolTCP, corev1.ServiceTypeNodePort, false, externalTrafficPolicy)
	}
}

func TearDownTestForIngressInNodePortMode(t *testing.T, svcName, model_Name string) {
	objects.SharedAviGraphLister().Delete(model_Name)
	integrationtest.DelSVC(t, "default", svcName)
}

func VerifyIngressDeletion(t *testing.T, g *gomega.WithT, aviModel interface{}, poolCount int) {
	var nodes []*avinodes.AviVsNode
	g.Eventually(func() []*avinodes.AviPoolNode {
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return nodes[0].PoolRefs
	}, 10*time.Second).Should(gomega.HaveLen(poolCount))

	g.Eventually(func() []*avinodes.AviPoolGroupNode {
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return nodes[0].PoolGroupRefs
	}, 10*time.Second).Should(gomega.HaveLen(poolCount))
}

func SetupDomain() {
	mcache := cache.SharedAviObjCache()
	cloudObj := &cache.AviCloudPropertyCache{Name: "Default-Cloud", VType: "mock"}
	subdomains := []string{"avi.internal", ".com"}
	cloudObj.NSIpamDNS = subdomains
	mcache.CloudKeyCache.AviCacheAdd("Default-Cloud", cloudObj)
}

func TearDownIngressForCacheSyncCheck(t *testing.T, secretName, ingressName, svcName, modelName string) {
	if err := KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), ingressName, metav1.DeleteOptions{}); err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	if secretName != "" {
		KubeClient.CoreV1().Secrets("default").Delete(context.TODO(), secretName, metav1.DeleteOptions{})
	}
	TearDownTestForIngress(t, svcName, modelName)
}

func TestFQDNCountInL7Model(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	modelName := "admin/cluster--foo.com-L7-dedicated"
	secretName := objNameMap.GenerateName("my-secret")
	ingressName := objNameMap.GenerateName("foo-with-targets")
	svcName := objNameMap.GenerateName("avisvc")
	ingTestObj := IngressTestObject{
		ingressName: ingressName,
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
	g.Expect(node.VSVIPRefs[0].FQDNs).To(gomega.HaveLen(1))
	for _, fqdn := range node.VSVIPRefs[0].FQDNs {
		g.Expect(fqdn).ShouldNot(gomega.ContainSubstring("L7-dedicated"))
	}

	TearDownIngressForCacheSyncCheck(t, secretName, ingressName, svcName, modelName)
}

func TestPortsForInsecureDedicatedShard(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	modelName := "admin/cluster--foo.com-L7-dedicated"

	ingressName := objNameMap.GenerateName("foo-with-targets")
	svcName := objNameMap.GenerateName("avisvc")

	ingTestObj := IngressTestObject{
		ingressName: ingressName,
		isTLS:       false,
		withSecret:  false,
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
	}, 20*time.Second).Should(gomega.Equal(1))

	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes[0].PortProto).To(gomega.HaveLen(1))
	g.Expect(int(nodes[0].PortProto[0].Port)).To(gomega.Equal(80))
	TearDownIngressForCacheSyncCheck(t, "", ingressName, svcName, modelName)
}

func TestPlacementNetworkDedicatedNodePort(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	nodeName := "testNodeNP"
	integrationtest.SetNodePortMode()
	defer integrationtest.SetClusterIPMode()
	nodeIP := "10.1.1.2"
	integrationtest.CreateNode(t, nodeName, nodeIP)
	defer integrationtest.DeleteNode(t, nodeName)

	modelName := "admin/cluster--foo.com-L7-dedicated"
	ingressName := objNameMap.GenerateName("foo-with-targets")
	svcName := objNameMap.GenerateName("avisvc")
	SetUpTestForIngressInNodePortMode(t, svcName, modelName, "")

	ingrFake := (integrationtest.FakeIngress{
		Name:        ingressName,
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

	g.Eventually(func() int {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes, ok := aviModel.(*avinodes.AviObjectGraph)
		if !ok {
			return 0
		}
		return len(nodes.GetAviVS())
	}, 20*time.Second).Should(gomega.Equal(1))

	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes[0].PoolRefs).To(gomega.HaveLen(1))
	// pool server is added for testNodeNP node even though endpointslice/endpoint does not exist
	g.Eventually(func() int {
		return len(nodes[0].PoolRefs[0].Servers)
	}, 30*time.Second).Should(gomega.Equal(1))
	g.Expect(*nodes[0].PoolRefs[0].Servers[0].Ip.Addr).To(gomega.Equal(nodeIP))
	g.Expect((nodes[0].PoolRefs[0].NetworkPlacementSettings)).To(gomega.HaveLen(1))
	_, ok := nodes[0].PoolRefs[0].NetworkPlacementSettings["net123"]
	g.Expect(ok).To(gomega.Equal(true))

	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), ingressName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	VerifyIngressDeletion(t, g, aviModel, 0)
	TearDownTestForIngressInNodePortMode(t, svcName, modelName)
}

// TestL7ModelDedicatedNodePortExternalTrafficPolicyLocal checks if pool servers are populated in model only for nodes that are running the app pod.
func TestL7ModelDedicatedNodePortExternalTrafficPolicyLocal(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	nodeName := "testNodeNP"
	integrationtest.SetNodePortMode()
	defer integrationtest.SetClusterIPMode()
	nodeIP := "10.1.1.2"
	integrationtest.CreateNode(t, nodeName, nodeIP)
	defer integrationtest.DeleteNode(t, nodeName)

	modelName := "admin/cluster--foo.com-L7-dedicated"
	ingressName := objNameMap.GenerateName("foo-with-targets")
	svcName := objNameMap.GenerateName("avisvc")
	SetUpTestForIngressInNodePortMode(t, svcName, modelName, "Local")

	ingrFake := (integrationtest.FakeIngress{
		Name:        ingressName,
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

	g.Eventually(func() int {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes, ok := aviModel.(*avinodes.AviObjectGraph)
		if !ok {
			return 0
		}
		return len(nodes.GetAviVS())
	}, 20*time.Second).Should(gomega.Equal(1))

	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes[0].PoolRefs).To(gomega.HaveLen(1))
	// No pool server is added as endpointslice/endpoint does not exist
	g.Expect(nodes[0].PoolRefs[0].Servers).To(gomega.HaveLen(0))
	g.Expect((nodes[0].PoolRefs[0].NetworkPlacementSettings)).To(gomega.HaveLen(1))
	_, ok := nodes[0].PoolRefs[0].NetworkPlacementSettings["net123"]
	g.Expect(ok).To(gomega.Equal(true))

	integrationtest.CreateEPorEPSNodeName(t, "default", svcName, false, false, "1.1.1", nodeName)
	// After creating the endpointslice/endpoint, pool server should be added for testNodeNP node
	g.Eventually(func() int {
		return len(nodes[0].PoolRefs[0].Servers)
	}, 30*time.Second).Should(gomega.Equal(1))
	g.Expect(*nodes[0].PoolRefs[0].Servers[0].Ip.Addr).To(gomega.Equal(nodeIP))

	err = KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), ingressName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	integrationtest.DelEPorEPS(t, "default", svcName)
	VerifyIngressDeletion(t, g, aviModel, 0)
	TearDownTestForIngressInNodePortMode(t, svcName, modelName)
}

func TestPortsForSecureDedicatedShard(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	modelName := "admin/cluster--foo.com-L7-dedicated"
	secretName := objNameMap.GenerateName("my-secret")
	ingressName := objNameMap.GenerateName("foo-with-targets")
	svcName := objNameMap.GenerateName("avisvc")

	ingTestObj := IngressTestObject{
		ingressName: ingressName,
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
	}, 20*time.Second).Should(gomega.Equal(1))

	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
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
	TearDownIngressForCacheSyncCheck(t, secretName, ingressName, svcName, modelName)
}
