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

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

func (l *leader) RestRespArrToObjByType(rest_op *utils.RestOp, obj_type string, key string) []map[string]interface{} {
	var resp_elems []map[string]interface{}
	if restResponse, ok := rest_op.Response.(map[string]interface{}); ok {
		resp_elems = append(resp_elems, restResponse)
		return resp_elems
	}
	return nil
}

func (l *follower) RestRespArrToObjByType(rest_op *utils.RestOp, obj_type string, key string) []map[string]interface{} {
	var resp_elems []map[string]interface{}
	if restResponse, ok := rest_op.Response.(map[string]interface{}); ok {
		if rest_op.Method == utils.RestPost {
			// response has format {count:1, results:[{ /* some data */ }]}
			response, ok := restResponse["results"].([]interface{})
			if !ok {
				return resp_elems
			}
			if len(response) == 0 {
				return resp_elems
			}
			restResponse, ok = response[0].(map[string]interface{})
			if !ok {
				return resp_elems
			}
		}
		utils.AviLog.Debugf("key: %s, msg: Got a response path %v tenant %v response %v",
			key, rest_op.Path, rest_op.Tenant, utils.Stringify(restResponse))
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
