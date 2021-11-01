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

package ingresstests

import (
	"context"
	"testing"
	"time"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	avinodes "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/testlib"

	"github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCreateDeleteHostRule(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName := "admin/cluster--Shared-L7-0"
	hrname := "samplehr-foo"
	SetUpIngressForCacheSyncCheck(t, true, true, modelName)

	testlib.SetupHostRule(t, hrname, "foo.com", true)

	g.Eventually(func() string {
		hostrule, _ := lib.AKOControlConfig().CRDClientset().AkoV1alpha1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 10*time.Second).Should(gomega.Equal("Accepted"))

	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--foo.com"}
	testlib.VerifyMetadataHostRule(t, g, sniVSKey, "default/samplehr-foo", true)
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
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
	g.Expect(nodes[0].SniNodes[0].VHDomainNames).To(gomega.ContainElement("bar.com"))

	hrUpdate := testlib.FakeHostRule{
		Name:              hrname,
		Namespace:         "default",
		Fqdn:              "foo.com",
		SslKeyCertificate: "thisisaviref-sslkey",
	}.HostRule()
	enableVirtualHost := false
	hrUpdate.Spec.VirtualHost.EnableVirtualHost = &enableVirtualHost
	hrUpdate.Spec.VirtualHost.Gslb.Fqdn = "baz.com"
	hrUpdate.ResourceVersion = "2"
	testlib.UpdateObjectOrFail(t, lib.HostRule, hrUpdate)

	g.Eventually(func() bool {
		if found, aviModel := objects.SharedAviGraphLister().Get(modelName); found {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
			return *nodes[0].SniNodes[0].Enabled
		}
		return true
	}, 25*time.Second).Should(gomega.Equal(false))
	g.Expect(nodes[0].SniNodes[0].VHDomainNames).To(gomega.ContainElement("baz.com"))
	testlib.TeardownHostRule(t, g, sniVSKey, hrname)
	testlib.VerifyMetadataHostRule(t, g, sniVSKey, "default/samplehr-foo", false)
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
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
	g.Expect(nodes[0].SniNodes[0].VHDomainNames).To(gomega.Not(gomega.ContainElement("baz.com")))

	TearDownIngressForCacheSyncCheck(t, modelName)
}

func TestCreateHostRuleBeforeIngress(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName := "admin/cluster--Shared-L7-0"
	hrname := "samplehr-foo"
	testlib.SetupHostRule(t, hrname, "foo.com", true)

	g.Eventually(func() string {
		hostrule, _ := lib.AKOControlConfig().CRDClientset().AkoV1alpha1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 10*time.Second).Should(gomega.Equal("Accepted"))

	SetUpIngressForCacheSyncCheck(t, true, true, modelName)

	g.Eventually(func() string {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		if len(nodes[0].SniNodes) == 1 {
			return nodes[0].SniNodes[0].SSLKeyCertAviRef
		}
		return ""
	}, 10*time.Second).Should(gomega.ContainSubstring("thisisaviref-sslkey"))

	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--foo.com"}
	testlib.TeardownHostRule(t, g, sniVSKey, hrname)

	g.Eventually(func() string {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		if len(nodes[0].SniNodes) == 1 {
			return nodes[0].SniNodes[0].SSLKeyCertAviRef
		}
		return ""
	}, 10*time.Second).Should(gomega.Equal(""))
	TearDownIngressForCacheSyncCheck(t, modelName)
}

func TestInsecureToSecureHostRule(t *testing.T) {
	// insecure ingress to secure VS via Hostrule
	g := gomega.NewGomegaWithT(t)

	modelName := "admin/cluster--Shared-L7-0"
	hrname := "samplehr-foo"
	SetUpIngressForCacheSyncCheck(t, false, false, modelName)

	mcache := cache.SharedAviObjCache()
	vsKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--Shared-L7-0"}
	g.Eventually(func() int {
		vsCache, _ := mcache.VsCacheMeta.AviCacheGet(vsKey)
		vsCacheObj, _ := vsCache.(*cache.AviVsCache)
		return len(vsCacheObj.SNIChildCollection)
	}, 15*time.Second).Should(gomega.Equal(0))

	testlib.SetupHostRule(t, hrname, "foo.com", true)

	g.Eventually(func() int {
		vsCache, _ := mcache.VsCacheMeta.AviCacheGet(vsKey)
		vsCacheObj, _ := vsCache.(*cache.AviVsCache)
		return len(vsCacheObj.SNIChildCollection)
	}, 15*time.Second).Should(gomega.Equal(1))

	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--foo.com"}
	testlib.VerifyMetadataHostRule(t, g, sniVSKey, "default/samplehr-foo", true)

	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes[0].SniNodes[0].SSLKeyCertAviRef).To(gomega.ContainSubstring("thisisaviref-sslkey"))
	g.Expect(nodes[0].SniNodes[0].WafPolicyRef).To(gomega.ContainSubstring("thisisaviref-waf"))
	g.Expect(nodes[0].HttpPolicyRefs[0].RedirectPorts[0].StatusCode).To(gomega.Equal("HTTP_REDIRECT_STATUS_CODE_302"))

	testlib.TeardownHostRule(t, g, sniVSKey, hrname)
	TearDownIngressForCacheSyncCheck(t, modelName)
}

func TestGSLBHostRewriteRule(t *testing.T) {
	// insecure ingress to secure VS via Hostrule
	g := gomega.NewGomegaWithT(t)

	modelName := "admin/cluster--Shared-L7-0"
	hrname := "samplehr-foo"
	SetUpIngressForCacheSyncCheck(t, false, false, modelName)

	mcache := cache.SharedAviObjCache()
	vsKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--Shared-L7-0"}
	g.Eventually(func() int {
		vsCache, _ := mcache.VsCacheMeta.AviCacheGet(vsKey)
		vsCacheObj, _ := vsCache.(*cache.AviVsCache)
		return len(vsCacheObj.SNIChildCollection)
	}, 30*time.Second).Should(gomega.Equal(0))

	testlib.SetupHostRule(t, hrname, "foo.com", false)
	g.Eventually(func() int {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return len(nodes[0].GetHttpPolicyRefs())
	}, 50*time.Second).Should(gomega.Equal(1))

	g.Eventually(func() int {
		vsCache, _ := mcache.VsCacheMeta.AviCacheGet(vsKey)
		vsCacheObj, _ := vsCache.(*cache.AviVsCache)
		return len(vsCacheObj.HTTPKeyCollection)
	}, 30*time.Second).Should(gomega.Equal(1))

	// Update the hostrule with a different GSLB host name
	testlib.SetupHostRule(t, hrname, "foo.com", false, "baz.com")

	g.Eventually(func() bool {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return nodes[0].GetHttpPolicyRefs()[0].HeaderReWrite.SourceHost == "baz.com"
	}, 50*time.Second).Should(gomega.Equal(true))

	testlib.DeleteObject(t, lib.HostRule, hrname, "default")
	g.Eventually(func() int {
		vsCache, _ := mcache.VsCacheMeta.AviCacheGet(vsKey)
		vsCacheObj, _ := vsCache.(*cache.AviVsCache)
		return len(vsCacheObj.HTTPKeyCollection)
	}, 30*time.Second).Should(gomega.Equal(0))

	TearDownIngressForCacheSyncCheck(t, modelName)
}

func TestMultiIngressToSecureHostRule(t *testing.T) {
	// 1 insecure ingress, 1 secure ingress -> secure VS via Hostrule
	g := gomega.NewGomegaWithT(t)

	modelName := "admin/cluster--Shared-L7-0"
	hrname := "samplehr-foo"

	// creating secure default/foo.com/foo
	SetUpIngressForCacheSyncCheck(t, true, true, modelName)

	// creating insecure red/foo.com/bar
	ingressObject := testlib.FakeIngress{
		Name:        "foo-with-targets-2",
		Namespace:   "red",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/bar"},
		ServiceName: "avisvc",
	}
	ingrFake := ingressObject.Ingress()
	if _, err := utils.GetInformers().ClientSet.NetworkingV1().Ingresses("red").Create(context.TODO(), ingrFake, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in updating Ingress: %v", err)
	}

	testlib.SetupHostRule(t, hrname, "foo.com", true)

	g.Eventually(func() int {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		if len(nodes[0].SniNodes) > 0 {
			return len(nodes[0].SniNodes[0].PoolGroupRefs)
		}
		return 0
	}, 50*time.Second).Should(gomega.Equal(2))
	VerifyPoolDeletionFromVsNode(g, modelName)

	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes[0].SniNodes[0].PoolRefs).To(gomega.HaveLen(2))
	g.Expect(nodes[0].SniNodes[0].SSLKeyCertAviRef).To(gomega.ContainSubstring("thisisaviref-sslkey"))
	g.Expect(nodes[0].SniNodes[0].SSLKeyCertRefs).To(gomega.HaveLen(0))
	if len(nodes[0].HttpPolicyRefs) > 0 {
		g.Expect(nodes[0].HttpPolicyRefs[0].RedirectPorts[0].StatusCode).To(gomega.Equal("HTTP_REDIRECT_STATUS_CODE_302"))
	}

	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--foo.com"}
	testlib.VerifyMetadataHostRule(t, g, sniVSKey, "default/samplehr-foo", true)

	testlib.DeleteObject(t, lib.Ingress, "foo-with-targets-2", "red")
	testlib.TeardownHostRule(t, g, sniVSKey, hrname)
	TearDownIngressForCacheSyncCheck(t, modelName)
}

