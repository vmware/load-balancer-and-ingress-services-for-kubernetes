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
	v1beta1crdfake "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/client/v1beta1/clientset/versioned/fake"

	v1alpha2crdfake "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/client/v1alpha2/clientset/versioned/fake"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

var objNameMap ObjectNameMap
var MODEL_REDNS_PREFIX = "admin/cluster--red-ns-"

func SetUpTestForSvcLB(t *testing.T, svcName string) {
	modelName := MODEL_REDNS_PREFIX + svcName
	objects.SharedAviGraphLister().Delete(modelName)
	CreateSVC(t, NAMESPACE, svcName, corev1.ProtocolTCP, corev1.ServiceTypeLoadBalancer, false)
	CreateEPS(t, NAMESPACE, svcName, false, false, "1.1.1")
	PollForCompletion(t, modelName, 5)
}

func SetUpTestForSvcLBWithExternalTrafficPolicy(t *testing.T, svcName, nodeName, externalTrafficPolicy string) {
	modelName := MODEL_REDNS_PREFIX + svcName
	objects.SharedAviGraphLister().Delete(modelName)
	if externalTrafficPolicy == "" {
		CreateSVC(t, NAMESPACE, svcName, corev1.ProtocolTCP, corev1.ServiceTypeLoadBalancer, false)
	} else {
		CreateSvcWithExternalTrafficPolicy(t, NAMESPACE, svcName, corev1.ProtocolTCP, corev1.ServiceTypeLoadBalancer, false, externalTrafficPolicy)
	}
	CreateEPSNodeName(t, NAMESPACE, svcName, false, false, "1.1.1", nodeName)
	PollForCompletion(t, modelName, 5)
}

func SetUpTestForSvcLBWithLBClass(t *testing.T, LBClass, svcName string) {
	modelName := MODEL_REDNS_PREFIX + svcName
	objects.SharedAviGraphLister().Delete(modelName)
	CreateSVCWithValidOrInvalidLBClass(t, NAMESPACE, svcName, corev1.ProtocolTCP, corev1.ServiceTypeLoadBalancer, false, LBClass)
	CreateEPS(t, NAMESPACE, svcName, false, false, "1.1.1")
	PollForCompletion(t, modelName, 5)
}

func TearDownTestForSvcLB(t *testing.T, g *gomega.GomegaWithT, svcName string) {
	modelName := MODEL_REDNS_PREFIX + svcName
	objects.SharedAviGraphLister().Delete(modelName)
	DelSVC(t, NAMESPACE, svcName)
	DelEPS(t, NAMESPACE, svcName)
	mcache := cache.SharedAviObjCache()
	vsKey := cache.NamespaceName{Namespace: AVINAMESPACE, Name: fmt.Sprintf("cluster--%s-%s", NAMESPACE, svcName)}
	g.Eventually(func() bool {
		_, found := mcache.VsCacheMeta.AviCacheGet(vsKey)
		return found
	}, 5*time.Second).Should(gomega.Equal(false))
}

func SetUpTestForSvcLBWithExtDNS(t *testing.T) {
	modelSvcDNS01 := MODEL_REDNS_PREFIX + EXTDNSSVC
	objects.SharedAviGraphLister().Delete(modelSvcDNS01)
	svcObj := ConstructService(NAMESPACE, EXTDNSSVC, corev1.ProtocolTCP, corev1.ServiceTypeLoadBalancer, false, make(map[string]string), "")
	svcObj.Annotations = map[string]string{lib.ExternalDNSAnnotation: EXTDNSANNOTATION}
	_, err := KubeClient.CoreV1().Services(NAMESPACE).Create(context.TODO(), svcObj, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Service: %v", err)
	}
	CreateEPS(t, NAMESPACE, EXTDNSSVC, false, false, "1.1.1")
	PollForCompletion(t, modelSvcDNS01, 5)
}

func TearDownTestForSvcLBWithExtDNS(t *testing.T, g *gomega.GomegaWithT) {
	modelSvcDNS01 := MODEL_REDNS_PREFIX + EXTDNSSVC
	objects.SharedAviGraphLister().Delete(modelSvcDNS01)
	DelSVC(t, NAMESPACE, EXTDNSSVC)
	DelEPS(t, NAMESPACE, EXTDNSSVC)
	mcache := cache.SharedAviObjCache()
	vsKey := cache.NamespaceName{Namespace: AVINAMESPACE, Name: fmt.Sprintf("cluster--%s-%s", NAMESPACE, EXTDNSSVC)}
	g.Eventually(func() bool {
		_, found := mcache.VsCacheMeta.AviCacheGet(vsKey)
		return found
	}, 5*time.Second).Should(gomega.Equal(false))
}

func SetUpTestForSvcLBMultiport(t *testing.T, svcName string) {
	modelName := MODEL_REDNS_PREFIX + svcName
	objects.SharedAviGraphLister().Delete(modelName)
	CreateSVC(t, NAMESPACE, svcName, corev1.ProtocolTCP, corev1.ServiceTypeLoadBalancer, true)
	CreateEPS(t, NAMESPACE, svcName, true, true, "1.1.1")
	PollForCompletion(t, modelName, 10)
}

func SetUpTestForSvcLBMixedProtocol(t *testing.T, svcName string, multiProtocol ...corev1.Protocol) {
	modelName := MODEL_REDNS_PREFIX + svcName
	objects.SharedAviGraphLister().Delete(modelName)
	CreateSVC(t, NAMESPACE, svcName, corev1.ProtocolTCP, corev1.ServiceTypeLoadBalancer, false, multiProtocol...)
	CreateEPS(t, NAMESPACE, svcName, false, false, "1.1.1", multiProtocol...)
	PollForCompletion(t, modelName, 10)
}

func TearDownTestForSvcLBMultiport(t *testing.T, g *gomega.GomegaWithT, svcName string) {
	modelName := MODEL_REDNS_PREFIX + svcName
	objects.SharedAviGraphLister().Delete(modelName)
	DelSVC(t, NAMESPACE, svcName)
	DelEPS(t, NAMESPACE, svcName)
	mcache := cache.SharedAviObjCache()
	vsKey := cache.NamespaceName{Namespace: AVINAMESPACE, Name: fmt.Sprintf("cluster--%s-%s", NAMESPACE, svcName)}
	g.Eventually(func() bool {
		_, found := mcache.VsCacheMeta.AviCacheGet(vsKey)
		return found
	}, 5*time.Second).Should(gomega.Equal(false))
}

func SetUpTestForSharedVIPSvcLB(t *testing.T, proto1, proto2 corev1.Protocol) {
	modelSvc01 := MODEL_REDNS_PREFIX + SHAREDVIPSVC01
	objects.SharedAviGraphLister().Delete(modelSvc01)
	svcObj := ConstructService(NAMESPACE, SHAREDVIPSVC01, proto1, corev1.ServiceTypeLoadBalancer, false, make(map[string]string), "")
	svcObj.Annotations = map[string]string{lib.SharedVipSvcLBAnnotation: SHAREDVIPKEY}
	_, err := KubeClient.CoreV1().Services(NAMESPACE).Create(context.TODO(), svcObj, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Service: %v", err)
	}
	CreateEPS(t, NAMESPACE, SHAREDVIPSVC01, false, false, "1.1.1")
	PollForCompletion(t, modelSvc01, 5)

	modelSvc02 := MODEL_REDNS_PREFIX + SHAREDVIPSVC02
	objects.SharedAviGraphLister().Delete(modelSvc01)
	svcObj = ConstructService(NAMESPACE, SHAREDVIPSVC02, proto2, corev1.ServiceTypeLoadBalancer, false, make(map[string]string), "")
	svcObj.Annotations = map[string]string{lib.SharedVipSvcLBAnnotation: SHAREDVIPKEY}
	_, err = KubeClient.CoreV1().Services(NAMESPACE).Create(context.TODO(), svcObj, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Service: %v", err)
	}
	CreateEPS(t, NAMESPACE, SHAREDVIPSVC02, false, false, "2.1.1")
	PollForCompletion(t, modelSvc02, 5)
}

func TearDownTestForSharedVIPSvcLB(t *testing.T, g *gomega.GomegaWithT) {
	modelSvc01 := MODEL_REDNS_PREFIX + SHAREDVIPSVC01
	objects.SharedAviGraphLister().Delete(modelSvc01)
	DelSVC(t, NAMESPACE, SHAREDVIPSVC01)
	DelEPS(t, NAMESPACE, SHAREDVIPSVC01)
	mcache := cache.SharedAviObjCache()
	vsKey := cache.NamespaceName{Namespace: AVINAMESPACE, Name: fmt.Sprintf("cluster--%s-%s", NAMESPACE, SHAREDVIPSVC01)}
	g.Eventually(func() bool {
		_, found := mcache.VsCacheMeta.AviCacheGet(vsKey)
		return found
	}, 5*time.Second).Should(gomega.Equal(false))

	modelSvc02 := MODEL_REDNS_PREFIX + SHAREDVIPSVC02
	objects.SharedAviGraphLister().Delete(modelSvc02)
	DelSVC(t, NAMESPACE, SHAREDVIPSVC02)
	DelEPS(t, NAMESPACE, SHAREDVIPSVC02)
	mcache = cache.SharedAviObjCache()
	vsKey = cache.NamespaceName{Namespace: AVINAMESPACE, Name: fmt.Sprintf("cluster--%s-%s", NAMESPACE, SHAREDVIPSVC02)}
	g.Eventually(func() bool {
		_, found := mcache.VsCacheMeta.AviCacheGet(vsKey)
		return found
	}, 5*time.Second).Should(gomega.Equal(false))
}

func SetUpTestForSharedVIPSvcLBWithExtDNS(t *testing.T, proto1, proto2 corev1.Protocol) {
	modelSvc01 := MODEL_REDNS_PREFIX + SHAREDVIPSVC01
	objects.SharedAviGraphLister().Delete(modelSvc01)
	svcObj := ConstructService(NAMESPACE, SHAREDVIPSVC01, proto1, corev1.ServiceTypeLoadBalancer, false, make(map[string]string), "")
	svcObj.Annotations = map[string]string{lib.SharedVipSvcLBAnnotation: SHAREDVIPKEY, lib.ExternalDNSAnnotation: EXTDNSANNOTATION}
	_, err := KubeClient.CoreV1().Services(NAMESPACE).Create(context.TODO(), svcObj, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Service: %v", err)
	}
	CreateEPS(t, NAMESPACE, SHAREDVIPSVC01, false, false, "1.1.1")
	PollForCompletion(t, modelSvc01, 5)

	modelSvc02 := MODEL_REDNS_PREFIX + SHAREDVIPSVC02
	objects.SharedAviGraphLister().Delete(modelSvc01)
	svcObj = ConstructService(NAMESPACE, SHAREDVIPSVC02, proto2, corev1.ServiceTypeLoadBalancer, false, make(map[string]string), "")
	svcObj.Annotations = map[string]string{lib.SharedVipSvcLBAnnotation: SHAREDVIPKEY}
	_, err = KubeClient.CoreV1().Services(NAMESPACE).Create(context.TODO(), svcObj, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Service: %v", err)
	}
	CreateEPS(t, NAMESPACE, SHAREDVIPSVC02, false, false, "2.1.1")
	PollForCompletion(t, modelSvc02, 5)
}

