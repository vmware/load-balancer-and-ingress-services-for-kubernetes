/*
 * Copyright 2019-2020 VMware, Inc.
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

package k8s

import (
	v1 "k8s.io/api/core/v1"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

func (c *AviController) OnStartedLeading(readyCh chan struct{}) {
	lib.AKOControlConfig().SetIsLeaderFlag(true)
	utils.AviLog.Debugf("AKO became a leader")
	lib.AKOControlConfig().PodEventf(v1.EventTypeNormal, "LeaderElection", "AKO became a leader")
	if isReady(readyCh) {
		c.publishAllVSKeysToRestLayer()
	}
}

func (c *AviController) OnNewLeader(readyCh chan struct{}) {
	lib.AKOControlConfig().SetIsLeaderFlag(false)
	utils.AviLog.Debugf("AKO became a follower")
	lib.AKOControlConfig().PodEventf(v1.EventTypeNormal, "LeaderElection", "AKO became a follower")
	if isReady(readyCh) {
		c.publishAllVSKeysToRestLayer()
	}
}

func (c *AviController) OnStoppedLeading(readyCh chan struct{}) {
	lib.AKOControlConfig().SetIsLeaderFlag(false)
	utils.AviLog.Debugf("AKO lost the leadership, rebooting the AKO")
	lib.AKOControlConfig().PodEventf(v1.EventTypeNormal, "LeaderElection", "AKO lost the leadership")
}

func isReady(readyCh chan struct{}) bool {
	// ok returns false if the ready channel is closed.
	_, ok := <-readyCh
	return !ok
}
