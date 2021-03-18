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

func TestHostnameCreateDeleteHostRuleForEvh(t *testing.T) {

	g := gomega.NewGomegaWithT(t)
	integrationtest.EnableEVH()
	defer integrationtest.DisableEVH()

	modelName := "admin/cluster--Shared-L7-EVH-0"
	hrname := "samplehr-foo"
	SetUpIngressForCacheSyncCheck(t, true, true, modelName)

	integrationtest.SetupHostRule(t, hrname, "foo.com", true)

	g.Eventually(func() string {
		hostrule, _ := CRDClient.AkoV1alpha1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 10*time.Second).Should(gomega.Equal("Accepted"))

	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--default-foo.com"}
	integrationtest.VerifyMetadataHostRule(g, sniVSKey, "default/samplehr-foo", true)
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(*nodes[0].EvhNodes[0].Enabled).To(gomega.Equal(true))
	g.Expect(nodes[0].EvhNodes[0].SSLKeyCertAviRef).To(gomega.ContainSubstring("thisisaviref-sslkey"))
	g.Expect(nodes[0].EvhNodes[0].WafPolicyRef).To(gomega.ContainSubstring("thisisaviref-waf"))
	g.Expect(nodes[0].EvhNodes[0].AppProfileRef).To(gomega.ContainSubstring("thisisaviref-appprof"))
	g.Expect(nodes[0].EvhNodes[0].AnalyticsProfileRef).To(gomega.ContainSubstring("thisisaviref-analyticsprof"))
	g.Expect(nodes[0].EvhNodes[0].ErrorPageProfileRef).To(gomega.ContainSubstring("thisisaviref-errorprof"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicySetRefs).To(gomega.HaveLen(2))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicySetRefs[0]).To(gomega.ContainSubstring("thisisaviref-httpps2"))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicySetRefs[1]).To(gomega.ContainSubstring("thisisaviref-httpps1"))
	g.Expect(nodes[0].EvhNodes[0].VsDatascriptRefs).To(gomega.HaveLen(2))
	g.Expect(nodes[0].EvhNodes[0].VsDatascriptRefs[0]).To(gomega.ContainSubstring("thisisaviref-ds2"))
	g.Expect(nodes[0].EvhNodes[0].VsDatascriptRefs[1]).To(gomega.ContainSubstring("thisisaviref-ds1"))
	g.Expect(nodes[0].EvhNodes[0].SSLProfileRef).To(gomega.ContainSubstring("thisisaviref-sslprof"))

	hrUpdate := integrationtest.FakeHostRule{
		Name:              hrname,
		Namespace:         "default",
		Fqdn:              "foo.com",
		SslKeyCertificate: "thisisaviref-sslkey",
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
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
			return *nodes[0].EvhNodes[0].Enabled
		}
		return true
	}, 25*time.Second).Should(gomega.Equal(false))

	integrationtest.TeardownHostRule(t, g, sniVSKey, hrname)
	integrationtest.VerifyMetadataHostRule(g, sniVSKey, "default/samplehr-foo", false)
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(nodes[0].EvhNodes[0].Enabled).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].SSLKeyCertAviRef).To(gomega.Equal(""))
	g.Expect(nodes[0].EvhNodes[0].WafPolicyRef).To(gomega.Equal(""))
	g.Expect(nodes[0].EvhNodes[0].AppProfileRef).To(gomega.Equal(""))
	g.Expect(nodes[0].EvhNodes[0].AnalyticsProfileRef).To(gomega.Equal(""))
	g.Expect(nodes[0].EvhNodes[0].ErrorPageProfileRef).To(gomega.Equal(""))
	g.Expect(nodes[0].EvhNodes[0].HttpPolicySetRefs).To(gomega.HaveLen(0))
	g.Expect(nodes[0].EvhNodes[0].VsDatascriptRefs).To(gomega.HaveLen(0))
	g.Expect(nodes[0].EvhNodes[0].SSLProfileRef).To(gomega.Equal(""))
	TearDownIngressForCacheSyncCheck(t, modelName)
}

