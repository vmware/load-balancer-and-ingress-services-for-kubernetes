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
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	avinodes "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/apis/ako/v1beta1"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/integrationtest"

	"github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCreateDeleteHostRule(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName := MODEL_NAME_PREFIX + "0"
	svcName := objNameMap.GenerateName("avisvc")
	hrname := objNameMap.GenerateName("samplehr-foo")
	secretName := objNameMap.GenerateName("my-secret")
	ingName := objNameMap.GenerateName("foo-with-targets")
	ingTestObj := IngressTestObject{
		ingressName: ingName,
		isTLS:       true,
		withSecret:  true,
		secretName:  secretName,
		serviceName: svcName,
		modelNames:  []string{modelName},
	}
	ingTestObj.FillParams()
	SetUpIngressForCacheSyncCheck(t, ingTestObj)

	integrationtest.SetupHostRule(t, hrname, "foo.com", true)

	g.Eventually(func() string {
		hostrule, _ := v1beta1CRDClient.AkoV1beta1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 20*time.Second).Should(gomega.Equal("Accepted"))

	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--foo.com"}
	integrationtest.VerifyMetadataHostRule(t, g, sniVSKey, "default/"+hrname, true)
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(*nodes[0].SniNodes[0].Enabled).To(gomega.Equal(true))
	g.Expect(nodes[0].SniNodes[0].SslKeyAndCertificateRefs).To(gomega.HaveLen(1))
	g.Expect(nodes[0].SniNodes[0].SslKeyAndCertificateRefs[0]).To(gomega.ContainSubstring("thisisaviref-sslkey"))
	g.Expect(nodes[0].SniNodes[0].ICAPProfileRefs).To(gomega.HaveLen(1))
	g.Expect(nodes[0].SniNodes[0].ICAPProfileRefs[0]).To(gomega.ContainSubstring("thisisaviref-icapprof"))
	g.Expect(*nodes[0].SniNodes[0].WafPolicyRef).To(gomega.ContainSubstring("thisisaviref-waf"))
	g.Expect(*nodes[0].SniNodes[0].ApplicationProfileRef).To(gomega.ContainSubstring("thisisaviref-appprof"))
	g.Expect(*nodes[0].SniNodes[0].AnalyticsProfileRef).To(gomega.ContainSubstring("thisisaviref-analyticsprof"))
	g.Expect(nodes[0].SniNodes[0].ErrorPageProfileRef).To(gomega.ContainSubstring("thisisaviref-errorprof"))
	g.Expect(nodes[0].SniNodes[0].HttpPolicySetRefs).To(gomega.HaveLen(2))
	g.Expect(nodes[0].SniNodes[0].HttpPolicySetRefs[0]).To(gomega.ContainSubstring("thisisaviref-httpps2"))
	g.Expect(nodes[0].SniNodes[0].HttpPolicySetRefs[1]).To(gomega.ContainSubstring("thisisaviref-httpps1"))
	g.Expect(nodes[0].SniNodes[0].VsDatascriptRefs).To(gomega.HaveLen(2))
	g.Expect(nodes[0].SniNodes[0].VsDatascriptRefs[0]).To(gomega.ContainSubstring("thisisaviref-ds2"))
	g.Expect(nodes[0].SniNodes[0].VsDatascriptRefs[1]).To(gomega.ContainSubstring("thisisaviref-ds1"))
	g.Expect(*nodes[0].SniNodes[0].SslProfileRef).To(gomega.ContainSubstring("thisisaviref-sslprof"))
	g.Expect(nodes[0].NetworkSecurityPolicyRef).Should(gomega.BeNil())
	g.Expect(nodes[0].SniNodes[0].VHDomainNames).To(gomega.ContainElement("bar.com"))

	hrUpdate := integrationtest.FakeHostRule{
		Name:              hrname,
		Namespace:         "default",
		Fqdn:              "foo.com",
		SslKeyCertificate: "thisisaviref-sslkey",
	}.HostRule()
	enableVirtualHost := false
	hrUpdate.Spec.VirtualHost.EnableVirtualHost = &enableVirtualHost
	hrUpdate.Spec.VirtualHost.Gslb.Fqdn = "baz.com"
	hrUpdate.ResourceVersion = "2"
	_, err := v1beta1CRDClient.AkoV1beta1().HostRules("default").Update(context.TODO(), hrUpdate, metav1.UpdateOptions{})
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
	g.Expect(nodes[0].SniNodes[0].VHDomainNames).To(gomega.ContainElement("baz.com"))
	integrationtest.TeardownHostRule(t, g, sniVSKey, hrname)
	integrationtest.VerifyMetadataHostRule(t, g, sniVSKey, "default/"+hrname, false)
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes[0].SniNodes[0].Enabled).To(gomega.BeNil())
	g.Expect(nodes[0].SniNodes[0].SslKeyAndCertificateRefs).To(gomega.HaveLen(0))
	g.Expect(nodes[0].SniNodes[0].ICAPProfileRefs).To(gomega.HaveLen(0))
	g.Expect(nodes[0].SniNodes[0].WafPolicyRef).To(gomega.BeNil())
	g.Expect(nodes[0].SniNodes[0].ApplicationProfileRef).To(gomega.BeNil())
	g.Expect(nodes[0].SniNodes[0].AnalyticsProfileRef).To(gomega.BeNil())
	g.Expect(nodes[0].SniNodes[0].ErrorPageProfileRef).To(gomega.Equal(""))
	g.Expect(nodes[0].SniNodes[0].HttpPolicySetRefs).To(gomega.HaveLen(0))
	g.Expect(nodes[0].SniNodes[0].VsDatascriptRefs).To(gomega.HaveLen(0))
	g.Expect(nodes[0].SniNodes[0].SslProfileRef).To(gomega.BeNil())
	g.Expect(nodes[0].SniNodes[0].VHDomainNames).To(gomega.Not(gomega.ContainElement("baz.com")))

	TearDownIngressForCacheSyncCheck(t, ingName, svcName, secretName, modelName)
}

func TestCreateDeleteSharedVSHostRule(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName := MODEL_NAME_PREFIX + "0"
	svcName := objNameMap.GenerateName("avisvc")
	hrname := objNameMap.GenerateName("samplehr-foo")
	secretName := objNameMap.GenerateName("my-secret")
	ingName := objNameMap.GenerateName("foo-with-targets")
	ingTestObj := IngressTestObject{
		ingressName: ingName,
		isTLS:       true,
		withSecret:  true,
		secretName:  secretName,
		serviceName: svcName,
		modelNames:  []string{modelName},
	}
	ingTestObj.FillParams()
	SetUpIngressForCacheSyncCheck(t, ingTestObj)

	hostrule := integrationtest.FakeHostRule{
		Name:                  hrname,
		Namespace:             "default",
		Fqdn:                  "cluster--Shared-L7-0.admin.com",
		WafPolicy:             "thisisaviref-waf",
		ApplicationProfile:    "thisisaviref-appprof",
		AnalyticsProfile:      "thisisaviref-analyticsprof",
		ErrorPageProfile:      "thisisaviref-errorprof",
		Datascripts:           []string{"thisisaviref-ds2", "thisisaviref-ds1"},
		HttpPolicySets:        []string{"thisisaviref-httpps2", "thisisaviref-httpps1"},
		NetworkSecurityPolicy: "thisisaviref-networksecuritypolicyref",
	}
	hrCreate := hostrule.HostRule()
	hrCreate.Spec.VirtualHost.TCPSettings = &v1beta1.HostRuleTCPSettings{
		Listeners: []v1beta1.HostRuleTCPListeners{
			{Port: 8081}, {Port: 8082}, {Port: 8083, EnableSSL: true},
		},
	}
	if _, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().HostRules("default").Create(context.TODO(), hrCreate, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding HostRule: %v", err)
	}

	g.Eventually(func() string {
		hostrule, _ := v1beta1CRDClient.AkoV1beta1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 10*time.Second).Should(gomega.Equal("Accepted"))

	vsKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--Shared-L7-0"}
	integrationtest.VerifyMetadataHostRule(t, g, vsKey, "default/"+hrname, true)
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(*nodes[0].Enabled).To(gomega.Equal(true))
	g.Expect(*nodes[0].WafPolicyRef).To(gomega.ContainSubstring("thisisaviref-waf"))
	g.Expect(*nodes[0].ApplicationProfileRef).To(gomega.ContainSubstring("thisisaviref-appprof"))
	g.Expect(*nodes[0].AnalyticsProfileRef).To(gomega.ContainSubstring("thisisaviref-analyticsprof"))
	g.Expect(nodes[0].ErrorPageProfileRef).To(gomega.ContainSubstring("thisisaviref-errorprof"))
	g.Expect(nodes[0].HttpPolicySetRefs).To(gomega.HaveLen(2))
	g.Expect(nodes[0].HttpPolicySetRefs[0]).To(gomega.ContainSubstring("thisisaviref-httpps2"))
	g.Expect(nodes[0].HttpPolicySetRefs[1]).To(gomega.ContainSubstring("thisisaviref-httpps1"))
	g.Expect(nodes[0].VsDatascriptRefs).To(gomega.HaveLen(2))
	g.Expect(nodes[0].VsDatascriptRefs[0]).To(gomega.ContainSubstring("thisisaviref-ds2"))
	g.Expect(nodes[0].VsDatascriptRefs[1]).To(gomega.ContainSubstring("thisisaviref-ds1"))
	g.Expect(*nodes[0].NetworkSecurityPolicyRef).To(gomega.ContainSubstring("thisisaviref-networksecuritypolicyref"))
	g.Expect(nodes[0].PortProto).To(gomega.HaveLen(3))
	var ports []int
	for _, port := range nodes[0].PortProto {
		ports = append(ports, int(port.Port))
		if port.EnableSSL {
			g.Expect(int(port.Port)).Should(gomega.Equal(8083))
		}
	}
	sort.Ints(ports)
	g.Expect(ports[0]).To(gomega.Equal(8081))
	g.Expect(ports[1]).To(gomega.Equal(8082))
	g.Expect(ports[2]).To(gomega.Equal(8083))

	integrationtest.TeardownHostRule(t, g, vsKey, hrname)
	integrationtest.VerifyMetadataHostRule(t, g, vsKey, "default/"+hrname, false)
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes[0].Enabled).To(gomega.BeNil())
	g.Expect(nodes[0].SslKeyAndCertificateRefs).To(gomega.HaveLen(0))
	g.Expect(nodes[0].WafPolicyRef).To(gomega.BeNil())
	g.Expect(nodes[0].ApplicationProfileRef).To(gomega.BeNil())
	g.Expect(nodes[0].AnalyticsProfileRef).To(gomega.BeNil())
	g.Expect(nodes[0].ErrorPageProfileRef).To(gomega.Equal(""))
	g.Expect(nodes[0].HttpPolicySetRefs).To(gomega.HaveLen(0))
	g.Expect(nodes[0].VsDatascriptRefs).To(gomega.HaveLen(0))
	g.Expect(nodes[0].SslProfileRef).To(gomega.BeNil())
	g.Expect(nodes[0].NetworkSecurityPolicyRef).To(gomega.BeNil())
	ports = []int{}
	for _, port := range nodes[0].PortProto {
		ports = append(ports, int(port.Port))
		if port.EnableSSL {
			g.Expect(int(port.Port)).To(gomega.Equal(443))
		}
	}
	sort.Ints(ports)
	g.Expect(ports[0]).To(gomega.Equal(80))
	g.Expect(ports[1]).To(gomega.Equal(443))

	TearDownIngressForCacheSyncCheck(t, ingName, svcName, secretName, modelName)
}

func TestCreateHostRuleBeforeIngress(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName := MODEL_NAME_PREFIX + "0"
	svcName := objNameMap.GenerateName("avisvc")
	hrname := objNameMap.GenerateName("samplehr-foo")
	secretName := objNameMap.GenerateName("my-secret")
	ingName := objNameMap.GenerateName("foo-with-targets")
	integrationtest.SetupHostRule(t, hrname, "foo.com", true)

	g.Eventually(func() string {
		hostrule, err := lib.AKOControlConfig().CRDInformers().HostRuleInformer.Lister().HostRules("default").Get(hrname)
		if err != nil {
			return ""
		}
		return hostrule.Status.Status
	}, 30*time.Second, 1*time.Second).Should(gomega.Equal("Accepted"))

	ingTestObj := IngressTestObject{
		ingressName: ingName,
		isTLS:       true,
		withSecret:  true,
		secretName:  secretName,
		serviceName: svcName,
		modelNames:  []string{modelName},
	}
	ingTestObj.FillParams()

	SetUpIngressForCacheSyncCheck(t, ingTestObj)

	g.Eventually(func() string {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		if len(nodes[0].SniNodes) == 1 && len(nodes[0].SniNodes[0].SslKeyAndCertificateRefs) == 1 {
			return nodes[0].SniNodes[0].SslKeyAndCertificateRefs[0]
		}
		return ""
	}, 30*time.Second, 1*time.Second).Should(gomega.ContainSubstring("thisisaviref-sslkey"))

	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--foo.com"}
	integrationtest.TeardownHostRule(t, g, sniVSKey, hrname)

	g.Eventually(func() string {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		if len(nodes[0].SniNodes) == 1 && len(nodes[0].SniNodes[0].SslKeyAndCertificateRefs) == 1 {
			return nodes[0].SniNodes[0].SslKeyAndCertificateRefs[0]
		}
		return ""
	}, 30*time.Second, 1*time.Second).Should(gomega.Equal(""))
	TearDownIngressForCacheSyncCheck(t, ingName, svcName, secretName, modelName)
}

func TestInsecureToSecureHostRule(t *testing.T) {
	// insecure ingress to secure VS via Hostrule
	g := gomega.NewGomegaWithT(t)

	modelName := MODEL_NAME_PREFIX + "0"
	svcName := objNameMap.GenerateName("avisvc")
	hrname := objNameMap.GenerateName("samplehr-foo")
	ingName := objNameMap.GenerateName("foo-with-targets")

	ingTestObj := IngressTestObject{
		ingressName: ingName,
		isTLS:       false,
		withSecret:  false,
		serviceName: svcName,
		modelNames:  []string{modelName},
	}
	ingTestObj.FillParams()
	SetUpIngressForCacheSyncCheck(t, ingTestObj)

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
	integrationtest.VerifyMetadataHostRule(t, g, sniVSKey, "default/"+hrname, true)

	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes[0].SniNodes[0].SslKeyAndCertificateRefs).To(gomega.HaveLen(1))
	g.Expect(nodes[0].SniNodes[0].SslKeyAndCertificateRefs[0]).To(gomega.ContainSubstring("thisisaviref-sslkey"))
	g.Expect(*nodes[0].SniNodes[0].WafPolicyRef).To(gomega.ContainSubstring("thisisaviref-waf"))
	g.Expect(nodes[0].HttpPolicyRefs[0].RedirectPorts[0].StatusCode).To(gomega.Equal("HTTP_REDIRECT_STATUS_CODE_302"))

	integrationtest.TeardownHostRule(t, g, sniVSKey, hrname)
	TearDownIngressForCacheSyncCheck(t, ingName, svcName, "", modelName)
}

