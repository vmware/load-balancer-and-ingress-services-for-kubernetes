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

package miscellaneous

import (
	"errors"
	"regexp"
	"strings"
	"testing"

	avimodels "github.com/vmware/alb-sdk/go/models"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
)

// ========== Tests from finalizer_utils_test.go ==========

func TestContainsFinalizer(t *testing.T) {
	tests := []struct {
		name       string
		finalizers []string
		finalizer  string
		expected   bool
	}{
		{
			name:       "Finalizer exists",
			finalizers: []string{"finalizer1", "finalizer2", "finalizer3"},
			finalizer:  "finalizer2",
			expected:   true,
		},
		{
			name:       "Finalizer does not exist",
			finalizers: []string{"finalizer1", "finalizer2"},
			finalizer:  "finalizer3",
			expected:   false,
		},
		{
			name:       "Empty finalizers",
			finalizers: []string{},
			finalizer:  "finalizer1",
			expected:   false,
		},
		{
			name:       "Nil finalizers",
			finalizers: nil,
			finalizer:  "finalizer1",
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &networkingv1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Finalizers: tt.finalizers,
				},
			}

			result := lib.ContainsFinalizer(obj, tt.finalizer)
			if result != tt.expected {
				t.Errorf("ContainsFinalizer() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestContainsFinalizerGatewayFinalizer(t *testing.T) {
	tests := []struct {
		name       string
		finalizers []string
		expected   bool
	}{
		{
			name:       "Contains gateway finalizer",
			finalizers: []string{lib.GatewayFinalizer, "other-finalizer"},
			expected:   true,
		},
		{
			name:       "Does not contain gateway finalizer",
			finalizers: []string{"other-finalizer"},
			expected:   false,
		},
		{
			name:       "Empty finalizers",
			finalizers: []string{},
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &networkingv1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Finalizers: tt.finalizers,
				},
			}

			result := lib.ContainsFinalizer(obj, lib.GatewayFinalizer)
			if result != tt.expected {
				t.Errorf("ContainsFinalizer(GatewayFinalizer) = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestContainsFinalizerIngressFinalizer(t *testing.T) {
	tests := []struct {
		name       string
		finalizers []string
		expected   bool
	}{
		{
			name:       "Contains ingress finalizer",
			finalizers: []string{lib.IngressFinalizer, "other-finalizer"},
			expected:   true,
		},
		{
			name:       "Does not contain ingress finalizer",
			finalizers: []string{"other-finalizer"},
			expected:   false,
		},
		{
			name:       "Empty finalizers",
			finalizers: []string{},
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &networkingv1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Finalizers: tt.finalizers,
				},
			}

			result := lib.ContainsFinalizer(obj, lib.IngressFinalizer)
			if result != tt.expected {
				t.Errorf("ContainsFinalizer(IngressFinalizer) = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestFinalizerConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant string
	}{
		{"GatewayFinalizer", lib.GatewayFinalizer},
		{"IngressFinalizer", lib.IngressFinalizer},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant == "" {
				t.Errorf("%s should not be empty", tt.name)
			}
		})
	}
}

func TestFinalizerWithMultipleObjects(t *testing.T) {
	finalizer := "test.finalizer.com"

	tests := []struct {
		name       string
		finalizers []string
		expected   bool
	}{
		{
			name:       "First position",
			finalizers: []string{finalizer, "other1", "other2"},
			expected:   true,
		},
		{
			name:       "Middle position",
			finalizers: []string{"other1", finalizer, "other2"},
			expected:   true,
		},
		{
			name:       "Last position",
			finalizers: []string{"other1", "other2", finalizer},
			expected:   true,
		},
		{
			name:       "Not present",
			finalizers: []string{"other1", "other2", "other3"},
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &networkingv1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Finalizers: tt.finalizers,
				},
			}

			result := lib.ContainsFinalizer(obj, finalizer)
			if result != tt.expected {
				t.Errorf("ContainsFinalizer() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// ========== Tests from parse_objects_test.go (exported functions only) ==========

func TestIPAddrIntfToObj(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected *avimodels.IPAddr
	}{
		{
			name: "Valid IPv4 address",
			input: map[string]interface{}{
				"addr": "192.168.1.1",
				"type": "V4",
			},
			expected: &avimodels.IPAddr{
				Addr: StringPtr("192.168.1.1"),
				Type: StringPtr("V4"),
			},
		},
		{
			name: "Valid IPv6 address",
			input: map[string]interface{}{
				"addr": "2001:db8::1",
				"type": "V6",
			},
			expected: &avimodels.IPAddr{
				Addr: StringPtr("2001:db8::1"),
				Type: StringPtr("V6"),
			},
		},
		{
			name: "Empty address",
			input: map[string]interface{}{
				"addr": "",
				"type": "V4",
			},
			expected: &avimodels.IPAddr{
				Addr: StringPtr(""),
				Type: StringPtr("V4"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := lib.IPAddrIntfToObj(tt.input)

			if result == nil {
				t.Fatal("IPAddrIntfToObj returned nil")
			}

			if *result.Addr != *tt.expected.Addr {
				t.Errorf("Addr = %v, want %v", *result.Addr, *tt.expected.Addr)
			}

			if *result.Type != *tt.expected.Type {
				t.Errorf("Type = %v, want %v", *result.Type, *tt.expected.Type)
			}
		})
	}
}

func TestIAddrPrefixIntfToObj(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected *avimodels.IPAddrPrefix
	}{
		{
			name: "Valid CIDR",
			input: map[string]interface{}{
				"ip_addr": map[string]interface{}{
					"addr": "192.168.1.0",
					"type": "V4",
				},
				"mask": float64(24),
			},
			expected: &avimodels.IPAddrPrefix{
				IPAddr: &avimodels.IPAddr{
					Addr: StringPtr("192.168.1.0"),
					Type: StringPtr("V4"),
				},
				Mask: Int32Ptr(24),
			},
		},
		{
			name: "IPv6 CIDR",
			input: map[string]interface{}{
				"ip_addr": map[string]interface{}{
					"addr": "2001:db8::",
					"type": "V6",
				},
				"mask": float64(64),
			},
			expected: &avimodels.IPAddrPrefix{
				IPAddr: &avimodels.IPAddr{
					Addr: StringPtr("2001:db8::"),
					Type: StringPtr("V6"),
				},
				Mask: Int32Ptr(64),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := lib.IAddrPrefixIntfToObj(tt.input)

			if result == nil {
				t.Fatal("IAddrPrefixIntfToObj returned nil")
			}

			if *result.IPAddr.Addr != *tt.expected.IPAddr.Addr {
				t.Errorf("IPAddr.Addr = %v, want %v", *result.IPAddr.Addr, *tt.expected.IPAddr.Addr)
			}

			if *result.Mask != *tt.expected.Mask {
				t.Errorf("Mask = %v, want %v", *result.Mask, *tt.expected.Mask)
			}
		})
	}
}

func TestLabelsIntfToObj(t *testing.T) {
	tests := []struct {
		name     string
		input    []interface{}
		expected int
	}{
		{
			name: "Single label",
			input: []interface{}{
				map[string]interface{}{
					"key":   "env",
					"value": "prod",
				},
			},
			expected: 1,
		},
		{
			name: "Multiple labels",
			input: []interface{}{
				map[string]interface{}{
					"key":   "env",
					"value": "prod",
				},
				map[string]interface{}{
					"key":   "app",
					"value": "web",
				},
			},
			expected: 2,
		},
		{
			name:     "Empty labels",
			input:    []interface{}{},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := lib.LabelsIntfToObj(tt.input)

			if len(result) != tt.expected {
				t.Errorf("Length = %v, want %v", len(result), tt.expected)
			}

			for i, label := range result {
				if i < len(tt.input) {
					inputMap := tt.input[i].(map[string]interface{})
					if *label.Key != inputMap["key"].(string) {
						t.Errorf("Label[%d].Key = %v, want %v", i, *label.Key, inputMap["key"])
					}
					if *label.Value != inputMap["value"].(string) {
						t.Errorf("Label[%d].Value = %v, want %v", i, *label.Value, inputMap["value"])
					}
				}
			}
		})
	}
}

func TestStaticRoutesIntfToObj(t *testing.T) {
	tests := []struct {
		name     string
		input    []interface{}
		expected int
	}{
		{
			name: "Single static route",
			input: []interface{}{
				map[string]interface{}{
					"route_id": "route-1",
					"prefix": map[string]interface{}{
						"ip_addr": map[string]interface{}{
							"addr": "10.0.0.0",
							"type": "V4",
						},
						"mask": float64(24),
					},
					"next_hop": map[string]interface{}{
						"addr": "192.168.1.1",
						"type": "V4",
					},
					"disable_gateway_monitor": false,
				},
			},
			expected: 1,
		},
		{
			name: "Multiple static routes",
			input: []interface{}{
				map[string]interface{}{
					"route_id": "route-1",
					"prefix": map[string]interface{}{
						"ip_addr": map[string]interface{}{
							"addr": "10.0.0.0",
							"type": "V4",
						},
						"mask": float64(24),
					},
					"next_hop": map[string]interface{}{
						"addr": "192.168.1.1",
						"type": "V4",
					},
				},
				map[string]interface{}{
					"route_id": "route-2",
					"prefix": map[string]interface{}{
						"ip_addr": map[string]interface{}{
							"addr": "172.16.0.0",
							"type": "V4",
						},
						"mask": float64(16),
					},
					"next_hop": map[string]interface{}{
						"addr": "192.168.1.2",
						"type": "V4",
					},
				},
			},
			expected: 2,
		},
		{
			name:     "Empty routes",
			input:    []interface{}{},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := lib.StaticRoutesIntfToObj(tt.input)

			if len(result) != tt.expected {
				t.Errorf("Length = %v, want %v", len(result), tt.expected)
			}

			for i, route := range result {
				if i < len(tt.input) {
					inputMap := tt.input[i].(map[string]interface{})
					if *route.RouteID != inputMap["route_id"].(string) {
						t.Errorf("Route[%d].RouteID = %v, want %v", i, *route.RouteID, inputMap["route_id"])
					}
				}
			}
		})
	}
}

func TestStaticRoutesIntfToObjWithLabels(t *testing.T) {
	input := []interface{}{
		map[string]interface{}{
			"route_id": "route-with-labels",
			"prefix": map[string]interface{}{
				"ip_addr": map[string]interface{}{
					"addr": "10.0.0.0",
					"type": "V4",
				},
				"mask": float64(24),
			},
			"next_hop": map[string]interface{}{
				"addr": "192.168.1.1",
				"type": "V4",
			},
			"labels": []interface{}{
				map[string]interface{}{
					"key":   "cluster",
					"value": "test-cluster",
				},
			},
		},
	}

	result := lib.StaticRoutesIntfToObj(input)

	if len(result) != 1 {
		t.Fatalf("Expected 1 route, got %d", len(result))
	}

	route := result[0]
	if len(route.Labels) != 1 {
		t.Errorf("Expected 1 label, got %d", len(route.Labels))
	}

	if route.Labels[0] != nil {
		if *route.Labels[0].Key != "cluster" {
			t.Errorf("Label key = %v, want cluster", *route.Labels[0].Key)
		}
		if *route.Labels[0].Value != "test-cluster" {
			t.Errorf("Label value = %v, want test-cluster", *route.Labels[0].Value)
		}
	}
}

// ========== Tests from constants_test.go ==========

func TestConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant interface{}
		expected interface{}
	}{
		{"DISABLE_STATIC_ROUTE_SYNC", lib.DISABLE_STATIC_ROUTE_SYNC, "DISABLE_STATIC_ROUTE_SYNC"},
		{"ENABLE_RHI", lib.ENABLE_RHI, "ENABLE_RHI"},
		{"ENABLE_EVH", lib.ENABLE_EVH, "ENABLE_EVH"},
		{"CNI_PLUGIN", lib.CNI_PLUGIN, "CNI_PLUGIN"},
		{"CALICO_CNI", lib.CALICO_CNI, "calico"},
		{"ANTREA_CNI", lib.ANTREA_CNI, "antrea"},
		{"NCP_CNI", lib.NCP_CNI, "ncp"},
		{"OPENSHIFT_CNI", lib.OPENSHIFT_CNI, "openshift"},
		{"OVN_KUBERNETES_CNI", lib.OVN_KUBERNETES_CNI, "ovn-kubernetes"},
		{"CILIUM_CNI", lib.CILIUM_CNI, "cilium"},
		{"INGRESS_API", lib.INGRESS_API, "INGRESS_API"},
		{"AviConfigMap", lib.AviConfigMap, "avi-k8s-config"},
		{"AviSecret", lib.AviSecret, "avi-secret"},
		{"AviInitSecret", lib.AviInitSecret, "avi-init-secret"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("%s = %v, want %v", tt.name, tt.constant, tt.expected)
			}
		})
	}
}

