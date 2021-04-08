package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// UpgradeOpsState upgrade ops state
// swagger:model UpgradeOpsState
type UpgradeOpsState struct {

	// The last time the state changed. Field introduced in 18.2.6.
	LastChangedTime *TimeStamp `json:"last_changed_time,omitempty"`

	// Descriptive reason for the state-change. Field introduced in 18.2.6.
	Reason *string `json:"reason,omitempty"`

	// State for keeping track of reboot status during upgrade operation. Field introduced in 20.1.2. Allowed in Basic edition, Essentials edition, Enterprise edition.
	Rebooted *bool `json:"rebooted,omitempty"`

	// The upgrade operations current fsm-state. Enum options - UPGRADE_FSM_INIT, UPGRADE_FSM_STARTED, UPGRADE_FSM_WAITING, UPGRADE_FSM_IN_PROGRESS, UPGRADE_FSM_ENQUEUED, UPGRADE_FSM_ERROR, UPGRADE_FSM_SUSPENDED, UPGRADE_FSM_ENQUEUE_FAILED, UPGRADE_FSM_PAUSED, UPGRADE_FSM_COMPLETED, UPGRADE_FSM_ABORT_IN_PROGRESS, UPGRADE_FSM_ABORTED, UPGRADE_FSM_SE_UPGRADE_IN_PROGRESS, UPGRADE_FSM_CONTROLLER_COMPLETED, UPGRADE_FSM_DUMMY_3, UPGRADE_FSM_DUMMY_4, UPGRADE_FSM_DUMMY_5. Field introduced in 18.2.6.
	State *string `json:"state,omitempty"`
}
