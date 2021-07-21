// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// CloudVipUpdate cloud vip update
// swagger:model CloudVipUpdate
type CloudVipUpdate struct {

	// cc_id of CloudVipUpdate.
	CcID *string `json:"cc_id,omitempty"`

	// error_string of CloudVipUpdate.
	ErrorString *string `json:"error_string,omitempty"`

	// Unique object identifier of vs.
	VsUUID *string `json:"vs_uuid,omitempty"`

	//  Enum options - CLOUD_NONE, CLOUD_VCENTER, CLOUD_OPENSTACK, CLOUD_AWS, CLOUD_VCA, CLOUD_APIC, CLOUD_MESOS, CLOUD_LINUXSERVER, CLOUD_DOCKER_UCP, CLOUD_RANCHER, CLOUD_OSHIFT_K8S, CLOUD_AZURE, CLOUD_GCP, CLOUD_NSXT.
	Vtype *string `json:"vtype,omitempty"`
}
