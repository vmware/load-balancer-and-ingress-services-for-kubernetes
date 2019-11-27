package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// URIParamToken URI param token
// swagger:model URIParamToken
type URIParamToken struct {

	// Index of the ending token in the incoming URI. Allowed values are 0-65534. Special values are 65535 - 'end of string'.
	EndIndex *int32 `json:"end_index,omitempty"`

	// Index of the starting token in the incoming URI.
	StartIndex *int32 `json:"start_index,omitempty"`

	// Constant *string to use as a token.
	StrValue *string `json:"str_value,omitempty"`

	// Token type for constructing the URI. Enum options - URI_TOKEN_TYPE_HOST, URI_TOKEN_TYPE_PATH, URI_TOKEN_TYPE_STRING, URI_TOKEN_TYPE_STRING_GROUP, URI_TOKEN_TYPE_REGEX.
	// Required: true
	Type *string `json:"type"`
}