func TestHostnameCreateHostRuleBeforeIngressForEvh(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	integrationtest.EnableEVH()
	defer integrationtest.DisableEVH()

	modelName := "admin/cluster--Shared-L7-EVH-0"
	hrname := "samplehr-foo"
	integrationtest.SetupHostRule(t, hrname, "foo.com", true)

	g.Eventually(func() string {
		hostrule, _ := CRDClient.AkoV1alpha1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 10*time.Second).Should(gomega.Equal("Accepted"))

	SetUpIngressForCacheSyncCheck(t, true, true, modelName)

	g.Eventually(func() string {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		if len(nodes[0].EvhNodes) == 1 {
			return nodes[0].EvhNodes[0].SSLKeyCertAviRef
		}
		return ""
	}, 10*time.Second).Should(gomega.ContainSubstring("thisisaviref-sslkey"))

	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--default-foo.com"}
	integrationtest.TeardownHostRule(t, g, sniVSKey, hrname)

	g.Eventually(func() string {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		if len(nodes[0].EvhNodes) == 1 {
			return nodes[0].EvhNodes[0].SSLKeyCertAviRef
		}
		return ""
	}, 10*time.Second).Should(gomega.Equal(""))
	TearDownIngressForCacheSyncCheck(t, modelName)
}

func TestHostnameGoodToBadHostRuleForEvh(t *testing.T) {
	// create insecure ingress, apply good secure hostrule, transition to bad
	g := gomega.NewGomegaWithT(t)
	integrationtest.EnableEVH()
	defer integrationtest.DisableEVH()

	modelName := "admin/cluster--Shared-L7-EVH-0"
	hrname := "samplehr-foo"
	SetUpIngressForCacheSyncCheck(t, false, false, modelName)
	integrationtest.SetupHostRule(t, hrname, "foo.com", true)

	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--default-foo.com"}
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
	}, 10*time.Second).Should(gomega.Equal("Rejected"))

	// the last applied hostrule values would exist
	g.Eventually(func() string {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return nodes[0].EvhNodes[0].SSLKeyCertAviRef
	}, 10*time.Second).Should(gomega.ContainSubstring("thisisaviref"))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(nodes[0].EvhNodes[0].WafPolicyRef).To(gomega.ContainSubstring("thisisaviref-waf"))
	g.Expect(nodes[0].EvhNodes[0].AppProfileRef).To(gomega.ContainSubstring("thisisaviref-appprof"))

	integrationtest.TeardownHostRule(t, g, sniVSKey, hrname)
	TearDownIngressForCacheSyncCheck(t, modelName)
}

func TestHostnameInsecureHostAndHostruleForEvh(t *testing.T) {
	// create insecure ingress, insecure hostrule, hostrule should be applied in case of EVH
	g := gomega.NewGomegaWithT(t)
	integrationtest.EnableEVH()
	defer integrationtest.DisableEVH()

	modelName := "admin/cluster--Shared-L7-EVH-0"
	hrname := "samplehr-foo"
	SetUpIngressForCacheSyncCheck(t, false, false, modelName)
	integrationtest.SetupHostRule(t, hrname, "foo.com", false)

	g.Eventually(func() int {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
		return len(nodes[0].EvhNodes)
	}, 10*time.Second).Should(gomega.Equal(1))
	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--default-foo.com"}
	integrationtest.VerifyMetadataHostRule(g, sniVSKey, "default/samplehr-foo", true)
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(nodes[0].EvhNodes).To(gomega.HaveLen(1))

	integrationtest.TeardownHostRule(t, g, sniVSKey, hrname)
	TearDownIngressForCacheSyncCheck(t, modelName)
}

// HttpRule tests

func TestHostnameHTTPRuleCreateDeleteForEvh(t *testing.T) {
	// ingress secure foo.com/foo /bar
	// create httprule /foo, nothing happens
	// create hostrule, httprule gets attached check on /foo /bar
	// delete hostrule, httprule gets detached
	g := gomega.NewGomegaWithT(t)
	integrationtest.EnableEVH()
	defer integrationtest.DisableEVH()

	modelName := "admin/cluster--Shared-L7-EVH-0"
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
			"my-secret": {"foo.com"},
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
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(nodes[0].EvhNodes[0].PoolRefs[0].LbAlgorithm).To(gomega.Equal("LB_ALGORITHM_CONSISTENT_HASH"))
	g.Expect(nodes[0].EvhNodes[0].PoolRefs[0].LbAlgorithmHash).To(gomega.Equal("LB_ALGORITHM_CONSISTENT_HASH_SOURCE_IP_ADDRESS"))
	g.Expect(nodes[0].EvhNodes[0].PoolRefs[0].SslProfileRef).To(gomega.ContainSubstring("thisisaviref-sslprofile"))
	g.Expect(nodes[0].EvhNodes[0].PoolRefs[0].PkiProfile.CACert).To(gomega.Equal("httprule-destinationCA"))
	g.Expect(nodes[0].EvhNodes[0].PoolRefs[0].HealthMonitors).To(gomega.HaveLen(2))
	g.Expect(nodes[0].EvhNodes[0].PoolRefs[0].HealthMonitors[0]).To(gomega.ContainSubstring("thisisaviref-hm2"))
	g.Expect(nodes[0].EvhNodes[0].PoolRefs[0].HealthMonitors[1]).To(gomega.ContainSubstring("thisisaviref-hm1"))

	// delete httprule deletes refs as well
	integrationtest.TeardownHTTPRule(t, rrname)
	integrationtest.VerifyMetadataHTTPRule(g, poolFooKey, "default/"+rrname, false)
	integrationtest.VerifyMetadataHTTPRule(g, poolBarKey, "default/"+rrname, false)
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviEvhVS()
	g.Expect(nodes[0].EvhNodes[0].PoolRefs[0].LbAlgorithm).To(gomega.Equal(""))
	g.Expect(nodes[0].EvhNodes[0].PoolRefs[0].SslProfileRef).To(gomega.Equal(""))
	g.Expect(nodes[0].EvhNodes[0].PoolRefs[0].PkiProfile).To(gomega.BeNil())
	g.Expect(nodes[0].EvhNodes[0].PoolRefs[0].HealthMonitors).To(gomega.HaveLen(0))

	TearDownIngressForCacheSyncCheck(t, modelName)
}
