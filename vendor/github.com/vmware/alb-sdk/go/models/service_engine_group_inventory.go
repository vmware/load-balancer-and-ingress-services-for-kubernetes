// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ServiceEngineGroupInventory service engine group inventory
// swagger:model ServiceEngineGroupInventory
type ServiceEngineGroupInventory struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Configuration summary of the service engine group. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Config *ServiceEngineGroup `json:"config,omitempty"`

	// Service engines the SE-Group is assigned to. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Serviceengins []*SeRefs `json:"serviceengins,omitempty"`

	// Upgrade status summary of the service engine group. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Upgradestatus *UpgradeStatusSummary `json:"upgradestatus,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// UUID of the service engine group. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	UUID *string `json:"uuid,omitempty"`

	// Virtual services the SE-Group is assigned to. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Virtualservices []*VsRefs `json:"virtualservices,omitempty"`
}
