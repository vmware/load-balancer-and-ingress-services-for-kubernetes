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
	"testing"
	"time"

	avinodes "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/integrationtest"

	"github.com/onsi/gomega"
	routev1 "github.com/openshift/api/route/v1"
	corev1 "k8s.io/api/core/v1"
)

func (rt FakeRoute) SecureRoute() *routev1.Route {
	routeExample := rt.Route()
	routeExample.Spec.TLS = &routev1.TLSConfig{
		Certificate:   "cert",
		CACertificate: "cacert",
		Key:           "key",
		Termination:   routev1.TLSTerminationEdge,
	}
	return routeExample
}

func (rt FakeRoute) SecureABRoute(ratio ...int) *routev1.Route {
	var routeExample *routev1.Route
	if len(ratio) > 0 {
		routeExample = rt.ABRoute(ratio[0])
	} else {
		routeExample = rt.ABRoute()
	}
	routeExample.Spec.TLS = &routev1.TLSConfig{
		Certificate:   "cert",
		CACertificate: "cacert",
		Key:           "key",
		Termination:   routev1.TLSTerminationEdge,
	}
	return routeExample
}

func VerifySecureRouteDeletion(t *testing.T, g *gomega.WithT, modelName string, poolCount, snicount int, nsname ...string) {
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	VerifyRouteDeletion(t, g, aviModel, poolCount, nsname...)
	g.Eventually(func() int {
		_, aviModel = objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return len(nodes[0].SniNodes)
	}, 20*time.Second).Should(gomega.Equal(snicount))
}

func VerifySniNode(g *gomega.WithT, sniVS *avinodes.AviVsNode) {
	g.Expect(sniVS.CACertRefs).To(gomega.HaveLen(1))
	g.Expect(sniVS.SSLKeyCertRefs).To(gomega.HaveLen(1))
	g.Expect(sniVS.PoolGroupRefs).To(gomega.HaveLen(1))
	g.Expect(sniVS.PoolRefs).To(gomega.HaveLen(1))
	g.Expect(sniVS.HttpPolicyRefs).To(gomega.HaveLen(1))
}

func ValidateSniModel(t *testing.T, g *gomega.GomegaWithT, modelName string, redirect ...bool) interface{} {
	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 50*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)

	g.Eventually(func() int {
		return len(aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes)
	}, 50*time.Second).Should(gomega.Equal(1))
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()

	g.Expect(len(nodes)).To(gomega.Equal(1))
	g.Expect(nodes[0].Name).To(gomega.ContainSubstring("Shared-L7"))
	g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))

	g.Expect(nodes[0].SharedVS).To(gomega.Equal(true))
	redirectPol := 0
	if len(redirect) > 0 {
		if redirect[0] == true {
			redirectPol = 1
		}
	}
	g.Expect(nodes[0].HttpPolicyRefs).To(gomega.HaveLen(redirectPol))
	dsNodes := aviModel.(*avinodes.AviObjectGraph).GetAviHTTPDSNode()
	g.Expect(len(dsNodes)).To(gomega.Equal(1))

	return aviModel
}

func CheckMultiSNIMultiNS(t *testing.T, g *gomega.GomegaWithT, aviModel interface{}) {
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
	g.Expect(sniVS.PoolGroupRefs).To(gomega.HaveLen(2))

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
}

func TestSecureRoute(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetUpTestForRoute(t, DefaultModelName)
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
	TearDownTestForRoute(t, DefaultModelName)
}

