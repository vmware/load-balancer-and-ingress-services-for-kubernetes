// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// CloudInfo cloud info
// swagger:model CloudInfo
type CloudInfo struct {

	// CloudConnectorAgent properties specific to this cloud type. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CcaProps *CCAgentProperties `json:"cca_props,omitempty"`

	// Controller properties specific to this cloud type. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ControllerProps *ControllerProperties `json:"controller_props,omitempty"`

	// Flavor properties specific to this cloud type. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	FlavorProps []*CloudFlavor `json:"flavor_props,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	FlavorRegexFilter *string `json:"flavor_regex_filter,omitempty"`

	// Supported hypervisors. Enum options - DEFAULT, VMWARE_ESX, KVM, VMWARE_VSAN, XEN. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Htypes []string `json:"htypes,omitempty"`

	// Cloud type. Enum options - CLOUD_NONE, CLOUD_VCENTER, CLOUD_OPENSTACK, CLOUD_AWS, CLOUD_VCA, CLOUD_APIC, CLOUD_MESOS, CLOUD_LINUXSERVER, CLOUD_DOCKER_UCP, CLOUD_RANCHER, CLOUD_OSHIFT_K8S, CLOUD_AZURE, CLOUD_GCP, CLOUD_NSXT. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Vtype *string `json:"vtype"`
}
