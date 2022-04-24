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

package rest

import (
	"regexp"

	"github.com/vmware/alb-sdk/go/models"
	avicache "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

func RestRespArrToObjByType(rest_op *utils.RestOp, obj_type string, key string) []map[string]interface{} {
	var resp_elems []map[string]interface{}
	if restResponse, ok := rest_op.Response.(map[string]interface{}); ok {
		resp_elems = append(resp_elems, restResponse)
		return resp_elems
	}
	return nil
}

func ExtractVsName(word string) string {
	r, _ := regexp.Compile("#.*")
	result := r.FindAllString(word, -1)
	if len(result) == 1 {
		return result[0][1:]
	}
	return ""
}

func GetLicenseTypeFromURI() (models.SystemConfiguration, error) {
	uri := "/api/systemconfiguration"
	response := models.SystemConfiguration{}
	client := avicache.SharedAVIClients()
	err := lib.AviGet(client.AviClient[0], uri, &response)

	if err != nil {
		utils.AviLog.Warnf("Unable to fetch system configuration, error %s", err.Error())
	}

	return response, err
}
