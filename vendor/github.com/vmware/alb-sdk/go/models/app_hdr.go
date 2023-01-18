// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// AppHdr app hdr
// swagger:model AppHdr
type AppHdr struct {

	//  Enum options - SENSITIVE, INSENSITIVE. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	HdrMatchCase *string `json:"hdr_match_case"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	HdrName *string `json:"hdr_name"`

	//  Enum options - BEGINS_WITH, DOES_NOT_BEGIN_WITH, CONTAINS, DOES_NOT_CONTAIN, ENDS_WITH, DOES_NOT_END_WITH, EQUALS, DOES_NOT_EQUAL, REGEX_MATCH, REGEX_DOES_NOT_MATCH. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- BEGINS_WITH,DOES_NOT_BEGIN_WITH,CONTAINS,DOES_NOT_CONTAIN,ENDS_WITH,DOES_NOT_END_WITH,EQUALS,DOES_NOT_EQUAL), Basic edition(Allowed values- BEGINS_WITH,DOES_NOT_BEGIN_WITH,CONTAINS,DOES_NOT_CONTAIN,ENDS_WITH,DOES_NOT_END_WITH,EQUALS,DOES_NOT_EQUAL), Enterprise with Cloud Services edition.
	// Required: true
	HdrStringOp *string `json:"hdr_string_op"`
}