func VerfiyL4Node(nodes *avinodes.AviVsNode, g *gomega.GomegaWithT, proto1, proto2 string) {
	g.Expect(nodes.Name).To(gomega.Equal(fmt.Sprintf("cluster--%s-%s", NAMESPACE, SHAREDVIPKEY)))
	g.Expect(nodes.Tenant).To(gomega.Equal(AVINAMESPACE))
	g.Expect(nodes.PortProto[0].Port).To(gomega.Equal(int32(8080)))
	g.Expect(nodes.PortProto[0].Protocol).To(gomega.Equal(proto1))
	g.Expect(nodes.PortProto[1].Port).To(gomega.Equal(int32(8080)))
	g.Expect(nodes.PortProto[1].Protocol).To(gomega.Equal(proto2))
	g.Expect(nodes.PoolRefs).To(gomega.HaveLen(2))
	g.Expect(nodes.NetworkProfile).To(gomega.Equal(utils.MIXED_NET_PROFILE))
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
	os.Setenv("POD_NAME", "ako-0")
	os.Setenv("DEFAULT_LB_CONTROLLER", "true")
	os.Setenv("AKO_CRD_OPERATOR_ENABLED", "true")

	akoControlConfig := lib.AKOControlConfig()
	KubeClient = k8sfake.NewSimpleClientset()
	CRDClient = crdfake.NewSimpleClientset()
	v1alpha2CRDClient = v1alpha2crdfake.NewSimpleClientset()
	v1beta1CRDClient = v1beta1crdfake.NewSimpleClientset()
	akoControlConfig.SetCRDClientset(CRDClient)
	akoControlConfig.Setv1alpha2CRDClientset(v1alpha2CRDClient)
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

	registeredInformers = append(registeredInformers, utils.EndpointSlicesInformer)

	utils.NewInformers(utils.KubeClientIntf{ClientSet: KubeClient}, registeredInformers)

	// Initialize fake dynamic client for HealthMonitor support (required for L4Rules)
	scheme := runtime.NewScheme()
	// Register HealthMonitor CRD schema with the fake client
	scheme.AddKnownTypeWithName(
		schema.GroupVersionKind{Group: "ako.vmware.com", Version: "v1alpha1", Kind: "HealthMonitor"},
		&unstructured.Unstructured{},
	)
	scheme.AddKnownTypeWithName(
		schema.GroupVersionKind{Group: "ako.vmware.com", Version: "v1alpha1", Kind: "HealthMonitorList"},
		&unstructured.UnstructuredList{},
	)
	dynamicClient := dynamicfake.NewSimpleDynamicClient(scheme)
	lib.SetDynamicClientSet(dynamicClient)

	informers := k8s.K8sinformers{Cs: KubeClient, DynamicClient: dynamicClient}
	k8s.NewCRDInformers()

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
	AddDefaultNamespace()
	AddDefaultNamespace(NAMESPACE)

	go ctrl.InitController(informers, registeredInformers, ctrlCh, stopCh, quickSyncCh, waitGroupMap)
	objNameMap.InitMap()
	os.Exit(m.Run())
}

func TestLBSvcWithLBClassWithAviDefaultLBController(t *testing.T) {
	// This test expects to start with defaultLBController = true
	g := gomega.NewWithT(t)

	svcName := objNameMap.GenerateName(SINGLEPORTSVC)
	modelName := MODEL_REDNS_PREFIX + svcName

	// test invalid service spec.LoadBalancerClass != ako.vmware.com/avi-lb
	SetUpTestForSvcLBWithLBClass(t, INVALID_LB_CLASS, svcName)
	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 10*time.Second).Should(gomega.Equal(false))
	TearDownTestForSvcLB(t, g, svcName)

	// test valid service with spec.LoadBalancerClass == ako.vmware.com/avi-lb
	SetUpTestForSvcLBWithLBClass(t, "ako.vmware.com/avi-lb", svcName)
	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 10*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].Name).To(gomega.Equal(fmt.Sprintf("cluster--%s-%s", NAMESPACE, svcName)))
	g.Expect(nodes[0].Tenant).To(gomega.Equal(AVINAMESPACE))
	g.Expect(nodes[0].PortProto[0].Port).To(gomega.Equal(int32(8080)))
	g.Expect(nodes[0].PoolRefs).To(gomega.HaveLen(1))
	address := "1.1.1.1"
	g.Expect(nodes[0].PoolRefs[0].Servers[0].Ip.Addr).To(gomega.Equal(&address))
	TearDownTestForSvcLB(t, g, svcName)

	// test valid service with spec.LoadBalancerClass == "" (unpopulated)
	SetUpTestForSvcLB(t, svcName)
	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 10*time.Second).Should(gomega.Equal(true))
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].Name).To(gomega.Equal(fmt.Sprintf("cluster--%s-%s", NAMESPACE, svcName)))
	g.Expect(nodes[0].Tenant).To(gomega.Equal(AVINAMESPACE))
	g.Expect(nodes[0].PortProto[0].Port).To(gomega.Equal(int32(8080)))
	g.Expect(nodes[0].PoolRefs).To(gomega.HaveLen(1))
	g.Expect(nodes[0].PoolRefs[0].Servers[0].Ip.Addr).To(gomega.Equal(&address))
	TearDownTestForSvcLB(t, g, svcName)
}

func TestLBSvcWithLBClassWithoutAviDefaultLBController(t *testing.T) {
	g := gomega.NewWithT(t)
	svcName := objNameMap.GenerateName(SINGLEPORTSVC)
	modelName := MODEL_REDNS_PREFIX + svcName
	lib.AKOControlConfig().SetDefaultLBController(false)

	// test invalid service spec.LoadBalancerClass != ako.vmware.com/avi-lb
	SetUpTestForSvcLBWithLBClass(t, INVALID_LB_CLASS, svcName)
	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 10*time.Second).Should(gomega.Equal(false))
	TearDownTestForSvcLB(t, g, svcName)

	// test valid service with spec.LoadBalancerClass = ako.vmware.com/avi-lb
	SetUpTestForSvcLBWithLBClass(t, "ako.vmware.com/avi-lb", svcName)
	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 10*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].Name).To(gomega.Equal(fmt.Sprintf("cluster--%s-%s", NAMESPACE, svcName)))
	g.Expect(nodes[0].Tenant).To(gomega.Equal(AVINAMESPACE))
	g.Expect(nodes[0].PortProto[0].Port).To(gomega.Equal(int32(8080)))
	g.Expect(nodes[0].PoolRefs).To(gomega.HaveLen(1))
	address := "1.1.1.1"
	g.Expect(nodes[0].PoolRefs[0].Servers[0].Ip.Addr).To(gomega.Equal(&address))
	TearDownTestForSvcLB(t, g, svcName)

	// test invalid service with spec.LoadBalancerClass == "" (unpopulated)
	SetUpTestForSvcLB(t, svcName)
	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 10*time.Second).Should(gomega.Equal(false))
	TearDownTestForSvcLB(t, g, svcName)

	lib.AKOControlConfig().SetDefaultLBController(true)
}

func TestAviSvcCreationSinglePort(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	svcName := objNameMap.GenerateName(SINGLEPORTSVC)
	modelName := MODEL_REDNS_PREFIX + svcName
	SetUpTestForSvcLB(t, svcName)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 10*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].Name).To(gomega.Equal(fmt.Sprintf("cluster--%s-%s", NAMESPACE, svcName)))
	g.Expect(nodes[0].Tenant).To(gomega.Equal(AVINAMESPACE))
	g.Expect(nodes[0].PortProto[0].Port).To(gomega.Equal(int32(8080)))

	// Check for the pools
	g.Expect(nodes[0].PoolRefs).To(gomega.HaveLen(1))
	address := "1.1.1.1"
	g.Expect(nodes[0].PoolRefs[0].Servers[0].Ip.Addr).To(gomega.Equal(&address))

	// If we transition the service from Loadbalancer to ClusterIP - it should get deleted.
	svcExample := (FakeService{
		Name:         svcName,
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
	vsKey := cache.NamespaceName{Namespace: AVINAMESPACE, Name: fmt.Sprintf("cluster--%s-%s", NAMESPACE, svcName)}
	g.Eventually(func() bool {
		_, found := mcache.VsCacheMeta.AviCacheGet(vsKey)
		return found
	}, 15*time.Second).Should(gomega.Equal(false))
	// If we transition the service from clusterIP to Loadbalancer - vs should get ceated
	svcExample = (FakeService{
		Name:         svcName,
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
	TearDownTestForSvcLB(t, g, svcName)
}

func TestAviSvcCreationSinglePortMultiTenantEnabled(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetAkoTenant()
	defer ResetAkoTenant()
	svcName := objNameMap.GenerateName(SINGLEPORTSVC)
	modelName := fmt.Sprintf("%s/cluster--red-ns-%s", AKOTENANT, svcName)
	objects.SharedAviGraphLister().Delete(modelName)
	CreateSVC(t, NAMESPACE, svcName, corev1.ProtocolTCP, corev1.ServiceTypeLoadBalancer, false)
	CreateEPS(t, NAMESPACE, svcName, false, false, "1.1.1")

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
	g.Expect(nodes[0].Name).To(gomega.Equal(fmt.Sprintf("cluster--%s-%s", NAMESPACE, svcName)))
	// Tenant should be akotenant instead of admin
	g.Expect(nodes[0].Tenant).To(gomega.Equal(AKOTENANT))
	g.Expect(nodes[0].PortProto[0].Port).To(gomega.Equal(int32(8080)))
	g.Expect(*nodes[0].TrafficEnabled).To(gomega.Equal(true))

	// Check for the pools
	g.Expect(nodes[0].PoolRefs).To(gomega.HaveLen(1))
	address := "1.1.1.1"
	g.Expect(nodes[0].PoolRefs[0].Servers[0].Ip.Addr).To(gomega.Equal(&address))

	objects.SharedAviGraphLister().Delete(modelName)
	DelSVC(t, NAMESPACE, svcName)
	DelEPS(t, NAMESPACE, svcName)
	mcache := cache.SharedAviObjCache()
	vsKey := cache.NamespaceName{Namespace: AVINAMESPACE, Name: fmt.Sprintf("cluster--%s-%s", svcName, NAMESPACE)}
	g.Eventually(func() bool {
		_, found := mcache.VsCacheMeta.AviCacheGet(vsKey)
		return found
	}, 5*time.Second).Should(gomega.Equal(false))
}

func TestAviSvcCreationMultiPort(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	svcName := objNameMap.GenerateName(MULTIPORTSVC)
	modelName := fmt.Sprintf("%s/cluster--%s-%s", AVINAMESPACE, NAMESPACE, svcName)

	SetUpTestForSvcLBMultiport(t, svcName)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 10*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].Name).To(gomega.Equal(fmt.Sprintf("cluster--%s-%s", NAMESPACE, svcName)))
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

	TearDownTestForSvcLBMultiport(t, g, svcName)
}

func TestL4NamingConvention(t *testing.T) {
	// checks naming convention of all generated nodes
	g := gomega.NewGomegaWithT(t)
	svcName := objNameMap.GenerateName(MULTIPORTSVC)
	modelName := fmt.Sprintf("%s/cluster--%s-%s", AVINAMESPACE, NAMESPACE, svcName)

	SetUpTestForSvcLBMultiport(t, svcName)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 15*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes[0].Name).To(gomega.Equal("cluster--red-ns-" + svcName))
	g.Expect(nodes[0].PoolRefs[0].Name).To(gomega.ContainSubstring("cluster--red-ns-" + svcName + "-TCP-808"))
	g.Expect(nodes[0].VSVIPRefs[0].Name).To(gomega.Equal("cluster--red-ns-" + svcName))
	g.Expect(nodes[0].L4PolicyRefs[0].Name).To(gomega.Equal("cluster--red-ns-" + svcName))

	TearDownTestForSvcLBMultiport(t, g, svcName)
}

