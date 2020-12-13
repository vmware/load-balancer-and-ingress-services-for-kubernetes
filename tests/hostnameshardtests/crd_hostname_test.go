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

package hostnameshardtests

import (
	"context"
	"testing"
	"time"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	avinodes "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/integrationtest"

	"github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestHostnameCreateDeleteHostRule(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName := "admin/cluster--Shared-L7-0"
	hrname := "samplehr-foo"
	SetUpIngressForCacheSyncCheck(t, modelName, true, true)

	integrationtest.SetupHostRule(t, hrname, "foo.com", true)

	g.Eventually(func() string {
		hostrule, _ := CRDClient.AkoV1alpha1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 10*time.Second).Should(gomega.Equal("Accepted"))

	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--foo.com"}
	integrationtest.VerifyMetadataHostRule(g, sniVSKey, "default/samplehr-foo", true)
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(*nodes[0].SniNodes[0].Enabled).To(gomega.Equal(true))
	g.Expect(nodes[0].SniNodes[0].SSLKeyCertAviRef).To(gomega.ContainSubstring("thisisahostruleref-sslkey"))
	g.Expect(nodes[0].SniNodes[0].WafPolicyRef).To(gomega.ContainSubstring("thisisahostruleref-waf"))
	g.Expect(nodes[0].SniNodes[0].AppProfileRef).To(gomega.ContainSubstring("thisisahostruleref-appprof"))
	g.Expect(nodes[0].SniNodes[0].AnalyticsProfileRef).To(gomega.ContainSubstring("thisisahostruleref-analyticsprof"))
	g.Expect(nodes[0].SniNodes[0].ErrorPageProfileRef).To(gomega.ContainSubstring("thisisahostruleref-errorprof"))
	g.Expect(nodes[0].SniNodes[0].HttpPolicySetRefs).To(gomega.HaveLen(2))
	g.Expect(nodes[0].SniNodes[0].HttpPolicySetRefs[0]).To(gomega.ContainSubstring("thisisahostruleref-httpps2"))
	g.Expect(nodes[0].SniNodes[0].HttpPolicySetRefs[1]).To(gomega.ContainSubstring("thisisahostruleref-httpps1"))
	g.Expect(nodes[0].SniNodes[0].VsDatascriptRefs).To(gomega.HaveLen(2))
	g.Expect(nodes[0].SniNodes[0].VsDatascriptRefs[0]).To(gomega.ContainSubstring("thisisahostruleref-ds2"))
	g.Expect(nodes[0].SniNodes[0].VsDatascriptRefs[1]).To(gomega.ContainSubstring("thisisahostruleref-ds1"))
	g.Expect(nodes[0].SniNodes[0].SSLProfileRef).To(gomega.ContainSubstring("thisisahostruleref-sslprof"))

	hrUpdate := integrationtest.FakeHostRule{
		Name:              hrname,
		Namespace:         "default",
		Fqdn:              "foo.com",
		SslKeyCertificate: "thisisahostruleref-sslkey",
	}.HostRule()
	enableVirtualHost := false
	hrUpdate.Spec.VirtualHost.EnableVirtualHost = &enableVirtualHost
	hrUpdate.ResourceVersion = "2"
	_, err := CRDClient.AkoV1alpha1().HostRules("default").Update(context.TODO(), hrUpdate, metav1.UpdateOptions{})
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
	TearDownIngressForCacheSyncCheck(t, modelName)
}

func TestHostnameCreateHostRuleBeforeIngress(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName := "admin/cluster--Shared-L7-0"
	hrname := "samplehr-foo"
	integrationtest.SetupHostRule(t, hrname, "foo.com", true)

	g.Eventually(func() string {
		hostrule, _ := CRDClient.AkoV1alpha1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 10*time.Second).Should(gomega.Equal("Accepted"))

	SetUpIngressForCacheSyncCheck(t, modelName, true, true)

	g.Eventually(func() string {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		if len(nodes[0].SniNodes) == 1 {
			return nodes[0].SniNodes[0].SSLKeyCertAviRef
		}
		return ""
	}, 10*time.Second).Should(gomega.ContainSubstring("thisisahostruleref-sslkey"))

	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--foo.com"}
	integrationtest.TeardownHostRule(t, g, sniVSKey, hrname)

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

func TestHostnameInsecureToSecureHostRule(t *testing.T) {
	// insecure ingress to secure VS via Hostrule
	g := gomega.NewGomegaWithT(t)

	modelName := "admin/cluster--Shared-L7-0"
	hrname := "samplehr-foo"
	SetUpIngressForCacheSyncCheck(t, modelName, false, false)

	mcache := cache.SharedAviObjCache()
	vsKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--Shared-L7-0"}
	g.Eventually(func() int {
		vsCache, _ := mcache.VsCacheMeta.AviCacheGet(vsKey)
		vsCacheObj, _ := vsCache.(*cache.AviVsCache)
		return len(vsCacheObj.SNIChildCollection)
	}, 15*time.Second).Should(gomega.Equal(0))

	integrationtest.SetupHostRule(t, hrname, "foo.com", true)

	g.Eventually(func() int {
		vsCache, _ := mcache.VsCacheMeta.AviCacheGet(vsKey)
		vsCacheObj, _ := vsCache.(*cache.AviVsCache)
		return len(vsCacheObj.SNIChildCollection)
	}, 15*time.Second).Should(gomega.Equal(1))

	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--foo.com"}
	integrationtest.VerifyMetadataHostRule(g, sniVSKey, "default/samplehr-foo", true)

	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes[0].SniNodes[0].SSLKeyCertAviRef).To(gomega.ContainSubstring("thisisahostruleref-sslkey"))
	g.Expect(nodes[0].SniNodes[0].WafPolicyRef).To(gomega.ContainSubstring("thisisahostruleref-waf"))
	g.Expect(nodes[0].HttpPolicyRefs[0].RedirectPorts[0].StatusCode).To(gomega.Equal("HTTP_REDIRECT_STATUS_CODE_302"))

	integrationtest.TeardownHostRule(t, g, sniVSKey, hrname)
	TearDownIngressForCacheSyncCheck(t, modelName)
}

func TestHostnameMultiIngressToSecureHostRule(t *testing.T) {
	// 1 insecure ingress, 1 secure ingress -> secure VS via Hostrule
	g := gomega.NewGomegaWithT(t)

	modelName := "admin/cluster--Shared-L7-0"
	hrname := "samplehr-foo"

	// creating secure default/foo.com/foo
	SetUpIngressForCacheSyncCheck(t, modelName, true, true)

	// creating insecure red/foo.com/bar
	ingressObject := integrationtest.FakeIngress{
		Name:        "foo-with-targets-2",
		Namespace:   "red",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/bar"},
		ServiceName: "avisvc",
	}
	ingrFake := ingressObject.Ingress()
	if _, err := KubeClient.NetworkingV1beta1().Ingresses("red").Create(context.TODO(), ingrFake, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in updating Ingress: %v", err)
	}

	integrationtest.SetupHostRule(t, hrname, "foo.com", true)

	g.Eventually(func() int {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		if len(nodes[0].SniNodes) > 0 {
			return len(nodes[0].SniNodes[0].PoolGroupRefs)
		}
		return 0
	}, 50*time.Second).Should(gomega.Equal(2))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes[0].PoolRefs).To(gomega.HaveLen(0))
	g.Expect(nodes[0].SniNodes[0].PoolRefs).To(gomega.HaveLen(2))
	g.Expect(nodes[0].SniNodes[0].SSLKeyCertAviRef).To(gomega.ContainSubstring("thisisahostruleref-sslkey"))
	g.Expect(nodes[0].SniNodes[0].SSLKeyCertRefs).To(gomega.HaveLen(0))
	if len(nodes[0].HttpPolicyRefs) > 0 {
		g.Expect(nodes[0].HttpPolicyRefs[0].RedirectPorts[0].StatusCode).To(gomega.Equal("HTTP_REDIRECT_STATUS_CODE_302"))
	}

	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--foo.com"}
	integrationtest.VerifyMetadataHostRule(g, sniVSKey, "default/samplehr-foo", true)

	if err := KubeClient.NetworkingV1beta1().Ingresses("red").Delete(context.TODO(), "foo-with-targets-2", metav1.DeleteOptions{}); err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	integrationtest.TeardownHostRule(t, g, sniVSKey, hrname)
	TearDownIngressForCacheSyncCheck(t, modelName)
}

func TestHostnameMultiIngressSwitchHostRuleFqdn(t *testing.T) {
	// 2 insecure ingresses -> VS via Hostrule
	g := gomega.NewGomegaWithT(t)

	modelName := "admin/cluster--Shared-L7-0"
	hrname := "samplehr-foo"

	// creating insecure default/foo.com/foo
	SetUpIngressForCacheSyncCheck(t, modelName, false, false)

	// creating insecure red/voo.com/voo
	ingressObject := integrationtest.FakeIngress{
		Name:        "voo-with-targets",
		Namespace:   "red",
		DnsNames:    []string{"voo.com"},
		Paths:       []string{"/voo"},
		ServiceName: "avisvc",
	}
	ingrFake := ingressObject.Ingress()
	if _, err := KubeClient.NetworkingV1beta1().Ingresses("red").Create(context.TODO(), ingrFake, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in updating Ingress: %v", err)
	}

	// hostrule for foo.com
	integrationtest.SetupHostRule(t, hrname, "foo.com", true)

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
	hrUpdate := integrationtest.FakeHostRule{
		Name:              hrname,
		Namespace:         "default",
		Fqdn:              "voo.com",
		SslKeyCertificate: "thisisahostruleref-sslkey",
	}.HostRule()
	hrUpdate.ResourceVersion = "2"
	if _, err := CRDClient.AkoV1alpha1().HostRules("default").Update(context.TODO(), hrUpdate, metav1.UpdateOptions{}); err != nil {
		t.Fatalf("error in updating HostRule: %v", err)
	}

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

	if err := KubeClient.NetworkingV1beta1().Ingresses("red").Delete(context.TODO(), "voo-with-targets", metav1.DeleteOptions{}); err != nil {
		t.Fatalf("Couldn't Delete the Ingress %v", err)
	}
	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--voo.com"}
	integrationtest.TeardownHostRule(t, g, sniVSKey, hrname)
	TearDownIngressForCacheSyncCheck(t, modelName)
}

func TestHostnameGoodToBadHostRule(t *testing.T) {
	// create insecure ingress, apply good secure hostrule, transition to bad
	g := gomega.NewGomegaWithT(t)

	modelName := "admin/cluster--Shared-L7-0"
	hrname := "samplehr-foo"
	SetUpIngressForCacheSyncCheck(t, modelName, false, false)
	integrationtest.SetupHostRule(t, hrname, "foo.com", true)

	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--foo.com"}
	integrationtest.VerifyMetadataHostRule(g, sniVSKey, "default/samplehr-foo", true)

	// update hostrule with bad ref
	hrUpdate := integrationtest.FakeHostRule{
		Name:               hrname,
		Namespace:          "default",
		Fqdn:               "voo.com",
		WafPolicy:          "BADREF",
		ApplicationProfile: "thisisahostruleref-appprof",
	}.HostRule()
	hrUpdate.ResourceVersion = "2"
	if _, err := CRDClient.AkoV1alpha1().HostRules("default").Update(context.TODO(), hrUpdate, metav1.UpdateOptions{}); err != nil {
		t.Fatalf("error in updating HostRule: %v", err)
	}

	g.Eventually(func() string {
		hostrule, _ := CRDClient.AkoV1alpha1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 10*time.Second).Should(gomega.Equal("Rejected"))

	// the last applied hostrule values would exist
	g.Eventually(func() string {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return nodes[0].SniNodes[0].SSLKeyCertAviRef
	}, 10*time.Second).Should(gomega.ContainSubstring("thisisahostruleref"))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes[0].SniNodes[0].WafPolicyRef).To(gomega.ContainSubstring("thisisahostruleref-waf"))
	g.Expect(nodes[0].SniNodes[0].AppProfileRef).To(gomega.ContainSubstring("thisisahostruleref-appprof"))

	integrationtest.TeardownHostRule(t, g, sniVSKey, hrname)
	TearDownIngressForCacheSyncCheck(t, modelName)
}

func TestHostnameInsecureHostAndHostrule(t *testing.T) {
	// create insecure ingress, insecure hostrule, nothing should be applied
	g := gomega.NewGomegaWithT(t)

	modelName := "admin/cluster--Shared-L7-0"
	hrname := "samplehr-foo"
	SetUpIngressForCacheSyncCheck(t, modelName, false, false)
	integrationtest.SetupHostRule(t, hrname, "foo.com", false)

	g.Eventually(func() int {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return len(nodes[0].PoolRefs)
	}, 10*time.Second).Should(gomega.Equal(1))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes[0].SniNodes).To(gomega.HaveLen(0))

	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--foo.com"}
	integrationtest.TeardownHostRule(t, g, sniVSKey, hrname)
	TearDownIngressForCacheSyncCheck(t, modelName)
}