func TestGSLBHostRewriteRule(t *testing.T) {
	// insecure ingress to secure VS via Hostrule
	g := gomega.NewGomegaWithT(t)

	modelName := MODEL_NAME_PREFIX + "0"
	svcName := objNameMap.GenerateName("avisvc")
	hrname := objNameMap.GenerateName("samplehr-foo")
	ingName := objNameMap.GenerateName("foo-with-targets")

	ingTestObj := IngressTestObject{
		ingressName: ingName,
		isTLS:       false,
		withSecret:  false,
		serviceName: svcName,
		modelNames:  []string{modelName},
	}
	ingTestObj.FillParams()
	SetUpIngressForCacheSyncCheck(t, ingTestObj)

	mcache := cache.SharedAviObjCache()
	vsKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--Shared-L7-0"}
	g.Eventually(func() int {
		vsCache, _ := mcache.VsCacheMeta.AviCacheGet(vsKey)
		vsCacheObj, _ := vsCache.(*cache.AviVsCache)
		return len(vsCacheObj.SNIChildCollection)
	}, 30*time.Second).Should(gomega.Equal(0))

	integrationtest.SetupHostRule(t, hrname, "foo.com", false)
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
	integrationtest.SetupHostRule(t, hrname, "foo.com", false, "baz.com")

	g.Eventually(func() bool {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return nodes[0].GetHttpPolicyRefs()[0].HeaderReWrite.SourceHost == "baz.com"
	}, 50*time.Second).Should(gomega.Equal(true))

	integrationtest.TearDownHostRuleWithNoVerify(t, g, hrname)
	g.Eventually(func() int {
		vsCache, _ := mcache.VsCacheMeta.AviCacheGet(vsKey)
		vsCacheObj, _ := vsCache.(*cache.AviVsCache)
		return len(vsCacheObj.HTTPKeyCollection)
	}, 30*time.Second).Should(gomega.Equal(0))

	TearDownIngressForCacheSyncCheck(t, ingName, svcName, "", modelName)
}

func TestMultiIngressToSecureHostRule(t *testing.T) {
	// 1 insecure ingress, 1 secure ingress -> secure VS via Hostrule
	g := gomega.NewGomegaWithT(t)

	modelName := MODEL_NAME_PREFIX + "0"
	svcName := objNameMap.GenerateName("avisvc")
	hrname := objNameMap.GenerateName("samplehr-foo")
	secretName := objNameMap.GenerateName("my-secret")
	ingName := objNameMap.GenerateName("foo-with-targets")
	ingName2 := objNameMap.GenerateName("foo-with-targets")

	// creating secure default/foo.com/foo
	ingTestObj := IngressTestObject{
		ingressName: ingName,
		isTLS:       true,
		withSecret:  true,
		secretName:  secretName,
		serviceName: svcName,
		modelNames:  []string{modelName},
	}
	ingTestObj.FillParams()
	SetUpIngressForCacheSyncCheck(t, ingTestObj)

	// creating insecure red/foo.com/bar
	ingressObject := integrationtest.FakeIngress{
		Name:        ingName2,
		Namespace:   "red",
		DnsNames:    []string{"foo.com"},
		Paths:       []string{"/bar"},
		ServiceName: svcName,
	}
	ingrFake := ingressObject.Ingress()
	if _, err := KubeClient.NetworkingV1().Ingresses("red").Create(context.TODO(), ingrFake, metav1.CreateOptions{}); err != nil {
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
	VerifyPoolDeletionFromVsNode(g, modelName)

	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes[0].SniNodes[0].PoolRefs).To(gomega.HaveLen(2))
	g.Expect(nodes[0].SniNodes[0].SslKeyAndCertificateRefs).To(gomega.HaveLen(1))
	g.Expect(nodes[0].SniNodes[0].SslKeyAndCertificateRefs[0]).To(gomega.ContainSubstring("thisisaviref-sslkey"))
	g.Expect(nodes[0].SniNodes[0].SSLKeyCertRefs).To(gomega.HaveLen(0))
	if len(nodes[0].HttpPolicyRefs) > 0 {
		g.Expect(nodes[0].HttpPolicyRefs[0].RedirectPorts[0].StatusCode).To(gomega.Equal("HTTP_REDIRECT_STATUS_CODE_302"))
	}

	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--foo.com"}
	integrationtest.VerifyMetadataHostRule(t, g, sniVSKey, "default/"+hrname, true)

	if err := KubeClient.NetworkingV1().Ingresses("red").Delete(context.TODO(), ingName2, metav1.DeleteOptions{}); err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	integrationtest.TeardownHostRule(t, g, sniVSKey, hrname)
	TearDownIngressForCacheSyncCheck(t, ingName, svcName, secretName, modelName)
}

