// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SeObjsyncPeerDownDetails se objsync peer down details
// swagger:model SeObjsyncPeerDownDetails
type SeObjsyncPeerDownDetails struct {

	// Objsync peer SE UUIDs. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	PeerSeUuids *string `json:"peer_se_uuids"`
}
