package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:resource:shortName=sslprofile,singular=sslprofile,path=sslprofiles
// SSLProfile is the Schema for the ssl profile API
type SSLProfile struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec SSLProfileSpec `json:"spec,omitempty"`
	// +optional
	Status SSLProfileStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// SSLProfileList contains a list of SSLProfile
type SSLProfileList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SSLProfile `json:"items"`
}

// SSLProfileSpec defines the desired state of SSLProfile
type SSLProfileSpec struct {
	// +kubebuilder:validation:Required
	AcceptedVersions []SSLVersionType `json:"accepted_versions,omitempty"`
	// +kubebuilder:validation:Optional
	AcceptedCiphers string `json:"accepted_ciphers,omitempty"`
	// +kubebuilder:validation:Optional
	CipherEnums []AcceptedCipherEnums `json:"cipher_enums,omitempty"`
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=true
	SendCloseNotify bool `json:"send_close_notify,omitempty"`
	// +kubebuilder:validation:Optional
	PreferClientCipherOrdering bool `json:"prefer_client_cipher_ordering,omitempty"`
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=true
	EnableSSLSessionReuse bool `json:"enable_ssl_session_reuse,omitempty"`
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=86400
	// +kubebuilder:default:=86400
	SSLSessionTimeout int32 `json:"ssl_session_timeout,omitempty"`
	// +kubebuilder:validation:Required
	// +kubebuilder:default:=SSL_PROFILE_TYPE_APPLICATION
	SSLProfileType SSLProfileType `json:"ssl_profile_type,omitempty"`
	// +kubebuilder:validation:Optional
	Ciphersuites string `json:"ciphersuites,omitempty"`
	// +kubebuilder:validation:Optional
	EnableEarlyData bool `json:"enable_early_data,omitempty"`
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=auto
	ECNamedCurve string `json:"ec_named_curve,omitempty"`
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=auto
	SignatureAlgorithm string `json:"signature_algorithm,omitempty"`
	// +kubebuilder:validation:Optional
	Description string `json:"description,omitempty"`
	// +kubebuilder:validation:Optional
	IsFederated bool `json:"is_federated,omitempty"`
}

// SSLProfileStatus defines the observed state of SSLProfile
type SSLProfileStatus struct {
	Accepted bool   `json:"accepted"`
	Url      string `json:"url"`
}

// SSLVersionType represents an ssl version type
// +kubebuilder:validation:Enum=SSL_VERSION_SSLV3;SSL_VERSION_TLS1;SSL_VERSION_TLS1_1;SSL_VERSION_TLS1_2;SSL_VERSION_TLS1_3
type SSLVersionType string

const (
	SSLVersionSSLV3  SSLVersionType = "SSL_VERSION_SSLV3"
	SSLVersionTLS1   SSLVersionType = "SSL_VERSION_TLS1"
	SSLVersionTLS1_1 SSLVersionType = "SSL_VERSION_TLS1_1"
	SSLVersionTLS1_2 SSLVersionType = "SSL_VERSION_TLS1_2"
	SSLVersionTLS1_3 SSLVersionType = "SSL_VERSION_TLS1_3"
)

// SSLProfileType represents an ssl profile type
// +kubebuilder:validation:Enum=SSL_PROFILE_TYPE_APPLICATION;SSL_PROFILE_TYPE_SYSTEM
type SSLProfileType string

const (
	SSLProfileTypeApplication SSLProfileType = "SSL_PROFILE_TYPE_APPLICATION"
	SSLProfileTypeSystem      SSLProfileType = "SSL_PROFILE_TYPE_SYSTEM"
)

// AcceptedCipherEnums represents an accepted cipher
// +kubebuilder:validation:Enum=TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256;TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384;TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256;TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384;TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256;TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA384;TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256;TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA384;TLS_RSA_WITH_AES_128_GCM_SHA256;TLS_RSA_WITH_AES_256_GCM_SHA384;TLS_RSA_WITH_AES_128_CBC_SHA256;TLS_RSA_WITH_AES_256_CBC_SHA256;TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA;TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA;TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA;TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA;TLS_RSA_WITH_AES_128_CBC_SHA;TLS_RSA_WITH_AES_256_CBC_SHA;TLS_RSA_WITH_3DES_EDE_CBC_SHA
type AcceptedCipherEnums string

