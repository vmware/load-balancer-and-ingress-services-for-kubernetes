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
	"testing"
	"time"

	avinodes "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/integrationtest"

	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func SetUpTestForRouteInNodePort(t *testing.T, modelName string, externalTrafficPolicy string) {
	AddLabelToNamespace(defaultKey, defaultValue, defaultNamespace, modelName, t)
	for retry := 0; retry < 3; retry++ {
		if !utils.CheckIfNamespaceAccepted(defaultNamespace) {
			time.Sleep(1 * time.Second)
		}
	}
	objects.SharedAviGraphLister().Delete(modelName)
	if externalTrafficPolicy == "" {
		integrationtest.CreateSVC(t, "default", "avisvc", corev1.ProtocolTCP, corev1.ServiceTypeNodePort, false)
	} else {
		integrationtest.CreateSvcWithExternalTrafficPolicy(t, "default", "avisvc", corev1.ProtocolTCP, corev1.ServiceTypeNodePort, false, externalTrafficPolicy)
	}
}

func TearDownTestForRouteInNodePort(t *testing.T, modelName string) {
	objects.SharedAviGraphLister().Delete(modelName)
	integrationtest.DelSVC(t, "default", "avisvc")
}

// TestRouteDefaultPathInNodePort tests route creation with nodeport service in NodePort mode with no path .

func TestRouteNoPathInNodePort(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	integrationtest.SetNodePortMode()
	defer integrationtest.SetClusterIPMode()
	nodeIP := "10.1.1.2"
	integrationtest.CreateNode(t, "testNodeNP", nodeIP)
	defer integrationtest.DeleteNode(t, "testNodeNP")

	SetUpTestForRouteInNodePort(t, defaultModelName, "")

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
	TearDownTestForRouteInNodePort(t, defaultModelName)
}

// TestRouteDefaultPathInNodePort tests route creation with nodeport service in NodePort mode with default path.
func TestRouteDefaultPathInNodePort(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	integrationtest.SetNodePortMode()
	defer integrationtest.SetClusterIPMode()
	nodeIP := "10.1.1.2"
	integrationtest.CreateNode(t, "testNodeNP", nodeIP)
	defer integrationtest.DeleteNode(t, "testNodeNP")

	SetUpTestForRouteInNodePort(t, defaultModelName, "")
	routeExample := FakeRoute{Path: "/foo"}.Route()
	_, err := OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	aviModel := ValidateModelCommon(t, g)
	pool := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].PoolRefs[0]

	// pool server is added for testNodeNP node even though endpointslice/endpoint does not exist
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
	TearDownTestForRouteInNodePort(t, defaultModelName)
}

// TestRouteNodePortExternalTrafficPolicyLocal checks if pool servers are populated in model only for nodes that are running the app pod.
func TestRouteNodePortExternalTrafficPolicyLocal(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	nodeName := "testNodeNP"
	integrationtest.SetNodePortMode()
	defer integrationtest.SetClusterIPMode()
	nodeIP := "10.1.1.2"
	integrationtest.CreateNode(t, nodeName, nodeIP)
	defer integrationtest.DeleteNode(t, nodeName)

	SetUpTestForRouteInNodePort(t, defaultModelName, "Local")
	routeExample := FakeRoute{Path: "/foo", TargetPort: 8080}.Route()
	_, err := OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	aviModel := ValidateModelCommon(t, g)
	pool := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].PoolRefs[0]

	g.Eventually(func() int {
		pool = aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].PoolRefs[0]
		return len(pool.Servers)
	}, 10*time.Second).Should(gomega.Equal(0))

	g.Expect(pool.Name).To(gomega.Equal("cluster--foo.com_foo-default-foo-avisvc"))
	g.Expect(pool.PriorityLabel).To(gomega.Equal("foo.com/foo"))

	poolgroups := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].PoolGroupRefs
	pgmember := poolgroups[0].Members[0]
	g.Expect(*pgmember.PoolRef).To(gomega.Equal("/api/pool?name=cluster--foo.com_foo-default-foo-avisvc"))
	g.Expect(*pgmember.PriorityLabel).To(gomega.Equal("foo.com/foo"))

	integrationtest.CreateEPorEPSNodeName(t, "default", "avisvc", false, false, "1.1.1", nodeName)
	// After creating the endpointslice/endpoint, pool server should be added for testNodeNP node
	g.Eventually(func() int {
		pool = aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].PoolRefs[0]
		return len(pool.Servers)
	}, 30*time.Second).Should(gomega.Equal(1))

	VerifyRouteDeletion(t, g, aviModel, 0)
	integrationtest.DelEPorEPS(t, "default", "avisvc")
	TearDownTestForRouteInNodePort(t, defaultModelName)
}

