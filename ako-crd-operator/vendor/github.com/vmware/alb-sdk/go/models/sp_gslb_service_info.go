// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SpGslbServiceInfo sp gslb service info
// swagger:model SpGslbServiceInfo
type SpGslbServiceInfo struct {

	// FQDNs associated with the GSLB service. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Fqdns []string `json:"fqdns,omitempty"`

	// GSLB service uuid associated with the site persistence pool. It is a reference to an object of type GslbService. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	GsRef *string `json:"gs_ref,omitempty"`
}
