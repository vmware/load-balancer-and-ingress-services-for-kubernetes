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

package advl4tests

import (
	"context"
	"os"
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

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/integrationtest"
	advl4fake "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/third_party/service-apis/client/clientset/versioned/fake"

	"github.com/onsi/gomega"
	advl4v1alpha1pre1 "github.com/vmware-tanzu/service-apis/apis/v1alpha1pre1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

var KubeClient *k8sfake.Clientset
var AdvL4Client *advl4fake.Clientset
var CRDClient *crdfake.Clientset
var V1beta1CRDClient *v1beta1crdfake.Clientset
var ctrl *k8s.AviController

func TestMain(m *testing.M) {
	os.Setenv("CLUSTER_ID", "abc:cluster")
	os.Setenv("CLOUD_NAME", "Default-Cloud")
	os.Setenv("ADVANCED_L4", "true")
	os.Setenv("POD_NAMESPACE", utils.AKO_DEFAULT_NS)
	os.Setenv("SHARD_VS_SIZE", "LARGE")
	os.Setenv("POD_NAME", "ako-0")

	akoControlConfig := lib.AKOControlConfig()
	KubeClient = k8sfake.NewSimpleClientset()
	AdvL4Client = advl4fake.NewSimpleClientset()
	CRDClient = crdfake.NewSimpleClientset()
	V1beta1CRDClient = v1beta1crdfake.NewSimpleClientset()
	akoControlConfig.SetAKOInstanceFlag(true)
	akoControlConfig.SetAdvL4Clientset(AdvL4Client)
	akoControlConfig.Setv1beta1CRDClientset(V1beta1CRDClient)
	akoControlConfig.SetCRDClientsetAndEnableInfraSettingParam(V1beta1CRDClient)
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
		utils.EndpointInformer,
		utils.SecretInformer,
		utils.NSInformer,
		utils.ConfigMapInformer,
	}
	utils.NewInformers(utils.KubeClientIntf{ClientSet: KubeClient}, registeredInformers)
	informers := k8s.K8sinformers{Cs: KubeClient}
	k8s.NewCRDInformers()
	k8s.NewAdvL4Informers(AdvL4Client)

	mcache := cache.SharedAviObjCache()
	cloudObj := &cache.AviCloudPropertyCache{Name: "Default-Cloud", VType: "mock"}
	subdomains := []string{"avi.internal", ".com"}
	cloudObj.NSIpamDNS = subdomains
	mcache.CloudKeyCache.AviCacheAdd("Default-Cloud", cloudObj)

	integrationtest.KubeClient = KubeClient
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

	ctrl.HandleConfigMap(informers, ctrlCh, stopCh, quickSyncCh)
	integrationtest.AddDefaultNamespace()
	go ctrl.InitController(informers, registeredInformers, ctrlCh, stopCh, quickSyncCh, waitGroupMap)
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
	Port     int32
	Protocol string
	Labels   map[string]string
}

func (gw FakeGateway) Gateway() *advl4v1alpha1pre1.Gateway {
	var fakeListeners []advl4v1alpha1pre1.Listener
	for _, listener := range gw.Listeners {
		fakeListeners = append(fakeListeners, advl4v1alpha1pre1.Listener{
			Port:     listener.Port,
			Protocol: advl4v1alpha1pre1.ProtocolType(listener.Protocol),
			Routes: advl4v1alpha1pre1.RouteBindingSelector{
				Resource: "services",
				RouteSelector: metav1.LabelSelector{
					MatchLabels: listener.Labels,
				},
			},
		})
	}

	gateway := &advl4v1alpha1pre1.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: gw.Namespace,
			Name:      gw.Name,
		},
		Spec: advl4v1alpha1pre1.GatewaySpec{
			Class:     gw.GWClass,
			Listeners: fakeListeners,
		},
	}

	if gw.IPAddress != "" {
		gateway.Spec.Addresses = []advl4v1alpha1pre1.GatewayAddress{{
			Type:  advl4v1alpha1pre1.IPAddressType,
			Value: gw.IPAddress,
		}}
	}

	return gateway
}

func SetupGateway(t *testing.T, gwname, namespace, gwclass string) {
	gateway := FakeGateway{
		Name:      gwname,
		Namespace: namespace,
		GWClass:   gwclass,
		Listeners: []FakeGWListener{{
			Port:     int32(8081),
			Protocol: "TCP",
			Labels: map[string]string{
				lib.GatewayNameLabelKey:      gwname,
				lib.GatewayNamespaceLabelKey: namespace,
				lib.GatewayTypeLabelKey:      "direct",
			},
		}},
	}

	gwCreate := gateway.Gateway()
	if _, err := lib.AKOControlConfig().AdvL4Clientset().NetworkingV1alpha1pre1().Gateways(namespace).Create(context.TODO(), gwCreate, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding Gateway: %v", err)
	}
}

func TeardownGateway(t *testing.T, gwname, namespace string) {
	if err := lib.AKOControlConfig().AdvL4Clientset().NetworkingV1alpha1pre1().Gateways(namespace).Delete(context.TODO(), gwname, metav1.DeleteOptions{}); err != nil {
		t.Fatalf("error in deleting Gateway: %v", err)
	}
}

type FakeGWClass struct {
	Name       string
	Controller string
}

func (gwclass FakeGWClass) GatewayClass() *advl4v1alpha1pre1.GatewayClass {
	gatewayclass := &advl4v1alpha1pre1.GatewayClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: gwclass.Name,
		},
		Spec: advl4v1alpha1pre1.GatewayClassSpec{
			Controller: gwclass.Controller,
		},
	}

	return gatewayclass
}

