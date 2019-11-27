package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// FailAction fail action
// swagger:model FailAction
type FailAction struct {

	// Backup Pool when pool experiences a failure. Field deprecated in 18.1.2.
	BackupPool *FailActionBackupPool `json:"backup_pool,omitempty"`

	// Local response to HTTP requests when pool experiences a failure.
	LocalRsp *FailActionHTTPLocalResponse `json:"local_rsp,omitempty"`

	// URL to redirect HTTP requests to when pool experiences a failure.
	Redirect *FailActionHTTPRedirect `json:"redirect,omitempty"`

	// Enables a response to client when pool experiences a failure. By default TCP connection is closed. Enum options - FAIL_ACTION_HTTP_REDIRECT, FAIL_ACTION_HTTP_LOCAL_RSP, FAIL_ACTION_CLOSE_CONN, FAIL_ACTION_BACKUP_POOL.
	// Required: true
	Type *string `json:"type"`
}