func TestHostnameValidToInvalidHostSwitch(t *testing.T) {
	// create insecure host foo.com, attach hostrule, change hostrule to non existing bar.com
	// foo.com should become insecure again
	// change hostrule back to foo.com and it should become secure again
	g := gomega.NewGomegaWithT(t)

	modelName := "admin/cluster--Shared-L7-0"
	hrname := "samplehr-foo"
	SetUpIngressForCacheSyncCheck(t, modelName, false, false)
	integrationtest.SetupHostRule(t, hrname, "foo.com", true)

	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--foo.com"}
	integrationtest.VerifyMetadataHostRule(g, sniVSKey, "default/samplehr-foo", true)

	hrUpdate := integrationtest.FakeHostRule{
		Name:              hrname,
		Namespace:         "default",
		Fqdn:              "bar.com",
		SslKeyCertificate: "thisisahostruleref-sslkey",
	}.HostRule()
	hrUpdate.ResourceVersion = "2"
	if _, err := CRDClient.AkoV1alpha1().HostRules("default").Update(context.TODO(), hrUpdate, metav1.UpdateOptions{}); err != nil {
		t.Fatalf("error in updating HostRule: %v", err)
	}

	g.Eventually(func() int {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return len(nodes[0].PoolRefs)
	}, 10*time.Second).Should(gomega.Equal(1))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes[0].PoolRefs[0].Name).To(gomega.Equal("cluster--foo.com_foo-default-foo-with-targets"))

	// change back to good host
	hrUpdate = integrationtest.FakeHostRule{
		Name:              hrname,
		Namespace:         "default",
		Fqdn:              "foo.com",
		SslKeyCertificate: "thisisahostruleref-sslkey",
	}.HostRule()
	hrUpdate.ResourceVersion = "3"
	if _, err := CRDClient.AkoV1alpha1().HostRules("default").Update(context.TODO(), hrUpdate, metav1.UpdateOptions{}); err != nil {
		t.Fatalf("error in updating HostRule: %v", err)
	}

	g.Eventually(func() int {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return len(nodes[0].PoolRefs)
	}, 10*time.Second).Should(gomega.Equal(0))
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes[0].SniNodes[0].PoolRefs[0].Name).To(gomega.Equal("cluster--default-foo.com_foo-foo-with-targets"))

	integrationtest.TeardownHostRule(t, g, sniVSKey, hrname)
	TearDownIngressForCacheSyncCheck(t, modelName)
}

