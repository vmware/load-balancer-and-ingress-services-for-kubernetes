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
	_ "fmt"
	"testing"
	"time"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	avinodes "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/integrationtest"

	"github.com/onsi/gomega"
	routev1 "github.com/openshift/api/route/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestRouteCreateDeleteHostRule(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	hrname := "samplehr-foo"
	modelName := "admin/cluster--Shared-L7-0"

	SetUpTestForRoute(t, modelName)
	routeExample := FakeRoute{Path: "/foo"}.SecureRoute()
	_, err := OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	aviModel := ValidateSniModel(t, g, modelName)
	integrationtest.SetupHostRule(t, hrname, "foo.com", true)

	g.Eventually(func() string {
		hostrule, _ := lib.GetCRDClientset().AkoV1alpha1().HostRules(defaultNamespace).Get(context.TODO(), hrname, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 50*time.Second).Should(gomega.Equal("Accepted"))

	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--foo.com"}
	integrationtest.VerifyMetadataHostRule(g, sniVSKey, "default/samplehr-foo", true)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(*nodes[0].SniNodes[0].Enabled).To(gomega.Equal(true))
	g.Expect(nodes[0].SniNodes[0].SSLKeyCertAviRef).To(gomega.ContainSubstring("thisisaviref-sslkey"))
	g.Expect(nodes[0].SniNodes[0].WafPolicyRef).To(gomega.ContainSubstring("thisisaviref-waf"))
	g.Expect(nodes[0].SniNodes[0].AppProfileRef).To(gomega.ContainSubstring("thisisaviref-appprof"))
	g.Expect(nodes[0].SniNodes[0].AnalyticsProfileRef).To(gomega.ContainSubstring("thisisaviref-analyticsprof"))
	g.Expect(nodes[0].SniNodes[0].ErrorPageProfileRef).To(gomega.ContainSubstring("thisisaviref-errorprof"))
	g.Expect(nodes[0].SniNodes[0].HttpPolicySetRefs).To(gomega.HaveLen(2))
	g.Expect(nodes[0].SniNodes[0].HttpPolicySetRefs[0]).To(gomega.ContainSubstring("thisisaviref-httpps2"))
	g.Expect(nodes[0].SniNodes[0].HttpPolicySetRefs[1]).To(gomega.ContainSubstring("thisisaviref-httpps1"))
	g.Expect(nodes[0].SniNodes[0].VsDatascriptRefs).To(gomega.HaveLen(2))
	g.Expect(nodes[0].SniNodes[0].VsDatascriptRefs[0]).To(gomega.ContainSubstring("thisisaviref-ds2"))
	g.Expect(nodes[0].SniNodes[0].VsDatascriptRefs[1]).To(gomega.ContainSubstring("thisisaviref-ds1"))
	g.Expect(nodes[0].SniNodes[0].SSLProfileRef).To(gomega.ContainSubstring("thisisaviref-sslprof"))

	hrUpdate := integrationtest.FakeHostRule{
		Name:              hrname,
		Namespace:         "default",
		Fqdn:              "foo.com",
		SslKeyCertificate: "thisisaviref-sslkey",
	}.HostRule()
	enableVirtualHost := false
	hrUpdate.Spec.VirtualHost.EnableVirtualHost = &enableVirtualHost
	hrUpdate.ResourceVersion = "2"
	_, err = CRDClient.AkoV1alpha1().HostRules("default").Update(context.TODO(), hrUpdate, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating HostRule: %v", err)
	}
	g.Eventually(func() bool {
		if found, aviModel := objects.SharedAviGraphLister().Get(modelName); found {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
			return *nodes[0].SniNodes[0].Enabled
		}
		return true
	}, 25*time.Second).Should(gomega.Equal(false))

	integrationtest.TeardownHostRule(t, g, sniVSKey, hrname)
	integrationtest.VerifyMetadataHostRule(g, sniVSKey, "default/samplehr-foo", false)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes[0].SniNodes[0].Enabled).To(gomega.BeNil())
	g.Expect(nodes[0].SniNodes[0].SSLKeyCertAviRef).To(gomega.Equal(""))
	g.Expect(nodes[0].SniNodes[0].WafPolicyRef).To(gomega.Equal(""))
	g.Expect(nodes[0].SniNodes[0].AppProfileRef).To(gomega.Equal(""))
	g.Expect(nodes[0].SniNodes[0].AnalyticsProfileRef).To(gomega.Equal(""))
	g.Expect(nodes[0].SniNodes[0].ErrorPageProfileRef).To(gomega.Equal(""))
	g.Expect(nodes[0].SniNodes[0].HttpPolicySetRefs).To(gomega.HaveLen(0))
	g.Expect(nodes[0].SniNodes[0].VsDatascriptRefs).To(gomega.HaveLen(0))
	g.Expect(nodes[0].SniNodes[0].SSLProfileRef).To(gomega.Equal(""))

	VerifySecureRouteDeletion(t, g, defaultModelName, 0, 0)
	TearDownTestForRoute(t, defaultModelName)
}

