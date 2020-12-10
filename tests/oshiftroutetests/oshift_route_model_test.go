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

package oshiftroutetests

import (
	"context"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	crdfake "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/client/v1alpha1/clientset/versioned/fake"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/k8s"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	avinodes "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/integrationtest"

	utils "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	"github.com/avinetworks/sdk/go/models"
	"github.com/onsi/gomega"
	routev1 "github.com/openshift/api/route/v1"
	oshiftfake "github.com/openshift/client-go/route/clientset/versioned/fake"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

var KubeClient *k8sfake.Clientset
var OshiftClient *oshiftfake.Clientset
var CRDClient *crdfake.Clientset
var ctrl *k8s.AviController

var defaultRouteName, defaultNamespace, defaultHostname, defaultService string
var defaultModelName string
var defaultKey, defaultValue string

// Candiate to move to lib
type FakeRoute struct {
	Name        string
	Namespace   string
	Hostname    string
	Path        string
	ServiceName string
	Backend2    string
}

func (rt FakeRoute) Route() *routev1.Route {
	if rt.Name == "" {
		rt.Name = defaultRouteName
	}
	if rt.Namespace == "" {
		rt.Namespace = defaultNamespace
	}
	if rt.Hostname == "" {
		rt.Hostname = defaultHostname
	}
	if rt.ServiceName == "" {
		rt.ServiceName = defaultService
	}
	weight := int32(100)
	routeExample := &routev1.Route{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Namespace:       rt.Namespace,
			Name:            rt.Name,
			ResourceVersion: "1",
		},
		Spec: routev1.RouteSpec{
			Host: rt.Hostname,
			To: routev1.RouteTargetReference{
				Kind:   "Service",
				Name:   rt.ServiceName,
				Weight: &weight,
			},
		},
	}
	if rt.Path != "" {
		routeExample.Spec.Path = rt.Path
	}
	return routeExample
}

func (rt FakeRoute) ABRoute(ratio ...int) *routev1.Route {
	routeExample := rt.Route()
	if rt.Backend2 == "" {
		rt.Backend2 = "absvc2"
	}
	weight2 := int32(200)
	if len(ratio) > 0 {
		weight2 = int32(ratio[0])
	}
	backend2 := routev1.RouteTargetReference{
		Kind:   "Service",
		Name:   rt.Backend2,
		Weight: &weight2,
	}
	routeExample.Spec.AlternateBackends = append(routeExample.Spec.AlternateBackends, backend2)
	return routeExample
}

func TestMain(m *testing.M) {
	os.Setenv("INGRESS_API", "extensionv1")
	os.Setenv("NETWORK_NAME", "net123")
	os.Setenv("CLUSTER_NAME", "cluster")
	os.Setenv("CLOUD_NAME", "CLOUD_VCENTER")
	os.Setenv("SEG_NAME", "Default-Group")
	os.Setenv("NODE_NETWORK_LIST", `[{"networkName":"net123","cidrs":["10.79.168.0/22"]}]`)
	KubeClient = k8sfake.NewSimpleClientset()
	CRDClient = crdfake.NewSimpleClientset()
	lib.SetCRDClientset(CRDClient)
	OshiftClient = oshiftfake.NewSimpleClientset()
	informersArg := make(map[string]interface{})
	informersArg[utils.INFORMERS_OPENSHIFT_CLIENT] = OshiftClient
	registeredInformers := []string{
		utils.ServiceInformer,
		utils.EndpointInformer,
		utils.RouteInformer,
		utils.SecretInformer,
		utils.NSInformer,
		utils.NodeInformer,
		utils.ConfigMapInformer,
	}
	utils.NewInformers(utils.KubeClientIntf{KubeClient}, registeredInformers, informersArg)
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
	wgSlowRetry := &sync.WaitGroup{}
	waitGroupMap["slowretry"] = wgSlowRetry
	wgGraph := &sync.WaitGroup{}
	waitGroupMap["graph"] = wgGraph

	AddConfigMap()
	ctrl.HandleConfigMap(informers, ctrlCh, stopCh, quickSyncCh)
	integrationtest.KubeClient = KubeClient
	integrationtest.AddDefaultIngressClass()
	defaultKey = "app"
	defaultValue = "migrate"
	SetupRouteNamespaceSync(defaultKey, defaultValue, "")

	go ctrl.InitController(informers, registeredInformers, ctrlCh, stopCh, quickSyncCh, waitGroupMap)

	defaultRouteName = "foo"
	defaultNamespace = "default"
	defaultHostname = "foo.com"
	defaultService = "avisvc"
	defaultModelName = "admin/cluster--Shared-L7-0"

	os.Exit(m.Run())
}

