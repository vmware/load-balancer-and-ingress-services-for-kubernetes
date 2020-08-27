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

package nodes

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/avinetworks/ako/internal/lib"
	"github.com/avinetworks/ako/internal/objects"
	"github.com/avinetworks/ako/pkg/utils"

	advl4v1alpha1pre1 "github.com/vmware-tanzu/service-apis/apis/v1alpha1pre1"
)

func (o *AviObjectGraph) BuildAdvancedL4Graph(namespace string, gatewayName string, key string) {
	o.Lock.Lock()
	defer o.Lock.Unlock()
	var VsNode *AviVsNode
	// TODO: work around gateway object and fetch services
	gateway, err := lib.GetAdvL4Informers().GatewayInformer.Lister().Gateways(namespace).Get(gatewayName)
	if err != nil {
		utils.AviLog.Warnf("key: %s, msg: error in obtaining networking.x-k8s.io/gateway: %s", key, gatewayName)
		return
	}

	found, services := objects.ServiceGWLister().GetGwToSvcs(namespace + "/" + gatewayName)
	if !found {
		utils.AviLog.Warnf("key: %s, msg: No services mapped to gateway %s/%s", key, namespace, gatewayName)
		return
	}
	utils.AviLog.Infof("key: %s, msg: Found Services %v for Gateway %s/%s", key, services, gateway.Namespace, gateway.Name)

	// for gatewayName, fetch all valid services for gw.listener
	VsNode = o.ConstructAdvL4VsNode(gatewayName, namespace, key)
	utils.AviLog.Infof("key: %s, msg: Created VSNode with gateway %v", key, utils.Stringify(VsNode))
	o.ConstructAdvL4PolPoolNodes(VsNode, gatewayName, namespace, key)
	o.AddModelNode(VsNode)
	VsNode.CalculateCheckSum()
	o.GraphChecksum = o.GraphChecksum + VsNode.GetCheckSum()
	utils.AviLog.Infof("key: %s, msg: checksum  for AVI VS object %v", key, VsNode.GetCheckSum())
	utils.AviLog.Infof("key: %s, msg: computed Graph checksum for VS is: %v", key, o.GraphChecksum)
}

func (o *AviObjectGraph) ConstructAdvL4VsNode(gatewayName, namespace, key string) *AviVsNode {
	// The logic: Each listener in the gateway is a listener port on the Avi VS.
	// A L4 policyset object is create where listener port --> pool. Pool gets it's server from the endpoints that has the same name as the 'service' pointed
	// by the listener port.
	found, listeners := objects.ServiceGWLister().GetGWListeners(gatewayName)
	if found {
		vsName := lib.GetL4VSName(gatewayName, namespace)
		// TODO: Add service metadata as all services associated with the VS.
		avi_vs_meta := &AviVsNode{Name: vsName, Tenant: lib.GetTenant(),
			EastWest: false}

		if lib.GetSEGName() != lib.DEFAULT_GROUP {
			avi_vs_meta.ServiceEngineGroup = lib.GetSEGName()
		}
		vrfcontext := lib.GetVrf()
		avi_vs_meta.VrfContext = vrfcontext

		isTCP := false
		var portProtocols []AviPortHostProtocol
		for _, listener := range listeners {
			// Listener format: protocol/port
			portProto := strings.Split(listener, "/")
			port, _ := strconv.Atoi(portProto[1])
			pp := AviPortHostProtocol{Port: int32(port), Protocol: fmt.Sprint(portProto[0])}
			portProtocols = append(portProtocols, pp)
			if portProto[0] == "" || portProto[1] == utils.TCP {
				isTCP = true
			}
		}
		avi_vs_meta.PortProto = portProtocols
		// Default case.
		avi_vs_meta.ApplicationProfile = utils.DEFAULT_L4_APP_PROFILE
		if !isTCP {
			avi_vs_meta.NetworkProfile = utils.SYSTEM_UDP_FAST_PATH
		} else {
			avi_vs_meta.NetworkProfile = utils.DEFAULT_TCP_NW_PROFILE
		}
		vsVipName := lib.GetL4VSVipName(gatewayName, namespace)
		vsVipNode := &AviVSVIPNode{Name: vsVipName, Tenant: lib.GetTenant(),
			EastWest: false, VrfContext: vrfcontext}
		avi_vs_meta.VSVIPRefs = append(avi_vs_meta.VSVIPRefs, vsVipNode)
		utils.AviLog.Infof("key: %s, msg: created vs object: %s", key, utils.Stringify(avi_vs_meta))
		return avi_vs_meta
	}
	return nil
}

func (o *AviObjectGraph) ConstructAdvL4PolPoolNodes(vsNode *AviVsNode, gwName, namespace, key string) {
	var l4Policies []*AviL4PolicyNode
	_, listeners := objects.ServiceGWLister().GetGWListeners(gwName)
	for _, listener := range listeners {
		// Listener format: protocol/port
		portProto := strings.Split(listener, "/")
		port, _ := strconv.Atoi(portProto[1])
		// TODO: Pool names to be created using the service name instead of VsName
		poolNode := &AviPoolNode{Name: lib.GetL4PoolName(vsNode.Name, int32(port)), Tenant: lib.GetTenant(), Protocol: portProto[1], PortName: ""}
		poolNode.VrfContext = lib.GetVrf()
		// TODO: Here the populate servers should take the service Name for the listener
		if servers := PopulateServers(poolNode, namespace, gwName, false, key); servers != nil {
			poolNode.Servers = servers
		}

		pool_ref := fmt.Sprintf("/api/pool?name=%s", poolNode.Name)
		var portPoolSet []AviHostPathPortPoolPG
		portPool := AviHostPathPortPoolPG{Port: uint32(port), Pool: pool_ref, Protocol: portProto[1]}
		portPoolSet = append(portPoolSet, portPool)
		l4policyNode := &AviL4PolicyNode{Name: lib.GetL4PolicyName(vsNode.Name, int32(port)), Tenant: lib.GetTenant(), PortPool: portPoolSet}

		vsNode.PoolRefs = append(vsNode.PoolRefs, poolNode)
		utils.AviLog.Infof("key: %s, msg: evaluated L4 pool values :%v", key, utils.Stringify(poolNode))

		l4Policies = append(l4Policies, l4policyNode)
		poolNode.CalculateCheckSum()
		l4policyNode.CalculateCheckSum()
		o.AddModelNode(poolNode)
		o.GraphChecksum = o.GraphChecksum + l4policyNode.GetCheckSum()
		o.GraphChecksum = o.GraphChecksum + poolNode.GetCheckSum()
	}
	vsNode.L4PolicyRefs = l4Policies
	utils.AviLog.Infof("key: %s, msg: evaluated L4 pool policies :%v", key, utils.Stringify(vsNode.L4PolicyRefs))

}

func validateGatewayObj(key string, gateway *advl4v1alpha1pre1.Gateway) error {
	gwClassObj, err := lib.GetAdvL4Informers().GatewayClassInformer.Lister().Get(gateway.Spec.Class)
	if err != nil {
		utils.AviLog.Errorf("key: %s, msg: Unable to fetch corresponding networking.x-k8s.io/gatewayclass %s %v",
			key, gateway.Spec.Class, err)
		return err
	}
	// Additional check to see if the gatewayclass is a valid avi gateway class or not.
	if gwClassObj.Spec.Controller != lib.AviGatewayController {
		// Return an error since this is not our object.
		return errors.New("Unexpected controller")
	}
	return nil
}
