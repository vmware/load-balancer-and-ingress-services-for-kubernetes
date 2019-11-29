package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// DNSQueryNameMatch Dns query name match
// swagger:model DnsQueryNameMatch
type DNSQueryNameMatch struct {

	// Criterion to use for *string matching the DNS query domain name in the question section. Enum options - BEGINS_WITH, DOES_NOT_BEGIN_WITH, CONTAINS, DOES_NOT_CONTAIN, ENDS_WITH, DOES_NOT_END_WITH, EQUALS, DOES_NOT_EQUAL, REGEX_MATCH, REGEX_DOES_NOT_MATCH. Field introduced in 17.1.1.
	// Required: true
	MatchCriteria *string `json:"match_criteria"`

	// Domain name to match against that specified in the question section of the DNS query. Field introduced in 17.1.1.
	QueryDomainNames []string `json:"query_domain_names,omitempty"`

	// UUID of the *string group(s) for matching against DNS query domain name in the question section. It is a reference to an object of type StringGroup. Field introduced in 17.1.1.
	StringGroupRefs []string `json:"string_group_refs,omitempty"`
}