const (
	AcceptedCipherEnumsTLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256 AcceptedCipherEnums = "TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256"
	AcceptedCipherEnumsTLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384 AcceptedCipherEnums = "TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384"
	AcceptedCipherEnumsTLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256   AcceptedCipherEnums = "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"
	AcceptedCipherEnumsTLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384   AcceptedCipherEnums = "TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384"
	AcceptedCipherEnumsTLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256 AcceptedCipherEnums = "TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256"
	AcceptedCipherEnumsTLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA384 AcceptedCipherEnums = "TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA384"
	AcceptedCipherEnumsTLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256   AcceptedCipherEnums = "TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256"
	AcceptedCipherEnumsTLS_ECDHE_RSA_WITH_AES_256_CBC_SHA384   AcceptedCipherEnums = "TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA384"
	AcceptedCipherEnumsTLS_RSA_WITH_AES_128_GCM_SHA256         AcceptedCipherEnums = "TLS_RSA_WITH_AES_128_GCM_SHA256"
	AcceptedCipherEnumsTLS_RSA_WITH_AES_256_GCM_SHA384         AcceptedCipherEnums = "TLS_RSA_WITH_AES_256_GCM_SHA384"
	AcceptedCipherEnumsTLS_RSA_WITH_AES_128_CBC_SHA256         AcceptedCipherEnums = "TLS_RSA_WITH_AES_128_CBC_SHA256"
	AcceptedCipherEnumsTLS_RSA_WITH_AES_256_CBC_SHA256         AcceptedCipherEnums = "TLS_RSA_WITH_AES_256_CBC_SHA256"
	AcceptedCipherEnumsTLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA    AcceptedCipherEnums = "TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA"
	AcceptedCipherEnumsTLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA    AcceptedCipherEnums = "TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA"
	AcceptedCipherEnumsTLS_ECDHE_RSA_WITH_AES_128_CBC_SHA      AcceptedCipherEnums = "TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA"
	AcceptedCipherEnumsTLS_ECDHE_RSA_WITH_AES_256_CBC_SHA      AcceptedCipherEnums = "TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA"
	AcceptedCipherEnumsTLS_RSA_WITH_AES_128_CBC_SHA            AcceptedCipherEnums = "TLS_RSA_WITH_AES_128_CBC_SHA"
	AcceptedCipherEnumsTLS_RSA_WITH_AES_256_CBC_SHA            AcceptedCipherEnums = "TLS_RSA_WITH_AES_256_CBC_SHA"
	AcceptedCipherEnumsTLS_RSA_WITH_3DES_EDE_CBC_SHA           AcceptedCipherEnums = "TLS_RSA_WITH_3DES_EDE_CBC_SHA"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:resource:shortName=sslkeyandcertificate,singular=sslkeyandcertificate,path=sslkeyandcertificates
// SSLKeyAndCertificate is the Schema for the ssl key and certificate API
type SSLKeyAndCertificate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec SSLKeyAndCertificateSpec `json:"spec,omitempty"`
	// +optional
	Status SSLKeyAndCertificateStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// SSLKeyAndCertificateList contains a list of SSLKeyAndCertificate
type SSLKeyAndCertificateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SSLKeyAndCertificate `json:"items"`
}

// +kubebuilder:validation:XValidation:rule="self.spec.self_signed_certificate.common_name != '' || self.spec.certificate_secret_ref != ''",message="If self signed certificate is not defined, you must define the certificate secret ref"
// +kubebuilder:validation:XValidation:rule="self.spec.certificate_secret_ref != '' && self.spec.key_secret_ref != ''",message="If certificate secret ref is defined, you must define the key secret ref"

// SSLKeyAndCertificateSpec defines the desired state of SSLKeyAndCertificate
type SSLKeyAndCertificateSpec struct {
	// +kubebuilder:validation:Optional
	CertificateSecretRef string `json:"certificate_secret_ref"`
	// +kubebuilder:validation:Optional
	KeySecretRef string `json:"key_secret_ref"`
	// +kubebuilder:validation:Optional
	SelfSignedCertificate SSLCertificateDescription `json:"self_signed_certificate"`
	// +optional
	KeyPassphrase string `json:"key_passphrase,omitempty"`
	// +optional
	EnableOCSPStapling bool `json:"enable_ocsp_stapling,omitempty"`
	// +optional
	OCSPConfig OCSPConfig `json:"ocsp_config,omitempty"`
	// +optional
	IsFederated bool `json:"is_federated,omitempty"`
}

// SSLCertificateDescription represents common fields in a certificate (subject, issuer).
type SSLCertificateDescription struct {
	// +kubebuilder:validation:Required
	CommonName string `json:"common_name,omitempty"`
	// +optional
	EmailAddress string `json:"email_address,omitempty"`
	// +optional
	OrganizationUnit string `json:"organization_unit,omitempty"`
	// +optional
	Organization string `json:"organization,omitempty"`
	// +optional
	Locality string `json:"locality,omitempty"`
	// +optional
	State string `json:"state,omitempty"`
	// +optional
	Country string `json:"country,omitempty"`
}