func TestCloudTypeConstants(t *testing.T) {
	cloudTypes := map[string]string{
		"CLOUD_VCENTER":   lib.CLOUD_VCENTER,
		"CLOUD_AZURE":     lib.CLOUD_AZURE,
		"CLOUD_AWS":       lib.CLOUD_AWS,
		"CLOUD_OPENSTACK": lib.CLOUD_OPENSTACK,
		"CLOUD_GCP":       lib.CLOUD_GCP,
		"CLOUD_NONE":      lib.CLOUD_NONE,
		"CLOUD_NSXT":      lib.CLOUD_NSXT,
	}

	for name, value := range cloudTypes {
		if value == "" {
			t.Errorf("%s should not be empty", name)
		}
	}
}

func TestShardSchemeConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{"DEFAULT_SHARD_SCHEME", lib.DEFAULT_SHARD_SCHEME, "hostname"},
		{"HOSTNAME_SHARD_SCHEME", lib.HOSTNAME_SHARD_SCHEME, "hostname"},
		{"NAMESPACE_SHARD_SCHEME", lib.NAMESPACE_SHARD_SCHEME, "namespace"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("%s = %v, want %v", tt.name, tt.constant, tt.expected)
			}
		})
	}
}

func TestRegexConstants(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		valid   []string
		invalid []string
	}{
		{
			name:    "IPCIDRRegex",
			pattern: lib.IPCIDRRegex,
			valid:   []string{"192.168.1.0/24", "10.0.0.0/8", "172.16.0.0/16"},
			invalid: []string{"256.1.1.1/24", "192.168.1.1/33", "invalid"},
		},
		{
			name:    "IPRegex",
			pattern: lib.IPRegex,
			valid:   []string{"192.168.1.1", "10.0.0.1", "172.16.0.1"},
			invalid: []string{"256.1.1.1", "192.168.1", "invalid"},
		},
		{
			name:    "IPV6CIDRRegex",
			pattern: lib.IPV6CIDRRegex,
			valid:   []string{"2001:db8::/32", "fe80::/10", "::1/128"},
			invalid: []string{"invalid", "192.168.1.1/24"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re, err := regexp.Compile(tt.pattern)
			if err != nil {
				t.Fatalf("Failed to compile regex %s: %v", tt.name, err)
			}

			for _, valid := range tt.valid {
				if !re.MatchString(valid) {
					t.Errorf("%s should match %s", tt.name, valid)
				}
			}

			for _, invalid := range tt.invalid {
				if re.MatchString(invalid) {
					t.Errorf("%s should not match %s", tt.name, invalid)
				}
			}
		})
	}
}

func TestAutoFQDNConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{"AutoFQDNDefault", lib.AutoFQDNDefault, "Default"},
		{"AutoFQDNFlat", lib.AutoFQDNFlat, "Flat"},
		{"AutoFQDNDisabled", lib.AutoFQDNDisabled, "Disabled"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("%s = %v, want %v", tt.name, tt.constant, tt.expected)
			}
		})
	}
}

func TestStatusConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{"StatusRejected", lib.StatusRejected, "Rejected"},
		{"StatusAccepted", lib.StatusAccepted, "Accepted"},
		{"CRDActive", lib.CRDActive, "ACTIVE"},
		{"CRDInactive", lib.CRDInactive, "INACTIVE"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("%s = %v, want %v", tt.name, tt.constant, tt.expected)
			}
		})
	}
}

func TestEventConstants(t *testing.T) {
	eventConstants := []string{
		lib.AKOEventComponent,
		lib.AKOShutdown,
		lib.SyncDisabled,
		lib.ValidatedUserInput,
		lib.StatusSync,
		lib.AKOReady,
		lib.AKOPause,
		lib.DuplicateHostPath,
		lib.DuplicateHost,
		lib.Removed,
		lib.Synced,
		lib.Attached,
		lib.Detached,
		lib.PatchFailed,
		lib.InvalidConfiguration,
	}

	for _, constant := range eventConstants {
		if constant == "" {
			t.Errorf("Event constant should not be empty")
		}
	}
}

func TestAnnotationConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant string
	}{
		{"DefaultIngressClassAnnotation", lib.DefaultIngressClassAnnotation},
		{"ExternalDNSAnnotation", lib.ExternalDNSAnnotation},
		{"GatewayFinalizer", lib.GatewayFinalizer},
		{"IngressFinalizer", lib.IngressFinalizer},
		{"AkoGroup", lib.AkoGroup},
		{"AviIngressController", lib.AviIngressController},
		{"InfraSettingNameAnnotation", lib.InfraSettingNameAnnotation},
		{"SkipNodePortAnnotation", lib.SkipNodePortAnnotation},
		{"PassthroughAnnotation", lib.PassthroughAnnotation},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant == "" {
				t.Errorf("%s should not be empty", tt.name)
			}
		})
	}
}

func TestNumericConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant int
		min      int
	}{
		{"SLOW_SYNC_TIME", lib.SLOW_SYNC_TIME, 0},
		{"SSLPort", lib.SSLPort, 0},
		{"DNS_LABEL_LENGTH", lib.DNS_LABEL_LENGTH, 0},
		{"NODE_NETWORK_MAX_ENTRIES", lib.NODE_NETWORK_MAX_ENTRIES, 0},
		{"AVI_OBJ_NAME_MAX_LENGTH", lib.AVI_OBJ_NAME_MAX_LENGTH, 0},
		{"ControllerReqWaitTime", lib.ControllerReqWaitTime, 0},
		{"AviObjDeletionTime", lib.AviObjDeletionTime, 0},
		{"FullSyncInterval", lib.FullSyncInterval, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant < tt.min {
				t.Errorf("%s = %d, should be >= %d", tt.name, tt.constant, tt.min)
			}
		})
	}
}

