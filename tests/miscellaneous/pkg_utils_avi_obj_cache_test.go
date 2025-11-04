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

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

func TestSharedCtrlProp(t *testing.T) {
	t.Run("Singleton instance - underlying cache is shared", func(t *testing.T) {
		// Get the singleton instance multiple times
		instance1 := utils.SharedCtrlProp()
		instance2 := utils.SharedCtrlProp()

		// Verify instances are not nil
		if instance1 == nil || instance2 == nil {
			t.Fatal("SharedCtrlProp() returned nil")
		}

		// Verify the underlying cache is shared by adding data through one instance
		// and retrieving through another
		testKey := "singleton-test-key"
		testValue := "singleton-test-value"

		instance1.AviCacheAdd(testKey, testValue)
		val, ok := instance2.AviCacheGet(testKey)

		if !ok {
			t.Errorf("SharedCtrlProp() underlying cache is not shared")
		}

		if val != testValue {
			t.Errorf("SharedCtrlProp() value from second instance = %v, want %v", val, testValue)
		}

		// Cleanup
		instance1.AviCacheDelete(testKey)
	})

	t.Run("Instance persists data", func(t *testing.T) {
		cache := utils.SharedCtrlProp()
		testKey := "test-key-singleton"
		testValue := "test-value-singleton"

		// Add data to the cache
		cache.AviCacheAdd(testKey, testValue)

		// Get a new reference and verify data persists
		cache2 := utils.SharedCtrlProp()
		val, ok := cache2.AviCacheGet(testKey)

		if !ok {
			t.Errorf("SharedCtrlProp() data not persisted across calls")
		}

		if val != testValue {
			t.Errorf("SharedCtrlProp() value = %v, want %v", val, testValue)
		}

		// Cleanup
		cache.AviCacheDelete(testKey)
	})
}

func TestPopulateCtrlProp(t *testing.T) {
	tests := []struct {
		name  string
		props map[string]string
	}{
		{
			name: "Single property",
			props: map[string]string{
				"key1": "value1",
			},
		},
		{
			name: "Multiple properties",
			props: map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			},
		},
		{
			name:  "Empty properties",
			props: map[string]string{},
		},
		{
			name: "Properties with empty values",
			props: map[string]string{
				"key1": "",
				"key2": "value2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := utils.SharedCtrlProp()

			// Populate the cache
			cache.PopulateCtrlProp(tt.props)

			// Verify all properties were added
			for k, expectedVal := range tt.props {
				val, ok := cache.AviCacheGet(k)
				if !ok {
					t.Errorf("PopulateCtrlProp() key %s not found in cache", k)
					continue
				}
				if val != expectedVal {
					t.Errorf("PopulateCtrlProp() key %s = %v, want %v", k, val, expectedVal)
				}
			}

			// Cleanup
			for k := range tt.props {
				cache.AviCacheDelete(k)
			}
		})
	}
}

func TestPopulateCtrlAPIUserHeaders(t *testing.T) {
	tests := []struct {
		name       string
		userHeader map[string]string
	}{
		{
			name: "Single header",
			userHeader: map[string]string{
				"X-Custom-Header": "value1",
			},
		},
		{
			name: "Multiple headers",
			userHeader: map[string]string{
				"X-Custom-Header-1": "value1",
				"X-Custom-Header-2": "value2",
				"Authorization":     "Bearer token",
			},
		},
		{
			name:       "Empty headers",
			userHeader: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := utils.SharedCtrlProp()

			// Populate user headers
			cache.PopulateCtrlAPIUserHeaders(tt.userHeader)

			// Verify headers were added
			val, ok := cache.AviCacheGet(utils.ControllerAPIHeader)
			if !ok {
				t.Errorf("PopulateCtrlAPIUserHeaders() header not found in cache")
				return
			}

			headers, ok := val.(map[string]string)
			if !ok {
				t.Errorf("PopulateCtrlAPIUserHeaders() value is not map[string]string")
				return
			}

			// Verify all headers match
			if len(headers) != len(tt.userHeader) {
				t.Errorf("PopulateCtrlAPIUserHeaders() header count = %v, want %v", len(headers), len(tt.userHeader))
			}

			for k, expectedVal := range tt.userHeader {
				if headers[k] != expectedVal {
					t.Errorf("PopulateCtrlAPIUserHeaders() header %s = %v, want %v", k, headers[k], expectedVal)
				}
			}

			// Cleanup
			cache.AviCacheDelete(utils.ControllerAPIHeader)
		})
	}
}

