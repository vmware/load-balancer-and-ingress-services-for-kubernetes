// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// DsrProfile dsr profile
// swagger:model DsrProfile
type DsrProfile struct {

	// Encapsulation type to use when DSR is L3. Enum options - ENCAP_IPINIP, ENCAP_GRE. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	DsrEncapType *string `json:"dsr_encap_type"`

	// DSR type L2/L3. Enum options - DSR_TYPE_L2, DSR_TYPE_L3. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	DsrType *string `json:"dsr_type"`
}
