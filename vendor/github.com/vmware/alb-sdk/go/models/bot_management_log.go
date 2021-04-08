package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// BotManagementLog bot management log
// swagger:model BotManagementLog
type BotManagementLog struct {

	// The final classification of the bot management module. Field introduced in 21.1.1.
	Classification *BotClassification `json:"classification,omitempty"`

	// The evaluation results of the various bot module components. Field introduced in 21.1.1.
	Results []*BotEvaluationResult `json:"results,omitempty"`
}