func TestUpdatePathSecureRoute(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetUpTestForRoute(t, DefaultModelName)
	routeExample := FakeRoute{Path: "/foo"}.SecureRoute()
	_, err := OshiftClient.RouteV1().Routes(DefaultNamespace).Create(routeExample)
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	routeExample = FakeRoute{Path: "/bar"}.SecureRoute()
	routeExample.ObjectMeta.ResourceVersion = "2"
	_, err = OshiftClient.RouteV1().Routes(DefaultNamespace).Update(routeExample)
	if err != nil {
		t.Fatalf("error in updating route: %v", err)
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
	TearDownTestForRoute(t, DefaultModelName)
}

func TestUpdateHostnameSecureRoute(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	SetUpTestForRoute(t, DefaultModelName)
	routeExample := FakeRoute{Path: "/foo"}.SecureRoute()
	_, err := OshiftClient.RouteV1().Routes(DefaultNamespace).Create(routeExample)
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	routeExample = FakeRoute{Hostname: "bar.com", Path: "/foo"}.SecureRoute()
	routeExample.ObjectMeta.ResourceVersion = "2"
	_, err = OshiftClient.RouteV1().Routes(DefaultNamespace).Update(routeExample)
	if err != nil {
		t.Fatalf("error in updating route: %v", err)
	}

	aviModel := ValidateSniModel(t, g, "admin/cluster--Shared-L7-1")

	g.Eventually(func() string {
		sniNodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes
		if len(sniNodes) == 0 {
			return ""
		}
		sniVS := sniNodes[0]
		return sniVS.VHDomainNames[0]
	}, 20*time.Second).Should(gomega.Equal("bar.com"))
	sniVS := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes[0]
	VerifySniNode(g, sniVS)

	VerifySecureRouteDeletion(t, g, "admin/cluster--Shared-L7-1", 0, 0)
	TearDownTestForRoute(t, "admin/cluster--Shared-L7-1")
}

func TestSecureToInsecureRoute(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	SetUpTestForRoute(t, DefaultModelName)
	routeExample := FakeRoute{Path: "/foo"}.SecureRoute()
	_, err := OshiftClient.RouteV1().Routes(DefaultNamespace).Create(routeExample)
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	routeExample = FakeRoute{Path: "/foo"}.Route()
	routeExample.ObjectMeta.ResourceVersion = "2"
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
	TearDownTestForRoute(t, DefaultModelName)
}

func TestInsecureToSecureRoute(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	SetUpTestForRoute(t, DefaultModelName)
	routeExample := FakeRoute{Path: "/foo"}.Route()
	_, err := OshiftClient.RouteV1().Routes(DefaultNamespace).Create(routeExample)
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	routeExample = FakeRoute{Path: "/foo"}.SecureRoute()
	routeExample.ObjectMeta.ResourceVersion = "2"
	_, err = OshiftClient.RouteV1().Routes(DefaultNamespace).Update(routeExample)
	if err != nil {
		t.Fatalf("error in updating route: %v", err)
	}

	aviModel := ValidateSniModel(t, g, DefaultModelName)

	g.Eventually(func() string {
		sniNodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes
		if len(sniNodes) == 0 {
			return ""
		}
		sniVS := sniNodes[0]
		return sniVS.VHDomainNames[0]
	}, 20*time.Second).Should(gomega.Equal("foo.com"))
	sniVS := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes[0]
	VerifySniNode(g, sniVS)

	VerifySecureRouteDeletion(t, g, DefaultModelName, 0, 0)
	TearDownTestForRoute(t, DefaultModelName)
}

func TestSecureRouteMultiNamespace(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetUpTestForRoute(t, DefaultModelName)
	route1 := FakeRoute{Path: "/foo"}.SecureRoute()
	_, err := OshiftClient.RouteV1().Routes(DefaultNamespace).Create(route1)
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	integrationtest.CreateSVC(t, "test", "avisvc", corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEP(t, "test", "avisvc", false, false, "1.1.1")
	route2 := FakeRoute{Namespace: "test", Path: "/bar"}.SecureRoute()
	_, err = OshiftClient.RouteV1().Routes("test").Create(route2)
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	aviModel := ValidateSniModel(t, g, DefaultModelName)

	CheckMultiSNIMultiNS(t, g, aviModel)

	err = OshiftClient.RouteV1().Routes("test").Delete(DefaultRouteName, nil)
	if err != nil {
		t.Fatalf("Couldn't DELETE the route %v", err)
	}
	VerifySecureRouteDeletion(t, g, DefaultModelName, 0, 0)
	TearDownTestForRoute(t, DefaultModelName)
	integrationtest.DelSVC(t, "test", "avisvc")
	integrationtest.DelEP(t, "test", "avisvc")
}

func TestSecureRouteAlternateBackend(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetUpTestForRoute(t, DefaultModelName)
	integrationtest.CreateSVC(t, "default", "absvc2", corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEP(t, "default", "absvc2", false, false, "3.3.3")
	routeExample := FakeRoute{Path: "/foo"}.SecureABRoute()
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

	g.Expect(sniVS.CACertRefs).To(gomega.HaveLen(1))
	g.Expect(sniVS.SSLKeyCertRefs).To(gomega.HaveLen(1))
	g.Expect(sniVS.PoolGroupRefs).To(gomega.HaveLen(1))
	g.Expect(sniVS.HttpPolicyRefs).To(gomega.HaveLen(1))
	g.Expect(sniVS.PoolRefs).To(gomega.HaveLen(2))

	for _, pool := range sniVS.PoolRefs {
		if pool.Name != "cluster--default-foo.com_foo-foo-avisvc" && pool.Name != "cluster--default-foo.com_foo-foo-absvc2" {
			t.Fatalf("Unexpected poolName found: %s", pool.Name)
		}
		g.Expect(pool.Servers).To(gomega.HaveLen(1))
	}
	for _, pgmember := range sniVS.PoolGroupRefs[0].Members {
		if *pgmember.PoolRef == "/api/pool?name=cluster--default-foo.com_foo-foo-avisvc" {
			g.Expect(*pgmember.Ratio).To(gomega.Equal(int32(100)))
		} else if *pgmember.PoolRef == "/api/pool?name=cluster--default-foo.com_foo-foo-absvc2" {
			g.Expect(*pgmember.Ratio).To(gomega.Equal(int32(200)))
		} else {
			t.Fatalf("Unexpected pg member: %s", *pgmember.PoolRef)
		}
	}

	VerifySecureRouteDeletion(t, g, DefaultModelName, 0, 0)
	TearDownTestForRoute(t, DefaultModelName)
	integrationtest.DelSVC(t, "default", "absvc2")
	integrationtest.DelEP(t, "default", "absvc2")
}

func TestSecureRouteAlternateBackendUpdateRatio(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetUpTestForRoute(t, DefaultModelName)
	integrationtest.CreateSVC(t, "default", "absvc2", corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEP(t, "default", "absvc2", false, false, "3.3.3")
	routeExample := FakeRoute{Path: "/foo"}.SecureABRoute()
	_, err := OshiftClient.RouteV1().Routes(DefaultNamespace).Create(routeExample)
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	routeExample = FakeRoute{Path: "/foo"}.SecureABRoute(150)
	routeExample.ObjectMeta.ResourceVersion = "2"
	_, err = OshiftClient.RouteV1().Routes(DefaultNamespace).Update(routeExample)
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
	g.Expect(sniVS.PoolGroupRefs).To(gomega.HaveLen(1))
	g.Expect(sniVS.HttpPolicyRefs).To(gomega.HaveLen(1))
	g.Expect(sniVS.PoolRefs).To(gomega.HaveLen(2))

	for _, pool := range sniVS.PoolRefs {
		if pool.Name != "cluster--default-foo.com_foo-foo-avisvc" && pool.Name != "cluster--default-foo.com_foo-foo-absvc2" {
			t.Fatalf("Unexpected poolName found: %s", pool.Name)
		}
		g.Expect(pool.Servers).To(gomega.HaveLen(1))
	}
	for _, pgmember := range sniVS.PoolGroupRefs[0].Members {
		if *pgmember.PoolRef == "/api/pool?name=cluster--default-foo.com_foo-foo-avisvc" {
			g.Expect(*pgmember.Ratio).To(gomega.Equal(int32(100)))
		} else if *pgmember.PoolRef == "/api/pool?name=cluster--default-foo.com_foo-foo-absvc2" {
			g.Expect(*pgmember.Ratio).To(gomega.Equal(int32(150)))
		} else {
			t.Fatalf("Unexpected pg member: %s", *pgmember.PoolRef)
		}
	}

	VerifySecureRouteDeletion(t, g, DefaultModelName, 0, 0)
	TearDownTestForRoute(t, DefaultModelName)
	integrationtest.DelSVC(t, "default", "absvc2")
	integrationtest.DelEP(t, "default", "absvc2")
}

func TestSecureRouteAlternateBackendUpdatePath(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetUpTestForRoute(t, DefaultModelName)
	integrationtest.CreateSVC(t, "default", "absvc2", corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEP(t, "default", "absvc2", false, false, "3.3.3")
	routeExample := FakeRoute{Path: "/foo"}.SecureABRoute()
	_, err := OshiftClient.RouteV1().Routes(DefaultNamespace).Create(routeExample)
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	routeExample = FakeRoute{Path: "/bar"}.SecureABRoute()
	routeExample.ObjectMeta.ResourceVersion = "2"
	_, err = OshiftClient.RouteV1().Routes(DefaultNamespace).Update(routeExample)
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
	g.Expect(sniVS.PoolGroupRefs).To(gomega.HaveLen(1))
	g.Expect(sniVS.HttpPolicyRefs).To(gomega.HaveLen(1))
	g.Expect(sniVS.PoolRefs).To(gomega.HaveLen(2))

	for _, pool := range sniVS.PoolRefs {
		if pool.Name != "cluster--default-foo.com_bar-foo-avisvc" && pool.Name != "cluster--default-foo.com_bar-foo-absvc2" {
			t.Fatalf("Unexpected poolName found: %s", pool.Name)
		}
		g.Expect(pool.Servers).To(gomega.HaveLen(1))
	}
	for _, pgmember := range sniVS.PoolGroupRefs[0].Members {
		if *pgmember.PoolRef == "/api/pool?name=cluster--default-foo.com_bar-foo-avisvc" {
			g.Expect(*pgmember.Ratio).To(gomega.Equal(int32(100)))
		} else if *pgmember.PoolRef == "/api/pool?name=cluster--default-foo.com_bar-foo-absvc2" {
			g.Expect(*pgmember.Ratio).To(gomega.Equal(int32(200)))
		} else {
			t.Fatalf("Unexpected pg member: %s", *pgmember.PoolRef)
		}
	}

	VerifySecureRouteDeletion(t, g, DefaultModelName, 0, 0)
	TearDownTestForRoute(t, DefaultModelName)
	integrationtest.DelSVC(t, "default", "absvc2")
	integrationtest.DelEP(t, "default", "absvc2")
}

func TestSecureRouteRemoveAlternateBackend(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetUpTestForRoute(t, DefaultModelName)
	integrationtest.CreateSVC(t, "default", "absvc2", corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEP(t, "default", "absvc2", false, false, "3.3.3")
	routeExample := FakeRoute{Path: "/foo"}.SecureABRoute()
	_, err := OshiftClient.RouteV1().Routes(DefaultNamespace).Create(routeExample)
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	routeExample = FakeRoute{Path: "/foo"}.SecureRoute()
	routeExample.ObjectMeta.ResourceVersion = "2"
	_, err = OshiftClient.RouteV1().Routes(DefaultNamespace).Update(routeExample)
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
	g.Expect(sniVS.PoolGroupRefs).To(gomega.HaveLen(1))
	g.Expect(sniVS.HttpPolicyRefs).To(gomega.HaveLen(1))
	g.Expect(sniVS.PoolRefs).To(gomega.HaveLen(1))

	for _, pool := range sniVS.PoolRefs {
		if pool.Name != "cluster--default-foo.com_foo-foo-avisvc" {
			t.Fatalf("Unexpected poolName found: %s", pool.Name)
		}
		g.Expect(pool.Servers).To(gomega.HaveLen(1))
	}
	for _, pgmember := range sniVS.PoolGroupRefs[0].Members {
		if *pgmember.PoolRef == "/api/pool?name=cluster--default-foo.com_foo-foo-avisvc" {
			g.Expect(*pgmember.Ratio).To(gomega.Equal(int32(100)))
		} else {
			t.Fatalf("Unexpected pg member: %s", *pgmember.PoolRef)
		}
	}

	VerifySecureRouteDeletion(t, g, DefaultModelName, 0, 0)
	TearDownTestForRoute(t, DefaultModelName)
	integrationtest.DelSVC(t, "default", "absvc2")
	integrationtest.DelEP(t, "default", "absvc2")
}

func TestSecureRouteInsecureRedirect(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetUpTestForRoute(t, DefaultModelName)
	routeExample := FakeRoute{Path: "/foo"}.SecureRoute()
	routeExample.Spec.TLS.InsecureEdgeTerminationPolicy = routev1.InsecureEdgeTerminationPolicyRedirect
	_, err := OshiftClient.RouteV1().Routes(DefaultNamespace).Create(routeExample)
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	aviModel := ValidateSniModel(t, g, DefaultModelName, true)

	g.Expect(aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes).To(gomega.HaveLen(1))
	sniVS := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes[0]
	g.Eventually(func() string {
		sniVS = aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes[0]
		return sniVS.VHDomainNames[0]
	}, 20*time.Second).Should(gomega.Equal(DefaultHostname))
	VerifySniNode(g, sniVS)

	VerifySecureRouteDeletion(t, g, DefaultModelName, 0, 0)
	TearDownTestForRoute(t, DefaultModelName)
}

func TestSecureRouteInsecureAllow(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetUpTestForRoute(t, DefaultModelName)
	routeExample := FakeRoute{Path: "/foo"}.SecureRoute()
	routeExample.Spec.TLS.InsecureEdgeTerminationPolicy = routev1.InsecureEdgeTerminationPolicyAllow
	_, err := OshiftClient.RouteV1().Routes(DefaultNamespace).Create(routeExample)
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	aviModel := ValidateSniModel(t, g, DefaultModelName)

	//shared vs
	ValidateModelCommon(t, g)
	g.Eventually(func() int {
		pool := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].PoolRefs[0]
		return len(pool.Servers)
	}, 10*time.Second).Should(gomega.Equal(1))

	poolgroups := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].PoolGroupRefs
	pgmember := poolgroups[0].Members[0]
	g.Expect(*pgmember.PoolRef).To(gomega.Equal("/api/pool?name=cluster--foo.com_foo-default-foo-avisvc"))
	g.Expect(*pgmember.PriorityLabel).To(gomega.Equal("foo.com/foo"))

	// sni vs
	g.Eventually(func() int {
		return len(aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes)
	}, 10*time.Second).Should(gomega.Equal(1))
	sniVS := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes[0]
	g.Eventually(func() string {
		sniVS = aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes[0]
		return sniVS.VHDomainNames[0]
	}, 20*time.Second).Should(gomega.Equal(DefaultHostname))
	VerifySniNode(g, sniVS)

	VerifySecureRouteDeletion(t, g, DefaultModelName, 0, 0)
	TearDownTestForRoute(t, DefaultModelName)
}

//Transition insecureEdgeTerminationPolicy from Allow to Redirect
func TestSecureRouteInsecureAllowToRedirect(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetUpTestForRoute(t, DefaultModelName)
	routeExample := FakeRoute{Path: "/foo"}.SecureRoute()
	routeExample.Spec.TLS.InsecureEdgeTerminationPolicy = routev1.InsecureEdgeTerminationPolicyAllow
	_, err := OshiftClient.RouteV1().Routes(DefaultNamespace).Create(routeExample)
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	routeExample.Spec.TLS.InsecureEdgeTerminationPolicy = routev1.InsecureEdgeTerminationPolicyRedirect
	routeExample.ObjectMeta.ResourceVersion = "2"
	_, err = OshiftClient.RouteV1().Routes(DefaultNamespace).Update(routeExample)
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	aviModel := ValidateSniModel(t, g, DefaultModelName, true)

	g.Expect(aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes).To(gomega.HaveLen(1))
	sniVS := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes[0]
	g.Eventually(func() string {
		sniVS = aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes[0]
		return sniVS.VHDomainNames[0]
	}, 20*time.Second).Should(gomega.Equal(DefaultHostname))
	VerifySniNode(g, sniVS)

	VerifySecureRouteDeletion(t, g, DefaultModelName, 0, 0)
	TearDownTestForRoute(t, DefaultModelName)
}

//Transition insecureEdgeTerminationPolicy from Allow to None
func TestSecureRouteInsecureAllowToNone(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetUpTestForRoute(t, DefaultModelName)
	routeExample := FakeRoute{Path: "/foo"}.SecureRoute()
	routeExample.Spec.TLS.InsecureEdgeTerminationPolicy = routev1.InsecureEdgeTerminationPolicyAllow
	_, err := OshiftClient.RouteV1().Routes(DefaultNamespace).Create(routeExample)
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	routeExample.Spec.TLS.InsecureEdgeTerminationPolicy = routev1.InsecureEdgeTerminationPolicyNone
	routeExample.ObjectMeta.ResourceVersion = "2"
	_, err = OshiftClient.RouteV1().Routes(DefaultNamespace).Update(routeExample)
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
	TearDownTestForRoute(t, DefaultModelName)
}

//Transition insecureEdgeTerminationPolicy from Redirect to Allow
func TestSecureRouteInsecureRedirectToAllow(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetUpTestForRoute(t, DefaultModelName)
	routeExample := FakeRoute{Path: "/foo"}.SecureRoute()
	routeExample.Spec.TLS.InsecureEdgeTerminationPolicy = routev1.InsecureEdgeTerminationPolicyRedirect
	_, err := OshiftClient.RouteV1().Routes(DefaultNamespace).Create(routeExample)
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	routeExample.Spec.TLS.InsecureEdgeTerminationPolicy = routev1.InsecureEdgeTerminationPolicyAllow
	routeExample.ObjectMeta.ResourceVersion = "2"
	_, err = OshiftClient.RouteV1().Routes(DefaultNamespace).Update(routeExample)
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	aviModel := ValidateSniModel(t, g, DefaultModelName)

	//shared vs
	ValidateModelCommon(t, g)
	g.Eventually(func() int {
		vslist := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		if len(vslist) == 0 {
			return 0
		}
		poolrefs := vslist[0].PoolRefs
		if len(poolrefs) == 0 {
			return 0
		}
		pool := poolrefs[0]
		return len(pool.Servers)
	}, 10*time.Second).Should(gomega.Equal(1))

	poolgroups := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].PoolGroupRefs
	g.Expect(poolgroups).To(gomega.HaveLen(1))
	pgmember := poolgroups[0].Members[0]
	g.Expect(*pgmember.PoolRef).To(gomega.Equal("/api/pool?name=cluster--foo.com_foo-default-foo-avisvc"))
	g.Expect(*pgmember.PriorityLabel).To(gomega.Equal("foo.com/foo"))

	// sni vs
	g.Eventually(func() int {
		return len(aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes)
	}, 20*time.Second).Should(gomega.Equal(1))
	sniVS := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes[0]
	g.Eventually(func() string {
		sniVS = aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes[0]
		return sniVS.VHDomainNames[0]
	}, 20*time.Second).Should(gomega.Equal(DefaultHostname))
	VerifySniNode(g, sniVS)

	VerifySecureRouteDeletion(t, g, DefaultModelName, 0, 0)
	TearDownTestForRoute(t, DefaultModelName)
}

//Transition insecureEdgeTerminationPolicy from Redirect to None
func TestSecureRouteInsecureRedirectToNone(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetUpTestForRoute(t, DefaultModelName)
	routeExample := FakeRoute{Path: "/foo"}.SecureRoute()
	routeExample.Spec.TLS.InsecureEdgeTerminationPolicy = routev1.InsecureEdgeTerminationPolicyRedirect
	_, err := OshiftClient.RouteV1().Routes(DefaultNamespace).Create(routeExample)
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	routeExample.Spec.TLS.InsecureEdgeTerminationPolicy = routev1.InsecureEdgeTerminationPolicyNone
	routeExample.ObjectMeta.ResourceVersion = "2"
	_, err = OshiftClient.RouteV1().Routes(DefaultNamespace).Update(routeExample)
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
	TearDownTestForRoute(t, DefaultModelName)
}

func TestSecureRouteInsecureRedirectMultiNamespace(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetUpTestForRoute(t, DefaultModelName)
	route1 := FakeRoute{Path: "/foo"}.SecureRoute()
	route1.Spec.TLS.InsecureEdgeTerminationPolicy = routev1.InsecureEdgeTerminationPolicyRedirect
	_, err := OshiftClient.RouteV1().Routes(DefaultNamespace).Create(route1)
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	integrationtest.CreateSVC(t, "test", "avisvc", corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEP(t, "test", "avisvc", false, false, "1.1.1")
	route2 := FakeRoute{Namespace: "test", Path: "/bar"}.SecureRoute()
	route2.Spec.TLS.InsecureEdgeTerminationPolicy = routev1.InsecureEdgeTerminationPolicyRedirect
	_, err = OshiftClient.RouteV1().Routes("test").Create(route2)
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	aviModel := ValidateSniModel(t, g, DefaultModelName, true)

	CheckMultiSNIMultiNS(t, g, aviModel)

	err = OshiftClient.RouteV1().Routes("test").Delete(DefaultRouteName, nil)
	if err != nil {
		t.Fatalf("Couldn't DELETE the route %v", err)
	}
	VerifySecureRouteDeletion(t, g, DefaultModelName, 0, 0)
	TearDownTestForRoute(t, DefaultModelName)
	integrationtest.DelSVC(t, "test", "avisvc")
	integrationtest.DelEP(t, "test", "avisvc")
}

func TestSecureRouteInsecureAllowMultiNamespace(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetUpTestForRoute(t, DefaultModelName)
	route1 := FakeRoute{Path: "/foo"}.SecureRoute()
	route1.Spec.TLS.InsecureEdgeTerminationPolicy = routev1.InsecureEdgeTerminationPolicyAllow
	_, err := OshiftClient.RouteV1().Routes(DefaultNamespace).Create(route1)
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	integrationtest.CreateSVC(t, "test", "avisvc", corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEP(t, "test", "avisvc", false, false, "1.1.1")
	route2 := FakeRoute{Namespace: "test", Path: "/bar"}.SecureRoute()
	route2.Spec.TLS.InsecureEdgeTerminationPolicy = routev1.InsecureEdgeTerminationPolicyAllow
	_, err = OshiftClient.RouteV1().Routes("test").Create(route2)
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	aviModel := ValidateSniModel(t, g, DefaultModelName)

	// shared VS
	ValidateModelCommon(t, g)
	g.Eventually(func() int {
		pools := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].PoolRefs
		return len(pools)
	}, 10*time.Second).Should(gomega.Equal(2))

	pools := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].PoolRefs
	for _, pool := range pools {
		if pool.Name != "cluster--foo.com_foo-default-foo-avisvc" && pool.Name != "cluster--foo.com_bar-test-foo-avisvc" {
			t.Fatalf("Unexpected pool found: %s", pool.Name)
		}
	}

	poolgroups := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].PoolGroupRefs
	for _, pgmember := range poolgroups[0].Members {
		if *pgmember.PoolRef == "/api/pool?name=cluster--foo.com_foo-default-foo-avisvc" {
			g.Expect(*pgmember.PriorityLabel).To(gomega.Equal("foo.com/foo"))
		} else if *pgmember.PoolRef == "/api/pool?name=cluster--foo.com_bar-test-foo-avisvc" {
			g.Expect(*pgmember.PriorityLabel).To(gomega.Equal("foo.com/bar"))
		} else {
			t.Fatalf("Unexpected PG member found: %s", *pgmember.PoolRef)
		}
	}

	// sni VS
	CheckMultiSNIMultiNS(t, g, aviModel)

	err = OshiftClient.RouteV1().Routes("test").Delete(DefaultRouteName, nil)
	if err != nil {
		t.Fatalf("Couldn't DELETE the route %v", err)
	}
	VerifySecureRouteDeletion(t, g, DefaultModelName, 0, 0)
	TearDownTestForRoute(t, DefaultModelName)
	integrationtest.DelSVC(t, "test", "avisvc")
	integrationtest.DelEP(t, "test", "avisvc")
}

func TestReencryptRoute(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetUpTestForRoute(t, DefaultModelName)
	routeExample := FakeRoute{Path: "/foo"}.SecureRoute()
	routeExample.Spec.TLS.Termination = routev1.TLSTerminationReencrypt
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
	}, 60*time.Second).Should(gomega.Equal(DefaultHostname))
	VerifySniNode(g, sniVS)

	g.Expect(sniVS.PoolRefs[0].SniEnabled).To(gomega.Equal(true))
	g.Expect(sniVS.PoolRefs[0].SslProfileRef).To(gomega.Equal("/api/sslprofile?name=System-Standard"))

	VerifySecureRouteDeletion(t, g, DefaultModelName, 0, 0)
	TearDownTestForRoute(t, DefaultModelName)
}

