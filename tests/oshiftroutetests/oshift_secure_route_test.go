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

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	avinodes "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/integrationtest"

	"github.com/onsi/gomega"
	routev1 "github.com/openshift/api/route/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func CheckMultiSNIMultiNS(t *testing.T, g *gomega.GomegaWithT, aviModel interface{}, httpsCount, hppmapCount int) {
	g.Expect(aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes).To(gomega.HaveLen(1))
	sniVS := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes[0]
	g.Eventually(func() string {
		sniVS = aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes[0]
		return sniVS.VHDomainNames[0]
	}, 40*time.Second).Should(gomega.Equal(defaultHostname))

	g.Expect(sniVS.CACertRefs).To(gomega.HaveLen(1))
	g.Expect(sniVS.SSLKeyCertRefs).To(gomega.HaveLen(1))

	g.Eventually(func() int {
		sniVS = aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes[0]
		return len(sniVS.PoolRefs)
	}, 40*time.Second).Should(gomega.Equal(2))
	g.Expect(sniVS.HttpPolicyRefs).To(gomega.HaveLen(httpsCount))
	g.Expect(sniVS.HttpPolicyRefs[0].HppMap).To(gomega.HaveLen(hppmapCount))
	g.Expect(sniVS.PoolGroupRefs).To(gomega.HaveLen(2))

	for _, pool := range sniVS.PoolRefs {
		if pool.Name != "cluster--default-foo.com_foo-foo-avisvc" && pool.Name != "cluster--test-foo.com_bar-foo-avisvc" {
			t.Fatalf("Unexpected poolName found: %s", pool.Name)
		}
	}
	for _, httpps := range sniVS.HttpPolicyRefs {
		if httpps.Name != "cluster--default-foo.com" && httpps.Name != "cluster--test-foo.com" {
			t.Fatalf("Unexpected http policyset found: %s", httpps.Name)
		}
	}
}

func TestSecureRoute(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetUpTestForRoute(t, defaultModelName)
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
	}, 20*time.Second).Should(gomega.Equal(defaultHostname))
	VerifySniNode(g, sniVS)

	VerifySecureRouteDeletion(t, g, defaultModelName, 0, 0)
	TearDownTestForRoute(t, defaultModelName)
}

func TestUpdatePathSecureRoute(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetUpTestForRoute(t, defaultModelName)
	routeExample := FakeRoute{Path: "/foo"}.SecureRoute()
	_, err := OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	routeExample = FakeRoute{Path: "/bar"}.SecureRoute()
	routeExample.ObjectMeta.ResourceVersion = "2"
	_, err = OshiftClient.RouteV1().Routes(defaultNamespace).Update(context.TODO(), routeExample, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating route: %v", err)
	}

	aviModel := ValidateSniModel(t, g, defaultModelName)

	g.Expect(aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes).To(gomega.HaveLen(1))
	sniVS := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes[0]
	g.Eventually(func() string {
		sniVS = aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes[0]
		return sniVS.VHDomainNames[0]
	}, 20*time.Second).Should(gomega.Equal(defaultHostname))
	VerifySniNode(g, sniVS)

	VerifySecureRouteDeletion(t, g, defaultModelName, 0, 0)
	TearDownTestForRoute(t, defaultModelName)
}

func TestUpdateHostnameSecureRoute(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	SetUpTestForRoute(t, defaultModelName)
	routeExample := FakeRoute{Path: "/foo"}.SecureRoute()
	_, err := OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	routeExample = FakeRoute{Hostname: "bar.com", Path: "/foo"}.SecureRoute()
	routeExample.ObjectMeta.ResourceVersion = "2"
	_, err = OshiftClient.RouteV1().Routes(defaultNamespace).Update(context.TODO(), routeExample, metav1.UpdateOptions{})
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

	SetUpTestForRoute(t, defaultModelName)
	routeExample := FakeRoute{Path: "/foo"}.SecureRoute()
	_, err := OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	routeExample = FakeRoute{Path: "/foo"}.Route()
	routeExample.ObjectMeta.ResourceVersion = "2"
	_, err = OshiftClient.RouteV1().Routes(defaultNamespace).Update(context.TODO(), routeExample, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating route: %v", err)
	}

	aviModel := ValidateModelCommon(t, g)

	g.Eventually(func() int {
		sniNodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes
		return len(sniNodes)
	}, 20*time.Second).Should(gomega.Equal(0))

	VerifySecureRouteDeletion(t, g, defaultModelName, 0, 0)
	TearDownTestForRoute(t, defaultModelName)
}

func TestInsecureToSecureRoute(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	SetUpTestForRoute(t, defaultModelName)
	routeExample := FakeRoute{Path: "/foo"}.Route()
	_, err := OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	routeExample = FakeRoute{Path: "/foo"}.SecureRoute()
	routeExample.ObjectMeta.ResourceVersion = "2"
	_, err = OshiftClient.RouteV1().Routes(defaultNamespace).Update(context.TODO(), routeExample, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating route: %v", err)
	}

	aviModel := ValidateSniModel(t, g, defaultModelName)

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

	VerifySecureRouteDeletion(t, g, defaultModelName, 0, 0)
	TearDownTestForRoute(t, defaultModelName)
}

