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

package oshiftroutetests

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

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/integrationtest"

	utils "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	"github.com/onsi/gomega"
	oshiftfake "github.com/openshift/client-go/route/clientset/versioned/fake"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

var KubeClient *k8sfake.Clientset
var CRDClient *crdfake.Clientset
var V1beta1CRDClient *v1beta1crdfake.Clientset
var ctrl *k8s.AviController

func TestMain(m *testing.M) {
	os.Setenv("INGRESS_API", "extensionv1")
	os.Setenv("VIP_NETWORK_LIST", `[{"networkName":"net123"}]`)
	os.Setenv("CLUSTER_NAME", "cluster")
	os.Setenv("CLOUD_NAME", "CLOUD_VCENTER")
	os.Setenv("SEG_NAME", "Default-Group")
	os.Setenv("NODE_NETWORK_LIST", `[{"networkName":"net123","cidrs":["10.79.168.0/22"]}]`)
	os.Setenv("POD_NAMESPACE", utils.AKO_DEFAULT_NS)
	os.Setenv("SHARD_VS_SIZE", "LARGE")
	os.Setenv("POD_NAME", "ako-0")
	os.Setenv("DEFAULT_DOMAIN", "com")
	os.Setenv("AUTO_L4_FQDN", "default")

	akoControlConfig := lib.AKOControlConfig()
	KubeClient = k8sfake.NewSimpleClientset()
	CRDClient = crdfake.NewSimpleClientset()
	V1beta1CRDClient = v1beta1crdfake.NewSimpleClientset()
	akoControlConfig.SetCRDClientset(CRDClient)
	akoControlConfig.Setv1beta1CRDClientset(V1beta1CRDClient)
	akoControlConfig.SetEventRecorder(lib.AKOEventComponent, KubeClient, true)
	akoControlConfig.SetAKOInstanceFlag(true)
	data := map[string][]byte{
		"username": []byte("admin"),
		"password": []byte("admin"),
	}
	object := metav1.ObjectMeta{Name: "avi-secret", Namespace: utils.GetAKONamespace()}
	secret := &corev1.Secret{Data: data, ObjectMeta: object}
	KubeClient.CoreV1().Secrets(utils.GetAKONamespace()).Create(context.TODO(), secret, metav1.CreateOptions{})
	OshiftClient = oshiftfake.NewSimpleClientset()
	informersArg := make(map[string]interface{})
	informersArg[utils.INFORMERS_OPENSHIFT_CLIENT] = OshiftClient
	registeredInformers := []string{
		utils.ServiceInformer,
		utils.RouteInformer,
		utils.SecretInformer,
		utils.NSInformer,
		utils.NodeInformer,
		utils.ConfigMapInformer,
	}

	registeredInformers = append(registeredInformers, utils.EndpointSlicesInformer)
	utils.NewInformers(utils.KubeClientIntf{ClientSet: KubeClient}, registeredInformers, informersArg)
	informers := k8s.K8sinformers{Cs: KubeClient}
	k8s.NewCRDInformers()

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
	integrationtest.PollForSyncStart(ctrl, 10)

	ctrl.HandleConfigMap(informers, ctrlCh, stopCh, quickSyncCh)
	integrationtest.KubeClient = KubeClient
	integrationtest.AddDefaultIngressClass()
	ctrl.SetSEGroupCloudNameFromNSAnnotations()

	SetupRouteNamespaceSync(defaultKey, defaultValue)
	integrationtest.AddDefaultNamespace()
	integrationtest.AddDefaultNamespace("red")

	go ctrl.InitController(informers, registeredInformers, ctrlCh, stopCh, quickSyncCh, waitGroupMap)

	os.Exit(m.Run())
}

func TestRouteNoPath(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	defaultNamespace = "default"
	SetUpTestForRoute(t, defaultModelName)

	routeExample := FakeRoute{}.Route()
	_, err := OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExample, metav1.CreateOptions{})

	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	aviModel := ValidateModelCommon(t, g)
	pool := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].PoolRefs[0]

	g.Eventually(func() int {
		pool = aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].PoolRefs[0]
		return len(pool.Servers)
	}, 10*time.Second).Should(gomega.Equal(1))

	g.Expect(pool.Name).To(gomega.Equal("cluster--foo.com-default-foo-avisvc"))
	g.Expect(pool.PriorityLabel).To(gomega.Equal("foo.com"))

	poolgroups := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].PoolGroupRefs
	pgmember := poolgroups[0].Members[0]
	g.Expect(*pgmember.PoolRef).To(gomega.Equal("/api/pool?name=cluster--foo.com-default-foo-avisvc"))
	g.Expect(*pgmember.PriorityLabel).To(gomega.Equal("foo.com"))

	VerifyRouteDeletion(t, g, aviModel, 0)
	TearDownTestForRoute(t, defaultModelName)
}

