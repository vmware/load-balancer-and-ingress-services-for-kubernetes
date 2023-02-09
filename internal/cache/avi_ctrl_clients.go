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
	corev1 "k8s.io/api/core/v1"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/api/models"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/third_party/github.com/vmware/alb-sdk/go/session"
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
	ctrlCAData := ctrlProp[utils.ENV_CTRL_CADATA]
	ctrlIpAddress := lib.GetControllerIP()
	if ctrlUsername == "" || (ctrlPassword == "" && ctrlAuthToken == "") || ctrlIpAddress == "" {
		var passwordLog, authTokenLog string
		if ctrlPassword != "" {
			passwordLog = "<sensitive>"
		}
		if ctrlAuthToken != "" {
			authTokenLog = "<sensitive>"
		}
		lib.AKOControlConfig().PodEventf(
			corev1.EventTypeWarning,
			lib.AKOShutdown, "Avi Controller information missing (username: %s, password: %s, authToken: %s, controller: %s)",
			ctrlUsername, passwordLog, authTokenLog, ctrlIpAddress,
		)
		utils.AviLog.Fatalf("Avi Controller information missing (username: %s, password: %s, authToken: %s, controller: %s). Update them in avi-secret.", ctrlUsername, passwordLog, authTokenLog, ctrlIpAddress)
	}

	if AviClientInstance == nil || len(AviClientInstance.AviClient) == 0 {
		// Always create 9 clients irrespective of shard size
		var currentControllerVersion string
		ctrlVersion := lib.AKOControlConfig().ControllerVersion()
		AviClientInstance, currentControllerVersion, err = utils.NewAviRestClientPool(
			9,
			ctrlIpAddress,
			ctrlUsername,
			ctrlPassword,
			ctrlAuthToken,
			ctrlVersion,
			ctrlCAData,
		)

		connectionStatus = utils.AVIAPI_CONNECTED
		if err != nil {
			connectionStatus = utils.AVIAPI_DISCONNECTED
			utils.AviLog.Errorf("AVI controller initialization failed")
			return nil
		}

		if ctrlVersion == "" {
			lib.AKOControlConfig().SetControllerVersion(currentControllerVersion)
			ctrlVersion = currentControllerVersion
		}
		// set the tenant and controller version in avisession obj
		for _, client := range AviClientInstance.AviClient {
			SetTenant := session.SetTenant(lib.GetTenant())
			SetTenant(client.AviSession)

			SetVersion := session.SetVersion(ctrlVersion)
			SetVersion(client.AviSession)
		}
	}

	models.RestStatus.UpdateAviApiRestStatus(connectionStatus, err)
	return AviClientInstance
}
