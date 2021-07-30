// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ContainerCloudService container cloud service
// swagger:model ContainerCloudService
type ContainerCloudService struct {

	// cc_id of ContainerCloudService.
	CcID *string `json:"cc_id,omitempty"`

	// object of ContainerCloudService.
	Object *string `json:"object,omitempty"`

	// reason of ContainerCloudService.
	Reason *string `json:"reason,omitempty"`

	// service of ContainerCloudService.
	Service *string `json:"service,omitempty"`

	// status of ContainerCloudService.
	Status *string `json:"status,omitempty"`
}
