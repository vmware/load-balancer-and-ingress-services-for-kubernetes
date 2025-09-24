package lib

import "time"

const (
	HealthMonitorFinalizer            = "healthmonitor.ako.vmware.com/finalizer"
	HealthMonitorURL                  = "/api/healthmonitor"
	Sensitive                         = "<sensitive>"
	ApplicationProfileFinalizer       = "applicationprofile.ako.vmware.com/finalizer"
	ApplicationProfileURL             = "/api/applicationprofile"
	PKIProfileFinalizer               = "pkiprofile.ako.vmware.com/finalizer"
	PKIProfileURL                     = "/api/pkiprofile"
	HealthMonitorSecretType           = "ako.vmware.com/basic-auth"
	ObservedResourceVersionAnnotation = "ako.vmware.com/observedResourceVersion"
	RequeueInterval                   = 5 * time.Minute
	ACCEPTED                          = "Accepted"
	REJECTED                          = "Rejected"
	AKOCRDController                  = "AKOCRDController"
	NoObject                          = "Object not found"
	Prefix                            = "ako-crd-operator-"
)
