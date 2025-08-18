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
	"testing"
	"time"

	avinodes "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/integrationtest"

	"github.com/onsi/gomega"
	routev1 "github.com/openshift/api/route/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var DefaultPassthroughModel = "admin/cluster--Shared-Passthrough-0"

func (rt FakeRoute) PassthroughRoute() *routev1.Route {
	routeExample := rt.Route()
	routeExample.Spec.TLS = &routev1.TLSConfig{
		Termination: routev1.TLSTerminationPassthrough,
	}
	return routeExample
}

func (rt FakeRoute) PassthroughABRoute(ratio ...int) *routev1.Route {
	routeExample := rt.PassthroughRoute()
	weight2 := int32(200)
	if len(ratio) > 0 {
		weight2 = int32(ratio[0])
	}
	backend2 := routev1.RouteTargetReference{
		Kind:   "Service",
		Name:   "absvc2",
		Weight: &weight2,
	}
	routeExample.Spec.AlternateBackends = append(routeExample.Spec.AlternateBackends, backend2)
	return routeExample
}

func ValidatePassthroughModel(t *testing.T, g *gomega.WithT, modelName string) interface{} {

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 60*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()

	g.Expect(len(nodes)).To(gomega.Equal(1))
	g.Expect(nodes[0].HTTPDSrefs).To(gomega.HaveLen(1))
	g.Expect(nodes[0].Tenant).To(gomega.Equal("admin"))

	return aviModel
}

func VerifyPassthroughRouteDeletion(t *testing.T, g *gomega.WithT, modelName string, poolCount, childcount int) {
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)

	err := OshiftClient.RouteV1().Routes(defaultNamespace).Delete(context.TODO(), defaultRouteName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the route %v", err)
	}
	var nodes []*avinodes.AviVsNode
	g.Eventually(func() []*avinodes.AviPoolNode {
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return nodes[0].PoolRefs
	}, 60*time.Second).Should(gomega.HaveLen(poolCount))

	g.Eventually(func() []*avinodes.AviPoolGroupNode {
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return nodes[0].PoolGroupRefs
	}, 10*time.Second).Should(gomega.HaveLen(poolCount))

	g.Eventually(func() int {
		_, aviModel = objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		if len(nodes[0].PassthroughChildNodes) == 0 {
			return 0
		}
		return len(nodes[0].PassthroughChildNodes[0].HttpPolicySetRefs)
	}, 60*time.Second).Should(gomega.Equal(childcount))
}

func VerifyOnePasthrough(t *testing.T, g *gomega.WithT, vs *avinodes.AviVsNode) {

	g.Eventually(func() int {
		if len(vs.HTTPDSrefs) < 1 {
			return 0
		}
		return len(vs.HTTPDSrefs[0].PoolGroupRefs)
	}, 60*time.Second).Should(gomega.Equal(1))

	g.Expect(vs.HTTPDSrefs[0].PoolGroupRefs[0]).To(gomega.Equal("cluster--foo.com"))

	g.Expect(vs.PoolGroupRefs).To(gomega.HaveLen(1))
	g.Expect(vs.PoolGroupRefs[0].Name).To(gomega.Equal("cluster--foo.com"))

	g.Eventually(func() int {
		return len(vs.PoolGroupRefs[0].Members)
	}, 60*time.Second).Should(gomega.Equal(1))
	g.Expect(*vs.PoolGroupRefs[0].Members[0].PoolRef).To(gomega.Equal("/api/pool?name=cluster--foo.com-avisvc"))

	g.Expect(vs.PoolRefs).To(gomega.HaveLen(1))
	g.Expect(vs.PoolRefs[0].Name).To(gomega.Equal("cluster--foo.com-avisvc"))

	g.Eventually(func() int {
		return len(vs.PoolRefs[0].Servers)
	}, 60*time.Second).Should(gomega.Equal(1))

	g.Expect(vs.VSVIPRefs[0].FQDNs[0]).To(gomega.Equal("foo.com"))
}

func TestPassthroughRoute(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetUpTestForRoute(t, DefaultPassthroughModel)
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
	TearDownTestForRoute(t, DefaultPassthroughModel)
}