func TestOshiftNamingConvention(t *testing.T) {
	// checks naming convention of all generated nodes
	g := gomega.NewGomegaWithT(t)
	SetUpTestForRoute(t, defaultModelName)
	routeExample := FakeRoute{Path: "/foo/bar"}.Route()
	_, err := OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	aviModel := ValidateModelCommon(t, g)
	vsNode := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0]

	g.Expect(vsNode.Name).To(gomega.Equal("cluster--Shared-L7-0"))
	g.Expect(vsNode.PoolGroupRefs[0].Name).To(gomega.Equal("cluster--Shared-L7-0"))
	g.Expect(vsNode.PoolRefs[0].Name).To(gomega.Equal("cluster--foo.com_foo_bar-default-foo-avisvc"))
	g.Expect(vsNode.HTTPDSrefs[0].Name).To(gomega.Equal("cluster--Shared-L7-0"))
	g.Expect(vsNode.VSVIPRefs[0].Name).To(gomega.Equal("cluster--Shared-L7-0"))

	VerifyRouteDeletion(t, g, aviModel, 0)
	TearDownTestForRoute(t, defaultModelName)
}

func TestRouteDefaultPath(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetUpTestForRoute(t, defaultModelName)
	routeExample := FakeRoute{Path: "/foo"}.Route()
	_, err := OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	aviModel := ValidateModelCommon(t, g)
	pool := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].PoolRefs[0]

	g.Eventually(func() int {
		pool = aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].PoolRefs[0]
		return len(pool.Servers)
	}, 10*time.Second).Should(gomega.Equal(1))

	g.Expect(pool.Name).To(gomega.Equal("cluster--foo.com_foo-default-foo-avisvc"))
	g.Expect(pool.PriorityLabel).To(gomega.Equal("foo.com/foo"))

	poolgroups := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].PoolGroupRefs
	pgmember := poolgroups[0].Members[0]
	g.Expect(*pgmember.PoolRef).To(gomega.Equal("/api/pool?name=cluster--foo.com_foo-default-foo-avisvc"))
	g.Expect(*pgmember.PriorityLabel).To(gomega.Equal("foo.com/foo"))

	VerifyRouteDeletion(t, g, aviModel, 0)
	TearDownTestForRoute(t, defaultModelName)
}