func TestMultiIngressSwitchHostRuleFqdn(t *testing.T) {
	// 2 insecure ingresses -> VS via Hostrule
	g := gomega.NewGomegaWithT(t)

	modelName := MODEL_NAME_PREFIX + "0"
	svcName := objNameMap.GenerateName("avisvc")
	hrname := objNameMap.GenerateName("samplehr-foo")
	ingName := objNameMap.GenerateName("foo-with-targets")

	// creating insecure default/foo.com/foo
	ingTestObj := IngressTestObject{
		ingressName: ingName,
		isTLS:       false,
		withSecret:  false,
		serviceName: svcName,
		modelNames:  []string{modelName},
	}
	ingTestObj.FillParams()
	SetUpIngressForCacheSyncCheck(t, ingTestObj)

	// creating insecure red/voo.com/voo
	ingressObject := integrationtest.FakeIngress{
		Name:        "voo-with-targets",
		Namespace:   "red",
		DnsNames:    []string{"voo.com"},
		Paths:       []string{"/voo"},
		ServiceName: svcName,
	}
	ingrFake := ingressObject.Ingress()
	if _, err := KubeClient.NetworkingV1().Ingresses("red").Create(context.TODO(), ingrFake, metav1.CreateOptions{}); err != nil {
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
	g.Expect(nodes[0].SniNodes[0].PoolRefs[0].Name).To(gomega.Equal("cluster--default-foo.com_foo-" + ingName))

	// change hostrule for foo.com to voo.com
	hrUpdate := integrationtest.FakeHostRule{
		Name:              hrname,
		Namespace:         "default",
		Fqdn:              "voo.com",
		SslKeyCertificate: "thisisaviref-sslkey",
	}.HostRule()
	hrUpdate.ResourceVersion = "2"
	if _, err := v1beta1CRDClient.AkoV1beta1().HostRules("default").Update(context.TODO(), hrUpdate, metav1.UpdateOptions{}); err != nil {
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
	g.Expect(nodes[0].PoolRefs[0].Name).To(gomega.Equal("cluster--foo.com_foo-default-" + ingName))
	g.Expect(nodes[0].SniNodes[0].PoolRefs[0].Name).To(gomega.Equal("cluster--red-voo.com_voo-voo-with-targets"))

	if err := KubeClient.NetworkingV1().Ingresses("red").Delete(context.TODO(), "voo-with-targets", metav1.DeleteOptions{}); err != nil {
		t.Fatalf("Couldn't Delete the Ingress %v", err)
	}
	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--voo.com"}
	integrationtest.TeardownHostRule(t, g, sniVSKey, hrname)
	TearDownIngressForCacheSyncCheck(t, ingName, svcName, "", modelName)
}

func TestGoodToBadHostRule(t *testing.T) {
	// create insecure ingress, apply good secure hostrule, transition to bad
	g := gomega.NewGomegaWithT(t)

	modelName := MODEL_NAME_PREFIX + "0"
	svcName := objNameMap.GenerateName("avisvc")
	hrname := objNameMap.GenerateName("samplehr-foo")
	ingName := objNameMap.GenerateName("foo-with-targets")
	ingTestObj := IngressTestObject{
		ingressName: ingName,
		isTLS:       false,
		withSecret:  false,
		serviceName: svcName,
		modelNames:  []string{modelName},
	}
	ingTestObj.FillParams()
	SetUpIngressForCacheSyncCheck(t, ingTestObj)
	integrationtest.SetupHostRule(t, hrname, "foo.com", true)

	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--foo.com"}
	integrationtest.VerifyMetadataHostRule(t, g, sniVSKey, "default/"+hrname, true)

	// update hostrule with bad ref
	hrUpdate := integrationtest.FakeHostRule{
		Name:               hrname,
		Namespace:          "default",
		Fqdn:               "voo.com",
		WafPolicy:          "thisisBADaviref",
		ApplicationProfile: "thisisaviref-appprof",
	}.HostRule()
	hrUpdate.ResourceVersion = "2"
	if _, err := v1beta1CRDClient.AkoV1beta1().HostRules("default").Update(context.TODO(), hrUpdate, metav1.UpdateOptions{}); err != nil {
		t.Fatalf("error in updating HostRule: %v", err)
	}

	g.Eventually(func() string {
		hostrule, _ := v1beta1CRDClient.AkoV1beta1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 10*time.Second).Should(gomega.Equal("Rejected"))

	// the last applied hostrule values would exist
	g.Eventually(func() string {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		if len(nodes[0].SniNodes[0].SslKeyAndCertificateRefs) == 1 {
			return nodes[0].SniNodes[0].SslKeyAndCertificateRefs[0]
		}
		return ""
	}, 10*time.Second).Should(gomega.ContainSubstring("thisisaviref"))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(*nodes[0].SniNodes[0].WafPolicyRef).To(gomega.ContainSubstring("thisisaviref-waf"))
	g.Expect(*nodes[0].SniNodes[0].ApplicationProfileRef).To(gomega.ContainSubstring("thisisaviref-appprof"))

	integrationtest.TeardownHostRule(t, g, sniVSKey, hrname)
	TearDownIngressForCacheSyncCheck(t, ingName, svcName, "", modelName)
}

func TestInsecureHostAndHostrule(t *testing.T) {
	// create insecure ingress, insecure hostrule, nothing should be applied
	g := gomega.NewGomegaWithT(t)

	modelName := MODEL_NAME_PREFIX + "0"
	svcName := objNameMap.GenerateName("avisvc")
	hrname := objNameMap.GenerateName("samplehr-foo")
	ingName := objNameMap.GenerateName("foo-with-targets")
	ingTestObj := IngressTestObject{
		ingressName: ingName,
		isTLS:       false,
		withSecret:  false,
		serviceName: svcName,
		modelNames:  []string{modelName},
	}
	ingTestObj.FillParams()
	SetUpIngressForCacheSyncCheck(t, ingTestObj)
	integrationtest.SetupHostRule(t, hrname, "foo.com", false)

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
	integrationtest.TeardownHostRule(t, g, sniVSKey, hrname)
	TearDownIngressForCacheSyncCheck(t, ingName, svcName, "", modelName)
}

func TestValidToInvalidHostSwitch(t *testing.T) {
	// create insecure host foo.com, attach hostrule, change hostrule to non existing bar.com
	// foo.com should become insecure again
	// change hostrule back to foo.com and it should become secure again
	g := gomega.NewGomegaWithT(t)

	modelName := MODEL_NAME_PREFIX + "0"
	svcName := objNameMap.GenerateName("avisvc")
	hrname := objNameMap.GenerateName("samplehr-foo")
	ingName := objNameMap.GenerateName("foo-with-targets")
	ingTestObj := IngressTestObject{
		ingressName: ingName,
		isTLS:       false,
		withSecret:  false,
		serviceName: svcName,
		modelNames:  []string{modelName},
	}
	ingTestObj.FillParams()
	SetUpIngressForCacheSyncCheck(t, ingTestObj)
	integrationtest.SetupHostRule(t, hrname, "foo.com", true)

	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--foo.com"}
	integrationtest.VerifyMetadataHostRule(t, g, sniVSKey, "default/"+hrname, true)

	hrUpdate := integrationtest.FakeHostRule{
		Name:              hrname,
		Namespace:         "default",
		Fqdn:              "bar.com",
		SslKeyCertificate: "thisisaviref-sslkey",
	}.HostRule()
	hrUpdate.ResourceVersion = "2"
	if _, err := v1beta1CRDClient.AkoV1beta1().HostRules("default").Update(context.TODO(), hrUpdate, metav1.UpdateOptions{}); err != nil {
		t.Fatalf("error in updating HostRule: %v", err)
	}

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
	g.Expect(nodes[0].PoolRefs[0].Name).To(gomega.Equal("cluster--foo.com_foo-default-" + ingName))

	// change back to good host
	hrUpdate = integrationtest.FakeHostRule{
		Name:              hrname,
		Namespace:         "default",
		Fqdn:              "foo.com",
		SslKeyCertificate: "thisisaviref-sslkey",
	}.HostRule()
	hrUpdate.ResourceVersion = "3"
	if _, err := v1beta1CRDClient.AkoV1beta1().HostRules("default").Update(context.TODO(), hrUpdate, metav1.UpdateOptions{}); err != nil {
		t.Fatalf("error in updating HostRule: %v", err)
	}

	VerifyPoolDeletionFromVsNode(g, modelName)
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes[0].SniNodes[0].PoolRefs[0].Name).To(gomega.Equal("cluster--default-foo.com_foo-" + ingName))

	integrationtest.TeardownHostRule(t, g, sniVSKey, hrname)
	TearDownIngressForCacheSyncCheck(t, ingName, svcName, "", modelName)
}

// This tc tests hostrule state if GSLB FQDN is same as that of Local FQDN/ Host.
func TestCreateHostRuleWithGSLBFqdn(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	hrname := objNameMap.GenerateName("samplehr-foo")
	integrationtest.SetupHostRule(t, hrname, "zoo.com", true)

	g.Eventually(func() string {
		hostrule, _ := v1beta1CRDClient.AkoV1beta1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 10*time.Second).Should(gomega.Equal("Accepted"))
	//Update hr
	integrationtest.SetupHostRule(t, hrname, "zoo.com", true, "zoo.com")
	g.Eventually(func() string {
		hostrule, _ := v1beta1CRDClient.AkoV1beta1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 10*time.Second).Should(gomega.Equal("Rejected"))
	integrationtest.TearDownHostRuleWithNoVerify(t, g, hrname)
}

func TestHostruleAnalyticsPolicyUpdate(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName := MODEL_NAME_PREFIX + "0"
	svcName := objNameMap.GenerateName("avisvc")
	hrname := objNameMap.GenerateName("samplehr-foo")
	secretName := objNameMap.GenerateName("my-secret")
	ingName := objNameMap.GenerateName("foo-with-targets")
	ingTestObj := IngressTestObject{
		ingressName: ingName,
		isTLS:       true,
		withSecret:  true,
		secretName:  secretName,
		serviceName: svcName,
		modelNames:  []string{modelName},
	}
	ingTestObj.FillParams()
	SetUpIngressForCacheSyncCheck(t, ingTestObj)

	integrationtest.SetupHostRule(t, hrname, "foo.com", true)

	g.Eventually(func() string {
		hostrule, _ := v1beta1CRDClient.AkoV1beta1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 10*time.Second).Should(gomega.Equal("Accepted"))

	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--foo.com"}
	integrationtest.VerifyMetadataHostRule(t, g, sniVSKey, "default/"+hrname, true)
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()

	// Check the default value of AnalyticsPolicy
	g.Expect(nodes[0].SniNodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].SniNodes[0].AnalyticsPolicy).To(gomega.BeNil())

	// Update host rule with AnalyticsPolicy - only LogAllHeaders
	hrUpdate := integrationtest.FakeHostRule{
		Name:      hrname,
		Namespace: "default",
		Fqdn:      "foo.com",
	}.HostRule()
	enabled := true
	analyticsPolicy := &v1beta1.HostRuleAnalyticsPolicy{
		LogAllHeaders: &enabled,
	}
	hrUpdate.Spec.VirtualHost.AnalyticsPolicy = analyticsPolicy
	hrUpdate.ResourceVersion = "2"
	_, err := v1beta1CRDClient.AkoV1beta1().HostRules("default").Update(context.TODO(), hrUpdate, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating HostRule: %v", err)
	}
	g.Eventually(func() string {
		hostrule, _ := v1beta1CRDClient.AkoV1beta1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 10*time.Second).Should(gomega.Equal("Accepted"))

	g.Eventually(func() bool {
		_, aviModel = objects.SharedAviGraphLister().Get(modelName)
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		if len(nodes[0].SniNodes) == 1 &&
			nodes[0].SniNodes[0].AnalyticsPolicy != nil {
			return *nodes[0].SniNodes[0].AnalyticsPolicy.AllHeaders
		}
		return false
	}, 10*time.Second).Should(gomega.Equal(true))

	// Update host rule with AnalyticsPolicy - only LogAllHeaders, FullClientLogs.Enabled.
	hrUpdate = integrationtest.FakeHostRule{
		Name:      hrname,
		Namespace: "default",
		Fqdn:      "foo.com",
	}.HostRule()
	enabled = true
	analyticsPolicy = &v1beta1.HostRuleAnalyticsPolicy{
		FullClientLogs: &v1beta1.FullClientLogs{
			Enabled: &enabled,
		},
		LogAllHeaders: &enabled,
	}
	hrUpdate.Spec.VirtualHost.AnalyticsPolicy = analyticsPolicy
	hrUpdate.ResourceVersion = "3"
	_, err = v1beta1CRDClient.AkoV1beta1().HostRules("default").Update(context.TODO(), hrUpdate, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating HostRule: %v", err)
	}
	g.Eventually(func() string {
		hostrule, _ := v1beta1CRDClient.AkoV1beta1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 10*time.Second).Should(gomega.Equal("Accepted"))

	g.Eventually(func() bool {
		_, aviModel = objects.SharedAviGraphLister().Get(modelName)
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		if len(nodes[0].SniNodes) == 1 && nodes[0].SniNodes[0].AnalyticsPolicy != nil &&
			nodes[0].SniNodes[0].AnalyticsPolicy.FullClientLogs != nil &&
			nodes[0].SniNodes[0].AnalyticsPolicy.FullClientLogs.Enabled != nil {
			return *nodes[0].SniNodes[0].AnalyticsPolicy.AllHeaders && *nodes[0].SniNodes[0].AnalyticsPolicy.FullClientLogs.Enabled
		}
		return false
	}, 10*time.Second).Should(gomega.Equal(true))

	// Update host rule with AnalyticsPolicy - with all fields
	hrUpdate = integrationtest.FakeHostRule{
		Name:      hrname,
		Namespace: "default",
		Fqdn:      "foo.com",
	}.HostRule()
	enabled = true
	analyticsPolicy = &v1beta1.HostRuleAnalyticsPolicy{
		FullClientLogs: &v1beta1.FullClientLogs{
			Enabled:  &enabled,
			Throttle: "LOW",
		},
		LogAllHeaders: &enabled,
	}
	hrUpdate.Spec.VirtualHost.AnalyticsPolicy = analyticsPolicy
	hrUpdate.ResourceVersion = "4"
	_, err = v1beta1CRDClient.AkoV1beta1().HostRules("default").Update(context.TODO(), hrUpdate, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating HostRule: %v", err)
	}
	g.Eventually(func() string {
		hostrule, _ := v1beta1CRDClient.AkoV1beta1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 10*time.Second).Should(gomega.Equal("Accepted"))

	g.Eventually(func() bool {
		_, aviModel = objects.SharedAviGraphLister().Get(modelName)
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		if len(nodes[0].SniNodes) == 1 && nodes[0].SniNodes[0].AnalyticsPolicy != nil &&
			nodes[0].SniNodes[0].AnalyticsPolicy.FullClientLogs != nil &&
			nodes[0].SniNodes[0].AnalyticsPolicy.FullClientLogs.Throttle != nil {
			return *nodes[0].SniNodes[0].AnalyticsPolicy.FullClientLogs.Throttle == *lib.GetThrottle("LOW")
		}
		return false
	}, 10*time.Second).Should(gomega.Equal(true))

	g.Expect(*nodes[0].SniNodes[0].AnalyticsPolicy.AllHeaders).To(gomega.BeTrue())
	g.Expect(*nodes[0].SniNodes[0].AnalyticsPolicy.FullClientLogs.Enabled).To(gomega.BeTrue())

	// Remove the analyticPolicy and check whether values are removed from VS
	hrUpdate.Spec.VirtualHost.AnalyticsPolicy = nil
	hrUpdate.ResourceVersion = "5"
	_, err = v1beta1CRDClient.AkoV1beta1().HostRules("default").Update(context.TODO(), hrUpdate, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating HostRule: %v", err)
	}
	g.Eventually(func() string {
		hostrule, _ := v1beta1CRDClient.AkoV1beta1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 10*time.Second).Should(gomega.Equal("Accepted"))

	g.Eventually(func() bool {
		_, aviModel = objects.SharedAviGraphLister().Get(modelName)
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		if len(nodes[0].SniNodes) == 1 {
			return nodes[0].SniNodes[0].AnalyticsPolicy == nil
		}
		return false
	}, 10*time.Second).Should(gomega.Equal(true))

	integrationtest.TeardownHostRule(t, g, sniVSKey, hrname)
	TearDownIngressForCacheSyncCheck(t, ingName, svcName, secretName, modelName)
}

func TestHostruleFQDNAliases(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName := MODEL_NAME_PREFIX + "0"
	svcName := objNameMap.GenerateName("avisvc")
	hrname := objNameMap.GenerateName("samplehr-foo")
	secretName := objNameMap.GenerateName("my-secret")
	ingName := objNameMap.GenerateName("foo-with-targets")
	ingTestObj := IngressTestObject{
		ingressName: ingName,
		isTLS:       true,
		withSecret:  true,
		secretName:  secretName,
		serviceName: svcName,
		modelNames:  []string{modelName},
	}
	ingTestObj.FillParams()
	SetUpIngressForCacheSyncCheck(t, ingTestObj)
	integrationtest.SetupHostRule(t, hrname, "foo.com", false)
	g.Eventually(func() int {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return len(nodes[0].SniNodes)
	}, 10*time.Second).Should(gomega.Equal(1))
	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: lib.Encode("cluster--foo.com", lib.EVHVS)}
	integrationtest.VerifyMetadataHostRule(t, g, sniVSKey, "default/"+hrname, true)
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()

	// Common function that takes care of all validations
	validateNode := func(node *avinodes.AviVsNode, aliases []string) {
		g.Expect(node.VSVIPRefs).To(gomega.HaveLen(1))
		g.Expect(node.VSVIPRefs[0].FQDNs).Should(gomega.ContainElements(aliases))
		g.Expect(node.HttpPolicyRefs).To(gomega.HaveLen(1))
		g.Expect(node.HttpPolicyRefs[0].RedirectPorts).To(gomega.HaveLen(1))
		g.Expect(node.HttpPolicyRefs[0].RedirectPorts[0].Hosts).Should(gomega.ContainElements(aliases))

		g.Expect(node.SniNodes).To(gomega.HaveLen(1))
		g.Expect(node.SniNodes[0].VHDomainNames).Should(gomega.ContainElements(aliases))
		g.Expect(node.SniNodes[0].HttpPolicyRefs).To(gomega.HaveLen(1))
		g.Expect(node.SniNodes[0].HttpPolicyRefs[0].HppMap).To(gomega.HaveLen(1))
		g.Expect(node.SniNodes[0].HttpPolicyRefs[0].HppMap[0].Host).Should(gomega.ContainElements(aliases))
		g.Expect(node.SniNodes[0].HttpPolicyRefs[0].AviMarkers.Host).Should(gomega.ContainElements(aliases))
	}

	// Check default values.
	validateNode(nodes[0], []string{"foo.com"})

	// Update host rule with a valid FQDN Aliases
	hrUpdate := integrationtest.FakeHostRule{
		Name:      hrname,
		Namespace: "default",
		Fqdn:      "foo.com",
	}.HostRule()
	aliases := []string{"alias1.com", "alias2.com"}
	hrUpdate.Spec.VirtualHost.FqdnType = v1beta1.Exact
	hrUpdate.Spec.VirtualHost.Aliases = aliases
	hrUpdate.ResourceVersion = "2"
	_, err := v1beta1CRDClient.AkoV1beta1().HostRules("default").Update(context.TODO(), hrUpdate, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating HostRule: %v", err)
	}
	g.Eventually(func() string {
		hostrule, _ := v1beta1CRDClient.AkoV1beta1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 10*time.Second).Should(gomega.Equal("Accepted"))

	g.Eventually(func() int {
		_, aviModel = objects.SharedAviGraphLister().Get(modelName)
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return len(nodes[0].SniNodes)
	}, 10*time.Second).Should(gomega.Equal(1))

	// update is not getting reflected on evh nodes immediately. Hence adding a sleep of 5 seconds.
	time.Sleep(5 * time.Second)

	// Check whether the Aliases are properly added to Parent and Child VSes.
	validateNode(nodes[0], aliases)

	// Append one more alias and check whether it is getting added to parent and child VS.
	aliases = append(aliases, "alias3.com")
	hrUpdate.Spec.VirtualHost.Aliases = aliases
	hrUpdate.ResourceVersion = "3"
	_, err = v1beta1CRDClient.AkoV1beta1().HostRules("default").Update(context.TODO(), hrUpdate, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating HostRule: %v", err)
	}
	g.Eventually(func() string {
		hostrule, _ := v1beta1CRDClient.AkoV1beta1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 10*time.Second).Should(gomega.Equal("Accepted"))

	g.Eventually(func() int {
		_, aviModel = objects.SharedAviGraphLister().Get(modelName)
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return len(nodes[0].SniNodes)
	}, 10*time.Second).Should(gomega.Equal(1))

	// update is not getting reflected on evh nodes immediately. Hence adding a sleep of 5 seconds.
	time.Sleep(5 * time.Second)

	// Check whether the Aliases are properly added to Parent and Child VSes.
	validateNode(nodes[0], aliases)

	// Remove one alias from hostrule and check whether its reference is removed properly.
	aliases = aliases[1:]
	hrUpdate.Spec.VirtualHost.Aliases = aliases
	hrUpdate.ResourceVersion = "4"
	_, err = v1beta1CRDClient.AkoV1beta1().HostRules("default").Update(context.TODO(), hrUpdate, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating HostRule: %v", err)
	}
	g.Eventually(func() string {
		hostrule, _ := v1beta1CRDClient.AkoV1beta1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 10*time.Second).Should(gomega.Equal("Accepted"))

	g.Eventually(func() int {
		_, aviModel = objects.SharedAviGraphLister().Get(modelName)
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return len(nodes[0].SniNodes)
	}, 10*time.Second).Should(gomega.Equal(1))

	// update is not getting reflected on evh nodes immediately. Hence adding a sleep of 5 seconds.
	time.Sleep(5 * time.Second)

	// Check whether the Alias reference is properly removed from Parent and Child VSes.
	validateNode(nodes[0], aliases)

	integrationtest.TeardownHostRule(t, g, sniVSKey, hrname)
	TearDownIngressForCacheSyncCheck(t, ingName, svcName, secretName, modelName)
}

func TestValidationsOfHostruleFQDNAliases(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName := MODEL_NAME_PREFIX + "0"
	svcName := objNameMap.GenerateName("avisvc")
	hrname := objNameMap.GenerateName("samplehr-foo")
	secretName := objNameMap.GenerateName("my-secret")
	ingName := objNameMap.GenerateName("foo-with-targets")
	ingTestObj := IngressTestObject{
		ingressName: ingName,
		isTLS:       true,
		withSecret:  true,
		secretName:  secretName,
		serviceName: svcName,
		modelNames:  []string{modelName},
	}
	ingTestObj.FillParams()
	SetUpIngressForCacheSyncCheck(t, ingTestObj)
	integrationtest.SetupHostRule(t, hrname, "foo.com", false)
	g.Eventually(func() int {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return len(nodes[0].SniNodes)
	}, 10*time.Second).Should(gomega.Equal(1))
	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: lib.Encode("cluster--foo.com", lib.EVHVS)}
	integrationtest.VerifyMetadataHostRule(t, g, sniVSKey, "default/"+hrname, true)
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()

	hosts := []string{"foo.com"}
	// Check default values.
	g.Expect(nodes[0].VSVIPRefs).To(gomega.HaveLen(1))
	g.Expect(nodes[0].VSVIPRefs[0].FQDNs).Should(gomega.ContainElements())
	g.Expect(nodes[0].HttpPolicyRefs).To(gomega.HaveLen(1))
	g.Expect(nodes[0].HttpPolicyRefs[0].RedirectPorts).To(gomega.HaveLen(1))
	g.Expect(nodes[0].HttpPolicyRefs[0].RedirectPorts[0].Hosts).Should(gomega.ContainElements(hosts))

	g.Expect(nodes[0].SniNodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].SniNodes[0].VHDomainNames).Should(gomega.ContainElements(hosts))
	g.Expect(nodes[0].SniNodes[0].HttpPolicyRefs).To(gomega.HaveLen(1))
	g.Expect(nodes[0].SniNodes[0].HttpPolicyRefs[0].HppMap).To(gomega.HaveLen(1))
	g.Expect(nodes[0].SniNodes[0].HttpPolicyRefs[0].HppMap[0].Host).Should(gomega.ContainElements(hosts))

	// Update host rule with duplicate Aliases
	hrUpdate := integrationtest.FakeHostRule{
		Name:      hrname,
		Namespace: "default",
		Fqdn:      "foo.com",
		GslbFqdn:  "bar.com",
	}.HostRule()
	aliases := []string{"alias1.com", "alias1.com"}
	hrUpdate.Spec.VirtualHost.FqdnType = v1beta1.Exact
	hrUpdate.Spec.VirtualHost.Aliases = aliases
	hrUpdate.ResourceVersion = "2"
	_, err := v1beta1CRDClient.AkoV1beta1().HostRules("default").Update(context.TODO(), hrUpdate, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating HostRule: %v", err)
	}
	g.Eventually(func() string {
		hostrule, _ := v1beta1CRDClient.AkoV1beta1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 10*time.Second).Should(gomega.Equal("Rejected"))

	// Update host rule with aliases that contains FQDN
	aliases = []string{"foo.com", "alias1.com"}
	hrUpdate.Spec.VirtualHost.Aliases = aliases
	hrUpdate.ResourceVersion = "3"
	_, err = v1beta1CRDClient.AkoV1beta1().HostRules("default").Update(context.TODO(), hrUpdate, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating HostRule: %v", err)
	}
	g.Eventually(func() string {
		hostrule, _ := v1beta1CRDClient.AkoV1beta1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 10*time.Second).Should(gomega.Equal("Rejected"))

	// Update host rule with aliases that contains GSLB FQDN
	aliases = []string{"bar.com", "alias1.com"}
	hrUpdate.Spec.VirtualHost.Aliases = aliases
	hrUpdate.ResourceVersion = "4"
	_, err = v1beta1CRDClient.AkoV1beta1().HostRules("default").Update(context.TODO(), hrUpdate, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating HostRule: %v", err)
	}
	g.Eventually(func() string {
		hostrule, _ := v1beta1CRDClient.AkoV1beta1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 10*time.Second).Should(gomega.Equal("Rejected"))

	// Update host rule with fqdn type other than Exact
	aliases = []string{"bar.com", "alias1.com"}
	hrUpdate.Spec.VirtualHost.FqdnType = v1beta1.Contains
	hrUpdate.Spec.VirtualHost.Aliases = aliases
	hrUpdate.ResourceVersion = "5"
	_, err = v1beta1CRDClient.AkoV1beta1().HostRules("default").Update(context.TODO(), hrUpdate, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating HostRule: %v", err)
	}
	g.Eventually(func() string {
		hostrule, _ := v1beta1CRDClient.AkoV1beta1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 10*time.Second).Should(gomega.Equal("Rejected"))

	// Create another host rule with same Aliases
	newHostRule := integrationtest.FakeHostRule{
		Name:      "new-" + hrname,
		Namespace: "default",
		Fqdn:      "baz.com",
	}.HostRule()
	newHostRule.Spec.VirtualHost.Aliases = aliases
	_, err = v1beta1CRDClient.AkoV1beta1().HostRules("default").Create(context.TODO(), newHostRule, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in updating HostRule: %v", err)
	}

	// creation of hostrule is taking some time. Hence adding a sleep of 5 seconds.
	time.Sleep(5 * time.Second)

	g.Eventually(func() string {
		hostrule, _ := v1beta1CRDClient.AkoV1beta1().HostRules("default").Get(context.TODO(), newHostRule.Name, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 10*time.Second).Should(gomega.Equal("Rejected"))

	integrationtest.TeardownHostRule(t, g, sniVSKey, newHostRule.Name)
	integrationtest.TeardownHostRule(t, g, sniVSKey, hrname)
	TearDownIngressForCacheSyncCheck(t, ingName, svcName, secretName, modelName)
}

func TestHostruleFQDNAliasesForMultiPathIngress(t *testing.T) {

	g := gomega.NewGomegaWithT(t)

	modelName := MODEL_NAME_PREFIX + "0"
	svcName := objNameMap.GenerateName("avisvc")
	hrname := objNameMap.GenerateName("samplehr-foo")
	secretName := objNameMap.GenerateName("my-secret")
	ingName := objNameMap.GenerateName("foo-with-targets")

	SetupDomain()
	SetUpTestForIngress(t, svcName, modelName)
	integrationtest.AddSecret(secretName, "default", "tlsCert", "tlsKey")
	integrationtest.PollForCompletion(t, modelName, 5)
	ingressObject := integrationtest.FakeIngress{
		Name:        ingName,
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Ips:         []string{"10.0.0.1"},
		HostNames:   []string{"v1"},
		Paths:       []string{"/foo", "/bar"},
		ServiceName: svcName,
		TlsSecretDNS: map[string][]string{
			secretName: {"foo.com"},
		},
	}

	ingrFake := ingressObject.Ingress(true)
	if _, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	integrationtest.PollForCompletion(t, modelName, 5)

	// Create the hostrule with a valid FQDN Aliases
	hrUpdate := integrationtest.FakeHostRule{
		Name:      hrname,
		Namespace: "default",
		Fqdn:      "foo.com",
	}.HostRule()
	aliases := []string{"alias1.foo.com", "alias2.foo.com"}
	hrUpdate.Spec.VirtualHost.FqdnType = v1beta1.Exact
	hrUpdate.Spec.VirtualHost.Aliases = aliases
	hrUpdate.ResourceVersion = "2"
	_, err := v1beta1CRDClient.AkoV1beta1().HostRules("default").Create(context.TODO(), hrUpdate, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in updating HostRule: %v", err)
	}
	g.Eventually(func() string {
		hostrule, _ := v1beta1CRDClient.AkoV1beta1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 10*time.Second).Should(gomega.Equal("Accepted"))

	// Waiting for the hostrule to get applied.
	time.Sleep(10 * time.Second)

	g.Eventually(func() int {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return len(nodes[0].SniNodes)
	}, 30*time.Second).Should(gomega.Equal(1))

	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()

	g.Expect(nodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].HttpPolicyRefs).To(gomega.HaveLen(1))
	g.Expect(nodes[0].HttpPolicyRefs[0].RedirectPorts).To(gomega.HaveLen(1))
	g.Expect(nodes[0].HttpPolicyRefs[0].RedirectPorts[0].Hosts).To(gomega.HaveLen(len(aliases) + 1)) // aliases + host
	g.Expect(nodes[0].HttpPolicyRefs[0].RedirectPorts[0].Hosts).Should(gomega.ContainElements(aliases))
	g.Expect(nodes[0].SniNodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].SniNodes[0].VHDomainNames).Should(gomega.ContainElements(aliases))
	g.Expect(nodes[0].SniNodes[0].AviMarkers).ShouldNot(gomega.BeNil())
	g.Expect(nodes[0].SniNodes[0].AviMarkers.Host).To(gomega.HaveLen(len(aliases) + 1)) // aliases + host
	g.Expect(nodes[0].SniNodes[0].AviMarkers.Host).Should(gomega.ContainElements(aliases))
	for _, httpPolicyRef := range nodes[0].SniNodes[0].HttpPolicyRefs {
		if httpPolicyRef.HppMap != nil {
			g.Expect(httpPolicyRef.HppMap).To(gomega.HaveLen(2))
			g.Expect(httpPolicyRef.HppMap[0].Host).Should(gomega.ContainElements(aliases))
			g.Expect(httpPolicyRef.HppMap[1].Host).Should(gomega.ContainElements(aliases))
		}
		g.Expect(httpPolicyRef.AviMarkers.Host).To(gomega.HaveLen(len(aliases) + 1)) // aliases + host
		g.Expect(httpPolicyRef.AviMarkers.Host).Should(gomega.ContainElements(aliases))
	}

	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: lib.Encode("cluster--foo.com", lib.EVHVS)}
	integrationtest.TeardownHostRule(t, g, sniVSKey, hrname)
	TearDownIngressForCacheSyncCheck(t, ingName, svcName, secretName, modelName)
}

func TestApplyHostruleToParentVS(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName := MODEL_NAME_PREFIX + "0"
	svcName := objNameMap.GenerateName("avisvc")
	hrname := objNameMap.GenerateName("samplehr-foo")
	secretName := objNameMap.GenerateName("my-secret")
	ingName := objNameMap.GenerateName("foo-with-targets")
	ingTestObj := IngressTestObject{
		ingressName: ingName,
		isTLS:       true,
		withSecret:  true,
		secretName:  secretName,
		serviceName: svcName,
		modelNames:  []string{modelName},
	}
	ingTestObj.FillParams()
	SetUpIngressForCacheSyncCheck(t, ingTestObj)

	hostrule := integrationtest.FakeHostRule{
		Name:               hrname,
		Namespace:          "default",
		WafPolicy:          "thisisaviref-waf",
		ApplicationProfile: "thisisaviref-appprof",
		AnalyticsProfile:   "thisisaviref-analyticsprof",
		ErrorPageProfile:   "thisisaviref-errorprof",
		Datascripts:        []string{"thisisaviref-ds2", "thisisaviref-ds1"},
		HttpPolicySets:     []string{"thisisaviref-httpps2", "thisisaviref-httpps1"},
	}
	hrObj := hostrule.HostRule()
	hrObj.Spec.VirtualHost.Fqdn = "Shared-L7-0"
	hrObj.Spec.VirtualHost.FqdnType = v1beta1.Contains

	if _, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().HostRules("default").Create(context.TODO(), hrObj, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding HostRule: %v", err)
	}

	g.Eventually(func() string {
		hostrule, _ := v1beta1CRDClient.AkoV1beta1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 30*time.Second).Should(gomega.Equal("Accepted"))

	vsKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--Shared-L7-0"}
	integrationtest.VerifyMetadataHostRule(t, g, vsKey, "default/"+hrname, true)
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(*nodes[0].Enabled).To(gomega.Equal(true))
	g.Expect(*nodes[0].WafPolicyRef).To(gomega.ContainSubstring("thisisaviref-waf"))
	g.Expect(*nodes[0].ApplicationProfileRef).To(gomega.ContainSubstring("thisisaviref-appprof"))
	g.Expect(*nodes[0].AnalyticsProfileRef).To(gomega.ContainSubstring("thisisaviref-analyticsprof"))
	g.Expect(nodes[0].ErrorPageProfileRef).To(gomega.ContainSubstring("thisisaviref-errorprof"))
	g.Expect(nodes[0].HttpPolicySetRefs).To(gomega.HaveLen(2))
	g.Expect(nodes[0].HttpPolicySetRefs[0]).To(gomega.ContainSubstring("thisisaviref-httpps2"))
	g.Expect(nodes[0].HttpPolicySetRefs[1]).To(gomega.ContainSubstring("thisisaviref-httpps1"))
	g.Expect(nodes[0].VsDatascriptRefs).To(gomega.HaveLen(2))
	g.Expect(nodes[0].VsDatascriptRefs[0]).To(gomega.ContainSubstring("thisisaviref-ds2"))
	g.Expect(nodes[0].VsDatascriptRefs[1]).To(gomega.ContainSubstring("thisisaviref-ds1"))

	integrationtest.TeardownHostRule(t, g, vsKey, hrname)
	integrationtest.VerifyMetadataHostRule(t, g, vsKey, "default/"+hrname, false)
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes[0].Enabled).To(gomega.BeNil())
	g.Expect(nodes[0].SslKeyAndCertificateRefs).To(gomega.HaveLen(0))
	g.Expect(nodes[0].WafPolicyRef).To(gomega.BeNil())
	g.Expect(nodes[0].ApplicationProfileRef).To(gomega.BeNil())
	g.Expect(nodes[0].AnalyticsProfileRef).To(gomega.BeNil())
	g.Expect(nodes[0].ErrorPageProfileRef).To(gomega.Equal(""))
	g.Expect(nodes[0].HttpPolicySetRefs).To(gomega.HaveLen(0))
	g.Expect(nodes[0].VsDatascriptRefs).To(gomega.HaveLen(0))
	g.Expect(nodes[0].SslProfileRef).To(gomega.BeNil())

	TearDownIngressForCacheSyncCheck(t, ingName, svcName, secretName, modelName)
}

func TestHostRuleWithEmptyConfig(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName := MODEL_NAME_PREFIX + "0"
	svcName := objNameMap.GenerateName("avisvc")
	hrname := objNameMap.GenerateName("samplehr-foo")
	secretName := objNameMap.GenerateName("my-secret")
	ingName := objNameMap.GenerateName("foo-with-targets")
	ingTestObj := IngressTestObject{
		ingressName: ingName,
		isTLS:       true,
		withSecret:  true,
		secretName:  secretName,
		serviceName: svcName,
		modelNames:  []string{modelName},
	}
	ingTestObj.FillParams()
	SetUpIngressForCacheSyncCheck(t, ingTestObj)

	hostrule := integrationtest.FakeHostRule{
		Name:      hrname,
		Namespace: "default",
		Fqdn:      "foo.com",
	}
	hrObj := hostrule.HostRule()
	hrObj.ResourceVersion = "1"
	if _, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().HostRules("default").Create(context.TODO(), hrObj, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in creating HostRule: %v", err)
	}
	g.Eventually(func() string {
		hostrule, _ := v1beta1CRDClient.AkoV1beta1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 20*time.Second).Should(gomega.Equal("Accepted"))

	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--foo.com"}
	integrationtest.VerifyMetadataHostRule(t, g, sniVSKey, "default/"+hrname, true)
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(*nodes[0].SniNodes[0].Enabled).To(gomega.Equal(true))
	g.Expect(nodes[0].SniNodes[0].SslKeyAndCertificateRefs).To(gomega.HaveLen(0))
	g.Expect(nodes[0].SniNodes[0].ICAPProfileRefs).To(gomega.HaveLen(0))
	g.Expect(nodes[0].SniNodes[0].WafPolicyRef).To(gomega.BeNil())
	g.Expect(nodes[0].SniNodes[0].ApplicationProfileRef).To(gomega.BeNil())
	g.Expect(nodes[0].SniNodes[0].AnalyticsProfileRef).To(gomega.BeNil())
	g.Expect(nodes[0].SniNodes[0].ErrorPageProfileRef).To(gomega.Equal(""))
	g.Expect(nodes[0].SniNodes[0].HttpPolicySetRefs).To(gomega.HaveLen(0))
	g.Expect(nodes[0].SniNodes[0].VsDatascriptRefs).To(gomega.HaveLen(0))
	g.Expect(nodes[0].SniNodes[0].SslProfileRef).To(gomega.BeNil())
	g.Expect(nodes[0].SniNodes[0].VHDomainNames).To(gomega.ContainElement("foo.com"))

	integrationtest.TeardownHostRule(t, g, sniVSKey, hrname)
	integrationtest.VerifyMetadataHostRule(t, g, sniVSKey, "default/"+hrname, false)
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes[0].SniNodes[0].Enabled).To(gomega.BeNil())
	g.Expect(nodes[0].SniNodes[0].SslKeyAndCertificateRefs).To(gomega.HaveLen(0))
	g.Expect(nodes[0].SniNodes[0].ICAPProfileRefs).To(gomega.HaveLen(0))
	g.Expect(nodes[0].SniNodes[0].WafPolicyRef).To(gomega.BeNil())
	g.Expect(nodes[0].SniNodes[0].ApplicationProfileRef).To(gomega.BeNil())
	g.Expect(nodes[0].SniNodes[0].AnalyticsProfileRef).To(gomega.BeNil())
	g.Expect(nodes[0].SniNodes[0].ErrorPageProfileRef).To(gomega.Equal(""))
	g.Expect(nodes[0].SniNodes[0].HttpPolicySetRefs).To(gomega.HaveLen(0))
	g.Expect(nodes[0].SniNodes[0].VsDatascriptRefs).To(gomega.HaveLen(0))
	g.Expect(nodes[0].SniNodes[0].SslProfileRef).To(gomega.BeNil())
	g.Expect(nodes[0].SniNodes[0].VHDomainNames).To(gomega.ContainElement("foo.com"))

	TearDownIngressForCacheSyncCheck(t, ingName, svcName, secretName, modelName)
}

func TestSharedVSHostRuleNoListenerForSNI(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName := MODEL_NAME_PREFIX + "0"
	svcName := objNameMap.GenerateName("avisvc")
	hrname := objNameMap.GenerateName("samplehr-foo")
	secretName := objNameMap.GenerateName("my-secret")
	ingName := objNameMap.GenerateName("foo-with-targets")
	ingTestObj := IngressTestObject{
		ingressName: ingName,
		isTLS:       true,
		withSecret:  true,
		secretName:  secretName,
		serviceName: svcName,
		modelNames:  []string{modelName},
	}
	ingTestObj.FillParams()
	SetUpIngressForCacheSyncCheck(t, ingTestObj)

	hostrule := integrationtest.FakeHostRule{
		Name:               hrname,
		Namespace:          "default",
		Fqdn:               "cluster--Shared-L7-0.admin.com",
		WafPolicy:          "thisisaviref-waf",
		ApplicationProfile: "thisisaviref-appprof",
		AnalyticsProfile:   "thisisaviref-analyticsprof",
		ErrorPageProfile:   "thisisaviref-errorprof",
		Datascripts:        []string{"thisisaviref-ds2", "thisisaviref-ds1"},
		HttpPolicySets:     []string{"thisisaviref-httpps2", "thisisaviref-httpps1"},
	}
	hrCreate := hostrule.HostRule()
	hrCreate.Spec.VirtualHost.TCPSettings = &v1beta1.HostRuleTCPSettings{
		LoadBalancerIP: "80.80.80.80",
	}
	if _, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().HostRules("default").Create(context.TODO(), hrCreate, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding HostRule: %v", err)
	}

	g.Eventually(func() string {
		hostrule, _ := v1beta1CRDClient.AkoV1beta1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 10*time.Second).Should(gomega.Equal("Accepted"))

	vsKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--Shared-L7-0"}
	integrationtest.VerifyMetadataHostRule(t, g, vsKey, "default/"+hrname, true)
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(*nodes[0].Enabled).To(gomega.Equal(true))
	g.Expect(nodes[0].PortProto).To(gomega.HaveLen(2))
	var ports []int
	for _, port := range nodes[0].PortProto {
		ports = append(ports, int(port.Port))
		if port.EnableSSL {
			g.Expect(int(port.Port)).To(gomega.Equal(443))
		}
	}
	sort.Ints(ports)
	g.Expect(ports[0]).To(gomega.Equal(80))
	g.Expect(ports[1]).To(gomega.Equal(443))
	g.Expect(nodes[0].VSVIPRefs[0].IPAddress).To(gomega.Equal("80.80.80.80"))

	integrationtest.TeardownHostRule(t, g, vsKey, hrname)
	integrationtest.VerifyMetadataHostRule(t, g, vsKey, "default/"+hrname, false)

	TearDownIngressForCacheSyncCheck(t, ingName, svcName, secretName, modelName)
}

// HttpRule tests

func TestHTTPRuleCreateDelete(t *testing.T) {
	// ingress secure foo.com/foo /bar
	// create httprule /foo, nothing happens
	// create hostrule, httprule gets attached check on /foo /bar
	// delete hostrule, httprule gets detached
	g := gomega.NewGomegaWithT(t)

	modelName := MODEL_NAME_PREFIX + "0"
	svcName := objNameMap.GenerateName("avisvc")
	rrname := objNameMap.GenerateName("samplerr-foo")
	secretName := objNameMap.GenerateName("my-secret")
	ingName := objNameMap.GenerateName("foo-with-targets")

	SetupDomain()
	SetUpTestForIngress(t, svcName, modelName)
	integrationtest.AddSecret(secretName, "default", "tlsCert", "tlsKey")
	integrationtest.PollForCompletion(t, modelName, 5)
	ingressObject := integrationtest.FakeIngress{
		Name:        ingName,
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		Paths:       []string{"/foo", "/bar"},
		ServiceName: svcName,
		TlsSecretDNS: map[string][]string{
			secretName: {"foo.com"},
		},
	}

	ingrFake := ingressObject.Ingress(true)
	if _, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	integrationtest.PollForCompletion(t, modelName, 5)

	poolFooKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--default-foo.com_foo-" + ingName}
	poolBarKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--default-foo.com_bar-" + ingName}
	httpRulePath := "/"
	integrationtest.SetupHTTPRule(t, rrname, "foo.com", httpRulePath)
	integrationtest.VerifyMetadataHTTPRule(t, g, poolFooKey, "default/"+rrname+"/"+httpRulePath, true)
	integrationtest.VerifyMetadataHTTPRule(t, g, poolBarKey, "default/"+rrname+"/"+httpRulePath, true)
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(*nodes[0].SniNodes[0].PoolRefs[0].LbAlgorithm).To(gomega.Equal("LB_ALGORITHM_CONSISTENT_HASH"))
	g.Expect(*nodes[0].SniNodes[0].PoolRefs[0].LbAlgorithmHash).To(gomega.Equal("LB_ALGORITHM_CONSISTENT_HASH_SOURCE_IP_ADDRESS"))
	g.Expect(*nodes[0].SniNodes[0].PoolRefs[0].SslProfileRef).To(gomega.ContainSubstring("thisisaviref-sslprofile"))
	g.Expect(nodes[0].SniNodes[0].PoolRefs[0].PkiProfile.CACert).To(gomega.Equal("httprule-destinationCA"))
	g.Expect(nodes[0].SniNodes[0].PoolRefs[0].HealthMonitorRefs).To(gomega.HaveLen(2))
	g.Expect(nodes[0].SniNodes[0].PoolRefs[0].HealthMonitorRefs[0]).To(gomega.ContainSubstring("thisisaviref-hm2"))
	g.Expect(nodes[0].SniNodes[0].PoolRefs[0].HealthMonitorRefs[1]).To(gomega.ContainSubstring("thisisaviref-hm1"))

	// delete httprule deletes refs as well
	integrationtest.TeardownHTTPRule(t, rrname)
	integrationtest.VerifyMetadataHTTPRule(t, g, poolFooKey, "default/"+rrname+"/"+httpRulePath, false)
	integrationtest.VerifyMetadataHTTPRule(t, g, poolBarKey, "default/"+rrname+"/"+httpRulePath, false)
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes[0].SniNodes[0].PoolRefs[0].LbAlgorithm).To(gomega.BeNil())
	g.Expect(nodes[0].SniNodes[0].PoolRefs[0].SslProfileRef).To(gomega.BeNil())
	g.Expect(nodes[0].SniNodes[0].PoolRefs[0].PkiProfile).To(gomega.BeNil())
	g.Expect(nodes[0].SniNodes[0].PoolRefs[0].HealthMonitorRefs).To(gomega.HaveLen(0))

	TearDownIngressForCacheSyncCheck(t, ingName, svcName, secretName, modelName)
}

func TestHTTPRuleCreateDeleteWithPkiRef(t *testing.T) {
	// ingress secure foo.com/foo /bar
	// create httprule /foo, nothing happens
	// create hostrule, httprule gets attached check on /foo /bar
	// delete hostrule, httprule gets detached
	g := gomega.NewGomegaWithT(t)

	modelName := MODEL_NAME_PREFIX + "0"
	svcName := objNameMap.GenerateName("avisvc")
	rrname := objNameMap.GenerateName("samplerr-foo")
	secretName := objNameMap.GenerateName("my-secret")
	ingName := objNameMap.GenerateName("foo-with-targets")

	SetupDomain()
	SetUpTestForIngress(t, svcName, modelName)
	integrationtest.AddSecret(secretName, "default", "tlsCert", "tlsKey")
	integrationtest.PollForCompletion(t, modelName, 5)
	ingressObject := integrationtest.FakeIngress{
		Name:        ingName,
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		Paths:       []string{"/foo", "/bar"},
		ServiceName: svcName,
		TlsSecretDNS: map[string][]string{
			secretName: {"foo.com"},
		},
	}

	ingrFake := ingressObject.Ingress(true)
	if _, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	integrationtest.PollForCompletion(t, modelName, 5)

	poolFooKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--default-foo.com_foo-" + ingName}
	poolBarKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--default-foo.com_bar-" + ingName}

	httpRulePath := "/"
	httprule := integrationtest.FakeHTTPRule{
		Name:      rrname,
		Namespace: "default",
		Fqdn:      "foo.com",
		PathProperties: []integrationtest.FakeHTTPRulePath{{
			Path:        httpRulePath,
			PkiProfile:  "thisisaviref-pkiprofile",
			LbAlgorithm: "LB_ALGORITHM_CONSISTENT_HASH",
		}},
	}

	rrCreate := httprule.HTTPRule()
	if _, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().HTTPRules("default").Create(context.TODO(), rrCreate, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding HTTPRule: %v", err)
	}

	integrationtest.VerifyMetadataHTTPRule(t, g, poolFooKey, "default/"+rrname+"/"+httpRulePath, true)
	integrationtest.VerifyMetadataHTTPRule(t, g, poolBarKey, "default/"+rrname+"/"+httpRulePath, true)
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(*nodes[0].SniNodes[0].PoolRefs[0].LbAlgorithm).To(gomega.Equal("LB_ALGORITHM_CONSISTENT_HASH"))
	g.Expect(*nodes[0].SniNodes[0].PoolRefs[0].PkiProfileRef).To(gomega.ContainSubstring("thisisaviref-pkiprofile"))
	g.Expect(nodes[0].SniNodes[0].PoolRefs[0].PkiProfile).To(gomega.BeNil())

	// delete httprule deletes refs as well
	integrationtest.TeardownHTTPRule(t, rrname)
	integrationtest.VerifyMetadataHTTPRule(t, g, poolFooKey, "default/"+rrname+"/"+httpRulePath, false)
	integrationtest.VerifyMetadataHTTPRule(t, g, poolBarKey, "default/"+rrname+"/"+httpRulePath, false)
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes[0].SniNodes[0].PoolRefs[0].LbAlgorithm).To(gomega.BeNil())
	g.Expect(nodes[0].SniNodes[0].PoolRefs[0].PkiProfileRef).To(gomega.BeNil())
	g.Expect(nodes[0].SniNodes[0].PoolRefs[0].PkiProfile).To(gomega.BeNil())

	TearDownIngressForCacheSyncCheck(t, ingName, svcName, secretName, modelName)
}

func TestHTTPRuleHostSwitch(t *testing.T) {
	// ingress foo.com/foo voo.com/foo
	// hr1: foo.com (secure), hr2: voo.com (insecure)
	// rr1: hr1/foo ALGO1
	// switch rr1 host to hr2
	g := gomega.NewGomegaWithT(t)

	modelName := MODEL_NAME_PREFIX + "0"
	svcName := objNameMap.GenerateName("avisvc")
	hrnameFoo := objNameMap.GenerateName("samplehr-foo")
	hrnameVoo := objNameMap.GenerateName("samplehr-voo")
	rrnameFoo := objNameMap.GenerateName("samplerr-foo")
	ingName := objNameMap.GenerateName("foo-with-targets")

	// creates foo.com insecure
	ingTestObj := IngressTestObject{
		ingressName: ingName,
		isTLS:       false,
		withSecret:  false,
		serviceName: svcName,
		modelNames:  []string{modelName},
	}
	ingTestObj.FillParams()
	SetUpIngressForCacheSyncCheck(t, ingTestObj)

	// creates voo.com insecure
	ingressObject := integrationtest.FakeIngress{
		Name:        "voo-with-targets",
		Namespace:   "default",
		DnsNames:    []string{"voo.com"},
		Paths:       []string{"/foo"},
		ServiceName: svcName,
	}
	ingrFake := ingressObject.Ingress()
	if _, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in updating Ingress: %v", err)
	}

	integrationtest.SetupHTTPRule(t, rrnameFoo, "foo.com", "/foo")
	integrationtest.SetupHostRule(t, hrnameFoo, "foo.com", true) // makes foo.com secure
	integrationtest.SetupHostRule(t, hrnameVoo, "voo.com", false)

	g.Eventually(func() bool {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		if len(nodes[0].PoolRefs) == 1 &&
			nodes[0].PoolRefs[0].LbAlgorithm == nil &&
			len(nodes[0].SniNodes) == 1 &&
			len(nodes[0].SniNodes[0].PoolRefs) == 1 &&
			*nodes[0].SniNodes[0].PoolRefs[0].LbAlgorithm == "LB_ALGORITHM_CONSISTENT_HASH" {
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
	if _, err := v1beta1CRDClient.AkoV1beta1().HTTPRules("default").Update(context.TODO(), rrUpdate, metav1.UpdateOptions{}); err != nil {
		t.Fatalf("error in updating HostRule: %v", err)
	}

	// httprule things should get attached to insecure Pools of bar.com now
	// earlier since the hostrule pointed to secure foo.com, it was attached to the SNI pools
	g.Eventually(func() bool {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		if len(nodes[0].PoolRefs) == 1 &&
			nodes[0].PoolRefs[0].LbAlgorithm != nil &&
			*nodes[0].PoolRefs[0].LbAlgorithm == "LB_ALGORITHM_CONSISTENT_HASH" &&
			len(nodes[0].SniNodes) == 1 &&
			len(nodes[0].SniNodes[0].PoolRefs) == 1 &&
			nodes[0].SniNodes[0].PoolRefs[0].LbAlgorithm == nil {
			return true
		}
		return false
	}, 25*time.Second).Should(gomega.Equal(true))

	if err := KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), "voo-with-targets", metav1.DeleteOptions{}); err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--foo.com"}
	integrationtest.TeardownHostRule(t, g, sniVSKey, hrnameFoo)
	integrationtest.TeardownHostRule(t, g, sniVSKey, hrnameVoo)
	integrationtest.TeardownHTTPRule(t, rrnameFoo)
	TearDownIngressForCacheSyncCheck(t, ingName, svcName, "", modelName)
}

func TestHTTPRuleWithInvalidPath(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName := MODEL_NAME_PREFIX + "0"
	svcName := objNameMap.GenerateName("avisvc")
	rrname := objNameMap.GenerateName("samplerr-foo-1")
	secretName := objNameMap.GenerateName("my-secret")
	ingName := objNameMap.GenerateName("foo-with-targets")

	SetupDomain()
	SetUpTestForIngress(t, svcName, modelName)
	integrationtest.AddSecret(secretName, "default", "tlsCert", "tlsKey")
	integrationtest.PollForCompletion(t, modelName, 5)
	ingressObject := integrationtest.FakeIngress{
		Name:        ingName,
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		Paths:       []string{"/foo", "/bar"},
		ServiceName: svcName,
		TlsSecretDNS: map[string][]string{
			secretName: {"foo.com"},
		},
	}

	ingrFake := ingressObject.Ingress(true)
	if _, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	integrationtest.PollForCompletion(t, modelName, 5)

	// create a httprule with a non-existing path
	integrationtest.SetupHTTPRule(t, rrname, "foo.com", "/invalidPath")

	time.Sleep(10 * time.Second)

	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes[0].SniNodes[0].PoolRefs).To(gomega.HaveLen(2))

	// pool corresponding to the path "foo"
	g.Expect(nodes[0].SniNodes[0].PoolRefs[0].LbAlgorithm).To(gomega.BeNil())
	g.Expect(nodes[0].SniNodes[0].PoolRefs[0].LbAlgorithmHash).To(gomega.BeNil())
	g.Expect(nodes[0].SniNodes[0].PoolRefs[0].SslProfileRef).To(gomega.BeNil())
	g.Expect(nodes[0].SniNodes[0].PoolRefs[0].PkiProfile).To(gomega.BeNil())
	g.Expect(nodes[0].SniNodes[0].PoolRefs[0].HealthMonitorRefs).To(gomega.HaveLen(0))

	// pool corresponding to the path "bar"
	g.Expect(nodes[0].SniNodes[0].PoolRefs[1].LbAlgorithm).To(gomega.BeNil())
	g.Expect(nodes[0].SniNodes[0].PoolRefs[1].LbAlgorithmHash).To(gomega.BeNil())
	g.Expect(nodes[0].SniNodes[0].PoolRefs[1].SslProfileRef).To(gomega.BeNil())
	g.Expect(nodes[0].SniNodes[0].PoolRefs[1].PkiProfile).To(gomega.BeNil())
	g.Expect(nodes[0].SniNodes[0].PoolRefs[1].HealthMonitorRefs).To(gomega.HaveLen(0))

	// delete httprule must not change any configs
	integrationtest.TeardownHTTPRule(t, rrname)

	time.Sleep(10 * time.Second)

	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes[0].SniNodes[0].PoolRefs).To(gomega.HaveLen(2))

	// pool corresponding to the path "foo"
	g.Expect(nodes[0].SniNodes[0].PoolRefs[0].LbAlgorithm).To(gomega.BeNil())
	g.Expect(nodes[0].SniNodes[0].PoolRefs[0].SslProfileRef).To(gomega.BeNil())
	g.Expect(nodes[0].SniNodes[0].PoolRefs[0].PkiProfile).To(gomega.BeNil())
	g.Expect(nodes[0].SniNodes[0].PoolRefs[0].HealthMonitorRefs).To(gomega.HaveLen(0))

	// pool corresponding to the path "bar"
	g.Expect(nodes[0].SniNodes[0].PoolRefs[1].LbAlgorithm).To(gomega.BeNil())
	g.Expect(nodes[0].SniNodes[0].PoolRefs[1].SslProfileRef).To(gomega.BeNil())
	g.Expect(nodes[0].SniNodes[0].PoolRefs[1].PkiProfile).To(gomega.BeNil())
	g.Expect(nodes[0].SniNodes[0].PoolRefs[1].HealthMonitorRefs).To(gomega.HaveLen(0))

	TearDownIngressForCacheSyncCheck(t, ingName, svcName, secretName, modelName)
}

func TestHTTPRuleCreateDeleteEnableHTTP2(t *testing.T) {

	g := gomega.NewGomegaWithT(t)

	modelName := MODEL_NAME_PREFIX + "0"
	svcName := objNameMap.GenerateName("avisvc")
	rrname := objNameMap.GenerateName("samplerr-foo")
	secretName := objNameMap.GenerateName("my-secret")
	ingName := objNameMap.GenerateName("foo-with-targets")

	SetupDomain()
	SetUpTestForIngress(t, svcName, modelName)
	integrationtest.AddSecret(secretName, "default", "tlsCert", "tlsKey")
	integrationtest.PollForCompletion(t, modelName, 5)
	ingressObject := integrationtest.FakeIngress{
		Name:        ingName,
		Namespace:   "default",
		DnsNames:    []string{"foo.com"},
		Ips:         []string{"8.8.8.8"},
		HostNames:   []string{"v1"},
		Paths:       []string{"/foo", "/bar"},
		ServiceName: svcName,
		TlsSecretDNS: map[string][]string{
			secretName: {"foo.com"},
		},
	}

	ingrFake := ingressObject.Ingress(true)
	if _, err := KubeClient.NetworkingV1().Ingresses("default").Create(context.TODO(), ingrFake, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	integrationtest.PollForCompletion(t, modelName, 5)

	poolFooKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--default-foo.com_foo-" + ingName}
	poolBarKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--default-foo.com_bar-" + ingName}

	httpRulePath := "/"
	httprule := integrationtest.FakeHTTPRule{
		Name:      rrname,
		Namespace: "default",
		Fqdn:      "foo.com",
		PathProperties: []integrationtest.FakeHTTPRulePath{{
			Path:        httpRulePath,
			EnableHTTP2: true,
		}},
	}

	rrCreate := httprule.HTTPRule()
	if _, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().HTTPRules("default").Create(context.TODO(), rrCreate, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding HTTPRule: %v", err)
	}

	integrationtest.VerifyMetadataHTTPRule(t, g, poolFooKey, "default/"+rrname+"/"+httpRulePath, true)
	integrationtest.VerifyMetadataHTTPRule(t, g, poolBarKey, "default/"+rrname+"/"+httpRulePath, true)
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(*nodes[0].SniNodes[0].PoolRefs[0].EnableHttp2).To(gomega.Equal(true))

	// delete httprule disables HTTP2
	integrationtest.TeardownHTTPRule(t, rrname)
	integrationtest.VerifyMetadataHTTPRule(t, g, poolFooKey, "default/"+rrname+"/"+httpRulePath, false)
	integrationtest.VerifyMetadataHTTPRule(t, g, poolBarKey, "default/"+rrname+"/"+httpRulePath, false)
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes[0].SniNodes[0].PoolRefs[0].EnableHttp2).To(gomega.BeNil())

	TearDownIngressForCacheSyncCheck(t, ingName, svcName, secretName, modelName)
}

func TestHostRuleUseRegex(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName := MODEL_NAME_PREFIX + "0"
	hrname := objNameMap.GenerateName("samplehr-foo")
	fqdn := "foo.com"
	namespace := integrationtest.DEFAULT_NAMESPACE

	svcName := objNameMap.GenerateName("avisvc")
	secretName := objNameMap.GenerateName("my-secret")
	ingName := objNameMap.GenerateName("foo-with-targets")

	ingTestObj := IngressTestObject{
		ingressName: ingName,
		isTLS:       true,
		withSecret:  true,
		secretName:  secretName,
		serviceName: svcName,
		modelNames:  []string{modelName},
	}
	ingTestObj.FillParams()
	SetUpIngressForCacheSyncCheck(t, ingTestObj)

	hostrule := integrationtest.FakeHostRule{
		Name:      hrname,
		Namespace: namespace,
		Fqdn:      fqdn,
		UseRegex:  true,
	}
	hrCreate := hostrule.HostRule()
	if _, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().HostRules(namespace).Create(context.TODO(), hrCreate, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding HostRule: %v", err)
	}
	g.Eventually(func() string {
		hostrule, _ := v1beta1CRDClient.AkoV1beta1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 20*time.Second).Should(gomega.Equal("Accepted"))

	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--foo.com"}
	integrationtest.VerifyMetadataHostRule(t, g, sniVSKey, "default/"+hrname, true)

	g.Eventually(func() int {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return len(nodes[0].SniNodes)
	}, 10*time.Second).Should(gomega.Equal(1))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	node := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0]

	g.Expect(node.HttpPolicyRefs).To(gomega.HaveLen(1))
	g.Expect(node.HttpPolicyRefs[0].RedirectPorts).To(gomega.HaveLen(1))
	g.Expect(node.HttpPolicyRefs[0].RedirectPorts[0].Hosts).Should(gomega.ContainElements(fqdn))

	g.Expect(node.SniNodes).To(gomega.HaveLen(1))
	g.Expect(node.SniNodes[0].HttpPolicyRefs).To(gomega.HaveLen(1))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].HppMap).To(gomega.HaveLen(1))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].HppMap[0].MatchCriteria).Should(gomega.Equal("REGEX_MATCH"))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].HppMap[0].MatchCase).Should(gomega.Equal("INSENSITIVE"))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].HppMap[0].StringGroupRefs).To(gomega.HaveLen(1))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].HppMap[0].PoolGroup).To(gomega.Equal(lib.GetEncodedSniPGPoolNameforRegex("cluster--default-foo.com_foo-" + ingName)))

	integrationtest.TeardownHostRule(t, g, sniVSKey, hrname)
	time.Sleep(2 * time.Second)

	g.Expect(node.SniNodes).To(gomega.HaveLen(1))
	g.Expect(node.SniNodes[0].HttpPolicyRefs).To(gomega.HaveLen(1))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].RedirectPorts).To(gomega.BeNil())
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].HppMap).To(gomega.HaveLen(1))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].HppMap[0].MatchCriteria).Should(gomega.Equal("BEGINS_WITH"))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].HppMap[0].StringGroupRefs).To(gomega.HaveLen(0))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].HppMap[0].PoolGroup).To(gomega.Equal("cluster--default-foo.com_foo-" + ingName))

	TearDownIngressForCacheSyncCheck(t, ingName, svcName, secretName, modelName)
}

