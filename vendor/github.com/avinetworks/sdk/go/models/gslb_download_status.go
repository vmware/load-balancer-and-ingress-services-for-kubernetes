package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// GslbDownloadStatus gslb download status
// swagger:model GslbDownloadStatus
type GslbDownloadStatus struct {

	//  Field introduced in 17.1.1.
	LastChangedTime *TimeStamp `json:"last_changed_time,omitempty"`

	// This field indicates the download state to a dns-vs(es) or a VS or a SE depending on the usage context. . Enum options - GSLB_DOWNLOAD_NONE, GSLB_DOWNLOAD_DONE, GSLB_DOWNLOAD_PENDING, GSLB_DOWNLOAD_ERROR. Field introduced in 17.1.1.
	State *string `json:"state,omitempty"`
}
