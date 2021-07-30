// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// MetricsDbRuntime metrics db runtime
// swagger:model MetricsDbRuntime
type MetricsDbRuntime struct {

	// Number of db_num_client_queries.
	DbNumClientQueries *int32 `json:"db_num_client_queries,omitempty"`

	// Number of db_num_client_resp.
	DbNumClientResp *int32 `json:"db_num_client_resp,omitempty"`

	// Number of db_num_db_queries.
	DbNumDbQueries *int32 `json:"db_num_db_queries,omitempty"`

	// Number of db_num_db_resp.
	DbNumDbResp *int32 `json:"db_num_db_resp,omitempty"`

	// Number of db_num_oom.
	DbNumOom *int32 `json:"db_num_oom,omitempty"`

	// Number of db_queue_size.
	DbQueueSize *int32 `json:"db_queue_size,omitempty"`

	// Number of db_rum_queries.
	DbRumQueries *int32 `json:"db_rum_queries,omitempty"`

	// Number of db_rum_rows.
	DbRumRows *int32 `json:"db_rum_rows,omitempty"`
}