func TestHostRuleUseRegexNoPath(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName := MODEL_NAME_PREFIX + "0"
	hrname := objNameMap.GenerateName("samplehr-foo")
	fqdn := "foo.com"
	namespace := integrationtest.DEFAULT_NAMESPACE

	svcName := objNameMap.GenerateName("avisvc")
	secretName := objNameMap.GenerateName("my-secret")
	ingName := objNameMap.GenerateName("foo-with-targets")

	ingTestObj := IngressTestObject{
		ingressName: ingName,
		isTLS:       true,
		withSecret:  true,
		secretName:  secretName,
		serviceName: svcName,
		modelNames:  []string{modelName},
	}
	ingTestObj.FillParams()
	ingTestObj.paths = []string{}
	SetUpIngressForCacheSyncCheck(t, ingTestObj)

	hostrule := integrationtest.FakeHostRule{
		Name:      hrname,
		Namespace: namespace,
		Fqdn:      fqdn,
		UseRegex:  true,
	}
	hrCreate := hostrule.HostRule()
	if _, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().HostRules(namespace).Create(context.TODO(), hrCreate, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding HostRule: %v", err)
	}
	g.Eventually(func() string {
		hostrule, _ := v1beta1CRDClient.AkoV1beta1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 20*time.Second).Should(gomega.Equal("Accepted"))

	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--foo.com"}
	integrationtest.VerifyMetadataHostRule(t, g, sniVSKey, "default/"+hrname, true)

	g.Eventually(func() int {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return len(nodes[0].SniNodes)
	}, 10*time.Second).Should(gomega.Equal(1))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	node := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0]

	g.Expect(node.HttpPolicyRefs).To(gomega.HaveLen(1))
	g.Expect(node.HttpPolicyRefs[0].RedirectPorts).To(gomega.HaveLen(1))
	g.Expect(node.HttpPolicyRefs[0].RedirectPorts[0].Hosts).Should(gomega.ContainElements(fqdn))

	g.Expect(node.SniNodes).To(gomega.HaveLen(1))
	g.Expect(node.SniNodes[0].HttpPolicyRefs).To(gomega.HaveLen(1))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].HppMap).To(gomega.HaveLen(1))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].HppMap[0].MatchCriteria).Should(gomega.Equal("BEGINS_WITH"))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].HppMap[0].StringGroupRefs).To(gomega.BeNil())
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].HppMap[0].PoolGroup).To(gomega.Equal(lib.GetEncodedSniPGPoolNameforRegex("cluster--default-foo.com-" + ingName)))

	integrationtest.TeardownHostRule(t, g, sniVSKey, hrname)
	time.Sleep(2 * time.Second)

	g.Expect(node.SniNodes).To(gomega.HaveLen(1))
	g.Expect(node.SniNodes[0].HttpPolicyRefs).To(gomega.HaveLen(1))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].RedirectPorts).To(gomega.BeNil())
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].HppMap).To(gomega.HaveLen(1))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].HppMap[0].MatchCriteria).Should(gomega.Equal("BEGINS_WITH"))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].HppMap[0].StringGroupRefs).To(gomega.HaveLen(0))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].HppMap[0].PoolGroup).To(gomega.Equal("cluster--default-foo.com-" + ingName))

	TearDownIngressForCacheSyncCheck(t, ingName, svcName, secretName, modelName)
}