func AddConfigMap() {
	aviCM := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "avi-system",
			Name:      "avi-k8s-config",
		},
	}
	KubeClient.CoreV1().ConfigMaps("avi-system").Create(context.TODO(), aviCM, metav1.CreateOptions{})

	integrationtest.PollForSyncStart(ctrl, 10)
}
func AddLabelToNamespace(key, value, namespace, modelName string, t *testing.T) {

	nsLabel := map[string]string{
		key: value,
	}
	integrationtest.AddNamespace(namespace, nsLabel)
}

func SetUpTestForRoute(t *testing.T, modelName string) {
	os.Setenv("SHARD_VS_SIZE", "LARGE")
	os.Setenv("L7_SHARD_SCHEME", "hostname")
	AddLabelToNamespace(defaultKey, defaultValue, defaultNamespace, modelName, t)
	objects.SharedAviGraphLister().Delete(modelName)

	integrationtest.CreateSVC(t, defaultNamespace, "avisvc", corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEP(t, defaultNamespace, "avisvc", false, false, "1.1.1")
	integrationtest.PollForCompletion(t, modelName, 5)
}

func TearDownTestForRoute(t *testing.T, modelName string) {
	os.Setenv("SHARD_VS_SIZE", "")
	os.Setenv("CLOUD_NAME", "")

	objects.SharedAviGraphLister().Delete(modelName)
	integrationtest.DelSVC(t, "default", "avisvc")
	integrationtest.DelEP(t, "default", "avisvc")
}

func VerifyRouteDeletion(t *testing.T, g *gomega.WithT, aviModel interface{}, poolCount int, nsname ...string) {
	namespace, name := defaultNamespace, defaultRouteName
	if len(nsname) > 0 {
		namespace, name = strings.Split(nsname[0], "/")[0], strings.Split(nsname[0], "/")[1]
	}

	err := OshiftClient.RouteV1().Routes(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the route %v", err)
	}
	var nodes []*avinodes.AviVsNode
	g.Eventually(func() []*avinodes.AviPoolNode {
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return nodes[0].PoolRefs
	}, 50*time.Second).Should(gomega.HaveLen(poolCount))

	g.Eventually(func() []*models.PoolGroupMember {
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return nodes[0].PoolGroupRefs[0].Members
	}, 50*time.Second).Should(gomega.HaveLen(poolCount))
}

func ValidateModelCommon(t *testing.T, g *gomega.GomegaWithT) interface{} {

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(defaultModelName)
		return found
	}, 30*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(defaultModelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()

	g.Expect(len(nodes)).To(gomega.Equal(1))
	g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
	g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))

	g.Expect(nodes[0].SharedVS).To(gomega.Equal(true))
	dsNodes := aviModel.(*avinodes.AviObjectGraph).GetAviHTTPDSNode()
	g.Expect(len(dsNodes)).To(gomega.Equal(1))

	g.Expect(len(nodes[0].PoolGroupRefs)).To(gomega.Equal(1))

	return aviModel
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
	integrationtest.CreateSVC(t, "default", "newsvc", corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEP(t, "default", "newsvc", false, false, "3.3.3")
	routeExample := FakeRoute{Path: "/foo", ServiceName: "newsvc"}.Route()
	_, err := OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	aviModel := ValidateModelCommon(t, g)

	integrationtest.DelSVC(t, "default", "newsvc")
	integrationtest.DelEP(t, "default", "newsvc")

	g.Eventually(func() int {
		_, aviModel = objects.SharedAviGraphLister().Get(defaultModelName)
		pool := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].PoolRefs[0]
		return len(pool.Servers)
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

	integrationtest.CreateSVC(t, "default", "newsvc", corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEP(t, "default", "newsvc", false, false, "3.3.3")

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
	integrationtest.DelEP(t, "default", "newsvc")
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

	integrationtest.ScaleCreateEP(t, "default", "avisvc")
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
	pools := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].PoolRefs

	g.Expect(pools).To(gomega.HaveLen(2))
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
	integrationtest.CreateSVC(t, "default", "absvc2", corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEP(t, "default", "absvc2", false, false, "3.3.3")
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
			g.Expect(*pgmember.Ratio).To(gomega.Equal(int32(100)))
		} else if *pgmember.PoolRef == "/api/pool?name=cluster--foo.com-default-foo-absvc2" {
			g.Expect(*pgmember.PriorityLabel).To(gomega.Equal("foo.com"))
			g.Expect(*pgmember.Ratio).To(gomega.Equal(int32(200)))
		} else {
			t.Fatalf("unexpected pgmember: %s", *pgmember.PoolRef)
		}
	}

	VerifyRouteDeletion(t, g, aviModel, 0)
	TearDownTestForRoute(t, defaultModelName)

	integrationtest.DelSVC(t, "default", "absvc2")
	integrationtest.DelEP(t, "default", "absvc2")
}