func TestSecureRouteMultiNamespace(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	integrationtest.DeleteNamespace("test")
	SetUpTestForRoute(t, defaultModelName)
	route1 := FakeRoute{Path: "/foo"}.SecureRoute()
	_, err := OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), route1, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}
	AddLabelToNamespace(defaultKey, defaultValue, "test", defaultModelName, t)
	defer integrationtest.DeleteNamespace("test")

	integrationtest.CreateSVC(t, "test", "avisvc", corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPorEPS(t, "test", "avisvc", false, false, "1.1.1")
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
	TearDownTestForRoute(t, defaultModelName)
	integrationtest.DelSVC(t, "test", "avisvc")
	integrationtest.DelEPorEPS(t, "test", "avisvc")
}

func TestMultiSecureRouteSameNamespace(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetUpTestForRoute(t, defaultModelName)
	route1 := FakeRoute{Path: "/foo"}.SecureRoute()
	_, err := OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), route1, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	route2 := FakeRoute{Name: "route2", Namespace: defaultNamespace, Path: "/bar"}.SecureRoute()
	_, err = OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), route2, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	aviModel := ValidateSniModel(t, g, defaultModelName)
	g.Expect(aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes).To(gomega.HaveLen(1))
	sniVS := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes[0]
	g.Eventually(func() int {
		sniVS = aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes[0]
		return len(sniVS.PoolRefs)
	}, 20*time.Second).Should(gomega.Equal(2))

	err = OshiftClient.RouteV1().Routes(defaultNamespace).Delete(context.TODO(), "route2", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the route %v", err)
	}

	mcache := cache.SharedAviObjCache()
	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--foo.com"}
	g.Eventually(func() int {
		vsCache, found := mcache.VsCacheMeta.AviCacheGet(sniVSKey)
		vsCacheObj, ok := vsCache.(*cache.AviVsCache)
		if found && ok {
			return len(vsCacheObj.PoolKeyCollection)
		}
		return 2
	}, 60*time.Second).Should(gomega.Equal(1))

	VerifySecureRouteDeletion(t, g, defaultModelName, 0, 0)
	TearDownTestForRoute(t, defaultModelName)
}

func TestSecureRouteAlternateBackend(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetUpTestForRoute(t, defaultModelName)
	integrationtest.CreateSVC(t, "default", "absvc2", corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPorEPS(t, "default", "absvc2", false, false, "3.3.3")
	routeExample := FakeRoute{Path: "/foo"}.SecureABRoute()
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
	}, 20*time.Second).Should(gomega.Equal(defaultHostname))

	g.Expect(sniVS.HttpPolicyRefs[0].HppMap[0].Host[0]).To(gomega.Equal("foo.com"))
	g.Expect(sniVS.HttpPolicyRefs[0].HppMap[0].PoolGroup).To(gomega.Equal("cluster--default-foo.com_foo-foo"))

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
			g.Expect(*pgmember.Ratio).To(gomega.Equal(uint32(100)))
		} else if *pgmember.PoolRef == "/api/pool?name=cluster--default-foo.com_foo-foo-absvc2" {
			g.Expect(*pgmember.Ratio).To(gomega.Equal(uint32(200)))
		} else {
			t.Fatalf("Unexpected pg member: %s", *pgmember.PoolRef)
		}
	}

	VerifySecureRouteDeletion(t, g, defaultModelName, 0, 0)
	TearDownTestForRoute(t, defaultModelName)
	integrationtest.DelSVC(t, "default", "absvc2")
	integrationtest.DelEPorEPS(t, "default", "absvc2")
}

func TestSecureRouteAlternateBackendUpdateRatio(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetUpTestForRoute(t, defaultModelName)
	integrationtest.CreateSVC(t, "default", "absvc2", corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPorEPS(t, "default", "absvc2", false, false, "3.3.3")
	routeExample := FakeRoute{Path: "/foo"}.SecureABRoute()
	_, err := OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	routeExample = FakeRoute{Path: "/foo"}.SecureABRoute(150)
	routeExample.ObjectMeta.ResourceVersion = "2"
	_, err = OshiftClient.RouteV1().Routes(defaultNamespace).Update(context.TODO(), routeExample, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	aviModel := ValidateSniModel(t, g, defaultModelName)

	g.Expect(aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes).To(gomega.HaveLen(1))
	sniVS := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes[0]
	g.Eventually(func() string {
		sniVS = aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes[0]
		return sniVS.VHDomainNames[0]
	}, 20*time.Second).Should(gomega.Equal(defaultHostname))

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
			g.Expect(*pgmember.Ratio).To(gomega.Equal(uint32(100)))
		} else if *pgmember.PoolRef == "/api/pool?name=cluster--default-foo.com_foo-foo-absvc2" {
			g.Expect(*pgmember.Ratio).To(gomega.Equal(uint32(150)))
		} else {
			t.Fatalf("Unexpected pg member: %s", *pgmember.PoolRef)
		}
	}

	VerifySecureRouteDeletion(t, g, defaultModelName, 0, 0)
	TearDownTestForRoute(t, defaultModelName)
	integrationtest.DelSVC(t, "default", "absvc2")
	integrationtest.DelEPorEPS(t, "default", "absvc2")
}