func TestHostRuleUseRegexMultiIngressNoPath(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName := MODEL_NAME_PREFIX + "0"
	hrname := objNameMap.GenerateName("samplehr-foo")
	fqdn := "foo.com"
	namespace := integrationtest.DEFAULT_NAMESPACE

	svcName := objNameMap.GenerateName("avisvc")
	secretName := objNameMap.GenerateName("my-secret")
	ingName := objNameMap.GenerateName("foo-with-targets")

	ingTestObj := IngressTestObject{
		ingressName: ingName,
		isTLS:       true,
		withSecret:  true,
		secretName:  secretName,
		serviceName: svcName,
		modelNames:  []string{modelName},
	}
	ingTestObj.FillParams()
	ingTestObj.paths = []string{}
	SetUpIngressForCacheSyncCheck(t, ingTestObj)

	ingressName2 := objNameMap.GenerateName("foo-with-targets")
	ingTestObj2 := IngressTestObject{
		ingressName: ingressName2,
		isTLS:       true,
		withSecret:  true,
		secretName:  secretName,
		serviceName: svcName,
		modelNames:  []string{modelName},
	}
	ingTestObj2.FillParams()
	// only creating ingress
	CreateIngress(t, ingTestObj2)

	hostrule := integrationtest.FakeHostRule{
		Name:      hrname,
		Namespace: namespace,
		Fqdn:      fqdn,
		UseRegex:  true,
	}
	hrCreate := hostrule.HostRule()
	if _, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().HostRules(namespace).Create(context.TODO(), hrCreate, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding HostRule: %v", err)
	}
	g.Eventually(func() string {
		hostrule, _ := v1beta1CRDClient.AkoV1beta1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 20*time.Second).Should(gomega.Equal("Accepted"))

	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--foo.com"}
	integrationtest.VerifyMetadataHostRule(t, g, sniVSKey, "default/"+hrname, true)

	g.Eventually(func() int {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return len(nodes[0].SniNodes)
	}, 10*time.Second).Should(gomega.Equal(1))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	node := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0]

	g.Eventually(func() bool {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return nodes[0].SniNodes[0].HttpPolicyRefs[0].HppMap[1].MatchCriteria == "REGEX_MATCH" || nodes[0].SniNodes[0].HttpPolicyRefs[0].HppMap[0].MatchCriteria == "REGEX_MATCH"
	}, 10*time.Second).Should(gomega.Equal(true))

	g.Expect(node.HttpPolicyRefs).To(gomega.HaveLen(1))
	g.Expect(node.HttpPolicyRefs[0].RedirectPorts).To(gomega.HaveLen(1))
	g.Expect(node.HttpPolicyRefs[0].RedirectPorts[0].Hosts).Should(gomega.ContainElements(fqdn))

	g.Expect(node.SniNodes).To(gomega.HaveLen(1))
	g.Expect(node.SniNodes[0].HttpPolicyRefs).To(gomega.HaveLen(1))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].HppMap).To(gomega.HaveLen(2))

	hppMapRegexMatch := node.SniNodes[0].HttpPolicyRefs[0].HppMap[1]
	hppMapBeginsWith := node.SniNodes[0].HttpPolicyRefs[0].HppMap[0]
	if node.SniNodes[0].HttpPolicyRefs[0].HppMap[0].MatchCriteria == "REGEX_MATCH" {
		hppMapRegexMatch = node.SniNodes[0].HttpPolicyRefs[0].HppMap[0]
		hppMapBeginsWith = node.SniNodes[0].HttpPolicyRefs[0].HppMap[1]
	}

	g.Expect(hppMapBeginsWith.MatchCriteria).Should(gomega.Equal("BEGINS_WITH"))
	g.Expect(hppMapBeginsWith.StringGroupRefs).To(gomega.BeNil())
	g.Expect(hppMapBeginsWith.PoolGroup).To(gomega.Equal(lib.GetEncodedSniPGPoolNameforRegex("cluster--default-foo.com-" + ingName)))
	g.Expect(hppMapRegexMatch.MatchCriteria).Should(gomega.Equal("REGEX_MATCH"))
	g.Expect(hppMapRegexMatch.MatchCase).Should(gomega.Equal("INSENSITIVE"))
	g.Expect(hppMapRegexMatch.StringGroupRefs).To(gomega.HaveLen(1))
	g.Expect(hppMapRegexMatch.PoolGroup).To(gomega.Equal(lib.GetEncodedSniPGPoolNameforRegex("cluster--default-foo.com_foo-" + ingressName2)))

	integrationtest.TeardownHostRule(t, g, sniVSKey, hrname)

	g.Eventually(func() bool {
		_, aviModel = objects.SharedAviGraphLister().Get(modelName)
		node = aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0]
		return (node.SniNodes[0].HttpPolicyRefs[0].HppMap[0].PoolGroup == "cluster--default-foo.com-"+ingName || node.SniNodes[0].HttpPolicyRefs[0].HppMap[1].PoolGroup == "cluster--default-foo.com-"+ingName) &&
			(node.SniNodes[0].HttpPolicyRefs[0].HppMap[0].PoolGroup == "cluster--default-foo.com_foo-"+ingressName2 || node.SniNodes[0].HttpPolicyRefs[0].HppMap[1].PoolGroup == "cluster--default-foo.com_foo-"+ingressName2)
	}, 10*time.Second).Should(gomega.Equal(true))

	g.Expect(node.SniNodes).To(gomega.HaveLen(1))
	g.Expect(node.SniNodes[0].HttpPolicyRefs).To(gomega.HaveLen(1))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].RedirectPorts).To(gomega.BeNil())
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].HppMap).To(gomega.HaveLen(2))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].HppMap[0].MatchCriteria).Should(gomega.Equal("BEGINS_WITH"))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].HppMap[0].StringGroupRefs).To(gomega.HaveLen(0))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].HppMap[1].MatchCriteria).Should(gomega.Equal("BEGINS_WITH"))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].HppMap[1].StringGroupRefs).To(gomega.HaveLen(0))

	// deleting the second created ingress
	if err := KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), ingressName2, metav1.DeleteOptions{}); err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	TearDownIngressForCacheSyncCheck(t, ingName, svcName, secretName, modelName)
}

