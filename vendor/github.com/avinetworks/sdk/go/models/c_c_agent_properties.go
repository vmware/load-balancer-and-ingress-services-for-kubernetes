package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// CCAgentProperties c c agent properties
// swagger:model CC_AgentProperties
type CCAgentProperties struct {

	// Maximum polls to check for async jobs to finish.
	AsyncRetries *int32 `json:"async_retries,omitempty"`

	// Delay between each async job status poll check.
	AsyncRetriesDelay *int32 `json:"async_retries_delay,omitempty"`

	// Discovery poll target duration; a scale factor of 1+ is computed with the actual discovery (actual/target) and used to tweak slow and fast poll intervals.
	PollDurationTarget *int32 `json:"poll_duration_target,omitempty"`

	// Fast poll interval.
	PollFastTarget *int32 `json:"poll_fast_target,omitempty"`

	// Slow poll interval.
	PollSlowTarget *int32 `json:"poll_slow_target,omitempty"`

	// Maximum polls to check for vnics to be attached to VM.
	VnicRetries *int32 `json:"vnic_retries,omitempty"`

	// Delay between each vnic status poll check.
	VnicRetriesDelay *int32 `json:"vnic_retries_delay,omitempty"`
}
