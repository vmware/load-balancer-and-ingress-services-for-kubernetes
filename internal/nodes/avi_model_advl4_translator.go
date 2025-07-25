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
	"strings"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	akov1alpha2 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/apis/ako/v1alpha2"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/apis/ako/v1beta1"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	advl4v1alpha1pre1 "github.com/vmware-tanzu/service-apis/apis/v1alpha1pre1"
	"google.golang.org/protobuf/proto"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	utilsnet "k8s.io/utils/net"
	svcapiv1alpha1 "sigs.k8s.io/service-apis/apis/v1alpha1"
)

func (o *AviObjectGraph) BuildAdvancedL4Graph(namespace, gatewayName, key string, sharedVipOnSvcLBUsecase bool) {
	o.Lock.Lock()
	defer o.Lock.Unlock()
	var vsNode *AviVsNode

	if sharedVipOnSvcLBUsecase {
		vsNode = o.ConstructSharedVipSvcLBNode(gatewayName, namespace, key)
	} else if lib.UseServicesAPI() {
		vsNode = o.ConstructSvcApiL4VsNode(gatewayName, namespace, key)
	} else {
		vsNode = o.ConstructAdvL4VsNode(gatewayName, namespace, key)
	}
	if vsNode != nil {
		if !sharedVipOnSvcLBUsecase {
			o.ConstructAdvL4PolPoolNodes(vsNode, gatewayName, namespace, key)
		}
		o.AddModelNode(vsNode)
		utils.AviLog.Infof("key: %s, msg: checksum for AVI VS object %v", key, vsNode.GetCheckSum())
	}
}

func (o *AviObjectGraph) ConstructAdvL4VsNode(gatewayName, namespace, key string) *AviVsNode {
	// The logic: Each listener in the gateway is a listener port on the Avi VS.
	// A L4 policyset object is create where listener port --> pool. Pool gets it's server from the endpoints that has the same name as the 'service' pointed
	// by the listener port.
	found, listeners := objects.ServiceGWLister().GetGWListeners(namespace + "/" + gatewayName)
	if !found {
		return nil
	}

	vsName := lib.GetL4VSName(gatewayName, namespace)
	gw, err := lib.AKOControlConfig().AdvL4Informers().GatewayInformer.Lister().Gateways(namespace).Get(gatewayName)
	if err != nil {
		utils.AviLog.Warnf("key: %s, msg: GatewayLister returned error for advancedL4: %s", err)
		return nil
	}

	var serviceNSNames []string
	if found, services := objects.ServiceGWLister().GetGwToSvcs(namespace + "/" + gatewayName); found {
		for svcListener, service := range services {
			// assume it to have only a single backend service, the check is in isGatewayDelete
			if utils.HasElem(listeners, svcListener) && len(service) == 1 && !utils.HasElem(serviceNSNames, service[0]) {
				serviceNSNames = append(serviceNSNames, service[0])
			}
		}
	}

	infraSetting, err := getL4InfraSetting(key, namespace, nil, &gw.Spec.Class)
	if err != nil {
		if !k8serrors.IsNotFound(err) {
			utils.AviLog.Warnf("key: %s, msg: Error while fetching infrasetting for Gateway %s", key, err.Error())
		}
		infraSetting = nil
	}

	tenant := lib.GetTenantInNamespace(namespace)

	DeleteStaleTenantModelData(gatewayName, namespace, key, tenant, lib.ADVANCED_L4)

	objects.SharedNamespaceTenantLister().UpdateNamespacedResourceToTenantStore(namespace+"/"+gatewayName, tenant)

	avi_vs_meta := &AviVsNode{
		Name:   vsName,
		Tenant: tenant,
		ServiceMetadata: lib.ServiceMetadataObj{
			NamespaceServiceName: serviceNSNames,
			Gateway:              namespace + "/" + gatewayName,
		},
		ServiceEngineGroup: lib.GetSEGName(),
		EnableRhi:          proto.Bool(lib.GetEnableRHI()),
	}

	var vrfcontext string
	t1lr := lib.GetT1LRPath()
	if infraSetting != nil && infraSetting.Spec.NSXSettings.T1LR != nil {
		t1lr = *infraSetting.Spec.NSXSettings.T1LR
	}
	if t1lr == "" {
		vrfcontext = lib.GetVrf()
		avi_vs_meta.VrfContext = vrfcontext
	}

	avi_vs_meta.AviMarkers = lib.PopulateAdvL4VSNodeMarkers(namespace, gatewayName)

	isTCP, isUDP, isSCTP := false, false, false
	var portProtocols []AviPortHostProtocol
	for _, listener := range listeners {
		portProto := strings.Split(listener, "/") // format: protocol/port
		port, _ := utilsnet.ParsePort(portProto[1], true)
		pp := AviPortHostProtocol{Port: int32(port), Protocol: portProto[0]}
		portProtocols = append(portProtocols, pp)
		if portProto[0] == "" || portProto[0] == utils.TCP {
			isTCP = true
		} else if portProto[0] == utils.UDP {
			isUDP = true
		} else if portProto[0] == utils.SCTP {
			if lib.GetServiceType() == lib.NodePortLocal {
				utils.AviLog.Warnf("key: %s, msg: SCTP protocol is not supported for service type NodePortLocal", key)
				return nil
			}
			isSCTP = true
		}
	}

	avi_vs_meta.PortProto = portProtocols
	avi_vs_meta.ApplicationProfile = utils.DEFAULT_L4_APP_PROFILE

	avi_vs_meta.NetworkProfile = getNetworkProfile(isSCTP, isTCP, isUDP)

	vsVipNode := &AviVSVIPNode{
		Name:        lib.GetL4VSVipName(gatewayName, namespace),
		Tenant:      tenant,
		VrfContext:  vrfcontext,
		VipNetworks: utils.GetVipNetworkList(),
	}

	if t1lr != "" {
		vsVipNode.T1Lr = t1lr
	}

	if avi_vs_meta.EnableRhi != nil && *avi_vs_meta.EnableRhi {
		vsVipNode.BGPPeerLabels = lib.GetGlobalBgpPeerLabels()
	}
	// configures VS and VsVip nodes using infraSetting object (via CRD).
	buildWithInfraSetting(key, namespace, avi_vs_meta, vsVipNode, infraSetting)

	if len(gw.Spec.Addresses) > 0 && gw.Spec.Addresses[0].Type == advl4v1alpha1pre1.IPAddressType {
		vsVipNode.IPAddress = gw.Spec.Addresses[0].Value
	}
	avi_vs_meta.VSVIPRefs = append(avi_vs_meta.VSVIPRefs, vsVipNode)
	return avi_vs_meta
}