func TestHostRuleAppRoot(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName := MODEL_NAME_PREFIX + "0"
	hrname := objNameMap.GenerateName("samplehr-foo")
	fqdn := "foo.com"
	namespace := integrationtest.DEFAULT_NAMESPACE
	appRootPath := "/foo"

	svcName := objNameMap.GenerateName("avisvc")
	secretName := objNameMap.GenerateName("my-secret")
	ingName := objNameMap.GenerateName("foo-with-targets")

	ingTestObj := IngressTestObject{
		ingressName: ingName,
		namespace:   namespace,
		dnsNames:    []string{fqdn},
		paths:       []string{"/"},
		isTLS:       true,
		withSecret:  true,
		secretName:  secretName,
		serviceName: svcName,
		modelNames:  []string{modelName},
	}
	ingTestObj.FillParams()
	SetUpIngressForCacheSyncCheck(t, ingTestObj)

	hostrule := integrationtest.FakeHostRule{
		Name:                hrname,
		Namespace:           namespace,
		Fqdn:                fqdn,
		ApplicationRootPath: appRootPath,
	}
	hrCreate := hostrule.HostRule()
	if _, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().HostRules(namespace).Create(context.TODO(), hrCreate, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding HostRule: %v", err)
	}
	g.Eventually(func() string {
		hostrule, _ := v1beta1CRDClient.AkoV1beta1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 20*time.Second).Should(gomega.Equal("Accepted"))

	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--foo.com"}
	integrationtest.VerifyMetadataHostRule(t, g, sniVSKey, "default/"+hrname, true)

	g.Eventually(func() int {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return len(nodes[0].SniNodes)
	}, 10*time.Second).Should(gomega.Equal(1))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	node := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0]

	g.Expect(node.HttpPolicyRefs).To(gomega.HaveLen(1))
	g.Expect(node.HttpPolicyRefs[0].RedirectPorts).To(gomega.HaveLen(1))
	g.Expect(node.HttpPolicyRefs[0].RedirectPorts[0].Hosts).Should(gomega.ContainElements(fqdn))

	g.Expect(node.SniNodes).To(gomega.HaveLen(1))
	g.Expect(node.SniNodes[0].HttpPolicyRefs).To(gomega.HaveLen(1))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].HppMap).To(gomega.HaveLen(1))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].HppMap[0].MatchCriteria).Should(gomega.Equal("BEGINS_WITH"))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].HppMap[0].Path[0]).Should(gomega.Equal(appRootPath))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].HppMap[0].StringGroupRefs).To(gomega.HaveLen(0))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].HppMap[0].PoolGroup).To(gomega.Equal("cluster--default-foo.com_-" + ingName))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].RedirectPorts).To(gomega.HaveLen(2))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].RedirectPorts[0].Path).To(gomega.Equal("/"))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].RedirectPorts[0].Protocol).To(gomega.Equal("HTTP"))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].RedirectPorts[0].RedirectPort).To(gomega.Equal(int32(80)))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].RedirectPorts[0].RedirectPath).To(gomega.Equal(strings.TrimPrefix(appRootPath, "/")))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].RedirectPorts[0].MatchCriteriaPath).To(gomega.Equal("EQUALS"))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].RedirectPorts[0].MatchCriteriaPort).To(gomega.Equal("IS_IN"))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].RedirectPorts[1].Path).To(gomega.Equal("/"))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].RedirectPorts[1].Protocol).To(gomega.Equal("HTTPS"))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].RedirectPorts[1].RedirectPort).To(gomega.Equal(int32(443)))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].RedirectPorts[1].RedirectPath).To(gomega.Equal(strings.TrimPrefix(appRootPath, "/")))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].RedirectPorts[1].MatchCriteriaPath).To(gomega.Equal("EQUALS"))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].RedirectPorts[1].MatchCriteriaPort).To(gomega.Equal("IS_IN"))

	integrationtest.TeardownHostRule(t, g, sniVSKey, hrname)
	g.Eventually(func() bool {
		return node.SniNodes[0].HttpPolicyRefs[0].RedirectPorts == nil
	}, 30*time.Second).Should(gomega.BeTrue())

	g.Expect(node.SniNodes).To(gomega.HaveLen(1))
	g.Expect(node.SniNodes[0].HttpPolicyRefs).To(gomega.HaveLen(1))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].RedirectPorts).To(gomega.BeNil())
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].HppMap).To(gomega.HaveLen(1))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].HppMap[0].MatchCriteria).Should(gomega.Equal("BEGINS_WITH"))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].HppMap[0].Path[0]).To(gomega.Equal("/"))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].HppMap[0].StringGroupRefs).To(gomega.HaveLen(0))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].HppMap[0].PoolGroup).To(gomega.Equal("cluster--default-foo.com_-" + ingName))

	TearDownIngressForCacheSyncCheck(t, ingName, svcName, secretName, modelName)
}

