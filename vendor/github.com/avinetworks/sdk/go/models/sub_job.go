package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SubJob sub job
// swagger:model SubJob
type SubJob struct {

	//  Field introduced in 18.1.1.
	// Required: true
	ExpiresAt *string `json:"expires_at"`

	//  Field introduced in 18.1.1.
	Metadata *string `json:"metadata,omitempty"`

	//  Enum options - JOB_TYPE_VS_FULL_LOGS, JOB_TYPE_VS_UDF, JOB_TYPE_VS_METRICS_RT, JOB_TYPE_SSL_CERT, JOB_TYPE_DEBUGVS_PKT_CAPTURE, JOB_TYPE_CONSISTENCY_CHECK, JOB_TYPE_TECHSUPPORT, JOB_TYPE_PKI_PROFILE, JOB_TYPE_NSP_RULE, JOB_TYPE_SEGROUP_METRICS_RT, JOB_TYPE_POSTGRES_STATUS, JOB_TYPE_VS_ROTATE_KEYS, JOB_TYPE_POOL_DNS, JOB_TYPE_GSLB_SERVICE, JOB_TYPE_APP_PERSISTENCE, JOB_TYPE_PROCESS_LOCKED_USER_ACCOUNTS, JOB_TYPE_SESSION, JOB_TYPE_AUTHTOKEN, JOB_TYPE_CLUSTER, JOB_TYPE_SE_SECURE_CHANNEL_CLEANUP. Field introduced in 18.1.1.
	// Required: true
	Type *string `json:"type"`
}
