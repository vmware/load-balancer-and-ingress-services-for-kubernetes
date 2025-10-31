// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ALBServicesJob a l b services job
// swagger:model ALBServicesJob
type ALBServicesJob struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// The command to be triggered by the albservicesjob. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	Command *string `json:"command"`

	// Protobuf versioning for config pbs. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	ConfigpbAttributes *ConfigPbAttributes `json:"configpb_attributes,omitempty"`

	// Time at which the albservicesjob ended. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	EndTime *TimeStamp `json:"end_time,omitempty"`

	// The name of the albservicesjob. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	Name *string `json:"name"`

	// Job params. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Params []*ALBServicesJobParam `json:"params,omitempty"`

	// A unique identifier for this job entry on the Pulse Cloud Services. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	PulseJobID *string `json:"pulse_job_id,omitempty"`

	// Status of sync to Pulse Cloud Services(result uploads/state updates). Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	PulseSyncStatus *bool `json:"pulse_sync_status,omitempty"`

	// Job result. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Result *string `json:"result,omitempty"`

	// Time at which the albservicesjob started. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	StartTime *TimeStamp `json:"start_time,omitempty"`

	// The status of the albservicesjob. Enum options - UNDETERMINED, PENDING, IN_PROGRESS, COMPLETED, FAILED, NOT_ENABLED. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Status *string `json:"status,omitempty"`

	// Time at which the status of albservicesjob updated. Field introduced in 22.1.6. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	StatusUpdateTime *TimeStamp `json:"status_update_time,omitempty"`

	// The unique identifier of the tenant to which this albservicesjob belongs. It is a reference to an object of type Tenant. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// Job token. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Token *string `json:"token,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// A unique identifier for this albservicesjob entry. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	UUID *string `json:"uuid,omitempty"`
}