func TestAviSvcMultiPortApplicationProf(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	svcName := objNameMap.GenerateName(MULTIPORTSVC)
	modelName := fmt.Sprintf("%s/cluster--%s-%s", AVINAMESPACE, NAMESPACE, svcName)

	SetUpTestForSvcLBMultiport(t, svcName)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 10*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].Name).To(gomega.Equal(fmt.Sprintf("cluster--%s-%s", NAMESPACE, svcName)))
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

	TearDownTestForSvcLBMultiport(t, g, svcName)
}

func TestAviSvcUpdateEndpoint(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	svcName := objNameMap.GenerateName(SINGLEPORTSVC)
	modelName := fmt.Sprintf("%s/cluster--%s-%s", AVINAMESPACE, NAMESPACE, svcName)

	SetUpTestForSvcLB(t, svcName)

	// epExample := &corev1.Endpoints{
	// 	ObjectMeta: metav1.ObjectMeta{Namespace: NAMESPACE, Name:svcName},
	// 	Subsets: []corev1.EndpointSubset{{
	// 		Addresses: []corev1.EndpointAddress{{IP: "1.2.3.14"}, {IP: "1.2.3.24"}},
	// 		Ports:     []corev1.EndpointPort{{Name: "foo", Port: 8080, Protocol: "TCP"}},
	// 	}},
	// }
	// if _, err = KubeClient.CoreV1().Endpoints(NAMESPACE).Update(context.TODO(), epExample, metav1.UpdateOptions{}); err != nil {
	// 	t.Fatalf("Error in updating the Endpoint: %v", err)
	// }
	ScaleCreateEPS(t, NAMESPACE, svcName)

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
			address := "1.2.3.5"
			g.Expect(pool.Servers).To(gomega.HaveLen(2))
			g.Expect(pool.Servers[1].Ip.Addr).To(gomega.Equal(&address))
		} else {
			g.Expect(pool.Servers).To(gomega.HaveLen(0))
		}
	}

	TearDownTestForSvcLB(t, g, svcName)
}

// Rest Cache sync tests

func TestCreateServiceLBCacheSync(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	svcName := objNameMap.GenerateName(SINGLEPORTSVC)

	SetUpTestForSvcLB(t, svcName)

	mcache := cache.SharedAviObjCache()
	vsKey := cache.NamespaceName{Namespace: AVINAMESPACE, Name: fmt.Sprintf("cluster--%s-%s", NAMESPACE, svcName)}
	vsCache, found := mcache.VsCacheMeta.AviCacheGet(vsKey)
	if !found {
		t.Fatalf("Cache not found for VS: %v", vsKey)
	} else {
		vsCacheObj, ok := vsCache.(*cache.AviVsCache)
		if !ok {
			t.Fatalf("Invalid VS object. Cannot cast.")
		}
		g.Expect(vsCacheObj.Name).To(gomega.Equal(fmt.Sprintf("cluster--%s-%s", NAMESPACE, svcName)))
		g.Expect(vsCacheObj.Tenant).To(gomega.Equal(AVINAMESPACE))
		g.Expect(vsCacheObj.PoolKeyCollection).To(gomega.HaveLen(1))
		g.Expect(vsCacheObj.PoolKeyCollection[0].Name).To(gomega.MatchRegexp("cluster--red-ns-" + svcName + "-TCP-8080"))
		g.Expect(vsCacheObj.L4PolicyCollection).To(gomega.HaveLen(1))
		g.Expect(vsCacheObj.L4PolicyCollection[0].Name).To(gomega.MatchRegexp("cluster--red-ns-" + svcName))
	}

	TearDownTestForSvcLB(t, g, svcName)
	g.Eventually(func() bool {
		_, found := mcache.VsCacheMeta.AviCacheGet(vsKey)
		return found
	}, 10*time.Second).Should(gomega.Equal(false))
}

func TestCreateServiceLBWithFaultCacheSync(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	svcName := objNameMap.GenerateName(SINGLEPORTSVC)

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

	SetUpTestForSvcLB(t, svcName)

	mcache := cache.SharedAviObjCache()
	vsKey := cache.NamespaceName{Namespace: AVINAMESPACE, Name: fmt.Sprintf("cluster--%s-%s", NAMESPACE, svcName)}
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
		g.Expect(vsCacheObj.Name).To(gomega.Equal(fmt.Sprintf("cluster--%s-%s", NAMESPACE, svcName)))
		g.Expect(vsCacheObj.Tenant).To(gomega.Equal(AVINAMESPACE))
		g.Expect(vsCacheObj.PoolKeyCollection).To(gomega.HaveLen(1))
		g.Expect(vsCacheObj.PoolKeyCollection[0].Name).To(gomega.MatchRegexp("cluster--red-ns-" + svcName + "-TCP-8080"))
		g.Expect(vsCacheObj.L4PolicyCollection).To(gomega.HaveLen(1))
		g.Expect(vsCacheObj.L4PolicyCollection[0].Name).To(gomega.MatchRegexp("cluster--red-ns-" + svcName))
	}

	TearDownTestForSvcLB(t, g, svcName)
}

func TestCreateMultiportServiceLBCacheSync(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	svcName := objNameMap.GenerateName(MULTIPORTSVC)
	NAMESPACE, AVINAMESPACE := "red-ns", "admin"

	SetUpTestForSvcLBMultiport(t, svcName)

	mcache := cache.SharedAviObjCache()
	vsKey := cache.NamespaceName{Namespace: AVINAMESPACE, Name: fmt.Sprintf("cluster--%s-%s", NAMESPACE, svcName)}
	vsCache, found := mcache.VsCacheMeta.AviCacheGet(vsKey)
	if !found {
		t.Fatalf("Cache not found for VS: %v", vsKey)
	}
	vsCacheObj, ok := vsCache.(*cache.AviVsCache)
	if !ok {
		t.Fatalf("Invalid VS object. Cannot cast.")
	}
	g.Expect(vsCacheObj.Name).To(gomega.Equal(fmt.Sprintf("cluster--%s-%s", NAMESPACE, svcName)))
	g.Expect(vsCacheObj.Tenant).To(gomega.Equal(AVINAMESPACE))
	g.Expect(vsCacheObj.PoolKeyCollection).To(gomega.HaveLen(3))
	g.Expect(vsCacheObj.PoolKeyCollection[0].Name).To(gomega.MatchRegexp(`^(cluster--[a-zA-Z0-9-]+-808(0|1|2))$`))
	g.Expect(vsCacheObj.L4PolicyCollection).To(gomega.HaveLen(1))
	g.Expect(vsCacheObj.L4PolicyCollection[0].Name).To(gomega.MatchRegexp(`^(cluster--[a-zA-Z0-9-]+)$`))

	TearDownTestForSvcLBMultiport(t, g, svcName)
}

func TestUpdateAndDeleteServiceLBCacheSync(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	svcName := objNameMap.GenerateName(SINGLEPORTSVC)

	SetUpTestForSvcLB(t, svcName)

	// Get hold of the pool checksum on CREATE
	poolName := "cluster--red-ns-" + svcName + "-TCP-8080"
	mcache := cache.SharedAviObjCache()
	poolKey := cache.NamespaceName{Namespace: AVINAMESPACE, Name: poolName}
	poolCacheBefore, _ := mcache.PoolCache.AviCacheGet(poolKey)
	poolCacheBeforeObj, _ := poolCacheBefore.(*cache.AviPoolCache)
	oldPoolCksum := poolCacheBeforeObj.CloudConfigCksum

	// UPDATE Test: After Endpoint update, Cache checksums must change
	// epExample := &corev1.Endpoints{
	// 	ObjectMeta: metav1.ObjectMeta{Namespace: NAMESPACE, Name:svcName},
	// 	Subsets: []corev1.EndpointSubset{{
	// 		Addresses: []corev1.EndpointAddress{{IP: "1.2.3.14"}, {IP: "1.2.3.24"}},
	// 		Ports:     []corev1.EndpointPort{{Name: "foo", Port: 8080, Protocol: "TCP"}},
	// 	}},
	// }
	// if _, err = KubeClient.CoreV1().Endpoints(NAMESPACE).Update(context.TODO(), epExample, metav1.UpdateOptions{}); err != nil {
	// 	t.Fatalf("Error in updating the Endpoint: %v", err)
	// }
	ScaleCreateEPS(t, NAMESPACE, svcName)

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
	TearDownTestForSvcLB(t, g, svcName)
	g.Eventually(func() bool {
		_, found = mcache.PoolCache.AviCacheGet(poolKey)
		return found
	}, 5*time.Second).Should(gomega.Equal(false))
}