func TestRouteServiceDel(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetUpTestForRoute(t, defaultModelName)
	integrationtest.CreateSVC(t, "default", "newsvc", corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPS(t, "default", "newsvc", false, false, "3.3.3")
	routeExample := FakeRoute{Path: "/foo", ServiceName: "newsvc"}.Route()
	_, err := OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	aviModel := ValidateModelCommon(t, g)

	integrationtest.DelSVC(t, "default", "newsvc")
	integrationtest.DelEPS(t, "default", "newsvc")

	g.Eventually(func() int {
		_, aviModel = objects.SharedAviGraphLister().Get(defaultModelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		if len(nodes) > 0 {
			return len(nodes[0].PoolRefs[0].Servers)
		}
		return -1
	}, 10*time.Second).Should(gomega.Equal(0))

	VerifyRouteDeletion(t, g, aviModel, 0)
	TearDownTestForRoute(t, defaultModelName)
}

func TestRouteBadService(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetUpTestForRoute(t, defaultModelName)
	routeExample := FakeRoute{Path: "/foo", ServiceName: "badsvc"}.Route()
	_, err := OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	aviModel := ValidateModelCommon(t, g)
	pool := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].PoolRefs[0]
	g.Expect(pool.Name).To(gomega.Equal("cluster--foo.com_foo-default-foo-badsvc"))
	g.Expect(pool.PriorityLabel).To(gomega.Equal("foo.com/foo"))
	g.Expect(len(pool.Servers)).To(gomega.Equal(0))

	VerifyRouteDeletion(t, g, aviModel, 0)
	TearDownTestForRoute(t, defaultModelName)
}

func TestRouteServiceAdd(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetUpTestForRoute(t, defaultModelName)
	routeExample := FakeRoute{Path: "/foo", ServiceName: "newsvc"}.Route()
	_, err := OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	integrationtest.CreateSVC(t, "default", "newsvc", corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPS(t, "default", "newsvc", false, false, "3.3.3")

	aviModel := ValidateModelCommon(t, g)
	g.Eventually(func() int {
		vslist := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		if len(vslist) == 0 {
			return 0
		}
		pools := vslist[0].PoolRefs
		if len(pools) == 0 {
			return 0
		}
		return len(pools[0].Servers)
	}, 60*time.Second).Should(gomega.Equal(1))

	poolgroups := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].PoolGroupRefs
	pgmember := poolgroups[0].Members[0]
	g.Expect(*pgmember.PoolRef).To(gomega.Equal("/api/pool?name=cluster--foo.com_foo-default-foo-newsvc"))
	g.Expect(*pgmember.PriorityLabel).To(gomega.Equal("foo.com/foo"))

	VerifyRouteDeletion(t, g, aviModel, 0)
	TearDownTestForRoute(t, defaultModelName)

	integrationtest.DelSVC(t, "default", "newsvc")
	integrationtest.DelEPS(t, "default", "newsvc")
}

func TestRouteScaleEndpoint(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetUpTestForRoute(t, defaultModelName)
	routeExample := FakeRoute{Path: "/foo"}.Route()
	_, err := OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	aviModel := ValidateModelCommon(t, g)
	pool := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].PoolRefs[0]

	integrationtest.ScaleCreateEPS(t, "default", "avisvc")
	g.Eventually(func() int {
		vslist := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		if len(vslist) == 0 {
			return 0
		}
		pools := vslist[0].PoolRefs
		if len(pools) == 0 {
			return 0
		}
		return len(pools[0].Servers)
	}, 60*time.Second).Should(gomega.Equal(2))

	g.Expect(pool.Name).To(gomega.Equal("cluster--foo.com_foo-default-foo-avisvc"))
	g.Expect(pool.PriorityLabel).To(gomega.Equal("foo.com/foo"))

	poolgroups := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].PoolGroupRefs
	pgmember := poolgroups[0].Members[0]
	g.Expect(*pgmember.PoolRef).To(gomega.Equal("/api/pool?name=cluster--foo.com_foo-default-foo-avisvc"))
	g.Expect(*pgmember.PriorityLabel).To(gomega.Equal("foo.com/foo"))

	VerifyRouteDeletion(t, g, aviModel, 0)
	TearDownTestForRoute(t, defaultModelName)
}

func TestMultiRouteSameHost(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetUpTestForRoute(t, defaultModelName)
	routeExample1 := FakeRoute{Path: "/foo", ServiceName: "avisvc"}.Route()
	_, err := OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExample1, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	routeExample2 := FakeRoute{Name: "bar", Path: "/bar", ServiceName: "avisvc"}.Route()
	_, err = OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExample2, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	aviModel := ValidateModelCommon(t, g)
	g.Eventually(func() int {
		_, aviModel := objects.SharedAviGraphLister().Get(defaultModelName)
		pools := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].PoolRefs
		return len(pools)
	}, 15*time.Second).Should(gomega.Equal(2))

	pools := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].PoolRefs
	for _, pool := range pools {
		if pool.Name == "cluster--foo.com_foo-default-foo-avisvc" {
			g.Expect(pool.PriorityLabel).To(gomega.Equal("foo.com/foo"))
			g.Expect(len(pool.Servers)).To(gomega.Equal(1))
		} else if pool.Name == "cluster--foo.com_bar-default-bar-avisvc" {
			g.Expect(pool.PriorityLabel).To(gomega.Equal("foo.com/bar"))
			g.Expect(len(pool.Servers)).To(gomega.Equal(1))
		} else {
			t.Fatalf("unexpected pool: %s", pool.Name)
		}
	}

	poolgroups := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].PoolGroupRefs
	for _, pgmember := range poolgroups[0].Members {
		if *pgmember.PoolRef == "/api/pool?name=cluster--foo.com_foo-default-foo-avisvc" {
			g.Expect(*pgmember.PriorityLabel).To(gomega.Equal("foo.com/foo"))
		} else if *pgmember.PoolRef == "/api/pool?name=cluster--foo.com_bar-default-bar-avisvc" {
			g.Expect(*pgmember.PriorityLabel).To(gomega.Equal("foo.com/bar"))
		} else {
			t.Fatalf("unexpected pgmember: %s", *pgmember.PoolRef)
		}
	}

	err = OshiftClient.RouteV1().Routes(defaultNamespace).Delete(context.TODO(), "bar", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the route %v", err)
	}

	VerifyRouteDeletion(t, g, aviModel, 0)
	TearDownTestForRoute(t, defaultModelName)
}

