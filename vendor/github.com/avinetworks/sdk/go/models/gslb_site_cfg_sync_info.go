package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// GslbSiteCfgSyncInfo gslb site cfg sync info
// swagger:model GslbSiteCfgSyncInfo
type GslbSiteCfgSyncInfo struct {

	// Objects that could NOT be synced to the site .
	ErroredObjects []*VersionInfo `json:"errored_objects,omitempty"`

	// Placeholder for description of property last_changed_time of obj type GslbSiteCfgSyncInfo field type str  type object
	LastChangedTime *TimeStamp `json:"last_changed_time,omitempty"`

	// Configuration sync-state of the site . Enum options - GSLB_SITE_CFG_IN_SYNC, GSLB_SITE_CFG_OUT_OF_SYNC, GSLB_SITE_CFG_SYNC_DISABLED, GSLB_SITE_CFG_SYNC_IN_PROGRESS, GSLB_SITE_CFG_SYNC_NOT_APPLICABLE.
	SyncState *string `json:"sync_state,omitempty"`
}
