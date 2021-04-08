package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// MarathonConfiguration marathon configuration
// swagger:model MarathonConfiguration
type MarathonConfiguration struct {

	// Framework tag to be used in Virtualservice name. Default is framework name from Mesos. If this tag is altered atruntime, Virtualservices will be deleted and re-created.
	FrameworkTag *string `json:"framework_tag,omitempty"`

	// Password for Marathon authentication.
	MarathonPassword *string `json:"marathon_password,omitempty"`

	// Marathon API URL of the form http //host port.
	MarathonURL *string `json:"marathon_url,omitempty"`

	// Username for Marathon authentication.
	MarathonUsername *string `json:"marathon_username,omitempty"`

	// Private port range allocated to this Marathon framework instance.
	PrivatePortRange *PortRange `json:"private_port_range,omitempty"`

	// Public port range allocated to this Marathon framework instance.
	PublicPortRange *PortRange `json:"public_port_range,omitempty"`

	// Tenant to pin this Marathon instance to. If set, a tenant object will be created in Avi bearing this name and all applications created in this marathon will be associated with this tenant regardless of, if any, tenant configuration in marathon label for this application.
	Tenant *string `json:"tenant,omitempty"`

	// Use Token based authentication instead of basic authentication. Token is refreshed every 5 minutes.
	UseTokenAuth *bool `json:"use_token_auth,omitempty"`

	// Tag VS name with framework name or framework_tag. Useful in deployments with multiple frameworks.
	VsNameTagFramework *bool `json:"vs_name_tag_framework,omitempty"`
}
