package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// OutOfBandRequestLog out of band request log
// swagger:model OutOfBandRequestLog
type OutOfBandRequestLog struct {

	// Logs for out-of-band requests sent from the DataScript. Field introduced in 20.1.3.
	DsReqLogs []*DSRequestLog `json:"ds_req_logs,omitempty"`
}
