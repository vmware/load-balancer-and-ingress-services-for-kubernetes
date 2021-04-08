package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// GslbDNSSeInfo gslb Dns se info
// swagger:model GslbDnsSeInfo
type GslbDNSSeInfo struct {

	// This field describes the fd download status to the SE. Field deprecated in 18.2.3. Field introduced in 17.1.1.
	FdDownload *GslbDownloadStatus `json:"fd_download,omitempty"`

	// Geo files queue for sequencing files to SE. Field deprecated in 18.2.3. Field introduced in 17.1.1.
	FdInfo *ConfigInfo `json:"fd_info,omitempty"`

	// Service engine's fabric IP used to push Geo files. Field deprecated in 18.2.3. Field introduced in 17.1.1.
	IP *IPAddr `json:"ip,omitempty"`

	// UUID of the service engine. Field deprecated in 18.2.3. Field introduced in 17.1.1.
	UUID *string `json:"uuid,omitempty"`
}