// HttpRule tests

func TestHostnameHTTPRuleCreateDelete(t *testing.T) {
	// ingress secure foo.com/foo /bar
	// create httprule /foo, nothing happens
	// create hostrule, httprule gets attached check on /foo /bar
	// delete hostrule, httprule gets detached
	g := gomega.NewGomegaWithT(t)

	modelName := "admin/cluster--Shared-L7-0"
	rrname := "samplerr-foo"

	SetupDomain()
	SetUpTestForIngress(t, modelName)
	integrationtest.AddSecret("my-secret", "default", "tlsCert", "tlsKey")
	integrationtest.PollForCompletion(t, modelName, 5)
	ingressObject := integrationtest.FakeIngress{
		Name:        "foo-with-targets",
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		Paths:       []string{"/foo", "/bar"},
		ServiceName: "avisvc",
		TlsSecretDNS: map[string][]string{
			"my-secret": []string{"foo.com"},
		},
	}

	ingrFake := ingressObject.Ingress(true)
	if _, err := KubeClient.NetworkingV1beta1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	integrationtest.PollForCompletion(t, modelName, 5)

	poolFooKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--default-foo.com_foo-foo-with-targets"}
	poolBarKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--default-foo.com_bar-foo-with-targets"}
	integrationtest.SetupHTTPRule(t, rrname, "foo.com", "/")
	integrationtest.VerifyMetadataHTTPRule(g, poolFooKey, "default/"+rrname, true)
	integrationtest.VerifyMetadataHTTPRule(g, poolBarKey, "default/"+rrname, true)
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes[0].SniNodes[0].PoolRefs[0].LbAlgorithm).To(gomega.Equal("LB_ALGORITHM_CONSISTENT_HASH"))
	g.Expect(nodes[0].SniNodes[0].PoolRefs[0].LbAlgorithmHash).To(gomega.Equal("LB_ALGORITHM_CONSISTENT_HASH_SOURCE_IP_ADDRESS"))
	g.Expect(nodes[0].SniNodes[0].PoolRefs[0].SslProfileRef).To(gomega.ContainSubstring("thisisahttpruleref-sslprofile"))
	g.Expect(nodes[0].SniNodes[0].PoolRefs[0].PkiProfile.CACert).To(gomega.Equal("httprule-destinationCA"))
	g.Expect(nodes[0].SniNodes[0].PoolRefs[0].HealthMonitors).To(gomega.HaveLen(2))
	g.Expect(nodes[0].SniNodes[0].PoolRefs[0].HealthMonitors[0]).To(gomega.ContainSubstring("thisisahttpruleref-hm2"))
	g.Expect(nodes[0].SniNodes[0].PoolRefs[0].HealthMonitors[1]).To(gomega.ContainSubstring("thisisahttpruleref-hm1"))

	// delete httprule deletes refs as well
	integrationtest.TeardownHTTPRule(t, rrname)
	integrationtest.VerifyMetadataHTTPRule(g, poolFooKey, "default/"+rrname, false)
	integrationtest.VerifyMetadataHTTPRule(g, poolBarKey, "default/"+rrname, false)
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes[0].SniNodes[0].PoolRefs[0].LbAlgorithm).To(gomega.Equal(""))
	g.Expect(nodes[0].SniNodes[0].PoolRefs[0].SslProfileRef).To(gomega.Equal(""))
	g.Expect(nodes[0].SniNodes[0].PoolRefs[0].PkiProfile).To(gomega.BeNil())
	g.Expect(nodes[0].SniNodes[0].PoolRefs[0].HealthMonitors).To(gomega.HaveLen(0))

	TearDownIngressForCacheSyncCheck(t, modelName)
}

