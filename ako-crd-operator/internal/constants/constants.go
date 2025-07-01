package constants

import "time"

const (
	HealthMonitorFinalizer      = "healthmonitor.ako.vmware.com/finalizer"
	HealthMonitorURL            = "/api/healthmonitor"
	Sensitive                   = "<sensitive>"
	ApplicationProfileFinalizer = "applicationprofile.ako.vmware.com/finalizer"
	ApplicationProfileURL       = "/api/applicationprofile"
	HealthMonitorSecretType     = "ako.vmware.com/basic-auth"
	RequeueInterval             = 5 * time.Minute
	ACCEPTED                    = "Accepted"
	REJECTED                    = "Rejected"
	AKOCRDController            = "AKOCRDController"
	NoObject                    = "Object not found"
)
