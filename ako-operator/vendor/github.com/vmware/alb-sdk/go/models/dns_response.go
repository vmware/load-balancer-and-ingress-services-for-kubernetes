// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// DNSResponse Dns response
// swagger:model DnsResponse
type DNSResponse struct {

	// Number of additional records. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AdditionalRecordsCount *uint32 `json:"additional_records_count,omitempty"`

	// Number of answer records. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AnswerRecordsCount *uint32 `json:"answer_records_count,omitempty"`

	// Flag to indicate the responding DNS is an authority for the requested FQDN. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AuthoritativeAnswer *bool `json:"authoritative_answer,omitempty"`

	// Flag to indicate whether fallback algorithm was used to serve this request. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	FallbackAlgorithmUsed *bool `json:"fallback_algorithm_used,omitempty"`

	// Resource records in the response are generated based on wildcard match. Field introduced in 18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	IsWildcard *bool `json:"is_wildcard,omitempty"`

	// Number of nameserver records. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NameserverRecordsCount *uint32 `json:"nameserver_records_count,omitempty"`

	// DNS response operation code e.g. QUERY, NOTIFY, etc. Enum options - DNS_OPCODE_QUERY, DNS_OPCODE_IQUERY, DNS_OPCODE_STATUS, DNS_OPCODE_NOTIFY, DNS_OPCODE_UPDATE. Field introduced in 17.1.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Opcode *string `json:"opcode,omitempty"`

	// Opt resource records in the response. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	OptRecord *DNSOptRecord `json:"opt_record,omitempty"`

	// Flag indicating response is a client query (false) or server response (true). Field introduced in 17.1.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	QueryOrResponse *bool `json:"query_or_response,omitempty"`

	// Number of questions in the client DNS request eliciting this DNS response. Field introduced in 17.1.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	QuestionCount *uint32 `json:"question_count,omitempty"`

	// Resource records in the response. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Records []*DNSResourceRecord `json:"records,omitempty"`

	// Flag indicating the DNS query is fully answered. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	RecursionAvailable *bool `json:"recursion_available,omitempty"`

	// Flag copied from the DNS query's recursion desired field by the responding DNS. Field introduced in 17.1.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	RecursionDesired *bool `json:"recursion_desired,omitempty"`

	// Response code in the response. Enum options - DNS_RCODE_NOERROR, DNS_RCODE_FORMERR, DNS_RCODE_SERVFAIL, DNS_RCODE_NXDOMAIN, DNS_RCODE_NOTIMP, DNS_RCODE_REFUSED, DNS_RCODE_YXDOMAIN, DNS_RCODE_YXRRSET, DNS_RCODE_NXRRSET, DNS_RCODE_NOTAUTH, DNS_RCODE_NOTZONE. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ResponseCode *string `json:"response_code,omitempty"`

	// Flag to indicate if the answer received from DNS is truncated (original answer exceeds 512 bytes UDP limit). Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Truncation *bool `json:"truncation,omitempty"`
}