func TestRemoveReencryptRoute(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetUpTestForRoute(t, DefaultModelName)
	routeExample := FakeRoute{Path: "/foo"}.SecureRoute()
	routeExample.Spec.TLS.Termination = routev1.TLSTerminationReencrypt
	_, err := OshiftClient.RouteV1().Routes(DefaultNamespace).Create(routeExample)
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	routeExample.Spec.TLS.Termination = routev1.TLSTerminationEdge
	routeExample.ObjectMeta.ResourceVersion = "2"
	_, err = OshiftClient.RouteV1().Routes(DefaultNamespace).Update(routeExample)
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	aviModel := ValidateSniModel(t, g, DefaultModelName)

	g.Expect(aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes).To(gomega.HaveLen(1))
	sniVS := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes[0]
	g.Eventually(func() string {
		sniVS = aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes[0]
		return sniVS.VHDomainNames[0]
	}, 60*time.Second).Should(gomega.Equal(DefaultHostname))
	VerifySniNode(g, sniVS)
	g.Eventually(func() bool {
		return sniVS.PoolRefs[0].SniEnabled
	}, 60*time.Second).Should(gomega.Equal(false))

	VerifySecureRouteDeletion(t, g, DefaultModelName, 0, 0)
	TearDownTestForRoute(t, DefaultModelName)
}