func SetupGatewayClass(t *testing.T, gwclassName, controller string) {
	gatewayclass := FakeGWClass{
		Name:       gwclassName,
		Controller: controller,
	}

	gwClassCreate := gatewayclass.GatewayClass()
	if _, err := lib.AKOControlConfig().AdvL4Clientset().NetworkingV1alpha1pre1().GatewayClasses().Create(context.TODO(), gwClassCreate, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding GatewayClass: %v", err)
	}
}

func TeardownGatewayClass(t *testing.T, gwClassName string) {
	if err := lib.AKOControlConfig().AdvL4Clientset().NetworkingV1alpha1pre1().GatewayClasses().Delete(context.TODO(), gwClassName, metav1.DeleteOptions{}); err != nil {
		t.Fatalf("error in deleting GatewayClass: %v", err)
	}
}

func SetupAdvLBService(t *testing.T, svcname, namespace, gwname, gwnamespace string) {
	svc := integrationtest.FakeService{
		Name:      svcname,
		Namespace: namespace,
		Labels: map[string]string{
			lib.GatewayNameLabelKey:      gwname,
			lib.GatewayNamespaceLabelKey: gwnamespace,
			lib.GatewayTypeLabelKey:      "direct",
		},
		Type:         corev1.ServiceTypeLoadBalancer,
		ServicePorts: []integrationtest.Serviceport{{PortName: "foo1", Protocol: "TCP", PortNumber: 8081, TargetPort: intstr.FromInt(8081)}},
	}

	svcCreate := svc.Service()
	if _, err := KubeClient.CoreV1().Services(namespace).Create(context.TODO(), svcCreate, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding Service: %v", err)
	}
	integrationtest.CreateEP(t, namespace, svcname, false, true, "1.1.1")
}

