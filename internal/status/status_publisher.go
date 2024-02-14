/*
 * Copyright 2022-2023 VMware, Inc.
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
)

type StatusPublisher interface {
	DequeueStatus(objIntf interface{}) error

	UpdateL4LBStatus(options []UpdateOptions, bulk bool)
	DeleteL4LBStatus(svc_mdata_obj lib.ServiceMetadataObj, vsName, key string) error

	UpdateIngressStatus(options []UpdateOptions, bulk bool)
	DeleteIngressStatus(options []UpdateOptions, isVSDelete bool, key string) error

	UpdateRouteStatus(options []UpdateOptions, bulk bool)
	DeleteRouteStatus(options []UpdateOptions, isVSDelete bool, key string) error

	UpdateRouteIngressStatus(options []UpdateOptions, bulk bool)
	DeleteRouteIngressStatus(options []UpdateOptions, isVSDelete bool, key string) error

	UpdateGatewayStatusAddress(options []UpdateOptions, bulk bool)
	DeleteGatewayStatusAddress(svcMetadataObj lib.ServiceMetadataObj, key string) error

	UpdateSvcApiGatewayStatusAddress(options []UpdateOptions, bulk bool)
	DeleteSvcApiGatewayStatusAddress(key string, svcMetadataObj lib.ServiceMetadataObj) error

	UpdateNPLAnnotation(key, namespace, name string)
	DeleteNPLAnnotation(key, namespace, name string)

	UpdateMultiClusterIngressStatusAndAnnotation(key string, option *UpdateOptions)
	DeleteMultiClusterIngressStatusAndAnnotation(key string, option *UpdateOptions)

	AddStatefulSetAnnotation(statusName string, reason string)
	ResetStatefulSetAnnotation(statusName string)
}

type (
	leader   struct{}
	follower struct{}
)

func NewStatusPublisher() StatusPublisher {
	if lib.AKOControlConfig().IsLeader() {
		return &leader{}
	}
	return &follower{}
}