// TestRouteClusterIPSvcDefaultPathInNodePort tests route creation with Cluster IP in NodePort mode.
func TestRouteClusterIPSvcDefaultPathInNodePort(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	integrationtest.SetNodePortMode()
	defer integrationtest.SetClusterIPMode()
	nodeIP := "10.1.1.2"
	integrationtest.CreateNode(t, "testNodeNP", nodeIP)
	defer integrationtest.DeleteNode(t, "testNodeNP")

	SetUpTestForRoute(t, defaultModelName)
	routeExample := FakeRoute{Path: "/foo"}.Route()
	_, err := OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	aviModel := ValidateModelCommon(t, g)
	pool := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].PoolRefs[0]

	// since the route is referred by a service of type cluster IP, no pool servers should be created in AVI
	g.Eventually(func() int {
		pool = aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].PoolRefs[0]
		return len(pool.Servers)
	}, 10*time.Second).Should(gomega.Equal(0))

	g.Expect(pool.Name).To(gomega.Equal("cluster--foo.com_foo-default-foo-avisvc"))
	g.Expect(pool.PriorityLabel).To(gomega.Equal("foo.com/foo"))

	poolgroups := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].PoolGroupRefs
	pgmember := poolgroups[0].Members[0]
	g.Expect(*pgmember.PoolRef).To(gomega.Equal("/api/pool?name=cluster--foo.com_foo-default-foo-avisvc"))
	g.Expect(*pgmember.PriorityLabel).To(gomega.Equal("foo.com/foo"))

	VerifyRouteDeletion(t, g, aviModel, 0)
	TearDownTestForRoute(t, defaultModelName)
}

func TestRouteBadServiceInNodePort(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	integrationtest.SetNodePortMode()
	defer integrationtest.SetClusterIPMode()
	nodeIP := "10.1.1.2"
	integrationtest.CreateNode(t, "testNodeNP", nodeIP)
	defer integrationtest.DeleteNode(t, "testNodeNP")

	SetUpTestForRouteInNodePort(t, defaultModelName, "")
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
	TearDownTestForRouteInNodePort(t, defaultModelName)
}

// TestRouteScaleEndpointInNodePort tests that scaling an endpoint has no effect on pool servers in nodeport
func TestRouteScaleEndpointInNodePort(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	integrationtest.SetNodePortMode()
	defer integrationtest.SetClusterIPMode()
	nodeIP := "10.1.1.2"
	integrationtest.CreateNode(t, "testNodeNP", nodeIP)
	defer integrationtest.DeleteNode(t, "testNodeNP")

	SetUpTestForRoute(t, defaultModelName)
	routeExample := FakeRoute{Path: "/foo"}.Route()
	_, err := OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	aviModel := ValidateModelCommon(t, g)
	pool := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].PoolRefs[0]

	integrationtest.ScaleCreateEPorEPS(t, "default", "avisvc")
	g.Eventually(func() int {
		pool = aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].PoolRefs[0]
		return len(pool.Servers)
	}, 10*time.Second).Should(gomega.Equal(0))

	g.Expect(pool.Name).To(gomega.Equal("cluster--foo.com_foo-default-foo-avisvc"))
	g.Expect(pool.PriorityLabel).To(gomega.Equal("foo.com/foo"))

	poolgroups := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].PoolGroupRefs
	pgmember := poolgroups[0].Members[0]
	g.Expect(*pgmember.PoolRef).To(gomega.Equal("/api/pool?name=cluster--foo.com_foo-default-foo-avisvc"))
	g.Expect(*pgmember.PriorityLabel).To(gomega.Equal("foo.com/foo"))

	VerifyRouteDeletion(t, g, aviModel, 0)
	TearDownTestForRoute(t, defaultModelName)
}

