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
	"errors"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	akogatewayapilib "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/status"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

type StatusUpdater interface {
	Update(key string, option status.StatusOptions)
	BulkUpdate(key string, options []status.StatusOptions)
	Patch(key string, obj runtime.Object, status *Status, retryNum ...int) (runtime.Object, error)
	Delete(key string, option status.StatusOptions)
}

type Status struct {
	*gatewayv1.GatewayClassStatus
	*gatewayv1.GatewayStatus
	*gatewayv1.HTTPRouteStatus
}

func New(ObjectType string) StatusUpdater {
	switch ObjectType {
	case lib.GatewayClass:
		return &gatewayClass{}
	case lib.Gateway:
		return &gateway{}
	case lib.HTTPRoute:
		return &httproute{}
	}
	return nil
}

func DequeueStatus(objIntf interface{}) error {
	option, ok := objIntf.(status.StatusOptions)
	if !ok {
		utils.AviLog.Warnf("Object is not of type StatusOptions, %T", objIntf)
		return nil
	}
	utils.AviLog.Infof("key: %s, msg: starting status Sync", option.Key)
	obj := New(option.ObjType)
	if obj == nil {
		utils.AviLog.Debugf("key: %s, msg: unknown object received", option.Key)
		return nil
	}
	if option.Op == lib.UpdateStatus {
		obj.Update(option.Key, option)
	} else if option.Op == lib.DeleteStatus {
		obj.Delete(option.Key, option)
	}
	return nil
}

func BulkUpdate(key string, objectType string, options []status.StatusOptions) error {
	obj := New(objectType)
	utils.AviLog.Debugf("key: %s, msg: Bulk update is in-progress for object %s", key, objectType)
	obj.BulkUpdate(key, options)
	utils.AviLog.Debugf("key: %s, msg: Bulk update successful for object %s", key, objectType)
	return nil
}

func Record(key string, obj runtime.Object, status *Status) (runtime.Object, error) {
	var objectType string
	switch obj.(type) {
	case *gatewayv1.GatewayClass:
		objectType = lib.GatewayClass
	case *gatewayv1.Gateway:
		objectType = lib.Gateway
	case *gatewayv1.HTTPRoute:
		objectType = lib.HTTPRoute
	default:
		utils.AviLog.Warnf("key %s, msg: Unsupported object received at the status layer, %T", key, obj)
		return obj, errors.New("Unsupported object received at the status layer")
	}
	o := New(objectType)
	updatedGateway, err := o.Patch(key, obj, status)
	if err != nil {
		akogatewayapilib.AKOControlConfig().EventRecorder().Eventf(obj, corev1.EventTypeWarning,
			lib.PatchFailed, "Patch of status failed after multiple retries")
	}
	return updatedGateway, err
}
