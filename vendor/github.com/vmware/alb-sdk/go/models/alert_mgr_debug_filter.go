// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// AlertMgrDebugFilter alert mgr debug filter
// swagger:model AlertMgrDebugFilter
type AlertMgrDebugFilter struct {

	// filter debugs for entity uuid.
	AlertObjid *string `json:"alert_objid,omitempty"`

	// filter debugs for an alert id.
	AlertUUID *string `json:"alert_uuid,omitempty"`

	// filter debugs for an alert config.
	CfgUUID *string `json:"cfg_uuid,omitempty"`
}
