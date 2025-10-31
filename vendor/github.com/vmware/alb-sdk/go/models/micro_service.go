// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// MicroService micro service
// swagger:model MicroService
type MicroService struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	ApplicationName *string `json:"application_name"`

	// Checksum of cloud configuration for Microservice. Internally set by cloud connector. Field introduced in 17.2.8. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CloudConfigCksum *string `json:"cloud_config_cksum,omitempty"`

	// Protobuf versioning for config pbs. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	ConfigpbAttributes *ConfigPbAttributes `json:"configpb_attributes,omitempty"`

	// The list of containers for this microservice. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Containers []*MicroServiceContainer `json:"containers,omitempty"`

	// Creator name. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CreatedBy *string `json:"created_by,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Description *string `json:"description,omitempty"`

	// Flag to indicate if container IP list is provided by cloud connectorThis is applicable for overlay cases. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	IPList *bool `json:"ip_list,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Name *string `json:"name"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	OrchestratorName *string `json:"orchestrator_name"`

	//  It is a reference to an object of type Tenant. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UUID *string `json:"uuid,omitempty"`
}