func TestOshiftCreateHostRuleBeforeIngress(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName := "admin/cluster--Shared-L7-0"
	hrname := "samplehr-foo"
	integrationtest.SetupHostRule(t, hrname, "foo.com", true)

	g.Eventually(func() string {
		hostrule, _ := CRDClient.AkoV1alpha1().HostRules(defaultNamespace).Get(context.TODO(), hrname, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 50*time.Second).Should(gomega.Equal("Accepted"))

	SetUpTestForRoute(t, modelName)
	routeExample := FakeRoute{Path: "/foo"}.SecureRoute()
	_, err := OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}
	ValidateSniModel(t, g, modelName)

	g.Eventually(func() string {
		if found, aviModel := objects.SharedAviGraphLister().Get(modelName); found {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
			if len(nodes[0].SniNodes) == 1 {
				return nodes[0].SniNodes[0].SSLKeyCertAviRef
			}
		}
		return ""
	}, 50*time.Second).Should(gomega.ContainSubstring("thisisaviref-sslkey"))

	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--foo.com"}
	integrationtest.TeardownHostRule(t, g, sniVSKey, hrname)

	g.Eventually(func() string {
		if found, aviModel := objects.SharedAviGraphLister().Get(modelName); found {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
			if len(nodes[0].SniNodes) == 1 {
				return nodes[0].SniNodes[0].SSLKeyCertAviRef
			}
		}
		return ""
	}, 50*time.Second).Should(gomega.Equal(""))
	VerifySecureRouteDeletion(t, g, defaultModelName, 0, 0)
	TearDownTestForRoute(t, defaultModelName)
}

func TestOShiftRouteInsecureToSecureHostRule(t *testing.T) {
	// insecure route to secure VS via Hostrule
	g := gomega.NewGomegaWithT(t)

	modelName := "admin/cluster--Shared-L7-0"
	hrname := "samplehr-foo"

	SetUpTestForRoute(t, modelName)
	routeExample := FakeRoute{}.Route()
	if _, err := OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExample, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	mcache := cache.SharedAviObjCache()
	vsKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--Shared-L7-0"}
	g.Eventually(func() int {
		vsCache, ok := mcache.VsCacheMeta.AviCacheGet(vsKey)
		vsCacheObj, found := vsCache.(*cache.AviVsCache)
		if ok && found {
			return len(vsCacheObj.SNIChildCollection)
		}
		return 100
	}, 50*time.Second).Should(gomega.Equal(0))

	integrationtest.SetupHostRule(t, hrname, "foo.com", true)

	g.Eventually(func() int {
		vsCache, found := mcache.VsCacheMeta.AviCacheGet(vsKey)
		vsCacheObj, ok := vsCache.(*cache.AviVsCache)
		if found && ok {
			return len(vsCacheObj.SNIChildCollection)
		}
		return 0
	}, 50*time.Second).Should(gomega.Equal(1))

	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--foo.com"}
	integrationtest.VerifyMetadataHostRule(g, sniVSKey, "default/samplehr-foo", true)

	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes[0].SniNodes[0].SSLKeyCertAviRef).To(gomega.ContainSubstring("thisisaviref-sslkey"))
	g.Expect(nodes[0].SniNodes[0].WafPolicyRef).To(gomega.ContainSubstring("thisisaviref-waf"))

	integrationtest.TeardownHostRule(t, g, sniVSKey, hrname)
	VerifyRouteDeletion(t, g, aviModel, 0)
	TearDownTestForRoute(t, defaultModelName)
}