func TestRouteUpdatePath(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetUpTestForRoute(t, defaultModelName)
	routeExample := FakeRoute{Path: "/foo"}.Route()
	_, err := OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	routeExample = FakeRoute{Path: "/bar"}.Route()
	_, err = OshiftClient.RouteV1().Routes(defaultNamespace).Update(context.TODO(), routeExample, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	aviModel := ValidateModelCommon(t, g)
	pool := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].PoolRefs[0]

	g.Eventually(func() int {
		vslist := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		if len(vslist) == 0 {
			return 0
		}
		pools := vslist[0].PoolRefs
		if len(pools) == 0 {
			return 0
		}
		return len(pools[0].Servers)
	}, 60*time.Second).Should(gomega.Equal(1))

	g.Expect(pool.Name).To(gomega.Equal("cluster--foo.com_bar-default-foo-avisvc"))

	poolgroups := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].PoolGroupRefs
	for _, pgmember := range poolgroups[0].Members {
		if *pgmember.PoolRef == "/api/pool?name=cluster--foo.com_bar-default-foo-avisvc" {
			g.Expect(*pgmember.PriorityLabel).To(gomega.Equal("foo.com/bar"))
		} else {
			t.Fatalf("unexpected pgmember: %s", *pgmember.PoolRef)
		}
	}

	VerifyRouteDeletion(t, g, aviModel, 0)
	TearDownTestForRoute(t, defaultModelName)
}

func TestAlternateBackendNoPath(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetUpTestForRoute(t, defaultModelName)
	integrationtest.CreateSVC(t, "default", "absvc2", corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPS(t, "default", "absvc2", false, false, "3.3.3")
	time.Sleep(2 * time.Second)
	routeExample := FakeRoute{}.ABRoute()
	_, err := OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	aviModel := ValidateModelCommon(t, g)

	pools := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].PoolRefs
	g.Expect(pools).To(gomega.HaveLen(2))
	for _, pool := range pools {
		if pool.Name == "cluster--foo.com-default-foo-avisvc" || pool.Name == "cluster--foo.com-default-foo-absvc2" {
			g.Expect(len(pool.Servers)).To(gomega.Equal(1))
		} else {
			t.Fatalf("unexpected pool: %s", pool.Name)
		}
	}

	poolgroups := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].PoolGroupRefs
	for _, pgmember := range poolgroups[0].Members {
		if *pgmember.PoolRef == "/api/pool?name=cluster--foo.com-default-foo-avisvc" {
			g.Expect(*pgmember.PriorityLabel).To(gomega.Equal("foo.com"))
			g.Expect(*pgmember.Ratio).To(gomega.Equal(uint32(100)))
		} else if *pgmember.PoolRef == "/api/pool?name=cluster--foo.com-default-foo-absvc2" {
			g.Expect(*pgmember.PriorityLabel).To(gomega.Equal("foo.com"))
			g.Expect(*pgmember.Ratio).To(gomega.Equal(uint32(200)))
		} else {
			t.Fatalf("unexpected pgmember: %s", *pgmember.PoolRef)
		}
	}

	VerifyRouteDeletion(t, g, aviModel, 0)
	TearDownTestForRoute(t, defaultModelName)

	integrationtest.DelSVC(t, "default", "absvc2")
	integrationtest.DelEPS(t, "default", "absvc2")
}

