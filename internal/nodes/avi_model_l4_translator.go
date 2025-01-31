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
	"sort"
	"strconv"
	"strings"

	"google.golang.org/protobuf/proto"

	discovery "k8s.io/api/discovery/v1"

	avicache "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	akov1alpha2 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/apis/ako/v1alpha2"
	akov1beta1 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/apis/ako/v1beta1"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	"github.com/jinzhu/copier"
	"github.com/vmware/alb-sdk/go/models"
	avimodels "github.com/vmware/alb-sdk/go/models"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/sets"
	k8net "k8s.io/utils/net"
)

func (o *AviObjectGraph) ConstructAviL4VsNode(svcObj *corev1.Service, key string) *AviVsNode {
	var avi_vs_meta *AviVsNode
	var fqdns []string
	autoFQDN := true
	if lib.GetL4FqdnFormat() == lib.AutoFQDNDisabled {
		autoFQDN = false
	}

	if extDNS, ok := svcObj.Annotations[lib.ExternalDNSAnnotation]; ok && autoFQDN {
		autoFQDN = false
		fqdns = append(fqdns, extDNS)
	}

	subDomains := GetDefaultSubDomain()
	if subDomains != nil && autoFQDN {
		if fqdn := getAutoFQDNForService(svcObj.Namespace, svcObj.Name); fqdn != "" {
			fqdns = append(fqdns, fqdn)
		}
	}

	tenant := lib.GetTenantInNamespace(svcObj.GetNamespace())

	DeleteStaleTenantModelData(svcObj.GetName(), svcObj.GetNamespace(), key, tenant, lib.L4VS)

	objects.SharedNamespaceTenantLister().UpdateNamespacedResourceToTenantStore(svcObj.GetNamespace()+"/"+svcObj.GetName(), tenant)

	infraSetting, err := getL4InfraSetting(key, svcObj.Namespace, svcObj, nil)
	if err != nil {
		utils.AviLog.Warnf("key: %s, msg: Error while fetching infrasetting for Service %s", key, err.Error())
	}

	vsName := lib.GetL4VSName(svcObj.ObjectMeta.Name, svcObj.ObjectMeta.Namespace)
	avi_vs_meta = &AviVsNode{
		Name:   vsName,
		Tenant: tenant,
		ServiceMetadata: lib.ServiceMetadataObj{
			NamespaceServiceName: []string{svcObj.ObjectMeta.Namespace + "/" + svcObj.ObjectMeta.Name},
			HostNames:            fqdns,
		},
		ServiceEngineGroup: lib.GetSEGName(),
		EnableRhi:          proto.Bool(lib.GetEnableRHI()),
	}

	vrfcontext := lib.GetVrf()
	t1lr := lib.GetT1LRPath()
	if infraSetting != nil && infraSetting.Spec.NSXSettings.T1LR != nil {
		t1lr = *infraSetting.Spec.NSXSettings.T1LR
	}
	if t1lr != "" {
		vrfcontext = ""
	} else {
		avi_vs_meta.VrfContext = vrfcontext
	}
	avi_vs_meta.AviMarkers = lib.PopulateL4VSNodeMarkers(svcObj.ObjectMeta.Namespace, svcObj.ObjectMeta.Name)
	isTCP, isSCTP, isUDP := false, false, false
	var portProtocols []AviPortHostProtocol
	for _, port := range svcObj.Spec.Ports {
		pp := AviPortHostProtocol{Port: int32(port.Port), Protocol: fmt.Sprint(port.Protocol), Name: port.Name, TargetPort: port.TargetPort}
		portProtocols = append(portProtocols, pp)
		if port.Protocol == "" || port.Protocol == utils.TCP {
			isTCP = true
		} else if port.Protocol == utils.SCTP {
			if lib.GetServiceType() == lib.NodePortLocal {
				utils.AviLog.Warnf("key: %s, msg: SCTP protocol is not supported for service type NodePortLocal", key)
				return nil
			}
			isSCTP = true
		} else if port.Protocol == utils.UDP {
			isUDP = true
		}
	}
	avi_vs_meta.PortProto = portProtocols

	if appProfile, ok := svcObj.GetAnnotations()[lib.LBSvcAppProfileAnnotation]; ok && appProfile != "" {
		avi_vs_meta.ApplicationProfile = appProfile
	} else {
		// Default case
		avi_vs_meta.ApplicationProfile = utils.DEFAULT_L4_APP_PROFILE
	}

	avi_vs_meta.NetworkProfile = getNetworkProfile(isSCTP, isTCP, isUDP)

	vsVipName := lib.GetL4VSVipName(svcObj.ObjectMeta.Name, svcObj.ObjectMeta.Namespace)
	vsVipNode := &AviVSVIPNode{
		Name:        vsVipName,
		Tenant:      tenant,
		FQDNs:       fqdns,
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
	buildWithInfraSetting(key, svcObj.Namespace, avi_vs_meta, vsVipNode, infraSetting)

	// Copy the VS properties from L4Rule object
	if l4Rule, err := getL4Rule(key, svcObj); err == nil {
		buildWithL4Rule(key, avi_vs_meta, l4Rule)
	}

	if lib.HasSpecLoadBalancerIP(svcObj) {
		vsVipNode.IPAddress = svcObj.Spec.LoadBalancerIP
	} else if lib.HasLoadBalancerIPAnnotation(svcObj) {
		vsVipNode.IPAddress = svcObj.Annotations[lib.LoadBalancerIP]
	} else if avi_vs_meta.LoadBalancerIP != nil {
		vsVipNode.IPAddress = *avi_vs_meta.LoadBalancerIP
	}

	avi_vs_meta.VSVIPRefs = append(avi_vs_meta.VSVIPRefs, vsVipNode)
	return avi_vs_meta
}

func (o *AviObjectGraph) ConstructAviL4PolPoolNodes(svcObj *corev1.Service, vsNode *AviVsNode, key string) {
	var l4Policies []*AviL4PolicyNode
	var portPoolSet []AviHostPathPortPoolPG

	l4Rule, err := getL4Rule(key, svcObj)
	if err != nil {
		utils.AviLog.Warnf("key: %s, msg: Error while fetching L4Rule. Err: %s", key, err.Error())
	}

	isSSLEnabled := false
	for _, aviSvc := range vsNode.Services {
		if *aviSvc.EnableSsl {
			isSSLEnabled = true
		}
	}
	infraSetting, err := getL4InfraSetting(key, svcObj.Namespace, svcObj, nil)
	if err != nil {
		utils.AviLog.Warnf("key: %s, msg: Error while fetching infrasetting for Service %s", key, err.Error())
	}
	tenant := lib.GetTenantInNamespace(svcObj.GetNamespace())

	protocolSet := sets.NewString()
	for _, portProto := range vsNode.PortProto {
		filterPort := portProto.Port
		poolNode := &AviPoolNode{
			Name:       lib.GetL4PoolName(svcObj.ObjectMeta.Name, svcObj.ObjectMeta.Namespace, portProto.Protocol, filterPort),
			Tenant:     tenant,
			Protocol:   portProto.Protocol,
			PortName:   portProto.Name,
			Port:       portProto.Port,
			TargetPort: portProto.TargetPort,
			VrfContext: lib.GetVrf(),
		}

		buildPoolWithL4Rule(key, poolNode, l4Rule)

		if lib.IsIstioEnabled() {
			poolNode.UpdatePoolNodeForIstio()
		}
		protocolSet.Insert(portProto.Protocol)
		poolNode.NetworkPlacementSettings = lib.GetNodeNetworkMap()

		t1lr := lib.GetT1LRPath()
		if infraSetting != nil && infraSetting.Spec.NSXSettings.T1LR != nil {
			t1lr = *infraSetting.Spec.NSXSettings.T1LR
		}
		if t1lr != "" {
			poolNode.T1Lr = t1lr
			// Unset the poolnode's vrfcontext.
			poolNode.VrfContext = ""
		}

		serviceType := lib.GetServiceType()
		if serviceType == lib.NodePortLocal {
			if svcObj.Spec.Type == "NodePort" {
				utils.AviLog.Warnf("key: %s, msg: Service of type NodePort is not supported when `serviceType` is NodePortLocal.", key)
			} else {
				if servers := PopulateServersForNPL(poolNode, svcObj.ObjectMeta.Namespace, svcObj.ObjectMeta.Name, false, key); servers != nil {
					poolNode.Servers = servers
				}
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

		poolNode.AviMarkers = lib.PopulateL4PoolNodeMarkers(svcObj.ObjectMeta.Namespace, svcObj.ObjectMeta.Name, strconv.Itoa(int(filterPort)))
		pool_ref := fmt.Sprintf("/api/pool?name=%s", poolNode.Name)
		portPool := AviHostPathPortPoolPG{Port: uint32(filterPort), Pool: pool_ref, Protocol: portProto.Protocol}
		portPoolSet = append(portPoolSet, portPool)

		buildPoolWithInfraSetting(key, poolNode, infraSetting)

		if isSSLEnabled {
			vsNode.DefaultPool = poolNode.Name
		}
		vsNode.PoolRefs = append(vsNode.PoolRefs, poolNode)
		utils.AviLog.Infof("key: %s, msg: evaluated L4 pool values :%v", key, utils.Stringify(poolNode))
	}

	if !isSSLEnabled {
		l4policyNode := &AviL4PolicyNode{Name: vsNode.Name, Tenant: vsNode.Tenant, PortPool: portPoolSet}
		sort.Strings(protocolSet.List())
		protocols := strings.Join(protocolSet.List(), ",")
		l4policyNode.AviMarkers = lib.PopulateL4PolicysetMarkers(svcObj.ObjectMeta.Namespace, svcObj.ObjectMeta.Name, protocols)
		l4Policies = append(l4Policies, l4policyNode)
		vsNode.L4PolicyRefs = l4Policies
	}

	//As pool naming covention changed for L4 pools marking flag, so that cksum will be changed
	vsNode.IsL4VS = true
	if len(vsNode.L4PolicyRefs) != 0 {
		utils.AviLog.Infof("key: %s, msg: evaluated L4 pool policies :%v", key, utils.Stringify(vsNode.L4PolicyRefs))
	}
}

func PopulateServersForNPL(poolNode *AviPoolNode, ns string, serviceName string, ingress bool, key string) []AviPoolMetaServer {
	if ingress {
		found, _ := objects.SharedClusterIpLister().Get(ns + "/" + serviceName)
		if !found {
			utils.AviLog.Warnf("key: %s, msg: service pointed by the ingress object is not found in ClusterIP store", key)
			return nil
		}
	}
	pods, targetPort := lib.GetPodsFromService(ns, serviceName, poolNode.TargetPort, key)
	if len(pods) == 0 {
		utils.AviLog.Infof("key: %s, msg: got no Pod for Service %s", key, serviceName)
		return make([]AviPoolMetaServer, 0)
	}
	ipFamily := lib.GetIPFamily()
	v4enabled := ipFamily == "V4" || ipFamily == "V4_V6"
	v6enabled := ipFamily == "V6" || ipFamily == "V4_V6"
	v4Family := false
	v6Family := false
	svcObj, _ := utils.GetInformers().ServiceInformer.Lister().Services(ns).Get(serviceName)
	if len(svcObj.Spec.IPFamilies) == 2 {
		v4Family = true
		v6Family = true
	} else if svcObj.Spec.IPFamilies[0] == "IPv6" {
		v6Family = true
	} else {
		v4Family = true
	}
	v4ServerCount := 0
	v6ServerCount := 0
	var poolMeta []AviPoolMetaServer

	// create a mapping from pod name to its endpoint condition
	conditionMap := map[string]*bool{}
	if lib.AKOControlConfig().GetEndpointSlicesEnabled() {
		epSliceIntList, err := utils.GetInformers().EpSlicesInformer.Informer().GetIndexer().ByIndex(discovery.LabelServiceName, ns+"/"+serviceName)
		if err != nil {
			utils.AviLog.Warnf("key: %s, msg: error while retrieving endpointsice: %s", key, err)
			return nil
		}
		for _, epSliceInt := range epSliceIntList {
			epSlice, isEpSliceClass := epSliceInt.(*discovery.EndpointSlice)
			if !isEpSliceClass {
				// not an epslice. continue
				utils.AviLog.Warnf("key: %s, msg: invalid endpointslice object", key)
				continue
			}
			// select epslice containing target port
			found := false
			for _, port := range epSlice.Ports {
				if (port.Port != nil && poolNode.TargetPort.IntVal == *port.Port) ||
					(port.Name != nil && poolNode.PortName == *port.Name) {
					found = true
					break
				}
			}
			if !found {
				continue
			}
			// we will go through the pods and build a map for condition
			for _, pod := range pods {
				for _, ep := range epSlice.Endpoints {
					if ep.TargetRef.Name == pod.Name {
						condition := enableServer(ep.Conditions)
						conditionMap[pod.Name] = condition
						utils.AviLog.Debugf("key: %s, msg: found pod %s with condition %t", key, pod.Name, *condition)
						break
					}
				}
			}
		}
		poolNode.GracefulShutdownTimeout = lib.AKOControlConfig().GetGracefulShutdownTimeout()
	}

	for _, pod := range pods {
		var annotations []lib.NPLAnnotation
		found, obj := objects.SharedNPLLister().Get(ns + "/" + pod.Name)
		if !found {
			continue
		}
		annotations = obj.([]lib.NPLAnnotation)
		for _, a := range annotations {
			var atype string
			if v4enabled && v4Family && utils.IsV4(a.NodeIP) {
				v4ServerCount++
				atype = "V4"
			} else if v6enabled && v6Family && k8net.IsIPv6String(a.NodeIP) {
				v6ServerCount++
				atype = "V6"
			} else {
				continue
			}
			if (poolNode.TargetPort.Type == intstr.Int && a.PodPort == poolNode.TargetPort.IntValue()) ||
				a.PodPort == int(targetPort) {
				enabled, ok := conditionMap[pod.Name]
				if !ok {
					// not disabling just because the pod was not found in condition map. i.e. ignoring false negatives
					utils.AviLog.Debugf("key: %s, msg: pod %s not found in condition map. enabling by default", key, pod.Name)
					enabled = proto.Bool(true)
				}
				server := AviPoolMetaServer{
					Port: int32(a.NodePort),
					Ip: models.IPAddr{
						Addr: &a.NodeIP,
						Type: &atype,
					},
					Enabled: enabled}
				poolMeta = append(poolMeta, server)
			}
		}
	}
	if len(poolMeta) == 0 {
		utils.AviLog.Warnf("key: %s, msg: no servers for port: %v (%v)", key, poolNode.Port, poolNode.PortName)
	} else {
		if v4Family && v4ServerCount == 0 {
			utils.AviLog.Warnf("key: %s, msg: expected IPv4 servers but found none for port %v (%v)", key, poolNode.Port, poolNode.PortName)
		}
		if v6Family && v6ServerCount == 0 {
			utils.AviLog.Warnf("key: %s, msg: expected IPv6 servers but found none for port %v (%v)", key, poolNode.Port, poolNode.PortName)
		}
		utils.AviLog.Infof("key: %s, msg: servers for port: %v (%v), are: %v", key, poolNode.Port, poolNode.PortName, utils.Stringify(poolMeta))
	}

	return poolMeta
}

func PopulateServersForNodePort(poolNode *AviPoolNode, ns string, serviceName string, ingress bool, key string) []AviPoolMetaServer {

	ipFamily := lib.GetIPFamily()
	v4enabled := ipFamily == "V4" || ipFamily == "V4_V6"
	v6enabled := ipFamily == "V6" || ipFamily == "V4_V6"
	v4Family := false
	v6Family := false
	v4ServerCount := 0
	v6ServerCount := 0
	// Get all nodes which match nodePortSelector
	nodePortSelector := lib.GetNodePortsSelector()
	nodePortFilter := map[string]string{}
	if len(nodePortSelector) == 2 && nodePortSelector["key"] != "" {
		nodePortFilter[nodePortSelector["key"]] = nodePortSelector["value"]
	} else {
		nodePortFilter = nil
	}
	allNodes := objects.SharedNodeLister().CopyAllObjects()

	var poolMeta []AviPoolMetaServer
	svcObj, err := utils.GetInformers().ServiceInformer.Lister().Services(ns).Get(serviceName)
	if err != nil {
		utils.AviLog.Warnf("key: %s, msg: error in obtaining the object for service: %s", key, serviceName)
		return poolMeta
	}
	if len(svcObj.Spec.IPFamilies) == 2 {
		v4Family = true
		v6Family = true
	} else if svcObj.Spec.IPFamilies[0] == "IPv6" {
		v6Family = true
	} else {
		v4Family = true
	}
	// Populate pool servers
	if lib.IsServiceClusterIPType(svcObj) {
		utils.AviLog.Debugf("key: %s, msg: ClusterIP is not processed in NodePort: %s", key, serviceName)
		return poolMeta
	}
	for _, port := range svcObj.Spec.Ports {
		if port.Name != poolNode.PortName && len(svcObj.Spec.Ports) != 1 {
			// continue only if port name does not match and its multiport svcobj
			continue
		}
		svcPort := int32(port.NodePort)
		poolNode.Port = svcPort
		for _, nodeIntf := range allNodes {
			node, ok := nodeIntf.(*corev1.Node)
			if !ok {
				utils.AviLog.Warnf("key: %s,msg: error in fetching node from node cache", key)
				continue
			}
			if nodePortFilter != nil {
				// skip the node if node does not have node port selector labels
				_, ok := node.ObjectMeta.Labels[nodePortSelector["key"]]
				if !ok {
					continue
				}
				if node.ObjectMeta.Labels[nodePortSelector["key"]] != nodePortSelector["value"] {
					continue
				}

			}
			nodeIP, nodeIP6 := lib.GetIPFromNode(node)
			var atype string
			var serverIP avimodels.IPAddr
			if v4enabled && v4Family && nodeIP != "" {
				v4ServerCount++
				atype = "V4"
				serverIP = avimodels.IPAddr{Type: &atype, Addr: &nodeIP}
			} else if v6enabled && v6Family && nodeIP6 != "" {
				v6ServerCount++
				atype = "V6"
				serverIP = avimodels.IPAddr{Type: &atype, Addr: &nodeIP6}
			} else {
				continue
			}
			server := AviPoolMetaServer{Ip: serverIP}
			poolMeta = append(poolMeta, server)
		}
	}

	if len(poolMeta) == 0 {
		utils.AviLog.Warnf("key: %s, msg: no servers for port: %v (%v)", key, poolNode.Port, poolNode.PortName)
	} else {
		if v4Family && v4ServerCount == 0 {
			utils.AviLog.Warnf("key: %s, msg: expected IPv4 servers but found none for port %v (%v)", key, poolNode.Port, poolNode.PortName)
		}
		if v6Family && v6ServerCount == 0 {
			utils.AviLog.Warnf("key: %s, msg: expected IPv6 servers but found none for port %v (%v)", key, poolNode.Port, poolNode.PortName)
		}
		utils.AviLog.Infof("key: %s, msg: servers for port: %v (%v), are: %v", key, poolNode.Port, poolNode.PortName, utils.Stringify(poolMeta))
	}

	return poolMeta
}

// an endpoint can be unique on four constraints
type endpointKey struct {
	address     string
	port        int32
	protocol    v1.Protocol
	addressType discovery.AddressType
}

func PopulateServers(poolNode *AviPoolNode, ns string, serviceName string, ingress bool, key string) []AviPoolMetaServer {

	// Find the servers that match the port.
	if ingress {
		// If it's an ingress case, check if the service of type clusterIP or not.
		found, _ := objects.SharedClusterIpLister().Get(ns + "/" + serviceName)
		if !found {
			utils.AviLog.Warnf("key: %s, msg: service pointed by the ingress object is not found in ClusterIP store", key)
			return nil
		}
	}
	svcObj, err := utils.GetInformers().ServiceInformer.Lister().Services(ns).Get(serviceName)
	if err != nil {
		utils.AviLog.Warnf("key: %s, msg: error while retrieving service: %s", key, err)
		return nil
	}
	ipFamily := lib.GetIPFamily()
	v4enabled := ipFamily == "V4" || ipFamily == "V4_V6"
	v6enabled := ipFamily == "V6" || ipFamily == "V4_V6"
	v4Family := false
	v6Family := false
	v4ServerCount := 0
	v6ServerCount := 0
	if len(svcObj.Spec.IPFamilies) == 2 {
		v4Family = true
		v6Family = true
	} else if len(svcObj.Spec.IPFamilies) > 0 && svcObj.Spec.IPFamilies[0] == "IPv6" {
		v6Family = true
	} else {
		v4Family = true
	}
	var pool_meta []AviPoolMetaServer
	if lib.AKOControlConfig().GetEndpointSlicesEnabled() {
		epSliceIntList, err := utils.GetInformers().EpSlicesInformer.Informer().GetIndexer().ByIndex(discovery.LabelServiceName, ns+"/"+serviceName)
		if err != nil {
			utils.AviLog.Warnf("key: %s, msg: error while retrieving endpointsice: %s", key, err)
			return nil
		}
		// create a map for deduplication of addresses
		uniqueEndpoints := map[endpointKey]struct{}{}
		for _, epSliceInt := range epSliceIntList {
			epSlice, isEpSliceClass := epSliceInt.(*discovery.EndpointSlice)
			if !isEpSliceClass {
				// not an epslice. continue
				utils.AviLog.Warnf("key: %s, msg: invalid endpointslice object", key)
				continue
			}
			port_match := false
			var epProtocol v1.Protocol
			for _, epp := range epSlice.Ports {
				if (epp.Name != nil && poolNode.PortName == *epp.Name) || (epp.Port != nil && poolNode.TargetPort.IntVal == *epp.Port) {
					port_match = true
					poolNode.Port = *epp.Port
					epProtocol = *epp.Protocol
					break
				}
			}
			if len(epSliceIntList) == 1 && len(epSlice.Ports) == 1 {
				port_match = true
				poolNode.Port = *epSlice.Ports[0].Port
				epProtocol = *epSlice.Ports[0].Protocol
			}
			if !port_match {
				continue
			}
			var atype string
			utils.AviLog.Infof("key: %s, msg: found port match for port %v", key, poolNode.Port)
			for _, addr := range epSlice.Endpoints {
				// use only first address. Refer to: https://issue.k8s.io/106267
				ip := addr.Addresses[0]
				epKey := endpointKey{
					address:     ip,
					port:        poolNode.Port,
					protocol:    epProtocol,
					addressType: epSlice.AddressType,
				}
				if _, ok := uniqueEndpoints[epKey]; ok {
					// found duplicate continue
					utils.AviLog.Debugf("key: %s, msg: found duplicate endpoint %v", key, epKey)
					continue
				}
				uniqueEndpoints[epKey] = struct{}{}
				if v4enabled && v4Family && utils.IsV4(ip) {
					v4ServerCount++
					atype = "V4"
				} else if v6enabled && v6Family && k8net.IsIPv6String(ip) {
					v6ServerCount++
					atype = "V6"
				} else {
					continue
				}
				// check condition
				enabled := enableServer(addr.Conditions)
				a := avimodels.IPAddr{Type: &atype, Addr: &ip}
				server := AviPoolMetaServer{Ip: a, Enabled: enabled}
				if addr.NodeName != nil {
					server.ServerNode = *addr.NodeName
				}
				pool_meta = append(pool_meta, server)
			}
		}
		poolNode.GracefulShutdownTimeout = lib.AKOControlConfig().GetGracefulShutdownTimeout()
	} else {
		epObj, err := utils.GetInformers().EpInformer.Lister().Endpoints(ns).Get(serviceName)
		if err != nil {
			utils.AviLog.Warnf("key: %s, msg: error while retrieving endpoints: %s", key, err)
			return nil
		}
		for _, ss := range epObj.Subsets {
			port_match := false
			for _, epp := range ss.Ports {
				if poolNode.PortName == epp.Name || int32(poolNode.TargetPort.IntValue()) == epp.Port {
					port_match = true
					poolNode.Port = epp.Port
					break
				}
			}
			if len(ss.Ports) == 1 && len(epObj.Subsets) == 1 {
				// If it's just a single port then we make that as the server port.
				port_match = true
				poolNode.Port = ss.Ports[0].Port
			}
			if !port_match {
				continue
			}
			var atype string
			utils.AviLog.Infof("key: %s, msg: found port match for port %v", key, poolNode.Port)
			for _, addr := range ss.Addresses {
				ip := addr.IP
				if v4enabled && v4Family && utils.IsV4(addr.IP) {
					v4ServerCount++
					atype = "V4"
				} else if v6enabled && v6Family && k8net.IsIPv6String(addr.IP) {
					v6ServerCount++
					atype = "V6"
				} else {
					continue
				}
				a := avimodels.IPAddr{Type: &atype, Addr: &ip}
				server := AviPoolMetaServer{Ip: a}
				if addr.NodeName != nil {
					server.ServerNode = *addr.NodeName
				}
				pool_meta = append(pool_meta, server)
			}
		}

	}
	if len(pool_meta) == 0 {
		utils.AviLog.Warnf("key: %s, msg: no servers for port: %v", key, poolNode.Port)
	} else {
		if v4Family && v4ServerCount == 0 {
			utils.AviLog.Warnf("key: %s, msg: expected IPv4 servers but found none for port %v", key, poolNode.Port)
		}
		if v6Family && v6ServerCount == 0 {
			utils.AviLog.Warnf("key: %s, msg: expected IPv6 servers but found none for port %v", key, poolNode.Port)
		}
		utils.AviLog.Infof("key: %s, msg: servers for port: %v , are: %v", key, poolNode.Port, utils.Stringify(pool_meta))
	}
	return pool_meta
}

func enableServer(condition discovery.EndpointConditions) *bool {
	var ready, terminating bool
	enabled := new(bool)
	*enabled = true
	if condition.Ready != nil {
		ready = *condition.Ready
	}
	if condition.Terminating != nil {
		terminating = *condition.Terminating
	}
	// giving benefit of doubt and not marking the server disabled if terminating state is not set
	if !ready || terminating {
		*enabled = false
	}
	return enabled
}

func PopulateServersForMultiClusterIngress(poolNode *AviPoolNode, ns, cluster, serviceNamespace, serviceName string, key string) []AviPoolMetaServer {

	var servers []AviPoolMetaServer
	svcName := generateMultiClusterKey(cluster, serviceNamespace, serviceName)
	success, siNames := objects.SharedMultiClusterIngressSvcLister().MultiClusterIngressMappings(ns).GetSvcToSI(svcName)
	if !success {
		utils.AviLog.Warnf("key: %s, msg: failed to get service imports mapped to service with name: %v", key, svcName)
		return servers
	}

	for _, siName := range siNames {
		serviceImport, err := utils.GetInformers().ServiceImportInformer.Lister().ServiceImports(ns).Get(siName)
		if err != nil {
			utils.AviLog.Warnf("key: %s, msg: failed to get service imports with name: %v", key, siName)
			continue
		}
		for _, backend := range serviceImport.Spec.SvcPorts {
			for _, ep := range backend.Endpoints {
				addr := ep.IP
				var addrType string
				if utils.IsV4(addr) {
					addrType = "V4"
				}
				Ip := avimodels.IPAddr{
					Addr: &addr,
					Type: &addrType,
				}
				server := AviPoolMetaServer{
					Port: ep.Port,
					Ip:   Ip,
				}
				servers = append(servers, server)
			}
		}
	}
	return servers
}

func (o *AviObjectGraph) BuildL4LBGraph(namespace string, svcName string, key string) {
	o.Lock.Lock()
	defer o.Lock.Unlock()
	var VsNode *AviVsNode
	svcObj, err := utils.GetInformers().ServiceInformer.Lister().Services(namespace).Get(svcName)
	if err != nil {
		utils.AviLog.Warnf("key: %s, msg: error in obtaining the object for service: %s", key, svcName)
		return
	}
	VsNode = o.ConstructAviL4VsNode(svcObj, key)
	if VsNode != nil {
		o.ConstructAviL4PolPoolNodes(svcObj, VsNode, key)
		o.AddModelNode(VsNode)
		utils.AviLog.Infof("key: %s, msg: checksum  for AVI VS object %v", key, VsNode.GetCheckSum())
		utils.AviLog.Infof("key: %s, msg: computed Graph checksum for VS is: %v", key, o.GraphChecksum)
	}
}

func getAutoFQDNForService(svcNamespace, svcName string) string {
	var fqdn string
	subDomains := GetDefaultSubDomain()

	// honour defaultSubDomain from values.yaml if specified.
	defaultSubDomain := lib.GetDomain()
	if defaultSubDomain != "" && utils.HasElem(subDomains, defaultSubDomain) {
		subDomains = []string{defaultSubDomain}
	}

	if subDomains == nil {
		// return empty string
		return fqdn
	}
	// subDomains[0] would either have the defaultSubDomain value
	// or would default to the first dns subdomain it gets from the dns profile
	subdomain := subDomains[0]
	if strings.HasPrefix(subDomains[0], ".") {
		subdomain = strings.Replace(subDomains[0], ".", "", 1)
	}

	//check each label for RFC 1035
	if !lib.CheckRFC1035(svcName) {
		lib.CorrectLabelToSatisfyRFC1035(&svcName, lib.FQDN_SVCNAME_PREFIX)
	}

	if !lib.CheckRFC1035(svcNamespace) {
		lib.CorrectLabelToSatisfyRFC1035(&svcNamespace, lib.FQDN_SVCNAMESPACE_PREFIX)
	}

	if lib.GetL4FqdnFormat() == lib.AutoFQDNDefault {
		// Generate the FQDN based on the logic: <svc_name>.<namespace>.<sub-domain>
		fqdn = svcName + "." + svcNamespace + "." + subdomain

	} else if lib.GetL4FqdnFormat() == lib.AutoFQDNFlat {

		// check and shorten the length of name and namespace to follow RFC 1035.
		svcName, svcNamespace := lib.CheckAndShortenLabelToFollowRFC1035(svcName, svcNamespace)

		// Generate the FQDN based on the logic: <svc_name>-<namespace>.<sub-domain>
		fqdn = svcName + "-" + svcNamespace + "." + subdomain
	}

	return fqdn
}

func GetDefaultSubDomain() []string {
	cache := avicache.SharedAviObjCache()
	cloud, ok := cache.CloudKeyCache.AviCacheGet(utils.CloudName)
	if !ok || cloud == nil {
		utils.AviLog.Warnf("Cloud object %s not found in cache", utils.CloudName)
		return nil
	}
	cloudProperty, ok := cloud.(*avicache.AviCloudPropertyCache)
	if !ok {
		utils.AviLog.Warnf("Cloud property object not found")
		return nil
	}

	if len(cloudProperty.NSIpamDNS) > 0 {
		sort.Strings(cloudProperty.NSIpamDNS)
	} else {
		return nil
	}
	return cloudProperty.NSIpamDNS
}

func getL4InfraSetting(key, namespace string, svc *corev1.Service, advl4GWClassName *string) (*akov1beta1.AviInfraSetting, error) {
	var err error
	var infraSetting *akov1beta1.AviInfraSetting

	if lib.UseServicesAPI() && advl4GWClassName != nil {
		gwClass, err := lib.AKOControlConfig().SvcAPIInformers().GatewayClassInformer.Lister().Get(*advl4GWClassName)
		if err != nil {
			utils.AviLog.Warnf("key: %s, msg: Unable to get corresponding GatewayClass %s", key, err.Error())
			return nil, err
		}
		if gwClass.Spec.ParametersRef != nil && gwClass.Spec.ParametersRef.Group == lib.AkoGroup && gwClass.Spec.ParametersRef.Kind == lib.AviInfraSetting {
			infraSetting, err = lib.AKOControlConfig().CRDInformers().AviInfraSettingInformer.Lister().Get(gwClass.Spec.ParametersRef.Name)
			if err != nil {
				utils.AviLog.Warnf("key: %s, msg: Unable to get corresponding AviInfraSetting via GatewayClass %s", key, err.Error())
				return nil, err
			}

		}
	} else if svc != nil {
		if infraSettingAnnotation, ok := svc.GetAnnotations()[lib.InfraSettingNameAnnotation]; ok && infraSettingAnnotation != "" {
			infraSetting, err = lib.AKOControlConfig().CRDInformers().AviInfraSettingInformer.Lister().Get(infraSettingAnnotation)
			if err != nil {
				utils.AviLog.Warnf("key: %s, msg: Unable to get corresponding AviInfraSetting via annotation %s", key, err.Error())
				return nil, err
			}
		}
	}

	if infraSetting != nil {
		if infraSetting.Status.Status != lib.StatusAccepted {
			utils.AviLog.Warnf("key: %s, msg: Referred AviInfraSetting %s is invalid", key, infraSetting.Name)
			return nil, fmt.Errorf("referred AviInfraSetting %s is invalid", infraSetting.Name)
		}
		return infraSetting, nil
	}

	//return namespace InfraSetting if global infraSetting is not present
	return getNamespaceAviInfraSetting(key, namespace)
}

func getL4Rule(key string, svc *corev1.Service) (*akov1alpha2.L4Rule, error) {
	var err error
	var l4Rule *akov1alpha2.L4Rule

	l4RuleName, ok := svc.GetAnnotations()[lib.L4RuleAnnotation]
	if !ok {
		// Annotation not present. Return error as nil in that case.
		return nil, nil
	}
	// fetch namespace and name from l4 annotation
	namespace, _, name := lib.ExtractTypeNameNamespace(l4RuleName)

	if namespace == "" {
		namespace = svc.GetNamespace()
	}
	l4Rule, err = lib.AKOControlConfig().CRDInformers().L4RuleInformer.Lister().L4Rules(namespace).Get(name)
	if err != nil {
		utils.AviLog.Warnf("key: %s, msg: Unable to get corresponding L4Rule via annotation. Err: %s", key, err.Error())
		return nil, err
	}

	if l4Rule != nil && l4Rule.Status.Status != lib.StatusAccepted {
		return nil, fmt.Errorf("referred L4Rule %s is invalid", l4Rule.Name)
	}
	svcPortsLen := len(svc.Spec.Ports)
	listenerPortProtoMap := make(map[string]bool)
	if len(l4Rule.Spec.Services) != 0 {
		if len(l4Rule.Spec.Services) != svcPortsLen {
			err := fmt.Errorf("Number of port definitions in %s l4rule listener spec does not match with the port definitons in %s service", l4RuleName, svc.Name)
			utils.AviLog.Warnf("key: %s, msg: %s", key, err.Error())
			return nil, err
		}
		for _, l4Svc := range l4Rule.Spec.Services {
			if *l4Svc.EnableSsl && svcPortsLen > 1 {
				err := fmt.Errorf("Port %d requires enabling SSL for L4 but there are multiple ports defined in %s service definition", int(*l4Svc.Port), svc.Name)
				utils.AviLog.Warnf("key: %s, msg: %s", key, err.Error())
				return nil, err
			}
			key := strconv.Itoa(int(*l4Svc.Port)) + *l4Svc.Protocol
			listenerPortProtoMap[key] = *l4Svc.EnableSsl
		}

		for _, port := range svc.Spec.Ports {
			portProtocol := strconv.Itoa(int(port.Port)) + fmt.Sprint(port.Protocol)
			if _, ok := listenerPortProtoMap[portProtocol]; !ok {
				err := fmt.Errorf("Port %d defined in %s service definition is not present in %s l4rule listener spec", int(port.Port), svc.Name, l4RuleName)
				utils.AviLog.Warnf("key: %s, msg: %s", key, err.Error())
				return nil, err
			}
		}
	}

	utils.AviLog.Debugf("key: %s, Got L4Rule %v", key, l4Rule)
	return l4Rule, nil
}

func buildWithL4Rule(key string, vs *AviVsNode, l4Rule *akov1alpha2.L4Rule) {

	if l4Rule == nil {
		return
	}
	isSSLEnabled := false
	for _, svc := range l4Rule.Spec.Services {
		if *svc.EnableSsl {
			isSSLEnabled = true
		}
	}
	if isSSLEnabled && vs.NetworkProfile != utils.DEFAULT_TCP_NW_PROFILE {
		utils.AviLog.Warnf("key: %s, msg: L4Rule %s cannot be applied to the service as network profile is not equal to %s", key, l4Rule.Name, utils.DEFAULT_TCP_NW_PROFILE)
		return
	}
	copier.Copy(vs, &l4Rule.Spec)
	if isSSLEnabled && *l4Rule.Spec.ApplicationProfileRef == utils.DEFAULT_L4_APP_PROFILE {
		defaultAppProfile := utils.DEFAULT_L4_SSL_APP_PROFILE
		vs.ApplicationProfileRef = &defaultAppProfile
	}
	vs.AviVsNodeCommonFields.ConvertToRef()
	vs.AviVsNodeGeneratedFields.ConvertToRef()

	utils.AviLog.Debugf("key: %s, msg: Applied L4Rule %s configuration over VS %s", key, l4Rule.Name, vs.Name)
}

func buildPoolWithL4Rule(key string, pool *AviPoolNode, l4Rule *akov1alpha2.L4Rule) {

	if l4Rule == nil {
		return
	}

	index := -1
	for i, poolProperty := range l4Rule.Spec.BackendProperties {
		if *poolProperty.Port == int(pool.Port) &&
			*poolProperty.Protocol == pool.Protocol {
			index = i
			break
		}
	}
	if index == -1 {
		utils.AviLog.Warnf("key: %s, msg: L4Rule %s doesn't match any pools present.", key, l4Rule.Name)
		return
	}
	copier.Copy(pool, l4Rule.Spec.BackendProperties[index])

	pool.AviPoolCommonFields.ConvertToRef()
	pool.AviPoolGeneratedFields.ConvertToRef()

	utils.AviLog.Debugf("key: %s, msg: Applied L4Rule %s configuration over Pool %s", key, l4Rule.Name, pool.Name)
}

// In case the VS has services that are a mix of TCP and UDP/SCTP sockets,
// we create the VS with global network profile TCP Proxy or Fast Path based on license,
// and override required services with UDP Fast Path or SCTP proxy. Having a separate
// internally used network profile (MIXED_NET_PROFILE) helps ensure PUT calls
// on existing VSes.
func getNetworkProfile(isSCTP, isTCP, isUDP bool) string {
	if isSCTP && !isTCP && !isUDP {
		return utils.SYSTEM_SCTP_PROXY
	}
	if isTCP && !isUDP && !isSCTP {
		license := lib.AKOControlConfig().GetLicenseType()
		if license == lib.LicenseTypeEnterprise || license == lib.LicenseTypeEnterpriseCloudServices {
			return utils.DEFAULT_TCP_NW_PROFILE
		}
		return utils.TCP_NW_FAST_PATH
	}
	if isUDP && !isTCP && !isSCTP {
		return utils.SYSTEM_UDP_FAST_PATH
	}
	return utils.MIXED_NET_PROFILE
}

// Delete Old Model when Tenant values changes in Namespace annotation
func DeleteStaleTenantModelData(objName, namespace, key, tenant, objType string) {
	oldTenant := objects.SharedNamespaceTenantLister().GetTenantInNamespace(namespace + "/" + objName)
	if oldTenant == "" {
		oldTenant = lib.GetTenant()
	}
	if oldTenant == tenant {
		return
	}
	// Old model in oldTenant can be safely deleted here
	oldModelName := lib.GetModelName(oldTenant, lib.Encode(lib.GetNamePrefix()+namespace+"-"+objName, objType))
	found, _ := objects.SharedAviGraphLister().Get(oldModelName)
	if !found {
		utils.AviLog.Debugf("key: %s, msg: Model not found in the Graph Lister, model: %s", key, oldModelName)
		return
	}
	utils.AviLog.Infof("key: %s, msg: Deleting old model data, model: %s", key, oldModelName)
	objects.SharedAviGraphLister().Save(oldModelName, nil)
	PublishKeyToRestLayer(oldModelName, key, utils.SharedWorkQueue().GetQueueByName(utils.GraphLayer))
}
