// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// PodToleration pod toleration
// swagger:model PodToleration
type PodToleration struct {

	// Effect to match. Enum options - NO_SCHEDULE, PREFER_NO_SCHEDULE, NO_EXECUTE. Field introduced in 17.2.14, 18.1.5, 18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Effect *string `json:"effect,omitempty"`

	// Key to match. Field introduced in 17.2.14, 18.1.5, 18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Key *string `json:"key,omitempty"`

	// Operator to match. Enum options - EQUAL, EXISTS. Field introduced in 17.2.14, 18.1.5, 18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Operator *string `json:"operator,omitempty"`

	// Pods that tolerate the taint with a specified toleration_seconds remain bound for the specified amount of time. Field introduced in 17.2.14, 18.1.5, 18.2.1. Unit is SECONDS. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	TolerationSeconds uint32 `json:"toleration_seconds,omitempty"`

	// Value to match. Field introduced in 17.2.14, 18.1.5, 18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Value *string `json:"value,omitempty"`
}
