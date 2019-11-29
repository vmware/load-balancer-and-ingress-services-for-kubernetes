package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// AWSASGDelete a w s a s g delete
// swagger:model AWSASGDelete
type AWSASGDelete struct {

	// List of Autoscale groups deleted from AWS. Field introduced in 17.2.10,18.1.2.
	Asgs []string `json:"asgs,omitempty"`

	// UUID of the cloud. Field introduced in 17.2.10,18.1.2.
	CcID *string `json:"cc_id,omitempty"`

	// UUID of the Pool. Field introduced in 17.2.10,18.1.2.
	PoolUUID *string `json:"pool_uuid,omitempty"`
}