// TestScaleUpAndDownServiceLBCacheSync tests the avi node graph and rest layer functionality when the
// multiport serviceLB is increased from 1 to 5 and then decreased back to 1
func TestScaleUpAndDownServiceLBCacheSync(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	svcName := objNameMap.GenerateName(SINGLEPORTSVC)
	var model, service string

	// Simulate a delay of 200ms in the Avi API
	AddMiddleware(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		NormalControllerServer(w, r)
	})
	defer ResetMiddleware()

	SetUpTestForSvcLB(t, svcName)

	svcName2 := objNameMap.GenerateName(MULTIPORTSVC)
	// create numScale more multiport service of type loadbalancer
	numScale := 5
	for i := 0; i < numScale; i++ {
		service = fmt.Sprintf("%s%d", svcName2, i)
		model = MODEL_REDNS_PREFIX + service

		objects.SharedAviGraphLister().Delete(model)
		CreateSVC(t, NAMESPACE, service, corev1.ProtocolTCP, corev1.ServiceTypeLoadBalancer, true)
		CreateEPS(t, NAMESPACE, service, true, true, "1.1.1")
	}

	// verify that numScale services are created on the graph and corresponding cache objects
	var found bool
	var vsKey cache.NamespaceName
	var aviModel interface{}

	mcache := cache.SharedAviObjCache()
	for i := 0; i < numScale; i++ {
		service = fmt.Sprintf("%s%d", svcName2, i)
		model = MODEL_REDNS_PREFIX + service

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
		service = fmt.Sprintf("%s%d", svcName2, i)
		model = MODEL_REDNS_PREFIX + service
		objects.SharedAviGraphLister().Delete(model)
		DelSVC(t, NAMESPACE, service)
		DelEPS(t, NAMESPACE, service)
	}

	// verify that the graph nodes and corresponding cache are deleted for the numScale services
	for i := 0; i < numScale; i++ {
		service = fmt.Sprintf("%s%d", svcName2, i)
		model = MODEL_REDNS_PREFIX + service
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
	vsKey = cache.NamespaceName{Namespace: AVINAMESPACE, Name: fmt.Sprintf("cluster--%s-%s", NAMESPACE, svcName)}
	g.Eventually(func() bool {
		_, found = mcache.VsCacheMeta.AviCacheGet(vsKey)
		return found
	}, 40*time.Second).Should(gomega.Equal(true))
	TearDownTestForSvcLB(t, g, svcName)
}

func TestAviSvcCreationWithStaticIP(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	svcName := objNameMap.GenerateName(SINGLEPORTSVC)
	modelName := MODEL_REDNS_PREFIX + svcName

	staticIP := "80.80.80.80"
	objects.SharedAviGraphLister().Delete(modelName)
	svcExample := (FakeService{
		Name:           svcName,
		Namespace:      NAMESPACE,
		Type:           corev1.ServiceTypeLoadBalancer,
		LoadBalancerIP: staticIP,
		ServicePorts:   []Serviceport{{PortName: "foo1", Protocol: "TCP", PortNumber: 8080, TargetPort: intstr.FromInt(8080)}},
	}).Service()
	_, err := KubeClient.CoreV1().Services(NAMESPACE).Create(context.TODO(), svcExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in creating Service: %v", err)
	}
	CreateEPS(t, NAMESPACE, svcName, false, false, "1.1.1")
	PollForCompletion(t, modelName, 5)

	g.Eventually(func() string {
		if found, aviModel := objects.SharedAviGraphLister().Get(modelName); found && aviModel != nil {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
			if len(nodes) > 0 && len(nodes[0].VSVIPRefs) > 0 {
				return nodes[0].VSVIPRefs[0].IPAddress
			}
		}
		return ""
	}, 20*time.Second).Should(gomega.Equal(staticIP))
	TearDownTestForSvcLB(t, g, svcName)
}

// Infra CRD tests via service annotation

