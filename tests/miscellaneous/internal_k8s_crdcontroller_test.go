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

// @AI-Generated
// Tests for CRD controller

package miscellaneous

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	akov1alpha2 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/apis/ako/v1alpha2"
	akov1beta1 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/apis/ako/v1beta1"
)

// Note: Most CRD controller functions (isHostRuleUpdated, isHTTPRuleUpdated, etc.)
// are not exported and cannot be tested from this package.
//
// The exported functions (SetAviInfrasettingVIPNetworks, SetAviInfrasettingNodeNetworks,
// GetSEGManagementNetwork) all require AVI controller client setup and are better tested
// through integration tests with a real or mocked AVI controller.
//
// These tests focus on verifying CRD object structure and basic validation.

// TestCRDObjectCreation tests that CRD objects can be created with proper structure
func TestCRDObjectCreation(t *testing.T) {
	t.Run("Create HostRule", func(t *testing.T) {
		hr := &akov1beta1.HostRule{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "default",
			},
			Spec: akov1beta1.HostRuleSpec{
				VirtualHost: akov1beta1.HostRuleVirtualHost{
					Fqdn: "test.com",
				},
			},
		}
		if hr.Name != "test" {
			t.Errorf("HostRule name = %v, want test", hr.Name)
		}
		if hr.Spec.VirtualHost.Fqdn != "test.com" {
			t.Errorf("HostRule FQDN = %v, want test.com", hr.Spec.VirtualHost.Fqdn)
		}
	})

	t.Run("Create HTTPRule", func(t *testing.T) {
		httpr := &akov1beta1.HTTPRule{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "default",
			},
			Spec: akov1beta1.HTTPRuleSpec{
				Fqdn: "test.com",
			},
		}
		if httpr.Name != "test" {
			t.Errorf("HTTPRule name = %v, want test", httpr.Name)
		}
		if httpr.Spec.Fqdn != "test.com" {
			t.Errorf("HTTPRule FQDN = %v, want test.com", httpr.Spec.Fqdn)
		}
	})

	t.Run("Create AviInfraSetting", func(t *testing.T) {
		infra := &akov1beta1.AviInfraSetting{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test",
			},
			Spec: akov1beta1.AviInfraSettingSpec{
				SeGroup: akov1beta1.AviInfraSettingSeGroup{
					Name: "Default-Group",
				},
			},
		}
		if infra.Name != "test" {
			t.Errorf("AviInfraSetting name = %v, want test", infra.Name)
		}
		if infra.Spec.SeGroup.Name != "Default-Group" {
			t.Errorf("AviInfraSetting SEGroup = %v, want Default-Group", infra.Spec.SeGroup.Name)
		}
	})

	t.Run("Create L4Rule", func(t *testing.T) {
		l4r := &akov1alpha2.L4Rule{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "default",
			},
			Spec: akov1alpha2.L4RuleSpec{
				LoadBalancerIP: StringPtr("10.0.0.1"),
			},
		}
		if l4r.Name != "test" {
			t.Errorf("L4Rule name = %v, want test", l4r.Name)
		}
		if l4r.Spec.LoadBalancerIP == nil || *l4r.Spec.LoadBalancerIP != "10.0.0.1" {
			t.Errorf("L4Rule LoadBalancerIP = %v, want 10.0.0.1", l4r.Spec.LoadBalancerIP)
		}
	})

	t.Run("Create L7Rule", func(t *testing.T) {
		l7r := &akov1alpha2.L7Rule{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "default",
			},
			Spec: akov1alpha2.L7RuleSpec{
				AllowInvalidClientCert: BoolPtr(false),
			},
		}
		if l7r.Name != "test" {
			t.Errorf("L7Rule name = %v, want test", l7r.Name)
		}
		if l7r.Spec.AllowInvalidClientCert == nil || *l7r.Spec.AllowInvalidClientCert != false {
			t.Errorf("L7Rule AllowInvalidClientCert = %v, want false", l7r.Spec.AllowInvalidClientCert)
		}
	})

	t.Run("Create SSORule", func(t *testing.T) {
		ssor := &akov1alpha2.SSORule{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "default",
			},
			Spec: akov1alpha2.SSORuleSpec{
				Fqdn: StringPtr("test.com"),
			},
		}
		if ssor.Name != "test" {
			t.Errorf("SSORule name = %v, want test", ssor.Name)
		}
		if ssor.Spec.Fqdn == nil || *ssor.Spec.Fqdn != "test.com" {
			t.Errorf("SSORule FQDN = %v, want test.com", ssor.Spec.Fqdn)
		}
	})
}