func TestAlternateBackendDefaultPath(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetUpTestForRoute(t, defaultModelName)
	integrationtest.CreateSVC(t, "default", "absvc2", corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEP(t, "default", "absvc2", false, false, "3.3.3")
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
			g.Expect(*pgmember.Ratio).To(gomega.Equal(int32(100)))
		} else if *pgmember.PoolRef == "/api/pool?name=cluster--foo.com_foo-default-foo-absvc2" {
			g.Expect(*pgmember.PriorityLabel).To(gomega.Equal("foo.com/foo"))
			g.Expect(*pgmember.Ratio).To(gomega.Equal(int32(200)))
		} else {
			t.Fatalf("unexpected pgmember: %s", *pgmember.PoolRef)
		}
	}

	VerifyRouteDeletion(t, g, aviModel, 0)
	TearDownTestForRoute(t, defaultModelName)

	integrationtest.DelSVC(t, "default", "absvc2")
	integrationtest.DelEP(t, "default", "absvc2")
}

func TestRemoveAlternateBackend(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetUpTestForRoute(t, defaultModelName)
	integrationtest.CreateSVC(t, "default", "absvc2", corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEP(t, "default", "absvc2", false, false, "3.3.3")
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
			g.Expect(*pgmember.Ratio).To(gomega.Equal(int32(100)))
		} else {
			t.Fatalf("unexpected pgmember: %s", *pgmember.PoolRef)
		}
	}

	VerifyRouteDeletion(t, g, aviModel, 0)
	TearDownTestForRoute(t, defaultModelName)

	integrationtest.DelSVC(t, "default", "absvc2")
	integrationtest.DelEP(t, "default", "absvc2")
}

func TestAlternateBackendUpdatePath(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetUpTestForRoute(t, defaultModelName)
	integrationtest.CreateSVC(t, "default", "absvc2", corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEP(t, "default", "absvc2", false, false, "3.3.3")
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
			g.Expect(*pgmember.Ratio).To(gomega.Equal(int32(100)))
		} else if *pgmember.PoolRef == "/api/pool?name=cluster--foo.com_bar-default-foo-absvc2" {
			g.Expect(*pgmember.PriorityLabel).To(gomega.Equal("foo.com/bar"))
			g.Expect(*pgmember.Ratio).To(gomega.Equal(int32(200)))
		} else {
			t.Fatalf("unexpected pgmember: %s", *pgmember.PoolRef)
		}
	}

	VerifyRouteDeletion(t, g, aviModel, 0)
	TearDownTestForRoute(t, defaultModelName)

	integrationtest.DelSVC(t, "default", "absvc2")
	integrationtest.DelEP(t, "default", "absvc2")
}

func TestAlternateBackendUpdateWeight(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetUpTestForRoute(t, defaultModelName)
	integrationtest.CreateSVC(t, "default", "absvc2", corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEP(t, "default", "absvc2", false, false, "3.3.3")
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
			g.Expect(*pgmember.Ratio).To(gomega.Equal(int32(100)))
		} else if *pgmember.PoolRef == "/api/pool?name=cluster--foo.com_foo-default-foo-absvc2" {
			g.Expect(*pgmember.PriorityLabel).To(gomega.Equal("foo.com/foo"))
			g.Expect(*pgmember.Ratio).To(gomega.Equal(int32(300)))
		} else {
			t.Fatalf("unexpected pgmember: %s", *pgmember.PoolRef)
		}
	}

	VerifyRouteDeletion(t, g, aviModel, 0)
	TearDownTestForRoute(t, defaultModelName)

	integrationtest.DelSVC(t, "default", "absvc2")
	integrationtest.DelEP(t, "default", "absvc2")
}