func SetupAdvLBServiceWithLoadBalancerClass(t *testing.T, svcname, namespace, gwname, gwnamespace string, LBClass string) {
	svc := integrationtest.FakeService{
		Name:      svcname,
		Namespace: namespace,
		Labels: map[string]string{
			lib.GatewayNameLabelKey:      gwname,
			lib.GatewayNamespaceLabelKey: gwnamespace,
			lib.GatewayTypeLabelKey:      "direct",
		},
		Type:              corev1.ServiceTypeLoadBalancer,
		ServicePorts:      []integrationtest.Serviceport{{PortName: "foo1", Protocol: "TCP", PortNumber: 8081, TargetPort: intstr.FromInt(8081)}},
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

func TestAdvL4BestCase(t *testing.T) {
	// create gwclass, create gw, create 1svc
	// check graph VsNode vals, check IP status
	// remove gwclasss, IP removed
	g := gomega.NewGomegaWithT(t)

	gwClassName, gatewayName, ns := "avi-lb", "my-gateway", "default"
	modelName := "admin/abc--default-my-gateway"

	SetupGatewayClass(t, gwClassName, lib.AviGatewayController)
	SetupGateway(t, gatewayName, ns, gwClassName)

	SetupAdvLBService(t, "svc", ns, gatewayName, ns)

	g.Eventually(func() string {
		gw, _ := AdvL4Client.NetworkingV1alpha1pre1().Gateways(ns).Get(context.TODO(), gatewayName, metav1.GetOptions{})
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
	g.Expect(nodes[0].ServiceMetadata.NamespaceServiceName[0]).To(gomega.Equal("default/svc"))
	g.Expect(nodes[0].ServiceMetadata.Gateway).To(gomega.Equal("default/my-gateway"))
	g.Expect(nodes[0].PoolRefs[0].Servers).To(gomega.HaveLen(3))

	TeardownGatewayClass(t, gwClassName)
	g.Eventually(func() int {
		gw, _ := AdvL4Client.NetworkingV1alpha1pre1().Gateways(ns).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		return len(gw.Status.Addresses)
	}, 20*time.Second).Should(gomega.Equal(0))

	TeardownAdvLBService(t, "svc", ns)
	TeardownGateway(t, gatewayName, ns)
	VerifyGatewayVSNodeDeletion(g, modelName)
}

func TestAdvL4WithInvalidLoadBalancerClass(t *testing.T) {
	// create gwclass, create gw
	// create svc with invalid LBClass
	// Ako should skip LBClass validation and VS should come up

	g := gomega.NewGomegaWithT(t)
	gwClassName, gatewayName, ns := "avi-lb", "my-gateway", "default"
	modelName := "admin/abc--default-my-gateway"

	SetupGatewayClass(t, gwClassName, lib.AviGatewayController)
	SetupGateway(t, gatewayName, ns, gwClassName)
	SetupAdvLBServiceWithLoadBalancerClass(t, "svc", ns, gatewayName, ns, integrationtest.INVALID_LB_CLASS)

	g.Eventually(func() string {
		gw, _ := AdvL4Client.NetworkingV1alpha1pre1().Gateways(ns).Get(context.TODO(), gatewayName, metav1.GetOptions{})
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
	g.Expect(nodes[0].ServiceMetadata.NamespaceServiceName[0]).To(gomega.Equal("default/svc"))
	g.Expect(nodes[0].ServiceMetadata.Gateway).To(gomega.Equal("default/my-gateway"))
	g.Expect(nodes[0].PoolRefs[0].Servers).To(gomega.HaveLen(3))

	TeardownGatewayClass(t, gwClassName)
	g.Eventually(func() int {
		gw, _ := AdvL4Client.NetworkingV1alpha1pre1().Gateways(ns).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		return len(gw.Status.Addresses)
	}, 20*time.Second).Should(gomega.Equal(0))

	TeardownAdvLBService(t, "svc", ns)
	TeardownGateway(t, gatewayName, ns)
	VerifyGatewayVSNodeDeletion(g, modelName)
}
func TestAdvL4NamingConvention(t *testing.T) {
	// create gwclass, create gw, create 1svc
	// check naming conventions for vs, pool, l4policy
	// checks naming convention of all generated nodes
	g := gomega.NewGomegaWithT(t)

	gwClassName, gatewayName, ns := "avi-lb", "my-gateway", "default"
	modelName := "admin/abc--default-my-gateway"

	SetupGatewayClass(t, gwClassName, lib.AviGatewayController)
	SetupGateway(t, gatewayName, ns, gwClassName)

	SetupAdvLBService(t, "svc", ns, gatewayName, ns)

	g.Eventually(func() string {
		gw, _ := AdvL4Client.NetworkingV1alpha1pre1().Gateways(ns).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		if len(gw.Status.Addresses) > 0 {
			return gw.Status.Addresses[0].Value
		}
		return ""
	}, 40*time.Second).Should(gomega.Equal("10.250.250.1"))

	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes[0].Name).To(gomega.Equal("abc--default-my-gateway"))
	g.Expect(nodes[0].PoolRefs[0].Name).To(gomega.Equal("abc--default-svc-my-gateway-TCP--8081"))
	g.Expect(nodes[0].L4PolicyRefs[0].Name).To(gomega.Equal("abc--default-my-gateway"))

	TeardownGatewayClass(t, gwClassName)
	TeardownAdvLBService(t, "svc", ns)
	TeardownGateway(t, gatewayName, ns)
	VerifyGatewayVSNodeDeletion(g, modelName)
}

func TestAdvL4WithStaticIP(t *testing.T) {
	// create gwclass, create gw, create 1svc
	// check graph VsNode IPAddress val in vsvip ref
	g := gomega.NewGomegaWithT(t)

	gwClassName, gatewayName, ns := "avi-lb", "my-gateway", "default"
	modelName := "admin/abc--default-my-gateway"
	staticIP := "80.80.80.80"

	SetupGatewayClass(t, gwClassName, lib.AviGatewayController)
	gateway := FakeGateway{
		Name:      gatewayName,
		Namespace: ns,
		GWClass:   gwClassName,
		IPAddress: staticIP,
		Listeners: []FakeGWListener{{
			Port:     int32(8081),
			Protocol: "TCP",
			Labels: map[string]string{
				lib.GatewayNameLabelKey:      gatewayName,
				lib.GatewayNamespaceLabelKey: ns,
				lib.GatewayTypeLabelKey:      "direct",
			},
		}},
	}
	gwCreate := gateway.Gateway()
	if _, err := lib.AKOControlConfig().AdvL4Clientset().NetworkingV1alpha1pre1().Gateways(ns).Create(context.TODO(), gwCreate, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding Gateway: %v", err)
	}

	SetupAdvLBService(t, "svc", ns, gatewayName, ns)

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

func TestAdvL4WrongControllerGWClass(t *testing.T) {
	// create gateway, nothing happens
	// create gatewayclass, VS created
	// update to bad gatewayclass (wrong controller), VS deleted
	g := gomega.NewGomegaWithT(t)

	gwClassName, gatewayName, ns := "avi-lb", "my-gateway", "default"
	modelName := "admin/abc--default-my-gateway"

	SetupGateway(t, gatewayName, ns, gwClassName)
	SetupAdvLBService(t, "svc", ns, gatewayName, ns)

	SetupGatewayClass(t, gwClassName, lib.AviGatewayController)

	g.Eventually(func() string {
		gw, _ := AdvL4Client.NetworkingV1alpha1pre1().Gateways(ns).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		if len(gw.Status.Addresses) > 0 {
			return gw.Status.Addresses[0].Value
		}
		return ""
	}, 50*time.Second).Should(gomega.Equal("10.250.250.1"))

	gwclassUpdate := FakeGWClass{
		Name:       gwClassName,
		Controller: "xyz",
	}.GatewayClass()
	gwclassUpdate.ResourceVersion = "2"
	if _, err := lib.AKOControlConfig().AdvL4Clientset().NetworkingV1alpha1pre1().GatewayClasses().Update(context.TODO(), gwclassUpdate, metav1.UpdateOptions{}); err != nil {
		t.Fatalf("error in updating GatewayClass: %v", err)
	}

	g.Eventually(func() int {
		gw, _ := AdvL4Client.NetworkingV1alpha1pre1().Gateways(ns).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		return len(gw.Status.Addresses)
	}, 10*time.Second).Should(gomega.Equal(0))
	g.Eventually(func() int {
		svc, _ := KubeClient.CoreV1().Services(ns).Get(context.TODO(), "svc", metav1.GetOptions{})
		return len(svc.Status.LoadBalancer.Ingress)
	}, 10*time.Second).Should(gomega.Equal(0))

	TeardownAdvLBService(t, "svc", ns)
	TeardownGateway(t, gatewayName, ns)
	TeardownGatewayClass(t, gwClassName)
	VerifyGatewayVSNodeDeletion(g, modelName)
}

func TestAdvL4WrongClassMappingInGateway(t *testing.T) {
	// create gwclass, gw
	// update wrong mapping of class in gw, VS deleted
	// fix class in gw, VS created
	g := gomega.NewGomegaWithT(t)

	gwClassName, gatewayName, ns := "avi-lb", "my-gateway", "default"
	modelName := "admin/abc--default-my-gateway"

	SetupGateway(t, gatewayName, ns, gwClassName)
	SetupAdvLBService(t, "svc", ns, gatewayName, ns)

	SetupGatewayClass(t, gwClassName, lib.AviGatewayController)

	g.Eventually(func() string {
		gw, _ := AdvL4Client.NetworkingV1alpha1pre1().Gateways(ns).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		if len(gw.Status.Addresses) > 0 {
			return gw.Status.Addresses[0].Value
		}
		return ""
	}, 10*time.Second).Should(gomega.Equal("10.250.250.1"))

	gwUpdate := FakeGateway{
		Name: gatewayName, Namespace: ns, GWClass: gwClassName,
		Listeners: []FakeGWListener{{
			Port: int32(8081), Protocol: "TCP",
			Labels: map[string]string{
				lib.GatewayNameLabelKey:      "BADGATEWAY",
				lib.GatewayNamespaceLabelKey: ns,
				lib.GatewayTypeLabelKey:      "direct",
			},
		}},
	}.Gateway()
	gwUpdate.ResourceVersion = "2"
	if _, err := lib.AKOControlConfig().AdvL4Clientset().NetworkingV1alpha1pre1().Gateways(ns).Update(context.TODO(), gwUpdate, metav1.UpdateOptions{}); err != nil {
		t.Fatalf("error in updating Gateway: %v", err)
	}

	// vsNode must get deleted
	VerifyGatewayVSNodeDeletion(g, modelName)
	g.Eventually(func() int {
		gw, _ := AdvL4Client.NetworkingV1alpha1pre1().Gateways(ns).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		return len(gw.Status.Addresses)
	}, 10*time.Second).Should(gomega.Equal(0))

	gwUpdate = FakeGateway{
		Name: gatewayName, Namespace: ns, GWClass: gwClassName,
		Listeners: []FakeGWListener{{
			Port: int32(8081), Protocol: "TCP",
			Labels: map[string]string{
				lib.GatewayNameLabelKey:      gatewayName,
				lib.GatewayNamespaceLabelKey: ns,
				lib.GatewayTypeLabelKey:      "direct",
			},
		}},
	}.Gateway()
	gwUpdate.ResourceVersion = "3"
	if _, err := lib.AKOControlConfig().AdvL4Clientset().NetworkingV1alpha1pre1().Gateways(ns).Update(context.TODO(), gwUpdate, metav1.UpdateOptions{}); err != nil {
		t.Fatalf("error in updating Gateway: %v", err)
	}

	// vsNode must come back up
	g.Eventually(func() int {
		gw, _ := AdvL4Client.NetworkingV1alpha1pre1().Gateways(ns).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		return len(gw.Status.Addresses)
	}, 10*time.Second).Should(gomega.Equal(1))

	TeardownAdvLBService(t, "svc", ns)
	TeardownGateway(t, gatewayName, ns)
	TeardownGatewayClass(t, gwClassName)
	VerifyGatewayVSNodeDeletion(g, modelName)
}

func TestAdvL4ProtocolChangeInService(t *testing.T) {
	// gw/tcp/8081 svc/tcp/8081  -> svc/udp/8081
	// service protocol changes Pool deleted
	g := gomega.NewGomegaWithT(t)

	gwClassName, gatewayName, ns := "avi-lb", "my-gateway", "default"
	modelName := "admin/abc--default-my-gateway"

	SetupGatewayClass(t, gwClassName, lib.AviGatewayController)
	SetupGateway(t, gatewayName, ns, gwClassName)
	SetupAdvLBService(t, "svc", ns, gatewayName, ns)

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
			lib.GatewayNameLabelKey:      gatewayName,
			lib.GatewayNamespaceLabelKey: ns,
			lib.GatewayTypeLabelKey:      "direct",
		},
		Type:         corev1.ServiceTypeLoadBalancer,
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
	}, 20*time.Second).Should(gomega.Equal(true))

	TeardownAdvLBService(t, "svc", ns)
	TeardownGateway(t, gatewayName, ns)
	TeardownGatewayClass(t, gwClassName)
	VerifyGatewayVSNodeDeletion(g, modelName)
}

func TestAdvL4PortChangeInService(t *testing.T) {
	// gw/tcp/8081 svc/tcp/8081 -> svc/tcp/8080
	// service port changes Pools deleted
	g := gomega.NewGomegaWithT(t)

	gwClassName, gatewayName, ns := "avi-lb", "my-gateway", "default"
	modelName := "admin/abc--default-my-gateway"

	SetupGatewayClass(t, gwClassName, lib.AviGatewayController)
	SetupGateway(t, gatewayName, ns, gwClassName)
	SetupAdvLBService(t, "svc", ns, gatewayName, ns)

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
			lib.GatewayNameLabelKey:      gatewayName,
			lib.GatewayNamespaceLabelKey: ns,
			lib.GatewayTypeLabelKey:      "direct",
		},
		Type:         corev1.ServiceTypeLoadBalancer,
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
	}, 20*time.Second).Should(gomega.Equal(true))

	TeardownAdvLBService(t, "svc", ns)
	TeardownGateway(t, gatewayName, ns)
	TeardownGatewayClass(t, gwClassName)
	VerifyGatewayVSNodeDeletion(g, modelName)
}

