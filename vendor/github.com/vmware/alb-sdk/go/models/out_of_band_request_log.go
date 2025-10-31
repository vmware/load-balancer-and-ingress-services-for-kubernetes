// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// OutOfBandRequestLog out of band request log
// swagger:model OutOfBandRequestLog
type OutOfBandRequestLog struct {

	// Logs for out-of-band requests sent from the DataScript. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	DsReqLogs []*DSRequestLog `json:"ds_req_logs,omitempty"`
}
