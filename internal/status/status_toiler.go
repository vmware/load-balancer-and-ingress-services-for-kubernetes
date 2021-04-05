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
	ObjType string
	Op      string
	IsVSDel bool
	Options *UpdateOptions
}

func DequeueStatus(objIntf interface{}) error {
	obj, ok := objIntf.(StatusOptions)
	if !ok {
		utils.AviLog.Warnf("key: %s, object is not of type StatusOptions, %T", obj.Options.Key, objIntf)
		return nil
	}
	if obj.ObjType == utils.L4LBService {
		if obj.Op == lib.UpdateOperation {
			UpdateL4LBStatus([]UpdateOptions{*obj.Options}, false)
		} else if obj.Op == lib.DeleteOperation {
			DeleteL4LBStatus(obj.Options.ServiceMetadata, obj.Options.Key)
		}
	} else if obj.ObjType == utils.Ingress {
		if obj.Op == lib.UpdateOperation {
			UpdateIngressStatus([]UpdateOptions{*obj.Options}, false)
		} else if obj.Op == lib.DeleteOperation {
			DeleteIngressStatus([]UpdateOptions{*obj.Options}, obj.IsVSDel, obj.Options.Key)
		}
	} else if obj.ObjType == utils.OshiftRoute {
		if obj.Op == lib.UpdateOperation {
			UpdateRouteStatus([]UpdateOptions{*obj.Options}, false)
		} else if obj.Op == lib.DeleteOperation {
			DeleteRouteStatus([]UpdateOptions{*obj.Options}, obj.IsVSDel, obj.Options.Key)
		}
	} else if obj.ObjType == lib.Gateway {
		if obj.Op == lib.UpdateOperation {
			UpdateGatewayStatusAddress([]UpdateOptions{*obj.Options}, false)
		} else if obj.Op == lib.DeleteOperation {
			DeleteGatewayStatusAddress(obj.Options.ServiceMetadata, "")
		}
	} else if obj.ObjType == lib.SERVICES_API {
		if obj.Op == lib.UpdateOperation {
			UpdateSvcApiGatewayStatusAddress([]UpdateOptions{*obj.Options}, false)
		} else if obj.Op == lib.DeleteOperation {
			DeleteSvcApiGatewayStatusAddress(obj.Options.Key, obj.Options.ServiceMetadata)
		}
	}

	return nil
}