func TestAdvL4LabelUpdatesInService(t *testing.T) {
	// correct labels, label mismatch, correct labels, delete labels
	g := gomega.NewGomegaWithT(t)

	gwClassName, gatewayName, ns := "avi-lb", "my-gateway", "default"
	modelName := "admin/abc--default-my-gateway"

	SetupGatewayClass(t, gwClassName, lib.AviGatewayController)
	SetupGateway(t, gatewayName, ns, gwClassName)
	SetupAdvLBService(t, "svc", ns, gatewayName, ns)

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
			lib.GatewayNameLabelKey:      "BADGATEWAY",
			lib.GatewayNamespaceLabelKey: ns,
			lib.GatewayTypeLabelKey:      "direct",
		},
		Type:         corev1.ServiceTypeLoadBalancer,
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
	}, 20*time.Second).Should(gomega.Equal(true))

	TeardownAdvLBService(t, "svc", ns)
	TeardownGateway(t, gatewayName, ns)
	TeardownGatewayClass(t, gwClassName)
	VerifyGatewayVSNodeDeletion(g, modelName)
}

func TestAdvL4LabelUpdatesInGateway(t *testing.T) {
	// correct labels, label mismatch, correct labels, delete labels
	g := gomega.NewGomegaWithT(t)

	gwClassName, gatewayName, ns := "avi-lb", "my-gateway", "default"
	modelName := "admin/abc--default-my-gateway"

	SetupGatewayClass(t, gwClassName, lib.AviGatewayController)
	SetupGateway(t, gatewayName, ns, gwClassName)
	SetupAdvLBService(t, "svc", ns, gatewayName, ns)

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
			Port:     int32(8081),
			Protocol: "TCP",
			Labels: map[string]string{
				lib.GatewayNameLabelKey:      "BADGATEWAY",
				lib.GatewayNamespaceLabelKey: ns,
				lib.GatewayTypeLabelKey:      "direct",
			},
		}},
	}.Gateway()
	gwUpdate.ResourceVersion = "2"
	if _, err := lib.AKOControlConfig().AdvL4Clientset().NetworkingV1alpha1pre1().Gateways(ns).Update(context.TODO(), gwUpdate, metav1.UpdateOptions{}); err != nil {
		t.Fatalf("error in updating Gateway: %v", err)
	}

	g.Eventually(func() bool {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if found && aviModel == nil {
			return false
		}
		return true
	}, 20*time.Second).Should(gomega.Equal(false))

	TeardownAdvLBService(t, "svc", ns)
	TeardownGateway(t, gatewayName, ns)
	TeardownGatewayClass(t, gwClassName)
	VerifyGatewayVSNodeDeletion(g, modelName)
}

