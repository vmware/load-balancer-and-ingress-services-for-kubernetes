/*
 * Copyright 2023-2024 VMware, Inc.
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
	"errors"
	"reflect"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	akogatewayapilib "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/status"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

type gatewayClass struct{}

func (o *gatewayClass) Get(key string, name string) *gatewayv1.GatewayClass {

	obj, err := akogatewayapilib.AKOControlConfig().GatewayApiInformers().GatewayClassInformer.Lister().Get(name)
	if err != nil {
		utils.AviLog.Warnf("key: %s, msg: unable to get the GatewayClass object. err: %s", key, err)
		return nil
	}
	utils.AviLog.Debugf("key: %s, msg: Successfully retrieved the GatewayClass object %s", key, name)
	return obj.DeepCopy()
}

func (o *gatewayClass) GetAll(key string) map[string]*gatewayv1.GatewayClass {

	objs, err := akogatewayapilib.AKOControlConfig().GatewayApiInformers().GatewayClassInformer.Lister().List(labels.Everything())
	if err != nil {
		utils.AviLog.Warnf("key: %s, msg: unable to get the GatewayClass objects. err: %s", key, err)
		return nil
	}

	gatewayClassMap := make(map[string]*gatewayv1.GatewayClass)
	for _, obj := range objs {
		gatewayClassMap[obj.Name] = obj.DeepCopy()
	}

	utils.AviLog.Debugf("key: %s, msg: Successfully retrieved the GatewayClass objects", key)
	return gatewayClassMap
}

func (o *gatewayClass) Delete(key string, option status.StatusOptions) {
	// TODO: Add this code when we publish the status from the rest layer
}

func (o *gatewayClass) Update(key string, option status.StatusOptions) {
	// TODO: Add this code when we publish the status from the rest layer
}

func (o *gatewayClass) BulkUpdate(key string, options []status.StatusOptions) {
	// TODO: Add this code when we publish the status from the rest layer
}

func (o *gatewayClass) Patch(key string, obj runtime.Object, status *status.Status, retryNum ...int) error {
	retry := 0
	if len(retryNum) > 0 {
		retry = retryNum[0]
		if retry >= 5 {
			utils.AviLog.Errorf("key: %s, msg: Patch retried 5 times, aborting", key)
			return errors.New("Patch retried 5 times, aborting")
		}
	}

	gatewayClass := obj.(*gatewayv1.GatewayClass)
	if o.isStatusEqual(&gatewayClass.Status, status.GatewayClassStatus) {
		return nil
	}

	patchPayload, _ := json.Marshal(map[string]interface{}{
		"status": status.GatewayClassStatus,
	})
	_, err := akogatewayapilib.AKOControlConfig().GatewayAPIClientset().GatewayV1().GatewayClasses().Patch(context.TODO(), gatewayClass.Name, types.MergePatchType, patchPayload, metav1.PatchOptions{}, "status")
	if err != nil {
		utils.AviLog.Warnf("key: %s, msg: there was an error in updating the GatewayClass status. err: %+v, retry: %d", key, err, retry)
		updatedObj, err := akogatewayapilib.AKOControlConfig().GatewayApiInformers().GatewayClassInformer.Lister().Get(gatewayClass.Name)
		if err != nil {
			utils.AviLog.Warnf("GatewayClass not found %v", err)
			return err
		}
		return o.Patch(key, updatedObj, status, retry+1)
	}
	utils.AviLog.Infof("key: %s, msg: Successfully updated the GatewayClass %s status %+v %v", key, gatewayClass.Name, utils.Stringify(status), err)
	return nil
}

func (o *gatewayClass) isStatusEqual(old, new *gatewayv1.GatewayClassStatus) bool {
	oldStatus, newStatus := old.DeepCopy(), new.DeepCopy()
	currentTime := metav1.Now()
	for i := range oldStatus.Conditions {
		oldStatus.Conditions[i].LastTransitionTime = currentTime
	}
	for i := range newStatus.Conditions {
		newStatus.Conditions[i].LastTransitionTime = currentTime
	}
	return reflect.DeepEqual(oldStatus, newStatus)
}
