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

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func CheckNPLSvcAnnotation(key, namespace, name string) bool {
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
	return false
}

// UpdateSvcAnnotation updates a Service with NPL annotation, if not already annotated.
// If the annotation is already pressent return true
func (l *leader) UpdateNPLAnnotation(key, namespace, name string) {
	service, err := utils.GetInformers().ServiceInformer.Lister().Services(namespace).Get(name)
	if err != nil {
		utils.AviLog.Infof("key: %s, returning without updating NPL annotation, err %v", key, err)
		return
	}
	if service.Spec.Type == corev1.ServiceTypeNodePort {
		utils.AviLog.Infof("key: %s, returning without updating NPL annotation for Service type NodePort", key)
		return
	}
	ann := service.GetAnnotations()
	if val, found := ann[lib.NPLSvcAnnotation]; found {
		if val == "true" {
			return
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
		return
	}
	utils.AviLog.Infof("key: %s, msg: updated NPL annotation for Service: %s/%s", key, namespace, name)
	return
}

func (l *leader) DeleteNPLAnnotation(key, namespace, name string) {
	service, err := utils.GetInformers().ServiceInformer.Lister().Services(namespace).Get(name)
	if err != nil {
		return
	}
	ann := service.GetAnnotations()
	if ann == nil {
		return
	}
	if _, found := ann[lib.NPLSvcAnnotation]; !found {
		return
	}

	payloadValue := make(map[string]*string)
	// To delete an annotation with patch call, the value has to be set to nil
	payloadValue[lib.NPLSvcAnnotation] = nil

	newPayload := map[string]interface{}{
		"metadata": map[string]map[string]*string{
			"annotations": payloadValue,
		},
	}

	payloadBytes, _ := json.Marshal(newPayload)
	_, err = utils.GetInformers().ClientSet.CoreV1().Services(service.Namespace).Patch(context.TODO(), service.Name, types.MergePatchType, payloadBytes, metav1.PatchOptions{})
	if err != nil {
		utils.AviLog.Errorf("key: %s, msg: there was an error in updating the Service annotation for NPL: %v", key, err)
		return
	}
	utils.AviLog.Infof("key: %s, msg: Deleted NPL annotation from Service: %s/%s", key, namespace, name)
}

func (f *follower) UpdateNPLAnnotation(key, namespace, name string) {
	utils.AviLog.Debugf("key: %s, AKO is not a leader, not updating the NPL Annotation", key)
}

func (f *follower) DeleteNPLAnnotation(key, namespace, name string) {
	utils.AviLog.Debugf("key: %s, AKO is not a leader, not deleting the NPL Annotation", key)
}
