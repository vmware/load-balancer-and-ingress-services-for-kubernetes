// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// HealthMonitorFtp health monitor ftp
// swagger:model HealthMonitorFtp
type HealthMonitorFtp struct {

	// Filename to download with full path. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	Filename *string `json:"filename"`

	// FTP data transfer process mode. Enum options - FTP_PASSIVE_MODE, FTP_PORT_MODE. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	Mode *string `json:"mode"`

	// SSL attributes for FTPS monitor. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SslAttributes *HealthMonitorSSlattributes `json:"ssl_attributes,omitempty"`
}
