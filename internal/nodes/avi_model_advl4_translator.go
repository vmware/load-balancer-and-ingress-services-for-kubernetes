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

	avicache "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	advl4v1alpha1pre1 "github.com/vmware-tanzu/service-apis/apis/v1alpha1pre1"
)

func (o *AviObjectGraph) BuildAdvancedL4Graph(namespace string, gatewayName string, key string) {
	o.Lock.Lock()
	defer o.Lock.Unlock()
	if VsNode := o.ConstructAdvL4VsNode(gatewayName, namespace, key); VsNode != nil {
		o.ConstructAdvL4PolPoolNodes(VsNode, gatewayName, namespace, key)
		o.AddModelNode(VsNode)
		VsNode.CalculateCheckSum()
		o.GraphChecksum = o.GraphChecksum + VsNode.GetCheckSum()
		utils.AviLog.Infof("key: %s, msg: checksum  for AVI VS object %v", key, VsNode.GetCheckSum())
		utils.AviLog.Infof("key: %s, msg: computed Graph checksum for VS is: %v", key, o.GraphChecksum)
	}
}

func (o *AviObjectGraph) ConstructAdvL4VsNode(gatewayName, namespace, key string) *AviVsNode {
	// The logic: Each listener in the gateway is a listener port on the Avi VS.
	// A L4 policyset object is create where listener port --> pool. Pool gets it's server from the endpoints that has the same name as the 'service' pointed
	// by the listener port.
	found, listeners := objects.ServiceGWLister().GetGWListeners(namespace + "/" + gatewayName)
	if found {
		vsName := lib.GetL4VSName(gatewayName, namespace)

		var serviceNSNames []string
		if found, services := objects.ServiceGWLister().GetGwToSvcs(namespace + "/" + gatewayName); found {
			for svcListener, service := range services {
				if utils.HasElem(listeners, svcListener) && !utils.HasElem(serviceNSNames, service) {
					serviceNSNames = append(serviceNSNames, service)
				}
			}
		}

		avi_vs_meta := &AviVsNode{
			Name:       vsName,
			Tenant:     lib.GetTenant(),
			EastWest:   false,
			VrfContext: lib.GetVrf(),
			ServiceMetadata: avicache.ServiceMetadataObj{
				NamespaceServiceName: serviceNSNames,
				Gateway:              namespace + "/" + gatewayName,
			},
		}

		if lib.GetSEGName() != lib.DEFAULT_GROUP {
			avi_vs_meta.ServiceEngineGroup = lib.GetSEGName()
		}

		isTCP := false
		var portProtocols []AviPortHostProtocol
		for _, listener := range listeners {
			portProto := strings.Split(listener, "/") // format: protocol/port
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

		vsVipNode := &AviVSVIPNode{
			Name:       lib.GetL4VSVipName(gatewayName, namespace),
			Tenant:     lib.GetTenant(),
			EastWest:   false,
			VrfContext: lib.GetVrf(),
		}
		avi_vs_meta.VSVIPRefs = append(avi_vs_meta.VSVIPRefs, vsVipNode)
		utils.AviLog.Infof("key: %s, msg: created vs object: %s", key, utils.Stringify(avi_vs_meta))
		return avi_vs_meta
	}
	return nil
}

func (o *AviObjectGraph) ConstructAdvL4PolPoolNodes(vsNode *AviVsNode, gwName, namespace, key string) {
	var l4Policies []*AviL4PolicyNode
	found, svcListeners := objects.ServiceGWLister().GetGwToSvcs(namespace + "/" + gwName)
	foundGW, gwListeners := objects.ServiceGWLister().GetGWListeners(namespace + "/" + gwName)
	if !found || !foundGW {
		return
	}
	var portPoolSet []AviHostPathPortPoolPG
	for listener, svc := range svcListeners {
		if !utils.HasElem(gwListeners, listener) {
			continue
		}
		portProto := strings.Split(listener, "/") // format: protocol/port
		svcNSName := strings.Split(svc, "/")
		port, _ := strconv.Atoi(portProto[1])

		poolNode := &AviPoolNode{
			Name:       lib.GetAdvL4PoolName(svcNSName[1], namespace, int32(port)),
			Tenant:     lib.GetTenant(),
			Protocol:   portProto[1],
			PortName:   "",
			VrfContext: lib.GetVrf(),
		}

		// If the service has multiple ports but the gateway specifies one of them as listeners then we pick the portname from the service and populate it in pool portname.
		svcObj, err := utils.GetInformers().ServiceInformer.Lister().Services(svcNSName[0]).Get(svcNSName[1])
		if err != nil {
			utils.AviLog.Warnf("key: %s, msg: error while retrieving service: %s", key, err)
			return
		}
		// Obtain the matching portname from the svcObj
		for _, svcPort := range svcObj.Spec.Ports {
			if svcPort.Port == int32(port) {
				poolNode.PortName = svcPort.Name
			}
		}

		if servers := PopulateServers(poolNode, svcNSName[0], svcNSName[1], false, key); servers != nil {
			poolNode.Servers = servers
		}

		pool_ref := fmt.Sprintf("/api/pool?name=%s", poolNode.Name)
		portPool := AviHostPathPortPoolPG{
			Port:     uint32(port),
			Pool:     pool_ref,
			Protocol: portProto[0],
		}
		portPoolSet = append(portPoolSet, portPool)
		vsNode.PoolRefs = append(vsNode.PoolRefs, poolNode)
		utils.AviLog.Infof("key: %s, msg: evaluated L4 pool values :%v", key, utils.Stringify(poolNode))
		poolNode.CalculateCheckSum()
		o.AddModelNode(poolNode)
		o.GraphChecksum = o.GraphChecksum + poolNode.GetCheckSum()
	}
	l4policyNode := &AviL4PolicyNode{
		Name:     vsNode.Name,
		Tenant:   lib.GetTenant(),
		PortPool: portPoolSet,
	}
	l4Policies = append(l4Policies, l4policyNode)
	l4policyNode.CalculateCheckSum()
	o.GraphChecksum = o.GraphChecksum + l4policyNode.GetCheckSum()
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

	for _, listener := range gateway.Spec.Listeners {
		gwName, nameOk := listener.Routes.RouteSelector.MatchLabels[lib.GatewayNameLabelKey]
		gwNamespace, nsOk := listener.Routes.RouteSelector.MatchLabels[lib.GatewayNamespaceLabelKey]
		if !nameOk || !nsOk ||
			(nameOk && gwName != gateway.Name) ||
			(nsOk && gwNamespace != gateway.Namespace) {
			return errors.New("Incorrect gateway matchLabels configuration")
		}
	}

	// Additional check to see if the gatewayclass is a valid avi gateway class or not.
	if gwClassObj.Spec.Controller != lib.AviGatewayController {
		// Return an error since this is not our object.
		return errors.New("Unexpected controller")
	}

	return nil
}