func TestSecureRouteAlternateBackendUpdatePath(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetUpTestForRoute(t, defaultModelName)
	integrationtest.CreateSVC(t, "default", "absvc2", corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPorEPS(t, "default", "absvc2", false, false, "3.3.3")
	routeExample := FakeRoute{Path: "/foo"}.SecureABRoute()
	_, err := OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	routeExample = FakeRoute{Path: "/bar"}.SecureABRoute()
	routeExample.ObjectMeta.ResourceVersion = "2"
	_, err = OshiftClient.RouteV1().Routes(defaultNamespace).Update(context.TODO(), routeExample, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	aviModel := ValidateSniModel(t, g, defaultModelName)

	g.Expect(aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes).To(gomega.HaveLen(1))
	sniVS := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes[0]
	g.Eventually(func() string {
		sniVS = aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes[0]
		return sniVS.VHDomainNames[0]
	}, 20*time.Second).Should(gomega.Equal(defaultHostname))

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
			g.Expect(*pgmember.Ratio).To(gomega.Equal(uint32(100)))
		} else if *pgmember.PoolRef == "/api/pool?name=cluster--default-foo.com_bar-foo-absvc2" {
			g.Expect(*pgmember.Ratio).To(gomega.Equal(uint32(200)))
		} else {
			t.Fatalf("Unexpected pg member: %s", *pgmember.PoolRef)
		}
	}

	VerifySecureRouteDeletion(t, g, defaultModelName, 0, 0)
	TearDownTestForRoute(t, defaultModelName)
	integrationtest.DelSVC(t, "default", "absvc2")
	integrationtest.DelEPorEPS(t, "default", "absvc2")
}

func TestSecureRouteRemoveAlternateBackend(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetUpTestForRoute(t, defaultModelName)
	integrationtest.CreateSVC(t, "default", "absvc2", corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPorEPS(t, "default", "absvc2", false, false, "3.3.3")
	routeExample := FakeRoute{Path: "/foo"}.SecureABRoute()
	_, err := OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	routeExample = FakeRoute{Path: "/foo"}.SecureRoute()
	routeExample.ObjectMeta.ResourceVersion = "2"
	_, err = OshiftClient.RouteV1().Routes(defaultNamespace).Update(context.TODO(), routeExample, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	aviModel := ValidateSniModel(t, g, defaultModelName)

	g.Expect(aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes).To(gomega.HaveLen(1))
	sniVS := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes[0]
	g.Eventually(func() string {
		sniVS = aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes[0]
		return sniVS.VHDomainNames[0]
	}, 20*time.Second).Should(gomega.Equal(defaultHostname))

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
			g.Expect(*pgmember.Ratio).To(gomega.Equal(uint32(100)))
		} else {
			t.Fatalf("Unexpected pg member: %s", *pgmember.PoolRef)
		}
	}

	VerifySecureRouteDeletion(t, g, defaultModelName, 0, 0)
	TearDownTestForRoute(t, defaultModelName)
	integrationtest.DelSVC(t, "default", "absvc2")
	integrationtest.DelEPorEPS(t, "default", "absvc2")
}

func TestSecureRouteInsecureRedirect(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetUpTestForRoute(t, defaultModelName)
	routeExample := FakeRoute{Path: "/foo"}.SecureRoute()
	routeExample.Spec.TLS.InsecureEdgeTerminationPolicy = routev1.InsecureEdgeTerminationPolicyRedirect
	_, err := OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	aviModel := ValidateSniModel(t, g, defaultModelName, true)

	g.Expect(aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes).To(gomega.HaveLen(1))
	sniVS := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes[0]
	g.Eventually(func() string {
		sniVS = aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes[0]
		return sniVS.VHDomainNames[0]
	}, 20*time.Second).Should(gomega.Equal(defaultHostname))
	VerifySniNode(g, sniVS)

	VerifySecureRouteDeletion(t, g, defaultModelName, 0, 0)
	TearDownTestForRoute(t, defaultModelName)
}

func TestSecureRouteInsecureAllow(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetUpTestForRoute(t, defaultModelName)
	routeExample := FakeRoute{Path: "/foo"}.SecureRoute()
	routeExample.Spec.TLS.InsecureEdgeTerminationPolicy = routev1.InsecureEdgeTerminationPolicyAllow
	_, err := OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	aviModel := ValidateSniModel(t, g, defaultModelName)

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
	}, 20*time.Second).Should(gomega.Equal(defaultHostname))
	VerifySniNode(g, sniVS)

	VerifySecureRouteDeletion(t, g, defaultModelName, 0, 0)
	TearDownTestForRoute(t, defaultModelName)
}

