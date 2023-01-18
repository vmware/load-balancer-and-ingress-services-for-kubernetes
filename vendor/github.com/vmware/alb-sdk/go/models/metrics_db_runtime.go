// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// MetricsDbRuntime metrics db runtime
// swagger:model MetricsDbRuntime
type MetricsDbRuntime struct {

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DbNumClientQueries *int32 `json:"db_num_client_queries,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DbNumClientResp *int32 `json:"db_num_client_resp,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DbNumDbQueries *int32 `json:"db_num_db_queries,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DbNumDbResp *int32 `json:"db_num_db_resp,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DbNumOom *int32 `json:"db_num_oom,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DbQueueSize *int32 `json:"db_queue_size,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DbRumQueries *int32 `json:"db_rum_queries,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DbRumRows *int32 `json:"db_rum_rows,omitempty"`
}