func TestOshiftMultiRouteToSecureHostRule(t *testing.T) {
	// 2 insecure route -> secure VS via Hostrule
	g := gomega.NewGomegaWithT(t)

	modelName := "admin/cluster--Shared-L7-0"
	hrname := "samplehr-foo"

	// creating insecure default/foo.com/foo
	SetUpTestForRoute(t, modelName, integrationtest.AllModels...)
	secRouteExample := FakeRoute{Path: "/foo"}.Route()
	if _, err := OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), secRouteExample, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	// creating insecure red/foo.com/bar
	routeExample := FakeRoute{
		Name: "insecure-foo",
		Path: "/bar",
	}.Route()
	_, err := OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExample, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	integrationtest.SetupHostRule(t, hrname, "foo.com", true)

	g.Eventually(func() bool {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if found {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
			if len(nodes[0].SniNodes) > 0 &&
				len(nodes[0].SniNodes[0].PoolRefs) == 2 {
				return true
			}
		}
		return false
	}, 90*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes[0].SniNodes[0].SSLKeyCertAviRef).To(gomega.ContainSubstring("thisisaviref-sslkey"))
	g.Expect(nodes[0].SniNodes[0].SSLKeyCertRefs).To(gomega.HaveLen(0))

	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--foo.com"}
	integrationtest.VerifyMetadataHostRule(g, sniVSKey, "default/samplehr-foo", true)

	integrationtest.TeardownHostRule(t, g, sniVSKey, hrname)
	VerifySecureRouteDeletion(t, g, modelName, 1, 0)
	VerifyRouteDeletion(t, g, aviModel, 0, "default/insecure-foo")
	TearDownTestForRoute(t, modelName)
}

func TestOshiftMultiRouteSwitchHostRuleFqdn(t *testing.T) {
	// 2 insecure routes -> secure VS via Hostrule
	g := gomega.NewGomegaWithT(t)

	modelName := "admin/cluster--Shared-L7-0"
	hrname := "samplehr-foo"

	// creating insecure default/foo.com/foo
	SetUpTestForRoute(t, modelName, integrationtest.AllModels...)
	routeExample := FakeRoute{Path: "/foo"}.Route()
	if _, err := OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExample, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	// creating insecure red/voo.com/voo
	routeExampleVoo := FakeRoute{
		Name:     "voo",
		Hostname: "voo.com",
		Path:     "/voo",
	}.Route()
	if _, err := OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExampleVoo, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	// hostrule for foo.com
	integrationtest.SetupHostRule(t, hrname, "foo.com", true)

	// voo.com must be insecure, foo.com must be secure
	// both foo.com and voo.com fall in the SAME shard
	g.Eventually(func() bool {
		if found, aviModel := objects.SharedAviGraphLister().Get(modelName); found {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
			if len(nodes[0].SniNodes) == 1 &&
				len(nodes[0].PoolRefs) == 1 &&
				nodes[0].PoolRefs[0].Name == "cluster--voo.com_voo-default-voo-avisvc" {
				return true
			}
		}
		return false
	}, 30*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes[0].SniNodes[0].Name).To(gomega.Equal("cluster--foo.com"))
	g.Expect(nodes[0].SniNodes[0].PoolRefs[0].Name).To(gomega.Equal("cluster--default-foo.com_foo-foo-avisvc"))

	// change hostrule for foo.com to voo.com
	hrUpdate := integrationtest.FakeHostRule{
		Name:              hrname,
		Namespace:         "default",
		Fqdn:              "voo.com",
		SslKeyCertificate: "thisisaviref-sslkey",
	}.HostRule()
	hrUpdate.ResourceVersion = "2"
	if _, err := CRDClient.AkoV1alpha1().HostRules("default").Update(context.TODO(), hrUpdate, metav1.UpdateOptions{}); err != nil {
		t.Fatalf("error in updating HostRule: %v", err)
	}

	// foo.com would be insecure, voo.com would become secure now
	g.Eventually(func() bool {
		if found, aviModel := objects.SharedAviGraphLister().Get(modelName); found {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
			if len(nodes[0].SniNodes) == 1 &&
				nodes[0].SniNodes[0].Name == "cluster--voo.com" &&
				len(nodes[0].PoolRefs) == 1 &&
				nodes[0].PoolRefs[0].Name == "cluster--foo.com_foo-default-foo-avisvc" {
				return true
			}
		}
		return false
	}, 30*time.Second).Should(gomega.Equal(true))
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes[0].SniNodes[0].PoolRefs[0].Name).To(gomega.Equal("cluster--default-voo.com_voo-voo-avisvc"))

	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--voo.com"}
	integrationtest.TeardownHostRule(t, g, sniVSKey, hrname)
	VerifySecureRouteDeletion(t, g, modelName, 1, 0, "default/voo")
	VerifyRouteDeletion(t, g, aviModel, 0)
	TearDownTestForRoute(t, defaultModelName)
}

