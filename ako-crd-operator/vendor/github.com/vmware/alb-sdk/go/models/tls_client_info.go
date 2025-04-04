// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// TLSClientInfo Tls client info
// swagger:model TlsClientInfo
type TLSClientInfo struct {

	// The list of Cipher Suites in the ClientHello as integers. For example, TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA (0xc009) will be shown as 49161. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	CipherSuites []int64 `json:"cipher_suites,omitempty,omitempty"`

	// The TLS version in the ClientHello as integer. For example, TLSv1.2 (0x0303) will be shown as 771. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ClientHelloTLSVersion *uint32 `json:"client_hello_tls_version,omitempty"`

	// The list of supported EC Point Formats in the ClientHello as integers. For example, uncompressed will be shown as 0 (zero). Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	PointFormats []int64 `json:"point_formats,omitempty,omitempty"`

	// The list of TLS Supported Groups in the ClientHello as integers. For example, secp256r1 will be shown as 23. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SupportedGroups []int64 `json:"supported_groups,omitempty,omitempty"`

	// The list of TLS Extensions in the ClientHello as integers. For example, signature_algorithms will be shown as 13. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	TLSExtensions []int64 `json:"tls_extensions,omitempty,omitempty"`

	// Indicates whether the ClientHello contained GREASE ciphers, extensions or groups. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	UsesGrease *bool `json:"uses_grease,omitempty"`
}