func TestPopulateCtrlAPIScheme(t *testing.T) {
	tests := []struct {
		name   string
		scheme string
	}{
		{
			name:   "HTTP scheme",
			scheme: "http",
		},
		{
			name:   "HTTPS scheme",
			scheme: "https",
		},
		{
			name:   "Empty scheme",
			scheme: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := utils.SharedCtrlProp()

			// Populate scheme
			cache.PopulateCtrlAPIScheme(tt.scheme)

			// Verify scheme was added
			val, ok := cache.AviCacheGet(utils.ControllerAPIScheme)
			if !ok {
				t.Errorf("PopulateCtrlAPIScheme() scheme not found in cache")
				return
			}

			scheme, ok := val.(string)
			if !ok {
				t.Errorf("PopulateCtrlAPIScheme() value is not string")
				return
			}

			if scheme != tt.scheme {
				t.Errorf("PopulateCtrlAPIScheme() = %v, want %v", scheme, tt.scheme)
			}

			// Cleanup
			cache.AviCacheDelete(utils.ControllerAPIScheme)
		})
	}
}

func TestGetAllCtrlProp(t *testing.T) {
	tests := []struct {
		name     string
		setup    map[string]string
		expected map[string]string
	}{
		{
			name: "All properties present",
			setup: map[string]string{
				utils.ENV_CTRL_USERNAME:  "admin",
				utils.ENV_CTRL_PASSWORD:  "password123",
				utils.ENV_CTRL_AUTHTOKEN: "token123",
				utils.ENV_CTRL_CADATA:    "ca-data",
			},
			expected: map[string]string{
				utils.ENV_CTRL_USERNAME:  "admin",
				utils.ENV_CTRL_PASSWORD:  "password123",
				utils.ENV_CTRL_AUTHTOKEN: "token123",
				utils.ENV_CTRL_CADATA:    "ca-data",
			},
		},
		{
			name: "Some properties missing",
			setup: map[string]string{
				utils.ENV_CTRL_USERNAME: "admin",
				utils.ENV_CTRL_PASSWORD: "password123",
			},
			expected: map[string]string{
				utils.ENV_CTRL_USERNAME:  "admin",
				utils.ENV_CTRL_PASSWORD:  "password123",
				utils.ENV_CTRL_AUTHTOKEN: "",
				utils.ENV_CTRL_CADATA:    "",
			},
		},
		{
			name:  "No properties present",
			setup: map[string]string{},
			expected: map[string]string{
				utils.ENV_CTRL_USERNAME:  "",
				utils.ENV_CTRL_PASSWORD:  "",
				utils.ENV_CTRL_AUTHTOKEN: "",
				utils.ENV_CTRL_CADATA:    "",
			},
		},
		{
			name: "Empty string values",
			setup: map[string]string{
				utils.ENV_CTRL_USERNAME:  "",
				utils.ENV_CTRL_PASSWORD:  "password123",
				utils.ENV_CTRL_AUTHTOKEN: "",
				utils.ENV_CTRL_CADATA:    "ca-data",
			},
			expected: map[string]string{
				utils.ENV_CTRL_USERNAME:  "",
				utils.ENV_CTRL_PASSWORD:  "password123",
				utils.ENV_CTRL_AUTHTOKEN: "",
				utils.ENV_CTRL_CADATA:    "ca-data",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := utils.SharedCtrlProp()

			// Setup cache with test data
			for k, v := range tt.setup {
				cache.AviCacheAdd(k, v)
			}

			// Get all properties
			result := cache.GetAllCtrlProp()

			// Verify all expected properties are present
			for k, expectedVal := range tt.expected {
				val, ok := result[k]
				if !ok {
					t.Errorf("GetAllCtrlProp() missing key %s", k)
					continue
				}
				if val != expectedVal {
					t.Errorf("GetAllCtrlProp() key %s = %v, want %v", k, val, expectedVal)
				}
			}

			// Verify no extra keys
			if len(result) != len(tt.expected) {
				t.Errorf("GetAllCtrlProp() returned %d keys, want %d", len(result), len(tt.expected))
			}

			// Cleanup
			for k := range tt.setup {
				cache.AviCacheDelete(k)
			}
		})
	}
}