func (o *AviObjectGraph) ConstructSvcApiL4VsNode(gatewayName, namespace, key string) *AviVsNode {
	// The logic: Each listener in the gateway is a listener port on the Avi VS.
	// A L4 policyset object is create where listener port --> pool. Pool gets it's server from the endpoints that has the same name as the 'service' pointed
	// by the listener port.
	found, listeners := objects.ServiceGWLister().GetGWListeners(namespace + "/" + gatewayName)
	if !found {
		return nil
	}
	vsName := lib.GetL4VSName(gatewayName, namespace)
	gw, err := lib.AKOControlConfig().SvcAPIInformers().GatewayInformer.Lister().Gateways(namespace).Get(gatewayName)
	if err != nil {
		utils.AviLog.Warnf("key: %s, msg: GatewayLister returned error for services APIs : %s", err)
		return nil
	}

	var serviceNSNames []string
	listenerSvcMapping := make(map[string][]string)
	if found, services := objects.ServiceGWLister().GetGwToSvcs(namespace + "/" + gatewayName); found {
		for svcListener, service := range services {
			// assume it to have only a single backend service, the check is in isGatewayDelete
			if utils.HasElem(listeners, svcListener) && len(service) == 1 && !utils.HasElem(serviceNSNames, service[0]) {
				serviceNSNames = append(serviceNSNames, service[0])
				if val, ok := listenerSvcMapping[svcListener]; ok {
					listenerSvcMapping[svcListener] = append(val, service[0])
				} else {
					listenerSvcMapping[svcListener] = []string{service[0]}
				}
			}
		}
	}

	var fqdns []string
	for _, listener := range gw.Spec.Listeners {
		autoFQDN := true
		// Honour the hostname if specified corresponding to the listener.
		if listener.Hostname != nil && string(*listener.Hostname) != "" {
			fqdns = append(fqdns, string(*listener.Hostname))
			autoFQDN = false
		}

		subDomains := GetDefaultSubDomain()
		if subDomains != nil && autoFQDN {
			services := listenerSvcMapping[fmt.Sprintf("%s/%d", listener.Protocol, listener.Port)]
			for _, service := range services {
				svcNsName := strings.Split(service, "/")
				if fqdn := getAutoFQDNForService(svcNsName[0], svcNsName[1]); fqdn != "" {
					fqdns = append(fqdns, fqdn)
				}
			}
		}
	}

	infraSetting, err := getL4InfraSetting(key, namespace, nil, &gw.Spec.GatewayClassName)
	if err != nil {
		if !k8serrors.IsNotFound(err) {
			utils.AviLog.Warnf("key: %s, msg: Error while fetching infrasetting for Gateway %s", key, err.Error())
		}
		infraSetting = nil
	}
	tenant := lib.GetTenantInNamespace(namespace)

	DeleteStaleTenantModelData(gatewayName, namespace, key, tenant, lib.ADVANCED_L4)

	objects.SharedNamespaceTenantLister().UpdateNamespacedResourceToTenantStore(namespace+"/"+gatewayName, tenant)

	avi_vs_meta := &AviVsNode{
		Name:   vsName,
		Tenant: tenant,
		ServiceMetadata: lib.ServiceMetadataObj{
			Gateway:   namespace + "/" + gatewayName,
			HostNames: fqdns,
		},
		ServiceEngineGroup: lib.GetSEGName(),
		EnableRhi:          proto.Bool(lib.GetEnableRHI()),
	}

	var vrfcontext string
	t1lr := lib.GetT1LRPath()
	if infraSetting != nil && infraSetting.Spec.NSXSettings.T1LR != nil {
		t1lr = *infraSetting.Spec.NSXSettings.T1LR
	}
	if t1lr == "" {
		vrfcontext = lib.GetVrf()
		avi_vs_meta.VrfContext = vrfcontext
	}

	isTCP, isUDP, isSCTP := false, false, false
	avi_vs_meta.AviMarkers = lib.PopulateAdvL4VSNodeMarkers(namespace, gatewayName)
	var portProtocols []AviPortHostProtocol
	for _, listener := range listeners {
		portProto := strings.Split(listener, "/") // format: protocol/port
		port, _ := utilsnet.ParsePort(portProto[1], true)
		pp := AviPortHostProtocol{Port: int32(port), Protocol: portProto[0]}
		portProtocols = append(portProtocols, pp)
		if portProto[0] == "" || portProto[0] == utils.TCP {
			isTCP = true
		} else if portProto[0] == utils.UDP {
			isUDP = true
		} else if portProto[0] == utils.SCTP {
			if lib.GetServiceType() == lib.NodePortLocal {
				utils.AviLog.Warnf("key: %s, msg: SCTP protocol is not supported for service type NodePortLocal", key)
				return nil
			}
			isSCTP = true
		}
	}

	avi_vs_meta.PortProto = portProtocols
	avi_vs_meta.ApplicationProfile = utils.DEFAULT_L4_APP_PROFILE

	avi_vs_meta.NetworkProfile = getNetworkProfile(isSCTP, isTCP, isUDP)

	vsVipNode := &AviVSVIPNode{
		Name:        lib.GetL4VSVipName(gatewayName, namespace),
		Tenant:      tenant,
		VrfContext:  vrfcontext,
		FQDNs:       fqdns,
		VipNetworks: utils.GetVipNetworkList(),
	}

	if t1lr != "" {
		vsVipNode.T1Lr = t1lr
	}

	if avi_vs_meta.EnableRhi != nil && *avi_vs_meta.EnableRhi {
		vsVipNode.BGPPeerLabels = lib.GetGlobalBgpPeerLabels()
	}
	// configures VS and VsVip nodes using infraSetting object (via CRD).
	buildWithInfraSetting(key, namespace, avi_vs_meta, vsVipNode, infraSetting)

	if len(gw.Spec.Addresses) > 0 && gw.Spec.Addresses[0].Type == svcapiv1alpha1.IPAddressType {
		vsVipNode.IPAddress = gw.Spec.Addresses[0].Value
	}

	avi_vs_meta.VSVIPRefs = append(avi_vs_meta.VSVIPRefs, vsVipNode)
	return avi_vs_meta
}

