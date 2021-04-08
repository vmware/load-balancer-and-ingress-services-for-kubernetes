package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// DNSZone Dns zone
// swagger:model DnsZone
type DNSZone struct {

	// Email address of the administrator responsible for this zone. This field is used in SOA records as rname (RFC 1035). If not configured, it is inherited from the DNS service profile. Field introduced in 18.2.6.
	AdminEmail *string `json:"admin_email,omitempty"`

	// Domain name authoritatively serviced by this Virtual Service. Queries for FQDNs that are sub domains of this domain and do not have any DNS record in Avi are dropped or NXDomain response sent. For domains which are present, SOA parameters are sent in answer section of response if query type is SOA. Field introduced in 18.2.6.
	// Required: true
	DomainName *string `json:"domain_name"`

	// The primary name server for this zone. This field is used in SOA records as mname (RFC 1035). If not configured, it is inherited from the DNS service profile. If even that is not configured, the domain name is used instead. Field introduced in 18.2.6.
	NameServer *string `json:"name_server,omitempty"`
}
