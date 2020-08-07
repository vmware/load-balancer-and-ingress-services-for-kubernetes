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
	"os"
	"testing"
	"time"

	avinodes "ako/internal/nodes"
	"ako/internal/objects"
	"ako/tests/integrationtest"

	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
)

func SetUpTestForRouteInNodePort(t *testing.T, modelName string) {
	os.Setenv("SHARD_VS_SIZE", "LARGE")
	os.Setenv("L7_SHARD_SCHEME", "hostname")

	objects.SharedAviGraphLister().Delete(modelName)
	integrationtest.CreateSVC(t, "default", "avisvc", corev1.ServiceTypeNodePort, false)
}

func TearDownTestForRouteInNodePort(t *testing.T, modelName string) {
	os.Setenv("SHARD_VS_SIZE", "")
	os.Setenv("CLOUD_NAME", "")

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

	SetUpTestForRouteInNodePort(t, DefaultModelName)

	routeExample := FakeRoute{}.Route()
	_, err := OshiftClient.RouteV1().Routes(DefaultNamespace).Create(routeExample)

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
	TearDownTestForRouteInNodePort(t, DefaultModelName)
}

// TestRouteDefaultPathInNodePort tests route creation with nodeport service in NodePort mode with default path.
func TestRouteDefaultPathInNodePort(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	integrationtest.SetNodePortMode()
	defer integrationtest.SetClusterIPMode()
	nodeIP := "10.1.1.2"
	integrationtest.CreateNode(t, "testNodeNP", nodeIP)
	defer integrationtest.DeleteNode(t, "testNodeNP")

	SetUpTestForRouteInNodePort(t, DefaultModelName)
	routeExample := FakeRoute{Path: "/foo"}.Route()
	_, err := OshiftClient.RouteV1().Routes(DefaultNamespace).Create(routeExample)
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
	TearDownTestForRouteInNodePort(t, DefaultModelName)
}

// TestRouteClusterIPSvcDefaultPathInNodePort tests route creation with Cluster IP in NodePort mode.
func TestRouteClusterIPSvcDefaultPathInNodePort(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	integrationtest.SetNodePortMode()
	defer integrationtest.SetClusterIPMode()
	nodeIP := "10.1.1.2"
	integrationtest.CreateNode(t, "testNodeNP", nodeIP)
	defer integrationtest.DeleteNode(t, "testNodeNP")

	SetUpTestForRoute(t, DefaultModelName)
	routeExample := FakeRoute{Path: "/foo"}.Route()
	_, err := OshiftClient.RouteV1().Routes(DefaultNamespace).Create(routeExample)
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
	TearDownTestForRoute(t, DefaultModelName)
}

func TestRouteBadServiceInNodePort(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	integrationtest.SetNodePortMode()
	defer integrationtest.SetClusterIPMode()
	nodeIP := "10.1.1.2"
	integrationtest.CreateNode(t, "testNodeNP", nodeIP)
	defer integrationtest.DeleteNode(t, "testNodeNP")

	SetUpTestForRouteInNodePort(t, DefaultModelName)
	routeExample := FakeRoute{Path: "/foo", ServiceName: "badsvc"}.Route()
	_, err := OshiftClient.RouteV1().Routes(DefaultNamespace).Create(routeExample)
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	aviModel := ValidateModelCommon(t, g)
	pool := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].PoolRefs[0]
	g.Expect(pool.Name).To(gomega.Equal("cluster--foo.com_foo-default-foo-badsvc"))
	g.Expect(pool.PriorityLabel).To(gomega.Equal("foo.com/foo"))
	g.Expect(len(pool.Servers)).To(gomega.Equal(0))

	VerifyRouteDeletion(t, g, aviModel, 0)
	TearDownTestForRouteInNodePort(t, DefaultModelName)
}

// TestRouteScaleEndpointInNodePort tests that scaling an endpoint has no effect on pool servers in nodeport
func TestRouteScaleEndpointInNodePort(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	integrationtest.SetNodePortMode()
	defer integrationtest.SetClusterIPMode()
	nodeIP := "10.1.1.2"
	integrationtest.CreateNode(t, "testNodeNP", nodeIP)
	defer integrationtest.DeleteNode(t, "testNodeNP")

	SetUpTestForRoute(t, DefaultModelName)
	routeExample := FakeRoute{Path: "/foo"}.Route()
	_, err := OshiftClient.RouteV1().Routes(DefaultNamespace).Create(routeExample)
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	aviModel := ValidateModelCommon(t, g)
	pool := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].PoolRefs[0]

	integrationtest.ScaleCreateEP(t, "default", "avisvc")
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
	TearDownTestForRoute(t, DefaultModelName)
}

