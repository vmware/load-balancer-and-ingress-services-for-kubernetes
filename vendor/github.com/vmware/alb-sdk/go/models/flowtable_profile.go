// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// FlowtableProfile flowtable profile
// swagger:model FlowtableProfile
type FlowtableProfile struct {

	// Idle timeout in seconds for ICMP flows. Allowed values are 1-36000. Field introduced in 20.1.3. Unit is SECONDS. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	IcmpIDLETimeout *int32 `json:"icmp_idle_timeout,omitempty"`

	// Idle timeout in seconds for TCP flows in closed state. Allowed values are 1-36000. Field introduced in 18.2.5. Unit is SECONDS. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	TCPClosedTimeout *int32 `json:"tcp_closed_timeout,omitempty"`

	// Idle timeout in seconds for nat TCP flows in connection setup state. Allowed values are 1-36000. Field introduced in 18.2.5. Unit is SECONDS. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	TCPConnectionSetupTimeout *int32 `json:"tcp_connection_setup_timeout,omitempty"`

	// Idle timeout in seconds for TCP flows in half closed state. Allowed values are 1-36000. Field introduced in 18.2.5. Unit is SECONDS. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	TCPHalfClosedTimeout *int32 `json:"tcp_half_closed_timeout,omitempty"`

	// Idle timeout in seconds for TCP flows. Allowed values are 1-36000. Field introduced in 18.2.5. Unit is SECONDS. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	TCPIDLETimeout *int32 `json:"tcp_idle_timeout,omitempty"`

	// Timeout in seconds for TCP flows after RST is seen.Within this timeout, if any non-syn packet is seenfrom the endpoint from which RST is received,nat-flow moves to established state. Otherwise nat-flowis cleaned up. This state helps to mitigate the impactof RST attacks. Allowed values are 1-36000. Field introduced in 18.2.5. Unit is SECONDS. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	TCPResetTimeout *int32 `json:"tcp_reset_timeout,omitempty"`

	// Idle timeout in seconds for UDP flows. Allowed values are 1-36000. Field introduced in 18.2.5. Unit is SECONDS. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UDPIDLETimeout *int32 `json:"udp_idle_timeout,omitempty"`
}
