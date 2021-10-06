/*
 * Copyright 2021 VMware, Inc.
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

// This file is used to create all Avi infra related changes and can be used as a library if required in other places.

package ingestion

import (
	avicache "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

type AviControllerInfra struct{}

func (a *AviControllerInfra) InitInfraController() {
	aviRestClientPool := avicache.SharedAVIClients()
	if aviRestClientPool == nil {
		utils.AviLog.Fatalf("Avi client not initialized during Infra bootup")
	}

	if aviRestClientPool != nil && !avicache.IsAviClusterActive(aviRestClientPool.AviClient[0]) {
		utils.AviLog.Fatalf("Avi Controller Cluster state is not Active, shutting down AKO info container")
	}
}