func TestMultiIngressSwitchHostRuleFqdn(t *testing.T) {
	// 2 insecure ingresses -> VS via Hostrule
	g := gomega.NewGomegaWithT(t)

	modelName := "admin/cluster--Shared-L7-0"
	hrname := "samplehr-foo"

	// creating insecure default/foo.com/foo
	SetUpIngressForCacheSyncCheck(t, false, false, modelName)

	// creating insecure red/voo.com/voo
	ingressObject := testlib.FakeIngress{
		Name:        "voo-with-targets",
		Namespace:   "red",
		DnsNames:    []string{"voo.com"},
		Paths:       []string{"/voo"},
		ServiceName: "avisvc",
	}
	ingrFake := ingressObject.Ingress()
	if _, err := utils.GetInformers().ClientSet.NetworkingV1().Ingresses("red").Create(context.TODO(), ingrFake, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in updating Ingress: %v", err)
	}

	// hostrule for foo.com
	testlib.SetupHostRule(t, hrname, "foo.com", true)

	// voo.com must be insecure, foo.com must be secure
	// both foo.com and voo.com fall in the SAME shard
	g.Eventually(func() bool {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		if len(nodes[0].SniNodes) == 1 &&
			len(nodes[0].PoolRefs) == 1 &&
			nodes[0].PoolRefs[0].Name == "cluster--voo.com_voo-red-voo-with-targets" {
			return true
		}
		return false
	}, 30*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes[0].SniNodes[0].Name).To(gomega.Equal("cluster--foo.com"))
	g.Expect(nodes[0].SniNodes[0].PoolRefs[0].Name).To(gomega.Equal("cluster--default-foo.com_foo-foo-with-targets"))

	// change hostrule for foo.com to voo.com
	hrUpdate := testlib.FakeHostRule{
		Name:              hrname,
		Namespace:         "default",
		Fqdn:              "voo.com",
		SslKeyCertificate: "thisisaviref-sslkey",
	}.HostRule()
	hrUpdate.ResourceVersion = "2"
	testlib.UpdateObjectOrFail(t, lib.HostRule, hrUpdate)
	// foo.com would be insecure, voo.com would become secure now
	g.Eventually(func() bool {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		if len(nodes[0].SniNodes) == 1 &&
			nodes[0].SniNodes[0].Name == "cluster--voo.com" &&
			len(nodes[0].PoolRefs) == 1 {
			return true
		}
		return false
	}, 30*time.Second).Should(gomega.Equal(true))
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes[0].PoolRefs[0].Name).To(gomega.Equal("cluster--foo.com_foo-default-foo-with-targets"))
	g.Expect(nodes[0].SniNodes[0].PoolRefs[0].Name).To(gomega.Equal("cluster--red-voo.com_voo-voo-with-targets"))

	testlib.DeleteObject(t, lib.Ingress, "voo-with-targets", "red")
	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--voo.com"}
	testlib.TeardownHostRule(t, g, sniVSKey, hrname)
	TearDownIngressForCacheSyncCheck(t, modelName)
}

