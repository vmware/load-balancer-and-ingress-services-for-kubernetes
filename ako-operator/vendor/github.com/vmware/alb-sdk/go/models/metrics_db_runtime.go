// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// MetricsDbRuntime metrics db runtime
// swagger:model MetricsDbRuntime
type MetricsDbRuntime struct {

	// Db Client name. Can be of DB_CLIENT_RT/DB_CLIENT_BATCH/DB_CLIENT_RT_ARR. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	DbClientName *string `json:"db_client_name,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DbNumClientQueries *uint32 `json:"db_num_client_queries,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DbNumClientResp *uint32 `json:"db_num_client_resp,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DbNumDbQueries *uint32 `json:"db_num_db_queries,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DbNumDbResp *uint32 `json:"db_num_db_resp,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DbNumOom *uint32 `json:"db_num_oom,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DbQueueSize *uint32 `json:"db_queue_size,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DbRumQueries *uint32 `json:"db_rum_queries,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DbRumRows *uint32 `json:"db_rum_rows,omitempty"`
}