func TestHostRuleRegexAppRoot(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName := MODEL_NAME_PREFIX + "0"
	hrname := objNameMap.GenerateName("samplehr-foo")
	fqdn := "foo.com"
	namespace := integrationtest.DEFAULT_NAMESPACE
	appRootPath := "/foo"

	svcName := objNameMap.GenerateName("avisvc")
	secretName := objNameMap.GenerateName("my-secret")
	ingName := objNameMap.GenerateName("foo-with-targets")

	ingTestObj := IngressTestObject{
		ingressName: ingName,
		namespace:   namespace,
		dnsNames:    []string{fqdn, fqdn},
		paths:       []string{"/something(/|$)(.*)", "/"},
		isTLS:       true,
		withSecret:  true,
		secretName:  secretName,
		serviceName: svcName,
		modelNames:  []string{modelName},
	}
	ingTestObj.FillParams()
	SetUpIngressForCacheSyncCheck(t, ingTestObj)

	hostrule := integrationtest.FakeHostRule{
		Name:                hrname,
		Namespace:           namespace,
		Fqdn:                fqdn,
		ApplicationRootPath: appRootPath,
		UseRegex:            true,
	}
	hrCreate := hostrule.HostRule()
	if _, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().HostRules(namespace).Create(context.TODO(), hrCreate, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding HostRule: %v", err)
	}
	g.Eventually(func() string {
		hostrule, _ := v1beta1CRDClient.AkoV1beta1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 20*time.Second).Should(gomega.Equal("Accepted"))

	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--foo.com"}
	integrationtest.VerifyMetadataHostRule(t, g, sniVSKey, "default/"+hrname, true)

	g.Eventually(func() int {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return len(nodes[0].SniNodes)
	}, 10*time.Second).Should(gomega.Equal(1))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	node := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0]

	g.Expect(node.HttpPolicyRefs).To(gomega.HaveLen(1))
	g.Expect(node.HttpPolicyRefs[0].RedirectPorts).To(gomega.HaveLen(1))
	g.Expect(node.HttpPolicyRefs[0].RedirectPorts[0].Hosts).Should(gomega.ContainElements(fqdn))

	g.Expect(node.SniNodes).To(gomega.HaveLen(1))
	g.Expect(node.SniNodes[0].HttpPolicyRefs).To(gomega.HaveLen(1))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].HppMap).To(gomega.HaveLen(2))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].HppMap[0].MatchCriteria).Should(gomega.Equal("REGEX_MATCH"))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].HppMap[0].MatchCase).Should(gomega.Equal("INSENSITIVE"))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].HppMap[0].StringGroupRefs).To(gomega.HaveLen(1))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].HppMap[0].PoolGroup).To(gomega.Equal(lib.GetEncodedSniPGPoolNameforRegex("cluster--default-foo.com_something(_|$)(.*)-" + ingName)))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].HppMap[1].MatchCriteria).Should(gomega.Equal("REGEX_MATCH"))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].HppMap[1].MatchCase).Should(gomega.Equal("INSENSITIVE"))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].HppMap[1].Path[0]).To(gomega.Equal(appRootPath))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].HppMap[1].StringGroupRefs).To(gomega.HaveLen(1))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].HppMap[1].PoolGroup).To(gomega.Equal(lib.GetEncodedSniPGPoolNameforRegex("cluster--default-foo.com_-" + ingName)))

	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].RedirectPorts).To(gomega.HaveLen(2))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].RedirectPorts[0].Path).To(gomega.Equal("/"))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].RedirectPorts[0].RedirectPort).To(gomega.Equal(int32(80)))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].RedirectPorts[0].Protocol).To(gomega.Equal("HTTP"))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].RedirectPorts[0].RedirectPath).To(gomega.Equal(strings.TrimPrefix(appRootPath, "/")))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].RedirectPorts[0].MatchCriteriaPath).To(gomega.Equal("EQUALS"))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].RedirectPorts[0].MatchCriteriaPort).To(gomega.Equal("IS_IN"))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].RedirectPorts[1].Path).To(gomega.Equal("/"))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].RedirectPorts[1].Protocol).To(gomega.Equal("HTTPS"))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].RedirectPorts[1].RedirectPort).To(gomega.Equal(int32(443)))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].RedirectPorts[1].RedirectPath).To(gomega.Equal(strings.TrimPrefix(appRootPath, "/")))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].RedirectPorts[1].MatchCriteriaPath).To(gomega.Equal("EQUALS"))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].RedirectPorts[1].MatchCriteriaPort).To(gomega.Equal("IS_IN"))

	integrationtest.TeardownHostRule(t, g, sniVSKey, hrname)
	g.Eventually(func() bool {
		return node.SniNodes[0].HttpPolicyRefs[0].RedirectPorts == nil
	}, 30*time.Second).Should(gomega.BeTrue())

	g.Expect(node.SniNodes).To(gomega.HaveLen(1))
	g.Expect(node.SniNodes[0].HttpPolicyRefs).To(gomega.HaveLen(1))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].RedirectPorts).To(gomega.BeNil())
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].HppMap).To(gomega.HaveLen(2))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].HppMap[0].MatchCriteria).Should(gomega.Equal("BEGINS_WITH"))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].HppMap[0].Path[0]).To(gomega.Equal("/something(/|$)(.*)"))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].HppMap[0].StringGroupRefs).To(gomega.HaveLen(0))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].HppMap[0].PoolGroup).To(gomega.Equal("cluster--default-foo.com_something(_|$)(.*)-" + ingName))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].HppMap[1].MatchCriteria).Should(gomega.Equal("BEGINS_WITH"))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].HppMap[1].Path[0]).To(gomega.Equal("/"))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].HppMap[1].StringGroupRefs).To(gomega.HaveLen(0))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].HppMap[1].PoolGroup).To(gomega.Equal("cluster--default-foo.com_-" + ingName))

	TearDownIngressForCacheSyncCheck(t, ingName, svcName, secretName, modelName)
}

func TestHostRuleAppRootListenerPorts(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName := MODEL_NAME_PREFIX + "0"
	hrname := objNameMap.GenerateName("samplehr-foo")
	sharedHrName := objNameMap.GenerateName("samplehr-shared")
	fqdn := "foo.com"
	namespace := integrationtest.DEFAULT_NAMESPACE
	appRootPath := "/foo"

	svcName := objNameMap.GenerateName("avisvc")
	secretName := objNameMap.GenerateName("my-secret")
	ingName := objNameMap.GenerateName("foo-with-targets")

	ingTestObj := IngressTestObject{
		ingressName: ingName,
		namespace:   namespace,
		dnsNames:    []string{fqdn},
		paths:       []string{"/"},
		isTLS:       true,
		withSecret:  true,
		secretName:  secretName,
		serviceName: svcName,
		modelNames:  []string{modelName},
	}
	ingTestObj.FillParams()
	SetUpIngressForCacheSyncCheck(t, ingTestObj)

	sharedFQDN := "Shared-L7"
	hostruleShared := integrationtest.FakeHostRule{
		Name:      sharedHrName,
		Namespace: namespace,
		Fqdn:      sharedFQDN,
		FqdnType:  "Contains",
		ListenerPorts: []integrationtest.ListenerPorts{
			{
				Port:      8081,
				EnableSSL: false,
			},
			{
				Port:      6443,
				EnableSSL: true,
			},
		},
	}

	hrCreate := hostruleShared.HostRule()
	if _, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().HostRules(namespace).Create(context.TODO(), hrCreate, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding HostRule: %v", err)
	}
	g.Eventually(func() string {
		hostrule, _ := v1beta1CRDClient.AkoV1beta1().HostRules("default").Get(context.TODO(), sharedHrName, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 20*time.Second).Should(gomega.Equal("Accepted"))

	hostrule := integrationtest.FakeHostRule{
		Name:                hrname,
		Namespace:           namespace,
		Fqdn:                fqdn,
		ApplicationRootPath: "/foo",
	}
	hrCreate = hostrule.HostRule()
	if _, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().HostRules(namespace).Create(context.TODO(), hrCreate, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding HostRule: %v", err)
	}
	g.Eventually(func() string {
		hostrule, _ := v1beta1CRDClient.AkoV1beta1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 20*time.Second).Should(gomega.Equal("Accepted"))

	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--foo.com"}
	integrationtest.VerifyMetadataHostRule(t, g, sniVSKey, "default/"+hrname, true)

	g.Eventually(func() int {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return len(nodes[0].SniNodes)
	}, 10*time.Second).Should(gomega.Equal(1))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	node := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0]

	g.Expect(node.HttpPolicyRefs).To(gomega.HaveLen(1))
	g.Expect(node.HttpPolicyRefs[0].RedirectPorts).To(gomega.HaveLen(1))
	g.Expect(node.HttpPolicyRefs[0].RedirectPorts[0].Hosts).Should(gomega.ContainElements(fqdn))

	g.Expect(node.SniNodes).To(gomega.HaveLen(1))
	g.Expect(node.SniNodes[0].HttpPolicyRefs).To(gomega.HaveLen(1))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].HppMap).To(gomega.HaveLen(1))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].HppMap[0].MatchCriteria).Should(gomega.Equal("BEGINS_WITH"))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].HppMap[0].Path[0]).Should(gomega.Equal(appRootPath))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].HppMap[0].StringGroupRefs).To(gomega.HaveLen(0))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].HppMap[0].PoolGroup).To(gomega.Equal("cluster--default-foo.com_-" + ingName))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].RedirectPorts).To(gomega.HaveLen(2))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].RedirectPorts[0].Path).To(gomega.Equal("/"))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].RedirectPorts[0].Protocol).To(gomega.Equal("HTTP"))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].RedirectPorts[0].RedirectPort).To(gomega.Equal(int32(8081)))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].RedirectPorts[0].RedirectPath).To(gomega.Equal(strings.TrimPrefix(appRootPath, "/")))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].RedirectPorts[0].MatchCriteriaPath).To(gomega.Equal("EQUALS"))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].RedirectPorts[0].MatchCriteriaPort).To(gomega.Equal("IS_IN"))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].RedirectPorts[1].Path).To(gomega.Equal("/"))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].RedirectPorts[1].Protocol).To(gomega.Equal("HTTPS"))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].RedirectPorts[1].RedirectPort).To(gomega.Equal(int32(6443)))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].RedirectPorts[1].RedirectPath).To(gomega.Equal(strings.TrimPrefix(appRootPath, "/")))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].RedirectPorts[1].MatchCriteriaPath).To(gomega.Equal("EQUALS"))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].RedirectPorts[1].MatchCriteriaPort).To(gomega.Equal("IS_IN"))

	integrationtest.TeardownHostRule(t, g, sniVSKey, hrname)
	integrationtest.TeardownHostRule(t, g, sniVSKey, sharedHrName)
	g.Eventually(func() bool {
		return node.SniNodes[0].HttpPolicyRefs[0].RedirectPorts == nil
	}, 30*time.Second).Should(gomega.BeTrue())

	g.Expect(node.SniNodes).To(gomega.HaveLen(1))
	g.Expect(node.SniNodes[0].HttpPolicyRefs).To(gomega.HaveLen(1))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].RedirectPorts).To(gomega.BeNil())
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].HppMap).To(gomega.HaveLen(1))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].HppMap[0].MatchCriteria).Should(gomega.Equal("BEGINS_WITH"))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].HppMap[0].Path[0]).To(gomega.Equal("/"))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].HppMap[0].StringGroupRefs).To(gomega.HaveLen(0))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].HppMap[0].PoolGroup).To(gomega.Equal("cluster--default-foo.com_-" + ingName))

	TearDownIngressForCacheSyncCheck(t, ingName, svcName, secretName, modelName)
}

