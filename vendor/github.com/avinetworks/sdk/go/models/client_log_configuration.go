package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ClientLogConfiguration client log configuration
// swagger:model ClientLogConfiguration
type ClientLogConfiguration struct {

	// Enable significant log collection. By default, this flag is enabled, which means that Avi SEs collect significant logs and forward them to Controller for further processing. For example, these logs correspond to error conditions such as when the response code for a request is 500. Users can deactivate this flag to turn off default significant log collection. Allowed in Basic(Allowed values- false) edition, Essentials(Allowed values- false) edition, Enterprise edition. Special default for Basic edition is false, Essentials edition is false, Enterprise is True.
	EnableSignificantLogCollection *bool `json:"enable_significant_log_collection,omitempty"`

	// Filtered logs are logs that match any client log filters or rules with logging enabled. Such logs are processed by the Logs Analytics system according to this setting. Enum options - LOGS_PROCESSING_NONE, LOGS_PROCESSING_SYNC_AND_INDEX_ON_DEMAND, LOGS_PROCESSING_AUTO_SYNC_AND_INDEX, LOGS_PROCESSING_AUTO_SYNC_BUT_INDEX_ON_DEMAND. Field introduced in 17.1.1.
	FilteredLogProcessing *string `json:"filtered_log_processing,omitempty"`

	// Logs that are neither significant nor filtered, are processed by the Logs Analytics system according to this setting. Enum options - LOGS_PROCESSING_NONE, LOGS_PROCESSING_SYNC_AND_INDEX_ON_DEMAND, LOGS_PROCESSING_AUTO_SYNC_AND_INDEX, LOGS_PROCESSING_AUTO_SYNC_BUT_INDEX_ON_DEMAND. Field introduced in 17.1.1.
	NonSignificantLogProcessing *string `json:"non_significant_log_processing,omitempty"`

	// Significant logs are processed by the Logs Analytics system according to this setting. Enum options - LOGS_PROCESSING_NONE, LOGS_PROCESSING_SYNC_AND_INDEX_ON_DEMAND, LOGS_PROCESSING_AUTO_SYNC_AND_INDEX, LOGS_PROCESSING_AUTO_SYNC_BUT_INDEX_ON_DEMAND. Field introduced in 17.1.1.
	SignificantLogProcessing *string `json:"significant_log_processing,omitempty"`
}