func TestWithInfraSettingStatusUpdates(t *testing.T) {
	// create infraSetting, svcLB with bad seGroup/networkName
	// check for Rejected status, check layer 2 for defaults
	// change to good seGroup/networkName, check for Accepted status
	// check layer 2 model

	g := gomega.NewGomegaWithT(t)
	settingName := "infra-setting"
	svcName := objNameMap.GenerateName(SINGLEPORTSVC)
	modelName := MODEL_REDNS_PREFIX + svcName

	objects.SharedAviGraphLister().Delete(modelName)
	svcExample := (FakeService{
		Name:         svcName,
		Namespace:    NAMESPACE,
		Type:         corev1.ServiceTypeLoadBalancer,
		ServicePorts: []Serviceport{{PortName: "foo1", Protocol: "TCP", PortNumber: 8080, TargetPort: intstr.FromInt(8080)}},
	}).Service()
	svcExample.Annotations = map[string]string{lib.InfraSettingNameAnnotation: settingName}
	_, err := KubeClient.CoreV1().Services(NAMESPACE).Create(context.TODO(), svcExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in creating Service: %v", err)
	}
	CreateEPS(t, NAMESPACE, svcName, false, false, "1.1.1")
	PollForCompletion(t, modelName, 5)

	// Create with bad seGroup ref.
	settingCreate := (FakeAviInfraSetting{
		Name:        settingName,
		SeGroupName: "thisisBADaviref-seGroup",
		Networks:    []string{"thisisaviref-networkName"},
		EnableRhi:   true,
	}).AviInfraSetting()
	if _, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().AviInfraSettings().Create(context.TODO(), settingCreate, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding AviInfraSetting: %v", err)
	}

	g.Eventually(func() string {
		setting, _ := v1beta1CRDClient.AkoV1beta1().AviInfraSettings().Get(context.TODO(), settingName, metav1.GetOptions{})
		return setting.Status.Status
	}, 15*time.Second).Should(gomega.Equal("Rejected"))

	// defaults to global seGroup and networkName.
	netList := utils.GetVipNetworkList()
	g.Eventually(func() bool {
		if found, aviModel := objects.SharedAviGraphLister().Get(modelName); found && aviModel != nil {
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
	if _, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().AviInfraSettings().Update(context.TODO(), settingUpdate, metav1.UpdateOptions{}); err != nil {
		t.Fatalf("error in updating AviInfraSetting: %v", err)
	}

	g.Eventually(func() string {
		setting, _ := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().AviInfraSettings().Get(context.TODO(), settingName, metav1.GetOptions{})
		return setting.Status.Status
	}, 15*time.Second).Should(gomega.Equal("Accepted"))

	g.Eventually(func() bool {
		if found, aviModel := objects.SharedAviGraphLister().Get(modelName); found && aviModel != nil {
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
	if _, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().AviInfraSettings().Update(context.TODO(), settingUpdate, metav1.UpdateOptions{}); err != nil {
		t.Fatalf("error in updating AviInfraSetting: %v", err)
	}

	g.Eventually(func() string {
		setting, _ := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().AviInfraSettings().Get(context.TODO(), settingName, metav1.GetOptions{})
		return setting.Status.Status
	}, 15*time.Second).Should(gomega.Equal("Rejected"))
	g.Eventually(func() bool {
		if found, aviModel := objects.SharedAviGraphLister().Get(modelName); found && aviModel != nil {
			if nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS(); len(nodes) > 0 {
				return nodes[0].ServiceEngineGroup == lib.GetSEGName() &&
					len(nodes[0].VSVIPRefs[0].VipNetworks) > 0 &&
					nodes[0].VSVIPRefs[0].VipNetworks[0].NetworkName == netList[0].NetworkName &&
					!*nodes[0].EnableRhi
			}
		}
		return false
	}, 45*time.Second).Should(gomega.Equal(true))

	settingUpdate = (FakeAviInfraSetting{
		Name:        settingName,
		SeGroupName: "thisisaviref-seGroup",
		Networks:    []string{"multivip-network1", "multivip-network2", "multivip-network3"},
		EnableRhi:   true,
	}).AviInfraSetting()
	settingUpdate.ResourceVersion = "4"
	if _, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().AviInfraSettings().Update(context.TODO(), settingUpdate, metav1.UpdateOptions{}); err != nil {
		t.Fatalf("error in updating AviInfraSetting: %v", err)
	}

	g.Eventually(func() string {
		setting, _ := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().AviInfraSettings().Get(context.TODO(), settingName, metav1.GetOptions{})
		return setting.Status.Status
	}, 15*time.Second).Should(gomega.Equal("Accepted"))

	g.Eventually(func() bool {
		if found, aviModel := objects.SharedAviGraphLister().Get(modelName); found && aviModel != nil {
			if nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS(); len(nodes) > 0 {
				if len(nodes[0].VSVIPRefs[0].VipNetworks) == 3 &&
					nodes[0].VSVIPRefs[0].VipNetworks[0].NetworkName == "multivip-network1" &&
					nodes[0].VSVIPRefs[0].VipNetworks[1].NetworkName == "multivip-network2" &&
					nodes[0].VSVIPRefs[0].VipNetworks[2].NetworkName == "multivip-network3" &&
					*nodes[0].EnableRhi == true &&
					nodes[0].ServiceEngineGroup == "thisisaviref-seGroup" {
					return true
				}
			}
		}
		return false
	}, 35*time.Second).Should(gomega.Equal(true))

	TeardownAviInfraSetting(t, settingName)
	TearDownTestForSvcLB(t, g, svcName)
}

func TestInfraSettingDelete(t *testing.T) {
	// create infraSetting, svcLB
	// delete infraSetting, fallback to defaults

	g := gomega.NewGomegaWithT(t)
	settingName := "infra-setting"
	svcName := objNameMap.GenerateName(SINGLEPORTSVC)
	modelName := MODEL_REDNS_PREFIX + svcName

	objects.SharedAviGraphLister().Delete(modelName)
	svcExample := (FakeService{
		Name:         svcName,
		Namespace:    NAMESPACE,
		Type:         corev1.ServiceTypeLoadBalancer,
		ServicePorts: []Serviceport{{PortName: "foo1", Protocol: "TCP", PortNumber: 8080, TargetPort: intstr.FromInt(8080)}},
	}).Service()
	svcExample.Annotations = map[string]string{lib.InfraSettingNameAnnotation: settingName}
	_, err := KubeClient.CoreV1().Services(NAMESPACE).Create(context.TODO(), svcExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in creating Service: %v", err)
	}
	CreateEPS(t, NAMESPACE, svcName, false, false, "1.1.1")
	PollForCompletion(t, modelName, 5)

	SetupAviInfraSetting(t, settingName, "")

	g.Eventually(func() string {
		if found, aviModel := objects.SharedAviGraphLister().Get(modelName); found && aviModel != nil {
			if nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS(); len(nodes) > 0 {
				return nodes[0].ServiceEngineGroup
			}
		}
		return ""
	}, 35*time.Second).Should(gomega.Equal("thisisaviref-" + settingName + "-seGroup"))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes[0].VSVIPRefs[0].VipNetworks[0].NetworkName).Should(gomega.Equal("thisisaviref-" + settingName + "-networkName"))
	g.Expect(*nodes[0].EnableRhi).Should(gomega.Equal(true))

	TeardownAviInfraSetting(t, settingName)

	// defaults to global seGroup and networkName.
	netList := utils.GetVipNetworkList()
	g.Eventually(func() bool {
		if found, aviModel := objects.SharedAviGraphLister().Get(modelName); found && aviModel != nil {
			if nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS(); len(nodes) > 0 {
				return nodes[0].ServiceEngineGroup == lib.GetSEGName() &&
					len(nodes[0].VSVIPRefs[0].VipNetworks) > 0 &&
					nodes[0].VSVIPRefs[0].VipNetworks[0].NetworkName == netList[0].NetworkName &&
					!*nodes[0].EnableRhi
			}
		}
		return true
	}, 20*time.Second).Should(gomega.Equal(true))

	TearDownTestForSvcLB(t, g, svcName)
}

func TestInfraSettingChangeMapping(t *testing.T) {
	// create 2 infraSettings, svcLB
	// update infraSetting from one to another in service annotation
	// check changed model

	g := gomega.NewGomegaWithT(t)

	settingName1, settingName2 := "infra-setting1", "infra-setting2"
	svcName := objNameMap.GenerateName(SINGLEPORTSVC)
	modelName := MODEL_REDNS_PREFIX + svcName

	objects.SharedAviGraphLister().Delete(modelName)

	svcExample := (FakeService{
		Name:         svcName,
		Namespace:    NAMESPACE,
		Type:         corev1.ServiceTypeLoadBalancer,
		ServicePorts: []Serviceport{{PortName: "foo1", Protocol: "TCP", PortNumber: 8080, TargetPort: intstr.FromInt(8080)}},
	}).Service()
	svcExample.Annotations = map[string]string{lib.InfraSettingNameAnnotation: settingName1}
	_, err := KubeClient.CoreV1().Services(NAMESPACE).Create(context.TODO(), svcExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in creating Service: %v", err)
	}
	CreateEPS(t, NAMESPACE, svcName, false, false, "1.1.1")
	PollForCompletion(t, modelName, 5)

	SetupAviInfraSetting(t, settingName1, "")
	SetupAviInfraSetting(t, settingName2, "")

	g.Eventually(func() string {
		if found, aviModel := objects.SharedAviGraphLister().Get(modelName); found && aviModel != nil {
			if nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS(); len(nodes) > 0 {
				return nodes[0].ServiceEngineGroup
			}
		}
		return ""
	}, 35*time.Second).Should(gomega.Equal("thisisaviref-" + settingName1 + "-seGroup"))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes[0].VSVIPRefs[0].VipNetworks[0].NetworkName).Should(gomega.Equal("thisisaviref-" + settingName1 + "-networkName"))

	// TODO: Change service annotation to have infraSettting2
	svcUpdate := (FakeService{
		Name:         svcName,
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
		if found, aviModel := objects.SharedAviGraphLister().Get(modelName); found && aviModel != nil {
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
	TearDownTestForSvcLB(t, g, svcName)
}
func TestSharedVIPSvcWithTCPUDPProtocols(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName := MODEL_REDNS_PREFIX + SHAREDVIPKEY

	SetUpTestForSharedVIPSvcLB(t, corev1.ProtocolTCP, corev1.ProtocolUDP)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 30*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes).To(gomega.HaveLen(1))
	VerfiyL4Node(nodes[0], g, "TCP", "UDP")
	TearDownTestForSharedVIPSvcLB(t, g)
}

func TestSharedVIPSvcTransitionSingle(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName := MODEL_REDNS_PREFIX + SHAREDVIPKEY

	SetUpTestForSharedVIPSvcLB(t, corev1.ProtocolTCP, corev1.ProtocolUDP)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 30*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes).To(gomega.HaveLen(1))
	VerfiyL4Node(nodes[0], g, "TCP", "UDP")

	// Initiating transition for one shared vip LB svc to type ClusterIP so the corresponfing pool and l4policyset should be deleted
	svcObj := ConstructService(NAMESPACE, SHAREDVIPSVC01, corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false, make(map[string]string), "")
	svcObj.ResourceVersion = "2"
	svcObj.Annotations = map[string]string{lib.SharedVipSvcLBAnnotation: SHAREDVIPKEY}
	_, err := KubeClient.CoreV1().Services(NAMESPACE).Update(context.TODO(), svcObj, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in adding Service: %v", err)
	}

	g.Eventually(func() int {
		if found, aviModel := objects.SharedAviGraphLister().Get(modelName); found {
			nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
			return len(nodes[0].PoolRefs)
		}
		return -1
	}, 30*time.Second).Should(gomega.Equal(1))
	g.Expect(nodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].Name).To(gomega.Equal(fmt.Sprintf("cluster--%s-%s", NAMESPACE, SHAREDVIPKEY)))
	g.Expect(nodes[0].Tenant).To(gomega.Equal(AVINAMESPACE))
	g.Expect(nodes[0].PortProto[0].Port).To(gomega.Equal(int32(8080)))
	g.Expect(nodes[0].PoolRefs).To(gomega.HaveLen(1))
	g.Expect(nodes[0].PortProto[0].Protocol).To(gomega.Equal("UDP"))
	g.Expect(nodes[0].NetworkProfile).To(gomega.Equal(utils.SYSTEM_UDP_FAST_PATH))

	// Initiating transition for same shared vip ClusterIP svc back to LB so the corresponfing pool and l4policyset should be re-created
	svcObj = ConstructService(NAMESPACE, SHAREDVIPSVC01, corev1.ProtocolTCP, corev1.ServiceTypeLoadBalancer, false, make(map[string]string), "")
	svcObj.ResourceVersion = "3"
	svcObj.Annotations = map[string]string{lib.SharedVipSvcLBAnnotation: SHAREDVIPKEY}
	_, err = KubeClient.CoreV1().Services(NAMESPACE).Update(context.TODO(), svcObj, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in adding Service: %v", err)
	}

	g.Eventually(func() int {
		if found, aviModel := objects.SharedAviGraphLister().Get(modelName); found {
			nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
			return len(nodes[0].PoolRefs)
		}
		return -1
	}, 30*time.Second).Should(gomega.Equal(2))
	VerfiyL4Node(nodes[0], g, "TCP", "UDP")

	TearDownTestForSharedVIPSvcLB(t, g)
}

func TestSharedVIPSvcTransitionAll(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName := MODEL_REDNS_PREFIX + SHAREDVIPKEY

	SetUpTestForSharedVIPSvcLB(t, corev1.ProtocolTCP, corev1.ProtocolUDP)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 30*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes).To(gomega.HaveLen(1))
	VerfiyL4Node(nodes[0], g, "TCP", "UDP")

	// Initiating transition for one shared vip LB svc to type ClusterIP so the corresponfing pool and l4policyset should be deleted
	svcObj := ConstructService(NAMESPACE, SHAREDVIPSVC01, corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false, make(map[string]string), "")
	svcObj.ResourceVersion = "2"
	svcObj.Annotations = map[string]string{lib.SharedVipSvcLBAnnotation: SHAREDVIPKEY}
	_, err := KubeClient.CoreV1().Services(NAMESPACE).Update(context.TODO(), svcObj, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in adding Service: %v", err)
	}

	g.Eventually(func() int {
		if found, aviModel := objects.SharedAviGraphLister().Get(modelName); found {
			nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
			return len(nodes[0].PoolRefs)
		}
		return -1
	}, 30*time.Second).Should(gomega.Equal(1))
	g.Expect(nodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].Name).To(gomega.Equal(fmt.Sprintf("cluster--%s-%s", NAMESPACE, SHAREDVIPKEY)))
	g.Expect(nodes[0].Tenant).To(gomega.Equal(AVINAMESPACE))
	g.Expect(nodes[0].PortProto[0].Port).To(gomega.Equal(int32(8080)))
	g.Expect(nodes[0].PoolRefs).To(gomega.HaveLen(1))
	g.Expect(nodes[0].PortProto[0].Protocol).To(gomega.Equal("UDP"))
	g.Expect(nodes[0].NetworkProfile).To(gomega.Equal(utils.SYSTEM_UDP_FAST_PATH))

	// Initiating transition for second shared vip LB svc to type ClusterIP so now the vs and all other configs should be deleted
	svcObj = ConstructService(NAMESPACE, SHAREDVIPSVC02, corev1.ProtocolUDP, corev1.ServiceTypeClusterIP, false, make(map[string]string), "")
	svcObj.ResourceVersion = "2"
	svcObj.Annotations = map[string]string{lib.SharedVipSvcLBAnnotation: SHAREDVIPKEY}
	_, err = KubeClient.CoreV1().Services(NAMESPACE).Update(context.TODO(), svcObj, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in adding Service: %v", err)
	}

	mcache := cache.SharedAviObjCache()
	vsKey := cache.NamespaceName{Namespace: AVINAMESPACE, Name: fmt.Sprintf("cluster--%s-%s", NAMESPACE, SHAREDVIPKEY)}
	g.Eventually(func() bool {
		_, found := mcache.VsCacheMeta.AviCacheGet(vsKey)
		return found
	}, 15*time.Second).Should(gomega.Equal(false))

	// Initiating transition for first shared vip ClusterIP svc back to LB so the corresponfing pool and l4policyset should be re-created
	svcObj = ConstructService(NAMESPACE, SHAREDVIPSVC01, corev1.ProtocolTCP, corev1.ServiceTypeLoadBalancer, false, make(map[string]string), "")
	svcObj.ResourceVersion = "3"
	svcObj.Annotations = map[string]string{lib.SharedVipSvcLBAnnotation: SHAREDVIPKEY}
	_, err = KubeClient.CoreV1().Services(NAMESPACE).Update(context.TODO(), svcObj, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in adding Service: %v", err)
	}
	// adding sleep of 2 seconds for vs node to get added to model
	time.Sleep(2 * time.Second)
	g.Eventually(func() int {
		if found, aviModel := objects.SharedAviGraphLister().Get(modelName); found {
			nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
			return len(nodes[0].PoolRefs)
		}
		return -1
	}, 30*time.Second).Should(gomega.Equal(1))
	g.Expect(nodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].Name).To(gomega.Equal(fmt.Sprintf("cluster--%s-%s", NAMESPACE, SHAREDVIPKEY)))
	g.Expect(nodes[0].Tenant).To(gomega.Equal(AVINAMESPACE))
	g.Expect(nodes[0].PortProto[0].Port).To(gomega.Equal(int32(8080)))
	g.Expect(nodes[0].PoolRefs).To(gomega.HaveLen(1))
	g.Expect(nodes[0].PortProto[0].Protocol).To(gomega.Equal("TCP"))
	g.Expect(nodes[0].NetworkProfile).To(gomega.Equal(utils.TCP_NW_FAST_PATH))

	// Initiating transition for second shared vip ClusterIP svc back to LB so the corresponfing pool and l4policyset should be re-created
	svcObj = ConstructService(NAMESPACE, SHAREDVIPSVC02, corev1.ProtocolUDP, corev1.ServiceTypeLoadBalancer, false, make(map[string]string), "")
	svcObj.ResourceVersion = "3"
	svcObj.Annotations = map[string]string{lib.SharedVipSvcLBAnnotation: SHAREDVIPKEY}
	_, err = KubeClient.CoreV1().Services(NAMESPACE).Update(context.TODO(), svcObj, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in adding Service: %v", err)
	}
	g.Eventually(func() int {
		if found, aviModel := objects.SharedAviGraphLister().Get(modelName); found {
			nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
			return len(nodes[0].PoolRefs)
		}
		return -1
	}, 30*time.Second).Should(gomega.Equal(2))
	VerfiyL4Node(nodes[0], g, "TCP", "UDP")

	TearDownTestForSharedVIPSvcLB(t, g)
}

func TestSharedVIPSvcWithTCPUDPProtocolsWithNSXT(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	// simulate nsx-t cloud by setting up t1lr
	os.Setenv("NSXT_T1_LR", "/infra/t1lr/sample")
	modelName := MODEL_REDNS_PREFIX + SHAREDVIPKEY

	SetUpTestForSharedVIPSvcLB(t, corev1.ProtocolTCP, corev1.ProtocolUDP)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 30*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes).To(gomega.HaveLen(1))
	VerfiyL4Node(nodes[0], g, "TCP", "UDP")

	// Check t1lr set at vsvip, pool. vrf empty at vsvip, pool and vs
	g.Expect(nodes[0].VSVIPRefs[0].T1Lr).Should(gomega.Equal("/infra/t1lr/sample"))
	g.Expect(nodes[0].VSVIPRefs[0].VrfContext).Should(gomega.Equal(""))
	g.Expect(nodes[0].PoolRefs[0].T1Lr).Should(gomega.Equal("/infra/t1lr/sample"))
	g.Expect(nodes[0].PoolRefs[0].VrfContext).Should(gomega.Equal(""))
	g.Expect(nodes[0].PoolRefs[1].T1Lr).Should(gomega.Equal("/infra/t1lr/sample"))
	g.Expect(nodes[0].PoolRefs[1].VrfContext).Should(gomega.Equal(""))
	g.Expect(nodes[0].VrfContext).Should(gomega.Equal(""))

	//reset the field
	os.Setenv("NSXT_T1_LR", "")
	TearDownTestForSharedVIPSvcLB(t, g)
}