func TestHostRuleRegexAppRootListenerPorts(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName := MODEL_NAME_PREFIX + "0"
	hrname := objNameMap.GenerateName("samplehr-foo")
	sharedHrName := objNameMap.GenerateName("samplehr-shared")
	fqdn := "foo.com"
	namespace := integrationtest.DEFAULT_NAMESPACE
	appRootPath := "/foo"

	svcName := objNameMap.GenerateName("avisvc")
	secretName := objNameMap.GenerateName("my-secret")
	ingName := objNameMap.GenerateName("foo-with-targets")

	ingTestObj := IngressTestObject{
		ingressName: ingName,
		namespace:   namespace,
		dnsNames:    []string{fqdn, fqdn},
		paths:       []string{"/something(/|$)(.*)", "/"},
		isTLS:       true,
		withSecret:  true,
		secretName:  secretName,
		serviceName: svcName,
		modelNames:  []string{modelName},
	}
	ingTestObj.FillParams()
	SetUpIngressForCacheSyncCheck(t, ingTestObj)

	hostrule := integrationtest.FakeHostRule{
		Name:                hrname,
		Namespace:           namespace,
		Fqdn:                fqdn,
		ApplicationRootPath: "/foo",
		UseRegex:            true,
	}
	hrCreate := hostrule.HostRule()
	if _, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().HostRules(namespace).Create(context.TODO(), hrCreate, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding HostRule: %v", err)
	}
	g.Eventually(func() string {
		hostrule, _ := v1beta1CRDClient.AkoV1beta1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 20*time.Second).Should(gomega.Equal("Accepted"))

	sharedFQDN := "Shared-L7"
	hostruleShared := integrationtest.FakeHostRule{
		Name:      sharedHrName,
		Namespace: namespace,
		Fqdn:      sharedFQDN,
		FqdnType:  "Contains",
		ListenerPorts: []integrationtest.ListenerPorts{
			{
				Port:      8081,
				EnableSSL: false,
			},
			{
				Port:      6443,
				EnableSSL: true,
			},
		},
	}

	hrCreate = hostruleShared.HostRule()
	if _, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().HostRules(namespace).Create(context.TODO(), hrCreate, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding HostRule: %v", err)
	}
	g.Eventually(func() string {
		hostrule, _ := v1beta1CRDClient.AkoV1beta1().HostRules("default").Get(context.TODO(), sharedHrName, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 20*time.Second).Should(gomega.Equal("Accepted"))

	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--foo.com"}
	integrationtest.VerifyMetadataHostRule(t, g, sniVSKey, "default/"+hrname, true)

	g.Eventually(func() int {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return len(nodes[0].SniNodes)
	}, 10*time.Second).Should(gomega.Equal(1))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	node := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0]

	g.Expect(node.HttpPolicyRefs).To(gomega.HaveLen(1))
	g.Expect(node.HttpPolicyRefs[0].RedirectPorts).To(gomega.HaveLen(1))
	g.Expect(node.HttpPolicyRefs[0].RedirectPorts[0].Hosts).Should(gomega.ContainElements(fqdn))

	g.Expect(node.SniNodes).To(gomega.HaveLen(1))
	g.Expect(node.SniNodes[0].HttpPolicyRefs).To(gomega.HaveLen(1))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].HppMap).To(gomega.HaveLen(2))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].HppMap[0].MatchCriteria).Should(gomega.Equal("REGEX_MATCH"))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].HppMap[0].MatchCase).Should(gomega.Equal("INSENSITIVE"))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].HppMap[0].StringGroupRefs).To(gomega.HaveLen(1))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].HppMap[0].PoolGroup).To(gomega.Equal(lib.GetEncodedSniPGPoolNameforRegex("cluster--default-foo.com_something(_|$)(.*)-" + ingName)))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].HppMap[1].MatchCriteria).Should(gomega.Equal("REGEX_MATCH"))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].HppMap[1].MatchCase).Should(gomega.Equal("INSENSITIVE"))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].HppMap[1].Path[0]).To(gomega.Equal(appRootPath))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].HppMap[1].StringGroupRefs).To(gomega.HaveLen(1))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].HppMap[1].PoolGroup).To(gomega.Equal(lib.GetEncodedSniPGPoolNameforRegex("cluster--default-foo.com_-" + ingName)))

	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].RedirectPorts).To(gomega.HaveLen(2))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].RedirectPorts[0].Path).To(gomega.Equal("/"))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].RedirectPorts[0].RedirectPort).To(gomega.Equal(int32(8081)))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].RedirectPorts[0].Protocol).To(gomega.Equal("HTTP"))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].RedirectPorts[0].RedirectPath).To(gomega.Equal(strings.TrimPrefix(appRootPath, "/")))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].RedirectPorts[0].MatchCriteriaPath).To(gomega.Equal("EQUALS"))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].RedirectPorts[0].MatchCriteriaPort).To(gomega.Equal("IS_IN"))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].RedirectPorts[1].Path).To(gomega.Equal("/"))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].RedirectPorts[1].Protocol).To(gomega.Equal("HTTPS"))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].RedirectPorts[1].RedirectPort).To(gomega.Equal(int32(6443)))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].RedirectPorts[1].RedirectPath).To(gomega.Equal(strings.TrimPrefix(appRootPath, "/")))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].RedirectPorts[1].MatchCriteriaPath).To(gomega.Equal("EQUALS"))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].RedirectPorts[1].MatchCriteriaPort).To(gomega.Equal("IS_IN"))

	integrationtest.TeardownHostRule(t, g, sniVSKey, hrname)
	integrationtest.TeardownHostRule(t, g, sniVSKey, sharedHrName)
	g.Eventually(func() bool {
		return node.SniNodes[0].HttpPolicyRefs[0].RedirectPorts == nil
	}, 30*time.Second).Should(gomega.BeTrue())

	g.Expect(node.SniNodes).To(gomega.HaveLen(1))
	g.Expect(node.SniNodes[0].HttpPolicyRefs).To(gomega.HaveLen(1))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].RedirectPorts).To(gomega.BeNil())
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].HppMap).To(gomega.HaveLen(2))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].HppMap[0].MatchCriteria).Should(gomega.Equal("BEGINS_WITH"))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].HppMap[0].Path[0]).To(gomega.Equal("/something(/|$)(.*)"))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].HppMap[0].StringGroupRefs).To(gomega.HaveLen(0))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].HppMap[0].PoolGroup).To(gomega.Equal("cluster--default-foo.com_something(_|$)(.*)-" + ingName))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].HppMap[1].MatchCriteria).Should(gomega.Equal("BEGINS_WITH"))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].HppMap[1].Path[0]).To(gomega.Equal("/"))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].HppMap[1].StringGroupRefs).To(gomega.HaveLen(0))
	g.Expect(node.SniNodes[0].HttpPolicyRefs[0].HppMap[1].PoolGroup).To(gomega.Equal("cluster--default-foo.com_-" + ingName))

	TearDownIngressForCacheSyncCheck(t, ingName, svcName, secretName, modelName)
}

func TestCreateDeleteSharedVSHostRuleWithoutDNS(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	integrationtest.SetEmptyDomainList()
	defer integrationtest.ResetMiddleware()
	modelName := MODEL_NAME_PREFIX + "0"
	svcName := objNameMap.GenerateName("avisvc")
	hrname := objNameMap.GenerateName("samplehr-foo")
	secretName := objNameMap.GenerateName("my-secret")
	ingName := objNameMap.GenerateName("foo-with-targets")
	ingTestObj := IngressTestObject{
		ingressName: ingName,
		isTLS:       true,
		withSecret:  true,
		secretName:  secretName,
		serviceName: svcName,
		modelNames:  []string{modelName},
	}
	ingTestObj.FillParams()
	SetUpIngressForCacheSyncCheck(t, ingTestObj)

	hostrule := integrationtest.FakeHostRule{
		Name:                  hrname,
		Namespace:             "default",
		Fqdn:                  "Shared-L7",
		WafPolicy:             "thisisaviref-waf",
		ApplicationProfile:    "thisisaviref-appprof",
		AnalyticsProfile:      "thisisaviref-analyticsprof",
		ErrorPageProfile:      "thisisaviref-errorprof",
		Datascripts:           []string{"thisisaviref-ds2", "thisisaviref-ds1"},
		HttpPolicySets:        []string{"thisisaviref-httpps2", "thisisaviref-httpps1"},
		NetworkSecurityPolicy: "thisisaviref-networksecuritypolicyref",
	}
	hrCreate := hostrule.HostRule()
	hrCreate.Spec.VirtualHost.FqdnType = "Contains"
	hrCreate.Spec.VirtualHost.TCPSettings = &v1beta1.HostRuleTCPSettings{
		Listeners: []v1beta1.HostRuleTCPListeners{
			{Port: 8081}, {Port: 8082}, {Port: 8083, EnableSSL: true},
		},
	}
	if _, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().HostRules("default").Create(context.TODO(), hrCreate, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding HostRule: %v", err)
	}

	g.Eventually(func() string {
		hostrule, _ := v1beta1CRDClient.AkoV1beta1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 10*time.Second).Should(gomega.Equal("Accepted"))

	vsKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--Shared-L7-0"}
	integrationtest.VerifyMetadataHostRule(t, g, vsKey, "default/"+hrname, true)
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(*nodes[0].Enabled).To(gomega.Equal(true))
	g.Expect(*nodes[0].WafPolicyRef).To(gomega.ContainSubstring("thisisaviref-waf"))
	g.Expect(*nodes[0].ApplicationProfileRef).To(gomega.ContainSubstring("thisisaviref-appprof"))
	g.Expect(*nodes[0].AnalyticsProfileRef).To(gomega.ContainSubstring("thisisaviref-analyticsprof"))
	g.Expect(nodes[0].ErrorPageProfileRef).To(gomega.ContainSubstring("thisisaviref-errorprof"))
	g.Expect(nodes[0].HttpPolicySetRefs).To(gomega.HaveLen(2))
	g.Expect(nodes[0].HttpPolicySetRefs[0]).To(gomega.ContainSubstring("thisisaviref-httpps2"))
	g.Expect(nodes[0].HttpPolicySetRefs[1]).To(gomega.ContainSubstring("thisisaviref-httpps1"))
	g.Expect(nodes[0].VsDatascriptRefs).To(gomega.HaveLen(2))
	g.Expect(nodes[0].VsDatascriptRefs[0]).To(gomega.ContainSubstring("thisisaviref-ds2"))
	g.Expect(nodes[0].VsDatascriptRefs[1]).To(gomega.ContainSubstring("thisisaviref-ds1"))
	g.Expect(*nodes[0].NetworkSecurityPolicyRef).To(gomega.ContainSubstring("thisisaviref-networksecuritypolicyref"))
	g.Expect(nodes[0].PortProto).To(gomega.HaveLen(3))
	var ports []int
	for _, port := range nodes[0].PortProto {
		ports = append(ports, int(port.Port))
		if port.EnableSSL {
			g.Expect(int(port.Port)).Should(gomega.Equal(8083))
		}
	}
	sort.Ints(ports)
	g.Expect(ports[0]).To(gomega.Equal(8081))
	g.Expect(ports[1]).To(gomega.Equal(8082))
	g.Expect(ports[2]).To(gomega.Equal(8083))

	integrationtest.TeardownHostRule(t, g, vsKey, hrname)
	integrationtest.VerifyMetadataHostRule(t, g, vsKey, "default/"+hrname, false)
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes[0].Enabled).To(gomega.BeNil())
	g.Expect(nodes[0].SslKeyAndCertificateRefs).To(gomega.HaveLen(0))
	g.Expect(nodes[0].WafPolicyRef).To(gomega.BeNil())
	g.Expect(nodes[0].ApplicationProfileRef).To(gomega.BeNil())
	g.Expect(nodes[0].AnalyticsProfileRef).To(gomega.BeNil())
	g.Expect(nodes[0].ErrorPageProfileRef).To(gomega.Equal(""))
	g.Expect(nodes[0].HttpPolicySetRefs).To(gomega.HaveLen(0))
	g.Expect(nodes[0].VsDatascriptRefs).To(gomega.HaveLen(0))
	g.Expect(nodes[0].SslProfileRef).To(gomega.BeNil())
	g.Expect(nodes[0].NetworkSecurityPolicyRef).To(gomega.BeNil())
	ports = []int{}
	for _, port := range nodes[0].PortProto {
		ports = append(ports, int(port.Port))
		if port.EnableSSL {
			g.Expect(int(port.Port)).To(gomega.Equal(443))
		}
	}
	sort.Ints(ports)
	g.Expect(ports[0]).To(gomega.Equal(80))
	g.Expect(ports[1]).To(gomega.Equal(443))

	TearDownIngressForCacheSyncCheck(t, ingName, svcName, secretName, modelName)
}