func (o *AviObjectGraph) ConstructAdvL4PolPoolNodes(vsNode *AviVsNode, gwName, namespace, key string) {
	var l4Policies []*AviL4PolicyNode
	found, svcListeners := objects.ServiceGWLister().GetGwToSvcs(namespace + "/" + gwName)
	foundGW, gwListeners := objects.ServiceGWLister().GetGWListeners(namespace + "/" + gwName)
	if !found || !foundGW {
		return
	}

	// create a mapping of portProto to hostname
	gwListenerHostNameMapping := make(map[string]string)
	gwClassName := ""
	if lib.UseServicesAPI() {
		// enable fqdn for gateway services only for non-advancedl4 usecases.
		gw, err := lib.AKOControlConfig().SvcAPIInformers().GatewayInformer.Lister().Gateways(namespace).Get(gwName)
		if err != nil {
			utils.AviLog.Warnf("key: %s, msg: Gateway deleted, not constructing the pool nodes", key)
			return
		}
		for _, gwlistener := range gw.Spec.Listeners {
			if gwlistener.Hostname != nil && string(*gwlistener.Hostname) != "" {
				gwListenerHostNameMapping[fmt.Sprintf("%s/%d", gwlistener.Protocol, gwlistener.Port)] = string(*gwlistener.Hostname)
			}
		}
		gwClassName = gw.Spec.GatewayClassName
	}

	infraSetting, err := getL4InfraSetting(key, namespace, nil, &gwClassName)
	if err != nil {
		if !k8serrors.IsNotFound(err) {
			utils.AviLog.Warnf("key: %s, msg: Error while fetching infrasetting for Gateway %s", key, err.Error())
		}
		infraSetting = nil
	}

	t1lr := lib.GetT1LRPath()
	if infraSetting != nil && infraSetting.Spec.NSXSettings.T1LR != nil {
		t1lr = *infraSetting.Spec.NSXSettings.T1LR
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

		var svcFQDN string
		if fqdn, ok := gwListenerHostNameMapping[listener]; ok {
			svcFQDN = fqdn
		}
		if lib.GetL4FqdnFormat() != lib.AutoFQDNDisabled && svcFQDN == "" {
			svcFQDN = getAutoFQDNForService(svcNSName[0], svcNSName[1])
		}

		poolName := lib.GetAdvL4PoolName(svcNSName[1], namespace, gwName, portProto[0], int32(port))
		if lib.UseServicesAPI() {
			poolName = lib.GetSvcApiL4PoolName(svcNSName[1], namespace, gwName, portProto[0], int32(port))
		}

		poolNode := &AviPoolNode{
			Name:     poolName,
			Tenant:   vsNode.Tenant,
			Protocol: portProto[0],
			PortName: "",
			ServiceMetadata: lib.ServiceMetadataObj{
				NamespaceServiceName: []string{svc[0]},
			},
			VrfContext: lib.GetVrf(),
		}

		// Unset the poolnode's vrfcontext.
		if t1lr != "" {
			poolNode.T1Lr = t1lr
			poolNode.VrfContext = ""
		}

		poolNode.NetworkPlacementSettings = lib.GetNodeNetworkMap()

		if svcFQDN != "" {
			poolNode.ServiceMetadata.HostNames = []string{svcFQDN}
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

		serviceType := lib.GetServiceType()
		if serviceType == lib.NodePortLocal {
			if servers := PopulateServersForNPL(poolNode, svcObj.ObjectMeta.Namespace, svcObj.ObjectMeta.Name, false, key); servers != nil {
				poolNode.Servers = servers
			}
		} else if serviceType == lib.NodePort {
			if servers := PopulateServersForNodePort(poolNode, svcObj.ObjectMeta.Namespace, svcObj.ObjectMeta.Name, false, key); servers != nil {
				poolNode.Servers = servers
			}
		} else {
			if servers := PopulateServers(poolNode, svcObj.ObjectMeta.Namespace, svcObj.ObjectMeta.Name, false, key); servers != nil {
				poolNode.Servers = servers
			}
		}

		if lib.UseServicesAPI() {
			poolNode.AviMarkers = lib.PopulateSvcApiL4PoolNodeMarkers(namespace, svcNSName[1], gwName, portProto[0], port)
		} else {
			poolNode.AviMarkers = lib.PopulateAdvL4PoolNodeMarkers(namespace, svcNSName[1], gwName, port)
		}

		poolRef := fmt.Sprintf("/api/pool?name=%s", poolNode.Name)
		portPool := AviHostPathPortPoolPG{
			Port:     uint32(port),
			Pool:     poolRef,
			Protocol: portProto[0],
		}
		portPoolSet = append(portPoolSet, portPool)

		buildPoolWithInfraSetting(key, poolNode, infraSetting)
		if lib.IsIstioEnabled() {
			poolNode.UpdatePoolNodeForIstio()
		}

		vsNode.PoolRefs = append(vsNode.PoolRefs, poolNode)
		utils.AviLog.Infof("key: %s, msg: evaluated L4 pool values :%v", key, utils.Stringify(poolNode))
	}

	l4policyNode := &AviL4PolicyNode{
		Name:       vsNode.Name,
		Tenant:     vsNode.Tenant,
		PortPool:   portPoolSet,
		AviMarkers: lib.PopulateAdvL4VSNodeMarkers(namespace, gwName),
	}

	l4Policies = append(l4Policies, l4policyNode)
	vsNode.L4PolicyRefs = l4Policies
	utils.AviLog.Infof("key: %s, msg: evaluated L4 pool policies :%v", key, utils.Stringify(vsNode.L4PolicyRefs))
}

func (o *AviObjectGraph) ConstructSharedVipSvcLBNode(sharedVipKey, namespace, key string) *AviVsNode {
	namespacedShareVipKey := namespace + "/" + sharedVipKey
	found, serviceNSNames := objects.SharedlbLister().GetSharedVipKeyToServices(namespacedShareVipKey)
	if !found {
		return nil
	}

	vsName := lib.GetL4VSName(sharedVipKey, namespace)

	var fqdns []string
	autoFQDN := true
	subDomains := GetDefaultSubDomain()
	if subDomains != nil && autoFQDN {
		for _, service := range serviceNSNames {
			svcNsName := strings.Split(service, "/")
			svcObj, err := utils.GetInformers().ServiceInformer.Lister().Services(namespace).Get(svcNsName[1])
			if err == nil {
				if extDNS, ok := svcObj.Annotations[lib.ExternalDNSAnnotation]; ok {
					fqdns = append(fqdns, extDNS)
					continue
				}
			}

			if fqdn := getAutoFQDNForService(svcNsName[0], svcNsName[1]); fqdn != "" {
				fqdns = append(fqdns, fqdn)
			}
		}
	}

	avi_vs_meta := &AviVsNode{
		Name:       vsName,
		Tenant:     lib.GetTenant(),
		VrfContext: lib.GetVrf(),
		ServiceMetadata: lib.ServiceMetadataObj{
			HostNames: fqdns,
		},
		ServiceEngineGroup: lib.GetSEGName(),
		EnableRhi:          proto.Bool(lib.GetEnableRHI()),
	}

	isTCP, isUDP, isSCTP := false, false, false
	avi_vs_meta.AviMarkers = lib.PopulateAdvL4VSNodeMarkers(namespace, sharedVipKey)
	var portProtocols []AviPortHostProtocol
	var sharedPreferredVIP string
	var serviceObject *v1.Service
	for i, serviceNSName := range serviceNSNames {
		svcNSName := strings.Split(serviceNSName, "/")
		svcObj, err := utils.GetInformers().ServiceInformer.Lister().Services(svcNSName[0]).Get(svcNSName[1])
		if err != nil {
			utils.AviLog.Debugf("key: %s, msg: there was an error in retrieving the service", key)
			return nil
		}

		if i == 0 {
			if lib.HasSpecLoadBalancerIP(svcObj) {
				sharedPreferredVIP = svcObj.Spec.LoadBalancerIP
			} else if lib.HasLoadBalancerIPAnnotation(svcObj) {
				sharedPreferredVIP = svcObj.Annotations[lib.LoadBalancerIP]
			}
			if infraSettingAnnotation, ok := svcObj.GetAnnotations()[lib.InfraSettingNameAnnotation]; ok && infraSettingAnnotation != "" {
				serviceObject = svcObj.DeepCopy()
			}
			if l4RuleName, ok := svcObj.GetAnnotations()[lib.L4RuleAnnotation]; ok && l4RuleName != "" {
				serviceObject = svcObj.DeepCopy()
			}
		}

		for _, listener := range svcObj.Spec.Ports {
			protocol := string(listener.Protocol)
			pp := AviPortHostProtocol{Port: listener.Port, Protocol: protocol}
			pp.Name = fmt.Sprintf("%s-%d", protocol, listener.Port)
			portProtocols = append(portProtocols, pp)
			if protocol == "" || protocol == utils.TCP {
				isTCP = true
			} else if protocol == utils.UDP {
				isUDP = true
			} else if protocol == utils.SCTP {
				if lib.GetServiceType() == lib.NodePortLocal {
					utils.AviLog.Warnf("key: %s, msg: SCTP protocol is not supported for service type NodePortLocal", key)
					return nil
				}
				isSCTP = true
			}
		}
	}

	avi_vs_meta.PortProto = portProtocols
	avi_vs_meta.ApplicationProfile = utils.DEFAULT_L4_APP_PROFILE

	avi_vs_meta.NetworkProfile = getNetworkProfile(isSCTP, isTCP, isUDP)

	vsVipNode := &AviVSVIPNode{
		Name:        lib.GetL4VSVipName(sharedVipKey, namespace),
		Tenant:      lib.GetTenant(),
		VrfContext:  lib.GetVrf(),
		FQDNs:       fqdns,
		VipNetworks: utils.GetVipNetworkList(),
	}

	if sharedPreferredVIP != "" {
		vsVipNode.IPAddress = sharedPreferredVIP
	}

	if avi_vs_meta.EnableRhi != nil && *avi_vs_meta.EnableRhi {
		vsVipNode.BGPPeerLabels = lib.GetGlobalBgpPeerLabels()
	}

	// configures VS and VsVip nodes using infraSetting object (via CRD).
	if serviceObject != nil {
		if infraSetting, err := getL4InfraSetting(key, namespace, serviceObject, nil); err == nil {
			buildWithInfraSetting(key, namespace, avi_vs_meta, vsVipNode, infraSetting)
		}

		// Copy the VS properties from L4Rule object
		if l4Rule, err := getL4Rule(key, serviceObject); err == nil {
			buildWithL4Rule(key, avi_vs_meta, l4Rule)

			// Copy the LoadBalancerIP if configured in L4Rule CRD.
			if avi_vs_meta.LoadBalancerIP != nil {
				vsVipNode.IPAddress = *avi_vs_meta.LoadBalancerIP
			}
		}
	}

	avi_vs_meta.VSVIPRefs = append(avi_vs_meta.VSVIPRefs, vsVipNode)

	o.ConstructSharedVipPolPoolNodes(avi_vs_meta, sharedVipKey, namespace, key)
	return avi_vs_meta
}

func (o *AviObjectGraph) ConstructSharedVipPolPoolNodes(vsNode *AviVsNode, sharedVipKey, namespace, key string) {
	namespacedShareVipKey := namespace + "/" + sharedVipKey
	found, serviceNSNames := objects.SharedlbLister().GetSharedVipKeyToServices(namespacedShareVipKey)
	if !found {
		return
	}

	var l4Policies []*AviL4PolicyNode
	var infraSetting *v1beta1.AviInfraSetting
	var l4Rule *akov1alpha2.L4Rule

	var portPoolSet []AviHostPathPortPoolPG
	for i, serviceNSName := range serviceNSNames {
		svcNSName := strings.Split(serviceNSName, "/")
		svcObj, err := utils.GetInformers().ServiceInformer.Lister().Services(svcNSName[0]).Get(svcNSName[1])
		if err != nil {
			utils.AviLog.Debugf("key: %s, msg: there was an error in retrieving the service", key)
			return
		}

		if i == 0 {
			infraSetting, err = getL4InfraSetting(key, namespace, svcObj, nil)
			if err != nil {
				utils.AviLog.Warnf("key: %s, msg: Error while fetching infrasetting for Service %s", key, err.Error())
			}

			l4Rule, err = getL4Rule(key, svcObj)
			if err != nil {
				utils.AviLog.Warnf("key: %s, msg: Error while fetching l4Rule for Service %s", key, err.Error())
				return
			}
		}

		for _, listener := range svcObj.Spec.Ports {
			protocol := string(listener.Protocol)
			port := listener.Port
			targetPort := listener.TargetPort

			var svcFQDN string
			if lib.GetL4FqdnFormat() != lib.AutoFQDNDisabled && svcFQDN == "" {
				svcFQDN = getAutoFQDNForService(svcNSName[0], svcNSName[1])
			}

			poolName := lib.GetSvcApiL4PoolName(svcNSName[1], namespace, sharedVipKey, protocol, port)
			poolNode := &AviPoolNode{
				Name:     poolName,
				Tenant:   vsNode.Tenant,
				Protocol: protocol,
				PortName: listener.Name,
				ServiceMetadata: lib.ServiceMetadataObj{
					NamespaceServiceName: []string{serviceNSName},
				},
				VrfContext: lib.GetVrf(),
				Port:       port,
				TargetPort: targetPort,
			}

			buildPoolWithL4Rule(key, poolNode, l4Rule)

			poolNode.NetworkPlacementSettings = lib.GetNodeNetworkMap()

			if svcFQDN != "" {
				poolNode.ServiceMetadata.HostNames = []string{svcFQDN}
			}

			serviceType := lib.GetServiceType()
			if serviceType == lib.NodePortLocal {
				if servers := PopulateServersForNPL(poolNode, svcObj.ObjectMeta.Namespace, svcObj.ObjectMeta.Name, false, key); servers != nil {
					poolNode.Servers = servers
				}
			} else if _, ok := svcObj.GetAnnotations()[lib.SkipNodePortAnnotation]; ok {
				// This annotation's presence on the svc object means that the node ports should be skipped.
				if servers := PopulateServers(poolNode, svcObj.ObjectMeta.Namespace, svcObj.ObjectMeta.Name, false, key); servers != nil {
					poolNode.Servers = servers
				}
			} else if serviceType == lib.NodePort {
				if servers := PopulateServersForNodePort(poolNode, svcObj.ObjectMeta.Namespace, svcObj.ObjectMeta.Name, false, key); servers != nil {
					poolNode.Servers = servers
				}
			} else {
				if servers := PopulateServers(poolNode, svcObj.ObjectMeta.Namespace, svcObj.ObjectMeta.Name, false, key); servers != nil {
					poolNode.Servers = servers
				}
			}

			poolNode.AviMarkers = lib.PopulateSvcApiL4PoolNodeMarkers(namespace, svcNSName[1], sharedVipKey, protocol, int(port))
			poolRef := fmt.Sprintf("/api/pool?name=%s", poolNode.Name)
			portPool := AviHostPathPortPoolPG{
				Port:     uint32(port),
				Pool:     poolRef,
				Protocol: protocol,
			}
			portPool.Name = fmt.Sprintf("%s-%d", protocol, port)
			portPoolSet = append(portPoolSet, portPool)

			buildPoolWithInfraSetting(key, poolNode, infraSetting)
			if lib.IsIstioEnabled() {
				poolNode.UpdatePoolNodeForIstio()
			}

			vsNode.PoolRefs = append(vsNode.PoolRefs, poolNode)
			utils.AviLog.Infof("key: %s, msg: evaluated L4 pool values :%v", key, utils.Stringify(poolNode))
		}
	}

	l4policyNode := &AviL4PolicyNode{
		Name:       vsNode.Name,
		Tenant:     vsNode.Tenant,
		PortPool:   portPoolSet,
		AviMarkers: lib.PopulateAdvL4VSNodeMarkers(namespace, sharedVipKey),
	}

	l4Policies = append(l4Policies, l4policyNode)
	vsNode.L4PolicyRefs = l4Policies
	utils.AviLog.Infof("key: %s, msg: evaluated L4 pool policies :%v", key, utils.Stringify(vsNode.L4PolicyRefs))
}
