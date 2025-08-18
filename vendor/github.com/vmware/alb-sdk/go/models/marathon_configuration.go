// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// MarathonConfiguration marathon configuration
// swagger:model MarathonConfiguration
type MarathonConfiguration struct {

	// Framework tag to be used in Virtualservice name. Default is framework name from Mesos. If this tag is altered atruntime, Virtualservices will be deleted and re-created. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	FrameworkTag *string `json:"framework_tag,omitempty"`

	// Password for Marathon authentication. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MarathonPassword *string `json:"marathon_password,omitempty"`

	// Marathon API URL of the form http //host port. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MarathonURL *string `json:"marathon_url,omitempty"`

	// Username for Marathon authentication. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MarathonUsername *string `json:"marathon_username,omitempty"`

	// Private port range allocated to this Marathon framework instance. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PrivatePortRange *PortRange `json:"private_port_range,omitempty"`

	// Public port range allocated to this Marathon framework instance. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PublicPortRange *PortRange `json:"public_port_range,omitempty"`

	// Tenant to pin this Marathon instance to. If set, a tenant object will be created in Avi bearing this name and all applications created in this marathon will be associated with this tenant regardless of, if any, tenant configuration in marathon label for this application. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Tenant *string `json:"tenant,omitempty"`

	// Use Token based authentication instead of basic authentication. Token is refreshed every 5 minutes. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UseTokenAuth *bool `json:"use_token_auth,omitempty"`

	// Tag VS name with framework name or framework_tag. Useful in deployments with multiple frameworks. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VsNameTagFramework *bool `json:"vs_name_tag_framework,omitempty"`
}
