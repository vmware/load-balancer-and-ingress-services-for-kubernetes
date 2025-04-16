// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// HSMSafenetLunaServer h s m safenet luna server
// swagger:model HSMSafenetLunaServer
type HSMSafenetLunaServer struct {

	//  Field introduced in 16.5.2,17.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Index *uint32 `json:"index"`

	// Password of the partition assigned to this client. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PartitionPasswd *string `json:"partition_passwd,omitempty"`

	// Serial number of the partition assigned to this client. Field introduced in 16.5.2,17.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PartitionSerialNumber *string `json:"partition_serial_number,omitempty"`

	// IP address of the Thales Luna HSM device. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	RemoteIP *string `json:"remote_ip"`

	// CA certificate of the server. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	ServerCert *string `json:"server_cert"`
}