func TestSharedVipSvcWithInvalidLBClass(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	modelName := MODEL_REDNS_PREFIX + SHAREDVIPKEY

	// create 1 valid and 1 invalid svc
	modelSvc01 := MODEL_REDNS_PREFIX + SHAREDVIPSVC01
	objects.SharedAviGraphLister().Delete(modelSvc01)
	svcObj := ConstructService(NAMESPACE, SHAREDVIPSVC01, corev1.ProtocolTCP, corev1.ServiceTypeLoadBalancer, false, make(map[string]string), lib.AviIngressController)
	svcObj.Annotations = map[string]string{lib.SharedVipSvcLBAnnotation: SHAREDVIPKEY}
	_, err := KubeClient.CoreV1().Services(NAMESPACE).Create(context.TODO(), svcObj, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Service: %v", err)
	}
	CreateEPS(t, NAMESPACE, SHAREDVIPSVC01, false, false, "1.1.1")
	PollForCompletion(t, modelSvc01, 5)

	// after valid svc creation VS should get created
	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 30*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes).To(gomega.HaveLen(1))

	modelSvc02 := MODEL_REDNS_PREFIX + SHAREDVIPSVC02
	objects.SharedAviGraphLister().Delete(modelSvc01)
	svcObj = ConstructService(NAMESPACE, SHAREDVIPSVC02, corev1.ProtocolSCTP, corev1.ServiceTypeLoadBalancer, false, make(map[string]string), INVALID_LB_CLASS)
	svcObj.Annotations = map[string]string{lib.SharedVipSvcLBAnnotation: SHAREDVIPKEY}
	_, err = KubeClient.CoreV1().Services(NAMESPACE).Create(context.TODO(), svcObj, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Service: %v", err)
	}
	CreateEPS(t, NAMESPACE, SHAREDVIPSVC02, false, false, "2.1.1")
	PollForCompletion(t, modelSvc02, 5)

	// after invalid svc is created, the model should get invalidated and be nil
	time.Sleep(5 * time.Second)
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	g.Expect(aviModel).To(gomega.BeNil())

	// after deletion of invalid svc, model should be valid again
	objects.SharedAviGraphLister().Delete(modelSvc02)
	DelSVC(t, NAMESPACE, SHAREDVIPSVC02)
	DelEPS(t, NAMESPACE, SHAREDVIPSVC02)
	mcache := cache.SharedAviObjCache()
	vsKey := cache.NamespaceName{Namespace: AVINAMESPACE, Name: fmt.Sprintf("cluster--%s-%s", NAMESPACE, SHAREDVIPSVC02)}
	g.Eventually(func() bool {
		_, found := mcache.VsCacheMeta.AviCacheGet(vsKey)
		return found
	}, 5*time.Second).Should(gomega.Equal(false))

	time.Sleep(5 * time.Second)
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes).To(gomega.HaveLen(1))

	// teardown valid svc
	objects.SharedAviGraphLister().Delete(modelSvc01)
	DelSVC(t, NAMESPACE, SHAREDVIPSVC01)
	DelEPS(t, NAMESPACE, SHAREDVIPSVC01)
	mcache = cache.SharedAviObjCache()
	vsKey = cache.NamespaceName{Namespace: AVINAMESPACE, Name: fmt.Sprintf("cluster--%s-%s", NAMESPACE, SHAREDVIPSVC01)}
	g.Eventually(func() bool {
		_, found := mcache.VsCacheMeta.AviCacheGet(vsKey)
		return found
	}, 5*time.Second).Should(gomega.Equal(false))
}
func TestSharedVIPSvcWithTCPSCTProtocols(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName := MODEL_REDNS_PREFIX + SHAREDVIPKEY

	SetUpTestForSharedVIPSvcLB(t, corev1.ProtocolTCP, corev1.ProtocolSCTP)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 30*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes).To(gomega.HaveLen(1))
	VerfiyL4Node(nodes[0], g, "SCTP", "TCP")
	TearDownTestForSharedVIPSvcLB(t, g)
}
func TestSharedVIPSvcWithUDPSCTProtocols(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	os.Setenv("AUTO_L4_FQDN", "default")
	modelName := MODEL_REDNS_PREFIX + SHAREDVIPKEY

	SetUpTestForSharedVIPSvcLB(t, corev1.ProtocolUDP, corev1.ProtocolSCTP)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 30*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes).To(gomega.HaveLen(1))
	VerfiyL4Node(nodes[0], g, "SCTP", "UDP")
	os.Setenv("AUTO_L4_FQDN", "disable")
	TearDownTestForSharedVIPSvcLB(t, g)
}

// this test checks if extDNS FQDN is being set properly when set alongside shared-vip annotation
func TestSvcExternalDNSWithSharedVIP(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	os.Setenv("AUTO_L4_FQDN", "default")
	modelName := MODEL_REDNS_PREFIX + SHAREDVIPKEY

	SetUpTestForSharedVIPSvcLBWithExtDNS(t, corev1.ProtocolUDP, corev1.ProtocolSCTP)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 30*time.Second).Should(gomega.Equal(true))

	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].ServiceMetadata.HostNames[0]).To(gomega.Equal(EXTDNSANNOTATION))
	os.Setenv("AUTO_L4_FQDN", "disable")
	TearDownTestForSharedVIPSvcLB(t, g)
}

// this test checks if extDNS FQDN is being set properly
func TestSvcExtDNSAddition(t *testing.T) {
	os.Setenv("AUTO_L4_FQDN", "default")

	g := gomega.NewGomegaWithT(t)
	SetUpTestForSvcLBWithExtDNS(t)

	modelSvcDNS01 := MODEL_REDNS_PREFIX + EXTDNSSVC

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelSvcDNS01)
		return found
	}, 10*time.Second).Should(gomega.Equal(true))

	_, aviModel := objects.SharedAviGraphLister().Get(modelSvcDNS01)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].ServiceMetadata.HostNames[0]).To(gomega.Equal(EXTDNSANNOTATION))
	os.Setenv("AUTO_L4_FQDN", "disable")
	TearDownTestForSvcLBWithExtDNS(t, g)
}

func TestLBSvcCreationMixedProtocol(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	svcName := objNameMap.GenerateName(SINGLEPORTSVC)
	modelName := MODEL_REDNS_PREFIX + svcName
	SetUpTestForSvcLBMixedProtocol(t, svcName, corev1.ProtocolTCP, corev1.ProtocolUDP, corev1.ProtocolSCTP)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 10*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].Name).To(gomega.Equal(fmt.Sprintf("cluster--%s-%s", NAMESPACE, svcName)))
	g.Expect(nodes[0].Tenant).To(gomega.Equal(AVINAMESPACE))
	g.Expect(nodes[0].PortProto[0].Port).To(gomega.Equal(int32(8080)))
	g.Expect(nodes[0].PortProto[0].Protocol).To(gomega.Equal("TCP"))
	g.Expect(nodes[0].PortProto[1].Port).To(gomega.Equal(int32(8080)))
	g.Expect(nodes[0].PortProto[1].Protocol).To(gomega.Equal("UDP"))
	g.Expect(nodes[0].PortProto[2].Port).To(gomega.Equal(int32(8080)))
	g.Expect(nodes[0].PortProto[2].Protocol).To(gomega.Equal("SCTP"))

	// Check for the pools
	g.Expect(nodes[0].PoolRefs).To(gomega.HaveLen(3))
	for _, node := range nodes[0].PoolRefs {
		if node.Port == 8080 {
			address := "1.1.1.1"
			g.Expect(node.Servers).To(gomega.HaveLen(1))
			g.Expect(node.Servers[0].Ip.Addr).To(gomega.Equal(&address))
		}
	}
	g.Expect(nodes[0].L4PolicyRefs).To(gomega.HaveLen(1))
	g.Expect(nodes[0].ApplicationProfile).To(gomega.Equal(utils.DEFAULT_L4_APP_PROFILE))
	g.Expect(nodes[0].NetworkProfile).To(gomega.Equal(utils.MIXED_NET_PROFILE))

	TearDownTestForSvcLB(t, g, svcName)
}