func TestGoodToBadHostRule(t *testing.T) {
	// create insecure ingress, apply good secure hostrule, transition to bad
	g := gomega.NewGomegaWithT(t)

	modelName := "admin/cluster--Shared-L7-0"
	hrname := "samplehr-foo"
	SetUpIngressForCacheSyncCheck(t, false, false, modelName)
	testlib.SetupHostRule(t, hrname, "foo.com", true)

	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--foo.com"}
	testlib.VerifyMetadataHostRule(t, g, sniVSKey, "default/samplehr-foo", true)

	// update hostrule with bad ref
	hrUpdate := testlib.FakeHostRule{
		Name:               hrname,
		Namespace:          "default",
		Fqdn:               "voo.com",
		WafPolicy:          "thisisBADaviref",
		ApplicationProfile: "thisisaviref-appprof",
	}.HostRule()
	hrUpdate.ResourceVersion = "2"
	testlib.UpdateObjectOrFail(t, lib.HostRule, hrUpdate)
	g.Eventually(func() string {
		hostrule, _ := lib.AKOControlConfig().CRDClientset().AkoV1alpha1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 10*time.Second).Should(gomega.Equal("Rejected"))

	// the last applied hostrule values would exist
	g.Eventually(func() string {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return nodes[0].SniNodes[0].SSLKeyCertAviRef
	}, 10*time.Second).Should(gomega.ContainSubstring("thisisaviref"))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes[0].SniNodes[0].WafPolicyRef).To(gomega.ContainSubstring("thisisaviref-waf"))
	g.Expect(nodes[0].SniNodes[0].AppProfileRef).To(gomega.ContainSubstring("thisisaviref-appprof"))

	testlib.TeardownHostRule(t, g, sniVSKey, hrname)
	TearDownIngressForCacheSyncCheck(t, modelName)
}