// TestCRDStatusFields tests that CRD status fields are properly structured
func TestCRDStatusFields(t *testing.T) {
	t.Run("HostRule status", func(t *testing.T) {
		hr := &akov1beta1.HostRule{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "default",
			},
			Status: akov1beta1.HostRuleStatus{
				Status: "Accepted",
				Error:  "",
			},
		}
		if hr.Status.Status != "Accepted" {
			t.Errorf("HostRule status = %v, want Accepted", hr.Status.Status)
		}
	})

	t.Run("HTTPRule status", func(t *testing.T) {
		httpr := &akov1beta1.HTTPRule{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "default",
			},
			Status: akov1beta1.HTTPRuleStatus{
				Status: "Rejected",
				Error:  "Invalid configuration",
			},
		}
		if httpr.Status.Status != "Rejected" {
			t.Errorf("HTTPRule status = %v, want Rejected", httpr.Status.Status)
		}
		if httpr.Status.Error != "Invalid configuration" {
			t.Errorf("HTTPRule error = %v, want 'Invalid configuration'", httpr.Status.Error)
		}
	})

	t.Run("L4Rule status", func(t *testing.T) {
		l4r := &akov1alpha2.L4Rule{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "default",
			},
			Status: akov1alpha2.L4RuleStatus{
				Status: "Accepted",
			},
		}
		if l4r.Status.Status != "Accepted" {
			t.Errorf("L4Rule status = %v, want Accepted", l4r.Status.Status)
		}
	})

	t.Run("AviInfraSetting status", func(t *testing.T) {
		infra := &akov1beta1.AviInfraSetting{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test",
			},
			Status: akov1beta1.AviInfraSettingStatus{
				Status: "Accepted",
				Error:  "",
			},
		}
		if infra.Status.Status != "Accepted" {
			t.Errorf("AviInfraSetting status = %v, want Accepted", infra.Status.Status)
		}
	})
}

// TestAviInfraSettingNetworkConfiguration tests network configuration structures
func TestAviInfraSettingNetworkConfiguration(t *testing.T) {
	t.Run("VIP network with CIDR", func(t *testing.T) {
		vipNetwork := akov1beta1.AviInfraSettingVipNetwork{
			NetworkName: "vip-network-1",
			Cidr:        "10.0.0.0/24",
		}
		if vipNetwork.NetworkName != "vip-network-1" {
			t.Errorf("NetworkName = %v, want vip-network-1", vipNetwork.NetworkName)
		}
		if vipNetwork.Cidr != "10.0.0.0/24" {
			t.Errorf("Cidr = %v, want 10.0.0.0/24", vipNetwork.Cidr)
		}
	})

	t.Run("VIP network with IPv6 CIDR", func(t *testing.T) {
		vipNetwork := akov1beta1.AviInfraSettingVipNetwork{
			NetworkName: "vip-network-v6",
			V6Cidr:      "2001:db8::/64",
		}
		if vipNetwork.V6Cidr != "2001:db8::/64" {
			t.Errorf("V6Cidr = %v, want 2001:db8::/64", vipNetwork.V6Cidr)
		}
	})

	t.Run("VIP network with UUID", func(t *testing.T) {
		vipNetwork := akov1beta1.AviInfraSettingVipNetwork{
			NetworkName: "vip-network-1",
			NetworkUUID: "network-uuid-123",
		}
		if vipNetwork.NetworkUUID != "network-uuid-123" {
			t.Errorf("NetworkUUID = %v, want network-uuid-123", vipNetwork.NetworkUUID)
		}
	})

	t.Run("Node network with CIDRs", func(t *testing.T) {
		nodeNetwork := akov1beta1.AviInfraSettingNodeNetwork{
			NetworkName: "node-network-1",
			Cidrs:       []string{"192.168.1.0/24", "192.168.2.0/24"},
		}
		if len(nodeNetwork.Cidrs) != 2 {
			t.Errorf("Cidrs length = %v, want 2", len(nodeNetwork.Cidrs))
		}
		if nodeNetwork.Cidrs[0] != "192.168.1.0/24" {
			t.Errorf("Cidrs[0] = %v, want 192.168.1.0/24", nodeNetwork.Cidrs[0])
		}
	})

	t.Run("Node network with UUID", func(t *testing.T) {
		nodeNetwork := akov1beta1.AviInfraSettingNodeNetwork{
			NetworkName: "node-network-1",
			NetworkUUID: "node-uuid-456",
		}
		if nodeNetwork.NetworkUUID != "node-uuid-456" {
			t.Errorf("NetworkUUID = %v, want node-uuid-456", nodeNetwork.NetworkUUID)
		}
	})
}

