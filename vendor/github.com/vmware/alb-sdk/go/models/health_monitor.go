// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// HealthMonitor health monitor
// swagger:model HealthMonitor
type HealthMonitor struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// By default, multiple instances of the same healthmonitor to the same server are suppressed intelligently. In rare cases, the monitor may have specific constructs that go beyond the server keys (ip, port, etc.) during which such suppression is not desired. Use this knob to allow duplicates. Field introduced in 18.2.8. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- true), Basic edition(Allowed values- true), Enterprise with Cloud Services edition.
	AllowDuplicateMonitors *bool `json:"allow_duplicate_monitors,omitempty"`

	// Authentication information for username/password. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Authentication *HealthMonitorAuthInfo `json:"authentication,omitempty"`

	// Protobuf versioning for config pbs. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	ConfigpbAttributes *ConfigPbAttributes `json:"configpb_attributes,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Description *string `json:"description,omitempty"`

	// During addition of a server or healthmonitors or during bootup, Avi performs sequential health checks rather than waiting for send-interval to kick in, to mark the server up as soon as possible. This knob may be used to turn this feature off. Field introduced in 18.2.7. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- false), Basic edition(Allowed values- false), Enterprise with Cloud Services edition.
	DisableQuickstart *bool `json:"disable_quickstart,omitempty"`

	//  Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	DNSMonitor *HealthMonitorDNS `json:"dns_monitor,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ExternalMonitor *HealthMonitorExternal `json:"external_monitor,omitempty"`

	// Number of continuous failed health checks before the server is marked down. Allowed values are 1-50. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	FailedChecks *int32 `json:"failed_checks,omitempty"`

	// Health monitor for FTP. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	FtpMonitor *HealthMonitorFtp `json:"ftp_monitor,omitempty"`

	// Health monitor for FTPS. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	FtpsMonitor *HealthMonitorFtp `json:"ftps_monitor,omitempty"`

	//  Allowed in Enterprise edition with any value, Basic, Enterprise with Cloud Services edition.
	HTTPMonitor *HealthMonitorHTTP `json:"http_monitor,omitempty"`

	//  Allowed in Enterprise edition with any value, Basic, Enterprise with Cloud Services edition.
	HTTPSMonitor *HealthMonitorHTTP `json:"https_monitor,omitempty"`

	// Health monitor for IMAP. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ImapMonitor *HealthMonitorImap `json:"imap_monitor,omitempty"`

	// Health monitor for IMAPS. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ImapsMonitor *HealthMonitorImap `json:"imaps_monitor,omitempty"`

	// This field describes the object's replication scope. If the field is set to false, then the object is visible within the controller-cluster and its associated service-engines.  If the field is set to true, then the object is replicated across the federation.  . Field introduced in 17.1.3. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- false), Basic edition(Allowed values- false), Enterprise with Cloud Services edition.
	IsFederated *bool `json:"is_federated,omitempty"`

	// Health monitor for LDAP. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	LdapMonitor *HealthMonitorLdap `json:"ldap_monitor,omitempty"`

	// Health monitor for LDAPS. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	LdapsMonitor *HealthMonitorLdap `json:"ldaps_monitor,omitempty"`

	// List of labels to be used for granular RBAC. Field introduced in 20.1.5. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	Markers []*RoleFilterMatchLabel `json:"markers,omitempty"`

	// Use this port instead of the port defined for the server in the Pool. If the monitor succeeds to this port, the load balanced traffic will still be sent to the port of the server defined within the Pool. Allowed values are 1-65535. Special values are 0 - Use server port. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MonitorPort *int32 `json:"monitor_port,omitempty"`

	// A user friendly name for this health monitor. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Name *string `json:"name"`

	// Health monitor for POP3. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Pop3Monitor *HealthMonitorPop3 `json:"pop3_monitor,omitempty"`

	// Health monitor for POP3S. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Pop3sMonitor *HealthMonitorPop3 `json:"pop3s_monitor,omitempty"`

	// Health monitor for Radius. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	RadiusMonitor *HealthMonitorRadius `json:"radius_monitor,omitempty"`

	// A valid response from the server is expected within the receive timeout window.  This timeout must be less than the send interval.  If server status is regularly flapping up and down, consider increasing this value. Allowed values are 1-2400. Unit is SEC. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ReceiveTimeout *int32 `json:"receive_timeout,omitempty"`

	// Health monitor for SCTP. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SctpMonitor *HealthMonitorSctp `json:"sctp_monitor,omitempty"`

	// Frequency, in seconds, that monitors are sent to a server. Allowed values are 1-3600. Unit is SEC. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SendInterval *int32 `json:"send_interval,omitempty"`

	// Health monitor for SIP. Field introduced in 17.2.8, 18.1.3, 18.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SipMonitor *HealthMonitorSIP `json:"sip_monitor,omitempty"`

	// Health monitor for SMTP. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SMTPMonitor *HealthMonitorSMTP `json:"smtp_monitor,omitempty"`

	// Health monitor for SMTPS. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SmtpsMonitor *HealthMonitorSMTP `json:"smtps_monitor,omitempty"`

	// Number of continuous successful health checks before server is marked up. Allowed values are 1-50. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SuccessfulChecks *int32 `json:"successful_checks,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	TCPMonitor *HealthMonitorTCP `json:"tcp_monitor,omitempty"`

	//  It is a reference to an object of type Tenant. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// Type of the health monitor. Enum options - HEALTH_MONITOR_PING, HEALTH_MONITOR_TCP, HEALTH_MONITOR_HTTP, HEALTH_MONITOR_HTTPS, HEALTH_MONITOR_EXTERNAL, HEALTH_MONITOR_UDP, HEALTH_MONITOR_DNS, HEALTH_MONITOR_GSLB, HEALTH_MONITOR_SIP, HEALTH_MONITOR_RADIUS, HEALTH_MONITOR_SMTP, HEALTH_MONITOR_SMTPS, HEALTH_MONITOR_POP3, HEALTH_MONITOR_POP3S, HEALTH_MONITOR_IMAP, HEALTH_MONITOR_IMAPS, HEALTH_MONITOR_FTP, HEALTH_MONITOR_FTPS, HEALTH_MONITOR_LDAP, HEALTH_MONITOR_LDAPS.... Allowed in Enterprise edition with any value, Essentials edition(Allowed values- HEALTH_MONITOR_PING,HEALTH_MONITOR_TCP,HEALTH_MONITOR_UDP), Basic edition(Allowed values- HEALTH_MONITOR_PING,HEALTH_MONITOR_TCP,HEALTH_MONITOR_UDP,HEALTH_MONITOR_HTTP,HEALTH_MONITOR_HTTPS), Enterprise with Cloud Services edition.
	// Required: true
	Type *string `json:"type"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UDPMonitor *HealthMonitorUDP `json:"udp_monitor,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// UUID of the health monitor. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UUID *string `json:"uuid,omitempty"`
}
