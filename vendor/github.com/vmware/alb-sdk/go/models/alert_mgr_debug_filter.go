// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// AlertMgrDebugFilter alert mgr debug filter
// swagger:model AlertMgrDebugFilter
type AlertMgrDebugFilter struct {

	// filter debugs for entity uuid. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AlertObjid *string `json:"alert_objid,omitempty"`

	// filter debugs for an alert id. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AlertUUID *string `json:"alert_uuid,omitempty"`

	// filter debugs for an alert config. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CfgUUID *string `json:"cfg_uuid,omitempty"`
}
