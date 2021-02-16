/*
 * Copyright 2020-2021 VMware, Inc.
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

package status

import (
	"context"
	"encoding/json"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// UpdateSvcAnnotation updates a Service with NPL annotation, if not already annotated.
// If the annotation is already pressent return true
func CheckUpdateSvcAnnotation(key, namespace, name string) bool {
	service, err := utils.GetInformers().ServiceInformer.Lister().Services(namespace).Get(name)
	if err != nil {
		return false
	}
	ann := service.GetAnnotations()
	if val, found := ann[lib.NPLSvcAnnotation]; found {
		if val == "true" {
			return true
		}
	}
	if ann == nil {
		ann = make(map[string]string)
	}
	ann[lib.NPLSvcAnnotation] = "true"
	patchPayload := map[string]interface{}{
		"metadata": map[string]map[string]string{
			"annotations": ann,
		},
	}

	payloadBytes, _ := json.Marshal(patchPayload)
	_, err = utils.GetInformers().ClientSet.CoreV1().Services(service.Namespace).Patch(context.TODO(), service.Name, types.MergePatchType, payloadBytes, metav1.PatchOptions{})
	if err != nil {
		utils.AviLog.Errorf("key: %s, msg: there was an error in updating the Service annotation for NPL: %v", key, err)
		return false
	}
	utils.AviLog.Debugf("key: %s, msg: updated NPL annotation from Service: %s/%s", namespace, name)
	return false
}

func DeleteSvcAnnotation(key, namespace, name string) {
	service, err := utils.GetInformers().ServiceInformer.Lister().Services(namespace).Get(name)
	if err != nil {
		return
	}
	ann := service.GetAnnotations()
	if _, found := ann[lib.NPLSvcAnnotation]; !found {
		return
	}
	if ann == nil {
		ann = make(map[string]string)
	}
	delete(ann, lib.NPLSvcAnnotation)
	patchPayload := map[string]interface{}{
		"metadata": map[string]map[string]string{
			"annotations": ann,
		},
	}

	payloadBytes, _ := json.Marshal(patchPayload)
	_, err = utils.GetInformers().ClientSet.CoreV1().Services(service.Namespace).Patch(context.TODO(), service.Name, types.MergePatchType, payloadBytes, metav1.PatchOptions{})
	if err != nil {
		utils.AviLog.Errorf("key: %s, msg: there was an error in updating the Service annotation for NPL: %v", key, err)
		return
	}
	utils.AviLog.Debugf("key: %s, msg: Deleted NPL annotation from Service: %s/%s", namespace, name)
}
