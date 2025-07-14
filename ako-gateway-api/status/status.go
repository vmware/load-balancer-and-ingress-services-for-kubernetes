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

package status

import (
	"k8s.io/apimachinery/pkg/runtime"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	akogatewayapiobjects "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/objects"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/status"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

type StatusUpdater interface {
	Update(key string, option status.StatusOptions)
	BulkUpdate(key string, options []status.StatusOptions)
	Patch(key string, obj runtime.Object, status *status.Status, retryNum ...int) error
	Delete(key string, option status.StatusOptions)
}

func New(ObjectType string) StatusUpdater {
	switch ObjectType {
	case lib.GatewayClass:
		return &gatewayClass{}
	case lib.Gateway:
		return &gateway{}
	case lib.HTTPRoute:
		return &httproute{}
	case lib.NPLService:
		return &nplservice{publisher: status.NewStatusPublisher()}
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
	if option.Options != nil && option.Options.ServiceMetadata.HTTPRoute != "" && option.Options.Status == nil {
		utils.AviLog.Debugf("key: %s, msg: Status update for ChildVs received", option.Options.ServiceMetadata.HTTPRoute)
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
func Record(key string, obj runtime.Object, objStatus *status.Status) {
	var objectType string
	var statusOption status.StatusOptions
	var updateOption status.UpdateOptions
	var serviceMetadata lib.ServiceMetadataObj

	switch gwObject := obj.(type) {
	case *gatewayv1.GatewayClass:
		objectType = lib.GatewayClass
		o := New(objectType)
		o.Patch(key, obj, objStatus)
		return
	case *gatewayv1.Gateway:
		objectType = lib.Gateway
		serviceMetadata.Gateway = gwObject.Namespace + "/" + gwObject.Name
		key = serviceMetadata.Gateway
		akogatewayapiobjects.GatewayApiLister().UpdateGatewayToGatewayStatusMapping(serviceMetadata.Gateway, objStatus.GatewayStatus)
	case *gatewayv1.HTTPRoute:
		objectType = lib.HTTPRoute
		serviceMetadata.HTTPRoute = gwObject.Namespace + "/" + gwObject.Name
		key = serviceMetadata.HTTPRoute
		akogatewayapiobjects.GatewayApiLister().UpdateRouteToRouteStatusMapping(objectType+"/"+serviceMetadata.HTTPRoute, objStatus.HTTPRouteStatus)
	default:
		utils.AviLog.Warnf("key %s, msg: Unsupported object received at the status layer, %T", key, obj)
		return
	}
	updateOption.Status = objStatus
	updateOption.ServiceMetadata = serviceMetadata
	statusOption.Options = &updateOption
	statusOption.Op = lib.UpdateStatus
	statusOption.ObjType = objectType
	statusOption.Key = key
	status.PublishToStatusQueue(key, statusOption)
}