func TestGetCtrlUserHeader(t *testing.T) {
	tests := []struct {
		name     string
		setup    interface{}
		expected map[string]string
	}{
		{
			name: "Valid headers present",
			setup: map[string]string{
				"X-Custom-Header": "value1",
				"Authorization":   "Bearer token",
			},
			expected: map[string]string{
				"X-Custom-Header": "value1",
				"Authorization":   "Bearer token",
			},
		},
		{
			name:     "No headers present",
			setup:    nil,
			expected: map[string]string{},
		},
		{
			name:     "Empty headers map",
			setup:    map[string]string{},
			expected: map[string]string{},
		},
		{
			name:     "Invalid type (not map[string]string)",
			setup:    "invalid-type",
			expected: map[string]string{},
		},
		{
			name:     "Invalid type (map with different types)",
			setup:    map[string]int{"key": 123},
			expected: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := utils.SharedCtrlProp()

			// Setup cache
			if tt.setup != nil {
				cache.AviCacheAdd(utils.ControllerAPIHeader, tt.setup)
			}

			// Get user headers
			result := cache.GetCtrlUserHeader()

			// Verify result
			if len(result) != len(tt.expected) {
				t.Errorf("GetCtrlUserHeader() returned %d headers, want %d", len(result), len(tt.expected))
			}

			for k, expectedVal := range tt.expected {
				val, ok := result[k]
				if !ok {
					t.Errorf("GetCtrlUserHeader() missing header %s", k)
					continue
				}
				if val != expectedVal {
					t.Errorf("GetCtrlUserHeader() header %s = %v, want %v", k, val, expectedVal)
				}
			}

			// Cleanup
			cache.AviCacheDelete(utils.ControllerAPIHeader)
		})
	}
}

func TestGetCtrlAPIScheme(t *testing.T) {
	tests := []struct {
		name     string
		setup    interface{}
		expected string
	}{
		{
			name:     "HTTPS scheme",
			setup:    "https",
			expected: "https",
		},
		{
			name:     "HTTP scheme",
			setup:    "http",
			expected: "http",
		},
		{
			name:     "Empty string scheme",
			setup:    "",
			expected: "",
		},
		{
			name:     "No scheme present",
			setup:    nil,
			expected: "",
		},
		{
			name:     "Invalid type (not string)",
			setup:    123,
			expected: "",
		},
		{
			name:     "Invalid type (map)",
			setup:    map[string]string{"scheme": "https"},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := utils.SharedCtrlProp()

			// Setup cache
			if tt.setup != nil {
				cache.AviCacheAdd(utils.ControllerAPIScheme, tt.setup)
			}

			// Get API scheme
			result := cache.GetCtrlAPIScheme()

			// Verify result
			if result != tt.expected {
				t.Errorf("GetCtrlAPIScheme() = %v, want %v", result, tt.expected)
			}

			// Cleanup
			cache.AviCacheDelete(utils.ControllerAPIScheme)
		})
	}
}

func TestCtrlPropCacheIntegration(t *testing.T) {
	t.Run("Complete workflow", func(t *testing.T) {
		cache := utils.SharedCtrlProp()

		// 1. Populate controller properties
		ctrlProps := map[string]string{
			utils.ENV_CTRL_USERNAME:  "test-user",
			utils.ENV_CTRL_PASSWORD:  "test-pass",
			utils.ENV_CTRL_AUTHTOKEN: "test-token",
			utils.ENV_CTRL_CADATA:    "test-ca-data",
		}
		cache.PopulateCtrlProp(ctrlProps)

		// 2. Populate API headers
		headers := map[string]string{
			"X-Custom-Header": "custom-value",
			"X-Request-ID":    "req-123",
		}
		cache.PopulateCtrlAPIUserHeaders(headers)

		// 3. Populate API scheme
		cache.PopulateCtrlAPIScheme("https")

		// 4. Verify all properties can be retrieved
		allProps := cache.GetAllCtrlProp()
		if allProps[utils.ENV_CTRL_USERNAME] != "test-user" {
			t.Errorf("Integration test: username not retrieved correctly")
		}

		// 5. Verify headers can be retrieved
		retrievedHeaders := cache.GetCtrlUserHeader()
		if retrievedHeaders["X-Custom-Header"] != "custom-value" {
			t.Errorf("Integration test: headers not retrieved correctly")
		}

		// 6. Verify scheme can be retrieved
		scheme := cache.GetCtrlAPIScheme()
		if scheme != "https" {
			t.Errorf("Integration test: scheme not retrieved correctly")
		}

		// Cleanup
		for k := range ctrlProps {
			cache.AviCacheDelete(k)
		}
		cache.AviCacheDelete(utils.ControllerAPIHeader)
		cache.AviCacheDelete(utils.ControllerAPIScheme)
	})
}

func TestCtrlPropCacheConcurrency(t *testing.T) {
	t.Run("Concurrent access", func(t *testing.T) {
		cache := utils.SharedCtrlProp()

		// Test concurrent reads and writes
		done := make(chan bool)
		iterations := 100

		// Writer goroutine
		go func() {
			for i := 0; i < iterations; i++ {
				cache.AviCacheAdd("concurrent-key", i)
			}
			done <- true
		}()

		// Reader goroutine
		go func() {
			for i := 0; i < iterations; i++ {
				cache.AviCacheGet("concurrent-key")
			}
			done <- true
		}()

		// Wait for both goroutines
		<-done
		<-done

		// Cleanup
		cache.AviCacheDelete("concurrent-key")
	})
}
