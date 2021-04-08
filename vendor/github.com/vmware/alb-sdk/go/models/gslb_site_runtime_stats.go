package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// GslbSiteRuntimeStats gslb site runtime stats
// swagger:model GslbSiteRuntimeStats
type GslbSiteRuntimeStats struct {

	//  Field introduced in 17.1.1.
	NumFileCrTxed *int32 `json:"num_file_cr_txed,omitempty"`

	//  Field introduced in 17.1.1.
	NumFileDelTxed *int32 `json:"num_file_del_txed,omitempty"`

	//  Field introduced in 17.1.1.
	NumGapCrRxed *int32 `json:"num_gap_cr_rxed,omitempty"`

	//  Field introduced in 17.1.1.
	NumGapCrTxed *int32 `json:"num_gap_cr_txed,omitempty"`

	//  Field introduced in 17.1.1.
	NumGapDelRxed *int32 `json:"num_gap_del_rxed,omitempty"`

	//  Field introduced in 17.1.1.
	NumGapDelTxed *int32 `json:"num_gap_del_txed,omitempty"`

	//  Field introduced in 17.1.1.
	NumGapUpdRxed *int32 `json:"num_gap_upd_rxed,omitempty"`

	//  Field introduced in 17.1.1.
	NumGapUpdTxed *int32 `json:"num_gap_upd_txed,omitempty"`

	//  Field introduced in 17.1.1.
	NumGeoCrRxed *int32 `json:"num_geo_cr_rxed,omitempty"`

	//  Field introduced in 17.1.1.
	NumGeoCrTxed *int32 `json:"num_geo_cr_txed,omitempty"`

	//  Field introduced in 17.1.1.
	NumGeoDelRxed *int32 `json:"num_geo_del_rxed,omitempty"`

	//  Field introduced in 17.1.1.
	NumGeoDelTxed *int32 `json:"num_geo_del_txed,omitempty"`

	//  Field introduced in 17.1.1.
	NumGeoUpdRxed *int32 `json:"num_geo_upd_rxed,omitempty"`

	//  Field introduced in 17.1.1.
	NumGeoUpdTxed *int32 `json:"num_geo_upd_txed,omitempty"`

	// Number of num_ghm_cr_rxed.
	NumGhmCrRxed *int32 `json:"num_ghm_cr_rxed,omitempty"`

	// Number of num_ghm_cr_txed.
	NumGhmCrTxed *int32 `json:"num_ghm_cr_txed,omitempty"`

	// Number of num_ghm_del_rxed.
	NumGhmDelRxed *int32 `json:"num_ghm_del_rxed,omitempty"`

	// Number of num_ghm_del_txed.
	NumGhmDelTxed *int32 `json:"num_ghm_del_txed,omitempty"`

	// Number of num_ghm_upd_rxed.
	NumGhmUpdRxed *int32 `json:"num_ghm_upd_rxed,omitempty"`

	// Number of num_ghm_upd_txed.
	NumGhmUpdTxed *int32 `json:"num_ghm_upd_txed,omitempty"`

	// Number of num_glb_cr_rxed.
	NumGlbCrRxed *int32 `json:"num_glb_cr_rxed,omitempty"`

	// Number of num_glb_cr_txed.
	NumGlbCrTxed *int32 `json:"num_glb_cr_txed,omitempty"`

	// Number of num_glb_del_rxed.
	NumGlbDelRxed *int32 `json:"num_glb_del_rxed,omitempty"`

	// Number of num_glb_del_txed.
	NumGlbDelTxed *int32 `json:"num_glb_del_txed,omitempty"`

	// Number of num_glb_upd_rxed.
	NumGlbUpdRxed *int32 `json:"num_glb_upd_rxed,omitempty"`

	// Number of num_glb_upd_txed.
	NumGlbUpdTxed *int32 `json:"num_glb_upd_txed,omitempty"`

	//  Field introduced in 17.1.3.
	NumGpkiCrRxed *int32 `json:"num_gpki_cr_rxed,omitempty"`

	//  Field introduced in 17.1.3.
	NumGpkiCrTxed *int32 `json:"num_gpki_cr_txed,omitempty"`

	//  Field introduced in 17.1.3.
	NumGpkiDelRxed *int32 `json:"num_gpki_del_rxed,omitempty"`

	//  Field introduced in 17.1.3.
	NumGpkiDelTxed *int32 `json:"num_gpki_del_txed,omitempty"`

	//  Field introduced in 17.1.3.
	NumGpkiUpdRxed *int32 `json:"num_gpki_upd_rxed,omitempty"`

	//  Field introduced in 17.1.3.
	NumGpkiUpdTxed *int32 `json:"num_gpki_upd_txed,omitempty"`

	// Number of num_gs_cr_rxed.
	NumGsCrRxed *int32 `json:"num_gs_cr_rxed,omitempty"`

	// Number of num_gs_cr_txed.
	NumGsCrTxed *int32 `json:"num_gs_cr_txed,omitempty"`

	// Number of num_gs_del_rxed.
	NumGsDelRxed *int32 `json:"num_gs_del_rxed,omitempty"`

	// Number of num_gs_del_txed.
	NumGsDelTxed *int32 `json:"num_gs_del_txed,omitempty"`

	// Number of num_gs_upd_rxed.
	NumGsUpdRxed *int32 `json:"num_gs_upd_rxed,omitempty"`

	// Number of num_gs_upd_txed.
	NumGsUpdTxed *int32 `json:"num_gs_upd_txed,omitempty"`

	// Number of num_health_msgs_rxed.
	NumHealthMsgsRxed *int32 `json:"num_health_msgs_rxed,omitempty"`

	// Number of num_health_msgs_txed.
	NumHealthMsgsTxed *int32 `json:"num_health_msgs_txed,omitempty"`

	// Number of num_of_bad_responses.
	NumOfBadResponses *int32 `json:"num_of_bad_responses,omitempty"`

	// Number of num_of_events_generated.
	NumOfEventsGenerated *int32 `json:"num_of_events_generated,omitempty"`

	// Number of num_of_skip_outstanding_requests.
	NumOfSkipOutstandingRequests *int32 `json:"num_of_skip_outstanding_requests,omitempty"`

	// Number of num_of_timeouts.
	NumOfTimeouts *int32 `json:"num_of_timeouts,omitempty"`
}
