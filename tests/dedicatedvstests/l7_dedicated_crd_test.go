/*
 * Copyright 2020-2021 VMware, Inc.
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
	"testing"
	"time"

	"github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	avinodes "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/apis/ako/v1alpha1"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/integrationtest"
)

func TestHostruleFQDNAliasesForDedicatedVS(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	modelName := "admin/cluster--foo.com-L7-dedicated"
	hrname := "fqdn-aliases-hr-foo"
	SetUpIngressForCacheSyncCheck(t, true, true, modelName)
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
	hrUpdate.Spec.VirtualHost.FqdnType = v1alpha1.Exact
	hrUpdate.Spec.VirtualHost.Aliases = aliases
	hrUpdate.ResourceVersion = "2"
	_, err := CRDClient.AkoV1alpha1().HostRules("default").Update(context.TODO(), hrUpdate, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating HostRule: %v", err)
	}
	g.Eventually(func() string {
		hostrule, _ := CRDClient.AkoV1alpha1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
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
	_, err = CRDClient.AkoV1alpha1().HostRules("default").Update(context.TODO(), hrUpdate, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating HostRule: %v", err)
	}
	g.Eventually(func() string {
		hostrule, _ := CRDClient.AkoV1alpha1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
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
	_, err = CRDClient.AkoV1alpha1().HostRules("default").Update(context.TODO(), hrUpdate, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating HostRule: %v", err)
	}
	g.Eventually(func() string {
		hostrule, _ := CRDClient.AkoV1alpha1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
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
	TearDownIngressForCacheSyncCheck(t, modelName)
}
