// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// CloudProperties cloud properties
// swagger:model CloudProperties
type CloudProperties struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// CloudConnector properties. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CcProps *CCProperties `json:"cc_props,omitempty"`

	// Cloud types supported by CloudConnector. Enum options - CLOUD_NONE, CLOUD_VCENTER, CLOUD_OPENSTACK, CLOUD_AWS, CLOUD_VCA, CLOUD_APIC, CLOUD_MESOS, CLOUD_LINUXSERVER, CLOUD_DOCKER_UCP, CLOUD_RANCHER, CLOUD_OSHIFT_K8S, CLOUD_AZURE, CLOUD_GCP, CLOUD_NSXT. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CcVtypes []string `json:"cc_vtypes,omitempty"`

	// Protobuf versioning for config pbs. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	ConfigpbAttributes *ConfigPbAttributes `json:"configpb_attributes,omitempty"`

	// Hypervisor properties. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	HypProps []*HypervisorProperties `json:"hyp_props,omitempty"`

	// Properties specific to a cloud type. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Info []*CloudInfo `json:"info,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UUID *string `json:"uuid,omitempty"`
}
