package rest

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

func (rest *RestOperations) AviHealthMonitorDel(uuid string, tenant string, key string) *utils.RestOp {
	path := "/api/healthmonitor/" + uuid
	rest_op := utils.RestOp{
		Path:   path,
		Method: "DELETE",
		Tenant: tenant,
		Model:  "HealthMonitor",
	}
	utils.AviLog.Debug(spew.Sprintf("HealthMonitor DELETE Restop %v ",
		utils.Stringify(rest_op)))
	return &rest_op
}