func TestMultiRouteSameHostInNodePort(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	integrationtest.SetNodePortMode()
	defer integrationtest.SetClusterIPMode()
	nodeIP := "10.1.1.2"
	integrationtest.CreateNode(t, "testNodeNP", nodeIP)
	defer integrationtest.DeleteNode(t, "testNodeNP")

	SetUpTestForRouteInNodePort(t, defaultModelName, "")
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

	g.Eventually(func() int {
		pools = aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].PoolRefs
		return len(pools)
	}, 10*time.Second).Should(gomega.Equal(2))
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
	TearDownTestForRouteInNodePort(t, defaultModelName)
}

func TestRouteUpdatePathInNodePort(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	integrationtest.SetNodePortMode()
	defer integrationtest.SetClusterIPMode()
	nodeIP := "10.1.1.2"
	integrationtest.CreateNode(t, "testNodeNP", nodeIP)
	defer integrationtest.DeleteNode(t, "testNodeNP")

	SetUpTestForRouteInNodePort(t, defaultModelName, "")
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
		pool = aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].PoolRefs[0]
		return len(pool.Servers)
	}, 10*time.Second).Should(gomega.Equal(1))

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
	TearDownTestForRouteInNodePort(t, defaultModelName)
}

func TestAlternateBackendNoPathInNodePort(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	integrationtest.SetNodePortMode()
	defer integrationtest.SetClusterIPMode()
	nodeIP := "10.1.1.2"
	integrationtest.CreateNode(t, "testNodeNP", nodeIP)
	defer integrationtest.DeleteNode(t, "testNodeNP")

	SetUpTestForRouteInNodePort(t, defaultModelName, "")
	integrationtest.CreateSVC(t, "default", "absvc2", corev1.ProtocolTCP, corev1.ServiceTypeNodePort, false)
	routeExample := FakeRoute{}.ABRoute()
	_, err := OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	aviModel := ValidateModelCommon(t, g)

	g.Eventually(func() int {
		pools := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].PoolRefs
		return len(pools)
	}, 60*time.Second).Should(gomega.Equal(2))
	pools := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].PoolRefs
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
	TearDownTestForRouteInNodePort(t, defaultModelName)

	integrationtest.DelSVC(t, "default", "absvc2")
}

func TestAlternateBackendDefaultPathInNodePort(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	integrationtest.SetNodePortMode()
	defer integrationtest.SetClusterIPMode()
	nodeIP := "10.1.1.2"
	integrationtest.CreateNode(t, "testNodeNP", nodeIP)
	defer integrationtest.DeleteNode(t, "testNodeNP")

	SetUpTestForRouteInNodePort(t, defaultModelName, "")
	integrationtest.CreateSVC(t, "default", "absvc2", corev1.ProtocolTCP, corev1.ServiceTypeNodePort, false)
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
	TearDownTestForRouteInNodePort(t, defaultModelName)

	integrationtest.DelSVC(t, "default", "absvc2")
}

