// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// DNSTxtRdata Dns txt rdata
// swagger:model DnsTxtRdata
type DNSTxtRdata struct {

	// Text data associated with the FQDN. Field introduced in 18.2.9, 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	TextStr *string `json:"text_str"`
}
