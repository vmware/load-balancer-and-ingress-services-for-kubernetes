// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// GCPVIPAllocation g c p v IP allocation
// swagger:model GCPVIPAllocation
type GCPVIPAllocation struct {

	// Configure Google Cloud Internal LoadBalancer for VIP. The VIP will be auto allocated from a Google Cloud VPC Subnet. Field introduced in 18.2.9, 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Ilb *GCPVIPILB `json:"ilb,omitempty"`

	// VIP Allocation Mode. Enum options - ROUTES, ILB. Field introduced in 18.2.9, 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Mode *string `json:"mode"`

	// Configure Google Cloud VPC Routes for VIP. The VIP can either be a static IP or auto allocted from AVI Internal Network. The VIP should not overlap with any of the subnet ranges in Google Cloud VPC. Field introduced in 18.2.9, 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Routes *GCPVIPRoutes `json:"routes,omitempty"`
}
