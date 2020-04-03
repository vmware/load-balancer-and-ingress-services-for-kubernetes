package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// EventLog event log
// swagger:model EventLog
type EventLog struct {

	//  Enum options - EVENT_CONTEXT_SYSTEM, EVENT_CONTEXT_CONFIG, EVENT_CONTEXT_APP, EVENT_CONTEXT_ALL.
	Context *string `json:"context,omitempty"`

	// Summary of event details.
	DetailsSummary *string `json:"details_summary,omitempty"`

	// Event Description for each Event  in the table view.
	EventDescription *string `json:"event_description,omitempty"`

	// Placeholder for description of property event_details of obj type EventLog field type str  type object
	EventDetails *EventDetails `json:"event_details,omitempty"`

	//  Enum options - VINFRA_DISC_DC. VINFRA_DISC_HOST. VINFRA_DISC_CLUSTER. VINFRA_DISC_VM. VINFRA_DISC_NW. MGMT_NW_NAME_CHANGED. DISCOVERY_DATACENTER_DEL. VM_ADDED. VM_REMOVED. VINFRA_DISC_COMPLETE. VCENTER_ADDRESS_ERROR. SE_GROUP_CLUSTER_DEL. SE_GROUP_MGMT_NW_DEL. MGMT_NW_DEL. VCENTER_BAD_CREDENTIALS. ESX_HOST_UNREACHABLE. SERVER_DELETED. SE_GROUP_HOST_DEL. VINFRA_DISC_FAILURE. ESX_HOST_POWERED_DOWN...
	// Required: true
	EventID *string `json:"event_id"`

	// Pages in which event should come up.
	EventPages []string `json:"event_pages,omitempty"`

	// Placeholder for description of property ignore_event_details_display of obj type EventLog field type str  type boolean
	IgnoreEventDetailsDisplay *bool `json:"ignore_event_details_display,omitempty"`

	//  Enum options - EVENT_INTERNAL, EVENT_EXTERNAL.
	Internal *string `json:"internal,omitempty"`

	// Placeholder for description of property is_security_event of obj type EventLog field type str  type boolean
	IsSecurityEvent *bool `json:"is_security_event,omitempty"`

	//  Enum options - UNKNOWN. VSMGR. SEMGR. RESMGR. VIMGR. METRICSMGR. CONFIG. SE_GENERAL. SE_FLOWTABLE. SE_HM. SE_POOL_PERSISTENCE. SE_POOL. VSERVER. CLOUD_CONNECTOR. CLUSTERMGR. HSMGR. NW_MGR. LICENSE_MGR. RES_MONITOR. STATEDBCACHE...
	// Required: true
	Module *string `json:"module"`

	// obj_name of EventLog.
	ObjName *string `json:"obj_name,omitempty"`

	//  Enum options - VIRTUALSERVICE. POOL. HEALTHMONITOR. NETWORKPROFILE. APPLICATIONPROFILE. HTTPPOLICYSET. DNSPOLICY. SECURITYPOLICY. IPADDRGROUP. STRINGGROUP. SSLPROFILE. SSLKEYANDCERTIFICATE. NETWORKSECURITYPOLICY. APPLICATIONPERSISTENCEPROFILE. ANALYTICSPROFILE. VSDATASCRIPTSET. TENANT. PKIPROFILE. AUTHPROFILE. CLOUD...
	ObjType *string `json:"obj_type,omitempty"`

	// Unique object identifier of obj.
	ObjUUID *string `json:"obj_uuid,omitempty"`

	// Reason code for generating the event. This would be added to the alert where it would say alert generated  on event with reason <reason code>. Enum options - SYSERR_SUCCESS, SYSERR_FAILURE, SYSERR_OUT_OF_MEMORY, SYSERR_NO_ENT, SYSERR_INVAL, SYSERR_ACCESS, SYSERR_FAULT, SYSERR_IO, SYSERR_TIMEOUT, SYSERR_NOT_SUPPORTED, SYSERR_NOT_READY, SYSERR_UPGRADE_IN_PROGRESS, SYSERR_WARM_START_IN_PROGRESS, SYSERR_TRY_AGAIN, SYSERR_NOT_UPGRADING, SYSERR_PENDING, SYSERR_EVENT_GEN_FAILURE, SYSERR_CONFIG_PARAM_MISSING, SYSERR_BAD_REQUEST, SYSERR_TEST1...
	ReasonCode *string `json:"reason_code,omitempty"`

	// related objects corresponding to the events.
	RelatedUuids []string `json:"related_uuids,omitempty"`

	// Number of report_timestamp.
	// Required: true
	ReportTimestamp *int64 `json:"report_timestamp"`

	// tenant of EventLog.
	Tenant *string `json:"tenant,omitempty"`

	//  Field introduced in 17.2.1.
	TenantName *string `json:"tenant_name,omitempty"`
}