// Transition insecureEdgeTerminationPolicy from Allow to Redirect
func TestSecureRouteInsecureAllowToRedirect(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetUpTestForRoute(t, defaultModelName)
	routeExample := FakeRoute{Path: "/foo"}.SecureRoute()
	routeExample.Spec.TLS.InsecureEdgeTerminationPolicy = routev1.InsecureEdgeTerminationPolicyAllow
	_, err := OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	routeExample.Spec.TLS.InsecureEdgeTerminationPolicy = routev1.InsecureEdgeTerminationPolicyRedirect
	routeExample.ObjectMeta.ResourceVersion = "2"
	_, err = OshiftClient.RouteV1().Routes(defaultNamespace).Update(context.TODO(), routeExample, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	aviModel := ValidateSniModel(t, g, defaultModelName, true)

	g.Expect(aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes).To(gomega.HaveLen(1))
	sniVS := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes[0]
	g.Eventually(func() string {
		sniVS = aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes[0]
		return sniVS.VHDomainNames[0]
	}, 20*time.Second).Should(gomega.Equal(defaultHostname))
	VerifySniNode(g, sniVS)

	VerifySecureRouteDeletion(t, g, defaultModelName, 0, 0)
	TearDownTestForRoute(t, defaultModelName)
}

// Transition insecureEdgeTerminationPolicy from Allow to None
func TestSecureRouteInsecureAllowToNone(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetUpTestForRoute(t, defaultModelName)
	routeExample := FakeRoute{Path: "/foo"}.SecureRoute()
	routeExample.Spec.TLS.InsecureEdgeTerminationPolicy = routev1.InsecureEdgeTerminationPolicyAllow
	_, err := OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	routeExample.Spec.TLS.InsecureEdgeTerminationPolicy = routev1.InsecureEdgeTerminationPolicyNone
	routeExample.ObjectMeta.ResourceVersion = "2"
	_, err = OshiftClient.RouteV1().Routes(defaultNamespace).Update(context.TODO(), routeExample, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	aviModel := ValidateSniModel(t, g, defaultModelName)

	g.Expect(aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes).To(gomega.HaveLen(1))
	sniVS := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes[0]
	g.Eventually(func() string {
		sniVS = aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes[0]
		return sniVS.VHDomainNames[0]
	}, 20*time.Second).Should(gomega.Equal(defaultHostname))
	VerifySniNode(g, sniVS)

	VerifySecureRouteDeletion(t, g, defaultModelName, 0, 0)
	TearDownTestForRoute(t, defaultModelName)
}

// Transition insecureEdgeTerminationPolicy from Redirect to Allow
func TestSecureRouteInsecureRedirectToAllow(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetUpTestForRoute(t, defaultModelName)
	routeExample := FakeRoute{Path: "/foo"}.SecureRoute()
	routeExample.Spec.TLS.InsecureEdgeTerminationPolicy = routev1.InsecureEdgeTerminationPolicyRedirect
	_, err := OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	routeExample.Spec.TLS.InsecureEdgeTerminationPolicy = routev1.InsecureEdgeTerminationPolicyAllow
	routeExample.ObjectMeta.ResourceVersion = "2"
	_, err = OshiftClient.RouteV1().Routes(defaultNamespace).Update(context.TODO(), routeExample, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	aviModel := ValidateSniModel(t, g, defaultModelName)

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
	}, 20*time.Second).Should(gomega.Equal(defaultHostname))
	VerifySniNode(g, sniVS)

	VerifySecureRouteDeletion(t, g, defaultModelName, 0, 0)
	TearDownTestForRoute(t, defaultModelName)
}

// Transition insecureEdgeTerminationPolicy from Redirect to None
func TestSecureRouteInsecureRedirectToNone(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetUpTestForRoute(t, defaultModelName)
	routeExample := FakeRoute{Path: "/foo"}.SecureRoute()
	routeExample.Spec.TLS.InsecureEdgeTerminationPolicy = routev1.InsecureEdgeTerminationPolicyRedirect
	_, err := OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	routeExample.Spec.TLS.InsecureEdgeTerminationPolicy = routev1.InsecureEdgeTerminationPolicyNone
	routeExample.ObjectMeta.ResourceVersion = "2"
	_, err = OshiftClient.RouteV1().Routes(defaultNamespace).Update(context.TODO(), routeExample, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	aviModel := ValidateSniModel(t, g, defaultModelName)

	g.Expect(aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes).To(gomega.HaveLen(1))
	sniVS := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes[0]
	g.Eventually(func() string {
		sniVS = aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes[0]
		return sniVS.VHDomainNames[0]
	}, 20*time.Second).Should(gomega.Equal(defaultHostname))
	VerifySniNode(g, sniVS)

	VerifySecureRouteDeletion(t, g, defaultModelName, 0, 0)
	TearDownTestForRoute(t, defaultModelName)
}

func TestSecureRouteInsecureRedirectMultiNamespace(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	integrationtest.DeleteNamespace("test")
	SetUpTestForRoute(t, defaultModelName)
	route1 := FakeRoute{Path: "/foo"}.SecureRoute()
	route1.Spec.TLS.InsecureEdgeTerminationPolicy = routev1.InsecureEdgeTerminationPolicyRedirect
	_, err := OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), route1, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}
	AddLabelToNamespace(defaultKey, defaultValue, "test", defaultModelName, t)
	defer integrationtest.DeleteNamespace("test")

	integrationtest.CreateSVC(t, "test", "avisvc", corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPorEPS(t, "test", "avisvc", false, false, "1.1.1")
	route2 := FakeRoute{Namespace: "test", Path: "/bar"}.SecureRoute()
	route2.Spec.TLS.InsecureEdgeTerminationPolicy = routev1.InsecureEdgeTerminationPolicyRedirect
	_, err = OshiftClient.RouteV1().Routes("test").Create(context.TODO(), route2, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	aviModel := ValidateSniModel(t, g, defaultModelName, true)

	CheckMultiSNIMultiNS(t, g, aviModel, 2, 1)

	err = OshiftClient.RouteV1().Routes("test").Delete(context.TODO(), defaultRouteName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the route %v", err)
	}
	VerifySecureRouteDeletion(t, g, defaultModelName, 0, 0)
	TearDownTestForRoute(t, defaultModelName)
	integrationtest.DelSVC(t, "test", "avisvc")
	integrationtest.DelEPorEPS(t, "test", "avisvc")
}

func TestSecureRouteInsecureAllowMultiNamespace(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	integrationtest.DeleteNamespace("test")

	SetUpTestForRoute(t, defaultModelName)
	route1 := FakeRoute{Path: "/foo"}.SecureRoute()
	route1.Spec.TLS.InsecureEdgeTerminationPolicy = routev1.InsecureEdgeTerminationPolicyAllow
	_, err := OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), route1, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}
	AddLabelToNamespace(defaultKey, defaultValue, "test", defaultModelName, t)
	defer integrationtest.DeleteNamespace("test")

	integrationtest.CreateSVC(t, "test", "avisvc", corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPorEPS(t, "test", "avisvc", false, false, "1.1.1")
	route2 := FakeRoute{Namespace: "test", Path: "/bar"}.SecureRoute()
	route2.Spec.TLS.InsecureEdgeTerminationPolicy = routev1.InsecureEdgeTerminationPolicyAllow
	_, err = OshiftClient.RouteV1().Routes("test").Create(context.TODO(), route2, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	aviModel := ValidateSniModel(t, g, defaultModelName)

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
	CheckMultiSNIMultiNS(t, g, aviModel, 2, 1)

	err = OshiftClient.RouteV1().Routes("test").Delete(context.TODO(), defaultRouteName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the route %v", err)
	}
	VerifySecureRouteDeletion(t, g, defaultModelName, 0, 0)
	TearDownTestForRoute(t, defaultModelName)
	integrationtest.DelSVC(t, "test", "avisvc")
	integrationtest.DelEPorEPS(t, "test", "avisvc")
}

func TestReencryptRoute(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetUpTestForRoute(t, defaultModelName)
	routeExample := FakeRoute{Path: "/foo"}.SecureRoute()
	routeExample.Spec.TLS.Termination = routev1.TLSTerminationReencrypt
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
	}, 60*time.Second).Should(gomega.Equal(defaultHostname))
	VerifySniNode(g, sniVS)

	g.Expect(sniVS.PoolRefs[0].SniEnabled).To(gomega.Equal(true))
	g.Expect(*sniVS.PoolRefs[0].SslProfileRef).To(gomega.Equal("/api/sslprofile?name=System-Standard"))

	VerifySecureRouteDeletion(t, g, defaultModelName, 0, 0)
	TearDownTestForRoute(t, defaultModelName)
}

