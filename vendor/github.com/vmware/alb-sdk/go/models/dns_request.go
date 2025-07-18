// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// DNSRequest Dns request
// swagger:model DnsRequest
type DNSRequest struct {

	// Number of additional records. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AdditionalRecordsCount *uint32 `json:"additional_records_count,omitempty"`

	// Number of answer records in the client DNS request. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AnswerRecordsCount *uint32 `json:"answer_records_count,omitempty"`

	// Flag indicating client understands AD bit and is interested in the value of AD bit in the response. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AuthenticData *bool `json:"authentic_data,omitempty"`

	// Flag indicating client does not want DNSSEC validation of the response. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CheckingDisabled *bool `json:"checking_disabled,omitempty"`

	// Geo Location of Client. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ClientLocation *GeoLocation `json:"client_location,omitempty"`

	// ID of the DNS request. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Identifier *uint32 `json:"identifier,omitempty"`

	// Number of nameserver records in the client DNS request. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NameserverRecordsCount *uint32 `json:"nameserver_records_count,omitempty"`

	// DNS request operation code e.g. QUERY, NOTIFY, etc. Enum options - DNS_OPCODE_QUERY, DNS_OPCODE_IQUERY, DNS_OPCODE_STATUS, DNS_OPCODE_NOTIFY, DNS_OPCODE_UPDATE. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Opcode *string `json:"opcode,omitempty"`

	// Opt resource records in the request. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	OptRecord *DNSOptRecord `json:"opt_record,omitempty"`

	// Flag indicating request is a client query (false) or server response (true). Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	QueryOrResponse *bool `json:"query_or_response,omitempty"`

	// Number of questions in the client DNS request. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	QuestionCount *uint32 `json:"question_count,omitempty"`

	// Flag indicating client request for recursive resolution. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	RecursionDesired *bool `json:"recursion_desired,omitempty"`
}
