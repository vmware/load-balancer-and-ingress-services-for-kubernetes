package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// UpgradeTask upgrade task
// swagger:model UpgradeTask
type UpgradeTask struct {

	// duration of UpgradeTask.
	Duration *string `json:"duration,omitempty"`

	// end_time of UpgradeTask.
	EndTime *string `json:"end_time,omitempty"`

	// start_time of UpgradeTask.
	StartTime *string `json:"start_time,omitempty"`

	//  Enum options - COPY_AND_VERIFY_IMAGE, INSTALL_IMAGE, POST_INSTALL_HOOKS, PREPARE_CONTROLLER_FOR_SHUTDOWN, STOP_CONTROLLER, EXTRACT_PATCH_IMAGE, EXECUTE_PRE_INSTALL_COMMANDS, INSTALL_PATCH_IMAGE, PREPARE_FOR_REBOOT_CONTROLLER_NODES, REBOOT_CONTROLLER_NODES, WAIT_FOR_ALL_CONTROLLER_NODES_ONLINE, PRE_UPGRADE_HOOKS, MIGRATE_CONFIG, START_PRIMARY_CONTROLLER, START_ALL_CONTROLLERS, POST_UPGRADE_HOOKS, EXECUTE_POST_INSTALL_COMMANDS, SET_CONTROLLER_UPGRADE_COMPLETED, SE_UPGRADE_START, COMMIT_UPGRADE, UNKNOWN_TASK, PATCH_CONTROLLER_HEALTH_CHECK.
	Task *string `json:"task,omitempty"`
}