func TestOshiftGoodToBadHostRule(t *testing.T) {
	// create insecure route, apply good secure hostrule, transition to bad
	g := gomega.NewGomegaWithT(t)
	modelName := "admin/cluster--Shared-L7-0"
	hrname := "samplehr-foo"

	SetUpTestForRoute(t, modelName)
	routeExample := FakeRoute{}.Route()
	if _, err := OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExample, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--foo.com"}
	integrationtest.VerifyMetadataHostRule(g, sniVSKey, "default/samplehr-foo", false)
	integrationtest.SetupHostRule(t, hrname, "foo.com", true)
	integrationtest.VerifyMetadataHostRule(g, sniVSKey, "default/samplehr-foo", true)

	// update hostrule with bad ref
	hrUpdate := integrationtest.FakeHostRule{
		Name:               hrname,
		Namespace:          "default",
		Fqdn:               "voo.com",
		WafPolicy:          "thisisBADaviref",
		ApplicationProfile: "thisisaviref-appprof",
	}.HostRule()
	hrUpdate.ResourceVersion = "2"
	if _, err := CRDClient.AkoV1alpha1().HostRules("default").Update(context.TODO(), hrUpdate, metav1.UpdateOptions{}); err != nil {
		t.Fatalf("error in updating HostRule: %v", err)
	}

	g.Eventually(func() string {
		hostrule, _ := CRDClient.AkoV1alpha1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 30*time.Second).Should(gomega.Equal("Rejected"))

	// the last applied hostrule values would exist
	g.Eventually(func() string {
		if found, aviModel := objects.SharedAviGraphLister().Get(modelName); found {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
			if len(nodes[0].SniNodes) == 1 {
				return nodes[0].SniNodes[0].SSLKeyCertAviRef
			}
		}
		return ""
	}, 30*time.Second).Should(gomega.ContainSubstring("thisisaviref"))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes[0].SniNodes[0].WafPolicyRef).To(gomega.ContainSubstring("thisisaviref-waf"))
	g.Expect(nodes[0].SniNodes[0].AppProfileRef).To(gomega.ContainSubstring("thisisaviref-appprof"))

	integrationtest.TeardownHostRule(t, g, sniVSKey, hrname)
	VerifyRouteDeletion(t, g, aviModel, 0)
	TearDownTestForRoute(t, defaultModelName)
}