func TestRemoveReencryptRoute(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetUpTestForRoute(t, defaultModelName)
	routeExample := FakeRoute{Path: "/foo"}.SecureRoute()
	routeExample.Spec.TLS.Termination = routev1.TLSTerminationReencrypt
	_, err := OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	routeExample.Spec.TLS.Termination = routev1.TLSTerminationEdge
	routeExample.ObjectMeta.ResourceVersion = "2"
	_, err = OshiftClient.RouteV1().Routes(defaultNamespace).Update(context.TODO(), routeExample, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	aviModel := ValidateSniModel(t, g, defaultModelName)

	g.Expect(aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes).To(gomega.HaveLen(1))
	sniVS := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes[0]
	g.Eventually(func() string {
		sniVS = aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes[0]
		return sniVS.VHDomainNames[0]
	}, 60*time.Second).Should(gomega.Equal(defaultHostname))
	VerifySniNode(g, sniVS)
	g.Eventually(func() bool {
		return sniVS.PoolRefs[0].SniEnabled
	}, 60*time.Second).Should(gomega.Equal(false))

	VerifySecureRouteDeletion(t, g, defaultModelName, 0, 0)
	TearDownTestForRoute(t, defaultModelName)
}

func TestRencryptRouteAlternateBackend(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetUpTestForRoute(t, defaultModelName)
	integrationtest.CreateSVC(t, "default", "absvc2", corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPorEPS(t, "default", "absvc2", false, false, "3.3.3")
	routeExample := FakeRoute{Path: "/foo"}.SecureABRoute()
	routeExample.Spec.TLS.Termination = routev1.TLSTerminationReencrypt
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
	}, 60*time.Second).Should(gomega.Equal(defaultHostname))

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
			g.Expect(*pool.SslProfileRef).To(gomega.Equal("/api/sslprofile?name=System-Standard"))
		}
		g.Expect(pool.Servers).To(gomega.HaveLen(1))
	}
	for _, pgmember := range sniVS.PoolGroupRefs[0].Members {
		if *pgmember.PoolRef == "/api/pool?name=cluster--default-foo.com_foo-foo-avisvc" {
			g.Expect(*pgmember.Ratio).To(gomega.Equal(uint32(100)))
		} else if *pgmember.PoolRef == "/api/pool?name=cluster--default-foo.com_foo-foo-absvc2" {
			g.Expect(*pgmember.Ratio).To(gomega.Equal(uint32(200)))
		} else {
			t.Fatalf("Unexpected pg member: %s", *pgmember.PoolRef)
		}
	}

	VerifySecureRouteDeletion(t, g, defaultModelName, 0, 0)
	TearDownTestForRoute(t, defaultModelName)
	integrationtest.DelSVC(t, "default", "absvc2")
	integrationtest.DelEPorEPS(t, "default", "absvc2")
}