func TestAdvL4GatewayListenerPortUpdate(t *testing.T) {
	// svc/tcp/8081
	// change gateway listener port to 8080, VS deletes
	// change svc port to 8080, VS creates, with 8080 exposed port
	g := gomega.NewGomegaWithT(t)

	gwClassName, gatewayName, ns := "avi-lb", "my-gateway", "default"
	modelName := "admin/abc--default-my-gateway"

	SetupGatewayClass(t, gwClassName, lib.AviGatewayController)
	SetupGateway(t, gatewayName, ns, gwClassName)
	SetupAdvLBService(t, "svc", ns, gatewayName, ns)

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
			Port:     int32(8080),
			Protocol: "TCP",
			Labels: map[string]string{
				lib.GatewayNameLabelKey:      gatewayName,
				lib.GatewayNamespaceLabelKey: ns,
				lib.GatewayTypeLabelKey:      "direct",
			},
		}},
	}.Gateway()
	gwUpdate.ResourceVersion = "2"
	if _, err := lib.AKOControlConfig().AdvL4Clientset().NetworkingV1alpha1pre1().Gateways(ns).Update(context.TODO(), gwUpdate, metav1.UpdateOptions{}); err != nil {
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
	}, 20*time.Second).Should(gomega.Equal(true))

	// match service to chaged gateway port: 8080
	svcUpdate := integrationtest.FakeService{
		Name:      "svc",
		Namespace: ns,
		Labels: map[string]string{
			lib.GatewayNameLabelKey:      gatewayName,
			lib.GatewayNamespaceLabelKey: ns,
			lib.GatewayTypeLabelKey:      "direct",
		},
		Type:         corev1.ServiceTypeLoadBalancer,
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

func TestAdvL4GatewayListenerProtocolUpdate(t *testing.T) {
	// svc/tcp/8080
	// change gateway listener protocol to UDP, VS deletes
	// change svc protocol to UDP, VS creates, with UDP protocol
	g := gomega.NewGomegaWithT(t)

	gwClassName, gatewayName, ns := "avi-lb", "my-gateway", "default"
	modelName := "admin/abc--default-my-gateway"

	SetupGatewayClass(t, gwClassName, lib.AviGatewayController)
	SetupGateway(t, gatewayName, ns, gwClassName)
	SetupAdvLBService(t, "svc", ns, gatewayName, ns)

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
			Port:     int32(8081),
			Protocol: "UDP",
			Labels: map[string]string{
				lib.GatewayNameLabelKey:      gatewayName,
				lib.GatewayNamespaceLabelKey: ns,
				lib.GatewayTypeLabelKey:      "direct",
			},
		}},
	}.Gateway()
	gwUpdate.ResourceVersion = "2"
	if _, err := lib.AKOControlConfig().AdvL4Clientset().NetworkingV1alpha1pre1().Gateways(ns).Update(context.TODO(), gwUpdate, metav1.UpdateOptions{}); err != nil {
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
	}, 20*time.Second).Should(gomega.Equal(true))

	// match service to chaged gateway protocol: UDP
	svcUpdate := integrationtest.FakeService{
		Name:      "svc",
		Namespace: ns,
		Labels: map[string]string{
			lib.GatewayNameLabelKey:      gatewayName,
			lib.GatewayNamespaceLabelKey: ns,
			lib.GatewayTypeLabelKey:      "direct",
		},
		Type:         corev1.ServiceTypeLoadBalancer,
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

func TestAdvL4MultiGatewayServiceUpdate(t *testing.T) {
	// svc/tcp/8081, gw1/tcp/8081, gw2/tcp/8081
	// change gateway from gw1 to gw2, gw1 VS deletes, gw2 VS is created
	g := gomega.NewGomegaWithT(t)

	gwClassName, gateway1Name, gateway2Name, ns := "avi-lb", "my-gateway1", "my-gateway2", "default"
	modelName1 := "admin/abc--default-my-gateway1"
	modelName2 := "admin/abc--default-my-gateway2"

	SetupGatewayClass(t, gwClassName, lib.AviGatewayController)
	SetupGateway(t, gateway1Name, ns, gwClassName)
	SetupGateway(t, gateway2Name, ns, gwClassName)
	SetupAdvLBService(t, "svc", ns, gateway1Name, ns)

	g.Eventually(func() bool {
		if found, aviModel := objects.SharedAviGraphLister().Get(modelName1); found && aviModel != nil {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
			if len(nodes) > 0 && nodes[0].Name == "abc--default-my-gateway1" {
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
			lib.GatewayNameLabelKey:      gateway2Name,
			lib.GatewayNamespaceLabelKey: ns,
			lib.GatewayTypeLabelKey:      "direct",
		},
		Type:         corev1.ServiceTypeLoadBalancer,
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
	}, 20*time.Second).Should(gomega.Equal(true))

	g.Eventually(func() bool {
		if found, aviModel := objects.SharedAviGraphLister().Get(modelName2); found && aviModel != nil {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
			if len(nodes) > 0 && nodes[0].Name == "abc--default-my-gateway2" {
				return true
			}
		}
		return false
	}, 20*time.Second).Should(gomega.Equal(true))

	TeardownAdvLBService(t, "svc", ns)
	TeardownGateway(t, gateway1Name, ns)
	TeardownGateway(t, gateway2Name, ns)
	TeardownGatewayClass(t, gwClassName)
	VerifyGatewayVSNodeDeletion(g, modelName2)
}

func TestAdvL4EndpointDeleteCreate(t *testing.T) {
	// svc/tcp/8081, gw1/tcp/8081
	// scale deployment to
	g := gomega.NewGomegaWithT(t)

	gwClassName, gatewayName, ns := "avi-lb", "my-gateway", "default"
	modelName := "admin/abc--default-my-gateway"

	SetupGatewayClass(t, gwClassName, lib.AviGatewayController)
	SetupGateway(t, gatewayName, ns, gwClassName)
	SetupAdvLBService(t, "svc", ns, gatewayName, ns)

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

func TestAdvL4MultiTenancyWithInfraSettting(t *testing.T) {
	// create a gw object, infrasetting and annotate a namespace with infrasetting
	// graph layer objects should come up with correct tenant
	// delete the gw object, graph layer object deletion
	g := gomega.NewGomegaWithT(t)

	infraSettingName := "my-infrasetting"
	gwClassName, gatewayName, ns := "avi-lb", "my-gateway", "default"
	modelName := "nonadmin/abc--default-my-gateway"

	integrationtest.SetupAviInfraSetting(t, infraSettingName, "DEDICATED")
	integrationtest.AnnotateAKONamespaceWithInfraSetting(t, ns, infraSettingName)
	integrationtest.AnnotateNamespaceWithTenant(t, ns, "nonadmin")

	SetupGatewayClass(t, gwClassName, lib.AviGatewayController)
	SetupGateway(t, gatewayName, ns, gwClassName)

	SetupAdvLBService(t, "svc", ns, gatewayName, ns)

	g.Eventually(func() string {
		gw, _ := AdvL4Client.NetworkingV1alpha1pre1().Gateways(ns).Get(context.TODO(), gatewayName, metav1.GetOptions{})
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
	g.Expect(nodes[0].Name).To(gomega.Equal("abc--default-my-gateway"))
	g.Expect(nodes[0].Tenant).To(gomega.Equal("nonadmin"))
	g.Expect(nodes[0].PortProto[0].Port).To(gomega.Equal(int32(8081)))
	g.Expect(nodes[0].HttpPolicySetRefs).To(gomega.HaveLen(0))
	g.Expect(nodes[0].L4PolicyRefs).To(gomega.HaveLen(1))
	g.Expect(nodes[0].L4PolicyRefs[0].PortPool[0].Port).To(gomega.Equal(uint32(8081)))
	g.Expect(nodes[0].L4PolicyRefs[0].PortPool[0].Protocol).To(gomega.Equal("TCP"))
	g.Expect(nodes[0].ServiceMetadata.NamespaceServiceName[0]).To(gomega.Equal("default/svc"))
	g.Expect(nodes[0].ServiceMetadata.Gateway).To(gomega.Equal("default/my-gateway"))
	g.Expect(nodes[0].PoolRefs[0].Servers).To(gomega.HaveLen(3))

	TeardownGatewayClass(t, gwClassName)
	g.Eventually(func() int {
		gw, _ := AdvL4Client.NetworkingV1alpha1pre1().Gateways(ns).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		return len(gw.Status.Addresses)
	}, 40*time.Second).Should(gomega.Equal(0))

	TeardownAdvLBService(t, "svc", ns)
	TeardownGateway(t, gatewayName, ns)
	VerifyGatewayVSNodeDeletion(g, modelName)
	integrationtest.RemoveAnnotateAKONamespaceWithInfraSetting(t, ns)
	integrationtest.TeardownAviInfraSetting(t, infraSettingName)
}

func TestAdvL4MultiTenancyWithTenantAddition(t *testing.T) {
	// create a gw object, infrasetting and annotate a namespace with infrasetting
	// graph layer objects should come up with correct tenant
	// delete the gw object, graph layer object deletion
	g := gomega.NewGomegaWithT(t)

	gwClassName, gatewayName, ns := "avi-lb", "my-gateway", "default"
	modelName := "admin/abc--default-my-gateway"

	SetupGatewayClass(t, gwClassName, lib.AviGatewayController)
	SetupGateway(t, gatewayName, ns, gwClassName)

	SetupAdvLBService(t, "svc", ns, gatewayName, ns)

	g.Eventually(func() string {
		gw, _ := AdvL4Client.NetworkingV1alpha1pre1().Gateways(ns).Get(context.TODO(), gatewayName, metav1.GetOptions{})
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
	g.Expect(nodes[0].Name).To(gomega.Equal("abc--default-my-gateway"))
	g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))
	g.Expect(nodes[0].PortProto[0].Port).To(gomega.Equal(int32(8081)))
	g.Expect(nodes[0].HttpPolicySetRefs).To(gomega.HaveLen(0))
	g.Expect(nodes[0].L4PolicyRefs).To(gomega.HaveLen(1))
	g.Expect(nodes[0].L4PolicyRefs[0].PortPool[0].Port).To(gomega.Equal(uint32(8081)))
	g.Expect(nodes[0].L4PolicyRefs[0].PortPool[0].Protocol).To(gomega.Equal("TCP"))
	g.Expect(nodes[0].ServiceMetadata.NamespaceServiceName[0]).To(gomega.Equal("default/svc"))
	g.Expect(nodes[0].ServiceMetadata.Gateway).To(gomega.Equal("default/my-gateway"))
	g.Expect(nodes[0].PoolRefs[0].Servers).To(gomega.HaveLen(3))

	integrationtest.AnnotateNamespaceWithTenant(t, ns, "nonadmin")
	g.Eventually(func() bool {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		return aviModel == nil
	}, 60*time.Second).Should(gomega.Equal(true))

	modelName = "nonadmin/abc--default-my-gateway"

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 30*time.Second).Should(gomega.Equal(true))

	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes[0].Name).To(gomega.Equal("abc--default-my-gateway"))
	g.Expect(nodes[0].Tenant).To(gomega.Equal("nonadmin"))
	g.Expect(nodes[0].PortProto[0].Port).To(gomega.Equal(int32(8081)))
	g.Expect(nodes[0].HttpPolicySetRefs).To(gomega.HaveLen(0))
	g.Expect(nodes[0].L4PolicyRefs).To(gomega.HaveLen(1))
	g.Expect(nodes[0].L4PolicyRefs[0].PortPool[0].Port).To(gomega.Equal(uint32(8081)))
	g.Expect(nodes[0].L4PolicyRefs[0].PortPool[0].Protocol).To(gomega.Equal("TCP"))
	g.Expect(nodes[0].ServiceMetadata.NamespaceServiceName[0]).To(gomega.Equal("default/svc"))
	g.Expect(nodes[0].ServiceMetadata.Gateway).To(gomega.Equal("default/my-gateway"))
	g.Expect(nodes[0].PoolRefs[0].Servers).To(gomega.HaveLen(3))

	TeardownGatewayClass(t, gwClassName)
	g.Eventually(func() int {
		gw, _ := AdvL4Client.NetworkingV1alpha1pre1().Gateways(ns).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		return len(gw.Status.Addresses)
	}, 40*time.Second).Should(gomega.Equal(0))

	TeardownAdvLBService(t, "svc", ns)
	TeardownGateway(t, gatewayName, ns)
	VerifyGatewayVSNodeDeletion(g, modelName)
}

func TestAdvL4MultiTenancyWithTenantDeannotationInNS(t *testing.T) {
	// create a gw object, infrasetting and annotate a namespace with infrasetting
	// graph layer objects should come up with correct tenant
	// delete the Infrasetting annotation from the namespace, old model should be deleted
	// new model in default tenant should get created
	g := gomega.NewGomegaWithT(t)

	gwClassName, gatewayName, ns := "avi-lb", "my-gateway", "default"
	modelName := "nonadmin/abc--default-my-gateway"

	integrationtest.AnnotateNamespaceWithTenant(t, ns, "nonadmin")

	SetupGatewayClass(t, gwClassName, lib.AviGatewayController)
	SetupGateway(t, gatewayName, ns, gwClassName)

	SetupAdvLBService(t, "svc", ns, gatewayName, ns)

	g.Eventually(func() string {
		gw, _ := AdvL4Client.NetworkingV1alpha1pre1().Gateways(ns).Get(context.TODO(), gatewayName, metav1.GetOptions{})
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
	g.Expect(nodes[0].ServiceMetadata.NamespaceServiceName[0]).To(gomega.Equal("default/svc"))
	g.Expect(nodes[0].ServiceMetadata.Gateway).To(gomega.Equal("default/my-gateway"))
	g.Expect(nodes[0].PoolRefs[0].Servers).To(gomega.HaveLen(3))

	integrationtest.RemoveAnnotateAKONamespaceWithInfraSetting(t, ns)

	g.Eventually(func() bool {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		return aviModel == nil
	}, 60*time.Second).Should(gomega.Equal(true))

	newModelName := "admin/abc--default-my-gateway"
	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(newModelName)
		return found
	}, 30*time.Second).Should(gomega.Equal(true))

	TeardownGatewayClass(t, gwClassName)

	g.Eventually(func() int {
		gw, _ := AdvL4Client.NetworkingV1alpha1pre1().Gateways(ns).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		return len(gw.Status.Addresses)
	}, 20*time.Second).Should(gomega.Equal(0))

	TeardownAdvLBService(t, "svc", ns)
	TeardownGateway(t, gatewayName, ns)
	VerifyGatewayVSNodeDeletion(g, newModelName)
}

// @AI-Generated
// [Generated by Google Gemini Code Assist]
// Description: UT to test Infrasetting transition from Invalid to Valid state
func TestAdvL4AviInfraSettingTransitionFromInvalidToValid(t *testing.T) {
	// This test verifies that when an AviInfraSetting transitions from invalid to valid,
	// the corresponding VirtualService is correctly configured.
	// 1. Initially, an invalid AviInfraSetting is applied. AKO should use default settings for the VS.
	// 2. The AviInfraSetting is then updated to be valid. AKO should update the VS with the new settings.
	g := gomega.NewGomegaWithT(t)

	// Setup unique names for test resources to avoid conflicts.
	infraSettingName := "my-infrasetting-transition"
	gwClassName := "avi-lb-transition"
	gatewayName := "my-gateway-transition"
	ns := "default"
	svcName := "svc-transition"
	tenant := "admin"
	modelName := tenant + "/abc--default-" + gatewayName

	// Defer cleanup of annotations and infrasetting to ensure test isolation.
	defer integrationtest.TeardownAviInfraSetting(t, infraSettingName)
	defer integrationtest.RemoveAnnotateAKONamespaceWithInfraSetting(t, ns)

	// Create an initially invalid AviInfraSetting (invalid network name).
	invalidSettings := integrationtest.FakeAviInfraSetting{
		Name:        infraSettingName,
		SeGroupName: "thisisaviref-" + infraSettingName + "-seGroup",
		Networks:    []string{"thisisBADaviref-" + infraSettingName + "-networkName"},
		ShardSize:   "DEDICATED",
		EnableRhi:   true,
	}
	settingCreate := invalidSettings.AviInfraSetting()
	if _, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().AviInfraSettings().Create(context.TODO(), settingCreate, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding AviInfraSetting: %v", err)
	}

	// Annotate namespace to use the AviInfraSetting.
	integrationtest.AnnotateAKONamespaceWithInfraSetting(t, ns, infraSettingName)

	// Setup Gateway and Service.
	SetupGatewayClass(t, gwClassName, lib.AviGatewayController)
	SetupGateway(t, gatewayName, ns, gwClassName)
	SetupAdvLBService(t, svcName, ns, gatewayName, ns)

	// STEP 1: Verify VS is created with default settings due to invalid AviInfraSetting.
	t.Log("Verifying VS is created with default settings for invalid AviInfraSetting")
	g.Eventually(func() {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		g.Expect(found).To(gomega.BeTrue(), "model should be found")
		g.Expect(aviModel).NotTo(gomega.BeNil(), "model should not be nil")

		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(nodes).To(gomega.HaveLen(1), "should have one VS node")

		vsNode := nodes[0]
		g.Expect(vsNode.Name).To(gomega.Equal("abc--default-" + gatewayName))
		g.Expect(vsNode.Tenant).To(gomega.Equal(tenant))
		g.Expect(vsNode.ServiceEngineGroup).To(gomega.Equal(lib.GetSEGName()), "should use default SE group")
		g.Expect(vsNode.VSVIPRefs[0].T1Lr).To(gomega.Equal(lib.GetT1LRPath()), "should use default T1 LR")
	}, 25*time.Second, 1*time.Second).Should(gomega.Succeed())

	// STEP 2: Update AviInfraSetting to be valid.
	t.Log("Updating AviInfraSetting to be valid")
	validSettings := integrationtest.FakeAviInfraSetting{
		Name:          infraSettingName,
		SeGroupName:   "thisisaviref-" + infraSettingName + "-seGroup",
		Networks:      []string{"thisisaviref-" + infraSettingName + "-networkName"},
		EnableRhi:     true,
		BGPPeerLabels: []string{"peer1", "peer2"},
		ShardSize:     "DEDICATED",
		T1LR:          "avi-domain-c9:1234",
	}.AviInfraSetting()
	validSettings.ResourceVersion = "2" // Must update the resource version for update operation.
	if _, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().AviInfraSettings().Update(context.TODO(), validSettings, metav1.UpdateOptions{}); err != nil {
		t.Fatalf("error in updating AviInfraSetting: %v", err)
	}

	// Verify AviInfraSetting status becomes "Accepted".
	g.Eventually(func() {
		setting, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().AviInfraSettings().Get(context.TODO(), infraSettingName, metav1.GetOptions{})
		g.Expect(err).ToNot(gomega.HaveOccurred())
		g.Expect(setting.Status.Status).To(gomega.Equal("Accepted"))
	}, 40*time.Second, 2*time.Second).Should(gomega.Succeed())

	// STEP 3: Verify VS is updated with values from the valid AviInfraSetting.
	t.Log("Verifying VS is updated with new settings")
	g.Eventually(func() {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		g.Expect(found).To(gomega.BeTrue())
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		g.Expect(nodes).To(gomega.HaveLen(1))

		vsNode := nodes[0]
		g.Expect(vsNode.ServiceEngineGroup).To(gomega.Equal("thisisaviref-" + infraSettingName + "-seGroup"))
		g.Expect(vsNode.VSVIPRefs[0].T1Lr).To(gomega.Equal("avi-domain-c9:1234"))
		g.Expect(vsNode.VSVIPRefs[0].VipNetworks).To(gomega.HaveLen(1))
		g.Expect(vsNode.VSVIPRefs[0].VipNetworks[0].NetworkName).To(gomega.Equal("thisisaviref-" + infraSettingName + "-networkName"))
	}, 25*time.Second, 1*time.Second).Should(gomega.Succeed())

	// STEP 4: Clean up K8s resources and verify VS deletion.
	TeardownGatewayClass(t, gwClassName)
	g.Eventually(func() int {
		gw, _ := AdvL4Client.NetworkingV1alpha1pre1().Gateways(ns).Get(context.TODO(), gatewayName, metav1.GetOptions{})
		return len(gw.Status.Addresses)
	}, 40*time.Second).Should(gomega.Equal(0))

	TeardownAdvLBService(t, svcName, ns)
	TeardownGateway(t, gatewayName, ns)
	VerifyGatewayVSNodeDeletion(g, modelName)
}
