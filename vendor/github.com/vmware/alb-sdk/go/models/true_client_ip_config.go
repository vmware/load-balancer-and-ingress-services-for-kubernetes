// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// TrueClientIPConfig true client IP config
// swagger:model TrueClientIPConfig
type TrueClientIPConfig struct {

	// Denotes the end from which to count the IPs in the specified header value. Enum options - LEFT, RIGHT. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Direction *string `json:"direction,omitempty"`

	// Headers to derive client IP from. The header value needs to be a comma-separated list of IP addresses. If none specified and use_true_client_ip is set to true, it will use X-Forwarded-For header, if present. Field introduced in 21.1.3. Maximum of 1 items allowed. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Headers []string `json:"headers,omitempty"`

	// Position in the configured direction, in the specified header's value, to be used to set true client IP. If the value is greater than the number of IP addresses in the header, then the last IP address in the configured direction in the header will be used. Allowed values are 1-1000. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	IndexInHeader *uint32 `json:"index_in_header,omitempty"`
}
