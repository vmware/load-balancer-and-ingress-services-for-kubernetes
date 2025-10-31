// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// Tier1LogicalRouterInfo tier1 logical router info
// swagger:model Tier1LogicalRouterInfo
type Tier1LogicalRouterInfo struct {

	// Locale-services configuration, holds T1 edge-cluster information. When VirtualService is enabled with preserve client IP, ServiceInsertion VirtualEndpoint will be created in this locale-service. By default Avi controller picks default locale-service on T1. If more than one locale-services are present, this will be used for resolving the same. Example locale-service path - /infra/tier-1s/London_Tier1Gateway1/locale-services/London_Tier1LocalServices-1. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	LocaleService *string `json:"locale_service,omitempty"`

	// Overlay segment path. Example- /infra/segments/Seg-Web-T1-01. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SegmentID *string `json:"segment_id,omitempty"`

	// Tier1 logical router path. Example- /infra/tier-1s/T1-01. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Tier1LrID *string `json:"tier1_lr_id"`
}
