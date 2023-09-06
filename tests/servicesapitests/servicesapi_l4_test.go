/*
 * Copyright 2020-2021 VMware, Inc.
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

package servicesapitests

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

	"google.golang.org/protobuf/proto"
	svcapifake "sigs.k8s.io/service-apis/pkg/client/clientset/versioned/fake"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/k8s"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	avinodes "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	crdfake "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/client/v1alpha1/clientset/versioned/fake"
	v1beta1crdfake "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/client/v1beta1/clientset/versioned/fake"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/integrationtest"

	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	servicesapi "sigs.k8s.io/service-apis/apis/v1alpha1"
)

const RANDOMUUID = "random-uuid"

var KubeClient *k8sfake.Clientset
var SvcAPIClient *svcapifake.Clientset
var ctrl *k8s.AviController
var CRDClient *crdfake.Clientset
var v1beta1CRDClient *v1beta1crdfake.Clientset

func TestMain(m *testing.M) {
	os.Setenv("SERVICES_API", "true")
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
	KubeClient = k8sfake.NewSimpleClientset()
	CRDClient = crdfake.NewSimpleClientset()
	v1beta1CRDClient = v1beta1crdfake.NewSimpleClientset()
	akoControlConfig.SetCRDClientset(CRDClient)
	akoControlConfig.Setv1beta1CRDClientset(v1beta1CRDClient)
	akoControlConfig.SetEventRecorder(lib.AKOEventComponent, KubeClient, true)
	akoControlConfig.SetAKOInstanceFlag(true)
	akoControlConfig.SetDefaultLBController(true)
	k8s.NewCRDInformers()

	data := map[string][]byte{
		"username": []byte("admin"),
		"password": []byte("admin"),
	}
	object := metav1.ObjectMeta{Name: "avi-secret", Namespace: utils.GetAKONamespace()}
	secret := &corev1.Secret{Data: data, ObjectMeta: object}
	KubeClient.CoreV1().Secrets(utils.GetAKONamespace()).Create(context.TODO(), secret, metav1.CreateOptions{})

	SvcAPIClient = svcapifake.NewSimpleClientset()
	akoControlConfig.SetServicesAPIClientset(SvcAPIClient)
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
	k8s.NewSvcApiInformers(SvcAPIClient)

	mcache := cache.SharedAviObjCache()
	cloudObj := &cache.AviCloudPropertyCache{Name: "Default-Cloud", VType: "mock"}
	subdomains := []string{"avi.internal", ".com"}
	cloudObj.NSIpamDNS = subdomains
	mcache.CloudKeyCache.AviCacheAdd("Default-Cloud", cloudObj)

	integrationtest.InitializeFakeAKOAPIServer()

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
	ctrl.SetSEGroupCloudNameFromNSAnnotations()

	integrationtest.PollForSyncStart(ctrl, 10)
	integrationtest.KubeClient = KubeClient
	integrationtest.AddDefaultNamespace()

	ctrl.HandleConfigMap(informers, ctrlCh, stopCh, quickSyncCh)
	go ctrl.InitController(informers, registeredInformers, ctrlCh, stopCh, quickSyncCh, waitGroupMap)
	integrationtest.KubeClient = KubeClient
	os.Exit(m.Run())
}

// Gateway/GatewayClass lib functions
type FakeGateway struct {
	Name      string
	Namespace string
	GWClass   string
	IPAddress string
	Listeners []FakeGWListener
}

type FakeGWListener struct {
	Port     servicesapi.PortNumber
	Protocol string
	Labels   map[string]string
	HostName *string
}

func (gw FakeGateway) Gateway() *servicesapi.Gateway {
	var fakeListeners []servicesapi.Listener
	for _, listener := range gw.Listeners {
		fakeListeners = append(fakeListeners, servicesapi.Listener{
			Port:     listener.Port,
			Protocol: servicesapi.ProtocolType(listener.Protocol),
			Hostname: (*servicesapi.Hostname)(listener.HostName),
			Routes: servicesapi.RouteBindingSelector{
				Kind: "services",
				Selector: metav1.LabelSelector{
					MatchLabels: listener.Labels,
				},
			},
		})
	}

	gateway := &servicesapi.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: gw.Namespace,
			Name:      gw.Name,
		},
		Spec: servicesapi.GatewaySpec{
			GatewayClassName: gw.GWClass,
			Listeners:        fakeListeners,
		},
	}

	if gw.IPAddress != "" {
		gateway.Spec.Addresses = []servicesapi.GatewayAddress{{
			Type:  servicesapi.IPAddressType,
			Value: gw.IPAddress,
		}}
	}

	return gateway
}

func SetupGateway(t *testing.T, gwname, namespace, gwclass string, protocols ...string) {
	var listeners []FakeGWListener
	if len(protocols) == 0 {
		protocols = append(protocols, "TCP")
	}
	for _, protocol := range protocols {
		listeners = append(listeners, FakeGWListener{
			Port:     8081,
			Protocol: protocol,
			Labels: map[string]string{
				lib.SvcApiGatewayNameLabelKey:      gwname,
				lib.SvcApiGatewayNamespaceLabelKey: namespace,
			},
		})
	}
	gateway := FakeGateway{
		Name:      gwname,
		Namespace: namespace,
		GWClass:   gwclass,
		Listeners: listeners,
	}

	gwCreate := gateway.Gateway()
	if _, err := lib.AKOControlConfig().ServicesAPIClientset().NetworkingV1alpha1().Gateways(namespace).Create(context.TODO(), gwCreate, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding Gateway: %v", err)
	}
}

func TeardownGateway(t *testing.T, gwname, namespace string) {
	if err := lib.AKOControlConfig().ServicesAPIClientset().NetworkingV1alpha1().Gateways(namespace).Delete(context.TODO(), gwname, metav1.DeleteOptions{}); err != nil {
		t.Fatalf("error in deleting Gateway: %v", err)
	}
}

type FakeGWClass struct {
	Name         string
	Controller   string
	InfraSetting string
}

func (gwclass FakeGWClass) GatewayClass() *servicesapi.GatewayClass {
	gatewayclass := &servicesapi.GatewayClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: gwclass.Name,
		},
		Spec: servicesapi.GatewayClassSpec{
			Controller: gwclass.Controller,
		},
	}

	if gwclass.InfraSetting != "" {
		gatewayclass.Spec.ParametersRef = &servicesapi.LocalObjectReference{
			Group: lib.AkoGroup,
			Kind:  lib.AviInfraSetting,
			Name:  gwclass.InfraSetting,
		}
	}

	return gatewayclass
}

func SetupGatewayClass(t *testing.T, gwclassName, controller, infraSetting string) {
	gatewayclass := FakeGWClass{
		Name:         gwclassName,
		Controller:   controller,
		InfraSetting: infraSetting,
	}

	gwClassCreate := gatewayclass.GatewayClass()
	if _, err := lib.AKOControlConfig().ServicesAPIClientset().NetworkingV1alpha1().GatewayClasses().Create(context.TODO(), gwClassCreate, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding GatewayClass: %v", err)
	}
}

func TeardownGatewayClass(t *testing.T, gwClassName string) {
	if err := lib.AKOControlConfig().ServicesAPIClientset().NetworkingV1alpha1().GatewayClasses().Delete(context.TODO(), gwClassName, metav1.DeleteOptions{}); err != nil {
		t.Fatalf("error in deleting GatewayClass: %v", err)
	}
}

func SetupSvcApiService(t *testing.T, svcname, namespace, gwname, gwnamespace, protocol string) {
	svc := integrationtest.FakeService{
		Name:      svcname,
		Namespace: namespace,
		Labels: map[string]string{
			lib.SvcApiGatewayNameLabelKey:      gwname,
			lib.SvcApiGatewayNamespaceLabelKey: gwnamespace,
		},
		Type:         corev1.ServiceTypeClusterIP,
		ServicePorts: []integrationtest.Serviceport{{PortName: "foo1", Protocol: corev1.Protocol(protocol), PortNumber: 8081, TargetPort: intstr.FromInt(8081)}},
	}

	svcCreate := svc.Service()
	if _, err := KubeClient.CoreV1().Services(namespace).Create(context.TODO(), svcCreate, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding Service: %v", err)
	}
	integrationtest.CreateEP(t, namespace, svcname, false, true, "1.1.1")
}

func SetupSvcApiLBServiceWithLBClass(t *testing.T, svcname, namespace, gwname, gwnamespace, protocol string, LBClass string) {
	svc := integrationtest.FakeService{
		Name:      svcname,
		Namespace: namespace,
		Labels: map[string]string{
			lib.SvcApiGatewayNameLabelKey:      gwname,
			lib.SvcApiGatewayNamespaceLabelKey: gwnamespace,
		},
		Type:              corev1.ServiceTypeLoadBalancer,
		ServicePorts:      []integrationtest.Serviceport{{PortName: "foo1", Protocol: corev1.Protocol(protocol), PortNumber: 8081, TargetPort: intstr.FromInt(8081)}},
		LoadBalancerClass: LBClass,
	}
	svcCreate := svc.Service()
	if _, err := KubeClient.CoreV1().Services(namespace).Create(context.TODO(), svcCreate, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding Service: %v", err)
	}
	integrationtest.CreateEP(t, namespace, svcname, false, true, "1.1.1")
}

func TeardownAdvLBService(t *testing.T, svcname, namespace string) {
	if err := KubeClient.CoreV1().Services(namespace).Delete(context.TODO(), svcname, metav1.DeleteOptions{}); err != nil {
		t.Fatalf("error in deleting AdvLB Service: %v", err)
	}
	integrationtest.DelEP(t, namespace, svcname)
}

func VerifyGatewayVSNodeDeletion(g *gomega.WithT, modelName string) {
	g.Eventually(func() interface{} {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		return aviModel
	}, 30*time.Second).Should(gomega.BeNil())
}

func TestServiceAPISvcWithLoadBalancerClass(t *testing.T) {
	// This test checks whether AKO ignores gateway labels for LB services in ServiceAPI scenario
	g := gomega.NewGomegaWithT(t)

	gwClassName, gatewayName, ns := "avi-lb", "my-gateway", "default"
	modelName := "admin/cluster--default-my-gateway"
	SetupGatewayClass(t, gwClassName, lib.SvcApiAviGatewayController, "")
	SetupGateway(t, gatewayName, ns, gwClassName)

	// LB Service with invalid LBClass should be processed for DedicatedVS and be invalidated with AKO ignoring gateway labels
	SetupSvcApiLBServiceWithLBClass(t, "svc", ns, gatewayName, ns, "TCP", integrationtest.INVALID_LB_CLASS)

	g.Eventually(func() string {
		gw, _ := SvcAPIClient.NetworkingV1alpha1().Gateways(ns).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		if len(gw.Status.Addresses) > 0 {
			return gw.Status.Addresses[0].Value
		}
		return ""
	}, 40*time.Second).Should(gomega.Equal("10.250.250.1"))

	_, aviModel := objects.SharedAviGraphLister().Get("admin/cluster--default-svc")
	g.Expect(aviModel).To(gomega.BeNil())

	TeardownAdvLBService(t, "svc", ns)

	// LB Service with valid LBClass should be processed for DedicatedVS and be validated with AKO ignoring gateway labels
	SetupSvcApiLBServiceWithLBClass(t, "svc", ns, gatewayName, ns, "TCP", lib.AviIngressController)
	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get("admin/cluster--default-svc")
		return found
	}, 30*time.Second).Should(gomega.Equal(true))
	_, aviModel = objects.SharedAviGraphLister().Get("admin/cluster--default-svc")
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes).To(gomega.HaveLen(1))

	TeardownGatewayClass(t, gwClassName)
	g.Eventually(func() int {
		gw, _ := SvcAPIClient.NetworkingV1alpha1().Gateways(ns).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		return len(gw.Status.Addresses)
	}, 40*time.Second).Should(gomega.Equal(0))

	TeardownGateway(t, gatewayName, ns)
	TeardownAdvLBService(t, "svc", ns)
	VerifyGatewayVSNodeDeletion(g, modelName)
}
func TestServicesAPISvcHostnameStatusUpdate(t *testing.T) {
	// create gw, svc1, svc2 on separate listeners
	// assign hostname to svc1, autofqdn for svc2, check model, check status
	// assign hostname to svc2 via listener, check model, check status

	g := gomega.NewGomegaWithT(t)

	gwClassName, gatewayName, ns := "avi-lb", "my-gateway", "default"
	svcName1, svcName2 := "svc1", "svc2"
	modelName := "admin/cluster--default-my-gateway"
	labels := map[string]string{lib.SvcApiGatewayNameLabelKey: gatewayName, lib.SvcApiGatewayNamespaceLabelKey: ns}

	SetupGatewayClass(t, gwClassName, lib.SvcApiAviGatewayController, "")

	gateway := FakeGateway{
		Name:      gatewayName,
		Namespace: ns,
		GWClass:   gwClassName,
		Listeners: []FakeGWListener{
			{Port: 8081, Protocol: "TCP", Labels: labels, HostName: proto.String("foo.avi.internal")},
			{Port: 8082, Protocol: "TCP", Labels: labels},
		},
	}
	if _, err := lib.AKOControlConfig().ServicesAPIClientset().NetworkingV1alpha1().Gateways(ns).Create(context.TODO(), gateway.Gateway(), metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding Gateway: %v", err)
	}

	svc1 := integrationtest.FakeService{
		Name:         svcName1,
		Namespace:    ns,
		Labels:       labels,
		Type:         corev1.ServiceTypeClusterIP,
		ServicePorts: []integrationtest.Serviceport{{PortName: "footcp", Protocol: "TCP", PortNumber: 8081, TargetPort: intstr.FromInt(80)}},
	}
	if _, err := KubeClient.CoreV1().Services(ns).Create(context.TODO(), svc1.Service(), metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding Service: %v", err)
	}

	svc2 := integrationtest.FakeService{
		Name:         svcName2,
		Namespace:    ns,
		Labels:       labels,
		Type:         corev1.ServiceTypeClusterIP,
		ServicePorts: []integrationtest.Serviceport{{PortName: "footcp", Protocol: "TCP", PortNumber: 8082, TargetPort: intstr.FromInt(80)}},
	}
	if _, err := KubeClient.CoreV1().Services(ns).Create(context.TODO(), svc2.Service(), metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding Service: %v", err)
	}

	integrationtest.CreateEP(t, ns, svcName1, false, true, "1.1.1")
	integrationtest.CreateEP(t, ns, svcName2, false, true, "1.1.1")

	g.Eventually(func() bool {
		svc1, _ := KubeClient.CoreV1().Services(ns).Get(context.TODO(), svcName1, metav1.GetOptions{})
		svc2, _ := KubeClient.CoreV1().Services(ns).Get(context.TODO(), svcName2, metav1.GetOptions{})
		if len(svc1.Status.LoadBalancer.Ingress) > 0 &&
			len(svc2.Status.LoadBalancer.Ingress) > 0 &&
			svc1.Status.LoadBalancer.Ingress[0].Hostname == "foo.avi.internal" &&
			svc2.Status.LoadBalancer.Ingress[0].Hostname == "svc2.default.com" {
			return true
		}
		return false
	}, 30*time.Second).Should(gomega.Equal(true))

	gatewayUpdate := FakeGateway{
		Name:      gatewayName,
		Namespace: ns,
		GWClass:   gwClassName,
		Listeners: []FakeGWListener{
			{Port: 8081, Protocol: "TCP", Labels: labels},
			{Port: 8082, Protocol: "TCP", Labels: labels, HostName: proto.String("bar.avi.internal")},
		},
	}
	if _, err := lib.AKOControlConfig().ServicesAPIClientset().NetworkingV1alpha1().Gateways(ns).Update(context.TODO(), gatewayUpdate.Gateway(), metav1.UpdateOptions{}); err != nil {
		t.Fatalf("error in updating Gateway: %v", err)
	}

	g.Eventually(func() bool {
		svc1, _ := KubeClient.CoreV1().Services(ns).Get(context.TODO(), svcName1, metav1.GetOptions{})
		svc2, _ := KubeClient.CoreV1().Services(ns).Get(context.TODO(), svcName2, metav1.GetOptions{})
		if len(svc1.Status.LoadBalancer.Ingress) > 0 &&
			len(svc2.Status.LoadBalancer.Ingress) > 0 &&
			svc1.Status.LoadBalancer.Ingress[0].Hostname == "svc1.default.com" &&
			svc2.Status.LoadBalancer.Ingress[0].Hostname == "bar.avi.internal" {
			return true
		}
		return false
	}, 30*time.Second).Should(gomega.Equal(true))

	TeardownAdvLBService(t, svcName1, ns)
	TeardownAdvLBService(t, svcName2, ns)
	TeardownGateway(t, gatewayName, ns)
	TeardownGatewayClass(t, gwClassName)
	VerifyGatewayVSNodeDeletion(g, modelName)
}

func TestServicesAPIBestCase(t *testing.T) {
	// create gwclass, create gw, create 1svc
	// check graph VsNode vals, check IP status
	// remove gwclasss, IP removed
	g := gomega.NewGomegaWithT(t)

	gwClassName, gatewayName, ns := "avi-lb", "my-gateway", "default"
	modelName := "admin/cluster--default-my-gateway"

	SetupGatewayClass(t, gwClassName, lib.SvcApiAviGatewayController, "")
	SetupGateway(t, gatewayName, ns, gwClassName)

	SetupSvcApiService(t, "svc", ns, gatewayName, ns, "TCP")

	g.Eventually(func() string {
		gw, _ := SvcAPIClient.NetworkingV1alpha1().Gateways(ns).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		if len(gw.Status.Addresses) > 0 {
			return gw.Status.Addresses[0].Value
		}
		return ""
	}, 40*time.Second).Should(gomega.Equal("10.250.250.1"))

	g.Eventually(func() string {
		svc, _ := KubeClient.CoreV1().Services(ns).Get(context.TODO(), "svc", metav1.GetOptions{})
		if len(svc.Status.LoadBalancer.Ingress) > 0 {
			return svc.Status.LoadBalancer.Ingress[0].IP
		}
		return ""
	}, 30*time.Second).Should(gomega.Equal("10.250.250.1"))

	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes[0].PortProto[0].Port).To(gomega.Equal(int32(8081)))
	g.Expect(nodes[0].HttpPolicySetRefs).To(gomega.HaveLen(0))
	g.Expect(nodes[0].L4PolicyRefs).To(gomega.HaveLen(1))
	g.Expect(nodes[0].L4PolicyRefs[0].PortPool[0].Port).To(gomega.Equal(uint32(8081)))
	g.Expect(nodes[0].L4PolicyRefs[0].PortPool[0].Protocol).To(gomega.Equal("TCP"))
	g.Expect(nodes[0].PoolRefs[0].ServiceMetadata.NamespaceServiceName[0]).To(gomega.Equal("default/svc"))
	g.Expect(nodes[0].ServiceMetadata.Gateway).To(gomega.Equal("default/my-gateway"))
	g.Expect(nodes[0].PoolRefs[0].Servers).To(gomega.HaveLen(3))

	TeardownGatewayClass(t, gwClassName)
	g.Eventually(func() int {
		gw, _ := SvcAPIClient.NetworkingV1alpha1().Gateways(ns).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		return len(gw.Status.Addresses)
	}, 40*time.Second).Should(gomega.Equal(0))

	TeardownAdvLBService(t, "svc", ns)
	TeardownGateway(t, gatewayName, ns)
	VerifyGatewayVSNodeDeletion(g, modelName)
}

func TestServicesAPINamingConvention(t *testing.T) {
	// create gwclass, create gw, create 1svc
	// check naming conventions for vs, pool, l4policy
	// checks naming convention of all generated nodes
	g := gomega.NewGomegaWithT(t)

	gwClassName, gatewayName, ns := "avi-lb", "my-gateway", "default"
	modelName := "admin/cluster--default-my-gateway"

	SetupGatewayClass(t, gwClassName, lib.SvcApiAviGatewayController, "")
	SetupGateway(t, gatewayName, ns, gwClassName)

	SetupSvcApiService(t, "svc", ns, gatewayName, ns, "TCP")

	g.Eventually(func() string {
		gw, _ := SvcAPIClient.NetworkingV1alpha1().Gateways(ns).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		if len(gw.Status.Addresses) > 0 {
			return gw.Status.Addresses[0].Value
		}
		return ""
	}, 40*time.Second).Should(gomega.Equal("10.250.250.1"))

	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes[0].Name).To(gomega.Equal("cluster--default-my-gateway"))
	g.Expect(nodes[0].PoolRefs[0].Name).To(gomega.Equal("cluster--default-svc-my-gateway-TCP-8081"))
	g.Expect(nodes[0].L4PolicyRefs[0].Name).To(gomega.Equal("cluster--default-my-gateway"))

	TeardownGatewayClass(t, gwClassName)
	TeardownAdvLBService(t, "svc", ns)
	TeardownGateway(t, gatewayName, ns)
	VerifyGatewayVSNodeDeletion(g, modelName)
}

func TestServicesAPIWithStaticIP(t *testing.T) {
	// create gwclass, create gw, create 1svc
	// check graph VsNode IPAddress val in vsvip ref
	g := gomega.NewGomegaWithT(t)

	gwClassName, gatewayName, ns := "avi-lb", "my-gateway", "default"
	modelName := "admin/cluster--default-my-gateway"
	staticIP := "80.80.80.80"

	SetupGatewayClass(t, gwClassName, lib.SvcApiAviGatewayController, "")
	gateway := FakeGateway{
		Name:      gatewayName,
		Namespace: ns,
		GWClass:   gwClassName,
		IPAddress: staticIP,
		Listeners: []FakeGWListener{{
			Port:     8081,
			Protocol: "TCP",
			Labels: map[string]string{
				lib.SvcApiGatewayNameLabelKey:      gatewayName,
				lib.SvcApiGatewayNamespaceLabelKey: ns,
			},
		}},
	}
	gwCreate := gateway.Gateway()
	if _, err := lib.AKOControlConfig().ServicesAPIClientset().NetworkingV1alpha1().Gateways(ns).Create(context.TODO(), gwCreate, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding Gateway: %v", err)
	}

	SetupSvcApiService(t, "svc", ns, gatewayName, ns, "TCP")

	g.Eventually(func() string {
		if found, aviModel := objects.SharedAviGraphLister().Get(modelName); found && aviModel != nil {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
			if len(nodes) > 0 && len(nodes[0].VSVIPRefs) > 0 {
				return nodes[0].VSVIPRefs[0].IPAddress
			}
		}
		return ""
	}, 40*time.Second).Should(gomega.Equal(staticIP))

	TeardownGatewayClass(t, gwClassName)
	TeardownAdvLBService(t, "svc", ns)
	TeardownGateway(t, gatewayName, ns)
	VerifyGatewayVSNodeDeletion(g, modelName)
}

// Gateway - GWClass mapping tests

func TestServicesAPIWrongControllerGWClass(t *testing.T) {
	// create gateway, nothing happens
	// create gatewayclass, VS created
	// update to bad gatewayclass (wrong controller), VS deleted
	g := gomega.NewGomegaWithT(t)

	gwClassName, gatewayName, ns := "avi-lb", "my-gateway", "default"
	modelName := "admin/cluster--default-my-gateway"

	SetupGateway(t, gatewayName, ns, gwClassName)
	SetupSvcApiService(t, "svc", ns, gatewayName, ns, "TCP")

	SetupGatewayClass(t, gwClassName, lib.SvcApiAviGatewayController, "")

	g.Eventually(func() string {
		gw, _ := SvcAPIClient.NetworkingV1alpha1().Gateways(ns).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		if len(gw.Status.Addresses) > 0 {
			return gw.Status.Addresses[0].Value
		}
		return ""
	}, 40*time.Second).Should(gomega.Equal("10.250.250.1"))

	gwclassUpdate := FakeGWClass{
		Name:       gwClassName,
		Controller: "xyz",
	}.GatewayClass()
	gwclassUpdate.ResourceVersion = "2"
	if _, err := lib.AKOControlConfig().ServicesAPIClientset().NetworkingV1alpha1().GatewayClasses().Update(context.TODO(), gwclassUpdate, metav1.UpdateOptions{}); err != nil {
		t.Fatalf("error in updating GatewayClass: %v", err)
	}

	g.Eventually(func() int {
		gw, _ := SvcAPIClient.NetworkingV1alpha1().Gateways(ns).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		return len(gw.Status.Addresses)
	}, 40*time.Second).Should(gomega.Equal(0))
	g.Eventually(func() int {
		svc, _ := KubeClient.CoreV1().Services(ns).Get(context.TODO(), "svc", metav1.GetOptions{})
		return len(svc.Status.LoadBalancer.Ingress)
	}, 40*time.Second).Should(gomega.Equal(0))

	TeardownAdvLBService(t, "svc", ns)
	TeardownGateway(t, gatewayName, ns)
	TeardownGatewayClass(t, gwClassName)
	VerifyGatewayVSNodeDeletion(g, modelName)
}

func TestServicesAPIWrongClassMappingInGateway(t *testing.T) {
	// create gwclass, gw
	// update wrong mapping of class in gw, VS deleted
	// fix class in gw, VS created
	g := gomega.NewGomegaWithT(t)

	gwClassName, gatewayName, ns := "avi-lb", "my-gateway", "default"
	modelName := "admin/cluster--default-my-gateway"

	SetupGateway(t, gatewayName, ns, gwClassName)
	SetupSvcApiService(t, "svc", ns, gatewayName, ns, "TCP")

	SetupGatewayClass(t, gwClassName, lib.SvcApiAviGatewayController, "")

	g.Eventually(func() string {
		gw, _ := SvcAPIClient.NetworkingV1alpha1().Gateways(ns).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		if len(gw.Status.Addresses) > 0 {
			return gw.Status.Addresses[0].Value
		}
		return ""
	}, 20*time.Second).Should(gomega.Equal("10.250.250.1"))

	gwUpdate := FakeGateway{
		Name: gatewayName, Namespace: ns, GWClass: gwClassName,
		Listeners: []FakeGWListener{{
			Port: 8081, Protocol: "TCP",
			Labels: map[string]string{
				lib.SvcApiGatewayNameLabelKey:      "BADGATEWAY",
				lib.SvcApiGatewayNamespaceLabelKey: ns,
			},
		}},
	}.Gateway()
	gwUpdate.ResourceVersion = "2"
	if _, err := lib.AKOControlConfig().ServicesAPIClientset().NetworkingV1alpha1().Gateways(ns).Update(context.TODO(), gwUpdate, metav1.UpdateOptions{}); err != nil {
		t.Fatalf("error in updating Gateway: %v", err)
	}

	// vsNode must get deleted
	VerifyGatewayVSNodeDeletion(g, modelName)
	g.Eventually(func() int {
		gw, _ := SvcAPIClient.NetworkingV1alpha1().Gateways(ns).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		return len(gw.Status.Addresses)
	}, 10*time.Second).Should(gomega.Equal(0))

	gwUpdate = FakeGateway{
		Name: gatewayName, Namespace: ns, GWClass: gwClassName,
		Listeners: []FakeGWListener{{
			Port: 8081, Protocol: "TCP",
			Labels: map[string]string{
				lib.SvcApiGatewayNameLabelKey:      gatewayName,
				lib.SvcApiGatewayNamespaceLabelKey: ns,
			},
		}},
	}.Gateway()
	gwUpdate.ResourceVersion = "3"
	if _, err := lib.AKOControlConfig().ServicesAPIClientset().NetworkingV1alpha1().Gateways(ns).Update(context.TODO(), gwUpdate, metav1.UpdateOptions{}); err != nil {
		t.Fatalf("error in updating Gateway: %v", err)
	}

	// vsNode must come back up
	g.Eventually(func() int {
		gw, _ := SvcAPIClient.NetworkingV1alpha1().Gateways(ns).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		return len(gw.Status.Addresses)
	}, 10*time.Second).Should(gomega.Equal(1))

	TeardownAdvLBService(t, "svc", ns)
	TeardownGateway(t, gatewayName, ns)
	TeardownGatewayClass(t, gwClassName)
	VerifyGatewayVSNodeDeletion(g, modelName)
}

func TestServicesAPIProtocolChangeInService(t *testing.T) {
	// gw/tcp/8081 svc/tcp/8081  -> svc/udp/8081
	// service protocol changes Pool deleted
	g := gomega.NewGomegaWithT(t)

	gwClassName, gatewayName, ns := "avi-lb", "my-gateway", "default"
	modelName := "admin/cluster--default-my-gateway"

	SetupGatewayClass(t, gwClassName, lib.SvcApiAviGatewayController, "")
	SetupGateway(t, gatewayName, ns, gwClassName)
	SetupSvcApiService(t, "svc", ns, gatewayName, ns, "TCP")

	g.Eventually(func() bool {
		if found, aviModel := objects.SharedAviGraphLister().Get(modelName); found && aviModel != nil {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
			if len(nodes) > 0 && len(nodes[0].PoolRefs) == 1 {
				return true
			}
		}
		return false
	}, 10*time.Second).Should(gomega.Equal(true))

	svcUpdate := integrationtest.FakeService{
		Name:      "svc",
		Namespace: ns,
		Labels: map[string]string{
			lib.SvcApiGatewayNameLabelKey:      gatewayName,
			lib.SvcApiGatewayNamespaceLabelKey: ns,
		},
		Type:         corev1.ServiceTypeClusterIP,
		ServicePorts: []integrationtest.Serviceport{{PortName: "foo1", Protocol: corev1.ProtocolUDP, PortNumber: 8081, TargetPort: intstr.FromInt(8081)}},
	}.Service()
	svcUpdate.ResourceVersion = "2"
	if _, err := KubeClient.CoreV1().Services(ns).Update(context.TODO(), svcUpdate, metav1.UpdateOptions{}); err != nil {
		t.Fatalf("error in updating Service: %v", err)
	}

	g.Eventually(func() bool {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if found && aviModel != nil {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
			if len(nodes[0].PoolRefs) == 0 {
				return true
			}
		}
		return false
	}, 50*time.Second).Should(gomega.Equal(true))

	TeardownAdvLBService(t, "svc", ns)
	TeardownGateway(t, gatewayName, ns)
	TeardownGatewayClass(t, gwClassName)
	VerifyGatewayVSNodeDeletion(g, modelName)
}

func TestServicesAPIPortChangeInService(t *testing.T) {
	// gw/tcp/8081 svc/tcp/8081 -> svc/tcp/8080
	// service port changes Pools deleted
	g := gomega.NewGomegaWithT(t)

	gwClassName, gatewayName, ns := "avi-lb", "my-gateway", "default"
	modelName := "admin/cluster--default-my-gateway"

	SetupGatewayClass(t, gwClassName, lib.SvcApiAviGatewayController, "")
	SetupGateway(t, gatewayName, ns, gwClassName)
	SetupSvcApiService(t, "svc", ns, gatewayName, ns, "TCP")

	g.Eventually(func() bool {
		if found, aviModel := objects.SharedAviGraphLister().Get(modelName); found && aviModel != nil {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
			if len(nodes) > 0 && len(nodes[0].PoolRefs) == 1 {
				return true
			}
		}
		return false
	}, 10*time.Second).Should(gomega.Equal(true))

	svcUpdate := integrationtest.FakeService{
		Name:      "svc",
		Namespace: ns,
		Labels: map[string]string{
			lib.SvcApiGatewayNameLabelKey:      gatewayName,
			lib.SvcApiGatewayNamespaceLabelKey: ns,
		},
		Type:         corev1.ServiceTypeClusterIP,
		ServicePorts: []integrationtest.Serviceport{{PortName: "foo1", Protocol: corev1.ProtocolTCP, PortNumber: 8080, TargetPort: intstr.FromInt(8081)}},
	}.Service()
	svcUpdate.ResourceVersion = "2"
	if _, err := KubeClient.CoreV1().Services(ns).Update(context.TODO(), svcUpdate, metav1.UpdateOptions{}); err != nil {
		t.Fatalf("error in updating Service: %v", err)
	}

	g.Eventually(func() bool {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if found && aviModel != nil {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
			if len(nodes[0].PoolRefs) == 0 {
				return true
			}
		}
		return false
	}, 40*time.Second).Should(gomega.Equal(true))

	TeardownAdvLBService(t, "svc", ns)
	TeardownGateway(t, gatewayName, ns)
	TeardownGatewayClass(t, gwClassName)
	VerifyGatewayVSNodeDeletion(g, modelName)
}

func TestServicesAPILabelUpdatesInService(t *testing.T) {
	// correct labels, label mismatch, correct labels, delete labels
	g := gomega.NewGomegaWithT(t)

	gwClassName, gatewayName, ns := "avi-lb", "my-gateway", "default"
	modelName := "admin/cluster--default-my-gateway"

	SetupGatewayClass(t, gwClassName, lib.SvcApiAviGatewayController, "")
	SetupGateway(t, gatewayName, ns, gwClassName)
	SetupSvcApiService(t, "svc", ns, gatewayName, ns, "TCP")

	g.Eventually(func() bool {
		if found, aviModel := objects.SharedAviGraphLister().Get(modelName); found && aviModel != nil {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
			if len(nodes) > 0 && len(nodes[0].PoolRefs) == 1 {
				return true
			}
		}
		return false
	}, 10*time.Second).Should(gomega.Equal(true))

	svcUpdate := integrationtest.FakeService{
		Name:      "svc",
		Namespace: ns,
		Labels: map[string]string{
			lib.SvcApiGatewayNameLabelKey:      "BADGATEWAY",
			lib.SvcApiGatewayNamespaceLabelKey: ns,
		},
		Type:         corev1.ServiceTypeClusterIP,
		ServicePorts: []integrationtest.Serviceport{{PortName: "foo1", Protocol: corev1.ProtocolTCP, PortNumber: 8081, TargetPort: intstr.FromInt(8081)}},
	}.Service()
	svcUpdate.ResourceVersion = "2"
	if _, err := KubeClient.CoreV1().Services(ns).Update(context.TODO(), svcUpdate, metav1.UpdateOptions{}); err != nil {
		t.Fatalf("error in updating Service: %v", err)
	}

	g.Eventually(func() bool {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if found && aviModel != nil {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
			if len(nodes[0].PoolRefs) == 0 {
				return true
			}
		}
		return false
	}, 40*time.Second).Should(gomega.Equal(true))

	TeardownAdvLBService(t, "svc", ns)
	TeardownGateway(t, gatewayName, ns)
	TeardownGatewayClass(t, gwClassName)
	VerifyGatewayVSNodeDeletion(g, modelName)
}

func TestServicesAPILabelUpdatesInGateway(t *testing.T) {
	// correct labels, label mismatch, correct labels, delete labels
	g := gomega.NewGomegaWithT(t)

	gwClassName, gatewayName, ns := "avi-lb", "my-gateway", "default"
	modelName := "admin/cluster--default-my-gateway"

	SetupGatewayClass(t, gwClassName, lib.SvcApiAviGatewayController, "")
	SetupGateway(t, gatewayName, ns, gwClassName)
	SetupSvcApiService(t, "svc", ns, gatewayName, ns, "TCP")

	g.Eventually(func() bool {
		if found, aviModel := objects.SharedAviGraphLister().Get(modelName); found && aviModel != nil {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
			if len(nodes) > 0 && len(nodes[0].PoolRefs) == 1 {
				return true
			}
		}
		return false
	}, 10*time.Second).Should(gomega.Equal(true))

	gwUpdate := FakeGateway{
		Name:      gatewayName,
		Namespace: ns,
		GWClass:   gwClassName,
		Listeners: []FakeGWListener{{
			Port:     8081,
			Protocol: "TCP",
			Labels: map[string]string{
				lib.SvcApiGatewayNameLabelKey:      "BADGATEWAY",
				lib.SvcApiGatewayNamespaceLabelKey: ns,
			},
		}},
	}.Gateway()
	gwUpdate.ResourceVersion = "2"
	if _, err := lib.AKOControlConfig().ServicesAPIClientset().NetworkingV1alpha1().Gateways(ns).Update(context.TODO(), gwUpdate, metav1.UpdateOptions{}); err != nil {
		t.Fatalf("error in updating Gateway: %v", err)
	}

	g.Eventually(func() bool {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if found && aviModel == nil {
			return false
		}
		return true
	}, 40*time.Second).Should(gomega.Equal(false))

	TeardownAdvLBService(t, "svc", ns)
	TeardownGateway(t, gatewayName, ns)
	TeardownGatewayClass(t, gwClassName)
	VerifyGatewayVSNodeDeletion(g, modelName)
}

func TestServicesAPIGatewayListenerPortUpdate(t *testing.T) {
	// svc/tcp/8081
	// change gateway listener port to 8080, Pools delete
	// change svc port to 8080, VS creates, with 8080 exposed port
	g := gomega.NewGomegaWithT(t)

	gwClassName, gatewayName, ns := "avi-lb", "my-gateway", "default"
	modelName := "admin/cluster--default-my-gateway"

	SetupGatewayClass(t, gwClassName, lib.SvcApiAviGatewayController, "")
	SetupGateway(t, gatewayName, ns, gwClassName)
	SetupSvcApiService(t, "svc", ns, gatewayName, ns, "TCP")

	g.Eventually(func() bool {
		if found, aviModel := objects.SharedAviGraphLister().Get(modelName); found && aviModel != nil {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
			if len(nodes) > 0 &&
				len(nodes[0].PoolRefs) == 1 &&
				nodes[0].PoolRefs[0].Protocol == "TCP" {
				return true
			}
		}
		return false
	}, 10*time.Second).Should(gomega.Equal(true))

	gwUpdate := FakeGateway{
		Name:      gatewayName,
		Namespace: ns,
		GWClass:   gwClassName,
		Listeners: []FakeGWListener{{
			Port:     8080,
			Protocol: "TCP",
			Labels: map[string]string{
				lib.SvcApiGatewayNameLabelKey:      gatewayName,
				lib.SvcApiGatewayNamespaceLabelKey: ns,
			},
		}},
	}.Gateway()
	gwUpdate.ResourceVersion = "2"
	if _, err := lib.AKOControlConfig().ServicesAPIClientset().NetworkingV1alpha1().Gateways(ns).Update(context.TODO(), gwUpdate, metav1.UpdateOptions{}); err != nil {
		t.Fatalf("error in updating Gateway: %v", err)
	}

	g.Eventually(func() bool {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if found && aviModel != nil {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
			if len(nodes[0].PoolRefs) == 0 {
				return true
			}
		}
		return false
	}, 40*time.Second).Should(gomega.Equal(true))

	// match service to chaged gateway port: 8080
	svcUpdate := integrationtest.FakeService{
		Name:      "svc",
		Namespace: ns,
		Labels: map[string]string{
			lib.SvcApiGatewayNameLabelKey:      gatewayName,
			lib.SvcApiGatewayNamespaceLabelKey: ns,
		},
		Type:         corev1.ServiceTypeClusterIP,
		ServicePorts: []integrationtest.Serviceport{{PortName: "foo1", Protocol: corev1.ProtocolTCP, PortNumber: 8080, TargetPort: intstr.FromInt(8081)}},
	}.Service()
	svcUpdate.ResourceVersion = "2"
	if _, err := KubeClient.CoreV1().Services(ns).Update(context.TODO(), svcUpdate, metav1.UpdateOptions{}); err != nil {
		t.Fatalf("error in updating Service: %v", err)
	}

	g.Eventually(func() bool {
		if found, aviModel := objects.SharedAviGraphLister().Get(modelName); found && aviModel != nil {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
			if len(nodes) > 0 &&
				len(nodes[0].PoolRefs) == 1 &&
				nodes[0].PoolRefs[0].Protocol == "TCP" {
				return true
			}
		}
		return false
	}, 10*time.Second).Should(gomega.Equal(true))

	TeardownAdvLBService(t, "svc", ns)
	TeardownGateway(t, gatewayName, ns)
	TeardownGatewayClass(t, gwClassName)
	VerifyGatewayVSNodeDeletion(g, modelName)
}

func TestServicesAPIGatewayListenerProtocolUpdate(t *testing.T) {
	// svc/tcp/8080
	// change gateway listener protocol to UDP, Pools delete
	// change svc protocol to UDP, VS creates, with UDP protocol
	g := gomega.NewGomegaWithT(t)

	gwClassName, gatewayName, ns := "avi-lb", "my-gateway", "default"
	modelName := "admin/cluster--default-my-gateway"

	SetupGatewayClass(t, gwClassName, lib.SvcApiAviGatewayController, "")
	SetupGateway(t, gatewayName, ns, gwClassName)
	SetupSvcApiService(t, "svc", ns, gatewayName, ns, "TCP")

	g.Eventually(func() bool {
		if found, aviModel := objects.SharedAviGraphLister().Get(modelName); found && aviModel != nil {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
			if len(nodes) > 0 &&
				len(nodes[0].PoolRefs) == 1 &&
				nodes[0].PoolRefs[0].Protocol == "TCP" {
				return true
			}
		}
		return false
	}, 10*time.Second).Should(gomega.Equal(true))

	gwUpdate := FakeGateway{
		Name:      gatewayName,
		Namespace: ns,
		GWClass:   gwClassName,
		Listeners: []FakeGWListener{{
			Port:     8081,
			Protocol: "UDP",
			Labels: map[string]string{
				lib.SvcApiGatewayNameLabelKey:      gatewayName,
				lib.SvcApiGatewayNamespaceLabelKey: ns,
			},
		}},
	}.Gateway()
	gwUpdate.ResourceVersion = "2"
	if _, err := lib.AKOControlConfig().ServicesAPIClientset().NetworkingV1alpha1().Gateways(ns).Update(context.TODO(), gwUpdate, metav1.UpdateOptions{}); err != nil {
		t.Fatalf("error in updating Gateway: %v", err)
	}

	g.Eventually(func() bool {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if found && aviModel != nil {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
			if len(nodes[0].PoolRefs) == 0 {
				return true
			}
		}
		return false
	}, 40*time.Second).Should(gomega.Equal(true))

	// match service to chaged gateway protocol: UDP
	svcUpdate := integrationtest.FakeService{
		Name:      "svc",
		Namespace: ns,
		Labels: map[string]string{
			lib.SvcApiGatewayNameLabelKey:      gatewayName,
			lib.SvcApiGatewayNamespaceLabelKey: ns,
		},
		Type:         corev1.ServiceTypeClusterIP,
		ServicePorts: []integrationtest.Serviceport{{PortName: "foo1", Protocol: corev1.ProtocolUDP, PortNumber: 8081, TargetPort: intstr.FromInt(8081)}},
	}.Service()
	svcUpdate.ResourceVersion = "2"
	if _, err := KubeClient.CoreV1().Services(ns).Update(context.TODO(), svcUpdate, metav1.UpdateOptions{}); err != nil {
		t.Fatalf("error in updating Service: %v", err)
	}

	g.Eventually(func() bool {
		if found, aviModel := objects.SharedAviGraphLister().Get(modelName); found && aviModel != nil {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
			if len(nodes) > 0 &&
				len(nodes[0].PoolRefs) == 1 &&
				nodes[0].PoolRefs[0].Protocol == "UDP" {
				return true
			}
		}
		return false
	}, 10*time.Second).Should(gomega.Equal(true))

	TeardownAdvLBService(t, "svc", ns)
	TeardownGateway(t, gatewayName, ns)
	TeardownGatewayClass(t, gwClassName)
	VerifyGatewayVSNodeDeletion(g, modelName)
}

func TestServicesAPIMultiGatewayServiceUpdate(t *testing.T) {
	// svc/tcp/8081, gw1/tcp/8081, gw2/tcp/8081
	// change gateway from gw1 to gw2, gw1 Pools delete, gw2 VS is created
	g := gomega.NewGomegaWithT(t)

	gwClassName, gateway1Name, gateway2Name, ns := "avi-lb", "my-gateway1", "my-gateway2", "default"
	modelName1 := "admin/cluster--default-my-gateway1"
	modelName2 := "admin/cluster--default-my-gateway2"

	SetupGatewayClass(t, gwClassName, lib.SvcApiAviGatewayController, "")
	SetupGateway(t, gateway1Name, ns, gwClassName)
	SetupGateway(t, gateway2Name, ns, gwClassName)
	SetupSvcApiService(t, "svc", ns, gateway1Name, ns, "TCP")

	g.Eventually(func() bool {
		if found, aviModel := objects.SharedAviGraphLister().Get(modelName1); found && aviModel != nil {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
			if len(nodes) > 0 && nodes[0].Name == "cluster--default-my-gateway1" {
				return true
			}
		}
		return false
	}, 10*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName1)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes[0].PoolRefs).Should(gomega.HaveLen(1))
	g.Expect(nodes[0].L4PolicyRefs).Should(gomega.HaveLen(1))

	// change service gw binding from gw1 to gw2
	svcUpdate := integrationtest.FakeService{
		Name:      "svc",
		Namespace: ns,
		Labels: map[string]string{
			lib.SvcApiGatewayNameLabelKey:      gateway2Name,
			lib.SvcApiGatewayNamespaceLabelKey: ns,
		},
		Type:         corev1.ServiceTypeClusterIP,
		ServicePorts: []integrationtest.Serviceport{{PortName: "foo1", Protocol: corev1.ProtocolTCP, PortNumber: 8081, TargetPort: intstr.FromInt(8081)}},
	}.Service()
	svcUpdate.ResourceVersion = "2"
	if _, err := KubeClient.CoreV1().Services(ns).Update(context.TODO(), svcUpdate, metav1.UpdateOptions{}); err != nil {
		t.Fatalf("error in updating Service: %v", err)
	}

	g.Eventually(func() bool {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName1)
		if found && aviModel != nil {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
			if len(nodes[0].PoolRefs) == 0 {
				return true
			}
		}
		return false
	}, 40*time.Second).Should(gomega.Equal(true))

	g.Eventually(func() bool {
		if found, aviModel := objects.SharedAviGraphLister().Get(modelName2); found && aviModel != nil {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
			if len(nodes) > 0 && nodes[0].Name == "cluster--default-my-gateway2" {
				return true
			}
		}
		return false
	}, 40*time.Second).Should(gomega.Equal(true))

	TeardownAdvLBService(t, "svc", ns)
	TeardownGateway(t, gateway1Name, ns)
	TeardownGateway(t, gateway2Name, ns)
	TeardownGatewayClass(t, gwClassName)
	VerifyGatewayVSNodeDeletion(g, modelName2)
}

func TestServicesAPIEndpointDeleteCreate(t *testing.T) {
	// svc/tcp/8081, gw1/tcp/8081
	// delete/create endpoints
	g := gomega.NewGomegaWithT(t)

	gwClassName, gatewayName, ns := "avi-lb", "my-gateway", "default"
	modelName := "admin/cluster--default-my-gateway"

	SetupGatewayClass(t, gwClassName, lib.SvcApiAviGatewayController, "")
	SetupGateway(t, gatewayName, ns, gwClassName)
	SetupSvcApiService(t, "svc", ns, gatewayName, ns, "TCP")

	// delete endpoints
	integrationtest.DelEP(t, ns, "svc")
	g.Eventually(func() bool {
		if found, aviModel := objects.SharedAviGraphLister().Get(modelName); found && aviModel != nil {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
			if len(nodes) > 0 &&
				len(nodes[0].PoolRefs) == 1 &&
				len(nodes[0].PoolRefs[0].Servers) == 0 {
				return true
			}
		}
		return false
	}, 40*time.Second).Should(gomega.Equal(true))

	// create new endpoints
	newIP := "2.2.2"
	integrationtest.CreateEP(t, ns, "svc", false, true, newIP)
	g.Eventually(func() bool {
		if found, aviModel := objects.SharedAviGraphLister().Get(modelName); found && aviModel != nil {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
			if len(nodes) > 0 &&
				len(nodes[0].PoolRefs) == 1 &&
				len(nodes[0].PoolRefs[0].Servers) == 3 {
				return true
			}
		}
		return false
	}, 40*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes[0].PoolRefs).Should(gomega.HaveLen(1))
	g.Expect(*nodes[0].PoolRefs[0].Servers[0].Ip.Addr).Should(gomega.ContainSubstring(newIP))

	TeardownAdvLBService(t, "svc", ns)
	TeardownGateway(t, gatewayName, ns)
	TeardownGatewayClass(t, gwClassName)
	VerifyGatewayVSNodeDeletion(g, modelName)
}

func TestServicesAPIMultiServiceMultiProtocol(t *testing.T) {
	// svc1/tcp/8081, svc2/udp/8082, gw1/tcp/8081 - gw1/udp/8082
	// creates services with network profile overrides (can't check).
	// remove udp listener from gateway, check for status delete on svc2.

	g := gomega.NewGomegaWithT(t)

	gwClassName, gatewayName, ns := "avi-lb", "my-gateway", "default"
	svcName1, svcName2 := "svc1", "svc2"
	modelName := "admin/cluster--default-my-gateway"
	labels := map[string]string{lib.SvcApiGatewayNameLabelKey: gatewayName, lib.SvcApiGatewayNamespaceLabelKey: ns}

	SetupGatewayClass(t, gwClassName, lib.SvcApiAviGatewayController, "")

	gateway := FakeGateway{
		Name:      gatewayName,
		Namespace: ns,
		GWClass:   gwClassName,
		Listeners: []FakeGWListener{
			{Port: 8081, Protocol: "TCP", Labels: labels},
			{Port: 8082, Protocol: "UDP", Labels: labels},
		},
	}
	if _, err := lib.AKOControlConfig().ServicesAPIClientset().NetworkingV1alpha1().Gateways(ns).Create(context.TODO(), gateway.Gateway(), metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding Gateway: %v", err)
	}

	svc1 := integrationtest.FakeService{
		Name:         svcName1,
		Namespace:    ns,
		Labels:       labels,
		Type:         corev1.ServiceTypeClusterIP,
		ServicePorts: []integrationtest.Serviceport{{PortName: "footcp", Protocol: "TCP", PortNumber: 8081, TargetPort: intstr.FromInt(80)}},
	}
	if _, err := KubeClient.CoreV1().Services(ns).Create(context.TODO(), svc1.Service(), metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding Service: %v", err)
	}
	integrationtest.CreateEP(t, ns, svcName1, false, true, "1.1.1")

	svc2 := integrationtest.FakeService{
		Name:         svcName2,
		Namespace:    ns,
		Labels:       labels,
		Type:         corev1.ServiceTypeClusterIP,
		ServicePorts: []integrationtest.Serviceport{{PortName: "fooudp", Protocol: "UDP", PortNumber: 8082, TargetPort: intstr.FromInt(80)}},
	}

	if _, err := KubeClient.CoreV1().Services(ns).Create(context.TODO(), svc2.Service(), metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding Service: %v", err)
	}

	integrationtest.CreateEP(t, ns, svcName2, false, true, "1.1.1")

	g.Eventually(func() string {
		gw, _ := SvcAPIClient.NetworkingV1alpha1().Gateways(ns).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		if len(gw.Status.Addresses) > 0 {
			return gw.Status.Addresses[0].Value
		}
		return ""
	}, 40*time.Second).Should(gomega.Equal("10.250.250.1"))

	g.Eventually(func() string {
		svc1, _ := KubeClient.CoreV1().Services(ns).Get(context.TODO(), svcName1, metav1.GetOptions{})
		svc2, _ := KubeClient.CoreV1().Services(ns).Get(context.TODO(), svcName2, metav1.GetOptions{})
		if len(svc1.Status.LoadBalancer.Ingress) > 0 &&
			len(svc2.Status.LoadBalancer.Ingress) > 0 &&
			svc1.Status.LoadBalancer.Ingress[0].IP == svc2.Status.LoadBalancer.Ingress[0].IP {
			return svc1.Status.LoadBalancer.Ingress[0].IP
		}
		return ""
	}, 30*time.Second).Should(gomega.Equal("10.250.250.1"))

	gatewayUpdate := FakeGateway{
		Name:      gatewayName,
		Namespace: ns,
		GWClass:   gwClassName,
		Listeners: []FakeGWListener{
			{Port: 8081, Protocol: "TCP", Labels: labels},
		},
	}.Gateway()
	gatewayUpdate.ResourceVersion = "2"
	if _, err := lib.AKOControlConfig().ServicesAPIClientset().NetworkingV1alpha1().Gateways(ns).Update(context.TODO(), gatewayUpdate, metav1.UpdateOptions{}); err != nil {
		t.Fatalf("error in updating Gateway: %v", err)
	}

	g.Eventually(func() int {
		svc2, _ := KubeClient.CoreV1().Services(ns).Get(context.TODO(), svcName2, metav1.GetOptions{})
		return len(svc2.Status.LoadBalancer.Ingress)
	}, 30*time.Second).Should(gomega.Equal(0))

	TeardownAdvLBService(t, svcName1, ns)
	TeardownAdvLBService(t, svcName2, ns)
	TeardownGateway(t, gatewayName, ns)
	TeardownGatewayClass(t, gwClassName)
	VerifyGatewayVSNodeDeletion(g, modelName)
}

// AviInfraSetting tests

func TestServicesAPIWithInfraSettingStatusUpdates(t *testing.T) {
	// create infraSetting, gwclass, gw with bad seGroup/networkName
	// check for Rejected status, check layer 2 for defaults
	// change to good seGroup/networkName, check for Accepted status
	// check layer 2 model

	g := gomega.NewGomegaWithT(t)

	gwClassName, gatewayName, settingName, ns := "avi-lb", "my-gateway", "infra-setting", "default"
	modelName := "admin/cluster--default-my-gateway"

	SetupGatewayClass(t, gwClassName, lib.SvcApiAviGatewayController, settingName)
	SetupGateway(t, gatewayName, ns, gwClassName)
	SetupSvcApiService(t, "svc", ns, gatewayName, ns, "TCP")

	// Create with bad seGroup ref.
	settingCreate := (integrationtest.FakeAviInfraSetting{
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
	}, 10*time.Second).Should(gomega.Equal("Rejected"))

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

	settingUpdate := (integrationtest.FakeAviInfraSetting{
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
	}, 45*time.Second).Should(gomega.Equal("Accepted"))

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

	settingUpdate = (integrationtest.FakeAviInfraSetting{
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

	TeardownAdvLBService(t, "svc", ns)
	TeardownGateway(t, gatewayName, ns)
	TeardownGatewayClass(t, gwClassName)
	VerifyGatewayVSNodeDeletion(g, modelName)
	integrationtest.TeardownAviInfraSetting(t, settingName)
}

func TestServicesAPInfraSettingDelete(t *testing.T) {
	// create infraSetting, gwclass, gw
	// delete infraSetting, fallback to defaults

	g := gomega.NewGomegaWithT(t)

	gwClassName, gatewayName, settingName, ns := "avi-lb", "my-gateway", "infra-setting", "default"
	modelName := "admin/cluster--default-my-gateway"

	SetupGatewayClass(t, gwClassName, lib.SvcApiAviGatewayController, settingName)
	SetupGateway(t, gatewayName, ns, gwClassName)
	SetupSvcApiService(t, "svc", ns, gatewayName, ns, "TCP")
	integrationtest.SetupAviInfraSetting(t, settingName, "")

	g.Eventually(func() bool {
		if found, aviModel := objects.SharedAviGraphLister().Get(modelName); found && aviModel != nil {
			if nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS(); len(nodes) > 0 {
				return nodes[0].ServiceEngineGroup == "thisisaviref-"+settingName+"-seGroup" &&
					len(nodes[0].VSVIPRefs[0].VipNetworks) > 0 &&
					nodes[0].VSVIPRefs[0].VipNetworks[0].NetworkName == "thisisaviref-"+settingName+"-networkName" &&
					*nodes[0].EnableRhi
			}
		}
		return false
	}, 45*time.Second).Should(gomega.Equal(true))

	integrationtest.TeardownAviInfraSetting(t, settingName)

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

	TeardownAdvLBService(t, "svc", ns)
	TeardownGateway(t, gatewayName, ns)
	TeardownGatewayClass(t, gwClassName)
	VerifyGatewayVSNodeDeletion(g, modelName)
}

func TestServicesAPIInfraSettingChangeMapping(t *testing.T) {
	// create 2 infraSettings, gwclass, gw
	// update infraSetting from one to another in gwclass
	// check changed model

	g := gomega.NewGomegaWithT(t)

	gwClassName, gatewayName, settingName1, settingName2, ns := "avi-lb", "my-gateway", "infra-setting1", "infra-setting2", "default"
	modelName := "admin/cluster--default-my-gateway"

	SetupGatewayClass(t, gwClassName, lib.SvcApiAviGatewayController, settingName1)
	SetupGateway(t, gatewayName, ns, gwClassName)
	SetupSvcApiService(t, "svc", ns, gatewayName, ns, "TCP")
	integrationtest.SetupAviInfraSetting(t, settingName1, "")
	integrationtest.SetupAviInfraSetting(t, settingName2, "")

	g.Eventually(func() bool {
		if found, aviModel := objects.SharedAviGraphLister().Get(modelName); found && aviModel != nil {
			if nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS(); len(nodes) > 0 {
				return nodes[0].ServiceEngineGroup == "thisisaviref-"+settingName1+"-seGroup" &&
					len(nodes[0].VSVIPRefs[0].VipNetworks) > 0 &&
					nodes[0].VSVIPRefs[0].VipNetworks[0].NetworkName == "thisisaviref-"+settingName1+"-networkName"
			}
		}
		return false
	}, 45*time.Second).Should(gomega.Equal(true))

	// Change gatewayclass to have infraSettting2
	gwClassUpdate := (FakeGWClass{
		Name:         gwClassName,
		Controller:   lib.SvcApiAviGatewayController,
		InfraSetting: settingName2,
	}).GatewayClass()
	gwClassUpdate.ResourceVersion = "2"
	if _, err := lib.AKOControlConfig().ServicesAPIClientset().NetworkingV1alpha1().GatewayClasses().Update(context.TODO(), gwClassUpdate, metav1.UpdateOptions{}); err != nil {
		t.Fatalf("error in updating GatewayClass: %v", err)
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
	}, 45*time.Second).Should(gomega.Equal(true))

	TeardownAdvLBService(t, "svc", ns)
	TeardownGateway(t, gatewayName, ns)
	TeardownGatewayClass(t, gwClassName)
	integrationtest.TeardownAviInfraSetting(t, settingName1)
	integrationtest.TeardownAviInfraSetting(t, settingName2)
	VerifyGatewayVSNodeDeletion(g, modelName)
}

func TestServicesAPINetworkProfileBasedOnLicense(t *testing.T) {
	// create gwclass, create gw, create 1svc
	// check graph VsNode vals, check IP status
	// check whether the `NetworkProfile` is based on the license
	// remove gwclasss, IP removed
	g := gomega.NewGomegaWithT(t)

	gwClassName, gatewayName, ns := "avi-lb", "my-gateway", "default"
	modelName := "admin/cluster--default-my-gateway"

	SetupLicense := func(license string) {
		integrationtest.AviFakeClientInstance = nil
		integrationtest.AddMiddleware(func(w http.ResponseWriter, r *http.Request) {
			url := r.URL.EscapedPath()
			if strings.Contains(url, "/api/systemconfiguration") {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"default_license_tier": "` + license + `"}`))
				return
			}
			integrationtest.NormalControllerServer(w, r)
		})
		integrationtest.NewAviFakeClientInstance(KubeClient)

		// Set the license
		aviRestClientPool := cache.SharedAVIClients(lib.GetTenant())
		lib.AKOControlConfig().SetLicenseType(aviRestClientPool.AviClient[0])
	}

	SetupLicense(lib.LicenseTypeEnterprise)
	SetupGatewayClass(t, gwClassName, lib.SvcApiAviGatewayController, "")
	SetupGateway(t, gatewayName, ns, gwClassName)

	SetupSvcApiService(t, "svc", ns, gatewayName, ns, "TCP")

	g.Eventually(func() string {
		gw, _ := SvcAPIClient.NetworkingV1alpha1().Gateways(ns).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		if len(gw.Status.Addresses) > 0 {
			return gw.Status.Addresses[0].Value
		}
		return ""
	}, 40*time.Second).Should(gomega.Equal("10.250.250.1"))

	g.Eventually(func() string {
		svc, _ := KubeClient.CoreV1().Services(ns).Get(context.TODO(), "svc", metav1.GetOptions{})
		if len(svc.Status.LoadBalancer.Ingress) > 0 {
			return svc.Status.LoadBalancer.Ingress[0].IP
		}
		return ""
	}, 30*time.Second).Should(gomega.Equal("10.250.250.1"))

	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes[0].PortProto[0].Port).To(gomega.Equal(int32(8081)))
	g.Expect(nodes[0].HttpPolicySetRefs).To(gomega.HaveLen(0))
	g.Expect(nodes[0].L4PolicyRefs).To(gomega.HaveLen(1))
	g.Expect(nodes[0].L4PolicyRefs[0].PortPool[0].Port).To(gomega.Equal(uint32(8081)))
	g.Expect(nodes[0].L4PolicyRefs[0].PortPool[0].Protocol).To(gomega.Equal("TCP"))
	g.Expect(nodes[0].PoolRefs[0].ServiceMetadata.NamespaceServiceName[0]).To(gomega.Equal("default/svc"))
	g.Expect(nodes[0].ServiceMetadata.Gateway).To(gomega.Equal("default/my-gateway"))
	g.Expect(nodes[0].PoolRefs[0].Servers).To(gomega.HaveLen(3))
	g.Expect(nodes[0].NetworkProfile).To(gomega.Equal(utils.DEFAULT_TCP_NW_PROFILE))

	TeardownGatewayClass(t, gwClassName)
	g.Eventually(func() int {
		gw, _ := SvcAPIClient.NetworkingV1alpha1().Gateways(ns).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		return len(gw.Status.Addresses)
	}, 40*time.Second).Should(gomega.Equal(0))

	TeardownAdvLBService(t, "svc", ns)
	TeardownGateway(t, gatewayName, ns)
	VerifyGatewayVSNodeDeletion(g, modelName)

	// Set the license as BASIC and verify the network profile.
	SetupLicense("BASIC")
	SetupGatewayClass(t, gwClassName, lib.SvcApiAviGatewayController, "")
	SetupGateway(t, gatewayName, ns, gwClassName)

	SetupSvcApiService(t, "svc", ns, gatewayName, ns, "TCP")

	g.Eventually(func() string {
		gw, _ := SvcAPIClient.NetworkingV1alpha1().Gateways(ns).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		if len(gw.Status.Addresses) > 0 {
			return gw.Status.Addresses[0].Value
		}
		return ""
	}, 40*time.Second).Should(gomega.Equal("10.250.250.1"))

	g.Eventually(func() string {
		svc, _ := KubeClient.CoreV1().Services(ns).Get(context.TODO(), "svc", metav1.GetOptions{})
		if len(svc.Status.LoadBalancer.Ingress) > 0 {
			return svc.Status.LoadBalancer.Ingress[0].IP
		}
		return ""
	}, 30*time.Second).Should(gomega.Equal("10.250.250.1"))

	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes[0].PortProto[0].Port).To(gomega.Equal(int32(8081)))
	g.Expect(nodes[0].HttpPolicySetRefs).To(gomega.HaveLen(0))
	g.Expect(nodes[0].L4PolicyRefs).To(gomega.HaveLen(1))
	g.Expect(nodes[0].L4PolicyRefs[0].PortPool[0].Port).To(gomega.Equal(uint32(8081)))
	g.Expect(nodes[0].L4PolicyRefs[0].PortPool[0].Protocol).To(gomega.Equal("TCP"))
	g.Expect(nodes[0].PoolRefs[0].ServiceMetadata.NamespaceServiceName[0]).To(gomega.Equal("default/svc"))
	g.Expect(nodes[0].ServiceMetadata.Gateway).To(gomega.Equal("default/my-gateway"))
	g.Expect(nodes[0].PoolRefs[0].Servers).To(gomega.HaveLen(3))
	g.Expect(nodes[0].NetworkProfile).To(gomega.Equal(utils.TCP_NW_FAST_PATH))

	TeardownGatewayClass(t, gwClassName)
	g.Eventually(func() int {
		gw, _ := SvcAPIClient.NetworkingV1alpha1().Gateways(ns).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		return len(gw.Status.Addresses)
	}, 40*time.Second).Should(gomega.Equal(0))

	TeardownAdvLBService(t, "svc", ns)
	TeardownGateway(t, gatewayName, ns)
	VerifyGatewayVSNodeDeletion(g, modelName)

	integrationtest.AviFakeClientInstance.Close()
}

func TestServicesAPIMutliProtocol(t *testing.T) {
	// create gwclass, create gw, create 1svc
	// check graph VsNode vals, check IP status
	// remove gwclasss, IP removed
	g := gomega.NewGomegaWithT(t)

	gwClassName, gatewayName, ns := "avi-lb", "my-gateway", "default"
	modelName := "admin/cluster--default-my-gateway"
	protocols := []string{"SCTP", "TCP", "UDP"}
	var namespacedSvcList []string
	for _, protocol := range protocols {
		namespacedSvcList = append(namespacedSvcList, ns+"/"+"svc"+"-"+protocol)
	}

	SetupGatewayClass(t, gwClassName, lib.SvcApiAviGatewayController, "")
	SetupGateway(t, gatewayName, ns, gwClassName, protocols...)

	for _, protocol := range protocols {
		svcname := "svc" + "-" + protocol
		SetupSvcApiService(t, svcname, ns, gatewayName, ns, protocol)
	}

	g.Eventually(func() string {
		gw, _ := SvcAPIClient.NetworkingV1alpha1().Gateways(ns).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		if len(gw.Status.Addresses) > 0 {
			return gw.Status.Addresses[0].Value
		}
		return ""
	}, 40*time.Second).Should(gomega.Equal("10.250.250.1"))

	for _, protocol := range protocols {
		svcname := "svc" + "-" + protocol
		g.Eventually(func() string {
			svc, _ := KubeClient.CoreV1().Services(ns).Get(context.TODO(), svcname, metav1.GetOptions{})
			if len(svc.Status.LoadBalancer.Ingress) > 0 {
				return svc.Status.LoadBalancer.Ingress[0].IP
			}
			return ""
		}, 30*time.Second).Should(gomega.Equal("10.250.250.1"))
	}

	for i, protocol := range protocols {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(nodes[0].PortProto).To(gomega.HaveLen(len(protocols)))
		g.Expect(nodes[0].PortProto[i].Port).To(gomega.Equal(int32(8081)))
		g.Expect(nodes[0].PortProto[i].Protocol).To(gomega.Equal(protocol))
		g.Expect(nodes[0].HttpPolicySetRefs).To(gomega.HaveLen(0))
		g.Expect(nodes[0].L4PolicyRefs).To(gomega.HaveLen(1))
		g.Expect(nodes[0].L4PolicyRefs[0].PortPool).To(gomega.HaveLen(len(protocols)))
		g.Expect(nodes[0].L4PolicyRefs[0].PortPool[i].Port).To(gomega.Equal(uint32(8081)))
		g.Expect(nodes[0].L4PolicyRefs[0].PortPool[i].Protocol).To(gomega.BeElementOf(protocols))
		g.Expect(nodes[0].PoolRefs).To(gomega.HaveLen(len(protocols)))
		g.Expect(nodes[0].PoolRefs[i].ServiceMetadata.NamespaceServiceName[0]).To(gomega.BeElementOf(namespacedSvcList))
		g.Expect(nodes[0].ServiceMetadata.Gateway).To(gomega.Equal("default/my-gateway"))
		g.Expect(nodes[0].PoolRefs[i].Servers).To(gomega.HaveLen(3))
	}

	TeardownGatewayClass(t, gwClassName)
	g.Eventually(func() int {
		gw, _ := SvcAPIClient.NetworkingV1alpha1().Gateways(ns).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		return len(gw.Status.Addresses)
	}, 40*time.Second).Should(gomega.Equal(0))

	for _, protocol := range protocols {
		svcname := "svc" + "-" + protocol
		TeardownAdvLBService(t, svcname, ns)
	}
	TeardownGateway(t, gatewayName, ns)
	VerifyGatewayVSNodeDeletion(g, modelName)
}

func TestServicesAPIMutliProtocolSCTPTCP(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	// middleware verifies the application and network profiles attached to the VS
	integrationtest.AddMiddleware(func(w http.ResponseWriter, r *http.Request) {
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
		integrationtest.NormalControllerServer(w, r)
	})

	gwClassName, gatewayName, ns := "avi-lb", "my-gateway", "default"
	modelName := "admin/cluster--default-my-gateway"
	protocols := []string{"SCTP", "TCP"}
	var namespacedSvcList []string
	for _, protocol := range protocols {
		namespacedSvcList = append(namespacedSvcList, ns+"/"+"svc"+"-"+protocol)
	}

	SetupGatewayClass(t, gwClassName, lib.SvcApiAviGatewayController, "")
	SetupGateway(t, gatewayName, ns, gwClassName, protocols...)

	for _, protocol := range protocols {
		svcname := "svc" + "-" + protocol
		SetupSvcApiService(t, svcname, ns, gatewayName, ns, protocol)
	}

	g.Eventually(func() string {
		gw, _ := SvcAPIClient.NetworkingV1alpha1().Gateways(ns).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		if len(gw.Status.Addresses) > 0 {
			return gw.Status.Addresses[0].Value
		}
		return ""
	}, 40*time.Second).Should(gomega.Equal("10.250.250.1"))

	for _, protocol := range protocols {
		svcname := "svc" + "-" + protocol
		g.Eventually(func() string {
			svc, _ := KubeClient.CoreV1().Services(ns).Get(context.TODO(), svcname, metav1.GetOptions{})
			if len(svc.Status.LoadBalancer.Ingress) > 0 {
				return svc.Status.LoadBalancer.Ingress[0].IP
			}
			return ""
		}, 30*time.Second).Should(gomega.Equal("10.250.250.1"))
	}

	for i, protocol := range protocols {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(nodes[0].PortProto).To(gomega.HaveLen(len(protocols)))
		g.Expect(nodes[0].PortProto[i].Port).To(gomega.Equal(int32(8081)))
		g.Expect(nodes[0].PortProto[i].Protocol).To(gomega.Equal(protocol))
		g.Expect(nodes[0].HttpPolicySetRefs).To(gomega.HaveLen(0))
		g.Expect(nodes[0].L4PolicyRefs).To(gomega.HaveLen(1))
		g.Expect(nodes[0].L4PolicyRefs[0].PortPool).To(gomega.HaveLen(len(protocols)))
		g.Expect(nodes[0].L4PolicyRefs[0].PortPool[i].Port).To(gomega.Equal(uint32(8081)))
		g.Expect(nodes[0].L4PolicyRefs[0].PortPool[i].Protocol).To(gomega.BeElementOf(protocols))
		g.Expect(nodes[0].PoolRefs).To(gomega.HaveLen(len(protocols)))
		g.Expect(nodes[0].PoolRefs[i].ServiceMetadata.NamespaceServiceName[0]).To(gomega.BeElementOf(namespacedSvcList))
		g.Expect(nodes[0].ServiceMetadata.Gateway).To(gomega.Equal("default/my-gateway"))
		g.Expect(nodes[0].PoolRefs[i].Servers).To(gomega.HaveLen(3))
	}

	TeardownGatewayClass(t, gwClassName)
	g.Eventually(func() int {
		gw, _ := SvcAPIClient.NetworkingV1alpha1().Gateways(ns).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		return len(gw.Status.Addresses)
	}, 40*time.Second).Should(gomega.Equal(0))

	for _, protocol := range protocols {
		svcname := "svc" + "-" + protocol
		TeardownAdvLBService(t, svcname, ns)
	}
	TeardownGateway(t, gatewayName, ns)
	VerifyGatewayVSNodeDeletion(g, modelName)
}

func TestServicesAPIMutliProtocolSCTPUDP(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	// middleware verifies the application and network profiles attached to the VS
	integrationtest.AddMiddleware(func(w http.ResponseWriter, r *http.Request) {
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
		integrationtest.NormalControllerServer(w, r)
	})

	gwClassName, gatewayName, ns := "avi-lb", "my-gateway", "default"
	modelName := "admin/cluster--default-my-gateway"
	protocols := []string{"SCTP", "UDP"}
	var namespacedSvcList []string
	for _, protocol := range protocols {
		namespacedSvcList = append(namespacedSvcList, ns+"/"+"svc"+"-"+protocol)
	}

	SetupGatewayClass(t, gwClassName, lib.SvcApiAviGatewayController, "")
	SetupGateway(t, gatewayName, ns, gwClassName, protocols...)

	for _, protocol := range protocols {
		svcname := "svc" + "-" + protocol
		SetupSvcApiService(t, svcname, ns, gatewayName, ns, protocol)
	}

	g.Eventually(func() string {
		gw, _ := SvcAPIClient.NetworkingV1alpha1().Gateways(ns).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		if len(gw.Status.Addresses) > 0 {
			return gw.Status.Addresses[0].Value
		}
		return ""
	}, 40*time.Second).Should(gomega.Equal("10.250.250.1"))

	for _, protocol := range protocols {
		svcname := "svc" + "-" + protocol
		g.Eventually(func() string {
			svc, _ := KubeClient.CoreV1().Services(ns).Get(context.TODO(), svcname, metav1.GetOptions{})
			if len(svc.Status.LoadBalancer.Ingress) > 0 {
				return svc.Status.LoadBalancer.Ingress[0].IP
			}
			return ""
		}, 30*time.Second).Should(gomega.Equal("10.250.250.1"))
	}

	for i, protocol := range protocols {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(nodes[0].PortProto).To(gomega.HaveLen(len(protocols)))
		g.Expect(nodes[0].PortProto[i].Port).To(gomega.Equal(int32(8081)))
		g.Expect(nodes[0].PortProto[i].Protocol).To(gomega.Equal(protocol))
		g.Expect(nodes[0].HttpPolicySetRefs).To(gomega.HaveLen(0))
		g.Expect(nodes[0].L4PolicyRefs).To(gomega.HaveLen(1))
		g.Expect(nodes[0].L4PolicyRefs[0].PortPool).To(gomega.HaveLen(len(protocols)))
		g.Expect(nodes[0].L4PolicyRefs[0].PortPool[i].Port).To(gomega.Equal(uint32(8081)))
		g.Expect(nodes[0].L4PolicyRefs[0].PortPool[i].Protocol).To(gomega.BeElementOf(protocols))
		g.Expect(nodes[0].PoolRefs).To(gomega.HaveLen(len(protocols)))
		g.Expect(nodes[0].PoolRefs[i].ServiceMetadata.NamespaceServiceName[0]).To(gomega.BeElementOf(namespacedSvcList))
		g.Expect(nodes[0].ServiceMetadata.Gateway).To(gomega.Equal("default/my-gateway"))
		g.Expect(nodes[0].PoolRefs[i].Servers).To(gomega.HaveLen(3))
	}

	TeardownGatewayClass(t, gwClassName)
	g.Eventually(func() int {
		gw, _ := SvcAPIClient.NetworkingV1alpha1().Gateways(ns).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		return len(gw.Status.Addresses)
	}, 40*time.Second).Should(gomega.Equal(0))

	for _, protocol := range protocols {
		svcname := "svc" + "-" + protocol
		TeardownAdvLBService(t, svcname, ns)
	}
	TeardownGateway(t, gatewayName, ns)
	VerifyGatewayVSNodeDeletion(g, modelName)
}

func TestServicesAPIMutliProtocolTCPUDP(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	// middleware verifies the application and network profiles attached to the VS
	integrationtest.AddMiddleware(func(w http.ResponseWriter, r *http.Request) {
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
		integrationtest.NormalControllerServer(w, r)
	})

	gwClassName, gatewayName, ns := "avi-lb", "my-gateway", "default"
	modelName := "admin/cluster--default-my-gateway"
	protocols := []string{"TCP", "UDP"}
	var namespacedSvcList []string
	for _, protocol := range protocols {
		namespacedSvcList = append(namespacedSvcList, ns+"/"+"svc"+"-"+protocol)
	}

	SetupGatewayClass(t, gwClassName, lib.SvcApiAviGatewayController, "")
	SetupGateway(t, gatewayName, ns, gwClassName, protocols...)

	for _, protocol := range protocols {
		svcname := "svc" + "-" + protocol
		SetupSvcApiService(t, svcname, ns, gatewayName, ns, protocol)
	}

	g.Eventually(func() string {
		gw, _ := SvcAPIClient.NetworkingV1alpha1().Gateways(ns).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		if len(gw.Status.Addresses) > 0 {
			return gw.Status.Addresses[0].Value
		}
		return ""
	}, 40*time.Second).Should(gomega.Equal("10.250.250.1"))

	for _, protocol := range protocols {
		svcname := "svc" + "-" + protocol
		g.Eventually(func() string {
			svc, _ := KubeClient.CoreV1().Services(ns).Get(context.TODO(), svcname, metav1.GetOptions{})
			if len(svc.Status.LoadBalancer.Ingress) > 0 {
				return svc.Status.LoadBalancer.Ingress[0].IP
			}
			return ""
		}, 30*time.Second).Should(gomega.Equal("10.250.250.1"))
	}

	for i, protocol := range protocols {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(nodes[0].PortProto).To(gomega.HaveLen(len(protocols)))
		g.Expect(nodes[0].PortProto[i].Port).To(gomega.Equal(int32(8081)))
		g.Expect(nodes[0].PortProto[i].Protocol).To(gomega.Equal(protocol))
		g.Expect(nodes[0].HttpPolicySetRefs).To(gomega.HaveLen(0))
		g.Expect(nodes[0].L4PolicyRefs).To(gomega.HaveLen(1))
		g.Expect(nodes[0].L4PolicyRefs[0].PortPool).To(gomega.HaveLen(len(protocols)))
		g.Expect(nodes[0].L4PolicyRefs[0].PortPool[i].Port).To(gomega.Equal(uint32(8081)))
		g.Expect(nodes[0].L4PolicyRefs[0].PortPool[i].Protocol).To(gomega.BeElementOf(protocols))
		g.Expect(nodes[0].PoolRefs).To(gomega.HaveLen(len(protocols)))
		g.Expect(nodes[0].PoolRefs[i].ServiceMetadata.NamespaceServiceName[0]).To(gomega.BeElementOf(namespacedSvcList))
		g.Expect(nodes[0].ServiceMetadata.Gateway).To(gomega.Equal("default/my-gateway"))
		g.Expect(nodes[0].PoolRefs[i].Servers).To(gomega.HaveLen(3))
	}

	TeardownGatewayClass(t, gwClassName)
	g.Eventually(func() int {
		gw, _ := SvcAPIClient.NetworkingV1alpha1().Gateways(ns).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		return len(gw.Status.Addresses)
	}, 40*time.Second).Should(gomega.Equal(0))

	for _, protocol := range protocols {
		svcname := "svc" + "-" + protocol
		TeardownAdvLBService(t, svcname, ns)
	}
	TeardownGateway(t, gatewayName, ns)
	VerifyGatewayVSNodeDeletion(g, modelName)
}