func TestSecureOshiftNamingConvention(t *testing.T) {
	// checks naming convention of all generated nodes
	g := gomega.NewGomegaWithT(t)
	SetUpTestForRoute(t, defaultModelName)
	routeExample := FakeRoute{Path: "/foo/bar"}.SecureRoute()
	routeExample.Spec.TLS.Termination = routev1.TLSTerminationReencrypt
	routeExample.Spec.TLS.DestinationCACertificate = "abc"
	_, err := OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	aviModel := ValidateSniModel(t, g, defaultModelName)
	vsNode := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0]

	g.Expect(vsNode.Name).To(gomega.Equal("cluster--Shared-L7-0"))
	g.Expect(vsNode.PoolGroupRefs[0].Name).To(gomega.Equal("cluster--Shared-L7-0"))
	g.Expect(vsNode.HTTPDSrefs[0].Name).To(gomega.Equal("cluster--Shared-L7-0"))
	g.Expect(vsNode.SniNodes[0].Name).To(gomega.Equal("cluster--foo.com"))
	g.Expect(vsNode.SniNodes[0].PoolGroupRefs[0].Name).To(gomega.Equal("cluster--default-foo.com_foo_bar-foo"))
	g.Expect(vsNode.SniNodes[0].PoolRefs[0].Name).To(gomega.Equal("cluster--default-foo.com_foo_bar-foo-avisvc"))
	g.Expect(vsNode.SniNodes[0].PoolRefs[0].PkiProfile.Name).To(gomega.Equal("cluster--default-foo.com_foo_bar-foo-avisvc-pkiprofile"))
	g.Expect(vsNode.SniNodes[0].CACertRefs[0].Name).To(gomega.Equal("cluster--foo.com-cacert"))
	g.Expect(vsNode.SniNodes[0].SSLKeyCertRefs[0].Name).To(gomega.Equal("cluster--foo.com"))
	g.Expect(vsNode.SniNodes[0].HttpPolicyRefs[0].Name).To(gomega.Equal("cluster--default-foo.com"))
	g.Expect(vsNode.SniNodes[0].HttpPolicyRefs[0].HppMap[0].Name).To(gomega.Equal("cluster--default-foo.com_foo_bar-foo"))
	g.Expect(vsNode.VSVIPRefs[0].Name).To(gomega.Equal("cluster--Shared-L7-0"))

	VerifySecureRouteDeletion(t, g, defaultModelName, 0, 0)
	TearDownTestForRoute(t, defaultModelName)
}

func TestReencryptRouteWithDestinationCA(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetUpTestForRoute(t, defaultModelName)
	routeExample := FakeRoute{Path: "/foo"}.SecureRoute()
	routeExample.Spec.TLS.Termination = routev1.TLSTerminationReencrypt
	routeExample.Spec.TLS.DestinationCACertificate = "abc"
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
	}, 60*time.Second).Should(gomega.Equal(defaultHostname))
	VerifySniNode(g, sniVS)
	g.Eventually(func() bool {
		return sniVS.PoolRefs[0].SniEnabled
	}, 60*time.Second).Should(gomega.Equal(true))

	g.Expect(*sniVS.PoolRefs[0].SslProfileRef).To(gomega.Equal("/api/sslprofile?name=System-Standard"))
	g.Expect(sniVS.PoolRefs[0].PkiProfile.Name).To(gomega.Equal("cluster--default-foo.com_foo-foo-avisvc-pkiprofile"))

	VerifySecureRouteDeletion(t, g, defaultModelName, 0, 0)
	TearDownTestForRoute(t, defaultModelName)
}