func TestOshiftInsecureHostAndHostrule(t *testing.T) {
	// create insecure route, insecure hostrule, nothing should be applied
	g := gomega.NewGomegaWithT(t)
	modelName := "admin/cluster--Shared-L7-0"
	hrname := "samplehr-foo"

	SetUpTestForRoute(t, modelName)
	routeExample := FakeRoute{}.Route()
	if _, err := OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExample, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding route: %v", err)
	}
	integrationtest.SetupHostRule(t, hrname, "foo.com", false)

	g.Eventually(func() int {
		if found, aviModel := objects.SharedAviGraphLister().Get(modelName); found {
			if nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS(); len(nodes) > 0 {
				return len(nodes[0].PoolRefs)
			}
		}
		return 0
	}, 10*time.Second).Should(gomega.Equal(1))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes[0].SniNodes).To(gomega.HaveLen(0))

	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--foo.com"}
	integrationtest.TeardownHostRule(t, g, sniVSKey, hrname)
	VerifyRouteDeletion(t, g, aviModel, 0)
	TearDownTestForRoute(t, defaultModelName)
}

func TestOshiftValidToInvalidHostSwitch(t *testing.T) {
	// create insecure host foo.com, attach hostrule, change hostrule to non existing bar.com
	// foo.com should become insecure again
	// change hostrule back to foo.com and it should become secure again
	g := gomega.NewGomegaWithT(t)
	modelName := "admin/cluster--Shared-L7-0"
	hrname := "samplehr-foo"

	SetUpTestForRoute(t, modelName)
	routeExample := FakeRoute{Path: "/foo"}.Route()
	if _, err := OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExample, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding route: %v", err)
	}
	integrationtest.SetupHostRule(t, hrname, "foo.com", true)

	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--foo.com"}
	integrationtest.VerifyMetadataHostRule(g, sniVSKey, "default/samplehr-foo", true)

	hrUpdate := integrationtest.FakeHostRule{
		Name:              hrname,
		Namespace:         "default",
		Fqdn:              "bar.com",
		SslKeyCertificate: "thisisaviref-sslkey",
	}.HostRule()
	hrUpdate.ResourceVersion = "2"
	if _, err := CRDClient.AkoV1alpha1().HostRules("default").Update(context.TODO(), hrUpdate, metav1.UpdateOptions{}); err != nil {
		t.Fatalf("error in updating HostRule: %v", err)
	}

	g.Eventually(func() int {
		if found, aviModel := objects.SharedAviGraphLister().Get(modelName); found {
			if nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS(); len(nodes) > 0 {
				return len(nodes[0].PoolRefs)
			}
		}
		return 0
	}, 10*time.Second).Should(gomega.Equal(1))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes[0].PoolRefs[0].Name).To(gomega.Equal("cluster--foo.com_foo-default-foo-avisvc"))

	// change back to good host
	hrUpdate = integrationtest.FakeHostRule{
		Name:              hrname,
		Namespace:         "default",
		Fqdn:              "foo.com",
		SslKeyCertificate: "thisisaviref-sslkey",
	}.HostRule()
	hrUpdate.ResourceVersion = "3"
	if _, err := CRDClient.AkoV1alpha1().HostRules("default").Update(context.TODO(), hrUpdate, metav1.UpdateOptions{}); err != nil {
		t.Fatalf("error in updating HostRule: %v", err)
	}

	g.Eventually(func() int {
		if found, aviModel := objects.SharedAviGraphLister().Get(modelName); found {
			if nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS(); len(nodes) > 0 {
				return len(nodes[0].PoolRefs)
			}
		}
		return -1
	}, 10*time.Second).Should(gomega.Equal(0))
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes[0].SniNodes[0].PoolRefs[0].Name).To(gomega.Equal("cluster--default-foo.com_foo-foo-avisvc"))

	integrationtest.TeardownHostRule(t, g, sniVSKey, hrname)
	VerifyRouteDeletion(t, g, aviModel, 0)
	TearDownTestForRoute(t, defaultModelName)
}