func TestAlternateBackendDefaultPath(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetUpTestForRoute(t, defaultModelName)
	integrationtest.CreateSVC(t, "default", "absvc2", corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPS(t, "default", "absvc2", false, false, "3.3.3")
	routeExample := FakeRoute{Path: "/foo"}.ABRoute()
	_, err := OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	aviModel := ValidateModelCommon(t, g)

	pools := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].PoolRefs
	g.Expect(pools).To(gomega.HaveLen(2))
	for _, pool := range pools {
		if pool.Name == "cluster--foo.com_foo-default-foo-avisvc" || pool.Name == "cluster--foo.com_foo-default-foo-absvc2" {
			g.Expect(len(pool.Servers)).To(gomega.Equal(1))
		} else {
			t.Fatalf("unexpected pool: %s", pool.Name)
		}
	}

	poolgroups := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].PoolGroupRefs
	for _, pgmember := range poolgroups[0].Members {
		if *pgmember.PoolRef == "/api/pool?name=cluster--foo.com_foo-default-foo-avisvc" {
			g.Expect(*pgmember.PriorityLabel).To(gomega.Equal("foo.com/foo"))
			g.Expect(*pgmember.Ratio).To(gomega.Equal(uint32(100)))
		} else if *pgmember.PoolRef == "/api/pool?name=cluster--foo.com_foo-default-foo-absvc2" {
			g.Expect(*pgmember.PriorityLabel).To(gomega.Equal("foo.com/foo"))
			g.Expect(*pgmember.Ratio).To(gomega.Equal(uint32(200)))
		} else {
			t.Fatalf("unexpected pgmember: %s", *pgmember.PoolRef)
		}
	}

	VerifyRouteDeletion(t, g, aviModel, 0)
	TearDownTestForRoute(t, defaultModelName)

	integrationtest.DelSVC(t, "default", "absvc2")
	integrationtest.DelEPS(t, "default", "absvc2")
}

func TestRemoveAlternateBackend(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetUpTestForRoute(t, defaultModelName)
	integrationtest.CreateSVC(t, "default", "absvc2", corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPS(t, "default", "absvc2", false, false, "3.3.3")
	routeExample := FakeRoute{Path: "/foo"}.ABRoute()
	_, err := OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	routeExample = FakeRoute{Path: "/foo"}.Route()
	_, err = OshiftClient.RouteV1().Routes(defaultNamespace).Update(context.TODO(), routeExample, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	aviModel := ValidateModelCommon(t, g)

	pools := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].PoolRefs
	g.Expect(pools).To(gomega.HaveLen(1))
	for _, pool := range pools {
		if pool.Name == "cluster--foo.com_foo-default-foo-avisvc" || pool.Name == "cluster--foo.com_foo-default-foo-absvc2" {
			g.Expect(len(pool.Servers)).To(gomega.Equal(1))
		} else {
			t.Fatalf("unexpected pool: %s", pool.Name)
		}
	}

	poolgroups := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].PoolGroupRefs
	for _, pgmember := range poolgroups[0].Members {
		if *pgmember.PoolRef == "/api/pool?name=cluster--foo.com_foo-default-foo-avisvc" {
			g.Expect(*pgmember.PriorityLabel).To(gomega.Equal("foo.com/foo"))
			g.Expect(*pgmember.Ratio).To(gomega.Equal(uint32(100)))
		} else {
			t.Fatalf("unexpected pgmember: %s", *pgmember.PoolRef)
		}
	}

	VerifyRouteDeletion(t, g, aviModel, 0)
	TearDownTestForRoute(t, defaultModelName)

	integrationtest.DelSVC(t, "default", "absvc2")
	integrationtest.DelEPS(t, "default", "absvc2")
}

func TestAlternateBackendUpdatePath(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetUpTestForRoute(t, defaultModelName)
	integrationtest.CreateSVC(t, "default", "absvc2", corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPS(t, "default", "absvc2", false, false, "3.3.3")
	routeExample := FakeRoute{Path: "/foo"}.ABRoute()
	_, err := OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	routeExample = FakeRoute{Path: "/bar"}.ABRoute()
	_, err = OshiftClient.RouteV1().Routes(defaultNamespace).Update(context.TODO(), routeExample, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	aviModel := ValidateModelCommon(t, g)

	var nodes []*avinodes.AviVsNode
	g.Eventually(func() []*avinodes.AviPoolNode {
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return nodes[0].PoolRefs
	}, 50*time.Second).Should(gomega.HaveLen(2))
	pools := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].PoolRefs
	for _, pool := range pools {
		if pool.Name == "cluster--foo.com_bar-default-foo-avisvc" || pool.Name == "cluster--foo.com_bar-default-foo-absvc2" {
			g.Expect(len(pool.Servers)).To(gomega.Equal(1))
		} else {
			t.Fatalf("unexpected pool: %s", pool.Name)
		}
	}

	poolgroups := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].PoolGroupRefs
	for _, pgmember := range poolgroups[0].Members {
		if *pgmember.PoolRef == "/api/pool?name=cluster--foo.com_bar-default-foo-avisvc" {
			g.Expect(*pgmember.PriorityLabel).To(gomega.Equal("foo.com/bar"))
			g.Expect(*pgmember.Ratio).To(gomega.Equal(uint32(100)))
		} else if *pgmember.PoolRef == "/api/pool?name=cluster--foo.com_bar-default-foo-absvc2" {
			g.Expect(*pgmember.PriorityLabel).To(gomega.Equal("foo.com/bar"))
			g.Expect(*pgmember.Ratio).To(gomega.Equal(uint32(200)))
		} else {
			t.Fatalf("unexpected pgmember: %s", *pgmember.PoolRef)
		}
	}

	VerifyRouteDeletion(t, g, aviModel, 0)
	TearDownTestForRoute(t, defaultModelName)

	integrationtest.DelSVC(t, "default", "absvc2")
	integrationtest.DelEPS(t, "default", "absvc2")
}