func TestPassthroughRedirectRoute(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetUpTestForRoute(t, DefaultPassthroughModel)
	routeExample := FakeRoute{Path: "/foo"}.PassthroughRoute()
	routeExample.Spec.TLS.InsecureEdgeTerminationPolicy = routev1.InsecureEdgeTerminationPolicyRedirect
	_, err := OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	aviModel := ValidatePassthroughModel(t, g, DefaultPassthroughModel)
	graph := aviModel.(*avinodes.AviObjectGraph)
	vs := graph.GetAviVS()[0]

	VerifyOnePasthrough(t, g, vs)

	g.Expect(vs.PassthroughChildNodes).To(gomega.HaveLen(1))
	passInsecureNode := vs.PassthroughChildNodes[0]
	g.Expect(passInsecureNode.Name).To(gomega.Equal("cluster--Shared-Passthrough-0-insecure"))
	g.Expect(passInsecureNode.HttpPolicyRefs[0].RedirectPorts[0].Hosts[0]).To(gomega.Equal("foo.com"))
	g.Expect(passInsecureNode.HttpPolicyRefs[0].RedirectPorts[0].StatusCode).To(gomega.Equal("HTTP_REDIRECT_STATUS_CODE_302"))

	VerifyPassthroughRouteDeletion(t, g, DefaultPassthroughModel, 0, 0)
	TearDownTestForRoute(t, DefaultPassthroughModel)
}

func TestPassthroughRemoveRedirectRoute(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetUpTestForRoute(t, DefaultPassthroughModel)
	routeExample := FakeRoute{Path: "/foo"}.PassthroughRoute()
	routeExample.Spec.TLS.InsecureEdgeTerminationPolicy = routev1.InsecureEdgeTerminationPolicyRedirect
	_, err := OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	routeExample.Spec.TLS.InsecureEdgeTerminationPolicy = routev1.InsecureEdgeTerminationPolicyNone
	routeExample.ObjectMeta.ResourceVersion = "2"
	_, err = OshiftClient.RouteV1().Routes(defaultNamespace).Update(context.TODO(), routeExample, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating route: %v", err)
	}

	aviModel := ValidatePassthroughModel(t, g, DefaultPassthroughModel)
	graph := aviModel.(*avinodes.AviObjectGraph)
	vs := graph.GetAviVS()[0]

	VerifyOnePasthrough(t, g, vs)

	g.Eventually(func() int {
		aviModel := ValidatePassthroughModel(t, g, DefaultPassthroughModel)
		graph := aviModel.(*avinodes.AviObjectGraph)
		vs := graph.GetAviVS()[0]
		return len(vs.PassthroughChildNodes)
	}, 40*time.Second).Should(gomega.Equal(0))

	g.Expect(vs.PassthroughChildNodes).To(gomega.HaveLen(0))

	VerifyPassthroughRouteDeletion(t, g, DefaultPassthroughModel, 0, 0)
	TearDownTestForRoute(t, DefaultPassthroughModel)
}

func TestPassthroughAddRedirectRoute(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetUpTestForRoute(t, DefaultPassthroughModel)
	routeExample := FakeRoute{Path: "/foo"}.PassthroughRoute()
	_, err := OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	routeExample.Spec.TLS.InsecureEdgeTerminationPolicy = routev1.InsecureEdgeTerminationPolicyRedirect
	routeExample.ObjectMeta.ResourceVersion = "2"
	_, err = OshiftClient.RouteV1().Routes(defaultNamespace).Update(context.TODO(), routeExample, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating route: %v", err)
	}

	aviModel := ValidatePassthroughModel(t, g, DefaultPassthroughModel)
	graph := aviModel.(*avinodes.AviObjectGraph)
	vs := graph.GetAviVS()[0]

	VerifyOnePasthrough(t, g, vs)

	g.Expect(vs.PassthroughChildNodes).To(gomega.HaveLen(1))
	passInsecureNode := vs.PassthroughChildNodes[0]
	g.Expect(passInsecureNode.Name).To(gomega.Equal("cluster--Shared-Passthrough-0-insecure"))
	g.Expect(passInsecureNode.HttpPolicyRefs[0].RedirectPorts[0].Hosts[0]).To(gomega.Equal("foo.com"))
	g.Expect(passInsecureNode.HttpPolicyRefs[0].RedirectPorts[0].StatusCode).To(gomega.Equal("HTTP_REDIRECT_STATUS_CODE_302"))

	VerifyPassthroughRouteDeletion(t, g, DefaultPassthroughModel, 0, 0)
	TearDownTestForRoute(t, DefaultPassthroughModel)
}

