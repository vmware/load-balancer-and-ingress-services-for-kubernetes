// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// HSMgrDebugFilter h s mgr debug filter
// swagger:model HSMgrDebugFilter
type HSMgrDebugFilter struct {

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Entity *string `json:"entity,omitempty"`

	//  Enum options - VSERVER_METRICS_ENTITY, VM_METRICS_ENTITY, SE_METRICS_ENTITY, CONTROLLER_METRICS_ENTITY, APPLICATION_METRICS_ENTITY, TENANT_METRICS_ENTITY, POOL_METRICS_ENTITY. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MetricEntity *string `json:"metric_entity,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Period *int32 `json:"period,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Pool *string `json:"pool,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Server *string `json:"server,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SkipHsDbWrites *bool `json:"skip_hs_db_writes,omitempty"`
}
