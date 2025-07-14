// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ClientLogStreamingConfig client log streaming config
// swagger:model ClientLogStreamingConfig
type ClientLogStreamingConfig struct {

	// IP address or hostnames (FQDNs) of destination servers. If an FQDN is provided, this should be resolvable on Avi Service Engines. Multiple servers are supported by furnishing a comma-separated list of IP addresses or host names, for example, 11.11.11.11,23.12.12.4,2001 123  1. Optionally, a separate port can be specified for each external server in the list, for example,11.11.11.11 234,12.12.12.12 343,[2001 123  1] 234. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	ExternalServer *string `json:"external_server"`

	// The service port to use for the external servers. If multiple external servers have been specified, the single port number specified here will apply to all those servers for which an explicit port number has not been specified in the external server list. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ExternalServerPort *uint32 `json:"external_server_port,omitempty"`

	// Configuration to specify the format of streamed logs. By default, each log is encoded in JSON format. Field introduced in 18.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	FormatConfig *ClientLogStreamingFormat `json:"format_config,omitempty"`

	// Type of logs to stream to the external server. Default is LOGS_ALL, i.e., send all logs. Enum options - LOGS_SIGNIFICANT_ONLY, LOGS_UDF_ONLY, LOGS_UDF_SIGNIFICANT, LOGS_ALL. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	LogTypesToSend *string `json:"log_types_to_send,omitempty"`

	// One or more keys which should exist in VirtualService RBAC markers. Key along with values will be streamed out in log. If key is not found in RBAC markers, it will not be streamed. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	MarkerKeys []*RoleFilterMatchLabel `json:"marker_keys,omitempty"`

	// Maximum number of logs per second streamed to the remote server. By default, 100 logs per second are streamed. Set this to zero(0) to not enforce any limit. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MaxLogsPerSecond *uint32 `json:"max_logs_per_second,omitempty"`

	// Protocol to use for streaming logs. Enum options - LOG_STREAMING_PROTOCOL_UDP, LOG_STREAMING_PROTOCOL_SYSLOG_OVER_UDP, LOG_STREAMING_PROTOCOL_TCP, LOG_STREAMING_PROTOCOL_SYSLOG_OVER_TCP, LOG_STREAMING_PROTOCOL_RAW_OVER_UDP, LOG_STREAMING_PROTOCOL_TLS, LOG_STREAMING_PROTOCOL_SYSLOG_OVER_TLS. Field introduced in 18.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Protocol *string `json:"protocol,omitempty"`

	// Syslog configuration if a Syslog-based protocol is specified for streaming. Field introduced in 18.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SyslogConfig *StreamingSyslogConfig `json:"syslog_config,omitempty"`
}