// OCSPConfig represents the OCSP configuration.
type OCSPConfig struct {
	// +kubebuilder:validation:Minimum=60
	// +kubebuilder:validation:Maximum=31536000
	// +kubebuilder:default:=86400
	// +optional
	OCSPReqInterval int32 `json:"ocsp_req_interval,omitempty"`
	// +optional
	OCSPRespTimeout int32 `json:"ocsp_resp_timeout,omitempty"`
	// +optional
	ResponderURLLists []string `json:"responder_url_lists,omitempty"`
	// +kubebuilder:validation:Enum=OCSP_RESPONDER_URL_FAILOVER;OCSP_RESPONDER_URL_OVERRIDE
	// +kubebuilder:default:="OCSP_RESPONDER_URL_FAILOVER"
	// +optional
	URLAction OCSPResponderUrlAction `json:"url_action,omitempty"`
	// +kubebuilder:validation:Minimum=60
	// +kubebuilder:validation:Maximum=86400
	// +kubebuilder:default:=3600
	// +optional
	FailedOCSPJobsRetryInterval int32 `json:"failed_ocsp_jobs_retry_interval,omitempty"`
	// +kubebuilder:default:=10
	// +optional
	MaxTries int32 `json:"max_tries,omitempty"`
}

// OCSPResponderUrlAction represents an OCSP responder action.
// +kubebuilder:validation:Enum=OCSP_RESPONDER_URL_FAILOVER;OCSP_RESPONDER_URL_OVERRIDE
type OCSPResponderUrlAction string

const (
	OCSPResponderURLFailover OCSPResponderUrlAction = "OCSP_RESPONDER_URL_FAILOVER"
	OCSPResponderURLOverride OCSPResponderUrlAction = "OCSP_RESPONDER_URL_OVERRIDE"
)

// SSLKeyAlgorithm represents an ssl key algorithm
// +kubebuilder:validation:Enum=SSL_KEY_ALGORITHM_RSA;SSL_KEY_ALGORITHM_EC
type SSLKeyAlgorithm string

const (
	SSLKeyAlgorithmRSA SSLKeyAlgorithm = "SSL_KEY_ALGORITHM_RSA"
	SSLKeyAlgorithmEC  SSLKeyAlgorithm = "SSL_KEY_ALGORITHM_EC"
)

// SSLKeyAndCertificateStatus defines the observed state of SSLKeyAndCertificate
type SSLKeyAndCertificateStatus struct {
	SSLCertificateStatus SSLCertificateStatus `json:"ssl_certificate_status"`
	Accepted             bool                 `json:"accepted"`
	Url                  string               `json:"url"`
}

// SSLCertificateStatus represents an ssl certificate status
// +kubebuilder:validation:Enum=SSL_CERTIFICATE_FINISHED;SSL_CERTIFICATE_PENDING
type SSLCertificateStatus string

const (
	SSLCertificateStatusFinished SSLCertificateStatus = "SSL_CERTIFICATE_FINISHED"
	SSLCertificateStatusPending  SSLCertificateStatus = "SSL_CERTIFICATE_PENDING"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:resource:shortName=pkiprofile,singular=pkiprofile,path=pkiprofiles
// PKIProfile is the Schema for the pki profile API
type PKIProfile struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec PKIProfileSpec `json:"spec,omitempty"`
	// +optional
	Status PKIProfileStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// PKIProfileList contains a list of PKIProfile
type PKIProfileList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PKIProfile `json:"items"`
}

// PKIProfileSpec defines the desired state of PKIProfile
type PKIProfileSpec struct {
	// +kubebuilder:validation:MinItems=1
	CACertsSecretRef []string `json:"ca_certs,omitempty"`
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:XValidation:rule="self.crl_check == false || self.crl_data != null",message="crl_data is mandatory if crl_check is set to true"
	CRLData CRLDataRef `json:"crl_data,omitempty"`
	// +kubebuilder:validation:Optional
	IgnorePeerChain bool `json:"ignore_peer_chain,omitempty"`
	// +kubebuilder:validation:Optional
	CRLCheck bool `json:"crl_check,omitempty"`
	// +kubebuilder:validation:Optional
	ValidateOnlyLeafCRL bool `json:"validate_only_leaf_crl,omitempty"`
	// +optional
	IsFederated bool `json:"is_federated,omitempty"`
}

// +kubebuilder:validation:XValidation:rule="self.crl_data == null || size(self.crl_data.crl_file_uuids) > 0 || size(self.crl_data.crl_server_url) > 0", message="When crl_data is specified, at least one of crl_file_uuids or crl_server_url must be specified"

type CRLDataRef struct {
	// Avi controller CRL fileobject ref
	CRLFileUUIDs []string `json:"crl_file_uuids,omitempty"`
	// CRL server URL
	CRLServerURL []string `json:"crl_server_url"`
}

// PKIProfileStatus defines the observed state of PKIProfile
type PKIProfileStatus struct {
	Accepted bool   `json:"accepted"`
	Url      string `json:"url"`
}
