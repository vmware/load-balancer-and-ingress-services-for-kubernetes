// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// UpgradeOpsState upgrade ops state
// swagger:model UpgradeOpsState
type UpgradeOpsState struct {

	// The last time the state changed. Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	LastChangedTime *TimeStamp `json:"last_changed_time,omitempty"`

	// Descriptive reason for the state-change. Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Reason *string `json:"reason,omitempty"`

	// State for keeping track of reboot status during upgrade operation. Field introduced in 20.1.2. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	Rebooted *bool `json:"rebooted,omitempty"`

	// The upgrade operations current fsm-state. Enum options - UPGRADE_FSM_INIT, UPGRADE_FSM_STARTED, UPGRADE_FSM_WAITING, UPGRADE_FSM_IN_PROGRESS, UPGRADE_FSM_ENQUEUED, UPGRADE_FSM_ERROR, UPGRADE_FSM_SUSPENDED, UPGRADE_FSM_ENQUEUE_FAILED, UPGRADE_FSM_PAUSED, UPGRADE_FSM_COMPLETED, UPGRADE_FSM_ABORT_IN_PROGRESS, UPGRADE_FSM_ABORTED, UPGRADE_FSM_SE_UPGRADE_IN_PROGRESS, UPGRADE_FSM_CONTROLLER_COMPLETED, UPGRADE_FSM_DUMMY_3, UPGRADE_FSM_DUMMY_4, UPGRADE_FSM_DUMMY_5, UPGRADE_PRE_CHECK_STARTED, UPGRADE_PRE_CHECK_IN_PROGRESS, UPGRADE_PRE_CHECK_SUCCESS.... Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	State *string `json:"state,omitempty"`
}
