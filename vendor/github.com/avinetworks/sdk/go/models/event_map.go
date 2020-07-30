package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// EventMap event map
// swagger:model EventMap
type EventMap struct {

	// List of all events node wise.(Not in use). Field introduced in 18.2.6.
	NodesEvents []*UpgradeEvent `json:"nodes_events,omitempty"`

	// List of all events node wise. Field introduced in 18.2.10, 20.1.1.
	SubEvents []*UpgradeEvent `json:"sub_events,omitempty"`

	// Enum representing the task.(Not in use). Enum options - PREPARE_FOR_SHUTDOWN, COPY_AND_VERIFY_IMAGE, INSTALL_IMAGE, POST_INSTALL_HOOKS, PREPARE_CONTROLLER_FOR_SHUTDOWN, STOP_CONTROLLER, EXTRACT_PATCH_IMAGE, EXECUTE_PRE_INSTALL_COMMANDS, INSTALL_PATCH_IMAGE, PREPARE_FOR_REBOOT_CONTROLLER_NODES, REBOOT_CONTROLLER_NODES, WAIT_FOR_ALL_CONTROLLER_NODES_ONLINE, PRE_UPGRADE_HOOKS, MIGRATE_CONFIG, START_PRIMARY_CONTROLLER, START_ALL_CONTROLLERS, POST_UPGRADE_HOOKS, EXECUTE_POST_INSTALL_COMMANDS, SET_CONTROLLER_UPGRADE_COMPLETED, STATE_NOT_USED_IN_V2.... Field introduced in 18.2.6.
	Task *string `json:"task,omitempty"`

	// Name representing the task. Field introduced in 18.2.10, 20.1.1.
	TaskName *string `json:"task_name,omitempty"`
}