func TestLBSvcCreationSCTPTCP(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	svcName := objNameMap.GenerateName(SINGLEPORTSVC)

	// middleware verifies the application and network profiles attached to the VS
	AddMiddleware(func(w http.ResponseWriter, r *http.Request) {
		var resp map[string]interface{}
		var finalResponse []byte
		url := r.URL.EscapedPath()

		rModelName := ""
		if r.Method == http.MethodPost &&
			strings.Contains(url, "/api/virtualservice") {
			data, _ := io.ReadAll(r.Body)
			json.Unmarshal(data, &resp)

			g.Expect(resp["application_profile_ref"]).Should(gomega.HaveSuffix("System-L4-Application"))
			g.Expect(resp["network_profile_ref"]).Should(gomega.HaveSuffix("System-TCP-Fast-Path"))

			rModelName = "virtualservice"
			rName := resp["name"].(string)
			objURL := fmt.Sprintf("https://localhost/api/%s/%s-%s#%s", rModelName, rModelName, RANDOMUUID, rName)

			// adding additional 'uuid' and 'url' (read-only) fields in the response
			resp["url"] = objURL
			resp["uuid"] = fmt.Sprintf("%s-%s-%s", rModelName, rName, RANDOMUUID)
			finalResponse, _ = json.Marshal(resp)
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, string(finalResponse))
			return
		}
		NormalControllerServer(w, r)
	})

	SetUpTestForSvcLBMixedProtocol(t, svcName, corev1.ProtocolTCP, corev1.ProtocolSCTP)

	mcache := cache.SharedAviObjCache()
	vsKey := cache.NamespaceName{Namespace: AVINAMESPACE, Name: fmt.Sprintf("cluster--%s-%s", NAMESPACE, svcName)}
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
		g.Expect(vsCacheObj.Name).To(gomega.Equal(fmt.Sprintf("cluster--%s-%s", NAMESPACE, svcName)))
		g.Expect(vsCacheObj.Tenant).To(gomega.Equal(AVINAMESPACE))
		g.Expect(vsCacheObj.PoolKeyCollection).To(gomega.HaveLen(2))
		g.Expect(vsCacheObj.PoolKeyCollection[0].Name).To(gomega.MatchRegexp("cluster--red-ns-" + svcName + "-TCP-8080"))
		g.Expect(vsCacheObj.PoolKeyCollection[1].Name).To(gomega.MatchRegexp("cluster--red-ns-" + svcName + "-SCTP-8080"))
		g.Expect(vsCacheObj.L4PolicyCollection).To(gomega.HaveLen(1))
		g.Expect(vsCacheObj.L4PolicyCollection[0].Name).To(gomega.MatchRegexp("cluster--red-ns-" + svcName))
	}

	defer ResetMiddleware()

	TearDownTestForSvcLB(t, g, svcName)
}

func TestLBSvcCreationSCTPUDP(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	svcName := objNameMap.GenerateName(SINGLEPORTSVC)

	// middleware verifies the application and network profiles attached to the VS
	AddMiddleware(func(w http.ResponseWriter, r *http.Request) {
		var resp map[string]interface{}
		var finalResponse []byte
		url := r.URL.EscapedPath()

		rModelName := ""
		if r.Method == http.MethodPost &&
			strings.Contains(url, "/api/virtualservice") {
			data, _ := io.ReadAll(r.Body)
			json.Unmarshal(data, &resp)

			g.Expect(resp["application_profile_ref"]).Should(gomega.HaveSuffix("System-L4-Application"))
			g.Expect(resp["network_profile_ref"]).Should(gomega.HaveSuffix("System-UDP-Fast-Path"))

			rModelName = "virtualservice"
			rName := resp["name"].(string)
			objURL := fmt.Sprintf("https://localhost/api/%s/%s-%s#%s", rModelName, rModelName, RANDOMUUID, rName)

			// adding additional 'uuid' and 'url' (read-only) fields in the response
			resp["url"] = objURL
			resp["uuid"] = fmt.Sprintf("%s-%s-%s", rModelName, rName, RANDOMUUID)
			finalResponse, _ = json.Marshal(resp)
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, string(finalResponse))
			return
		}
		NormalControllerServer(w, r)
	})

	SetUpTestForSvcLBMixedProtocol(t, svcName, corev1.ProtocolUDP, corev1.ProtocolSCTP)

	mcache := cache.SharedAviObjCache()
	vsKey := cache.NamespaceName{Namespace: AVINAMESPACE, Name: fmt.Sprintf("cluster--%s-%s", NAMESPACE, svcName)}
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
		g.Expect(vsCacheObj.Name).To(gomega.Equal(fmt.Sprintf("cluster--%s-%s", NAMESPACE, svcName)))
		g.Expect(vsCacheObj.Tenant).To(gomega.Equal(AVINAMESPACE))
		g.Expect(vsCacheObj.PoolKeyCollection).To(gomega.HaveLen(2))
		g.Expect(vsCacheObj.PoolKeyCollection[0].Name).To(gomega.MatchRegexp("cluster--red-ns-" + svcName + "-UDP-8080"))
		g.Expect(vsCacheObj.PoolKeyCollection[1].Name).To(gomega.MatchRegexp("cluster--red-ns-" + svcName + "-SCTP-8080"))
		g.Expect(vsCacheObj.L4PolicyCollection).To(gomega.HaveLen(1))
		g.Expect(vsCacheObj.L4PolicyCollection[0].Name).To(gomega.MatchRegexp("cluster--red-ns-" + svcName))
	}

	defer ResetMiddleware()

	TearDownTestForSvcLB(t, g, svcName)
}

func TestLBSvcCreationTCPUDP(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	svcName := objNameMap.GenerateName(SINGLEPORTSVC)

	// middleware verifies the application and network profiles attached to the VS
	AddMiddleware(func(w http.ResponseWriter, r *http.Request) {
		var resp map[string]interface{}
		var finalResponse []byte
		url := r.URL.EscapedPath()

		rModelName := ""
		if r.Method == http.MethodPost &&
			strings.Contains(url, "/api/virtualservice") {
			data, _ := io.ReadAll(r.Body)
			json.Unmarshal(data, &resp)

			g.Expect(resp["application_profile_ref"]).Should(gomega.HaveSuffix("System-L4-Application"))
			g.Expect(resp["network_profile_ref"]).Should(gomega.HaveSuffix("System-TCP-Fast-Path"))

			rModelName = "virtualservice"
			rName := resp["name"].(string)
			objURL := fmt.Sprintf("https://localhost/api/%s/%s-%s#%s", rModelName, rModelName, RANDOMUUID, rName)

			// adding additional 'uuid' and 'url' (read-only) fields in the response
			resp["url"] = objURL
			resp["uuid"] = fmt.Sprintf("%s-%s-%s", rModelName, rName, RANDOMUUID)
			finalResponse, _ = json.Marshal(resp)
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, string(finalResponse))
			return
		}
		NormalControllerServer(w, r)
	})

	SetUpTestForSvcLBMixedProtocol(t, svcName, corev1.ProtocolTCP, corev1.ProtocolUDP)

	mcache := cache.SharedAviObjCache()
	vsKey := cache.NamespaceName{Namespace: AVINAMESPACE, Name: fmt.Sprintf("cluster--%s-%s", NAMESPACE, svcName)}
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
		g.Expect(vsCacheObj.Name).To(gomega.Equal(fmt.Sprintf("cluster--%s-%s", NAMESPACE, svcName)))
		g.Expect(vsCacheObj.Tenant).To(gomega.Equal(AVINAMESPACE))
		g.Expect(vsCacheObj.PoolKeyCollection).To(gomega.HaveLen(2))
		g.Expect(vsCacheObj.PoolKeyCollection[0].Name).To(gomega.MatchRegexp("cluster--red-ns-" + svcName + "-TCP-8080"))
		g.Expect(vsCacheObj.PoolKeyCollection[1].Name).To(gomega.MatchRegexp("cluster--red-ns-" + svcName + "-UDP-8080"))
		g.Expect(vsCacheObj.L4PolicyCollection).To(gomega.HaveLen(1))
		g.Expect(vsCacheObj.L4PolicyCollection[0].Name).To(gomega.MatchRegexp("cluster--red-ns-" + svcName))
	}

	defer ResetMiddleware()

	TearDownTestForSvcLB(t, g, svcName)
}

func TestLBSvcWithAutoFQDNAsFlat(t *testing.T) {
	os.Setenv("AUTO_L4_FQDN", "flat")

	svcName := "service-01"
	svcNamespace := "red-ns"

	g := gomega.NewGomegaWithT(t)
	modelName := "admin/cluster--" + svcNamespace + "-" + svcName
	objects.SharedAviGraphLister().Delete(modelName)
	svcObj := ConstructService(svcNamespace, svcName, corev1.ProtocolTCP, corev1.ServiceTypeLoadBalancer, false, make(map[string]string), "")
	_, err := KubeClient.CoreV1().Services(svcNamespace).Create(context.TODO(), svcObj, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Service: %v", err)
	}
	CreateEPS(t, svcNamespace, svcName, false, false, "1.1.1")
	PollForCompletion(t, modelName, 5)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 10*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].Name).To(gomega.Equal(fmt.Sprintf("cluster--%s-%s", svcNamespace, svcName)))
	g.Expect(nodes[0].Tenant).To(gomega.Equal(AVINAMESPACE))
	g.Expect(nodes[0].PortProto[0].Port).To(gomega.Equal(int32(8080)))
	g.Expect(nodes[0].VSVIPRefs[0].FQDNs).To(gomega.HaveLen(1))
	g.Expect(nodes[0].VSVIPRefs[0].FQDNs[0]).To(gomega.HaveLen(len(svcName) + 1 + len(svcNamespace) + len(".com")))

	DelSVC(t, svcNamespace, svcName)
	DelEPS(t, svcNamespace, svcName)
	os.Setenv("AUTO_L4_FQDN", "disable")
}

func TestLBSvcFQDNLengthValidation(t *testing.T) {
	os.Setenv("AUTO_L4_FQDN", "flat")

	svcName := "python-flask-consumer-api-poc"
	svcNamespace := "service-ascend2-bookings-bi-int-dev"

	g := gomega.NewGomegaWithT(t)
	modelName := "admin/cluster--" + svcNamespace + "-" + svcName
	objects.SharedAviGraphLister().Delete(modelName)
	svcObj := ConstructService(svcNamespace, svcName, corev1.ProtocolTCP, corev1.ServiceTypeLoadBalancer, false, make(map[string]string), "")
	_, err := KubeClient.CoreV1().Services(svcNamespace).Create(context.TODO(), svcObj, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Service: %v", err)
	}
	CreateEPS(t, svcNamespace, svcName, false, false, "1.1.1")
	PollForCompletion(t, modelName, 5)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 10*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].Name).To(gomega.Equal(fmt.Sprintf("cluster--%s-%s", svcNamespace, svcName)))
	g.Expect(nodes[0].Tenant).To(gomega.Equal(AVINAMESPACE))
	g.Expect(nodes[0].PortProto[0].Port).To(gomega.Equal(int32(8080)))
	g.Expect(nodes[0].VSVIPRefs[0].FQDNs).To(gomega.HaveLen(1))
	g.Expect(nodes[0].VSVIPRefs[0].FQDNs[0]).To(gomega.HaveLen(63 + len(".com")))

	DelSVC(t, svcNamespace, svcName)
	DelEPS(t, svcNamespace, svcName)
	os.Setenv("AUTO_L4_FQDN", "disable")
}