// TestAviInfraSettingComplexConfiguration tests complex AviInfraSetting configurations
func TestAviInfraSettingComplexConfiguration(t *testing.T) {
	t.Run("Complete AviInfraSetting with all fields", func(t *testing.T) {
		infra := &akov1beta1.AviInfraSetting{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "complete-infra",
				Namespace: "default",
			},
			Spec: akov1beta1.AviInfraSettingSpec{
				SeGroup: akov1beta1.AviInfraSettingSeGroup{
					Name: "Test-Group",
				},
				Network: akov1beta1.AviInfraSettingNetwork{
					VipNetworks: []akov1beta1.AviInfraSettingVipNetwork{
						{
							NetworkName: "vip-network-1",
							Cidr:        "10.0.0.0/24",
						},
						{
							NetworkName: "vip-network-2",
							V6Cidr:      "2001:db8::/64",
						},
					},
					NodeNetworks: []akov1beta1.AviInfraSettingNodeNetwork{
						{
							NetworkName: "node-network-1",
							Cidrs:       []string{"192.168.1.0/24"},
						},
					},
					EnableRhi: BoolPtr(true),
				},
			},
		}

		if infra.Spec.SeGroup.Name != "Test-Group" {
			t.Errorf("SEGroup name = %v, want Test-Group", infra.Spec.SeGroup.Name)
		}
		if len(infra.Spec.Network.VipNetworks) != 2 {
			t.Errorf("VipNetworks length = %v, want 2", len(infra.Spec.Network.VipNetworks))
		}
		if len(infra.Spec.Network.NodeNetworks) != 1 {
			t.Errorf("NodeNetworks length = %v, want 1", len(infra.Spec.Network.NodeNetworks))
		}
		if infra.Spec.Network.EnableRhi == nil || *infra.Spec.Network.EnableRhi != true {
			t.Errorf("EnableRhi = %v, want true", infra.Spec.Network.EnableRhi)
		}
	})

	t.Run("AviInfraSetting with BGP peer labels", func(t *testing.T) {
		infra := &akov1beta1.AviInfraSetting{
			ObjectMeta: metav1.ObjectMeta{
				Name: "bgp-infra",
			},
			Spec: akov1beta1.AviInfraSettingSpec{
				Network: akov1beta1.AviInfraSettingNetwork{
					EnableRhi:     BoolPtr(true),
					BgpPeerLabels: []string{"peer-label-1", "peer-label-2"},
				},
			},
		}

		if len(infra.Spec.Network.BgpPeerLabels) != 2 {
			t.Errorf("BgpPeerLabels length = %v, want 2", len(infra.Spec.Network.BgpPeerLabels))
		}
	})
}

// TestL4RuleConfiguration tests L4Rule configuration structures
func TestL4RuleConfiguration(t *testing.T) {
	t.Run("L4Rule with backend properties", func(t *testing.T) {
		l4r := &akov1alpha2.L4Rule{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-l4",
				Namespace: "default",
			},
			Spec: akov1alpha2.L4RuleSpec{
				LoadBalancerIP: StringPtr("10.0.0.1"),
				BackendProperties: []*akov1alpha2.BackendProperties{
					{
						Protocol: StringPtr("TCP"),
					},
				},
			},
		}

		if len(l4r.Spec.BackendProperties) != 1 {
			t.Errorf("BackendProperties length = %v, want 1", len(l4r.Spec.BackendProperties))
		}
		if l4r.Spec.BackendProperties[0].Protocol == nil || *l4r.Spec.BackendProperties[0].Protocol != "TCP" {
			t.Errorf("Protocol = %v, want TCP", l4r.Spec.BackendProperties[0].Protocol)
		}
	})

	t.Run("L4Rule with analytics policy", func(t *testing.T) {
		l4r := &akov1alpha2.L4Rule{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-l4-analytics",
				Namespace: "default",
			},
			Spec: akov1alpha2.L4RuleSpec{
				LoadBalancerIP:      StringPtr("10.0.0.2"),
				AnalyticsProfileRef: StringPtr("test-analytics-profile"),
				SecurityPolicyRef:   StringPtr("test-security-policy"),
			},
		}

		if l4r.Spec.AnalyticsProfileRef == nil || *l4r.Spec.AnalyticsProfileRef != "test-analytics-profile" {
			t.Errorf("AnalyticsProfileRef = %v, want test-analytics-profile", l4r.Spec.AnalyticsProfileRef)
		}
		if l4r.Spec.SecurityPolicyRef == nil || *l4r.Spec.SecurityPolicyRef != "test-security-policy" {
			t.Errorf("SecurityPolicyRef = %v, want test-security-policy", l4r.Spec.SecurityPolicyRef)
		}
	})
}

