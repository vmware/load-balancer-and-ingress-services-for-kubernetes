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

package dedicatedvstests

import (
	"context"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	avinodes "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/apis/ako/v1beta1"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/integrationtest"
)

func TestHostruleFQDNAliasesForDedicatedVS(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	modelName := "admin/cluster--foo.com-L7-dedicated"
	hrname := "fqdn-aliases-hr-foo"
	secretName := objNameMap.GenerateName("my-secret")
	ingressName := objNameMap.GenerateName("foo-with-targets")
	svcName := objNameMap.GenerateName("avisvc")
	ingTestObj := IngressTestObject{
		ingressName: ingressName,
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
		nodes, ok := aviModel.(*avinodes.AviObjectGraph)
		if !ok {
			return 0
		}
		return len(nodes.GetAviVS())
	}, 20*time.Second).Should(gomega.Equal(1))
	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: lib.Encode("cluster--foo.com", lib.EVHVS)}
	integrationtest.VerifyMetadataHostRule(t, g, sniVSKey, "default/fqdn-aliases-hr-foo", true)
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()

	// Common function that takes care of all validations
	validateNode := func(node *avinodes.AviVsNode, aliases []string) {
		g.Expect(node.VSVIPRefs).To(gomega.HaveLen(1))
		g.Expect(node.VSVIPRefs[0].FQDNs).Should(gomega.ContainElements(aliases))

		g.Expect(node.VHDomainNames).Should(gomega.ContainElements(aliases))
		g.Expect(node.HttpPolicyRefs).To(gomega.HaveLen(2))
		for _, httpPolicyRef := range nodes[0].HttpPolicyRefs {
			if httpPolicyRef.HppMap != nil {
				g.Expect(httpPolicyRef.HppMap).To(gomega.HaveLen(1))
				g.Expect(httpPolicyRef.HppMap[0].Host).Should(gomega.ContainElements(aliases))
			}
			if httpPolicyRef.RedirectPorts != nil {
				g.Expect(httpPolicyRef.RedirectPorts).To(gomega.HaveLen(1))
				g.Expect(httpPolicyRef.RedirectPorts[0].Hosts).Should(gomega.ContainElements(aliases))
			}
			g.Expect(httpPolicyRef.AviMarkers.Host).Should(gomega.ContainElements(aliases))
		}
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
	_, err := V1beta1Client.AkoV1beta1().HostRules("default").Update(context.TODO(), hrUpdate, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating HostRule: %v", err)
	}
	g.Eventually(func() string {
		hostrule, _ := V1beta1Client.AkoV1beta1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 10*time.Second).Should(gomega.Equal("Accepted"))

	g.Eventually(func() int {
		_, aviModel = objects.SharedAviGraphLister().Get(modelName)
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return len(nodes)
	}, 10*time.Second).Should(gomega.Equal(1))

	// update is not getting reflected on evh nodes immediately. Hence adding a sleep of 5 seconds.
	time.Sleep(5 * time.Second)

	// Check whether the Aliases are properly added to dedicated VS.
	validateNode(nodes[0], aliases)

	// Append one more alias and check whether it is getting added to parent and child VS.
	aliases = append(aliases, "alias3.com")
	hrUpdate.Spec.VirtualHost.Aliases = aliases
	hrUpdate.ResourceVersion = "3"
	_, err = V1beta1Client.AkoV1beta1().HostRules("default").Update(context.TODO(), hrUpdate, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating HostRule: %v", err)
	}
	g.Eventually(func() string {
		hostrule, _ := V1beta1Client.AkoV1beta1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 10*time.Second).Should(gomega.Equal("Accepted"))

	g.Eventually(func() int {
		_, aviModel = objects.SharedAviGraphLister().Get(modelName)
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return len(nodes)
	}, 10*time.Second).Should(gomega.Equal(1))

	// update is not getting reflected on evh nodes immediately. Hence adding a sleep of 5 seconds.
	time.Sleep(5 * time.Second)

	// Check whether the Aliases are properly added to dedicated VS.
	validateNode(nodes[0], aliases)

	// Remove one alias from hostrule and check whether its reference is removed properly.
	aliases = aliases[1:]
	hrUpdate.Spec.VirtualHost.Aliases = aliases
	hrUpdate.ResourceVersion = "4"
	_, err = V1beta1Client.AkoV1beta1().HostRules("default").Update(context.TODO(), hrUpdate, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating HostRule: %v", err)
	}
	g.Eventually(func() string {
		hostrule, _ := V1beta1Client.AkoV1beta1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 10*time.Second).Should(gomega.Equal("Accepted"))

	g.Eventually(func() int {
		_, aviModel = objects.SharedAviGraphLister().Get(modelName)
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return len(nodes)
	}, 10*time.Second).Should(gomega.Equal(1))

	// update is not getting reflected on evh nodes immediately. Hence adding a sleep of 5 seconds.
	time.Sleep(5 * time.Second)

	// Check whether the Alias reference is properly removed from dedicated VS.
	validateNode(nodes[0], aliases)

	integrationtest.TeardownHostRule(t, g, sniVSKey, hrname)
	TearDownIngressForCacheSyncCheck(t, secretName, ingressName, svcName, modelName)
}

func TestApplyHostruleToDedicatedVS(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName := "admin/cluster--foo.com-L7-dedicated"
	hrname := "hr-cluster--foo.com-L7-dedicated"
	secretName := objNameMap.GenerateName("my-secret")
	ingressName := objNameMap.GenerateName("foo-with-targets")
	svcName := objNameMap.GenerateName("avisvc")

	ingTestObj := IngressTestObject{
		ingressName: ingressName,
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
		WafPolicy:             "thisisaviref-waf",
		ApplicationProfile:    "thisisaviref-appprof",
		AnalyticsProfile:      "thisisaviref-analyticsprof",
		ErrorPageProfile:      "thisisaviref-errorprof",
		Datascripts:           []string{"thisisaviref-ds2", "thisisaviref-ds1"},
		HttpPolicySets:        []string{"thisisaviref-httpps2", "thisisaviref-httpps1"},
		NetworkSecurityPolicy: "thisisaviref-networksecuritypolicyref",
	}
	hrObj := hostrule.HostRule()
	hrObj.Spec.VirtualHost.Fqdn = "foo.com"
	hrObj.Spec.VirtualHost.FqdnType = v1beta1.Contains
	hrObj.Spec.VirtualHost.TCPSettings = &v1beta1.HostRuleTCPSettings{
		Listeners: []v1beta1.HostRuleTCPListeners{
			{Port: 8081}, {Port: 8082, EnableSSL: true},
		},
	}

	if _, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().HostRules("default").Create(context.TODO(), hrObj, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding HostRule: %v", err)
	}

	g.Eventually(func() string {
		hostrule, _ := V1beta1Client.AkoV1beta1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 30*time.Second).Should(gomega.Equal("Accepted"))

	vsKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--foo.com-L7-dedicated"}
	integrationtest.VerifyMetadataHostRule(t, g, vsKey, "default/hr-cluster--foo.com-L7-dedicated", true)
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

	g.Expect(nodes[0].PortProto).To(gomega.HaveLen(2))
	var portsWithHostRule []int
	for _, port := range nodes[0].PortProto {
		portsWithHostRule = append(portsWithHostRule, int(port.Port))
		if port.EnableSSL {
			g.Expect(int(port.Port)).Should(gomega.Equal(8082))
		}
	}
	sort.Ints(portsWithHostRule)
	g.Expect(portsWithHostRule[0]).To(gomega.Equal(8081))
	g.Expect(portsWithHostRule[1]).To(gomega.Equal(8082))

	integrationtest.TeardownHostRule(t, g, vsKey, hrname)
	integrationtest.VerifyMetadataHostRule(t, g, vsKey, "default/hr-cluster--foo.com-L7-dedicated", false)
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
	g.Expect(nodes[0].PortProto).To(gomega.HaveLen(2))
	var portWithoutHostRule []int
	for _, port := range nodes[0].PortProto {
		portWithoutHostRule = append(portWithoutHostRule, int(port.Port))
		if port.EnableSSL {
			g.Expect(int(port.Port)).To(gomega.Equal(443))
		}
	}
	sort.Ints(portWithoutHostRule)
	g.Expect(portWithoutHostRule[0]).To(gomega.Equal(80))
	g.Expect(portWithoutHostRule[1]).To(gomega.Equal(443))

	TearDownIngressForCacheSyncCheck(t, secretName, ingressName, svcName, modelName)

}

func TestHostruleSSLKeyCertToDedicatedVS(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName := "admin/cluster--foo.com-L7-dedicated"
	hrname := "hr-cluster--foo.com-L7-dedicated"
	secretName := objNameMap.GenerateName("my-secret")
	ingressName := objNameMap.GenerateName("foo-with-targets")
	svcName := objNameMap.GenerateName("avisvc")

	ingTestObj := IngressTestObject{
		ingressName: ingressName,
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
		hostrule, _ := V1beta1Client.AkoV1beta1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 30*time.Second).Should(gomega.Equal("Accepted"))

	vsKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--foo.com-L7-dedicated"}
	integrationtest.VerifyMetadataHostRule(t, g, vsKey, "default/hr-cluster--foo.com-L7-dedicated", true)
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(*nodes[0].Enabled).To(gomega.Equal(true))
	g.Expect(nodes[0].SslKeyAndCertificateRefs).To(gomega.HaveLen(1))
	g.Expect(nodes[0].SslKeyAndCertificateRefs[0]).To(gomega.ContainSubstring("thisisaviref-sslkey"))

	integrationtest.TeardownHostRule(t, g, vsKey, hrname)
	integrationtest.VerifyMetadataHostRule(t, g, vsKey, "default/hr-cluster--foo.com-L7-dedicated", false)
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes[0].Enabled).To(gomega.BeNil())
	g.Expect(nodes[0].SslKeyAndCertificateRefs).To(gomega.HaveLen(0))

	TearDownIngressForCacheSyncCheck(t, secretName, ingressName, svcName, modelName)
}

func TestHostruleNoListenerDedicatedVS(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName := "admin/cluster--foo.com-L7-dedicated"
	hrname := "hr-cluster--foo.com-L7-dedicated"
	secretName := objNameMap.GenerateName("my-secret")
	ingressName := objNameMap.GenerateName("foo-with-targets")
	svcName := objNameMap.GenerateName("avisvc")

	ingTestObj := IngressTestObject{
		ingressName: ingressName,
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
	hrObj.Spec.VirtualHost.Fqdn = "foo.com"
	hrObj.Spec.VirtualHost.FqdnType = v1beta1.Contains
	hrObj.Spec.VirtualHost.TCPSettings = &v1beta1.HostRuleTCPSettings{
		LoadBalancerIP: "80.80.80.80",
	}

	if _, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().HostRules("default").Create(context.TODO(), hrObj, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding HostRule: %v", err)
	}

	g.Eventually(func() string {
		hostrule, _ := V1beta1Client.AkoV1beta1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 30*time.Second).Should(gomega.Equal("Accepted"))

	vsKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--foo.com-L7-dedicated"}
	integrationtest.VerifyMetadataHostRule(t, g, vsKey, "default/hr-cluster--foo.com-L7-dedicated", true)
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(*nodes[0].Enabled).To(gomega.Equal(true))
	g.Expect(nodes[0].PortProto).To(gomega.HaveLen(2))
	var portsWithHostRule []int
	for _, port := range nodes[0].PortProto {
		portsWithHostRule = append(portsWithHostRule, int(port.Port))
		if port.EnableSSL {
			g.Expect(int(port.Port)).To(gomega.Equal(443))
		}
	}
	sort.Ints(portsWithHostRule)
	g.Expect(portsWithHostRule[0]).To(gomega.Equal(80))
	g.Expect(portsWithHostRule[1]).To(gomega.Equal(443))
	g.Expect(nodes[0].VSVIPRefs[0].IPAddress).To(gomega.Equal("80.80.80.80"))

	integrationtest.TeardownHostRule(t, g, vsKey, hrname)
	integrationtest.VerifyMetadataHostRule(t, g, vsKey, "default/hr-cluster--foo.com-L7-dedicated", false)

	TearDownIngressForCacheSyncCheck(t, secretName, ingressName, svcName, modelName)

}

func TestApplySSLHostruleToInsecureDedicatedVS(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName := "admin/cluster--foo.com-L7-dedicated"
	hrname := "hr-cluster--foo.com-L7-dedicated"
	ingressName := objNameMap.GenerateName("foo-with-targets")
	svcName := objNameMap.GenerateName("avisvc")

	ingTestObj := IngressTestObject{
		ingressName: ingressName,
		isTLS:       false,
		withSecret:  false,
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
	hostrule.SslKeyCertificate = "thisisaviref-sslkey"
	hostrule.SslProfile = "thisisaviref-sslprof"
	hrObj := hostrule.HostRule()
	hrObj.Spec.VirtualHost.Fqdn = "foo.com"
	hrObj.Spec.VirtualHost.FqdnType = v1beta1.Contains
	hrObj.Spec.VirtualHost.TCPSettings = &v1beta1.HostRuleTCPSettings{
		LoadBalancerIP: "80.80.80.80",
	}

	if _, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().HostRules("default").Create(context.TODO(), hrObj, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding HostRule: %v", err)
	}

	g.Eventually(func() string {
		hostrule, _ := V1beta1Client.AkoV1beta1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 30*time.Second).Should(gomega.Equal("Accepted"))

	vsKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--foo.com-L7-dedicated"}
	integrationtest.VerifyMetadataHostRule(t, g, vsKey, "default/hr-cluster--foo.com-L7-dedicated", true)
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(*nodes[0].Enabled).To(gomega.Equal(true))
	g.Expect(nodes[0].PortProto).To(gomega.HaveLen(2))
	var portsWithHostRule []int
	for _, port := range nodes[0].PortProto {
		portsWithHostRule = append(portsWithHostRule, int(port.Port))
		if port.EnableSSL {
			g.Expect(int(port.Port)).To(gomega.Equal(443))
		}
	}
	sort.Ints(portsWithHostRule)
	g.Expect(portsWithHostRule[0]).To(gomega.Equal(80))
	g.Expect(portsWithHostRule[1]).To(gomega.Equal(443))
	g.Expect(nodes[0].VSVIPRefs[0].IPAddress).To(gomega.Equal("80.80.80.80"))

	integrationtest.TeardownHostRule(t, g, vsKey, hrname)
	integrationtest.VerifyMetadataHostRule(t, g, vsKey, "default/hr-cluster--foo.com-L7-dedicated", false)

	TearDownIngressForCacheSyncCheck(t, "", ingressName, svcName, modelName)

}

// AviInfraSetting CRD

func TestFQDNsCountForAviInfraSettingWithDedicatedShardSize(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	ingClassName := objNameMap.GenerateName("avi-lb")
	ingressName := objNameMap.GenerateName("foo-with-class")
	ns := "default"
	settingName := objNameMap.GenerateName("my-infrasetting")
	secretName := objNameMap.GenerateName("my-secret")
	svcName := objNameMap.GenerateName("avisvc")
	modelName := "admin/cluster--" + settingName + "-foo.com-L7-dedicated"

	SetUpTestForIngress(t, svcName, modelName)
	integrationtest.RemoveDefaultIngressClass()
	defer integrationtest.AddDefaultIngressClass()

	integrationtest.SetupAviInfraSetting(t, settingName, "DEDICATED")
	integrationtest.SetupIngressClass(t, ingClassName, lib.AviIngressController, settingName)
	integrationtest.AddSecret(secretName, ns, "tlsCert", "tlsKey")

	ingressCreate := (integrationtest.FakeIngress{
		Name:        ingressName,
		Namespace:   ns,
		ClassName:   ingClassName,
		DnsNames:    []string{"foo.com"},
		ServiceName: svcName,
		TlsSecretDNS: map[string][]string{
			secretName: {"foo.com"},
		},
	}).Ingress()
	_, err := KubeClient.NetworkingV1().Ingresses(ns).Create(context.TODO(), ingressCreate, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	g.Eventually(func() int {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if aviModel == nil {
			return 0
		}
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return len(nodes)
	}, 30*time.Second, 1*time.Second).Should(gomega.Equal(1))

	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	node := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0]

	g.Expect(node.VSVIPRefs).To(gomega.HaveLen(1))
	g.Expect(node.VSVIPRefs[0].FQDNs).To(gomega.HaveLen(1))
	for _, fqdn := range node.VSVIPRefs[0].FQDNs {
		g.Expect(fqdn).ShouldNot(gomega.ContainSubstring("L7-dedicated"))
	}
	integrationtest.TeardownAviInfraSetting(t, settingName)
	integrationtest.TeardownIngressClass(t, ingClassName)
	err = KubeClient.NetworkingV1().Ingresses(ns).Delete(context.TODO(), ingressName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	integrationtest.DeleteSecret(secretName, ns)

	mcache := cache.SharedAviObjCache()
	vsKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--" + settingName + "-foo.com-L7-dedicated"}
	// verify removal of VS.
	g.Eventually(func() bool {
		_, found := mcache.VsCacheMeta.AviCacheGet(vsKey)
		return found
	}, 50*time.Second, 5*time.Second).Should(gomega.Equal(false))
	TearDownTestForIngress(t, svcName, modelName)
}

func TestFQDNsCountForAviInfraSettingWithLargeShardSize(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	ingClassName := objNameMap.GenerateName("avi-lb")
	ingressName := objNameMap.GenerateName("foo-with-class")
	ns := "default"
	settingName := objNameMap.GenerateName("my-infrasetting")
	secretName := objNameMap.GenerateName("my-secret")
	svcName := objNameMap.GenerateName("avisvc")
	modelName := "admin/cluster--Shared-L7-" + settingName + "-0"

	SetUpTestForIngress(t, svcName, modelName)
	integrationtest.RemoveDefaultIngressClass()
	defer integrationtest.AddDefaultIngressClass()

	integrationtest.SetupAviInfraSetting(t, settingName, "LARGE")
	integrationtest.SetupIngressClass(t, ingClassName, lib.AviIngressController, settingName)
	integrationtest.AddSecret(secretName, ns, "tlsCert", "tlsKey")

	ingressCreate := (integrationtest.FakeIngress{
		Name:        ingressName,
		Namespace:   ns,
		ClassName:   ingClassName,
		DnsNames:    []string{"foo.com"},
		ServiceName: svcName,
		TlsSecretDNS: map[string][]string{
			secretName: {"foo.com"},
		},
	}).Ingress()
	_, err := KubeClient.NetworkingV1().Ingresses(ns).Create(context.TODO(), ingressCreate, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}

	g.Eventually(func() int {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if aviModel == nil {
			return 0
		}
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return len(nodes)
	}, 10*time.Second).Should(gomega.Equal(1))

	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	node := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0]

	g.Expect(node.VSVIPRefs).To(gomega.HaveLen(1))
	g.Expect(node.VSVIPRefs[0].FQDNs).To(gomega.HaveLen(2))
	for _, fqdn := range node.VSVIPRefs[0].FQDNs {
		if fqdn == "foo.com" {
			continue
		}
		g.Expect(fqdn).Should(gomega.ContainSubstring("Shared-L7"))
	}
	integrationtest.TeardownAviInfraSetting(t, settingName)
	integrationtest.TeardownIngressClass(t, ingClassName)
	err = KubeClient.NetworkingV1().Ingresses(ns).Delete(context.TODO(), ingressName, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	integrationtest.DeleteSecret(secretName, ns)

	mcache := cache.SharedAviObjCache()
	vsKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--Shared-L7-" + settingName + "-0"}
	// Shard VS remains, Pools are moved/removed
	g.Eventually(func() bool {
		sniCache1, found := mcache.VsCacheMeta.AviCacheGet(vsKey)
		sniCacheObj1, _ := sniCache1.(*cache.AviVsCache)
		if found {
			return len(sniCacheObj1.PoolKeyCollection) == 0
		}
		return false
	}, 50*time.Second).Should(gomega.Equal(true))
	TearDownTestForIngress(t, svcName, modelName)
}

func TestHostRuleUseRegex(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName := "admin/cluster--foo.com-L7-dedicated"
	hrname := "hr-cluster--foo.com-L7-dedicated"
	fqdn := "foo.com"
	namespace := "default"
	secretName := objNameMap.GenerateName("my-secret")
	ingressName := objNameMap.GenerateName("foo-with-targets")
	svcName := objNameMap.GenerateName("avisvc")

	ingTestObj := IngressTestObject{
		ingressName: ingressName,
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
		hostrule, _ := V1beta1Client.AkoV1beta1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 20*time.Second).Should(gomega.Equal("Accepted"))

	vsKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--foo.com-L7-dedicated"}
	integrationtest.VerifyMetadataHostRule(t, g, vsKey, "default/hr-cluster--foo.com-L7-dedicated", true)
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	node := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0]

	g.Expect(node.HttpPolicyRefs).To(gomega.HaveLen(2))
	g.Expect(node.HttpPolicyRefs[0].RedirectPorts).To(gomega.BeNil())
	g.Expect(node.HttpPolicyRefs[0].HppMap).To(gomega.HaveLen(1))
	g.Expect(node.HttpPolicyRefs[0].HppMap[0].MatchCriteria).Should(gomega.Equal("REGEX_MATCH"))
	g.Expect(node.HttpPolicyRefs[0].HppMap[0].MatchCase).Should(gomega.Equal("INSENSITIVE"))
	g.Expect(node.HttpPolicyRefs[0].HppMap[0].StringGroupRefs).To(gomega.HaveLen(1))
	g.Expect(node.HttpPolicyRefs[0].HppMap[0].PoolGroup).To(gomega.Equal(lib.GetEncodedSniPGPoolNameforRegex("cluster--default-foo.com_foo-" + ingressName + "-L7-dedicated")))
	g.Expect(node.HttpPolicyRefs[1].HppMap).To(gomega.BeNil())
	g.Expect(node.HttpPolicyRefs[1].RedirectPorts).To(gomega.HaveLen(1))
	g.Expect(node.HttpPolicyRefs[1].RedirectPorts[0].Path).To(gomega.Equal(""))
	g.Expect(node.HttpPolicyRefs[1].RedirectPorts[0].RedirectPort).To(gomega.Equal(int32(443)))
	g.Expect(node.HttpPolicyRefs[1].RedirectPorts[0].RedirectPath).To(gomega.Equal(""))
	g.Expect(node.HttpPolicyRefs[1].RedirectPorts[0].MatchCriteriaPath).To(gomega.Equal(""))
	g.Expect(node.HttpPolicyRefs[1].RedirectPorts[0].Hosts[0]).To(gomega.Equal("foo.com"))

	integrationtest.TeardownHostRule(t, g, vsKey, hrname)
	time.Sleep(2 * time.Second)

	g.Expect(node.HttpPolicyRefs).To(gomega.HaveLen(2))
	g.Expect(node.HttpPolicyRefs[0].RedirectPorts).To(gomega.BeNil())
	g.Expect(node.HttpPolicyRefs[0].HppMap).To(gomega.HaveLen(1))
	g.Expect(node.HttpPolicyRefs[0].HppMap[0].MatchCriteria).Should(gomega.Equal("BEGINS_WITH"))
	g.Expect(node.HttpPolicyRefs[0].HppMap[0].StringGroupRefs).To(gomega.HaveLen(0))
	g.Expect(node.HttpPolicyRefs[0].HppMap[0].PoolGroup).To(gomega.Equal("cluster--default-foo.com_foo-" + ingressName + "-L7-dedicated"))
	g.Expect(node.HttpPolicyRefs[1].HppMap).To(gomega.BeNil())
	g.Expect(node.HttpPolicyRefs[1].RedirectPorts).To(gomega.HaveLen(1))
	g.Expect(node.HttpPolicyRefs[1].RedirectPorts[0].Path).To(gomega.Equal(""))
	g.Expect(node.HttpPolicyRefs[1].RedirectPorts[0].RedirectPort).To(gomega.Equal(int32(443)))
	g.Expect(node.HttpPolicyRefs[1].RedirectPorts[0].RedirectPath).To(gomega.Equal(""))
	g.Expect(node.HttpPolicyRefs[1].RedirectPorts[0].MatchCriteriaPath).To(gomega.Equal(""))
	g.Expect(node.HttpPolicyRefs[1].RedirectPorts[0].Hosts[0]).To(gomega.Equal("foo.com"))

	TearDownIngressForCacheSyncCheck(t, secretName, ingressName, svcName, modelName)
}

func TestHostRuleAppRoot(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName := "admin/cluster--foo.com-L7-dedicated"
	hrname := "hr-cluster--foo.com-L7-dedicated"
	fqdn := "foo.com"
	namespace := "default"
	appRootPath := "/foo"
	secretName := objNameMap.GenerateName("my-secret")
	ingressName := objNameMap.GenerateName("app-root-test")
	svcName := objNameMap.GenerateName("avisvc")
	ingTestObj := IngressTestObject{
		ingressName: ingressName,
		isTLS:       true,
		withSecret:  true,
		secretName:  secretName,
		serviceName: svcName,
		modelNames:  []string{modelName},
		dnsNames:    []string{fqdn},
		paths:       []string{"/"},
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
		hostrule, _ := V1beta1Client.AkoV1beta1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 20*time.Second).Should(gomega.Equal("Accepted"))

	vsKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--foo.com-L7-dedicated"}
	integrationtest.VerifyMetadataHostRule(t, g, vsKey, "default/hr-cluster--foo.com-L7-dedicated", true)
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	node := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0]

	g.Expect(node.HttpPolicyRefs).To(gomega.HaveLen(2))
	g.Expect(node.HttpPolicyRefs[0].HppMap).To(gomega.HaveLen(1))
	g.Expect(node.HttpPolicyRefs[0].HppMap[0].MatchCriteria).Should(gomega.Equal("BEGINS_WITH"))
	g.Expect(node.HttpPolicyRefs[0].HppMap[0].Path[0]).Should(gomega.Equal(appRootPath))
	g.Expect(node.HttpPolicyRefs[0].HppMap[0].StringGroupRefs).To(gomega.HaveLen(0))
	g.Expect(node.HttpPolicyRefs[0].HppMap[0].PoolGroup).To(gomega.Equal("cluster--default-foo.com_-" + ingressName + "-L7-dedicated"))
	g.Expect(node.HttpPolicyRefs[0].RedirectPorts).To(gomega.HaveLen(2))
	g.Expect(node.HttpPolicyRefs[0].RedirectPorts[0].Path).To(gomega.Equal("/"))
	g.Expect(node.HttpPolicyRefs[0].RedirectPorts[0].Protocol).To(gomega.Equal("HTTP"))
	g.Expect(node.HttpPolicyRefs[0].RedirectPorts[0].RedirectPort).To(gomega.Equal(int32(80)))
	g.Expect(node.HttpPolicyRefs[0].RedirectPorts[0].RedirectPath).To(gomega.Equal(strings.TrimPrefix(appRootPath, "/")))
	g.Expect(node.HttpPolicyRefs[0].RedirectPorts[0].MatchCriteriaPath).To(gomega.Equal("EQUALS"))
	g.Expect(node.HttpPolicyRefs[0].RedirectPorts[1].Path).To(gomega.Equal("/"))
	g.Expect(node.HttpPolicyRefs[0].RedirectPorts[1].Protocol).To(gomega.Equal("HTTPS"))
	g.Expect(node.HttpPolicyRefs[0].RedirectPorts[1].RedirectPort).To(gomega.Equal(int32(443)))
	g.Expect(node.HttpPolicyRefs[0].RedirectPorts[1].RedirectPath).To(gomega.Equal(strings.TrimPrefix(appRootPath, "/")))
	g.Expect(node.HttpPolicyRefs[0].RedirectPorts[1].MatchCriteriaPath).To(gomega.Equal("EQUALS"))
	g.Expect(node.HttpPolicyRefs[0].RedirectPorts[1].MatchCriteriaPort).To(gomega.Equal("IS_IN"))
	g.Expect(node.HttpPolicyRefs[1].RedirectPorts).To(gomega.HaveLen(1))
	g.Expect(node.HttpPolicyRefs[1].RedirectPorts[0].Hosts).Should(gomega.ContainElements(fqdn))
	g.Expect(node.HttpPolicyRefs[1].HppMap).To(gomega.BeNil())

	integrationtest.TeardownHostRule(t, g, vsKey, hrname)
	g.Eventually(func() bool {
		return node.HttpPolicyRefs[0].RedirectPorts == nil
	}, 30*time.Second).Should(gomega.BeTrue())

	g.Expect(node.HttpPolicyRefs).To(gomega.HaveLen(2))
	g.Expect(node.HttpPolicyRefs[0].RedirectPorts).To(gomega.BeNil())
	g.Expect(node.HttpPolicyRefs[0].HppMap).To(gomega.HaveLen(1))
	g.Expect(node.HttpPolicyRefs[0].HppMap[0].MatchCriteria).Should(gomega.Equal("BEGINS_WITH"))
	g.Expect(node.HttpPolicyRefs[0].HppMap[0].Path[0]).To(gomega.Equal("/"))
	g.Expect(node.HttpPolicyRefs[0].HppMap[0].StringGroupRefs).To(gomega.HaveLen(0))
	g.Expect(node.HttpPolicyRefs[0].HppMap[0].PoolGroup).To(gomega.Equal("cluster--default-foo.com_-" + ingressName + "-L7-dedicated"))
	g.Expect(node.HttpPolicyRefs[1].RedirectPorts).To(gomega.HaveLen(1))
	g.Expect(node.HttpPolicyRefs[1].RedirectPorts[0].Hosts).Should(gomega.ContainElements(fqdn))
	g.Expect(node.HttpPolicyRefs[1].HppMap).To(gomega.BeNil())

	TearDownIngressForCacheSyncCheck(t, secretName, ingressName, svcName, modelName)
}

func TestHostRuleRegexAppRoot(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName := "admin/cluster--foo.com-L7-dedicated"
	hrname := "hr-cluster--foo.com-L7-dedicated"
	fqdn := "foo.com"
	namespace := "default"
	appRootPath := "/foo"
	secretName := objNameMap.GenerateName("my-secret")
	ingressName := objNameMap.GenerateName("app-root-test")
	svcName := objNameMap.GenerateName("avisvc")
	ingTestObj := IngressTestObject{
		ingressName: ingressName,
		isTLS:       true,
		withSecret:  true,
		secretName:  secretName,
		serviceName: svcName,
		modelNames:  []string{modelName},
		dnsNames:    []string{fqdn, fqdn},
		paths:       []string{"/something(/|$)(.*)", "/"},
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
		hostrule, _ := V1beta1Client.AkoV1beta1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 20*time.Second).Should(gomega.Equal("Accepted"))

	vsKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--foo.com-L7-dedicated"}
	integrationtest.VerifyMetadataHostRule(t, g, vsKey, "default/hr-cluster--foo.com-L7-dedicated", true)
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	node := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0]

	g.Expect(node.HttpPolicyRefs).To(gomega.HaveLen(2))
	g.Expect(node.HttpPolicyRefs[0].HppMap).To(gomega.HaveLen(2))
	g.Expect(node.HttpPolicyRefs[0].HppMap[0].MatchCriteria).Should(gomega.Equal("REGEX_MATCH"))
	g.Expect(node.HttpPolicyRefs[0].HppMap[0].MatchCase).Should(gomega.Equal("INSENSITIVE"))
	g.Expect(node.HttpPolicyRefs[0].HppMap[0].StringGroupRefs).To(gomega.HaveLen(1))
	g.Expect(node.HttpPolicyRefs[0].HppMap[0].PoolGroup).To(gomega.Equal(lib.GetEncodedSniPGPoolNameforRegex("cluster--default-foo.com_something(_|$)(.*)-" + ingressName + "-L7-dedicated")))
	g.Expect(node.HttpPolicyRefs[0].HppMap[1].MatchCriteria).Should(gomega.Equal("REGEX_MATCH"))
	g.Expect(node.HttpPolicyRefs[0].HppMap[1].MatchCase).Should(gomega.Equal("INSENSITIVE"))
	g.Expect(node.HttpPolicyRefs[0].HppMap[1].Path[0]).To(gomega.Equal(appRootPath))
	g.Expect(node.HttpPolicyRefs[0].HppMap[1].StringGroupRefs).To(gomega.HaveLen(1))
	g.Expect(node.HttpPolicyRefs[0].HppMap[1].PoolGroup).To(gomega.Equal(lib.GetEncodedSniPGPoolNameforRegex("cluster--default-foo.com_-" + ingressName + "-L7-dedicated")))
	g.Expect(node.HttpPolicyRefs[0].RedirectPorts).To(gomega.HaveLen(2))
	g.Expect(node.HttpPolicyRefs[0].RedirectPorts[0].Path).To(gomega.Equal("/"))
	g.Expect(node.HttpPolicyRefs[0].RedirectPorts[0].RedirectPort).To(gomega.Equal(int32(80)))
	g.Expect(node.HttpPolicyRefs[0].RedirectPorts[0].Protocol).To(gomega.Equal("HTTP"))
	g.Expect(node.HttpPolicyRefs[0].RedirectPorts[0].RedirectPath).To(gomega.Equal(strings.TrimPrefix(appRootPath, "/")))
	g.Expect(node.HttpPolicyRefs[0].RedirectPorts[0].MatchCriteriaPath).To(gomega.Equal("EQUALS"))
	g.Expect(node.HttpPolicyRefs[0].RedirectPorts[1].MatchCriteriaPort).To(gomega.Equal("IS_IN"))
	g.Expect(node.HttpPolicyRefs[0].RedirectPorts[1].Path).To(gomega.Equal("/"))
	g.Expect(node.HttpPolicyRefs[0].RedirectPorts[1].Protocol).To(gomega.Equal("HTTPS"))
	g.Expect(node.HttpPolicyRefs[0].RedirectPorts[1].RedirectPort).To(gomega.Equal(int32(443)))
	g.Expect(node.HttpPolicyRefs[0].RedirectPorts[1].RedirectPath).To(gomega.Equal(strings.TrimPrefix(appRootPath, "/")))
	g.Expect(node.HttpPolicyRefs[0].RedirectPorts[1].MatchCriteriaPath).To(gomega.Equal("EQUALS"))
	g.Expect(node.HttpPolicyRefs[0].RedirectPorts[1].MatchCriteriaPort).To(gomega.Equal("IS_IN"))
	g.Expect(node.HttpPolicyRefs[1].RedirectPorts).To(gomega.HaveLen(1))
	g.Expect(node.HttpPolicyRefs[1].RedirectPorts[0].Hosts).Should(gomega.ContainElements(fqdn))
	g.Expect(node.HttpPolicyRefs[1].HppMap).To(gomega.BeNil())

	integrationtest.TeardownHostRule(t, g, vsKey, hrname)
	g.Eventually(func() bool {
		return node.HttpPolicyRefs[0].RedirectPorts == nil
	}, 30*time.Second).Should(gomega.BeTrue())

	g.Expect(node.HttpPolicyRefs).To(gomega.HaveLen(2))
	g.Expect(node.HttpPolicyRefs[0].RedirectPorts).To(gomega.BeNil())
	g.Expect(node.HttpPolicyRefs[0].HppMap).To(gomega.HaveLen(2))
	g.Expect(node.HttpPolicyRefs[0].HppMap[0].MatchCriteria).Should(gomega.Equal("BEGINS_WITH"))
	g.Expect(node.HttpPolicyRefs[0].HppMap[0].Path[0]).To(gomega.Equal("/something(/|$)(.*)"))
	g.Expect(node.HttpPolicyRefs[0].HppMap[0].StringGroupRefs).To(gomega.HaveLen(0))
	g.Expect(node.HttpPolicyRefs[0].HppMap[0].PoolGroup).To(gomega.Equal("cluster--default-foo.com_something(_|$)(.*)-" + ingressName + "-L7-dedicated"))
	g.Expect(node.HttpPolicyRefs[0].HppMap[1].MatchCriteria).Should(gomega.Equal("BEGINS_WITH"))
	g.Expect(node.HttpPolicyRefs[0].HppMap[1].Path[0]).To(gomega.Equal("/"))
	g.Expect(node.HttpPolicyRefs[0].HppMap[1].StringGroupRefs).To(gomega.HaveLen(0))
	g.Expect(node.HttpPolicyRefs[0].HppMap[1].PoolGroup).To(gomega.Equal("cluster--default-foo.com_-" + ingressName + "-L7-dedicated"))
	g.Expect(node.HttpPolicyRefs[1].RedirectPorts).To(gomega.HaveLen(1))
	g.Expect(node.HttpPolicyRefs[1].RedirectPorts[0].Hosts).Should(gomega.ContainElements(fqdn))
	g.Expect(node.HttpPolicyRefs[1].HppMap).To(gomega.BeNil())

	TearDownIngressForCacheSyncCheck(t, secretName, ingressName, svcName, modelName)
}

func TestHostRuleAppRootListenerPorts(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName := "admin/cluster--foo.com-L7-dedicated"
	hrname := "hr-cluster--foo.com-L7-dedicated"
	fqdn := "foo.com"
	namespace := "default"
	appRootPath := "/foo"
	secretName := objNameMap.GenerateName("my-secret")
	ingressName := objNameMap.GenerateName("app-root-test")
	svcName := objNameMap.GenerateName("avisvc")
	ingTestObj := IngressTestObject{
		ingressName: ingressName,
		isTLS:       true,
		withSecret:  true,
		secretName:  secretName,
		serviceName: svcName,
		modelNames:  []string{modelName},
		dnsNames:    []string{fqdn},
		paths:       []string{"/"},
	}
	ingTestObj.FillParams()
	SetUpIngressForCacheSyncCheck(t, ingTestObj)

	hostrule := integrationtest.FakeHostRule{
		Name:                hrname,
		Namespace:           namespace,
		Fqdn:                fqdn,
		ApplicationRootPath: appRootPath,
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
	hrCreate := hostrule.HostRule()
	if _, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().HostRules(namespace).Create(context.TODO(), hrCreate, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding HostRule: %v", err)
	}
	g.Eventually(func() string {
		hostrule, _ := V1beta1Client.AkoV1beta1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 20*time.Second).Should(gomega.Equal("Accepted"))

	vsKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--foo.com-L7-dedicated"}
	integrationtest.VerifyMetadataHostRule(t, g, vsKey, "default/hr-cluster--foo.com-L7-dedicated", true)
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	node := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0]

	g.Expect(node.HttpPolicyRefs).To(gomega.HaveLen(2))
	g.Expect(node.HttpPolicyRefs[0].HppMap).To(gomega.HaveLen(1))
	g.Expect(node.HttpPolicyRefs[0].HppMap[0].MatchCriteria).Should(gomega.Equal("BEGINS_WITH"))
	g.Expect(node.HttpPolicyRefs[0].HppMap[0].Path[0]).Should(gomega.Equal(appRootPath))
	g.Expect(node.HttpPolicyRefs[0].HppMap[0].StringGroupRefs).To(gomega.HaveLen(0))
	g.Expect(node.HttpPolicyRefs[0].HppMap[0].PoolGroup).To(gomega.Equal("cluster--default-foo.com_-" + ingressName + "-L7-dedicated"))
	g.Expect(node.HttpPolicyRefs[0].RedirectPorts).To(gomega.HaveLen(2))
	g.Expect(node.HttpPolicyRefs[0].RedirectPorts[0].Path).To(gomega.Equal("/"))
	g.Expect(node.HttpPolicyRefs[0].RedirectPorts[0].Protocol).To(gomega.Equal("HTTP"))
	g.Expect(node.HttpPolicyRefs[0].RedirectPorts[0].RedirectPort).To(gomega.Equal(int32(8081)))
	g.Expect(node.HttpPolicyRefs[0].RedirectPorts[0].RedirectPath).To(gomega.Equal(strings.TrimPrefix(appRootPath, "/")))
	g.Expect(node.HttpPolicyRefs[0].RedirectPorts[0].MatchCriteriaPath).To(gomega.Equal("EQUALS"))
	g.Expect(node.HttpPolicyRefs[0].RedirectPorts[1].Path).To(gomega.Equal("/"))
	g.Expect(node.HttpPolicyRefs[0].RedirectPorts[1].Protocol).To(gomega.Equal("HTTPS"))
	g.Expect(node.HttpPolicyRefs[0].RedirectPorts[1].RedirectPort).To(gomega.Equal(int32(6443)))
	g.Expect(node.HttpPolicyRefs[0].RedirectPorts[1].RedirectPath).To(gomega.Equal(strings.TrimPrefix(appRootPath, "/")))
	g.Expect(node.HttpPolicyRefs[0].RedirectPorts[1].MatchCriteriaPath).To(gomega.Equal("EQUALS"))
	g.Expect(node.HttpPolicyRefs[0].RedirectPorts[1].MatchCriteriaPort).To(gomega.Equal("IS_IN"))
	g.Expect(node.HttpPolicyRefs[1].RedirectPorts).To(gomega.HaveLen(1))
	g.Expect(node.HttpPolicyRefs[1].RedirectPorts[0].Hosts).Should(gomega.ContainElements(fqdn))
	g.Expect(node.HttpPolicyRefs[1].HppMap).To(gomega.BeNil())

	integrationtest.TeardownHostRule(t, g, vsKey, hrname)
	g.Eventually(func() bool {
		return node.HttpPolicyRefs[0].RedirectPorts == nil
	}, 30*time.Second).Should(gomega.BeTrue())

	g.Expect(node.HttpPolicyRefs).To(gomega.HaveLen(2))
	g.Expect(node.HttpPolicyRefs[0].RedirectPorts).To(gomega.BeNil())
	g.Expect(node.HttpPolicyRefs[0].HppMap).To(gomega.HaveLen(1))
	g.Expect(node.HttpPolicyRefs[0].HppMap[0].MatchCriteria).Should(gomega.Equal("BEGINS_WITH"))
	g.Expect(node.HttpPolicyRefs[0].HppMap[0].Path[0]).To(gomega.Equal("/"))
	g.Expect(node.HttpPolicyRefs[0].HppMap[0].StringGroupRefs).To(gomega.HaveLen(0))
	g.Expect(node.HttpPolicyRefs[0].HppMap[0].PoolGroup).To(gomega.Equal("cluster--default-foo.com_-" + ingressName + "-L7-dedicated"))
	g.Expect(node.HttpPolicyRefs[1].RedirectPorts).To(gomega.HaveLen(1))
	g.Expect(node.HttpPolicyRefs[1].RedirectPorts[0].Hosts).Should(gomega.ContainElements(fqdn))
	g.Expect(node.HttpPolicyRefs[1].HppMap).To(gomega.BeNil())

	TearDownIngressForCacheSyncCheck(t, secretName, ingressName, svcName, modelName)
}

func TestHostRuleRegexAppRootListenerPorts(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName := "admin/cluster--foo.com-L7-dedicated"
	hrname := "hr-cluster--foo.com-L7-dedicated"
	fqdn := "foo.com"
	namespace := "default"
	appRootPath := "/foo"
	secretName := objNameMap.GenerateName("my-secret")
	ingressName := objNameMap.GenerateName("app-root-test")
	svcName := objNameMap.GenerateName("avisvc")
	ingTestObj := IngressTestObject{
		ingressName: ingressName,
		isTLS:       true,
		withSecret:  true,
		secretName:  secretName,
		serviceName: svcName,
		modelNames:  []string{modelName},
		dnsNames:    []string{fqdn, fqdn},
		paths:       []string{"/something(/|$)(.*)", "/"},
	}
	ingTestObj.FillParams()
	SetUpIngressForCacheSyncCheck(t, ingTestObj)

	hostrule := integrationtest.FakeHostRule{
		Name:                hrname,
		Namespace:           namespace,
		Fqdn:                fqdn,
		ApplicationRootPath: appRootPath,
		UseRegex:            true,
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
	hrCreate := hostrule.HostRule()
	if _, err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().HostRules(namespace).Create(context.TODO(), hrCreate, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding HostRule: %v", err)
	}
	g.Eventually(func() string {
		hostrule, _ := V1beta1Client.AkoV1beta1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 20*time.Second).Should(gomega.Equal("Accepted"))

	vsKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--foo.com-L7-dedicated"}
	integrationtest.VerifyMetadataHostRule(t, g, vsKey, "default/hr-cluster--foo.com-L7-dedicated", true)
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	node := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0]

	g.Expect(node.HttpPolicyRefs).To(gomega.HaveLen(2))
	g.Expect(node.HttpPolicyRefs[0].HppMap).To(gomega.HaveLen(2))
	g.Expect(node.HttpPolicyRefs[0].HppMap[0].MatchCriteria).Should(gomega.Equal("REGEX_MATCH"))
	g.Expect(node.HttpPolicyRefs[0].HppMap[0].MatchCase).Should(gomega.Equal("INSENSITIVE"))
	g.Expect(node.HttpPolicyRefs[0].HppMap[0].StringGroupRefs).To(gomega.HaveLen(1))
	g.Expect(node.HttpPolicyRefs[0].HppMap[0].PoolGroup).To(gomega.Equal(lib.GetEncodedSniPGPoolNameforRegex("cluster--default-foo.com_something(_|$)(.*)-" + ingressName + "-L7-dedicated")))
	g.Expect(node.HttpPolicyRefs[0].HppMap[1].MatchCriteria).Should(gomega.Equal("REGEX_MATCH"))
	g.Expect(node.HttpPolicyRefs[0].HppMap[1].MatchCase).Should(gomega.Equal("INSENSITIVE"))
	g.Expect(node.HttpPolicyRefs[0].HppMap[1].Path[0]).To(gomega.Equal(appRootPath))
	g.Expect(node.HttpPolicyRefs[0].HppMap[1].StringGroupRefs).To(gomega.HaveLen(1))
	g.Expect(node.HttpPolicyRefs[0].HppMap[1].PoolGroup).To(gomega.Equal(lib.GetEncodedSniPGPoolNameforRegex("cluster--default-foo.com_-" + ingressName + "-L7-dedicated")))
	g.Expect(node.HttpPolicyRefs[0].RedirectPorts).To(gomega.HaveLen(2))
	g.Expect(node.HttpPolicyRefs[0].RedirectPorts[0].Path).To(gomega.Equal("/"))
	g.Expect(node.HttpPolicyRefs[0].RedirectPorts[0].RedirectPort).To(gomega.Equal(int32(8081)))
	g.Expect(node.HttpPolicyRefs[0].RedirectPorts[0].Protocol).To(gomega.Equal("HTTP"))
	g.Expect(node.HttpPolicyRefs[0].RedirectPorts[0].RedirectPath).To(gomega.Equal(strings.TrimPrefix(appRootPath, "/")))
	g.Expect(node.HttpPolicyRefs[0].RedirectPorts[0].MatchCriteriaPath).To(gomega.Equal("EQUALS"))
	g.Expect(node.HttpPolicyRefs[0].RedirectPorts[1].MatchCriteriaPort).To(gomega.Equal("IS_IN"))
	g.Expect(node.HttpPolicyRefs[0].RedirectPorts[1].Path).To(gomega.Equal("/"))
	g.Expect(node.HttpPolicyRefs[0].RedirectPorts[1].Protocol).To(gomega.Equal("HTTPS"))
	g.Expect(node.HttpPolicyRefs[0].RedirectPorts[1].RedirectPort).To(gomega.Equal(int32(6443)))
	g.Expect(node.HttpPolicyRefs[0].RedirectPorts[1].RedirectPath).To(gomega.Equal(strings.TrimPrefix(appRootPath, "/")))
	g.Expect(node.HttpPolicyRefs[0].RedirectPorts[1].MatchCriteriaPath).To(gomega.Equal("EQUALS"))
	g.Expect(node.HttpPolicyRefs[0].RedirectPorts[1].MatchCriteriaPort).To(gomega.Equal("IS_IN"))
	g.Expect(node.HttpPolicyRefs[1].RedirectPorts).To(gomega.HaveLen(1))
	g.Expect(node.HttpPolicyRefs[1].RedirectPorts[0].Hosts).Should(gomega.ContainElements(fqdn))
	g.Expect(node.HttpPolicyRefs[1].HppMap).To(gomega.BeNil())

	integrationtest.TeardownHostRule(t, g, vsKey, hrname)
	g.Eventually(func() bool {
		return node.HttpPolicyRefs[0].RedirectPorts == nil
	}, 30*time.Second).Should(gomega.BeTrue())

	g.Expect(node.HttpPolicyRefs).To(gomega.HaveLen(2))
	g.Expect(node.HttpPolicyRefs[0].RedirectPorts).To(gomega.BeNil())
	g.Expect(node.HttpPolicyRefs[0].HppMap).To(gomega.HaveLen(2))
	g.Expect(node.HttpPolicyRefs[0].HppMap[0].MatchCriteria).Should(gomega.Equal("BEGINS_WITH"))
	g.Expect(node.HttpPolicyRefs[0].HppMap[0].Path[0]).To(gomega.Equal("/something(/|$)(.*)"))
	g.Expect(node.HttpPolicyRefs[0].HppMap[0].StringGroupRefs).To(gomega.HaveLen(0))
	g.Expect(node.HttpPolicyRefs[0].HppMap[0].PoolGroup).To(gomega.Equal("cluster--default-foo.com_something(_|$)(.*)-" + ingressName + "-L7-dedicated"))
	g.Expect(node.HttpPolicyRefs[0].HppMap[1].MatchCriteria).Should(gomega.Equal("BEGINS_WITH"))
	g.Expect(node.HttpPolicyRefs[0].HppMap[1].Path[0]).To(gomega.Equal("/"))
	g.Expect(node.HttpPolicyRefs[0].HppMap[1].StringGroupRefs).To(gomega.HaveLen(0))
	g.Expect(node.HttpPolicyRefs[0].HppMap[1].PoolGroup).To(gomega.Equal("cluster--default-foo.com_-" + ingressName + "-L7-dedicated"))
	g.Expect(node.HttpPolicyRefs[1].RedirectPorts).To(gomega.HaveLen(1))
	g.Expect(node.HttpPolicyRefs[1].RedirectPorts[0].Hosts).Should(gomega.ContainElements(fqdn))
	g.Expect(node.HttpPolicyRefs[1].HppMap).To(gomega.BeNil())

	TearDownIngressForCacheSyncCheck(t, secretName, ingressName, svcName, modelName)
}

func TestHostRuleUseRegexNoPath(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName := "admin/cluster--foo.com-L7-dedicated"
	hrname := "hr-cluster--foo.com-L7-dedicated"
	fqdn := "foo.com"
	namespace := "default"
	secretName := objNameMap.GenerateName("my-secret")
	ingressName := objNameMap.GenerateName("foo-with-targets")
	svcName := objNameMap.GenerateName("avisvc")

	ingTestObj := IngressTestObject{
		ingressName: ingressName,
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
		hostrule, _ := V1beta1Client.AkoV1beta1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 20*time.Second).Should(gomega.Equal("Accepted"))

	vsKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--foo.com-L7-dedicated"}
	integrationtest.VerifyMetadataHostRule(t, g, vsKey, "default/hr-cluster--foo.com-L7-dedicated", true)
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	node := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0]

	g.Expect(node.HttpPolicyRefs).To(gomega.HaveLen(2))
	g.Expect(node.HttpPolicyRefs[0].RedirectPorts).To(gomega.BeNil())
	g.Expect(node.HttpPolicyRefs[0].HppMap).To(gomega.HaveLen(1))
	g.Expect(node.HttpPolicyRefs[0].HppMap[0].MatchCriteria).Should(gomega.Equal("BEGINS_WITH"))
	g.Expect(node.HttpPolicyRefs[0].HppMap[0].StringGroupRefs).To(gomega.BeNil())
	g.Expect(node.HttpPolicyRefs[0].HppMap[0].PoolGroup).To(gomega.Equal(lib.GetEncodedSniPGPoolNameforRegex("cluster--default-foo.com-" + ingressName + "-L7-dedicated")))
	g.Expect(node.HttpPolicyRefs[1].HppMap).To(gomega.BeNil())
	g.Expect(node.HttpPolicyRefs[1].RedirectPorts).To(gomega.HaveLen(1))
	g.Expect(node.HttpPolicyRefs[1].RedirectPorts[0].Path).To(gomega.Equal(""))
	g.Expect(node.HttpPolicyRefs[1].RedirectPorts[0].RedirectPort).To(gomega.Equal(int32(443)))
	g.Expect(node.HttpPolicyRefs[1].RedirectPorts[0].RedirectPath).To(gomega.Equal(""))
	g.Expect(node.HttpPolicyRefs[1].RedirectPorts[0].MatchCriteriaPath).To(gomega.Equal(""))
	g.Expect(node.HttpPolicyRefs[1].RedirectPorts[0].Hosts[0]).To(gomega.Equal("foo.com"))

	integrationtest.TeardownHostRule(t, g, vsKey, hrname)
	time.Sleep(2 * time.Second)

	g.Expect(node.HttpPolicyRefs).To(gomega.HaveLen(2))
	g.Expect(node.HttpPolicyRefs[0].RedirectPorts).To(gomega.BeNil())
	g.Expect(node.HttpPolicyRefs[0].HppMap).To(gomega.HaveLen(1))
	g.Expect(node.HttpPolicyRefs[0].HppMap[0].MatchCriteria).Should(gomega.Equal("BEGINS_WITH"))
	g.Expect(node.HttpPolicyRefs[0].HppMap[0].StringGroupRefs).To(gomega.HaveLen(0))
	g.Expect(node.HttpPolicyRefs[0].HppMap[0].PoolGroup).To(gomega.Equal("cluster--default-foo.com-" + ingressName + "-L7-dedicated"))
	g.Expect(node.HttpPolicyRefs[1].HppMap).To(gomega.BeNil())
	g.Expect(node.HttpPolicyRefs[1].RedirectPorts).To(gomega.HaveLen(1))
	g.Expect(node.HttpPolicyRefs[1].RedirectPorts[0].Path).To(gomega.Equal(""))
	g.Expect(node.HttpPolicyRefs[1].RedirectPorts[0].RedirectPort).To(gomega.Equal(int32(443)))
	g.Expect(node.HttpPolicyRefs[1].RedirectPorts[0].RedirectPath).To(gomega.Equal(""))
	g.Expect(node.HttpPolicyRefs[1].RedirectPorts[0].MatchCriteriaPath).To(gomega.Equal(""))
	g.Expect(node.HttpPolicyRefs[1].RedirectPorts[0].Hosts[0]).To(gomega.Equal("foo.com"))

	TearDownIngressForCacheSyncCheck(t, secretName, ingressName, svcName, modelName)
}

func TestHostRuleUseRegexMultiIngressNoPath(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName := "admin/cluster--foo.com-L7-dedicated"
	hrname := "hr-cluster--foo.com-L7-dedicated"
	fqdn := "foo.com"
	namespace := "default"
	secretName := objNameMap.GenerateName("my-secret")
	ingressName := objNameMap.GenerateName("foo-with-targets")
	svcName := objNameMap.GenerateName("avisvc")

	ingTestObj := IngressTestObject{
		ingressName: ingressName,
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
		hostrule, _ := V1beta1Client.AkoV1beta1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 20*time.Second).Should(gomega.Equal("Accepted"))

	vsKey := cache.NamespaceName{Namespace: "admin", Name: "cluster--foo.com-L7-dedicated"}
	integrationtest.VerifyMetadataHostRule(t, g, vsKey, "default/hr-cluster--foo.com-L7-dedicated", true)

	g.Eventually(func() bool {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return nodes[0].HttpPolicyRefs[0].HppMap[1].MatchCriteria == "REGEX_MATCH" || nodes[0].HttpPolicyRefs[0].HppMap[0].MatchCriteria == "REGEX_MATCH"
	}, 10*time.Second).Should(gomega.Equal(true))

	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	node := aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0]

	g.Expect(node.HttpPolicyRefs).To(gomega.HaveLen(2))
	g.Expect(node.HttpPolicyRefs[0].RedirectPorts).To(gomega.BeNil())
	g.Expect(node.HttpPolicyRefs[0].HppMap).To(gomega.HaveLen(2))
	hppMapRegexMatch := node.HttpPolicyRefs[0].HppMap[1]
	hppMapBeginsWith := node.HttpPolicyRefs[0].HppMap[0]
	if node.HttpPolicyRefs[0].HppMap[0].MatchCriteria == "REGEX_MATCH" {
		hppMapRegexMatch = node.HttpPolicyRefs[0].HppMap[0]
		hppMapBeginsWith = node.HttpPolicyRefs[0].HppMap[1]
	}
	g.Expect(hppMapBeginsWith.MatchCriteria).Should(gomega.Equal("BEGINS_WITH"))
	g.Expect(hppMapBeginsWith.StringGroupRefs).To(gomega.BeNil())
	g.Expect(hppMapBeginsWith.PoolGroup).To(gomega.Equal(lib.GetEncodedSniPGPoolNameforRegex("cluster--default-foo.com-" + ingressName + "-L7-dedicated")))
	g.Expect(hppMapRegexMatch.MatchCriteria).Should(gomega.Equal("REGEX_MATCH"))
	g.Expect(hppMapRegexMatch.MatchCase).Should(gomega.Equal("INSENSITIVE"))
	g.Expect(hppMapRegexMatch.StringGroupRefs).To(gomega.HaveLen(1))
	g.Expect(hppMapRegexMatch.PoolGroup).To(gomega.Equal(lib.GetEncodedSniPGPoolNameforRegex("cluster--default-foo.com_foo-" + ingressName2 + "-L7-dedicated")))
	g.Expect(node.HttpPolicyRefs[1].HppMap).To(gomega.BeNil())
	g.Expect(node.HttpPolicyRefs[1].RedirectPorts).To(gomega.HaveLen(1))
	g.Expect(node.HttpPolicyRefs[1].RedirectPorts[0].Path).To(gomega.Equal(""))
	g.Expect(node.HttpPolicyRefs[1].RedirectPorts[0].RedirectPort).To(gomega.Equal(int32(443)))
	g.Expect(node.HttpPolicyRefs[1].RedirectPorts[0].RedirectPath).To(gomega.Equal(""))
	g.Expect(node.HttpPolicyRefs[1].RedirectPorts[0].MatchCriteriaPath).To(gomega.Equal(""))
	g.Expect(node.HttpPolicyRefs[1].RedirectPorts[0].Hosts[0]).To(gomega.Equal("foo.com"))

	integrationtest.TeardownHostRule(t, g, vsKey, hrname)

	g.Eventually(func() bool {
		_, aviModel = objects.SharedAviGraphLister().Get(modelName)
		node = aviModel.(*avinodes.AviObjectGraph).GetAviVS()[0]
		return (node.HttpPolicyRefs[0].HppMap[0].PoolGroup == "cluster--default-foo.com-"+ingressName+"-L7-dedicated" || node.HttpPolicyRefs[0].HppMap[1].PoolGroup == "cluster--default-foo.com-"+ingressName+"-L7-dedicated") &&
			(node.HttpPolicyRefs[0].HppMap[0].PoolGroup == "cluster--default-foo.com_foo-"+ingressName2+"-L7-dedicated" || node.HttpPolicyRefs[0].HppMap[1].PoolGroup == "cluster--default-foo.com_foo-"+ingressName2+"-L7-dedicated")
	}, 10*time.Second).Should(gomega.Equal(true))

	g.Expect(node.HttpPolicyRefs).To(gomega.HaveLen(2))
	g.Expect(node.HttpPolicyRefs[0].RedirectPorts).To(gomega.BeNil())
	g.Expect(node.HttpPolicyRefs[0].HppMap).To(gomega.HaveLen(2))
	g.Expect(node.HttpPolicyRefs[0].HppMap[0].MatchCriteria).Should(gomega.Equal("BEGINS_WITH"))
	g.Expect(node.HttpPolicyRefs[0].HppMap[0].StringGroupRefs).To(gomega.HaveLen(0))
	g.Expect(node.HttpPolicyRefs[0].HppMap[1].MatchCriteria).Should(gomega.Equal("BEGINS_WITH"))
	g.Expect(node.HttpPolicyRefs[0].HppMap[1].StringGroupRefs).To(gomega.HaveLen(0))
	g.Expect(node.HttpPolicyRefs[1].HppMap).To(gomega.BeNil())
	g.Expect(node.HttpPolicyRefs[1].RedirectPorts).To(gomega.HaveLen(1))
	g.Expect(node.HttpPolicyRefs[1].RedirectPorts[0].Path).To(gomega.Equal(""))
	g.Expect(node.HttpPolicyRefs[1].RedirectPorts[0].RedirectPort).To(gomega.Equal(int32(443)))
	g.Expect(node.HttpPolicyRefs[1].RedirectPorts[0].RedirectPath).To(gomega.Equal(""))
	g.Expect(node.HttpPolicyRefs[1].RedirectPorts[0].MatchCriteriaPath).To(gomega.Equal(""))
	g.Expect(node.HttpPolicyRefs[1].RedirectPorts[0].Hosts[0]).To(gomega.Equal("foo.com"))

	// deleting the second created ingress
	if err := KubeClient.NetworkingV1().Ingresses("default").Delete(context.TODO(), ingressName2, metav1.DeleteOptions{}); err != nil {
		t.Fatalf("Couldn't DELETE the Ingress %v", err)
	}
	TearDownIngressForCacheSyncCheck(t, secretName, ingressName, svcName, modelName)
}

func TestHTTPRuleCreateDeleteEnableHTTP2(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName := "admin/cluster--foo.com-L7-dedicated"
	svcName := "avisvc"
	rrname := "samplerr-foo"
	secretName := "my-secret"
	ingName := "foo-with-targets"

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
	g.Expect(*nodes[0].PoolRefs[0].EnableHttp2).To(gomega.Equal(true))

	// delete httprule disables HTTP2
	integrationtest.TeardownHTTPRule(t, rrname)
	integrationtest.VerifyMetadataHTTPRule(t, g, poolFooKey, "default/"+rrname+"/"+httpRulePath, false)
	integrationtest.VerifyMetadataHTTPRule(t, g, poolBarKey, "default/"+rrname+"/"+httpRulePath, false)
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes[0].PoolRefs[0].EnableHttp2).To(gomega.BeNil())

	TearDownIngressForCacheSyncCheck(t, secretName, ingName, svcName, modelName)
}

/*
This test case tests following scenario
1. Create an ingress with foo.com
2. Create Hostrule with Fqdn foo.com
3. Delete the ingress.
4. Delete the hostrule
5. Create an hostrule with fqdn foo1.com and aliases as `foo.com`.
6. Check CRD is accepted or not.
*/
/*
func TestCreateUpdateDeleteHostRuleWithFqdnAlias_Dedicated(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	modelName := "admin/cluster--foo.com-L7-dedicated"
	hrname := "fqdn-aliases-hr-foo"
	secretName := objNameMap.GenerateName("my-secret")
	ingressName := objNameMap.GenerateName("foo-with-targets")
	svcName := objNameMap.GenerateName("avisvc")
	ingTestObj := IngressTestObject{
		ingressName: ingressName,
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
		nodes, ok := aviModel.(*avinodes.AviObjectGraph)
		if !ok {
			return 0
		}
		return len(nodes.GetAviVS())
	}, 20*time.Second).Should(gomega.Equal(1))
	sniVSKey := cache.NamespaceName{Namespace: "admin", Name: lib.Encode("cluster--foo.com", lib.EVHVS)}
	integrationtest.VerifyMetadataHostRule(t, g, sniVSKey, "default/fqdn-aliases-hr-foo", true)
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()

	// Common function that takes care of all validations
	validateNode := func(node *avinodes.AviVsNode, aliases []string) {
		g.Expect(node.VSVIPRefs).To(gomega.HaveLen(1))
		g.Expect(node.VSVIPRefs[0].FQDNs).Should(gomega.ContainElements(aliases))

		g.Expect(node.VHDomainNames).Should(gomega.ContainElements(aliases))
		g.Expect(node.HttpPolicyRefs).To(gomega.HaveLen(2))
		for _, httpPolicyRef := range nodes[0].HttpPolicyRefs {
			if httpPolicyRef.HppMap != nil {
				g.Expect(httpPolicyRef.HppMap).To(gomega.HaveLen(1))
				g.Expect(httpPolicyRef.HppMap[0].Host).Should(gomega.ContainElements(aliases))
			}
			if httpPolicyRef.RedirectPorts != nil {
				g.Expect(httpPolicyRef.RedirectPorts).To(gomega.HaveLen(1))
				g.Expect(httpPolicyRef.RedirectPorts[0].Hosts).Should(gomega.ContainElements(aliases))
			}
			g.Expect(httpPolicyRef.AviMarkers.Host).Should(gomega.ContainElements(aliases))
		}
	}

	// Check default values.
	validateNode(nodes[0], []string{"foo.com"})
	// now tear down hostrule and ingress
	integrationtest.TeardownHostRule(t, g, sniVSKey, hrname)
	TearDownIngressForCacheSyncCheckForAlias(t, secretName, ingressName, svcName)
	// sleep introduced so that cache will be synced
	time.Sleep(30 * time.Second)
	g.Eventually(func() int {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes, ok := aviModel.(*avinodes.AviObjectGraph)
		if !ok {
			return 0
		}
		return len(nodes.GetAviVS())
	}, 20*time.Second).Should(gomega.Equal(0))
	// Create new host rule with a valid FQDN Aliases
	hrNew := integrationtest.FakeHostRule{
		Name:      hrname,
		Namespace: "default",
		Fqdn:      "foo1.com",
	}.HostRule()
	aliases := []string{"foo.com", "alias2.com"}
	hrNew.Spec.VirtualHost.FqdnType = v1beta1.Exact
	hrNew.Spec.VirtualHost.Aliases = aliases
	hrNew.ResourceVersion = "1"
	_, err := V1beta1Client.AkoV1beta1().HostRules("default").Create(context.TODO(), hrNew, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in updating HostRule: %v", err)
	}
	g.Eventually(func() string {
		hostrule, _ := V1beta1Client.AkoV1beta1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 10*time.Second).Should(gomega.Equal("Accepted"))

	integrationtest.TeardownHostRule(t, g, sniVSKey, hrname)
}
*/
