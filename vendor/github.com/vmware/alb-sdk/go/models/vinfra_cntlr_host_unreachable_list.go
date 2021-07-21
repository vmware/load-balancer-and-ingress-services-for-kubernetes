// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// VinfraCntlrHostUnreachableList vinfra cntlr host unreachable list
// swagger:model VinfraCntlrHostUnreachableList
type VinfraCntlrHostUnreachableList struct {

	// host_name of VinfraCntlrHostUnreachableList.
	HostName []string `json:"host_name,omitempty"`

	// vcenter of VinfraCntlrHostUnreachableList.
	// Required: true
	Vcenter *string `json:"vcenter"`
}
