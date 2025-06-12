package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

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
