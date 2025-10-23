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
	"encoding/json"
	"testing"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

func TestIsV4(t *testing.T) {
	tests := []struct {
		name string
		addr string
		want bool
	}{
		{
			name: "Valid IPv4 address",
			addr: "192.168.1.1",
			want: true,
		},
		{
			name: "Valid IPv4 localhost",
			addr: "127.0.0.1",
			want: true,
		},
		{
			name: "Valid IPv4 zero address",
			addr: "0.0.0.0",
			want: true,
		},
		{
			name: "IPv6 address",
			addr: "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
			want: false,
		},
		{
			name: "IPv6 localhost",
			addr: "::1",
			want: false,
		},
		{
			name: "Invalid IP address",
			addr: "invalid",
			want: false,
		},
		{
			name: "Empty string",
			addr: "",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := utils.IsV4(tt.addr); got != tt.want {
				t.Errorf("IsV4() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsSvcHttp(t *testing.T) {
	tests := []struct {
		name     string
		svcName  string
		port     int32
		expected bool
	}{
		{
			name:     "Service name is http",
			svcName:  "http",
			port:     9000,
			expected: true,
		},
		{
			name:     "Service name starts with http-",
			svcName:  "http-web",
			port:     9000,
			expected: true,
		},
		{
			name:     "Port 80",
			svcName:  "web",
			port:     80,
			expected: true,
		},
		{
			name:     "Port 443",
			svcName:  "web",
			port:     443,
			expected: true,
		},
		{
			name:     "Port 8080",
			svcName:  "web",
			port:     8080,
			expected: true,
		},
		{
			name:     "Port 8443",
			svcName:  "web",
			port:     8443,
			expected: true,
		},
		{
			name:     "Non-HTTP service and port",
			svcName:  "grpc",
			port:     9090,
			expected: false,
		},
		{
			name:     "Empty service name with non-standard port",
			svcName:  "",
			port:     9000,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := utils.IsSvcHttp(tt.svcName, tt.port)
			if got != tt.expected {
				t.Errorf("IsSvcHttp(%s, %d) = %v, want %v", tt.svcName, tt.port, got, tt.expected)
			}
		})
	}
}

func TestAviUrlToObjType(t *testing.T) {
	tests := []struct {
		name    string
		aviurl  string
		want    string
		wantErr bool
	}{
		{
			name:    "Valid virtualservice URL",
			aviurl:  "https://controller.local/api/virtualservice/virtualservice-123",
			want:    "virtualservice",
			wantErr: false,
		},
		{
			name:    "Valid pool URL",
			aviurl:  "https://controller.local/api/pool/pool-456",
			want:    "pool",
			wantErr: false,
		},
		{
			name:    "Valid poolgroup URL",
			aviurl:  "https://controller.local/api/poolgroup/pg-789",
			want:    "poolgroup",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := utils.AviUrlToObjType(tt.aviurl)
			if (err != nil) != tt.wantErr {
				t.Errorf("AviUrlToObjType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("AviUrlToObjType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHash(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "Simple string",
			input: "test",
		},
		{
			name:  "Empty string",
			input: "",
		},
		{
			name:  "Long string",
			input: "this is a very long string for testing hash function",
		},
		{
			name:  "String with special characters",
			input: "test@#$%^&*()",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash1 := utils.Hash(tt.input)
			hash2 := utils.Hash(tt.input)

			// Hash should be deterministic
			if hash1 != hash2 {
				t.Errorf("Hash() not deterministic: got %v and %v for same input", hash1, hash2)
			}
		})
	}

	// Test that different strings produce different hashes
	hash1 := utils.Hash("string1")
	hash2 := utils.Hash("string2")
	if hash1 == hash2 {
		t.Errorf("Hash() produced same hash for different strings")
	}
}

func TestBkt(t *testing.T) {
	tests := []struct {
		name        string
		key         string
		numWorkers  uint32
		expectRange bool
	}{
		{
			name:        "4 workers",
			key:         "test-key",
			numWorkers:  4,
			expectRange: true,
		},
		{
			name:        "8 workers",
			key:         "another-key",
			numWorkers:  8,
			expectRange: true,
		},
		{
			name:        "16 workers",
			key:         "namespace/name",
			numWorkers:  16,
			expectRange: true,
		},
		{
			name:        "1 worker",
			key:         "single",
			numWorkers:  1,
			expectRange: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := utils.Bkt(tt.key, tt.numWorkers)

			// Result should be less than numWorkers
			if got >= tt.numWorkers {
				t.Errorf("Bkt() = %v, should be less than %v", got, tt.numWorkers)
			}

			// Same key should always return same bucket
			got2 := utils.Bkt(tt.key, tt.numWorkers)
			if got != got2 {
				t.Errorf("Bkt() not deterministic: got %v and %v for same input", got, got2)
			}
		})
	}
}

func TestDeepCopy(t *testing.T) {
	type TestStruct struct {
		Name  string
		Value int
		Tags  []string
	}

	tests := []struct {
		name string
		src  TestStruct
	}{
		{
			name: "Simple struct",
			src: TestStruct{
				Name:  "test",
				Value: 42,
				Tags:  []string{"tag1", "tag2"},
			},
		},
		{
			name: "Empty struct",
			src:  TestStruct{},
		},
		{
			name: "Struct with nil slice",
			src: TestStruct{
				Name:  "test",
				Value: 10,
				Tags:  nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var dst TestStruct
			utils.DeepCopy(tt.src, &dst)

			// Check that values are copied
			if dst.Name != tt.src.Name {
				t.Errorf("DeepCopy() Name = %v, want %v", dst.Name, tt.src.Name)
			}
			if dst.Value != tt.src.Value {
				t.Errorf("DeepCopy() Value = %v, want %v", dst.Value, tt.src.Value)
			}

			// Verify it's a deep copy by modifying source
			tt.src.Name = "modified"
			if dst.Name == "modified" {
				t.Errorf("DeepCopy() did not create independent copy")
			}
		})
	}
}

func TestStringify(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{
			name:     "Simple map",
			input:    map[string]string{"key": "value"},
			expected: `{"key":"value"}`,
		},
		{
			name:     "Empty map",
			input:    map[string]string{},
			expected: `{}`,
		},
		{
			name:     "String",
			input:    "test",
			expected: `"test"`,
		},
		{
			name:     "Integer",
			input:    42,
			expected: `42`,
		},
		{
			name:     "Slice",
			input:    []string{"a", "b", "c"},
			expected: `["a","b","c"]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := utils.Stringify(tt.input)
			if got != tt.expected {
				t.Errorf("Stringify() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestExtractNamespaceObjectName(t *testing.T) {
	tests := []struct {
		name          string
		key           string
		wantNamespace string
		wantName      string
	}{
		{
			name:          "Two segments",
			key:           "default/my-service",
			wantNamespace: "default",
			wantName:      "my-service",
		},
		{
			name:          "Three segments",
			key:           "default/subpath/my-service",
			wantNamespace: "default",
			wantName:      "subpath/my-service",
		},
		{
			name:          "Single segment",
			key:           "single",
			wantNamespace: "",
			wantName:      "",
		},
		{
			name:          "Empty string",
			key:           "",
			wantNamespace: "",
			wantName:      "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotNs, gotName := utils.ExtractNamespaceObjectName(tt.key)
			if gotNs != tt.wantNamespace {
				t.Errorf("ExtractNamespaceObjectName() namespace = %v, want %v", gotNs, tt.wantNamespace)
			}
			if gotName != tt.wantName {
				t.Errorf("ExtractNamespaceObjectName() name = %v, want %v", gotName, tt.wantName)
			}
		})
	}
}

func TestHasElem(t *testing.T) {
	tests := []struct {
		name  string
		slice interface{}
		elem  interface{}
		want  bool
	}{
		{
			name:  "Element exists in string slice",
			slice: []string{"a", "b", "c"},
			elem:  "b",
			want:  true,
		},
		{
			name:  "Element does not exist in string slice",
			slice: []string{"a", "b", "c"},
			elem:  "d",
			want:  false,
		},
		{
			name:  "Element exists in int slice",
			slice: []int{1, 2, 3},
			elem:  2,
			want:  true,
		},
		{
			name:  "Element does not exist in int slice",
			slice: []int{1, 2, 3},
			elem:  4,
			want:  false,
		},
		{
			name:  "Empty slice",
			slice: []string{},
			elem:  "a",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := utils.HasElem(tt.slice, tt.elem)
			if got != tt.want {
				t.Errorf("HasElem() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRemove(t *testing.T) {
	tests := []struct {
		name     string
		arr      []string
		item     string
		expected []string
	}{
		{
			name:     "Remove existing item",
			arr:      []string{"a", "b", "c"},
			item:     "b",
			expected: []string{"a", "c"},
		},
		{
			name:     "Remove first item",
			arr:      []string{"a", "b", "c"},
			item:     "a",
			expected: []string{"b", "c"},
		},
		{
			name:     "Remove last item",
			arr:      []string{"a", "b", "c"},
			item:     "c",
			expected: []string{"a", "b"},
		},
		{
			name:     "Remove non-existing item",
			arr:      []string{"a", "b", "c"},
			item:     "d",
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "Remove from empty slice",
			arr:      []string{},
			item:     "a",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Make a copy to avoid modifying test data
			arrCopy := make([]string, len(tt.arr))
			copy(arrCopy, tt.arr)

			got := utils.Remove(arrCopy, tt.item)

			if len(got) != len(tt.expected) {
				t.Errorf("Remove() length = %v, want %v", len(got), len(tt.expected))
				return
			}

			for i := range got {
				if got[i] != tt.expected[i] {
					t.Errorf("Remove() = %v, want %v", got, tt.expected)
					break
				}
			}
		})
	}
}

func TestFindAndRemove(t *testing.T) {
	tests := []struct {
		name        string
		arr         []string
		item        string
		expectFound bool
		expectedArr []string
	}{
		{
			name:        "Find and remove existing item",
			arr:         []string{"a", "b", "c"},
			item:        "b",
			expectFound: true,
			expectedArr: []string{"a", "c"},
		},
		{
			name:        "Item not found",
			arr:         []string{"a", "b", "c"},
			item:        "d",
			expectFound: false,
			expectedArr: []string{"a", "b", "c"},
		},
		{
			name:        "Empty slice",
			arr:         []string{},
			item:        "a",
			expectFound: false,
			expectedArr: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Make a copy to avoid modifying test data
			arrCopy := make([]string, len(tt.arr))
			copy(arrCopy, tt.arr)

			found, got := utils.FindAndRemove(arrCopy, tt.item)

			if found != tt.expectFound {
				t.Errorf("FindAndRemove() found = %v, want %v", found, tt.expectFound)
			}

			if len(got) != len(tt.expectedArr) {
				t.Errorf("FindAndRemove() length = %v, want %v", len(got), len(tt.expectedArr))
				return
			}

			for i := range got {
				if got[i] != tt.expectedArr[i] {
					t.Errorf("FindAndRemove() = %v, want %v", got, tt.expectedArr)
					break
				}
			}
		})
	}
}

func TestRandomSeq(t *testing.T) {
	tests := []struct {
		name   string
		length int
	}{
		{
			name:   "Length 5",
			length: 5,
		},
		{
			name:   "Length 10",
			length: 10,
		},
		{
			name:   "Length 0",
			length: 0,
		},
		{
			name:   "Length 1",
			length: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := utils.RandomSeq(tt.length)

			if len(got) != tt.length {
				t.Errorf("RandomSeq(%d) length = %v, want %v", tt.length, len(got), tt.length)
			}

			// Verify all characters are from the allowed set
			allowedChars := "abcdefghijklmnopqrstuvwxyz1234567890"
			for _, char := range got {
				if !containsRune(allowedChars, char) {
					t.Errorf("RandomSeq() contains invalid character: %c", char)
				}
			}
		})
	}

	// Test that multiple calls produce different results (probabilistically)
	seq1 := utils.RandomSeq(10)
	seq2 := utils.RandomSeq(10)
	if seq1 == seq2 {
		// This could theoretically fail, but probability is very low
		t.Log("Warning: Two random sequences are identical (this is possible but unlikely)")
	}
}

func containsRune(s string, r rune) bool {
	for _, c := range s {
		if c == r {
			return true
		}
	}
	return false
}

func TestDeepCopyJSON(t *testing.T) {
	// Test that DeepCopy works with complex nested structures
	type Nested struct {
		Field1 string
		Field2 int
	}

	type Complex struct {
		Name   string
		Nested Nested
		List   []string
	}

	src := Complex{
		Name: "test",
		Nested: Nested{
			Field1: "nested",
			Field2: 42,
		},
		List: []string{"a", "b"},
	}

	var dst Complex
	utils.DeepCopy(src, &dst)

	// Verify deep copy
	if dst.Name != src.Name {
		t.Errorf("DeepCopy() Name not copied correctly")
	}
	if dst.Nested.Field1 != src.Nested.Field1 {
		t.Errorf("DeepCopy() Nested.Field1 not copied correctly")
	}
	if dst.Nested.Field2 != src.Nested.Field2 {
		t.Errorf("DeepCopy() Nested.Field2 not copied correctly")
	}

	// Modify source and verify destination is independent
	src.Nested.Field1 = "modified"
	if dst.Nested.Field1 == "modified" {
		t.Errorf("DeepCopy() did not create independent nested copy")
	}
}

func TestStringifyJSON(t *testing.T) {
	// Test that Stringify produces valid JSON
	input := map[string]interface{}{
		"string": "value",
		"number": 42,
		"bool":   true,
		"null":   nil,
	}

	result := utils.Stringify(input)

	// Verify it's valid JSON by unmarshaling
	var decoded map[string]interface{}
	err := json.Unmarshal([]byte(result), &decoded)
	if err != nil {
		t.Errorf("Stringify() produced invalid JSON: %v", err)
	}

	// Verify values
	if decoded["string"] != "value" {
		t.Errorf("Stringify() string value incorrect")
	}
	if decoded["number"].(float64) != 42 {
		t.Errorf("Stringify() number value incorrect")
	}
	if decoded["bool"] != true {
		t.Errorf("Stringify() bool value incorrect")
	}
}
