package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// OshiftSharedVirtualService oshift shared virtual service
// swagger:model OshiftSharedVirtualService
type OshiftSharedVirtualService struct {

	// Name of shared virtualservice. VirtualService will be created automatically by Cloud Connector. Field introduced in 17.1.1.
	// Required: true
	VirtualserviceName *string `json:"virtualservice_name"`
}
