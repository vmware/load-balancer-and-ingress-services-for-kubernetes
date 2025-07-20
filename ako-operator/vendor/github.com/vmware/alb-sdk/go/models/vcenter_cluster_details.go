// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// VcenterClusterDetails vcenter cluster details
// swagger:model VcenterClusterDetails
type VcenterClusterDetails struct {

	// Cloud Id. Field introduced in 20.1.7, 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	CcID *string `json:"cc_id,omitempty"`

	// Cluster name in vCenter. Field introduced in 20.1.7, 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Cluster *string `json:"cluster,omitempty"`

	// Error message. Field introduced in 20.1.7, 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ErrorString *string `json:"error_string,omitempty"`

	// Hosts in vCenter Cluster. Field introduced in 20.1.7, 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Hosts []string `json:"hosts,omitempty"`

	// VC url. Field introduced in 20.1.7, 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	VcURL *string `json:"vc_url,omitempty"`
}
