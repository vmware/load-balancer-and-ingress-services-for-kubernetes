package cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-crd-operator/internal/session"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

func TestNewCache(t *testing.T) {
	mockSession := &session.Session{}
	clusterName := "test-cluster"

	cache := NewCache(mockSession, clusterName)

	assert.NotNil(t, cache)
	assert.Implements(t, (*CacheOperation)(nil), cache)
}

func TestDataMap_GetLastModifiedTimeStamp(t *testing.T) {
	tests := []struct {
		name     string
		dataMap  DataMap
		expected time.Time
	}{
		{
			name: "valid timestamp",
			dataMap: DataMap{
				"_last_modified": "1640995200000000", // 2022-01-01 00:00:00 UTC in microseconds
			},
			expected: time.UnixMicro(1640995200000000).UTC(),
		},
		{
			name:     "missing timestamp",
			dataMap:  DataMap{},
			expected: time.Unix(0, 0).UTC(),
		},
		{
			name: "invalid timestamp format",
			dataMap: DataMap{
				"_last_modified": "invalid",
			},
			expected: time.Unix(0, 0).UTC(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.dataMap.GetLastModifiedTimeStamp()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCache_GetObjectByUUID(t *testing.T) {
	tests := []struct {
		name       string
		uuid       string
		setupCache func(*cache)
		expectData DataMap
		expectOk   bool
	}{
		{
			name: "existing object",
			uuid: "test-uuid-1",
			setupCache: func(c *cache) {
				c.dataStore.Store("test-uuid-1", map[string]interface{}{
					"uuid":           "test-uuid-1",
					"_last_modified": "1640995200000000",
					"name":           "test-object",
				})
			},
			expectData: DataMap{
				"uuid":           "test-uuid-1",
				"_last_modified": "1640995200000000",
				"name":           "test-object",
			},
			expectOk: true,
		},
		{
			name: "non-existing object",
			uuid: "non-existing-uuid",
			setupCache: func(c *cache) {
				// Don't store anything
			},
			expectData: nil,
			expectOk:   false,
		},
		{
			name: "object with different UUID",
			uuid: "different-uuid",
			setupCache: func(c *cache) {
				c.dataStore.Store("test-uuid-1", map[string]interface{}{
					"uuid": "test-uuid-1",
					"name": "test-object",
				})
			},
			expectData: nil,
			expectOk:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := &cache{
				clusterName: "test-cluster",
			}

			if tt.setupCache != nil {
				tt.setupCache(cache)
			}

			ctx := utils.LoggerWithContext(context.Background(), utils.AviLog)
			data, ok := cache.GetObjectByUUID(ctx, tt.uuid)

			assert.Equal(t, tt.expectOk, ok)
			if tt.expectOk {
				assert.Equal(t, tt.expectData, data)
			} else {
				assert.Nil(t, data)
			}
		})
	}
}

func TestCache_PopulateCache(t *testing.T) {
	tests := []struct {
		name         string
		urls         []string
		setupCache   func(*cache)
		expectError  bool
		expectedData map[string]interface{}
	}{
		{
			name: "successful population with single URL",
			urls: []string{"api/healthmonitor"},
			setupCache: func(c *cache) {
				c.dataStore.Store("test-uuid-1", map[string]interface{}{
					"uuid":           "test-uuid-1",
					"_last_modified": "1640995200000000",
					"name":           "test-object",
				})
			},
			expectError: false,
			expectedData: map[string]interface{}{
				"uuid":           "test-uuid-1",
				"_last_modified": "1640995200000000",
				"name":           "test-object",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := &cache{
				clusterName: "test-cluster",
			}

			if tt.setupCache != nil {
				tt.setupCache(cache)
			}

			ctx := utils.LoggerWithContext(context.Background(), utils.AviLog)
			err := cache.PopulateCache(ctx, tt.urls...)
			assert.Equal(t, tt.expectError, err != nil)
			if !tt.expectError {
				assert.NoError(t, err)
			}
		})
	}
}
