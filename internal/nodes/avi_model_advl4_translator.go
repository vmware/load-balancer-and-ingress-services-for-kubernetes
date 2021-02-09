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
	"fmt"
	"strconv"
	"strings"

	svcapiv1alpha1 "sigs.k8s.io/service-apis/apis/v1alpha1"

	utilsnet "k8s.io/utils/net"

	avicache "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	advl4v1alpha1pre1 "github.com/vmware-tanzu/service-apis/apis/v1alpha1pre1"
)

func (o *AviObjectGraph) BuildAdvancedL4Graph(namespace string, gatewayName string, key string) {
	o.Lock.Lock()
	defer o.Lock.Unlock()
	var vsNode *AviVsNode
	if lib.UseServicesAPI() {
		vsNode = o.ConstructSvcApiL4VsNode(gatewayName, namespace, key)
	} else {
		vsNode = o.ConstructAdvL4VsNode(gatewayName, namespace, key)
	}
	if vsNode != nil {
		o.ConstructAdvL4PolPoolNodes(vsNode, gatewayName, namespace, key)
		o.AddModelNode(vsNode)
		vsNode.CalculateCheckSum()
		o.GraphChecksum = o.GraphChecksum + vsNode.GetCheckSum()
		utils.AviLog.Infof("key: %s, msg: checksum  for AVI VS object %v", key, vsNode.GetCheckSum())
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
		gw, _ := lib.GetAdvL4Informers().GatewayInformer.Lister().Gateways(namespace).Get(gatewayName)

		var serviceNSNames []string
		if found, services := objects.ServiceGWLister().GetGwToSvcs(namespace + "/" + gatewayName); found {
			for svcListener, service := range services {
				// assume it to have only a single backend service, the check is in isGatewayDelete
				if utils.HasElem(listeners, svcListener) && len(service) == 1 && !utils.HasElem(serviceNSNames, service[0]) {
					serviceNSNames = append(serviceNSNames, service[0])
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

		if lib.GetSEGName() != lib.DEFAULT_SE_GROUP {
			avi_vs_meta.ServiceEngineGroup = lib.GetSEGName()
		}

		isTCP := false
		var portProtocols []AviPortHostProtocol
		for _, listener := range listeners {
			portProto := strings.Split(listener, "/") // format: protocol/port
			port, _ := utilsnet.ParsePort(portProto[1], true)
			pp := AviPortHostProtocol{Port: int32(port), Protocol: portProto[0]}
			portProtocols = append(portProtocols, pp)
			if portProto[0] == "" || portProto[0] == utils.TCP {
				isTCP = true
			}
		}
		avi_vs_meta.PortProto = portProtocols
		// Default case.
		avi_vs_meta.ApplicationProfile = utils.DEFAULT_L4_APP_PROFILE
		if !isTCP {
			avi_vs_meta.NetworkProfile = utils.SYSTEM_UDP_FAST_PATH
		} else {
			avi_vs_meta.NetworkProfile = utils.TCP_NW_FAST_PATH
		}

		vsVipNode := &AviVSVIPNode{
			Name:       lib.GetL4VSVipName(gatewayName, namespace),
			Tenant:     lib.GetTenant(),
			EastWest:   false,
			VrfContext: lib.GetVrf(),
		}

		if len(gw.Spec.Addresses) > 0 && gw.Spec.Addresses[0].Type == advl4v1alpha1pre1.IPAddressType {
			vsVipNode.IPAddress = gw.Spec.Addresses[0].Value
		}

		avi_vs_meta.VSVIPRefs = append(avi_vs_meta.VSVIPRefs, vsVipNode)
		utils.AviLog.Infof("key: %s, msg: created vs object: %s", key, utils.Stringify(avi_vs_meta))
		return avi_vs_meta
	}
	return nil
}

func (o *AviObjectGraph) ConstructSvcApiL4VsNode(gatewayName, namespace, key string) *AviVsNode {
	// The logic: Each listener in the gateway is a listener port on the Avi VS.
	// A L4 policyset object is create where listener port --> pool. Pool gets it's server from the endpoints that has the same name as the 'service' pointed
	// by the listener port.
	found, listeners := objects.ServiceGWLister().GetGWListeners(namespace + "/" + gatewayName)
	if found {
		vsName := lib.GetL4VSName(gatewayName, namespace)
		gw, _ := lib.GetSvcAPIInformers().GatewayInformer.Lister().Gateways(namespace).Get(gatewayName)

		var serviceNSNames []string
		if found, services := objects.ServiceGWLister().GetGwToSvcs(namespace + "/" + gatewayName); found {
			for svcListener, service := range services {
				// assume it to have only a single backend service, the check is in isGatewayDelete
				if utils.HasElem(listeners, svcListener) && len(service) == 1 && !utils.HasElem(serviceNSNames, service[0]) {
					serviceNSNames = append(serviceNSNames, service[0])
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

		// Set SeGroup configuration either from NsxAlbInfraSetting Spec or
		// env variable (provided during installation)
		gwClass, err := lib.GetSvcAPIInformers().GatewayClassInformer.Lister().Get(gw.Spec.GatewayClassName)
		if err != nil {
			utils.AviLog.Warnf("Unable to get corresponding GatewayClass %s", err.Error())
		} else {
			if gwClass.Spec.ParametersRef != nil && gwClass.Spec.ParametersRef.Group == lib.AkoGroup && gwClass.Spec.ParametersRef.Kind == lib.NsxAlbInfraSetting {
				infraSetting, err := lib.GetCRDInformers().NsxAlbInfraSettingInformer.Lister().Get(gwClass.Spec.ParametersRef.Name)
				if err != nil {
					utils.AviLog.Warnf("Unable to get corresponding NsxAlbInfraSetting: %s", err.Error())
				} else if infraSetting.Spec.SegGroup.Name != "" {
					// This assumes that the SeGroup has the appropriate labels configured
					avi_vs_meta.ServiceEngineGroup = infraSetting.Spec.SegGroup.Name
				}
			}
		}

		if avi_vs_meta.ServiceEngineGroup == "" {
			avi_vs_meta.ServiceEngineGroup = lib.GetSEGName()
		}

		isTCP := false
		var portProtocols []AviPortHostProtocol
		for _, listener := range listeners {
			portProto := strings.Split(listener, "/") // format: protocol/port
			port, _ := strconv.Atoi(portProto[1])
			pp := AviPortHostProtocol{Port: int32(port), Protocol: portProto[0]}
			portProtocols = append(portProtocols, pp)
			if portProto[0] == "" || portProto[0] == utils.TCP {
				isTCP = true
			}
		}
		avi_vs_meta.PortProto = portProtocols
		// Default case.
		avi_vs_meta.ApplicationProfile = utils.DEFAULT_L4_APP_PROFILE
		if !isTCP {
			avi_vs_meta.NetworkProfile = utils.SYSTEM_UDP_FAST_PATH
		} else {
			avi_vs_meta.NetworkProfile = utils.TCP_NW_FAST_PATH
		}

		vsVipNode := &AviVSVIPNode{
			Name:       lib.GetL4VSVipName(gatewayName, namespace),
			Tenant:     lib.GetTenant(),
			EastWest:   false,
			VrfContext: lib.GetVrf(),
		}

		if len(gw.Spec.Addresses) > 0 && gw.Spec.Addresses[0].Type == svcapiv1alpha1.IPAddressType {
			vsVipNode.IPAddress = gw.Spec.Addresses[0].Value
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
		if !utils.HasElem(gwListeners, listener) || len(svc) != 1 {
			continue
		}
		portProto := strings.Split(listener, "/") // format: protocol/port
		// assume it to have only a single backed service, the check is in isGatewayDelete
		svcNSName := strings.Split(svc[0], "/")
		port, _ := utilsnet.ParsePort(portProto[1], true)

		poolNode := &AviPoolNode{
			Name:       lib.GetAdvL4PoolName(svcNSName[1], namespace, gwName, int32(port)),
			Tenant:     lib.GetTenant(),
			Protocol:   portProto[0],
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
