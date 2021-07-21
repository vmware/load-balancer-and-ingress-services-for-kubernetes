// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// VsResyncParams vs resync params
// swagger:model VsResyncParams
type VsResyncParams struct {

	//  It is a reference to an object of type ServiceEngine.
	SeRef []string `json:"se_ref,omitempty"`

	// Unique object identifier of the object.
	UUID *string `json:"uuid,omitempty"`
}
