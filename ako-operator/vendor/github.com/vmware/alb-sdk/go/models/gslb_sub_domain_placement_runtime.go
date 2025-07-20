// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// GslbSubDomainPlacementRuntime gslb sub domain placement runtime
// swagger:model GslbSubDomainPlacementRuntime
type GslbSubDomainPlacementRuntime struct {

	// This field describes the placement status of fqdns mapping to the above Subdomain.  If placement allowed is true, then the fqdn/GslbService will be placed on the DNS-VS. Otherwise, it shall not be placed on the DNS-VS. . Field introduced in 17.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PlacementAllowed *bool `json:"placement_allowed,omitempty"`

	// This field identifies the Subdomain. Field introduced in 17.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SubDomain *string `json:"sub_domain,omitempty"`

	// This field describes the transition operation to be initiated downstream when subdomain placement rules change. For example  if a.com was not placed on dns-vs-1 and due to configuration change if it is to be placed on dns-vs-1, then all the GslbServices whose fqdn maps a.com will be pushed to dns-vs-1. In this scenario, the transition ops will be GSLB_CREATE. If there is a configuration change where a.com is not placed on dns-vs-1 then the transition ops will be GSLB_DELETE. . Enum options - GSLB_NONE, GSLB_CREATE, GSLB_UPDATE, GSLB_DELETE, GSLB_PURGE, GSLB_DECL. Field introduced in 17.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	TransitionOps *string `json:"transition_ops,omitempty"`
}