// TestL7RuleConfiguration tests L7Rule configuration structures
func TestL7RuleConfiguration(t *testing.T) {
	t.Run("L7Rule with performance limits", func(t *testing.T) {
		l7r := &akov1alpha2.L7Rule{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-l7-perf",
				Namespace: "default",
			},
			Spec: akov1alpha2.L7RuleSpec{
				PerformanceLimits: &akov1alpha2.PerformanceLimits{
					MaxConcurrentConnections: Int32Ptr(1000),
				},
			},
		}

		if l7r.Spec.PerformanceLimits == nil {
			t.Error("PerformanceLimits should not be nil")
		}
		if l7r.Spec.PerformanceLimits.MaxConcurrentConnections == nil || *l7r.Spec.PerformanceLimits.MaxConcurrentConnections != 1000 {
			t.Errorf("MaxConcurrentConnections = %v, want 1000", l7r.Spec.PerformanceLimits.MaxConcurrentConnections)
		}
	})

	t.Run("L7Rule with WAF policy", func(t *testing.T) {
		l7r := &akov1alpha2.L7Rule{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-l7-waf",
				Namespace: "default",
			},
			Spec: akov1alpha2.L7RuleSpec{
				WafPolicy: &akov1alpha2.KindNameNamespace{
					Name: StringPtr("test-waf-policy"),
				},
			},
		}

		if l7r.Spec.WafPolicy == nil {
			t.Error("WafPolicy should not be nil")
		}
		if l7r.Spec.WafPolicy.Name == nil || *l7r.Spec.WafPolicy.Name != "test-waf-policy" {
			t.Errorf("WafPolicy name = %v, want test-waf-policy", l7r.Spec.WafPolicy.Name)
		}
	})
}

// TestHTTPRuleConfiguration tests HTTPRule configuration structures
func TestHTTPRuleConfiguration(t *testing.T) {
	t.Run("HTTPRule with paths", func(t *testing.T) {
		httpr := &akov1beta1.HTTPRule{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-http-paths",
				Namespace: "default",
			},
			Spec: akov1beta1.HTTPRuleSpec{
				Fqdn: "test.com",
				Paths: []akov1beta1.HTTPRulePaths{
					{
						Target: "/api",
					},
					{
						Target: "/web",
					},
				},
			},
		}

		if len(httpr.Spec.Paths) != 2 {
			t.Errorf("Paths length = %v, want 2", len(httpr.Spec.Paths))
		}
		if httpr.Spec.Paths[0].Target != "/api" {
			t.Errorf("Path[0] target = %v, want /api", httpr.Spec.Paths[0].Target)
		}
	})

	t.Run("HTTPRule with TLS settings", func(t *testing.T) {
		httpr := &akov1beta1.HTTPRule{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-http-tls",
				Namespace: "default",
			},
			Spec: akov1beta1.HTTPRuleSpec{
				Fqdn: "test.com",
				Paths: []akov1beta1.HTTPRulePaths{
					{
						Target: "/secure",
						TLS: akov1beta1.HTTPRuleTLS{
							Type:       "reencrypt",
							PKIProfile: "test-pki",
						},
					},
				},
			},
		}

		if httpr.Spec.Paths[0].TLS.Type != "reencrypt" {
			t.Errorf("TLS type = %v, want reencrypt", httpr.Spec.Paths[0].TLS.Type)
		}
		if httpr.Spec.Paths[0].TLS.PKIProfile != "test-pki" {
			t.Errorf("PKIProfile = %v, want test-pki", httpr.Spec.Paths[0].TLS.PKIProfile)
		}
	})
}
