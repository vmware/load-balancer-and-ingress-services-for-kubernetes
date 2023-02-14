// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SubJob sub job
// swagger:model SubJob
type SubJob struct {

	//  Field introduced in 18.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	ExpiresAt *string `json:"expires_at"`

	//  Field introduced in 18.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Metadata *string `json:"metadata,omitempty"`

	// Number of times the sub job got scheduled. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumRetries *int32 `json:"num_retries,omitempty"`

	//  Enum options - JOB_TYPE_VS_FULL_LOGS, JOB_TYPE_VS_UDF, JOB_TYPE_VS_METRICS_RT, JOB_TYPE_SSL_CERT, JOB_TYPE_DEBUGVS_PKT_CAPTURE, JOB_TYPE_CONSISTENCY_CHECK, JOB_TYPE_TECHSUPPORT, JOB_TYPE_PKI_PROFILE, JOB_TYPE_NSP_RULE, JOB_TYPE_SEGROUP_METRICS_RT, JOB_TYPE_POSTGRES_STATUS, JOB_TYPE_VS_ROTATE_KEYS, JOB_TYPE_POOL_DNS, JOB_TYPE_GSLB_SERVICE, JOB_TYPE_APP_PERSISTENCE, JOB_TYPE_PROCESS_LOCKED_USER_ACCOUNTS, JOB_TYPE_SESSION, JOB_TYPE_AUTHTOKEN, JOB_TYPE_CLUSTER, JOB_TYPE_SE_SECURE_CHANNEL_CLEANUP, JOB_TYPE_OCSP_STAPLE_STATUS, JOB_TYPE_FILE_OBJECT_CLEANUP, JOB_TYPE_WAF_POLICY_REFRESH_APPLICATION_SIGNATURES, JOB_TYPE_POOL_ASYNC, JOB_TYPE_PROCESS_BASELINE_BENCHMARK, JOB_TYPE_GEODB_REFRESH_CONTROLLER_DATABASES, JOB_TYPE_POSTGRES_VACUUM, JOB_TYPE_WAF_POLICY_AUTO_UPDATE_CRS. Field introduced in 18.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Type *string `json:"type"`
}
