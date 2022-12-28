// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// VSDataScripts v s data scripts
// swagger:model VSDataScripts
type VSDataScripts struct {

	// Index of the virtual service datascript collection. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Index *int32 `json:"index"`

	// UUID of the virtual service datascript collection. It is a reference to an object of type VSDataScriptSet. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	VsDatascriptSetRef *string `json:"vs_datascript_set_ref"`
}