func TestOshiftHTTPRuleCreateDelete(t *testing.T) {
	// route secure foo.com/foo /bar
	// create httprule /, httprule gets attached check on /foo /bar
	// delete hostrule, httprule gets detached
	g := gomega.NewGomegaWithT(t)

	modelName := "admin/cluster--Shared-L7-0"
	rrname := "samplerr-foo"

	SetUpTestForRoute(t, modelName)
	routeExampleFoo := FakeRoute{Path: "/foo"}.SecureRoute()
	if _, err := OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExampleFoo, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding route: %v", err)
	}
	routeExampleBar := FakeRoute{Name: "foobar", Path: "/bar"}.SecureRoute()
	if _, err := OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExampleBar, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	poolFooKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--default-foo.com_foo-foo-avisvc"}
	poolBarKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--default-foo.com_bar-foobar-avisvc"}
	integrationtest.SetupHTTPRule(t, rrname, "foo.com", "/")
	integrationtest.VerifyMetadataHTTPRule(g, poolFooKey, "default/"+rrname, true)
	integrationtest.VerifyMetadataHTTPRule(g, poolBarKey, "default/"+rrname, true)
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes[0].SniNodes[0].PoolRefs[0].LbAlgorithm).To(gomega.Equal("LB_ALGORITHM_CONSISTENT_HASH"))
	g.Expect(nodes[0].SniNodes[0].PoolRefs[0].LbAlgorithmHash).To(gomega.Equal("LB_ALGORITHM_CONSISTENT_HASH_SOURCE_IP_ADDRESS"))
	g.Expect(nodes[0].SniNodes[0].PoolRefs[0].SslProfileRef).To(gomega.ContainSubstring("thisisaviref-sslprofile"))
	g.Expect(nodes[0].SniNodes[0].PoolRefs[0].PkiProfile.CACert).To(gomega.Equal("httprule-destinationCA"))
	g.Expect(nodes[0].SniNodes[0].PoolRefs[0].HealthMonitors).To(gomega.HaveLen(2))
	g.Expect(nodes[0].SniNodes[0].PoolRefs[0].HealthMonitors[0]).To(gomega.ContainSubstring("thisisaviref-hm2"))
	g.Expect(nodes[0].SniNodes[0].PoolRefs[0].HealthMonitors[1]).To(gomega.ContainSubstring("thisisaviref-hm1"))

	// delete httprule deletes refs as well
	integrationtest.TeardownHTTPRule(t, rrname)
	integrationtest.VerifyMetadataHTTPRule(g, poolFooKey, "default/"+rrname, false)
	integrationtest.VerifyMetadataHTTPRule(g, poolBarKey, "default/"+rrname, false)
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes[0].SniNodes[0].PoolRefs[1].LbAlgorithm).To(gomega.Equal(""))
	g.Expect(nodes[0].SniNodes[0].PoolRefs[0].SslProfileRef).To(gomega.Equal(""))
	g.Expect(nodes[0].SniNodes[0].PoolRefs[0].PkiProfile).To(gomega.BeNil())
	g.Expect(nodes[0].SniNodes[0].PoolRefs[0].HealthMonitors).To(gomega.HaveLen(0))

	VerifySecureRouteDeletion(t, g, modelName, 0, 1)
	VerifySecureRouteDeletion(t, g, modelName, 0, 0, "default/foobar")
	TearDownTestForRoute(t, defaultModelName)
}

