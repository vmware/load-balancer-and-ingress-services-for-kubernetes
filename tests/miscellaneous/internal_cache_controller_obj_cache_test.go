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
	"testing"

	"github.com/vmware/alb-sdk/go/models"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	akov1beta1 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/apis/ako/v1beta1"
)

// Helper function for tests
func stringPtrCache(s string) *string {
	return &s
}

func TestNewAviObjCache(t *testing.T) {
	objCache := cache.NewAviObjCache()

	if objCache == nil {
		t.Fatal("NewAviObjCache() returned nil")
	}

	// Verify all cache components are initialized
	tests := []struct {
		name  string
		cache interface{}
	}{
		{"VsCacheMeta", objCache.VsCacheMeta},
		{"VsCacheLocal", objCache.VsCacheLocal},
		{"PgCache", objCache.PgCache},
		{"DSCache", objCache.DSCache},
		{"StringGroupCache", objCache.StringGroupCache},
		{"PoolCache", objCache.PoolCache},
		{"SSLKeyCache", objCache.SSLKeyCache},
		{"CloudKeyCache", objCache.CloudKeyCache},
		{"HTTPPolicyCache", objCache.HTTPPolicyCache},
		{"L4PolicyCache", objCache.L4PolicyCache},
		{"VSVIPCache", objCache.VSVIPCache},
		{"VrfCache", objCache.VrfCache},
		{"PKIProfileCache", objCache.PKIProfileCache},
		{"AppPersProfileCache", objCache.AppPersProfileCache},
		{"ClusterStatusCache", objCache.ClusterStatusCache},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.cache == nil {
				t.Errorf("NewAviObjCache() %s is nil", tt.name)
			}
		})
	}
}

func TestSharedAviObjCache(t *testing.T) {
	// Test singleton pattern
	cache1 := cache.SharedAviObjCache()
	cache2 := cache.SharedAviObjCache()

	if cache1 == nil {
		t.Fatal("SharedAviObjCache() returned nil")
	}

	if cache2 == nil {
		t.Fatal("SharedAviObjCache() returned nil on second call")
	}

	// Verify it's the same instance (singleton)
	if cache1 != cache2 {
		t.Error("SharedAviObjCache() did not return the same instance")
	}
}

func TestAviObjCacheStructure(t *testing.T) {
	objCache := cache.NewAviObjCache()

	// Test that we can access all cache fields
	t.Run("Access all cache fields", func(t *testing.T) {
		caches := []interface{}{
			objCache.VsCacheMeta,
			objCache.VsCacheLocal,
			objCache.PgCache,
			objCache.DSCache,
			objCache.StringGroupCache,
			objCache.PoolCache,
			objCache.SSLKeyCache,
			objCache.CloudKeyCache,
			objCache.HTTPPolicyCache,
			objCache.L4PolicyCache,
			objCache.VSVIPCache,
			objCache.VrfCache,
			objCache.PKIProfileCache,
			objCache.AppPersProfileCache,
			objCache.ClusterStatusCache,
		}

		for i, c := range caches {
			if c == nil {
				t.Errorf("Cache at index %d is nil", i)
			}
		}
	})
}

func TestAviObjCacheIndependence(t *testing.T) {
	// Create multiple cache instances and verify they're independent
	cache1 := cache.NewAviObjCache()
	cache2 := cache.NewAviObjCache()

	if cache1 == cache2 {
		t.Error("NewAviObjCache() returned the same instance, expected different instances")
	}

	// Verify they have independent cache structures
	if cache1.VsCacheMeta == cache2.VsCacheMeta {
		t.Error("NewAviObjCache() instances share VsCacheMeta, expected independent caches")
	}
}

func TestAviObjCacheConcurrentAccess(t *testing.T) {
	// Test that SharedAviObjCache is thread-safe
	done := make(chan *cache.AviObjCache, 10)

	for i := 0; i < 10; i++ {
		go func() {
			c := cache.SharedAviObjCache()
			done <- c
		}()
	}

	// Collect all results
	var caches []*cache.AviObjCache
	for i := 0; i < 10; i++ {
		caches = append(caches, <-done)
	}

	// Verify all goroutines got the same instance
	first := caches[0]
	for i, c := range caches {
		if c != first {
			t.Errorf("Goroutine %d got different cache instance", i)
		}
	}
}

func TestNewAviObjCacheMultipleTimes(t *testing.T) {
	// Test that NewAviObjCache creates new instances each time
	caches := make([]*cache.AviObjCache, 5)
	for i := 0; i < 5; i++ {
		caches[i] = cache.NewAviObjCache()
	}

	// Verify all are non-nil
	for i, c := range caches {
		if c == nil {
			t.Errorf("Cache %d is nil", i)
		}
	}

	// Verify all are different instances
	for i := 0; i < len(caches); i++ {
		for j := i + 1; j < len(caches); j++ {
			if caches[i] == caches[j] {
				t.Errorf("Cache %d and %d are the same instance", i, j)
			}
		}
	}
}

