package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// HSMSafenetLunaServer h s m safenet luna server
// swagger:model HSMSafenetLunaServer
type HSMSafenetLunaServer struct {

	//  Field introduced in 16.5.2,17.2.3.
	// Required: true
	Index *int32 `json:"index"`

	// Password of the partition assigned to this client.
	PartitionPasswd *string `json:"partition_passwd,omitempty"`

	// Serial number of the partition assigned to this client. Field introduced in 16.5.2,17.2.3.
	PartitionSerialNumber *string `json:"partition_serial_number,omitempty"`

	// IP address of the Safenet/Gemalto HSM device.
	// Required: true
	RemoteIP *string `json:"remote_ip"`

	// CA certificate of the server.
	// Required: true
	ServerCert *string `json:"server_cert"`
}
