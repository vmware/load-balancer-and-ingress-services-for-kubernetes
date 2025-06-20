package constants

import "time"

const (
	HealthMonitorFinalizer      = "healthmonitor.ako.vmware.com/finalizer"
	HealthMonitorURL            = "/api/healthmonitor"
	Sensitive                   = "<sensitive>"
	ApplicationProfileFinalizer = "applicationprofile.ako.vmware.com/finalizer"
	ApplicationProfileURL       = "/api/applicationprofile"
	RequeueInterval             = 5 * time.Minute
)