// TestNodeCUDForOshiftRouteInNodePort tests Node CUD with route creation in NodePort mode.
func TestNodeCUDForOshiftRouteInNodePort(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	integrationtest.SetNodePortMode()
	defer integrationtest.SetClusterIPMode()
	nodeIP := "10.1.1.2"
	integrationtest.CreateNode(t, "testNodeNP", nodeIP)
	defer integrationtest.DeleteNode(t, "testNodeNP")

	SetUpTestForRouteInNodePort(t, defaultModelName, "")
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

	// Update the Node's resource version
	objects.SharedAviGraphLister().Delete(defaultModelName)
	nodeExample := (integrationtest.FakeNode{
		Name:     "testNodeNP",
		PodCIDR:  "10.244.0.0/24",
		PodCIDRs: []string{"10.244.0.0/24"},
		Version:  "1",
		NodeIP:   nodeIP,
	}).Node()
	nodeExample.ResourceVersion = "2"

	_, err = KubeClient.CoreV1().Nodes().Update(context.TODO(), nodeExample, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in adding Node: %v", err)
	}
	integrationtest.PollForCompletion(t, defaultModelName, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(defaultModelName)
	if found {
		// model should not be sent for this update as only the resource ver is changed.
		t.Fatalf("Model found for node add %v", defaultModelName)
	}

	// Add another node and check if that is added in the pool servers
	nodeIP2 := "10.1.1.3"
	integrationtest.CreateNode(t, "testNodeNP2", nodeIP2)

	aviModel = ValidateModelCommon(t, g)
	pool = aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].PoolRefs[0]
	g.Eventually(func() int {
		pool = aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].PoolRefs[0]
		return len(pool.Servers)
	}, 10*time.Second).Should(gomega.Equal(2))

	// Delete the previously added node and check if that is removed from pool servers
	integrationtest.DeleteNode(t, "testNodeNP2")
	aviModel = ValidateModelCommon(t, g)
	pool = aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].PoolRefs[0]
	g.Eventually(func() int {
		pool = aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].PoolRefs[0]
		return len(pool.Servers)
	}, 10*time.Second).Should(gomega.Equal(1))

	VerifyRouteDeletion(t, g, aviModel, 0)
	TearDownTestForRouteInNodePort(t, defaultModelName)
}

func TestSecureRouteInNodePort(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	integrationtest.SetNodePortMode()
	defer integrationtest.SetClusterIPMode()
	nodeIP := "10.1.1.2"
	integrationtest.CreateNode(t, "testNodeNP", nodeIP)
	defer integrationtest.DeleteNode(t, "testNodeNP")

	SetUpTestForRouteInNodePort(t, defaultModelName, "")
	routeExample := FakeRoute{Path: "/foo"}.SecureRoute()
	_, err := OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	aviModel := ValidateSniModel(t, g, defaultModelName)

	g.Expect(aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes).To(gomega.HaveLen(1))
	sniVS := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes[0]
	g.Eventually(func() string {
		sniVS = aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes[0]
		return sniVS.VHDomainNames[0]
	}, 50*time.Second).Should(gomega.Equal(defaultHostname))
	VerifySniNode(g, sniVS)

	VerifySecureRouteDeletion(t, g, defaultModelName, 0, 0)
	TearDownTestForRouteInNodePort(t, defaultModelName)
}

func TestSecureToInsecureRouteInNodePort(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	integrationtest.SetNodePortMode()
	defer integrationtest.SetClusterIPMode()
	nodeIP := "10.1.1.2"
	integrationtest.CreateNode(t, "testNodeNP", nodeIP)
	defer integrationtest.DeleteNode(t, "testNodeNP")

	SetUpTestForRouteInNodePort(t, defaultModelName, "")
	routeExample := FakeRoute{Path: "/foo"}.SecureRoute()
	_, err := OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	routeExample = FakeRoute{Path: "/foo"}.Route()
	_, err = OshiftClient.RouteV1().Routes(defaultNamespace).Update(context.TODO(), routeExample, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating route: %v", err)
	}

	aviModel := ValidateModelCommon(t, g)

	g.Eventually(func() int {
		sniNodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes
		return len(sniNodes)
	}, 50*time.Second).Should(gomega.Equal(0))

	VerifyRouteDeletion(t, g, aviModel, 0)
	TearDownTestForRouteInNodePort(t, defaultModelName)
}

