/*
 * Copyright 2025 VMware, Inc.
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

package multitenancytests

import (
	"context"
	"strings"
	"testing"
	"time"

	avinodes "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/ingresstests"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/integrationtest"

	"github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestTenantNameSanitizationWithSpecialCharacters(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	// Test case: Create a tenant with special characters that should be sanitized
	problematicTenant := "test@1234$"
	expectedSanitizedTenant := "test-1234"
	ns := "default"
	serviceName := objNameMap.GenerateName("test-service")
	ingressName := objNameMap.GenerateName("test-ingress")

	// Setup test for ingress, which creates service and endpoints
	ingresstests.SetUpTestForIngress(t, serviceName)
	defer ingresstests.TearDownTestForIngress(t, serviceName)

	// Annotate the default namespace with the problematic tenant name
	integrationtest.AnnotateNamespaceWithTenant(t, ns, problematicTenant)
	defer integrationtest.RemoveAnnotateAKONamespaceWithInfraSetting(t, ns)

	// Create an ingress
	ingressCreate := (integrationtest.FakeIngress{
		Name:        ingressName,
		Namespace:   ns,
		DnsNames:    []string{"test-host.avi.internal"},
		ServiceName: serviceName,
	}).Ingress()
	_, err := integrationtest.KubeClient.NetworkingV1().Ingresses(ns).Create(context.TODO(), ingressCreate, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	defer integrationtest.KubeClient.NetworkingV1().Ingresses(ns).Delete(context.TODO(), ingressName, metav1.DeleteOptions{})

	// Wait for the ingress to be processed and verify sanitization
	g.Eventually(func() bool {
		// The model name format is: tenant/cluster--Shared-L7-1
		modelName := problematicTenant + "/cluster--Shared-L7-1"
		if found, aviModel := objects.SharedAviGraphLister().Get(modelName); found {
			if nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS(); len(nodes) > 0 {
				if len(nodes[0].VSVIPRefs) > 0 && len(nodes[0].VSVIPRefs[0].FQDNs) > 0 {
					for _, fqdn := range nodes[0].VSVIPRefs[0].FQDNs {
						if fqdn != "" {
							if strings.Contains(fqdn, expectedSanitizedTenant) && !strings.Contains(fqdn, "@") && !strings.Contains(fqdn, "$") {
								return true
							}
						}
					}
				}
			}
		}
		return false
	}, 40*time.Second).Should(gomega.Equal(true), "Failed to verify FQDN sanitization for special characters")
}

func TestTenantNameSanitizationWithUnderscores(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	// Test case: Create a tenant with underscores that should be sanitized for SAN compliance
	problematicTenant := "my_tenant_name"
	expectedSanitizedTenant := "my-tenant-name"
	ns := "default"
	serviceName := objNameMap.GenerateName("test-service-underscore")
	ingressName := objNameMap.GenerateName("test-ingress-underscore")

	// Setup test for ingress, which creates service and endpoints
	ingresstests.SetUpTestForIngress(t, serviceName)
	defer ingresstests.TearDownTestForIngress(t, serviceName)

	// Annotate the default namespace with the problematic tenant name
	integrationtest.AnnotateNamespaceWithTenant(t, ns, problematicTenant)
	defer integrationtest.RemoveAnnotateAKONamespaceWithInfraSetting(t, ns)

	// Create an ingress
	ingressCreate := (integrationtest.FakeIngress{
		Name:        ingressName,
		Namespace:   ns,
		DnsNames:    []string{"test-host-underscore.avi.internal"},
		ServiceName: serviceName,
	}).Ingress()
	_, err := integrationtest.KubeClient.NetworkingV1().Ingresses(ns).Create(context.TODO(), ingressCreate, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error in adding Ingress: %v", err)
	}
	defer integrationtest.KubeClient.NetworkingV1().Ingresses(ns).Delete(context.TODO(), ingressName, metav1.DeleteOptions{})

	// Wait for the ingress to be processed and verify sanitization
	g.Eventually(func() bool {
		// The model name format is: tenant/cluster--Shared-L7-0
		modelName := problematicTenant + "/cluster--Shared-L7-0"
		if found, aviModel := objects.SharedAviGraphLister().Get(modelName); found {
			if nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS(); len(nodes) > 0 {
				if len(nodes[0].VSVIPRefs) > 0 && len(nodes[0].VSVIPRefs[0].FQDNs) > 0 {
					for _, fqdn := range nodes[0].VSVIPRefs[0].FQDNs {
						if fqdn != "" {
							if strings.Contains(fqdn, expectedSanitizedTenant) && !strings.Contains(fqdn, "_") {
								return true
							}
						}
					}
				}
			}
		}
		return false
	}, 40*time.Second).Should(gomega.Equal(true), "Failed to verify FQDN sanitization for underscores")
}
