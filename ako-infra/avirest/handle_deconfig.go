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

package avirest

import (
	"strings"

	"github.com/vmware/alb-sdk/go/models"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

func DeleteServiceEngines(nextURI ...string) error {
	client := InfraAviClientInstance()
	segName := lib.GetClusterID()
	cloudName := utils.CloudName

	seListUri := "/api/serviceengine/?page_size=100&include_name&se_group_ref.name=" + segName + "&cloud_ref.name=" + cloudName
	if len(nextURI) > 0 {
		seListUri = nextURI[0]
	}

	response := models.ServiceEngineAPIResponse{}
	err := lib.AviGet(client, seListUri, &response)
	if err != nil {
		utils.AviLog.Warnf("Error during Get call for the SEs: %v", err.Error())
		return err
	}

	if len(response.Results) == 0 {
		utils.AviLog.Infof("No service engines found for service engine group %s in cloud %s", segName, cloudName)
		return nil
	}

	seUUIDs := []string{}
	for _, se := range response.Results {
		seUUIDs = append(seUUIDs, *se.UUID)
	}

	for _, seUUID := range seUUIDs {
		deleteUri := "/api/serviceengine/" + seUUID
		utils.AviLog.Infof("Deleting Service Engine %s", seUUID)
		if err := lib.AviDelete(client, deleteUri); err != nil {
			utils.AviLog.Errorf("Error during Delete call for the SE: %s, %s", seUUID, err.Error())
			return err
		}
	}

	if response.Next != nil {
		// The GET call response had a next page, let's recursively call the same method.
		nextUri := strings.Split(*response.Next, "/api/serviceengine")
		if len(nextUri) > 1 {
			overrideUri := "/api/serviceengine" + nextUri[1]
			if err := DeleteServiceEngines(overrideUri); err != nil {
				return err
			}
		}
	}
	return nil
}

func DeleteServiceEngineGroup() error {
	client := InfraAviClientInstance()
	segName := lib.GetClusterID()
	cloudName := utils.CloudName

	segListUri := "/api/serviceenginegroup/?name=" + segName + "&cloud_ref.name=" + cloudName
	response := models.ServiceEngineGroupAPIResponse{}
	err := lib.AviGet(client, segListUri, &response)
	if err != nil {
		utils.AviLog.Warnf("Error during Get call for the SE group :%v", err.Error())
		return err
	}

	if len(response.Results) == 0 {
		utils.AviLog.Infof("No service engine groups found in cloud %s", cloudName)
		return nil
	}

	segUuid := *response.Results[0].UUID
	deleteSEGUri := "/api/serviceenginegroup/" + segUuid
	utils.AviLog.Infof("Deleting Service Engine Group %s", segUuid)
	if err := lib.AviDelete(client, deleteSEGUri); err != nil {
		utils.AviLog.Errorf("Error during Delete call for the SE: %s, %s", segUuid, err.Error())
		return err
	}

	return nil
}
