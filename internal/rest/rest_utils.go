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
	"errors"
	"regexp"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

func RestRespArrToObjByType(rest_op *utils.RestOp, obj_type string, key string) ([]map[string]interface{}, error) {
	var resp_elems []map[string]interface{}
	if rest_op.Method == utils.RestPost {
		resp_arr, ok := rest_op.Response.([]interface{})
		if !ok {
			utils.AviLog.Warnf("key: %s, msg: response has unknown type %T", key, rest_op.Response)
			return nil, errors.New("Malformed response")
		}

		for _, resp_elem := range resp_arr {
			resp, ok := resp_elem.(map[string]interface{})
			if !ok {
				utils.AviLog.Warnf("key: %s, msg: response has unknown type %T", key, resp_elem)
				continue
			}

			avi_url, ok := resp["url"].(string)
			if !ok {
				utils.AviLog.Warnf("key: %s, msg:url not present in response %v", key, resp)
				continue
			}

			avi_obj_type, err := utils.AviUrlToObjType(avi_url)
			if err == nil && avi_obj_type == obj_type {
				resp_elems = append(resp_elems, resp)
			}
		}
	} else {
		// The PUT calls are specific for the resource
		resp_elems = append(resp_elems, rest_op.Response.(map[string]interface{}))
	}

	return resp_elems, nil
}

func ExtractVsName(word string) string {
	r, _ := regexp.Compile("#.*")
	result := r.FindAllString(word, -1)
	if len(result) == 1 {
		return result[0][1:]
	}
	return ""
}
