/*
 * Copyright 2019-2020 VMware, Inc.
 * All Rights Reserved.
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
*   http://www.apache.org/licenses/LICENSE-2.0
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
*/

package cache

import (
	"os"

	"github.com/vmware/alb-sdk/go/session"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/api/models"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

var AviClientInstance *utils.AviRestClientPool

// This class is in control of AKC. It uses utils from the common project.
func SharedAVIClients() *utils.AviRestClientPool {
	var err error
	var connectionStatus string

	ctrlProp := utils.SharedCtrlProp().GetAllCtrlProp()
	ctrlUsername := ctrlProp[utils.ENV_CTRL_USERNAME]
	ctrlPassword := ctrlProp[utils.ENV_CTRL_PASSWORD]
	ctrlAuthToken := ctrlProp[utils.ENV_CTRL_AUTHTOKEN]
	ctrlIpAddress := os.Getenv(utils.ENV_CTRL_IPADDRESS)
	if ctrlUsername == "" || (ctrlPassword == "" && ctrlAuthToken == "") || ctrlIpAddress == "" {
		utils.AviLog.Fatal("AVI controller information missing. Update them in kubernetes secret or via environment variables.")
	}

	if AviClientInstance == nil || len(AviClientInstance.AviClient) == 0 {

		// Always create 9 clients irrespective of shard size
		AviClientInstance, err = utils.NewAviRestClientPool(
			9,
			ctrlIpAddress,
			ctrlUsername,
			ctrlPassword,
			ctrlAuthToken,
		)
		connectionStatus = utils.AVIAPI_CONNECTED
		if err != nil {
			connectionStatus = utils.AVIAPI_DISCONNECTED
			utils.AviLog.Error("AVI controller initilization failed")
			return nil
		}
		controllerVersion := lib.SetControllerVersion()
		// set the tenant and controller version in avisession obj
		for _, client := range AviClientInstance.AviClient {
			SetTenant := session.SetTenant(lib.GetTenant())
			SetTenant(client.AviSession)

			lib.SetEnableCtrl2014Features(controllerVersion)
			SetVersion := session.SetVersion(controllerVersion)
			SetVersion(client.AviSession)

		}

	}

	models.RestStatus.UpdateAviApiRestStatus(connectionStatus, err)
	return AviClientInstance
}