func TestReencryptRouteRemoveDestinationCA(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetUpTestForRoute(t, defaultModelName)
	routeExample := FakeRoute{Path: "/foo"}.SecureRoute()
	routeExample.Spec.TLS.Termination = routev1.TLSTerminationReencrypt
	routeExample.Spec.TLS.DestinationCACertificate = "abc"
	_, err := OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	routeExample.Spec.TLS.DestinationCACertificate = ""
	routeExample.ObjectMeta.ResourceVersion = "2"
	_, err = OshiftClient.RouteV1().Routes(defaultNamespace).Update(context.TODO(), routeExample, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	aviModel := ValidateSniModel(t, g, defaultModelName)

	g.Expect(aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes).To(gomega.HaveLen(1))
	sniVS := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes[0]
	g.Eventually(func() string {
		sniVS = aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes[0]
		return sniVS.VHDomainNames[0]
	}, 60*time.Second).Should(gomega.Equal(defaultHostname))
	VerifySniNode(g, sniVS)
	g.Eventually(func() bool {
		return sniVS.PoolRefs[0].SniEnabled
	}, 60*time.Second).Should(gomega.Equal(true))
	g.Expect(*sniVS.PoolRefs[0].SslProfileRef).To(gomega.Equal("/api/sslprofile?name=System-Standard"))

	var nilPki *avinodes.AviPkiProfileNode
	g.Eventually(func() *avinodes.AviPkiProfileNode {
		return sniVS.PoolRefs[0].PkiProfile
	}, 60*time.Second).Should(gomega.Equal(nilPki))

	VerifySecureRouteDeletion(t, g, defaultModelName, 0, 0)
	TearDownTestForRoute(t, defaultModelName)
}

func TestAddPathSecureRouteNoKeyCert(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetUpTestForRoute(t, defaultModelName)
	routeExample := FakeRoute{Path: "/foo"}.SecureRouteNoCertKey()
	_, err := OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}
	integrationtest.AddSecret(integrationtest.DefaultRouteCert, utils.GetAKONamespace(), "tlsCert", "tlsKey")

	aviModel := ValidateSniModel(t, g, defaultModelName)

	g.Expect(aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes).To(gomega.HaveLen(1))
	sniVS := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes[0]
	g.Eventually(func() string {
		sniVS = aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes[0]
		return sniVS.VHDomainNames[0]
	}, 20*time.Second).Should(gomega.Equal(defaultHostname))
	VerifySniNodeNoCA(g, sniVS)

	KubeClient.CoreV1().Secrets(utils.GetAKONamespace()).Delete(context.TODO(), integrationtest.DefaultRouteCert, metav1.DeleteOptions{})
	VerifySecureRouteDeletion(t, g, defaultModelName, 0, 0)
	TearDownTestForRoute(t, defaultModelName)
}

func TestUpdatePathSecureRouteNoKeyCert(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetUpTestForRoute(t, defaultModelName)
	routeExample := FakeRoute{Path: "/foo"}.SecureRouteNoCertKey()
	_, err := OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}
	integrationtest.AddSecret(integrationtest.DefaultRouteCert, utils.GetAKONamespace(), "tlsCert", "tlsKey")

	routeExample = FakeRoute{Path: "/bar"}.SecureRoute()
	routeExample.ObjectMeta.ResourceVersion = "2"
	_, err = OshiftClient.RouteV1().Routes(defaultNamespace).Update(context.TODO(), routeExample, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating route: %v", err)
	}

	aviModel := ValidateSniModel(t, g, defaultModelName)

	g.Expect(aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes).To(gomega.HaveLen(1))
	sniVS := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes[0]
	g.Eventually(func() string {
		sniVS = aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes[0]
		return sniVS.VHDomainNames[0]
	}, 20*time.Second).Should(gomega.Equal(defaultHostname))
	VerifySniNodeNoCA(g, sniVS)

	KubeClient.CoreV1().Secrets(utils.GetAKONamespace()).Delete(context.TODO(), integrationtest.DefaultRouteCert, metav1.DeleteOptions{})
	VerifySecureRouteDeletion(t, g, defaultModelName, 0, 0)
	TearDownTestForRoute(t, defaultModelName)
}

func TestUpdateSecureRouteToNoKeyCert(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetUpTestForRoute(t, defaultModelName)
	integrationtest.AddSecret(integrationtest.DefaultRouteCert, utils.GetAKONamespace(), "tlsCert", "tlsKey")
	routeExample := FakeRoute{Path: "/foo"}.SecureRoute()
	_, err := OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}
	routeExample = FakeRoute{Path: "/foo"}.SecureRouteNoCertKey()
	_, err = OshiftClient.RouteV1().Routes(defaultNamespace).Update(context.TODO(), routeExample, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	routeExample = FakeRoute{Path: "/bar"}.SecureRoute()
	routeExample.ObjectMeta.ResourceVersion = "2"
	_, err = OshiftClient.RouteV1().Routes(defaultNamespace).Update(context.TODO(), routeExample, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating route: %v", err)
	}

	aviModel := ValidateSniModel(t, g, defaultModelName)

	g.Expect(aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes).To(gomega.HaveLen(1))
	sniVS := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes[0]
	g.Eventually(func() string {
		sniVS = aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes[0]
		return sniVS.VHDomainNames[0]
	}, 20*time.Second).Should(gomega.Equal(defaultHostname))
	VerifySniNodeNoCA(g, sniVS)

	KubeClient.CoreV1().Secrets(utils.GetAKONamespace()).Delete(context.TODO(), integrationtest.DefaultRouteCert, metav1.DeleteOptions{})
	VerifySecureRouteDeletion(t, g, defaultModelName, 0, 0)
	TearDownTestForRoute(t, defaultModelName)
}

