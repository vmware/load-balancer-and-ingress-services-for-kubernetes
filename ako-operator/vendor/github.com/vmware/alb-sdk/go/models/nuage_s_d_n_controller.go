// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// NuageSDNController nuage s d n controller
// swagger:model NuageSDNController
type NuageSDNController struct {

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NuageOrganization *string `json:"nuage_organization,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NuagePassword *string `json:"nuage_password,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NuagePort *uint32 `json:"nuage_port,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NuageUsername *string `json:"nuage_username,omitempty"`

	// Nuage VSD host name or IP address. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NuageVsdHost *string `json:"nuage_vsd_host,omitempty"`

	// Domain to be used for SE creation. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeDomain *string `json:"se_domain,omitempty"`

	// Enterprise to be used for SE creation. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeEnterprise *string `json:"se_enterprise,omitempty"`

	// Network to be used for SE creation. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeNetwork *string `json:"se_network,omitempty"`

	// Policy Group to be used for SE creation. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SePolicyGroup *string `json:"se_policy_group,omitempty"`

	// User to be used for SE creation. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeUser *string `json:"se_user,omitempty"`

	// Zone to be used for SE creation. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeZone *string `json:"se_zone,omitempty"`
}
