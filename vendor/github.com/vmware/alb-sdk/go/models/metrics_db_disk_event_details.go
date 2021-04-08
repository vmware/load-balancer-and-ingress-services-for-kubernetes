package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// MetricsDbDiskEventDetails metrics db disk event details
// swagger:model MetricsDbDiskEventDetails
type MetricsDbDiskEventDetails struct {

	// metrics_deleted_tables of MetricsDbDiskEventDetails.
	MetricsDeletedTables []string `json:"metrics_deleted_tables,omitempty"`

	// Number of metrics_free_sz.
	// Required: true
	MetricsFreeSz *int64 `json:"metrics_free_sz"`

	// Number of metrics_quota.
	// Required: true
	MetricsQuota *int64 `json:"metrics_quota"`
}
