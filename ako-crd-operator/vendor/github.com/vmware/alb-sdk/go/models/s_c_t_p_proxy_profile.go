// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SCTPProxyProfile s c t p proxy profile
// swagger:model SCTPProxyProfile
type SCTPProxyProfile struct {

	// SCTP cookie expiration timeout. Allowed values are 60-3600. Field introduced in 22.1.3. Unit is SEC. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	CookieExpirationTimeout *int32 `json:"cookie_expiration_timeout,omitempty"`

	// SCTP heartbeat interval. Allowed values are 30-247483647. Field introduced in 22.1.3. Unit is SEC. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	HeartbeatInterval *int32 `json:"heartbeat_interval,omitempty"`

	// SCTP autoclose timeout. 0 means autoclose deactivated. Allowed values are 0-247483647. Field introduced in 22.1.3. Unit is SEC. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	IDLETimeout *int32 `json:"idle_timeout,omitempty"`

	// SCTP maximum retransmissions for association. Allowed values are 1-247483647. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	MaxRetransmissionsAssociation *int32 `json:"max_retransmissions_association,omitempty"`

	// SCTP maximum retransmissions for INIT chunks. Allowed values are 1-247483647. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	MaxRetransmissionsInitChunks *int32 `json:"max_retransmissions_init_chunks,omitempty"`

	// Number of incoming SCTP Streams. Allowed values are 1-65535. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	NumberOfStreams *int32 `json:"number_of_streams,omitempty"`

	// SCTP send and receive buffer size. Allowed values are 2-65536. Field introduced in 22.1.3. Unit is KB. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ReceiveWindow *int32 `json:"receive_window,omitempty"`

	// SCTP reset timeout. 0 means 5 times RTO max. Allowed values are 0-247483647. Field introduced in 22.1.3. Unit is SEC. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ResetTimeout *int32 `json:"reset_timeout,omitempty"`
}
