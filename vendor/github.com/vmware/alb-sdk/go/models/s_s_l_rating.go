package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SSLRating s s l rating
// swagger:model SSLRating
type SSLRating struct {

	//  Enum options - SSL_SCORE_NOT_SECURE, SSL_SCORE_VERY_BAD, SSL_SCORE_BAD, SSL_SCORE_AVERAGE, SSL_SCORE_GOOD, SSL_SCORE_EXCELLENT.
	CompatibilityRating *string `json:"compatibility_rating,omitempty"`

	//  Enum options - SSL_SCORE_NOT_SECURE, SSL_SCORE_VERY_BAD, SSL_SCORE_BAD, SSL_SCORE_AVERAGE, SSL_SCORE_GOOD, SSL_SCORE_EXCELLENT.
	PerformanceRating *string `json:"performance_rating,omitempty"`

	// security_score of SSLRating.
	SecurityScore *string `json:"security_score,omitempty"`
}