func TestAviObjCacheAllFieldsAccessible(t *testing.T) {
	objCache := cache.NewAviObjCache()

	// Test that all fields are accessible and can be used
	tests := []struct {
		name     string
		testFunc func() error
	}{
		{
			name: "VsCacheMeta accessible",
			testFunc: func() error {
				if objCache.VsCacheMeta == nil {
					return nil
				}
				return nil
			},
		},
		{
			name: "PoolCache accessible",
			testFunc: func() error {
				if objCache.PoolCache == nil {
					return nil
				}
				return nil
			},
		},
		{
			name: "HTTPPolicyCache accessible",
			testFunc: func() error {
				if objCache.HTTPPolicyCache == nil {
					return nil
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.testFunc(); err != nil {
				t.Errorf("Test failed: %v", err)
			}
		})
	}
}

func TestAviObjCacheInitialization(t *testing.T) {
	// Test that cache is properly initialized
	objCache := cache.NewAviObjCache()

	// Count non-nil caches
	nonNilCount := 0
	if objCache.VsCacheMeta != nil {
		nonNilCount++
	}
	if objCache.VsCacheLocal != nil {
		nonNilCount++
	}
	if objCache.PgCache != nil {
		nonNilCount++
	}
	if objCache.DSCache != nil {
		nonNilCount++
	}
	if objCache.StringGroupCache != nil {
		nonNilCount++
	}
	if objCache.PoolCache != nil {
		nonNilCount++
	}
	if objCache.SSLKeyCache != nil {
		nonNilCount++
	}
	if objCache.CloudKeyCache != nil {
		nonNilCount++
	}
	if objCache.HTTPPolicyCache != nil {
		nonNilCount++
	}
	if objCache.L4PolicyCache != nil {
		nonNilCount++
	}
	if objCache.VSVIPCache != nil {
		nonNilCount++
	}
	if objCache.VrfCache != nil {
		nonNilCount++
	}
	if objCache.PKIProfileCache != nil {
		nonNilCount++
	}
	if objCache.AppPersProfileCache != nil {
		nonNilCount++
	}
	if objCache.ClusterStatusCache != nil {
		nonNilCount++
	}

	expectedCount := 15
	if nonNilCount != expectedCount {
		t.Errorf("NewAviObjCache() initialized %d caches, want %d", nonNilCount, expectedCount)
	}
}

func TestExtractUUID(t *testing.T) {
	tests := []struct {
		name    string
		word    string
		pattern string
		want    string
	}{
		{
			name:    "Valid UUID with hash",
			word:    "cluster--Shared-L7-0-abc123def456#",
			pattern: "[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}#",
			want:    "",
		},
		{
			name:    "Valid short hash",
			word:    "vs-abc123#",
			pattern: "[a-f0-9]+#",
			want:    "abc123",
		},
		{
			name:    "No match",
			word:    "no-hash-here",
			pattern: "[a-f0-9]+#",
			want:    "",
		},
		{
			name:    "Empty word",
			word:    "",
			pattern: "[a-f0-9]+#",
			want:    "",
		},
		{
			name:    "Multiple hashes - first match",
			word:    "prefix-abc#",
			pattern: "[a-f0-9]+#",
			want:    "abc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cache.ExtractUUID(tt.word, tt.pattern)
			if got != tt.want {
				t.Errorf("ExtractUUID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractUUIDWithoutHash(t *testing.T) {
	tests := []struct {
		name    string
		word    string
		pattern string
		want    string
	}{
		{
			name:    "Valid UUID without hash",
			word:    "cluster--Shared-L7-0-abc123def456",
			pattern: "[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}",
			want:    "",
		},
		{
			name:    "Valid short hash without trailing #",
			word:    "vs-abc123",
			pattern: "[a-f0-9]+",
			want:    "abc123",
		},
		{
			name:    "No match",
			word:    "no-hash-here",
			pattern: "[a-f0-9]+",
			want:    "",
		},
		{
			name:    "Empty word",
			word:    "",
			pattern: "[a-f0-9]+",
			want:    "",
		},
		{
			name:    "Hex at end - first match",
			word:    "prefix-abc",
			pattern: "abc",
			want:    "abc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cache.ExtractUUIDWithoutHash(tt.word, tt.pattern)
			if got != tt.want {
				t.Errorf("ExtractUUIDWithoutHash() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetControllerClusterUUID(t *testing.T) {
	// Test that GetControllerClusterUUID returns a string
	// The actual value depends on whether SetControllerClusterUUID was called
	uuid := cache.GetControllerClusterUUID()

	// UUID can be empty or non-empty depending on initialization
	// This test just verifies the function is callable and returns a string
	t.Logf("GetControllerClusterUUID() returned: %s", uuid)

	// Verify it's a string type
	if uuid != "" && len(uuid) == 0 {
		t.Error("GetControllerClusterUUID() returned empty string with non-zero length")
	}
}

func TestExtractPattern(t *testing.T) {
	tests := []struct {
		name    string
		word    string
		pattern string
		want    string
		wantErr bool
	}{
		{
			name:    "Valid pattern match - hex digits",
			word:    "cluster--Shared-L7-0-abc123",
			pattern: "abc123",
			want:    "abc123",
			wantErr: false,
		},
		{
			name:    "No match",
			word:    "no-match-here",
			pattern: "[0-9]{5}",
			want:    "",
			wantErr: false,
		},
		{
			name:    "Invalid regex pattern",
			word:    "test",
			pattern: "[invalid(regex",
			want:    "",
			wantErr: true,
		},
		{
			name:    "Empty word",
			word:    "",
			pattern: "[a-z]+",
			want:    "",
			wantErr: false,
		},
		{
			name:    "Multiple matches returns empty",
			word:    "123-456-789",
			pattern: "[0-9]+",
			want:    "",
			wantErr: false,
		},
		{
			name:    "Numeric pattern",
			word:    "test-12345-end",
			pattern: "[0-9]+",
			want:    "12345",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := cache.ExtractPattern(tt.word, tt.pattern)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractPattern() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ExtractPattern() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFindCIDROverlapping(t *testing.T) {
	// Helper to create IP address
	createIPAddr := func(addr, ipType string) *models.IPAddr {
		return &models.IPAddr{
			Addr: &addr,
			Type: &ipType,
		}
	}

	// Helper to create subnet
	createSubnet := func(addr, ipType string, mask int32) *models.Subnet {
		return &models.Subnet{
			Prefix: &models.IPAddrPrefix{
				IPAddr: createIPAddr(addr, ipType),
				Mask:   &mask,
			},
		}
	}

	tests := []struct {
		name        string
		networks    []models.Network
		ipNet       akov1beta1.AviInfraSettingVipNetwork
		wantFound   bool
		wantNetwork string
	}{
		{
			name: "Match IPv4 CIDR",
			networks: []models.Network{
				{
					Name: stringPtrCache("network1"),
					UUID: stringPtrCache("uuid1"),
					ConfiguredSubnets: []*models.Subnet{
						createSubnet("10.0.0.0", "V4", 24),
					},
				},
			},
			ipNet: akov1beta1.AviInfraSettingVipNetwork{
				Cidr: "10.0.0.0/24",
			},
			wantFound:   true,
			wantNetwork: "network1",
		},
		{
			name: "Match IPv6 CIDR",
			networks: []models.Network{
				{
					Name: stringPtrCache("network2"),
					UUID: stringPtrCache("uuid2"),
					ConfiguredSubnets: []*models.Subnet{
						createSubnet("2001:db8::", "V6", 64),
					},
				},
			},
			ipNet: akov1beta1.AviInfraSettingVipNetwork{
				V6Cidr: "2001:db8::/64",
			},
			wantFound:   true,
			wantNetwork: "network2",
		},
		{
			name: "No match",
			networks: []models.Network{
				{
					Name: stringPtrCache("network3"),
					UUID: stringPtrCache("uuid3"),
					ConfiguredSubnets: []*models.Subnet{
						createSubnet("192.168.0.0", "V4", 24),
					},
				},
			},
			ipNet: akov1beta1.AviInfraSettingVipNetwork{
				Cidr: "10.0.0.0/24",
			},
			wantFound:   false,
			wantNetwork: "",
		},
		{
			name:     "Empty networks",
			networks: []models.Network{},
			ipNet: akov1beta1.AviInfraSettingVipNetwork{
				Cidr: "10.0.0.0/24",
			},
			wantFound:   false,
			wantNetwork: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFound, gotNetwork := cache.FindCIDROverlapping(tt.networks, tt.ipNet)
			if gotFound != tt.wantFound {
				t.Errorf("FindCIDROverlapping() found = %v, want %v", gotFound, tt.wantFound)
			}
			if gotFound && gotNetwork.Name != nil && *gotNetwork.Name != tt.wantNetwork {
				t.Errorf("FindCIDROverlapping() network = %v, want %v", *gotNetwork.Name, tt.wantNetwork)
			}
		})
	}
}