func TestMultiRouteSameHostInNodePort(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	integrationtest.SetNodePortMode()
	defer integrationtest.SetClusterIPMode()
	nodeIP := "10.1.1.2"
	integrationtest.CreateNode(t, "testNodeNP", nodeIP)
	defer integrationtest.DeleteNode(t, "testNodeNP")

	SetUpTestForRouteInNodePort(t, DefaultModelName)
	routeExample1 := FakeRoute{Path: "/foo", ServiceName: "avisvc"}.Route()
	_, err := OshiftClient.RouteV1().Routes(DefaultNamespace).Create(routeExample1)
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	routeExample2 := FakeRoute{Name: "bar", Path: "/bar", ServiceName: "avisvc"}.Route()
	_, err = OshiftClient.RouteV1().Routes(DefaultNamespace).Create(routeExample2)
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

	err = OshiftClient.RouteV1().Routes(DefaultNamespace).Delete("bar", nil)
	if err != nil {
		t.Fatalf("Couldn't DELETE the route %v", err)
	}

	VerifyRouteDeletion(t, g, aviModel, 0)
	TearDownTestForRouteInNodePort(t, DefaultModelName)
}

func TestRouteUpdatePathInNodePort(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	integrationtest.SetNodePortMode()
	defer integrationtest.SetClusterIPMode()
	nodeIP := "10.1.1.2"
	integrationtest.CreateNode(t, "testNodeNP", nodeIP)
	defer integrationtest.DeleteNode(t, "testNodeNP")

	SetUpTestForRouteInNodePort(t, DefaultModelName)
	routeExample := FakeRoute{Path: "/foo"}.Route()
	_, err := OshiftClient.RouteV1().Routes(DefaultNamespace).Create(routeExample)
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	routeExample = FakeRoute{Path: "/bar"}.Route()
	_, err = OshiftClient.RouteV1().Routes(DefaultNamespace).Update(routeExample)
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
	TearDownTestForRouteInNodePort(t, DefaultModelName)
}

func TestAlternateBackendNoPathInNodePort(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	integrationtest.SetNodePortMode()
	defer integrationtest.SetClusterIPMode()
	nodeIP := "10.1.1.2"
	integrationtest.CreateNode(t, "testNodeNP", nodeIP)
	defer integrationtest.DeleteNode(t, "testNodeNP")

	SetUpTestForRouteInNodePort(t, DefaultModelName)
	integrationtest.CreateSVC(t, "default", "absvc2", corev1.ServiceTypeNodePort, false)
	routeExample := FakeRoute{}.ABRoute()
	_, err := OshiftClient.RouteV1().Routes(DefaultNamespace).Create(routeExample)
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
	TearDownTestForRouteInNodePort(t, DefaultModelName)

	integrationtest.DelSVC(t, "default", "absvc2")
}

