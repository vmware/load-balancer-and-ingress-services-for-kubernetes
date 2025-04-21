// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// CaptureTCP capture TCP
// swagger:model CaptureTCP
type CaptureTCP struct {

	// TCP flags filter. Or'ed internally and And'ed amongst each other. . Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Tcpflag *CaptureTCPFlags `json:"tcpflag,omitempty"`
}
