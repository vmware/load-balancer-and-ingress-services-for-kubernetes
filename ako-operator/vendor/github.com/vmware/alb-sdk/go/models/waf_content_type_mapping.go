// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// WafContentTypeMapping waf content type mapping
// swagger:model WafContentTypeMapping
type WafContentTypeMapping struct {

	// Request Content-Type. When it is equal to request Content-Type header value, the specified request_body_parser is used. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	ContentType *string `json:"content_type"`

	// String operation to use for matching the content_type. Only EQUALS and REGEX_MATCH are supported *string operations here. All EQUALS matches are checked before REGEX_MATCH matches. Enum options - BEGINS_WITH, DOES_NOT_BEGIN_WITH, CONTAINS, DOES_NOT_CONTAIN, ENDS_WITH, DOES_NOT_END_WITH, EQUALS, DOES_NOT_EQUAL, REGEX_MATCH, REGEX_DOES_NOT_MATCH. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	MatchOp *string `json:"match_op,omitempty"`

	// Request body parser. Enum options - WAF_REQUEST_PARSER_URLENCODED, WAF_REQUEST_PARSER_MULTIPART, WAF_REQUEST_PARSER_JSON, WAF_REQUEST_PARSER_XML, WAF_REQUEST_PARSER_HANDLE_AS_STRING, WAF_REQUEST_PARSER_DO_NOT_PARSE. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	RequestBodyParser *string `json:"request_body_parser"`
}