func TestPassthroughDatascript(t *testing.T) {
	if lib.PassthroughDatascript == "" {
		t.Error("PassthroughDatascript should not be empty")
	}

	// Check for key elements in the datascript
	requiredElements := []string{
		"avi_tls",
		"SNI",
		"poolgroup.select",
		"CLUSTER--AVIINFRA",
	}

	for _, element := range requiredElements {
		if !regexp.MustCompile(element).MatchString(lib.PassthroughDatascript) {
			t.Errorf("PassthroughDatascript should contain %s", element)
		}
	}
}

func TestCacheIndexerConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant string
	}{
		{"AviSettingGWClassIndex", lib.AviSettingGWClassIndex},
		{"GatewayClassGatewayIndex", lib.GatewayClassGatewayIndex},
		{"SeGroupAviSettingIndex", lib.SeGroupAviSettingIndex},
		{"AviSettingServicesIndex", lib.AviSettingServicesIndex},
		{"AviSettingIngClassIndex", lib.AviSettingIngClassIndex},
		{"AviSettingRouteIndex", lib.AviSettingRouteIndex},
		{"L4RuleToServicesIndex", lib.L4RuleToServicesIndex},
		{"AviSettingNamespaceIndex", lib.AviSettingNamespaceIndex},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant == "" {
				t.Errorf("%s should not be empty", tt.name)
			}
		})
	}
}

func TestLicenseTypeConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{"LicenseTypeEnterprise", lib.LicenseTypeEnterprise, "ENTERPRISE"},
		{"LicenseTypeEnterpriseCloudServices", lib.LicenseTypeEnterpriseCloudServices, "ENTERPRISE_WITH_CLOUD_SERVICES"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("%s = %v, want %v", tt.name, tt.constant, tt.expected)
			}
		})
	}
}

// ========== Tests from avi_api_test.go ==========

func TestCheckForInvalidCredentials(t *testing.T) {
	tests := []struct {
		name string
		uri  string
		err  error
	}{
		{
			name: "Nil error",
			uri:  "/api/test",
			err:  nil,
		},
		{
			name: "Non-401 error",
			uri:  "/api/test",
			err:  errors.New("some error"),
		},
		{
			name: "401 error without invalid credentials message",
			uri:  "/api/test",
			err:  errors.New("401 Unauthorized"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This function should not panic
			lib.CheckForInvalidCredentials(tt.uri, tt.err)
		})
	}
}

func TestGetControllerIP(t *testing.T) {
	tests := []struct {
		name     string
		setIP    string
		expected string
	}{
		{
			name:     "Set and get IP",
			setIP:    "10.10.10.10",
			expected: "10.10.10.10",
		},
		{
			name:     "Set and get another IP",
			setIP:    "192.168.1.1",
			expected: "192.168.1.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lib.SetControllerIP(tt.setIP)
			got := lib.GetControllerIP()
			if got != tt.expected {
				t.Errorf("GetControllerIP() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestVSVIPNotFoundError(t *testing.T) {
	if lib.VSVIPNotFoundError == "" {
		t.Error("VSVIPNotFoundError should not be empty")
	}

	if !strings.Contains(lib.VSVIPNotFoundError, "VsVip") {
		t.Error("VSVIPNotFoundError should contain 'VsVip'")
	}
}
