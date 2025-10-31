// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ACSubjectInfo a c subject info
// swagger:model ACSubjectInfo
type ACSubjectInfo struct {

	// Subject type for the audit event (e.g. DNS Hostname, IP address, URI). Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	Type *string `json:"type"`

	// Subject value for the audit event (e.g. www.example.com, 10.10.10.10, www.foo.com/index.html). Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	Value *string `json:"value"`
}