func TestHostNameHTTPRuleHostSwitch(t *testing.T) {
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
	SetUpIngressForCacheSyncCheck(t, modelName, false, false)

	// creates voo.com insecure
	ingressObject := integrationtest.FakeIngress{
		Name:        "voo-with-targets",
		Namespace:   "default",
		DnsNames:    []string{"voo.com"},
		Paths:       []string{"/foo"},
		ServiceName: "avisvc",
	}
	ingrFake := ingressObject.Ingress()
	if _, err := KubeClient.NetworkingV1beta1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in updating Ingress: %v", err)
	}

	integrationtest.SetupHTTPRule(t, rrnameFoo, "foo.com", "/foo")
	integrationtest.SetupHostRule(t, hrnameFoo, "foo.com", true) // makes foo.com secure
	integrationtest.SetupHostRule(t, hrnameVoo, "voo.com", false)

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

	if err := KubeClient.NetworkingV1beta1().Ingresses("default").Delete(context.TODO(), "voo-with-targets", metav1.DeleteOptions{}); err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--foo.com"}
	integrationtest.TeardownHostRule(t, g, sniVSKey, hrnameFoo)
	integrationtest.TeardownHostRule(t, g, sniVSKey, hrnameVoo)
	integrationtest.TeardownHTTPRule(t, rrnameFoo)
	TearDownIngressForCacheSyncCheck(t, modelName)
}
