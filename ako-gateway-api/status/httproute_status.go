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
	"reflect"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	gatewayv1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	akogatewayapilib "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/status"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

type httproute struct{}

func (o *httproute) Get(key string, name string, namespace string) *gatewayv1beta1.HTTPRoute {

	obj, err := akogatewayapilib.AKOControlConfig().GatewayApiInformers().HTTPRouteInformer.Lister().HTTPRoutes(namespace).Get(name)
	if err != nil {
		utils.AviLog.Warnf("key: %s, msg: unable to get the HTTPRoute object. err: %s", key, err)
		return nil
	}
	utils.AviLog.Debugf("key: %s, msg: Successfully retrieved the HTTPRoute object %s", key, name)
	return obj.DeepCopy()
}

func (o *httproute) GetAll(key string) map[string]*gatewayv1beta1.HTTPRoute {

	objs, err := akogatewayapilib.AKOControlConfig().GatewayApiInformers().HTTPRouteInformer.Lister().List(labels.Everything())
	if err != nil {
		utils.AviLog.Warnf("key: %s, msg: unable to get the HTTPRoute objects. err: %s", key, err)
		return nil
	}

	httpRouteMap := make(map[string]*gatewayv1beta1.HTTPRoute)
	for _, obj := range objs {
		httpRouteMap[obj.Namespace+"/"+obj.Name] = obj.DeepCopy()
	}

	utils.AviLog.Debugf("key: %s, msg: Successfully retrieved the HTTPRoute objects", key)
	return httpRouteMap
}

func (o *httproute) Delete(key string, option status.StatusOptions) {
	// TODO: Add this code when we publish the status from the rest layer
}

func (o *httproute) Update(key string, option status.StatusOptions) {
	// TODO: Add this code when we publish the status from the rest layer
}

func (o *httproute) BulkUpdate(key string, options []status.StatusOptions) {
	// TODO: Add this code when we publish the status from the rest layer
}

func (o *httproute) Patch(key string, obj runtime.Object, status *Status, retryNum ...int) {
	retry := 0
	if len(retryNum) > 0 {
		retry = retryNum[0]
		if retry >= 5 {
			utils.AviLog.Errorf("key: %s, msg: Patch retried 5 times, aborting", key)
			return
		}
	}

	httpRoute := obj.(*gatewayv1beta1.HTTPRoute)
	if o.isStatusEqual(&httpRoute.Status, &status.HTTPRouteStatus) {
		return
	}

	patchPayload, _ := json.Marshal(map[string]interface{}{
		"status": status.HTTPRouteStatus,
	})
	_, err := akogatewayapilib.AKOControlConfig().GatewayAPIClientset().GatewayV1beta1().HTTPRoutes(httpRoute.Namespace).Patch(context.TODO(), httpRoute.Name, types.MergePatchType, patchPayload, metav1.PatchOptions{}, "status")
	if err != nil {
		utils.AviLog.Warnf("key: %s, msg: there was an error in updating the HTTPRoute status. err: %+v, retry: %d", key, err, retry)
		updatedObj, err := akogatewayapilib.AKOControlConfig().GatewayApiInformers().HTTPRouteInformer.Lister().HTTPRoutes(httpRoute.Namespace).Get(httpRoute.Name)
		if err != nil {
			utils.AviLog.Warnf("HTTPRoute not found %v", err)
			return
		}
		o.Patch(key, updatedObj, status, retry+1)
		return
	}

	utils.AviLog.Infof("key: %s, msg: Successfully updated the HTTPRoute %s/%s status %+v", key, httpRoute.Namespace, httpRoute.Name, utils.Stringify(status))
}

func (o *httproute) isStatusEqual(old, new *gatewayv1beta1.HTTPRouteStatus) bool {
	oldStatus, newStatus := old.DeepCopy(), new.DeepCopy()
	currentTime := metav1.Now()
	for i := range oldStatus.Parents {
		for j := range oldStatus.Parents[i].Conditions {
			oldStatus.Parents[i].Conditions[j].LastTransitionTime = currentTime
		}
	}
	for i := range newStatus.Parents {
		for j := range newStatus.Parents[i].Conditions {
			newStatus.Parents[i].Conditions[j].LastTransitionTime = currentTime
		}
	}
	return reflect.DeepEqual(oldStatus, newStatus)
}
