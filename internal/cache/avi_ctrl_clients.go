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
	"errors"
	"os"
	"sync"

	"github.com/avinetworks/ako/internal/lib"
	"github.com/avinetworks/sdk/go/session"

	"github.com/avinetworks/ako/pkg/api/models"
	"github.com/avinetworks/ako/pkg/utils"
)

var AviClientInstance *utils.AviRestClientPool
var clientonce sync.Once

// This class is in control of AKC. It uses utils from the common project.
func SharedAVIClients() *utils.AviRestClientPool {
	var err error
	var connectionStatus string

	ctrlUsername := os.Getenv("CTRL_USERNAME")
	ctrlPassword := os.Getenv("CTRL_PASSWORD")
	ctrlIpAddress := os.Getenv("CTRL_IPADDRESS")
	if ctrlUsername == "" || ctrlPassword == "" || ctrlIpAddress == "" {
		utils.AviLog.Fatal("AVI controller information missing. Update them in kubernetes secret or via environment variables.")
	}

	if AviClientInstance == nil || len(AviClientInstance.AviClient) == 0 {
		shardSize := lib.GetshardSize()
		if shardSize != 0 {
			if AviClientInstance == nil || len(AviClientInstance.AviClient) == 0 {
				// initializing shardSize+1 clients in pool, the +1 is used by CRD ref verification calls
				AviClientInstance, err = utils.NewAviRestClientPool(
					shardSize+1,
					ctrlIpAddress,
					ctrlUsername,
					ctrlPassword,
				)
				connectionStatus = utils.AVIAPI_CONNECTED
				if err != nil {
					connectionStatus = utils.AVIAPI_DISCONNECTED
					utils.AviLog.Error("AVI controller initilization failed")
				}
				// set the tenant and controller version in avisession obj
				for _, client := range AviClientInstance.AviClient {
					SetTenant := session.SetTenant(lib.GetTenant())
					SetTenant(client.AviSession)
					SetVersion := session.SetVersion(utils.CtrlVersion)
					SetVersion(client.AviSession)
				}
			}
		} else {
			connectionStatus = utils.AVIAPI_DISCONNECTED
			err = errors.New("Unable to initialize the Avi controller because the shard vs size is indeterministic")
			utils.AviLog.Error(err)
		}
	}

	models.RestStatus.UpdateAviApiRestStatus(connectionStatus, err)
	return AviClientInstance
}