func TestUpdateSecureRouteNoKeyCertToKeyCert(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetUpTestForRoute(t, defaultModelName)
	integrationtest.AddSecret(integrationtest.DefaultRouteCert, utils.GetAKONamespace(), "tlsCert", "tlsKey")
	routeExample := FakeRoute{Path: "/foo"}.SecureRouteNoCertKey()
	_, err := OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}
	routeExample = FakeRoute{Path: "/foo"}.SecureRoute()
	_, err = OshiftClient.RouteV1().Routes(defaultNamespace).Update(context.TODO(), routeExample, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	routeExample = FakeRoute{Path: "/bar"}.SecureRoute()
	routeExample.ObjectMeta.ResourceVersion = "2"
	_, err = OshiftClient.RouteV1().Routes(defaultNamespace).Update(context.TODO(), routeExample, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating route: %v", err)
	}

	aviModel := ValidateSniModel(t, g, defaultModelName)

	g.Expect(aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes).To(gomega.HaveLen(1))
	sniVS := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes[0]
	g.Eventually(func() string {
		sniVS = aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes[0]
		return sniVS.VHDomainNames[0]
	}, 20*time.Second).Should(gomega.Equal(defaultHostname))
	VerifySniNodeNoCA(g, sniVS)

	KubeClient.CoreV1().Secrets(utils.GetAKONamespace()).Delete(context.TODO(), integrationtest.DefaultRouteCert, metav1.DeleteOptions{})
	VerifySecureRouteDeletion(t, g, defaultModelName, 0, 0)
	TearDownTestForRoute(t, defaultModelName)
}

func TestSecureRouteMultiNamespaceWithStrictRestrictFqdn(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	lib.AKOControlConfig().SetAKOFQDNReusePolicy("strict")
	integrationtest.DeleteNamespace("test")
	SetUpTestForRoute(t, defaultModelName)
	route1 := FakeRoute{Path: "/foo"}.SecureRoute()
	_, err := OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), route1, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}
	defer integrationtest.DeleteNamespace("test")

	integrationtest.CreateSVC(t, "test", "avisvc", corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPorEPS(t, "test", "avisvc", false, false, "1.1.1")
	route2 := FakeRoute{Namespace: "test", Path: "/bar"}.SecureRoute()
	_, err = OshiftClient.RouteV1().Routes("test").Create(context.TODO(), route2, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	aviModel := ValidateSniModel(t, g, defaultModelName)

	g.Expect(aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes).To(gomega.HaveLen(1))
	sniVS := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes[0]
	g.Eventually(func() string {
		sniVS = aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes[0]
		return sniVS.VHDomainNames[0]
	}, 40*time.Second).Should(gomega.Equal(defaultHostname))

	g.Expect(sniVS.CACertRefs).To(gomega.HaveLen(1))
	g.Expect(sniVS.SSLKeyCertRefs).To(gomega.HaveLen(1))

	g.Eventually(func() int {
		sniVS = aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes[0]
		return len(sniVS.PoolRefs)
	}, 40*time.Second).Should(gomega.Equal(1))
	g.Expect(sniVS.HttpPolicyRefs).To(gomega.HaveLen(1))
	g.Expect(sniVS.HttpPolicyRefs[0].HppMap).To(gomega.HaveLen(1))
	g.Expect(sniVS.PoolGroupRefs).To(gomega.HaveLen(1))

	for _, pool := range sniVS.PoolRefs {
		if pool.Name != "cluster--default-foo.com_foo-foo-avisvc" {
			t.Fatalf("Unexpected poolName found: %s", pool.Name)
		}
	}
	for _, httpps := range sniVS.HttpPolicyRefs {
		if httpps.Name != "cluster--default-foo.com" {
			t.Fatalf("Unexpected http policyset found: %s", httpps.Name)
		}
	}

	err = OshiftClient.RouteV1().Routes("test").Delete(context.TODO(), defaultRouteName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the route %v", err)
	}
	lib.AKOControlConfig().SetAKOFQDNReusePolicy("internamespaceallowed")
	VerifySecureRouteDeletion(t, g, defaultModelName, 0, 0)
	TearDownTestForRoute(t, defaultModelName)
	integrationtest.DelSVC(t, "test", "avisvc")
	integrationtest.DelEPorEPS(t, "test", "avisvc")
}
