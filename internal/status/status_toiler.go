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
	lib.IncrementQueueCounter(utils.StatusQueue)
	statusQueue.Workqueue[bkt].AddRateLimited(statusOption)
}

func (l *leader) DequeueStatus(objIntf interface{}) error {
	obj, ok := objIntf.(StatusOptions)
	lib.DecrementQueueCounter(utils.StatusQueue)
	if !ok {
		utils.AviLog.Warnf("Object is not of type StatusOptions, %T", objIntf)
		return nil
	}
	utils.AviLog.Infof("key: %s, msg: start status layer sync.", obj.Key)
	switch obj.ObjType {
	case utils.L4LBService:
		if obj.Op == lib.UpdateStatus {
			l.UpdateL4LBStatus([]UpdateOptions{*obj.Options}, false)
		} else if obj.Op == lib.DeleteStatus {
			l.DeleteL4LBStatus(obj.Options.ServiceMetadata, "", obj.Options.Key)
		}
	case utils.Ingress:
		if obj.Op == lib.UpdateStatus {
			l.UpdateIngressStatus([]UpdateOptions{*obj.Options}, false)
		} else if obj.Op == lib.DeleteStatus {
			l.DeleteIngressStatus([]UpdateOptions{*obj.Options}, obj.IsVSDel, obj.Options.Key)
		}
	case utils.OshiftRoute:
		if obj.Op == lib.UpdateStatus {
			l.UpdateRouteStatus([]UpdateOptions{*obj.Options}, false)
		} else if obj.Op == lib.DeleteStatus {
			l.DeleteRouteStatus([]UpdateOptions{*obj.Options}, obj.IsVSDel, obj.Options.Key)
		}
	case lib.Gateway:
		if obj.Op == lib.UpdateStatus {
			l.UpdateGatewayStatusAddress([]UpdateOptions{*obj.Options}, false)
		} else if obj.Op == lib.DeleteStatus {
			l.DeleteGatewayStatusAddress(obj.Options.ServiceMetadata, "")
		}
	case lib.SERVICES_API:
		if obj.Op == lib.UpdateStatus {
			l.UpdateSvcApiGatewayStatusAddress([]UpdateOptions{*obj.Options}, false)
		} else if obj.Op == lib.DeleteStatus {
			l.DeleteSvcApiGatewayStatusAddress(obj.Options.Key, obj.Options.ServiceMetadata)
		}
	case lib.NPLService:
		if obj.Op == lib.UpdateStatus {
			l.UpdateNPLAnnotation(obj.Key, obj.Namespace, obj.ObjName)
		} else if obj.Op == lib.DeleteStatus {
			l.DeleteNPLAnnotation(obj.Key, obj.Namespace, obj.ObjName)
		}
	case lib.MultiClusterIngress:
		if obj.Op == lib.UpdateStatus {
			l.UpdateMultiClusterIngressStatusAndAnnotation(obj.Key, obj.Options)
		} else if obj.Op == lib.DeleteStatus {
			l.DeleteMultiClusterIngressStatusAndAnnotation(obj.Key, obj.Options)
		}
	}
	return nil
}

func (f *follower) DequeueStatus(objIntf interface{}) error {
	obj, ok := objIntf.(StatusOptions)
	if !ok {
		utils.AviLog.Warnf("Object is not of type StatusOptions, %T", objIntf)
		return nil
	}
	utils.AviLog.Debugf("key: %s, AKO is not running as a leader", obj.Key)
	return nil
}