func TestOshiftHTTPRuleHostSwitch(t *testing.T) {
	// ingress foo.com/foo voo.com/foo
	// hr1: foo.com (secure), hr2: voo.com (insecure)
	// rr1: hr1/foo ALGO1
	// switch rr1 host to hr2
	g := gomega.NewGomegaWithT(t)

	modelName := "admin/cluster--Shared-L7-0"
	rrnameFoo := "samplerr-foo"

	// creates foo.com insecure
	SetUpTestForRoute(t, modelName)
	routeExampleFoo := FakeRoute{Path: "/foo"}.SecureRoute()
	if _, err := OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExampleFoo, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding route: %v", err)
	}
	routeExampleVoo := FakeRoute{Name: "voo", Hostname: "voo.com", Path: "/foo"}.Route()
	if _, err := OshiftClient.RouteV1().Routes(defaultNamespace).Create(context.TODO(), routeExampleVoo, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding route: %v", err)
	}

	integrationtest.SetupHTTPRule(t, rrnameFoo, "foo.com", "/foo")
	g.Eventually(func() bool {
		if found, aviModel := objects.SharedAviGraphLister().Get(modelName); found {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
			if len(nodes[0].PoolRefs) == 1 &&
				nodes[0].PoolRefs[0].LbAlgorithm == "" &&
				len(nodes[0].SniNodes) == 1 &&
				len(nodes[0].SniNodes[0].PoolRefs) == 1 &&
				nodes[0].SniNodes[0].PoolRefs[0].LbAlgorithm == "LB_ALGORITHM_CONSISTENT_HASH" {
				return true
			}
		}
		return false
	}, 25*time.Second).Should(gomega.Equal(true))

	// update httprule's hostrule pointer from FOO to VOO
	rrUpdate := integrationtest.FakeHTTPRule{
		Name:      rrnameFoo,
		Namespace: "default",
		Fqdn:      "voo.com",
		PathProperties: []integrationtest.FakeHTTPRulePath{{
			Path:        "/foo",
			LbAlgorithm: "LB_ALGORITHM_CONSISTENT_HASH",
		}},
	}.HTTPRule()
	rrUpdate.ResourceVersion = "2"
	if _, err := CRDClient.AkoV1alpha1().HTTPRules("default").Update(context.TODO(), rrUpdate, metav1.UpdateOptions{}); err != nil {
		t.Fatalf("error in updating HostRule: %v", err)
	}

	// httprule things should get attached to insecure Pools of bar.com now
	// earlier since the hostrule pointed to secure foo.com, it was attached to the SNI pools
	g.Eventually(func() bool {
		if found, aviModel := objects.SharedAviGraphLister().Get(modelName); found {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
			if len(nodes[0].PoolRefs) == 1 &&
				nodes[0].PoolRefs[0].LbAlgorithm == "LB_ALGORITHM_CONSISTENT_HASH" &&
				len(nodes[0].SniNodes) == 1 &&
				len(nodes[0].SniNodes[0].PoolRefs) == 1 &&
				nodes[0].SniNodes[0].PoolRefs[0].LbAlgorithm == "" {
				return true
			}
		}
		return false
	}, 25*time.Second).Should(gomega.Equal(true))

	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	integrationtest.TeardownHTTPRule(t, rrnameFoo)
	VerifyRouteDeletion(t, g, aviModel, 0, "default/voo")
	VerifySecureRouteDeletion(t, g, modelName, 0, 0)
	TearDownTestForRoute(t, defaultModelName)
}

func TestOshiftHTTPRuleReencryptWithDestinationCA(t *testing.T) {
	// create route foo.com/foo, with destinationCA in Route,
	// add destinationCA via httprule, overwrites Route, delete httprule, fallback to Route
	g := gomega.NewGomegaWithT(t)
	SetUpTestForRoute(t, defaultModelName)
	rrname := "samplerr-foo"

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

	g.Expect(sniVS.PoolRefs[0].SslProfileRef).To(gomega.Equal("/api/sslprofile?name=System-Standard"))
	g.Expect(sniVS.PoolRefs[0].PkiProfile.Name).To(gomega.Equal("cluster--default-foo.com_foo-foo-avisvc-pkiprofile"))
	g.Expect(sniVS.PoolRefs[0].PkiProfile.CACert).To(gomega.Equal("abc"))

	integrationtest.SetupHTTPRule(t, rrname, "foo.com", "/")
	g.Eventually(func() bool {
		if sniVS.PoolRefs[0].PkiProfile.CACert == "httprule-destinationCA" {
			return true
		}
		return false
	}, 50*time.Second).Should(gomega.Equal(true))

	integrationtest.TeardownHTTPRule(t, rrname)
	g.Eventually(func() bool {
		if sniVS.PoolRefs[0].PkiProfile.CACert == "abc" {
			return true
		}
		return false
	}, 50*time.Second).Should(gomega.Equal(true))
	g.Expect(sniVS.PoolRefs[0].PkiProfile.Name).To(gomega.Equal("cluster--default-foo.com_foo-foo-avisvc-pkiprofile"))

	VerifySecureRouteDeletion(t, g, defaultModelName, 0, 0)
	TearDownTestForRoute(t, defaultModelName)
}