func TestSecureRouteMultiNamespaceInNodePort(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	integrationtest.DeleteNamespace("test")

	integrationtest.SetNodePortMode()
	defer integrationtest.SetClusterIPMode()
	nodeIP := "10.1.1.2"
	integrationtest.CreateNode(t, "testNodeNP", nodeIP)
	defer integrationtest.DeleteNode(t, "testNodeNP")

	SetUpTestForRouteInNodePort(t, defaultModelName, "")
	route1 := FakeRoute{Path: "/foo"}.SecureRoute()
	_, err := OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), route1, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}
	AddLabelToNamespace(defaultKey, defaultValue, "test", defaultModelName, t)
	defer integrationtest.DeleteNamespace("test")
	if !utils.CheckIfNamespaceAccepted("test") {
		time.Sleep(time.Second * 2)
	}
	integrationtest.CreateSVC(t, "test", "avisvc", corev1.ProtocolTCP, corev1.ServiceTypeNodePort, false)
	route2 := FakeRoute{Namespace: "test", Path: "/bar"}.SecureRoute()
	_, err = OshiftClient.RouteV1().Routes("test").Create(context.TODO(), route2, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	aviModel := ValidateSniModel(t, g, defaultModelName)

	CheckMultiSNIMultiNS(t, g, aviModel, 2, 1)

	err = OshiftClient.RouteV1().Routes("test").Delete(context.TODO(), defaultRouteName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the route %v", err)
	}
	VerifySecureRouteDeletion(t, g, defaultModelName, 0, 0)
	TearDownTestForRouteInNodePort(t, defaultModelName)
	integrationtest.DelSVC(t, "test", "avisvc")
}

func TestPassthroughRouteInNodePort(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	integrationtest.SetNodePortMode()
	defer integrationtest.SetClusterIPMode()
	nodeIP := "10.1.1.2"
	integrationtest.CreateNode(t, "testNodeNP", nodeIP)
	defer integrationtest.DeleteNode(t, "testNodeNP")

	SetUpTestForRouteInNodePort(t, defaultModelName, "")
	routeExample := FakeRoute{Path: "/foo"}.PassthroughRoute()
	_, err := OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	aviModel := ValidatePassthroughModel(t, g, DefaultPassthroughModel)
	graph := aviModel.(*avinodes.AviObjectGraph)
	vs := graph.GetAviVS()[0]
	VerifyOnePasthrough(t, g, vs)

	g.Expect(vs.PassthroughChildNodes).To(gomega.HaveLen(0))

	VerifyPassthroughRouteDeletion(t, g, DefaultPassthroughModel, 0, 0)
	TearDownTestForRouteInNodePort(t, DefaultPassthroughModel)
}

func TestPassthroughRouteDelSvcInNodePort(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	integrationtest.SetNodePortMode()
	defer integrationtest.SetClusterIPMode()
	nodeIP := "10.1.1.2"
	integrationtest.CreateNode(t, "testNodeNP", nodeIP)
	defer integrationtest.DeleteNode(t, "testNodeNP")

	SetUpTestForRouteInNodePort(t, defaultModelName, "")
	routeExample := FakeRoute{Path: "/foo"}.PassthroughRoute()
	_, err := OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	aviModel := ValidatePassthroughModel(t, g, DefaultPassthroughModel)
	graph := aviModel.(*avinodes.AviObjectGraph)
	vs := graph.GetAviVS()[0]

	VerifyOnePasthrough(t, g, vs)

	// verify server is deleted from pool after deleting service
	integrationtest.DelSVC(t, "default", "avisvc")
	g.Eventually(func() int {
		return len(vs.PoolRefs[0].Servers)
	}, 60*time.Second).Should(gomega.Equal(0))

	VerifyPassthroughRouteDeletion(t, g, DefaultPassthroughModel, 0, 0)
	objects.SharedAviGraphLister().Delete(DefaultPassthroughModel)
}
