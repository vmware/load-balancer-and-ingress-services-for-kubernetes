// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// GslbSiteRuntimeStats gslb site runtime stats
// swagger:model GslbSiteRuntimeStats
type GslbSiteRuntimeStats struct {

	//  Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumFileCrTxed *int32 `json:"num_file_cr_txed,omitempty"`

	//  Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumFileDelTxed *int32 `json:"num_file_del_txed,omitempty"`

	//  Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumGapCrRxed *int32 `json:"num_gap_cr_rxed,omitempty"`

	//  Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumGapCrTxed *int32 `json:"num_gap_cr_txed,omitempty"`

	//  Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumGapDelRxed *int32 `json:"num_gap_del_rxed,omitempty"`

	//  Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumGapDelTxed *int32 `json:"num_gap_del_txed,omitempty"`

	//  Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumGapUpdRxed *int32 `json:"num_gap_upd_rxed,omitempty"`

	//  Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumGapUpdTxed *int32 `json:"num_gap_upd_txed,omitempty"`

	//  Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumGeoCrRxed *int32 `json:"num_geo_cr_rxed,omitempty"`

	//  Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumGeoCrTxed *int32 `json:"num_geo_cr_txed,omitempty"`

	//  Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumGeoDelRxed *int32 `json:"num_geo_del_rxed,omitempty"`

	//  Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumGeoDelTxed *int32 `json:"num_geo_del_txed,omitempty"`

	//  Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumGeoUpdRxed *int32 `json:"num_geo_upd_rxed,omitempty"`

	//  Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumGeoUpdTxed *int32 `json:"num_geo_upd_txed,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumGhmCrRxed *int32 `json:"num_ghm_cr_rxed,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumGhmCrTxed *int32 `json:"num_ghm_cr_txed,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumGhmDelRxed *int32 `json:"num_ghm_del_rxed,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumGhmDelTxed *int32 `json:"num_ghm_del_txed,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumGhmUpdRxed *int32 `json:"num_ghm_upd_rxed,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumGhmUpdTxed *int32 `json:"num_ghm_upd_txed,omitempty"`

	//  Field introduced in 20.1.5. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	NumGjwtCrRxed *int32 `json:"num_gjwt_cr_rxed,omitempty"`

	//  Field introduced in 20.1.5. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	NumGjwtCrTxed *int32 `json:"num_gjwt_cr_txed,omitempty"`

	//  Field introduced in 20.1.5. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	NumGjwtDelRxed *int32 `json:"num_gjwt_del_rxed,omitempty"`

	//  Field introduced in 20.1.5. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	NumGjwtDelTxed *int32 `json:"num_gjwt_del_txed,omitempty"`

	//  Field introduced in 20.1.5. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	NumGjwtUpdRxed *int32 `json:"num_gjwt_upd_rxed,omitempty"`

	//  Field introduced in 20.1.5. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	NumGjwtUpdTxed *int32 `json:"num_gjwt_upd_txed,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumGlbCrRxed *int32 `json:"num_glb_cr_rxed,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumGlbCrTxed *int32 `json:"num_glb_cr_txed,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumGlbDelRxed *int32 `json:"num_glb_del_rxed,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumGlbDelTxed *int32 `json:"num_glb_del_txed,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumGlbUpdRxed *int32 `json:"num_glb_upd_rxed,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumGlbUpdTxed *int32 `json:"num_glb_upd_txed,omitempty"`

	//  Field introduced in 17.1.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumGpkiCrRxed *int32 `json:"num_gpki_cr_rxed,omitempty"`

	//  Field introduced in 17.1.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumGpkiCrTxed *int32 `json:"num_gpki_cr_txed,omitempty"`

	//  Field introduced in 17.1.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumGpkiDelRxed *int32 `json:"num_gpki_del_rxed,omitempty"`

	//  Field introduced in 17.1.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumGpkiDelTxed *int32 `json:"num_gpki_del_txed,omitempty"`

	//  Field introduced in 17.1.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumGpkiUpdRxed *int32 `json:"num_gpki_upd_rxed,omitempty"`

	//  Field introduced in 17.1.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumGpkiUpdTxed *int32 `json:"num_gpki_upd_txed,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumGsCrRxed *int32 `json:"num_gs_cr_rxed,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumGsCrTxed *int32 `json:"num_gs_cr_txed,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumGsDelRxed *int32 `json:"num_gs_del_rxed,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumGsDelTxed *int32 `json:"num_gs_del_txed,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumGsUpdRxed *int32 `json:"num_gs_upd_rxed,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumGsUpdTxed *int32 `json:"num_gs_upd_txed,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumHealthMsgsRxed *int32 `json:"num_health_msgs_rxed,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumHealthMsgsTxed *int32 `json:"num_health_msgs_txed,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumOfBadResponses *int32 `json:"num_of_bad_responses,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumOfEventsGenerated *int32 `json:"num_of_events_generated,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumOfSkipOutstandingRequests *int32 `json:"num_of_skip_outstanding_requests,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumOfTimeouts *int32 `json:"num_of_timeouts,omitempty"`
}
