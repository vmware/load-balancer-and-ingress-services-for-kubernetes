package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// DebugVirtualServiceCapture debug virtual service capture
// swagger:model DebugVirtualServiceCapture
type DebugVirtualServiceCapture struct {

	// Maximum allowed size of PCAP Capture File per SE. Max(absolute_size, percentage_size) will be final value. Set both to 0 for avi default size. DOS, IPC and DROP pcaps not applicaple. Field introduced in 18.2.8.
	CaptureFileSize *CaptureFileSize `json:"capture_file_size,omitempty"`

	// Number of minutes to capture packets. Use 0 to capture until manually stopped. Special values are 0 - 'infinite'. Unit is MIN.
	Duration *int32 `json:"duration,omitempty"`

	// Enable SSL session key capture. Field introduced in 18.2.3.
	EnableSslSessionKeyCapture *bool `json:"enable_ssl_session_key_capture,omitempty"`

	// Number of files to maintain for SE pcap file rotation.file count set to 1 indicates no rotate. Allowed values are 1-10. Field introduced in 20.1.1.
	FileCount *int32 `json:"file_count,omitempty"`

	// Total number of packets to capture.
	NumPkts *int32 `json:"num_pkts,omitempty"`

	// Enable PcapNg for packet capture. Field introduced in 18.2.5.
	PcapNg *bool `json:"pcap_ng,omitempty"`

	// Number of bytes of each packet to capture. Use 0 to capture the entire packet. Allowed values are 64-1514. Special values are 0 - 'full capture'. Unit is BYTES.
	PktSize *int32 `json:"pkt_size,omitempty"`
}