func TestLBSvcWithNameLen63(t *testing.T) {
	os.Setenv("AUTO_L4_FQDN", "flat")

	svcName := "service-0123456789012345678901234567890123456789012345678901234"
	svcNamespace := "red-ns"

	g := gomega.NewGomegaWithT(t)
	modelName := "admin/cluster--" + svcNamespace + "-" + svcName
	objects.SharedAviGraphLister().Delete(modelName)
	svcObj := ConstructService(svcNamespace, svcName, corev1.ProtocolTCP, corev1.ServiceTypeLoadBalancer, false, make(map[string]string), "")
	_, err := KubeClient.CoreV1().Services(svcNamespace).Create(context.TODO(), svcObj, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Service: %v", err)
	}
	CreateEPS(t, svcNamespace, svcName, false, false, "1.1.1")
	PollForCompletion(t, modelName, 5)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 10*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].Name).To(gomega.Equal(fmt.Sprintf("cluster--%s-%s", svcNamespace, svcName)))
	g.Expect(nodes[0].Tenant).To(gomega.Equal(AVINAMESPACE))
	g.Expect(nodes[0].PortProto[0].Port).To(gomega.Equal(int32(8080)))
	g.Expect(nodes[0].VSVIPRefs[0].FQDNs).To(gomega.HaveLen(1))
	g.Expect(nodes[0].VSVIPRefs[0].FQDNs[0]).To(gomega.HaveLen(63 + len(".com")))

	DelSVC(t, svcNamespace, svcName)
	DelEPS(t, svcNamespace, svcName)
	os.Setenv("AUTO_L4_FQDN", "disable")
}

func TestLBSvcWithNamespaceNameLen63(t *testing.T) {
	os.Setenv("AUTO_L4_FQDN", "flat")

	svcName := "service-01"
	svcNamespace := "red-ns-01234567890123456789012345678901234567890123456789012345"

	g := gomega.NewGomegaWithT(t)
	modelName := "admin/cluster--" + svcNamespace + "-" + svcName
	objects.SharedAviGraphLister().Delete(modelName)
	svcObj := ConstructService(svcNamespace, svcName, corev1.ProtocolTCP, corev1.ServiceTypeLoadBalancer, false, make(map[string]string), "")
	_, err := KubeClient.CoreV1().Services(svcNamespace).Create(context.TODO(), svcObj, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Service: %v", err)
	}
	CreateEPS(t, svcNamespace, svcName, false, false, "1.1.1")
	PollForCompletion(t, modelName, 5)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 10*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].Name).To(gomega.Equal(fmt.Sprintf("cluster--%s-%s", svcNamespace, svcName)))
	g.Expect(nodes[0].Tenant).To(gomega.Equal(AVINAMESPACE))
	g.Expect(nodes[0].PortProto[0].Port).To(gomega.Equal(int32(8080)))
	g.Expect(nodes[0].VSVIPRefs[0].FQDNs).To(gomega.HaveLen(1))
	g.Expect(nodes[0].VSVIPRefs[0].FQDNs[0]).To(gomega.HaveLen(63 + len(".com")))

	DelSVC(t, svcNamespace, svcName)
	DelEPS(t, svcNamespace, svcName)
	os.Setenv("AUTO_L4_FQDN", "disable")
}

func TestLBSvcWithNameLen63AndNamespaceNameLen63(t *testing.T) {
	os.Setenv("AUTO_L4_FQDN", "flat")

	svcName := "service-0123456789012345678901234567890123456789012345678901234"
	svcNamespace := "red-ns-012345678901234567890123456789012345-----01234567890123"

	g := gomega.NewGomegaWithT(t)
	modelName := "admin/cluster--" + svcNamespace + "-" + svcName
	objects.SharedAviGraphLister().Delete(modelName)
	svcObj := ConstructService(svcNamespace, svcName, corev1.ProtocolTCP, corev1.ServiceTypeLoadBalancer, false, make(map[string]string), "")
	_, err := KubeClient.CoreV1().Services(svcNamespace).Create(context.TODO(), svcObj, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Service: %v", err)
	}
	CreateEPS(t, svcNamespace, svcName, false, false, "1.1.1")
	PollForCompletion(t, modelName, 5)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 10*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].Name).To(gomega.Equal(fmt.Sprintf("cluster--%s-%s", svcNamespace, svcName)))
	g.Expect(nodes[0].Tenant).To(gomega.Equal(AVINAMESPACE))
	g.Expect(nodes[0].PortProto[0].Port).To(gomega.Equal(int32(8080)))
	g.Expect(nodes[0].VSVIPRefs[0].FQDNs).To(gomega.HaveLen(1))
	g.Expect(nodes[0].VSVIPRefs[0].FQDNs[0]).To(gomega.HaveLen(63 + len(".com")))
	g.Expect(nodes[0].VSVIPRefs[0].FQDNs[0]).To(gomega.HaveSuffix("red-ns-012345678901234567890123456789012345.com"))

	DelSVC(t, svcNamespace, svcName)
	DelEPS(t, svcNamespace, svcName)
	os.Setenv("AUTO_L4_FQDN", "disable")
}

func TestLBSvcWithExtDNSAndAutoFQDNAsFlat(t *testing.T) {
	os.Setenv("AUTO_L4_FQDN", "flat")

	g := gomega.NewGomegaWithT(t)
	SetUpTestForSvcLBWithExtDNS(t)

	modelName := MODEL_REDNS_PREFIX + EXTDNSSVC

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 10*time.Second).Should(gomega.Equal(true))

	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].ServiceMetadata.HostNames[0]).To(gomega.Equal(EXTDNSANNOTATION))
	g.Expect(nodes[0].VSVIPRefs[0].FQDNs).To(gomega.HaveLen(1))
	g.Expect(nodes[0].VSVIPRefs[0].FQDNs[0]).To(gomega.Equal(EXTDNSANNOTATION))

	// remove the external-dns annotation and verfiy the auto-generated fqdn
	svcExample := (FakeService{
		Name:         EXTDNSSVC,
		Namespace:    NAMESPACE,
		Type:         corev1.ServiceTypeLoadBalancer,
		ServicePorts: []Serviceport{{PortName: "foo1", Protocol: "TCP", PortNumber: 8080, TargetPort: intstr.FromInt(8080)}},
	}).Service()
	svcExample.ResourceVersion = "2"
	_, err := KubeClient.CoreV1().Services(NAMESPACE).Update(context.TODO(), svcExample, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in adding Service: %v", err)
	}

	g.Eventually(func() bool {
		_, aviModel = objects.SharedAviGraphLister().Get(modelName)
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return g.Expect(nodes).To(gomega.HaveLen(1)) &&
			g.Expect(nodes[0].VSVIPRefs[0].FQDNs).To(gomega.HaveLen(1)) &&
			g.Expect(nodes[0].ServiceMetadata.HostNames[0]).NotTo(gomega.Equal(EXTDNSANNOTATION)) &&
			g.Expect(nodes[0].VSVIPRefs[0].FQDNs[0]).NotTo(gomega.Equal(EXTDNSANNOTATION)) &&
			g.Expect(nodes[0].VSVIPRefs[0].FQDNs[0]).To(gomega.Equal(EXTDNSSVC+"-"+NAMESPACE+".com"))
	}, 30*time.Second).Should(gomega.BeTrue())

	TearDownTestForSvcLBWithExtDNS(t, g)
	os.Setenv("AUTO_L4_FQDN", "disable")
}
func TestAviSvcCreationSinglePortMultiTenantEnabledWithTrafficDisabledAnnotation(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetAkoTenant()
	defer ResetAkoTenant()
	svcName := objNameMap.GenerateName(SINGLEPORTSVC)
	modelName := fmt.Sprintf("%s/cluster--red-ns-%s", AKOTENANT, svcName)
	objects.SharedAviGraphLister().Delete(modelName)
	svcExample := (FakeService{
		Name:         svcName,
		Namespace:    NAMESPACE,
		Type:         corev1.ServiceTypeLoadBalancer,
		ServicePorts: []Serviceport{{PortName: "foo1", Protocol: "TCP", PortNumber: 8080, TargetPort: intstr.FromInt(8080)}},
	}).Service()
	// Disable Traffic
	svcExample.Annotations = map[string]string{lib.VSTrafficDisabled: "true"}
	_, err := KubeClient.CoreV1().Services(NAMESPACE).Create(context.TODO(), svcExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in creating Service: %v", err)
	}
	CreateEPS(t, NAMESPACE, svcName, false, false, "1.1.1")

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
	g.Expect(nodes[0].Name).To(gomega.Equal(fmt.Sprintf("cluster--%s-%s", NAMESPACE, svcName)))
	// Tenant should be akotenant instead of admin
	g.Expect(nodes[0].Tenant).To(gomega.Equal(AKOTENANT))
	g.Expect(nodes[0].PortProto[0].Port).To(gomega.Equal(int32(8080)))
	g.Expect(*nodes[0].TrafficEnabled).To(gomega.Equal(false))

	// Check for the pools
	g.Expect(nodes[0].PoolRefs).To(gomega.HaveLen(1))
	address := "1.1.1.1"
	g.Expect(nodes[0].PoolRefs[0].Servers[0].Ip.Addr).To(gomega.Equal(&address))

	// Enable Traffic
	svcExample.Annotations = map[string]string{lib.VSTrafficDisabled: "false"}
	svcExample.ResourceVersion = "2"
	_, err = KubeClient.CoreV1().Services(NAMESPACE).Update(context.TODO(), svcExample, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating Service: %v", err)
	}

	// Traffic should be enabled.
	g.Eventually(func() bool {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return nodes[0].TrafficEnabled != nil && *nodes[0].TrafficEnabled
	}, 40*time.Second).Should(gomega.Equal(true))

	// Remove annotations
	svcExample.Annotations = nil
	svcExample.ResourceVersion = "3"
	_, err = KubeClient.CoreV1().Services(NAMESPACE).Update(context.TODO(), svcExample, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating Service: %v", err)
	}

	// Traffic should be enabled.
	g.Eventually(func() bool {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return nodes[0].TrafficEnabled != nil && *nodes[0].TrafficEnabled
	}, 40*time.Second).Should(gomega.Equal(true))

	objects.SharedAviGraphLister().Delete(modelName)
	objects.SharedAviGraphLister().Delete(modelName)
	DelSVC(t, NAMESPACE, svcName)
	DelEPS(t, NAMESPACE, svcName)
	mcache := cache.SharedAviObjCache()
	vsKey := cache.NamespaceName{Namespace: AVINAMESPACE, Name: fmt.Sprintf("cluster--%s-%s", svcName, NAMESPACE)}
	g.Eventually(func() bool {
		_, found := mcache.VsCacheMeta.AviCacheGet(vsKey)
		return found
	}, 5*time.Second).Should(gomega.Equal(false))
}
