// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ReportTask report task
// swagger:model ReportTask
type ReportTask struct {

	// Name for the Task journal. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	Name *string `json:"name"`

	// Reason in case of failure. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Reason *string `json:"reason,omitempty"`

	// Copy of journal summary for immediate visibility. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Summary *JournalSummary `json:"summary,omitempty"`

	// Journal reference for the task. It is a reference to an object of type TaskJournal. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	TaskJournalRef *string `json:"task_journal_ref,omitempty"`
}