func TestAlternateBackendUpdateWeight(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetUpTestForRoute(t, defaultModelName)
	integrationtest.CreateSVC(t, "default", "absvc2", corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPS(t, "default", "absvc2", false, false, "3.3.3")
	time.Sleep(2 * time.Second)
	routeExample := FakeRoute{Path: "/foo"}.ABRoute()
	_, err := OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	routeExample = FakeRoute{Path: "/foo"}.ABRoute(300)
	_, err = OshiftClient.RouteV1().Routes(defaultNamespace).Update(context.TODO(), routeExample, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	aviModel := ValidateModelCommon(t, g)

	pools := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].PoolRefs
	g.Expect(pools).To(gomega.HaveLen(2))
	for _, pool := range pools {
		if pool.Name == "cluster--foo.com_foo-default-foo-avisvc" || pool.Name == "cluster--foo.com_foo-default-foo-absvc2" {
			g.Expect(len(pool.Servers)).To(gomega.Equal(1))
		} else {
			t.Fatalf("unexpected pool: %s", pool.Name)
		}
	}

	poolgroups := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].PoolGroupRefs
	for _, pgmember := range poolgroups[0].Members {
		if *pgmember.PoolRef == "/api/pool?name=cluster--foo.com_foo-default-foo-avisvc" {
			g.Expect(*pgmember.PriorityLabel).To(gomega.Equal("foo.com/foo"))
			g.Expect(*pgmember.Ratio).To(gomega.Equal(uint32(100)))
		} else if *pgmember.PoolRef == "/api/pool?name=cluster--foo.com_foo-default-foo-absvc2" {
			g.Expect(*pgmember.PriorityLabel).To(gomega.Equal("foo.com/foo"))
			g.Expect(*pgmember.Ratio).To(gomega.Equal(uint32(300)))
		} else {
			t.Fatalf("unexpected pgmember: %s", *pgmember.PoolRef)
		}
	}

	VerifyRouteDeletion(t, g, aviModel, 0)
	TearDownTestForRoute(t, defaultModelName)

	integrationtest.DelSVC(t, "default", "absvc2")
	integrationtest.DelEPS(t, "default", "absvc2")
}

func TestRouteWithSubdomainNoHost(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	defaultNamespace = "default"
	SetUpTestForRoute(t, defaultModelName)
	path := "/foo"

	routeExample := FakeRoute{Path: path}.RouteWithSubdomainAndNoHost()
	_, err := OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExample, metav1.CreateOptions{})

	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	fqdnFromSubdomain := defaultSubdomain + "." + os.Getenv("DEFAULT_DOMAIN")
	aviModel := ValidateModelCommon(t, g)
	pool := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].PoolRefs[0]

	g.Eventually(func() int {
		pool = aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].PoolRefs[0]
		return len(pool.Servers)
	}, 10*time.Second).Should(gomega.Equal(1))

	g.Expect(pool.Name).To(gomega.Equal("cluster--foo.com_foo-default-foo-avisvc"))
	g.Expect(pool.PriorityLabel).To(gomega.Equal(fqdnFromSubdomain + path))

	poolgroups := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].PoolGroupRefs
	pgmember := poolgroups[0].Members[0]
	g.Expect(*pgmember.PoolRef).To(gomega.Equal("/api/pool?name=cluster--foo.com_foo-default-foo-avisvc"))
	g.Expect(*pgmember.PriorityLabel).To(gomega.Equal(fqdnFromSubdomain + path))

	VerifyRouteDeletion(t, g, aviModel, 0)
	TearDownTestForRoute(t, defaultModelName)
}
