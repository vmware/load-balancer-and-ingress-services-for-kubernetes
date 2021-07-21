// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// MesosMetricsDebugFilter mesos metrics debug filter
// swagger:model MesosMetricsDebugFilter
type MesosMetricsDebugFilter struct {

	// mesos_master of MesosMetricsDebugFilter.
	MesosMaster *string `json:"mesos_master,omitempty"`

	// mesos_slave of MesosMetricsDebugFilter.
	MesosSLAVE *string `json:"mesos_slave,omitempty"`

	//  Enum options - VSERVER_METRICS_ENTITY, VM_METRICS_ENTITY, SE_METRICS_ENTITY, CONTROLLER_METRICS_ENTITY, APPLICATION_METRICS_ENTITY, TENANT_METRICS_ENTITY, POOL_METRICS_ENTITY.
	MetricEntity *string `json:"metric_entity,omitempty"`

	// Number of metrics_collection_frq.
	MetricsCollectionFrq *int32 `json:"metrics_collection_frq,omitempty"`
}