func TestRencryptRouteAlternateBackend(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetUpTestForRoute(t, DefaultModelName)
	integrationtest.CreateSVC(t, "default", "absvc2", corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEP(t, "default", "absvc2", false, false, "3.3.3")
	routeExample := FakeRoute{Path: "/foo"}.SecureABRoute()
	routeExample.Spec.TLS.Termination = routev1.TLSTerminationReencrypt
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
	}, 60*time.Second).Should(gomega.Equal(DefaultHostname))

	g.Expect(sniVS.CACertRefs).To(gomega.HaveLen(1))
	g.Expect(sniVS.SSLKeyCertRefs).To(gomega.HaveLen(1))
	g.Expect(sniVS.PoolGroupRefs).To(gomega.HaveLen(1))
	g.Expect(sniVS.HttpPolicyRefs).To(gomega.HaveLen(1))
	g.Expect(sniVS.PoolRefs).To(gomega.HaveLen(2))

	for _, pool := range sniVS.PoolRefs {
		if pool.Name != "cluster--default-foo.com_foo-foo-avisvc" && pool.Name != "cluster--default-foo.com_foo-foo-absvc2" {
			t.Fatalf("Unexpected poolName found: %s", pool.Name)
		} else {
			g.Eventually(func() bool {
				return pool.SniEnabled
			}, 60*time.Second).Should(gomega.Equal(true))
			g.Expect(pool.SslProfileRef).To(gomega.Equal("/api/sslprofile?name=System-Standard"))
		}
		g.Expect(pool.Servers).To(gomega.HaveLen(1))
	}
	for _, pgmember := range sniVS.PoolGroupRefs[0].Members {
		if *pgmember.PoolRef == "/api/pool?name=cluster--default-foo.com_foo-foo-avisvc" {
			g.Expect(*pgmember.Ratio).To(gomega.Equal(int32(100)))
		} else if *pgmember.PoolRef == "/api/pool?name=cluster--default-foo.com_foo-foo-absvc2" {
			g.Expect(*pgmember.Ratio).To(gomega.Equal(int32(200)))
		} else {
			t.Fatalf("Unexpected pg member: %s", *pgmember.PoolRef)
		}
	}

	VerifySecureRouteDeletion(t, g, DefaultModelName, 0, 0)
	TearDownTestForRoute(t, DefaultModelName)
	integrationtest.DelSVC(t, "default", "absvc2")
	integrationtest.DelEP(t, "default", "absvc2")
}

