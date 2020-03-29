/*
 * [2013] - [2018] Avi Networks Incorporated
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
	"sync"

	"ako/pkg/lib"

	"github.com/avinetworks/container-lib/utils"
)

var AviClientInstance *utils.AviRestClientPool
var clientonce sync.Once

// This class is in control of AKC. It uses utils from the common project.
func SharedAVIClients() *utils.AviRestClientPool {
	var err error
	ctrlUsername := os.Getenv("CTRL_USERNAME")
	ctrlPassword := os.Getenv("CTRL_PASSWORD")
	ctrlIpAddress := os.Getenv("CTRL_IPADDRESS")
	if ctrlUsername == "" || ctrlPassword == "" || ctrlIpAddress == "" {
		utils.AviLog.Error.Panic("AVI controller information missing. Update them in kubernetes secret or via environment variables.")
	}

	if AviClientInstance == nil || len(AviClientInstance.AviClient) == 0 {
		shardSize := lib.GetshardSize()
		if shardSize != 0 {
			if AviClientInstance == nil || len(AviClientInstance.AviClient) == 0 {
				AviClientInstance, err = utils.NewAviRestClientPool(shardSize,
					ctrlIpAddress, ctrlUsername, ctrlPassword)
				if err != nil {
					utils.AviLog.Error.Print("AVI controller initilization failed")
				}
			}
		} else {
			utils.AviLog.Error.Print("Unable to initialize the Avi controller because the shard vs size is indeterministic")
		}
	}
	return AviClientInstance
}