func TestPassthroughToInsecureRoute(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetUpTestForRoute(t, DefaultPassthroughModel)
	routeExample := FakeRoute{Path: "/foo"}.PassthroughRoute()
	_, err := OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}
	passthroughModel := ValidatePassthroughModel(t, g, DefaultPassthroughModel)

	routeExample.Spec.TLS = nil
	routeExample.ObjectMeta.ResourceVersion = "2"
	_, err = OshiftClient.RouteV1().Routes(defaultNamespace).Update(context.TODO(), routeExample, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating route: %v", err)
	}

	g.Eventually(func() int {
		passnodes := passthroughModel.(*avinodes.AviObjectGraph).GetAviVS()
		vsvipNode := passnodes[0].VSVIPRefs[0]
		return len(vsvipNode.FQDNs)
	}, 50*time.Second).Should(gomega.Equal(0))

	aviModel := ValidateModelCommon(t, g)
	pool := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].PoolRefs[0]

	g.Eventually(func() int {
		pool = aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].PoolRefs[0]
		return len(pool.Servers)
	}, 60*time.Second).Should(gomega.Equal(1))

	g.Expect(pool.Name).To(gomega.Equal("cluster--foo.com_foo-default-foo-avisvc"))
	g.Expect(pool.PriorityLabel).To(gomega.Equal("foo.com/foo"))

	poolgroups := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].PoolGroupRefs
	pgmember := poolgroups[0].Members[0]
	g.Expect(*pgmember.PoolRef).To(gomega.Equal("/api/pool?name=cluster--foo.com_foo-default-foo-avisvc"))
	g.Expect(*pgmember.PriorityLabel).To(gomega.Equal("foo.com/foo"))

	VerifyRouteDeletion(t, g, aviModel, 0)
	TearDownTestForRoute(t, DefaultPassthroughModel)
	objects.SharedAviGraphLister().Delete(defaultModelName)
}

func TestInsecureToPassthroughRoute(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetUpTestForRoute(t, DefaultPassthroughModel)

	routeExample := FakeRoute{Path: "/foo"}.Route()
	_, err := OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	aviL7Model := ValidateModelCommon(t, g)

	routeExample = FakeRoute{Path: "/foo"}.PassthroughRoute()
	routeExample.ObjectMeta.ResourceVersion = "2"
	_, err = OshiftClient.RouteV1().Routes(defaultNamespace).Update(context.TODO(), routeExample, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	g.Eventually(func() bool {
		l7nodes := aviL7Model.(*avinodes.AviObjectGraph).GetAviVS()
		vsvipNode := l7nodes[0].VSVIPRefs[0]
		for _, fqdn := range vsvipNode.FQDNs {
			if fqdn == "foo.com" {
				return true
			}
		}
		return false
	}, 50*time.Second).Should(gomega.Equal(false))

	aviModel := ValidatePassthroughModel(t, g, DefaultPassthroughModel)
	graph := aviModel.(*avinodes.AviObjectGraph)
	vs := graph.GetAviVS()[0]
	VerifyOnePasthrough(t, g, vs)

	g.Expect(vs.PassthroughChildNodes).To(gomega.HaveLen(0))

	VerifyPassthroughRouteDeletion(t, g, DefaultPassthroughModel, 0, 0)
	TearDownTestForRoute(t, DefaultPassthroughModel)
	objects.SharedAviGraphLister().Delete(defaultModelName)
}

func TestPassthroughToSecureRoute(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetUpTestForRoute(t, DefaultPassthroughModel)
	routeExample := FakeRoute{Path: "/foo"}.PassthroughRoute()
	_, err := OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}
	passthroughModel := ValidatePassthroughModel(t, g, DefaultPassthroughModel)

	routeExample = FakeRoute{Path: "/foo"}.SecureRoute()
	routeExample.ObjectMeta.ResourceVersion = "2"
	_, err = OshiftClient.RouteV1().Routes(defaultNamespace).Update(context.TODO(), routeExample, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	g.Eventually(func() int {
		passnodes := passthroughModel.(*avinodes.AviObjectGraph).GetAviVS()
		vsvipNode := passnodes[0].VSVIPRefs[0]
		return len(vsvipNode.FQDNs)
	}, 50*time.Second).Should(gomega.Equal(0))

	aviModel := ValidateSniModel(t, g, defaultModelName)

	g.Eventually(func() int {
		return len(aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes)
	}, 60*time.Second).Should(gomega.Equal(1))
	sniVS := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes[0]
	g.Eventually(func() string {
		sniVS = aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0].SniNodes[0]
		return sniVS.VHDomainNames[0]
	}, 60*time.Second).Should(gomega.Equal(defaultHostname))
	VerifySniNode(g, sniVS)

	VerifySecureRouteDeletion(t, g, defaultModelName, 0, 0)
	TearDownTestForRoute(t, defaultModelName)
}

