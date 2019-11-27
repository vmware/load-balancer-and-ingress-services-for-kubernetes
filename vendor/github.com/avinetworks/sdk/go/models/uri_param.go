package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// URIParam URI param
// swagger:model URIParam
type URIParam struct {

	// Token config either for the URI components or a constant string.
	Tokens []*URIParamToken `json:"tokens,omitempty"`

	// URI param type. Enum options - URI_PARAM_TYPE_TOKENIZED.
	// Required: true
	Type *string `json:"type"`
}