func TestInsecureHostAndHostrule(t *testing.T) {
	// create insecure ingress, insecure hostrule, nothing should be applied
	g := gomega.NewGomegaWithT(t)

	modelName := "admin/cluster--Shared-L7-0"
	hrname := "samplehr-foo"
	SetUpIngressForCacheSyncCheck(t, false, false, modelName)
	testlib.SetupHostRule(t, hrname, "foo.com", false)

	g.Eventually(func() int {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS(); len(nodes) > 0 {
			return len(nodes[0].PoolRefs)
		}
		return 0
	}, 10*time.Second).Should(gomega.Equal(1))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes[0].SniNodes).To(gomega.HaveLen(0))

	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--foo.com"}
	testlib.TeardownHostRule(t, g, sniVSKey, hrname)
	TearDownIngressForCacheSyncCheck(t, modelName)
}

func TestValidToInvalidHostSwitch(t *testing.T) {
	// create insecure host foo.com, attach hostrule, change hostrule to non existing bar.com
	// foo.com should become insecure again
	// change hostrule back to foo.com and it should become secure again
	g := gomega.NewGomegaWithT(t)

	modelName := "admin/cluster--Shared-L7-0"
	hrname := "samplehr-foo"
	SetUpIngressForCacheSyncCheck(t, false, false, modelName)
	testlib.SetupHostRule(t, hrname, "foo.com", true)

	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--foo.com"}
	testlib.VerifyMetadataHostRule(t, g, sniVSKey, "default/samplehr-foo", true)

	hrUpdate := testlib.FakeHostRule{
		Name:              hrname,
		Namespace:         "default",
		Fqdn:              "bar.com",
		SslKeyCertificate: "thisisaviref-sslkey",
	}.HostRule()
	hrUpdate.ResourceVersion = "2"
	testlib.UpdateObjectOrFail(t, lib.HostRule, hrUpdate)
	g.Eventually(func() int {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		if len(nodes) > 0 {
			return len(nodes[0].PoolRefs)
		}
		return 0
	}, 10*time.Second).Should(gomega.Equal(1))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes[0].PoolRefs[0].Name).To(gomega.Equal("cluster--foo.com_foo-default-foo-with-targets"))

	// change back to good host
	hrUpdate = testlib.FakeHostRule{
		Name:              hrname,
		Namespace:         "default",
		Fqdn:              "foo.com",
		SslKeyCertificate: "thisisaviref-sslkey",
	}.HostRule()
	hrUpdate.ResourceVersion = "3"
	testlib.UpdateObjectOrFail(t, lib.HostRule, hrUpdate)
	VerifyPoolDeletionFromVsNode(g, modelName)
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes[0].SniNodes[0].PoolRefs[0].Name).To(gomega.Equal("cluster--default-foo.com_foo-foo-with-targets"))

	testlib.TeardownHostRule(t, g, sniVSKey, hrname)
	TearDownIngressForCacheSyncCheck(t, modelName)
}