func TestReencryptRouteWithDestinationCA(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetUpTestForRoute(t, DefaultModelName)
	routeExample := FakeRoute{Path: "/foo"}.SecureRoute()
	routeExample.Spec.TLS.Termination = routev1.TLSTerminationReencrypt
	routeExample.Spec.TLS.DestinationCACertificate = "abc"
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
	}, 60*time.Second).Should(gomega.Equal(DefaultHostname))
	VerifySniNode(g, sniVS)
	g.Eventually(func() bool {
		return sniVS.PoolRefs[0].SniEnabled
	}, 60*time.Second).Should(gomega.Equal(true))

	g.Expect(sniVS.PoolRefs[0].SslProfileRef).To(gomega.Equal("/api/sslprofile?name=System-Standard"))
	g.Expect(sniVS.PoolRefs[0].PkiProfile.Name).To(gomega.Equal("cluster--default-foo.com_foo-foo-avisvc-pkiprofile"))

	VerifySecureRouteDeletion(t, g, DefaultModelName, 0, 0)
	TearDownTestForRoute(t, DefaultModelName)
}

func TestReencryptRouteRemoveDestinationCA(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetUpTestForRoute(t, DefaultModelName)
	routeExample := FakeRoute{Path: "/foo"}.SecureRoute()
	routeExample.Spec.TLS.Termination = routev1.TLSTerminationReencrypt
	routeExample.Spec.TLS.DestinationCACertificate = "abc"
	_, err := OshiftClient.RouteV1().Routes(DefaultNamespace).Create(routeExample)
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	routeExample.Spec.TLS.DestinationCACertificate = ""
	routeExample.ObjectMeta.ResourceVersion = "2"
	_, err = OshiftClient.RouteV1().Routes(DefaultNamespace).Update(routeExample)
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	aviModel := ValidateSniModel(t, g, DefaultModelName)

	g.Expect(aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes).To(gomega.HaveLen(1))
	sniVS := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes[0]
	g.Eventually(func() string {
		sniVS = aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes[0]
		return sniVS.VHDomainNames[0]
	}, 60*time.Second).Should(gomega.Equal(DefaultHostname))
	VerifySniNode(g, sniVS)
	g.Eventually(func() bool {
		return sniVS.PoolRefs[0].SniEnabled
	}, 60*time.Second).Should(gomega.Equal(true))
	g.Expect(sniVS.PoolRefs[0].SslProfileRef).To(gomega.Equal("/api/sslprofile?name=System-Standard"))

	var nilPki *avinodes.AviPkiProfileNode
	g.Eventually(func() *avinodes.AviPkiProfileNode {
		return sniVS.PoolRefs[0].PkiProfile
	}, 60*time.Second).Should(gomega.Equal(nilPki))

	VerifySecureRouteDeletion(t, g, DefaultModelName, 0, 0)
	TearDownTestForRoute(t, DefaultModelName)
}
