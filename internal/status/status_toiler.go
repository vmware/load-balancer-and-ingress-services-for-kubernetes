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
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

type StatusOptions struct {
	ObjType   string
	Op        string
	IsVSDel   bool
	ObjName   string
	Namespace string
	Key       string
	Options   *UpdateOptions
}

func PublishToStatusQueue(key string, statusOption StatusOptions) {
	statusQueue := utils.SharedWorkQueue().GetQueueByName(utils.StatusQueue)
	bkt := utils.Bkt(key, statusQueue.NumWorkers)
	statusQueue.Workqueue[bkt].AddRateLimited(statusOption)
}

func DequeueStatus(objIntf interface{}) error {
	obj, ok := objIntf.(StatusOptions)
	if !ok {
		utils.AviLog.Warnf("key: %s, object is not of type StatusOptions, %T", obj.Options.Key, objIntf)
		return nil
	}
	switch obj.ObjType {
	case utils.L4LBService:
		if obj.Op == lib.UpdateStatus {
			UpdateL4LBStatus([]UpdateOptions{*obj.Options}, false)
		} else if obj.Op == lib.DeleteStatus {
			DeleteL4LBStatus(obj.Options.ServiceMetadata, "", obj.Options.Key)
		}
	case utils.Ingress:
		if obj.Op == lib.UpdateStatus {
			UpdateIngressStatus([]UpdateOptions{*obj.Options}, false)
		} else if obj.Op == lib.DeleteStatus {
			DeleteIngressStatus([]UpdateOptions{*obj.Options}, obj.IsVSDel, obj.Options.Key)
		}
	case utils.OshiftRoute:
		if obj.Op == lib.UpdateStatus {
			UpdateRouteStatus([]UpdateOptions{*obj.Options}, false)
		} else if obj.Op == lib.DeleteStatus {
			DeleteRouteStatus([]UpdateOptions{*obj.Options}, obj.IsVSDel, obj.Options.Key)
		}
	case lib.Gateway:
		if obj.Op == lib.UpdateStatus {
			UpdateGatewayStatusAddress([]UpdateOptions{*obj.Options}, false)
		} else if obj.Op == lib.DeleteStatus {
			DeleteGatewayStatusAddress(obj.Options.ServiceMetadata, "")
		}
	case lib.SERVICES_API:
		if obj.Op == lib.UpdateStatus {
			UpdateSvcApiGatewayStatusAddress([]UpdateOptions{*obj.Options}, false)
		} else if obj.Op == lib.DeleteStatus {
			DeleteSvcApiGatewayStatusAddress(obj.Options.Key, obj.Options.ServiceMetadata)
		}
	case lib.NPLService:
		if obj.Op == lib.UpdateStatus {
			UpdateNPLAnnotation(obj.Key, obj.Namespace, obj.ObjName)
		} else if obj.Op == lib.DeleteStatus {
			DeleteNPLAnnotation(obj.Key, obj.Namespace, obj.ObjName)
		}
	case lib.MultiClusterIngress:
		if obj.Op == lib.UpdateStatus {
			UpdateMultiClusterIngressStatusAndAnnotation(obj.Key, obj.Options)
		} else if obj.Op == lib.DeleteStatus {
			DeleteMultiClusterIngressStatusAndAnnotation(obj.Key, obj.Options)
		}
	}
	return nil
}
