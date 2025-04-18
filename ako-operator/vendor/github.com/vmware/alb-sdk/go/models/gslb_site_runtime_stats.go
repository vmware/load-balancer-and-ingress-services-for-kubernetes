// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// GslbSiteRuntimeStats gslb site runtime stats
// swagger:model GslbSiteRuntimeStats
type GslbSiteRuntimeStats struct {

	//  Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumFileCrTxed *uint32 `json:"num_file_cr_txed,omitempty"`

	//  Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumFileDelTxed *uint32 `json:"num_file_del_txed,omitempty"`

	//  Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumGapCrRxed *uint32 `json:"num_gap_cr_rxed,omitempty"`

	//  Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumGapCrTxed *uint32 `json:"num_gap_cr_txed,omitempty"`

	//  Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumGapDelRxed *uint32 `json:"num_gap_del_rxed,omitempty"`

	//  Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumGapDelTxed *uint32 `json:"num_gap_del_txed,omitempty"`

	//  Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumGapUpdRxed *uint32 `json:"num_gap_upd_rxed,omitempty"`

	//  Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumGapUpdTxed *uint32 `json:"num_gap_upd_txed,omitempty"`

	//  Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumGeoCrRxed *uint32 `json:"num_geo_cr_rxed,omitempty"`

	//  Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumGeoCrTxed *uint32 `json:"num_geo_cr_txed,omitempty"`

	//  Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumGeoDelRxed *uint32 `json:"num_geo_del_rxed,omitempty"`

	//  Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumGeoDelTxed *uint32 `json:"num_geo_del_txed,omitempty"`

	//  Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumGeoUpdRxed *uint32 `json:"num_geo_upd_rxed,omitempty"`

	//  Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumGeoUpdTxed *uint32 `json:"num_geo_upd_txed,omitempty"`

	// Used for federated file object stats for create. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	NumGfoCrRxed *uint32 `json:"num_gfo_cr_rxed,omitempty"`

	// Used for federated file object stats for delete. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	NumGfoDelRxed *uint32 `json:"num_gfo_del_rxed,omitempty"`

	// Used for federated file object stats for update. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	NumGfoUpdRxed *uint32 `json:"num_gfo_upd_rxed,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumGhmCrRxed *uint32 `json:"num_ghm_cr_rxed,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumGhmCrTxed *uint32 `json:"num_ghm_cr_txed,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumGhmDelRxed *uint32 `json:"num_ghm_del_rxed,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumGhmDelTxed *uint32 `json:"num_ghm_del_txed,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumGhmUpdRxed *uint32 `json:"num_ghm_upd_rxed,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumGhmUpdTxed *uint32 `json:"num_ghm_upd_txed,omitempty"`

	//  Field introduced in 20.1.5. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	NumGjwtCrRxed *uint32 `json:"num_gjwt_cr_rxed,omitempty"`

	//  Field introduced in 20.1.5. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	NumGjwtCrTxed *uint32 `json:"num_gjwt_cr_txed,omitempty"`

	//  Field introduced in 20.1.5. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	NumGjwtDelRxed *uint32 `json:"num_gjwt_del_rxed,omitempty"`

	//  Field introduced in 20.1.5. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	NumGjwtDelTxed *uint32 `json:"num_gjwt_del_txed,omitempty"`

	//  Field introduced in 20.1.5. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	NumGjwtUpdRxed *uint32 `json:"num_gjwt_upd_rxed,omitempty"`

	//  Field introduced in 20.1.5. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	NumGjwtUpdTxed *uint32 `json:"num_gjwt_upd_txed,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumGlbCrRxed *uint32 `json:"num_glb_cr_rxed,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumGlbCrTxed *uint32 `json:"num_glb_cr_txed,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumGlbDelRxed *uint32 `json:"num_glb_del_rxed,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumGlbDelTxed *uint32 `json:"num_glb_del_txed,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumGlbUpdRxed *uint32 `json:"num_glb_upd_rxed,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumGlbUpdTxed *uint32 `json:"num_glb_upd_txed,omitempty"`

	//  Field introduced in 17.1.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumGpkiCrRxed *uint32 `json:"num_gpki_cr_rxed,omitempty"`

	//  Field introduced in 17.1.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumGpkiCrTxed *uint32 `json:"num_gpki_cr_txed,omitempty"`

	//  Field introduced in 17.1.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumGpkiDelRxed *uint32 `json:"num_gpki_del_rxed,omitempty"`

	//  Field introduced in 17.1.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumGpkiDelTxed *uint32 `json:"num_gpki_del_txed,omitempty"`

	//  Field introduced in 17.1.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumGpkiUpdRxed *uint32 `json:"num_gpki_upd_rxed,omitempty"`

	//  Field introduced in 17.1.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumGpkiUpdTxed *uint32 `json:"num_gpki_upd_txed,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumGsCrRxed *uint32 `json:"num_gs_cr_rxed,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumGsCrTxed *uint32 `json:"num_gs_cr_txed,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumGsDelRxed *uint32 `json:"num_gs_del_rxed,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumGsDelTxed *uint32 `json:"num_gs_del_txed,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumGsUpdRxed *uint32 `json:"num_gs_upd_rxed,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumGsUpdTxed *uint32 `json:"num_gs_upd_txed,omitempty"`

	// Used for federated ssl key and cert stats for create. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	NumGsslCertCrRxed *uint32 `json:"num_gssl_cert_cr_rxed,omitempty"`

	// Used for federated ssl key and cert stats for delete. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	NumGsslCertDelRxed *uint32 `json:"num_gssl_cert_del_rxed,omitempty"`

	// Used for federated ssl key and cert stats for update. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	NumGsslCertUpdRxed *uint32 `json:"num_gssl_cert_upd_rxed,omitempty"`

	// Used for federated ssl profile stats for create. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	NumGsslCrRxed *uint32 `json:"num_gssl_cr_rxed,omitempty"`

	// Used for federated ssl profile stats for delete. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	NumGsslDelRxed *uint32 `json:"num_gssl_del_rxed,omitempty"`

	// Used for federated ssl profile stats for update. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	NumGsslUpdRxed *uint32 `json:"num_gssl_upd_rxed,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumHealthMsgsRxed *uint32 `json:"num_health_msgs_rxed,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumHealthMsgsTxed *uint32 `json:"num_health_msgs_txed,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumOfBadResponses *uint32 `json:"num_of_bad_responses,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumOfEventsGenerated *uint32 `json:"num_of_events_generated,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumOfSkipOutstandingRequests *uint32 `json:"num_of_skip_outstanding_requests,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumOfTimeouts *uint32 `json:"num_of_timeouts,omitempty"`
}
