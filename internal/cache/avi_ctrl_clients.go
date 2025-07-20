/*
 * Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
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
	"sync"

	corev1 "k8s.io/api/core/v1"

	"github.com/vmware/alb-sdk/go/session"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/api/models"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

var AviClientInstanceMap sync.Map

// This class is in control of AKC. It uses utils from the common project.
func SharedAVIClients(tenant string) *utils.AviRestClientPool {
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
		if ctrlIpAddress == "" {
			utils.AviLog.Fatalf("Avi Controller information missing (username: %s, password: %s, authToken: %s, controller: %s). Update the controller IP in ConfigMap : avi-k8s-config", ctrlUsername, passwordLog, authTokenLog, ctrlIpAddress)
		}
		utils.AviLog.Fatalf("Avi Controller information missing (username: %s, password: %s, authToken: %s, controller: %s). Update them in avi-secret.", ctrlUsername, passwordLog, authTokenLog, ctrlIpAddress)
	}

	aviClientInstance, ok := AviClientInstanceMap.Load(tenant)
	if ok {
		models.RestStatus.UpdateAviApiRestStatus(connectionStatus, err)
		return aviClientInstance.(*utils.AviRestClientPool)
	}

	userHeaders := utils.SharedCtrlProp().GetCtrlUserHeader()
	userHeaders[utils.XAviUserAgentHeader] = "AKO"
	apiScheme := utils.SharedCtrlProp().GetCtrlAPIScheme()

	// Always create 9 clients irrespective of shard size
	var currentControllerVersion string
	var aviRestClientPool *utils.AviRestClientPool
	ctrlVersion := lib.AKOControlConfig().ControllerVersion()
	aviRestClientPool, currentControllerVersion, err = utils.NewAviRestClientPool(
		9,
		ctrlIpAddress,
		ctrlUsername,
		ctrlPassword,
		ctrlAuthToken,
		ctrlVersion,
		ctrlCAData,
		tenant,
		apiScheme,
		userHeaders,
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
	for _, client := range aviRestClientPool.AviClient {
		SetVersion := session.SetVersion(ctrlVersion)
		SetVersion(client.AviSession)
	}
	AviClientInstanceMap.Store(tenant, aviRestClientPool)
	models.RestStatus.UpdateAviApiRestStatus(connectionStatus, err)
	return aviRestClientPool
}