func TestAlternateBackendDefaultPathInNodePort(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	integrationtest.SetNodePortMode()
	defer integrationtest.SetClusterIPMode()
	nodeIP := "10.1.1.2"
	integrationtest.CreateNode(t, "testNodeNP", nodeIP)
	defer integrationtest.DeleteNode(t, "testNodeNP")

	SetUpTestForRouteInNodePort(t, DefaultModelName)
	integrationtest.CreateSVC(t, "default", "absvc2", corev1.ServiceTypeNodePort, false)
	routeExample := FakeRoute{Path: "/foo"}.ABRoute()
	_, err := OshiftClient.RouteV1().Routes(DefaultNamespace).Create(routeExample)
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
	TearDownTestForRouteInNodePort(t, DefaultModelName)

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

	SetUpTestForRouteInNodePort(t, DefaultModelName)
	routeExample := FakeRoute{Path: "/foo"}.Route()
	_, err := OshiftClient.RouteV1().Routes(DefaultNamespace).Create(routeExample)
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
	objects.SharedAviGraphLister().Delete(DefaultModelName)
	nodeExample := (integrationtest.FakeNode{
		Name:    "testNodeNP",
		PodCIDR: "10.244.0.0/24",
		Version: "1",
		NodeIP:  nodeIP,
	}).Node()
	nodeExample.ResourceVersion = "2"

	_, err = KubeClient.CoreV1().Nodes().Update(nodeExample)
	if err != nil {
		t.Fatalf("error in adding Node: %v", err)
	}
	integrationtest.PollForCompletion(t, DefaultModelName, 5)
	found, aviModel := objects.SharedAviGraphLister().Get(DefaultModelName)
	if found {
		// model should not be sent for this update as only the resource ver is changed.
		t.Fatalf("Model found for node add %v", DefaultModelName)
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
	TearDownTestForRouteInNodePort(t, DefaultModelName)
}

func TestSecureRouteInNodePort(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	integrationtest.SetNodePortMode()
	defer integrationtest.SetClusterIPMode()
	nodeIP := "10.1.1.2"
	integrationtest.CreateNode(t, "testNodeNP", nodeIP)
	defer integrationtest.DeleteNode(t, "testNodeNP")

	SetUpTestForRouteInNodePort(t, DefaultModelName)
	routeExample := FakeRoute{Path: "/foo"}.SecureRoute()
	_, err := OshiftClient.RouteV1().Routes(DefaultNamespace).Create(routeExample)
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	aviModel := ValidateSniModel(t, g, DefaultModelName)

	g.Expect(aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes).To(gomega.HaveLen(1))
	sniVS := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes[0]
	g.Eventually(func() string {
		sniVS = aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes[0]
		return sniVS.VHDomainNames[0]
	}, 20*time.Second).Should(gomega.Equal(DefaultHostname))
	VerifySniNode(g, sniVS)

	VerifySecureRouteDeletion(t, g, DefaultModelName, 0, 0)
	TearDownTestForRouteInNodePort(t, DefaultModelName)
}

func TestSecureToInsecureRouteInNodePort(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	integrationtest.SetNodePortMode()
	defer integrationtest.SetClusterIPMode()
	nodeIP := "10.1.1.2"
	integrationtest.CreateNode(t, "testNodeNP", nodeIP)
	defer integrationtest.DeleteNode(t, "testNodeNP")

	SetUpTestForRouteInNodePort(t, DefaultModelName)
	routeExample := FakeRoute{Path: "/foo"}.SecureRoute()
	_, err := OshiftClient.RouteV1().Routes(DefaultNamespace).Create(routeExample)
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	routeExample = FakeRoute{Path: "/foo"}.Route()
	_, err = OshiftClient.RouteV1().Routes(DefaultNamespace).Update(routeExample)
	if err != nil {
		t.Fatalf("error in updating route: %v", err)
	}

	aviModel := ValidateModelCommon(t, g)

	g.Eventually(func() int {
		sniNodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes
		return len(sniNodes)
	}, 20*time.Second).Should(gomega.Equal(0))

	VerifySecureRouteDeletion(t, g, DefaultModelName, 0, 0)
	TearDownTestForRouteInNodePort(t, DefaultModelName)
}

func TestSecureRouteMultiNamespaceInNodePort(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	integrationtest.SetNodePortMode()
	defer integrationtest.SetClusterIPMode()
	nodeIP := "10.1.1.2"
	integrationtest.CreateNode(t, "testNodeNP", nodeIP)
	defer integrationtest.DeleteNode(t, "testNodeNP")

	SetUpTestForRouteInNodePort(t, DefaultModelName)
	route1 := FakeRoute{Path: "/foo"}.SecureRoute()
	_, err := OshiftClient.RouteV1().Routes(DefaultNamespace).Create(route1)
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	integrationtest.CreateSVC(t, "test", "avisvc", corev1.ServiceTypeNodePort, false)
	route2 := FakeRoute{Namespace: "test", Path: "/bar"}.SecureRoute()
	_, err = OshiftClient.RouteV1().Routes("test").Create(route2)
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	aviModel := ValidateSniModel(t, g, DefaultModelName)

	g.Expect(aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes).To(gomega.HaveLen(1))
	sniVS := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes[0]
	g.Eventually(func() string {
		sniVS = aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes[0]
		return sniVS.VHDomainNames[0]
	}, 20*time.Second).Should(gomega.Equal(DefaultHostname))

	g.Expect(sniVS.CACertRefs).To(gomega.HaveLen(1))
	g.Expect(sniVS.SSLKeyCertRefs).To(gomega.HaveLen(1))

	g.Eventually(func() int {
		sniVS = aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes[0]
		return len(sniVS.PoolRefs)
	}, 20*time.Second).Should(gomega.Equal(2))
	g.Expect(sniVS.HttpPolicyRefs).To(gomega.HaveLen(2))

	for _, pool := range sniVS.PoolRefs {
		if pool.Name != "cluster--default-foo.com_foo-foo-avisvc" && pool.Name != "cluster--test-foo.com_bar-foo-avisvc" {
			t.Fatalf("Unexpected poolName found: %s", pool.Name)
		}
	}
	for _, httpps := range sniVS.HttpPolicyRefs {
		if httpps.Name != "cluster--default-foo.com_foo-foo" && httpps.Name != "cluster--test-foo.com_bar-foo" {
			t.Fatalf("Unexpected http policyset found: %s", httpps.Name)
		}
	}

	err = OshiftClient.RouteV1().Routes("test").Delete(DefaultRouteName, nil)
	if err != nil {
		t.Fatalf("Couldn't DELETE the route %v", err)
	}
	VerifySecureRouteDeletion(t, g, DefaultModelName, 0, 0)
	TearDownTestForRouteInNodePort(t, DefaultModelName)
	integrationtest.DelSVC(t, "test", "avisvc")
}