//This tc tests hostrule state if GSLB FQDN is same as that of Local FQDN/ Host.
func TestCreateHostRuleWithGSLBFqdn(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	hrname := "samplehr-foo-1"
	testlib.SetupHostRule(t, hrname, "zoo.com", true)

	g.Eventually(func() string {
		hostrule, _ := lib.AKOControlConfig().CRDClientset().AkoV1alpha1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 10*time.Second).Should(gomega.Equal("Accepted"))
	//Update hr
	testlib.SetupHostRule(t, hrname, "zoo.com", true, "zoo.com")
	g.Eventually(func() string {
		hostrule, _ := lib.AKOControlConfig().CRDClientset().AkoV1alpha1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 10*time.Second).Should(gomega.Equal("Rejected"))

	testlib.DeleteObject(t, lib.HostRule, hrname, "default")
}

// HttpRule tests

func TestHTTPRuleCreateDelete(t *testing.T) {
	// ingress secure foo.com/foo /bar
	// create httprule /foo, nothing happens
	// create hostrule, httprule gets attached check on /foo /bar
	// delete hostrule, httprule gets detached
	g := gomega.NewGomegaWithT(t)

	modelName := "admin/cluster--Shared-L7-0"
	rrname := "samplerr-foo"

	SetupDomain()
	SetUpTestForIngress(t, modelName)
	testlib.AddSecret("my-secret", "default", "tlsCert", "tlsKey")
	testlib.PollForCompletion(t, modelName, 5)
	ingressObject := testlib.FakeIngress{
		Name:        "foo-with-targets",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		Paths:       []string{"/foo", "/bar"},
		ServiceName: "avisvc",
		TlsSecretDNS: map[string][]string{
			"my-secret": {"foo.com"},
		},
	}

	ingrFake := ingressObject.Ingress(true)
	if _, err := utils.GetInformers().ClientSet.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	testlib.PollForCompletion(t, modelName, 5)

	poolFooKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--default-foo.com_foo-foo-with-targets"}
	poolBarKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--default-foo.com_bar-foo-with-targets"}
	httpRulePath := "/"
	testlib.SetupHTTPRule(t, rrname, "foo.com", httpRulePath)
	testlib.VerifyMetadataHTTPRule(t, g, poolFooKey, "default/"+rrname+"/"+httpRulePath, true)
	testlib.VerifyMetadataHTTPRule(t, g, poolBarKey, "default/"+rrname+"/"+httpRulePath, true)
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
	testlib.DeleteObject(t, lib.HTTPRule, rrname, "default")
	testlib.VerifyMetadataHTTPRule(t, g, poolFooKey, "default/"+rrname+"/"+httpRulePath, false)
	testlib.VerifyMetadataHTTPRule(t, g, poolBarKey, "default/"+rrname+"/"+httpRulePath, false)
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes[0].SniNodes[0].PoolRefs[0].LbAlgorithm).To(gomega.Equal(""))
	g.Expect(nodes[0].SniNodes[0].PoolRefs[0].SslProfileRef).To(gomega.Equal(""))
	g.Expect(nodes[0].SniNodes[0].PoolRefs[0].PkiProfile).To(gomega.BeNil())
	g.Expect(nodes[0].SniNodes[0].PoolRefs[0].HealthMonitors).To(gomega.HaveLen(0))

	TearDownIngressForCacheSyncCheck(t, modelName)
}

func TestHTTPRuleHostSwitch(t *testing.T) {
	// ingress foo.com/foo voo.com/foo
	// hr1: foo.com (secure), hr2: voo.com (insecure)
	// rr1: hr1/foo ALGO1
	// switch rr1 host to hr2
	g := gomega.NewGomegaWithT(t)

	modelName := "admin/cluster--Shared-L7-0"
	hrnameFoo := "samplehr-foo"
	hrnameVoo := "samplehr-voo"
	rrnameFoo := "samplerr-foo"

	// creates foo.com insecure
	SetUpIngressForCacheSyncCheck(t, false, false, modelName)

	// creates voo.com insecure
	ingressObject := testlib.FakeIngress{
		Name:        "voo-with-targets",
		Namespace:   "default",
		DnsNames:    []string{"voo.com"},
		Paths:       []string{"/foo"},
		ServiceName: "avisvc",
	}
	ingrFake := ingressObject.Ingress()
	if _, err := utils.GetInformers().ClientSet.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in updating Ingress: %v", err)
	}

	testlib.SetupHTTPRule(t, rrnameFoo, "foo.com", "/foo")
	testlib.SetupHostRule(t, hrnameFoo, "foo.com", true) // makes foo.com secure
	testlib.SetupHostRule(t, hrnameVoo, "voo.com", false)

	g.Eventually(func() bool {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		if len(nodes[0].PoolRefs) == 1 &&
			nodes[0].PoolRefs[0].LbAlgorithm == "" &&
			len(nodes[0].SniNodes) == 1 &&
			len(nodes[0].SniNodes[0].PoolRefs) == 1 &&
			nodes[0].SniNodes[0].PoolRefs[0].LbAlgorithm == "LB_ALGORITHM_CONSISTENT_HASH" {
			return true
		}
		return false
	}, 25*time.Second).Should(gomega.Equal(true))

	// update httprule's hostrule pointer from FOO to VOO
	rrUpdate := testlib.FakeHTTPRule{
		Name:      rrnameFoo,
		Namespace: "default",
		Fqdn:      "voo.com",
		PathProperties: []testlib.FakeHTTPRulePath{{
			Path:        "/foo",
			LbAlgorithm: "LB_ALGORITHM_CONSISTENT_HASH",
		}},
	}.HTTPRule()
	rrUpdate.ResourceVersion = "2"
	testlib.UpdateObjectOrFail(t, lib.HTTPRule, rrUpdate)

	// httprule things should get attached to insecure Pools of bar.com now
	// earlier since the hostrule pointed to secure foo.com, it was attached to the SNI pools
	g.Eventually(func() bool {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		if len(nodes[0].PoolRefs) == 1 &&
			nodes[0].PoolRefs[0].LbAlgorithm == "LB_ALGORITHM_CONSISTENT_HASH" &&
			len(nodes[0].SniNodes) == 1 &&
			len(nodes[0].SniNodes[0].PoolRefs) == 1 &&
			nodes[0].SniNodes[0].PoolRefs[0].LbAlgorithm == "" {
			return true
		}
		return false
	}, 25*time.Second).Should(gomega.Equal(true))

	testlib.DeleteObject(t, lib.Ingress, "voo-with-targets", "default")
	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--foo.com"}
	testlib.TeardownHostRule(t, g, sniVSKey, hrnameFoo)
	testlib.TeardownHostRule(t, g, sniVSKey, hrnameVoo)
	testlib.DeleteObject(t, lib.HTTPRule, rrnameFoo, "default")
	TearDownIngressForCacheSyncCheck(t, modelName)
}