func TestSecureToPassthroughRoute(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetUpTestForRoute(t, DefaultPassthroughModel)
	routeExample := FakeRoute{Path: "/foo"}.SecureRoute()
	_, err := OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	routeExample = FakeRoute{Path: "/foo"}.PassthroughRoute()
	routeExample.ObjectMeta.ResourceVersion = "2"
	_, err = OshiftClient.RouteV1().Routes(defaultNamespace).Update(context.TODO(), routeExample, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	aviModel := ValidatePassthroughModel(t, g, DefaultPassthroughModel)
	graph := aviModel.(*avinodes.AviObjectGraph)
	vs := graph.GetAviVS()[0]
	VerifyOnePasthrough(t, g, vs)

	g.Expect(vs.PassthroughChildNodes).To(gomega.HaveLen(0))

	VerifyPassthroughRouteDeletion(t, g, DefaultPassthroughModel, 0, 0)
	TearDownTestForRoute(t, DefaultPassthroughModel)
	objects.SharedAviGraphLister().Delete(defaultModelName)
}

func TestPassthroughRouteWithAlternateBackends(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetUpTestForRoute(t, DefaultPassthroughModel)

	integrationtest.CreateSVC(t, "default", "absvc2", corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEPS(t, "default", "absvc2", false, false, "3.3.3")

	routeExample := FakeRoute{Path: "/foo"}.PassthroughABRoute()
	routeExample.Spec.TLS.InsecureEdgeTerminationPolicy = routev1.InsecureEdgeTerminationPolicyRedirect
	_, err := OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	aviModel := ValidatePassthroughModel(t, g, DefaultPassthroughModel)
	graph := aviModel.(*avinodes.AviObjectGraph)
	vs := graph.GetAviVS()[0]

	g.Expect(vs.HTTPDSrefs[0].PoolGroupRefs).To(gomega.HaveLen(1))
	g.Expect(vs.HTTPDSrefs[0].PoolGroupRefs[0]).To(gomega.Equal("cluster--foo.com"))

	g.Expect(vs.PoolGroupRefs).To(gomega.HaveLen(1))
	g.Expect(vs.PoolGroupRefs[0].Name).To(gomega.Equal("cluster--foo.com"))
	g.Expect(vs.PoolGroupRefs[0].Members).To(gomega.HaveLen(2))
	for _, member := range vs.PoolGroupRefs[0].Members {
		if *member.PoolRef == "/api/pool?name=cluster--foo.com-avisvc" {
			g.Expect(*member.Ratio).To(gomega.Equal(uint32(100)))
		} else if *member.PoolRef == "/api/pool?name=cluster--foo.com-absvc2" {
			g.Expect(*member.Ratio).To(gomega.Equal(uint32(200)))
		} else {
			t.Fatalf("Unexpected Pg member: %s", *member.PoolRef)
		}
	}

	g.Expect(vs.PoolRefs).To(gomega.HaveLen(2))
	for _, pool := range vs.PoolRefs {
		if pool.Name == "cluster--foo.com-avisvc" || pool.Name == "cluster--foo.com-absvc2" {
			g.Expect(pool.Servers).To(gomega.HaveLen(1))
		} else {
			t.Fatalf("Unexpected Pool: %s", pool.Name)
		}
	}

	g.Expect(vs.PassthroughChildNodes).To(gomega.HaveLen(1))
	passInsecureNode := vs.PassthroughChildNodes[0]
	g.Expect(passInsecureNode.Name).To(gomega.Equal("cluster--Shared-Passthrough-0-insecure"))
	g.Expect(passInsecureNode.HttpPolicyRefs[0].RedirectPorts[0].Hosts[0]).To(gomega.Equal("foo.com"))
	g.Expect(passInsecureNode.HttpPolicyRefs[0].RedirectPorts[0].StatusCode).To(gomega.Equal("HTTP_REDIRECT_STATUS_CODE_302"))

	VerifyPassthroughRouteDeletion(t, g, DefaultPassthroughModel, 0, 0)
	TearDownTestForRoute(t, DefaultPassthroughModel)

	integrationtest.DelSVC(t, "default", "absvc2")
	integrationtest.DelEPS(t, "default", "absvc2")
}

func TestPassthroughRouteRemoveAlternateBackends(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetUpTestForRoute(t, DefaultPassthroughModel)
	routeExample := FakeRoute{Path: "/foo"}.PassthroughABRoute()
	routeExample.Spec.TLS.InsecureEdgeTerminationPolicy = routev1.InsecureEdgeTerminationPolicyRedirect
	_, err := OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	routeExample = FakeRoute{Path: "/foo"}.PassthroughRoute()
	routeExample.ObjectMeta.ResourceVersion = "2"
	_, err = OshiftClient.RouteV1().Routes(defaultNamespace).Update(context.TODO(), routeExample, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating route: %v", err)
	}

	aviModel := ValidatePassthroughModel(t, g, DefaultPassthroughModel)
	graph := aviModel.(*avinodes.AviObjectGraph)
	vs := graph.GetAviVS()[0]
	VerifyOnePasthrough(t, g, vs)

	g.Expect(vs.PassthroughChildNodes).To(gomega.HaveLen(0))

	VerifyPassthroughRouteDeletion(t, g, DefaultPassthroughModel, 0, 0)
	TearDownTestForRoute(t, DefaultPassthroughModel)
}

func TestMultiplePassthroughRoutes(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	SetUpTestForRoute(t, DefaultPassthroughModel)
	routeExample := FakeRoute{Path: "/foo"}.PassthroughRoute()
	routeExample.Spec.TLS.InsecureEdgeTerminationPolicy = routev1.InsecureEdgeTerminationPolicyRedirect
	_, err := OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	routeExample2 := FakeRoute{Name: "bar", Hostname: "bar.com", Path: "/bar"}.PassthroughRoute()
	routeExample2.Spec.TLS.InsecureEdgeTerminationPolicy = routev1.InsecureEdgeTerminationPolicyRedirect
	_, err = OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExample2, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	aviModel := ValidatePassthroughModel(t, g, DefaultPassthroughModel)
	graph := aviModel.(*avinodes.AviObjectGraph)
	vs := graph.GetAviVS()[0]

	g.Eventually(func() int {
		return len(vs.HTTPDSrefs[0].PoolGroupRefs)
	}, 60*time.Second).Should(gomega.Equal(2))

	for _, pgname := range vs.HTTPDSrefs[0].PoolGroupRefs {
		if pgname != "cluster--foo.com" && pgname != "cluster--bar.com" {
			t.Fatalf("Unexpected pg ref in datascript: %s", pgname)
		}
	}

	g.Expect(vs.PoolGroupRefs).To(gomega.HaveLen(2))
	for _, pg := range vs.PoolGroupRefs {
		if pg.Name == "cluster--foo.com" {
			g.Expect(pg.Members).To(gomega.HaveLen(1))
			g.Expect(*pg.Members[0].PoolRef).To(gomega.Equal("/api/pool?name=cluster--foo.com-avisvc"))
		} else if pg.Name == "cluster--bar.com" {
			g.Expect(pg.Members).To(gomega.HaveLen(1))
			g.Expect(*pg.Members[0].PoolRef).To(gomega.Equal("/api/pool?name=cluster--bar.com-avisvc"))
		} else {
			t.Fatalf("Unexpected PG: %s", pg.Name)
		}
	}

	g.Expect(vs.PoolRefs).To(gomega.HaveLen(2))
	for _, pool := range vs.PoolRefs {
		if pool.Name == "cluster--foo.com-avisvc" || pool.Name == "cluster--bar.com-avisvc" {
			g.Expect(pool.Servers).To(gomega.HaveLen(1))
		} else {
			t.Fatalf("Unexpected Pool: %s", pool.Name)
		}
	}

	g.Expect(vs.PassthroughChildNodes).To(gomega.HaveLen(1))
	passInsecureNode := vs.PassthroughChildNodes[0]
	g.Expect(passInsecureNode.Name).To(gomega.Equal("cluster--Shared-Passthrough-0-insecure"))
	for _, redir := range passInsecureNode.HttpPolicyRefs {
		if redir.RedirectPorts[0].Hosts[0] != "foo.com" && redir.RedirectPorts[0].Hosts[0] != "bar.com" {
			t.Fatalf("unexpected redirect policy for: %s", redir.RedirectPorts[0].Hosts[0])
		}
		g.Expect(redir.RedirectPorts[0].StatusCode).To(gomega.Equal("HTTP_REDIRECT_STATUS_CODE_302"))
	}

	err = OshiftClient.RouteV1().Routes(defaultNamespace).Delete(context.TODO(), defaultRouteName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the route %v", err)
	}
	err = OshiftClient.RouteV1().Routes(defaultNamespace).Delete(context.TODO(), "bar", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}
	TearDownTestForRoute(t, DefaultPassthroughModel)
}
