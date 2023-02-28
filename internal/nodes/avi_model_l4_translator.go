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

	avicache "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	akov1alpha1 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/apis/ako/v1alpha1"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	"github.com/vmware/alb-sdk/go/models"
	avimodels "github.com/vmware/alb-sdk/go/models"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/sets"
)

func (o *AviObjectGraph) ConstructAviL4VsNode(svcObj *corev1.Service, key string) *AviVsNode {
	var avi_vs_meta *AviVsNode
	var fqdns []string
	autoFQDN := true
	if lib.GetL4FqdnFormat() == lib.AutoFQDNDisabled {
		autoFQDN = false
	}

	if extDNS, ok := svcObj.Annotations[lib.ExternalDNSAnnotation]; ok && autoFQDN {
		fqdns = append(fqdns, extDNS)
	}

	subDomains := GetDefaultSubDomain()
	if subDomains != nil && autoFQDN {
		if fqdn := getAutoFQDNForService(svcObj.Namespace, svcObj.Name); fqdn != "" {
			fqdns = append(fqdns, fqdn)
		}
	}

	vsName := lib.GetL4VSName(svcObj.ObjectMeta.Name, svcObj.ObjectMeta.Namespace)
	avi_vs_meta = &AviVsNode{
		Name:   vsName,
		Tenant: lib.GetTenant(),
		ServiceMetadata: lib.ServiceMetadataObj{
			NamespaceServiceName: []string{svcObj.ObjectMeta.Namespace + "/" + svcObj.ObjectMeta.Name},
			HostNames:            fqdns,
		},
		ServiceEngineGroup: lib.GetSEGName(),
		EnableRhi:          proto.Bool(lib.GetEnableRHI()),
	}

	vrfcontext := lib.GetVrf()
	t1lr := lib.SharedWCPLister().GetT1LrForNamespace(svcObj.Namespace)
	if t1lr != "" {
		vrfcontext = ""
	} else {
		avi_vs_meta.VrfContext = vrfcontext
	}
	avi_vs_meta.AviMarkers = lib.PopulateL4VSNodeMarkers(svcObj.ObjectMeta.Namespace, svcObj.ObjectMeta.Name)
	isTCP, isSCTP := false, false
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
		}
	}
	avi_vs_meta.PortProto = portProtocols

	if appProfile, ok := svcObj.GetAnnotations()[lib.LBSvcAppProfileAnnotation]; ok && appProfile != "" {
		avi_vs_meta.ApplicationProfile = appProfile
	} else {
		// Default case
		avi_vs_meta.ApplicationProfile = utils.DEFAULT_L4_APP_PROFILE
	}
	if !isTCP {
		if isSCTP {
			avi_vs_meta.NetworkProfile = utils.SYSTEM_SCTP_PROXY
		} else {
			avi_vs_meta.NetworkProfile = utils.SYSTEM_UDP_FAST_PATH
		}

	} else {
		license := lib.AKOControlConfig().GetLicenseType()

		if license == "ENTERPRISE" {
			avi_vs_meta.NetworkProfile = utils.DEFAULT_TCP_NW_PROFILE
		} else {
			avi_vs_meta.NetworkProfile = utils.TCP_NW_FAST_PATH
		}
	}

	vsVipName := lib.GetL4VSVipName(svcObj.ObjectMeta.Name, svcObj.ObjectMeta.Namespace)
	vsVipNode := &AviVSVIPNode{
		Name:        vsVipName,
		Tenant:      lib.GetTenant(),
		FQDNs:       fqdns,
		VrfContext:  vrfcontext,
		VipNetworks: lib.SharedWCPLister().GetNetworkForNamespace(svcObj.Namespace),
	}
	if t1lr != "" {
		vsVipNode.T1Lr = t1lr
	}

	if avi_vs_meta.EnableRhi != nil && *avi_vs_meta.EnableRhi {
		vsVipNode.BGPPeerLabels = lib.GetGlobalBgpPeerLabels()
	}

	// configures VS and VsVip nodes using infraSetting object (via CRD).
	if infraSetting, err := getL4InfraSetting(key, svcObj, nil); err == nil {
		buildWithInfraSetting(key, svcObj.Namespace, avi_vs_meta, vsVipNode, infraSetting)
	}

	if lib.HasSpecLoadBalancerIP(svcObj) {
		vsVipNode.IPAddress = svcObj.Spec.LoadBalancerIP
	} else if lib.HasLoadBalancerIPAnnotation(svcObj) {
		vsVipNode.IPAddress = svcObj.Annotations[lib.LoadBalancerIP]
	}

	avi_vs_meta.VSVIPRefs = append(avi_vs_meta.VSVIPRefs, vsVipNode)
	return avi_vs_meta
}

func (o *AviObjectGraph) ConstructAviL4PolPoolNodes(svcObj *corev1.Service, vsNode *AviVsNode, key string) {
	var l4Policies []*AviL4PolicyNode
	var portPoolSet []AviHostPathPortPoolPG

	infraSetting, err := getL4InfraSetting(key, svcObj, nil)
	if err != nil {
		utils.AviLog.Warnf("key: %s, msg: Error while fetching infrasetting for Gateway %s", key, err.Error())
		return
	}
	protocolSet := sets.NewString()
	for _, portProto := range vsNode.PortProto {
		filterPort := portProto.Port
		poolNode := &AviPoolNode{
			Name:       lib.GetL4PoolName(svcObj.ObjectMeta.Name, svcObj.ObjectMeta.Namespace, portProto.Protocol, filterPort),
			Tenant:     lib.GetTenant(),
			Protocol:   portProto.Protocol,
			PortName:   portProto.Name,
			Port:       portProto.Port,
			TargetPort: portProto.TargetPort,
			VrfContext: lib.GetVrf(),
		}
		if lib.IsIstioEnabled() {
			poolNode.UpdatePoolNodeForIstio()
		}
		protocolSet.Insert(portProto.Protocol)
		poolNode.NetworkPlacementSettings, _ = lib.GetNodeNetworkMap()
		t1lr := lib.SharedWCPLister().GetT1LrForNamespace(svcObj.Namespace)
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

		vsNode.PoolRefs = append(vsNode.PoolRefs, poolNode)
		utils.AviLog.Infof("key: %s, msg: evaluated L4 pool values :%v", key, utils.Stringify(poolNode))
	}

	l4policyNode := &AviL4PolicyNode{Name: vsNode.Name, Tenant: lib.GetTenant(), PortPool: portPoolSet}
	sort.Strings(protocolSet.List())
	protocols := strings.Join(protocolSet.List(), ",")
	l4policyNode.AviMarkers = lib.PopulateL4PolicysetMarkers(svcObj.ObjectMeta.Namespace, svcObj.ObjectMeta.Name, protocols)
	l4Policies = append(l4Policies, l4policyNode)
	vsNode.L4PolicyRefs = l4Policies
	//As pool naming covention changed for L4 pools marking flag, so that cksum will be changed
	vsNode.IsL4VS = true
	utils.AviLog.Infof("key: %s, msg: evaluated L4 pool policies :%v", key, utils.Stringify(vsNode.L4PolicyRefs))
}

func PopulateServersForNPL(poolNode *AviPoolNode, ns string, serviceName string, ingress bool, key string) []AviPoolMetaServer {
	ipFamily := lib.GetIPFamily()
	if ingress {
		found, _ := objects.SharedClusterIpLister().Get(ns + "/" + serviceName)
		if !found {
			utils.AviLog.Warnf("key: %s, msg: service pointed by the ingress object is not found in ClusterIP store", key)
			return nil
		}
	}
	pods, targetPort := lib.GetPodsFromService(ns, serviceName, poolNode.TargetPort)
	if len(pods) == 0 {
		utils.AviLog.Infof("key: %s, msg: got no Pod for Service %s", key, serviceName)
		return make([]AviPoolMetaServer, 0)
	}

	var poolMeta []AviPoolMetaServer

	for _, pod := range pods {
		var annotations []lib.NPLAnnotation
		found, obj := objects.SharedNPLLister().Get(ns + "/" + pod.Name)
		if !found {
			continue
		}
		annotations = obj.([]lib.NPLAnnotation)
		for _, a := range annotations {
			var atype string
			if utils.IsV4(a.NodeIP) {
				if ipFamily != "V4" {
					utils.AviLog.Infof("Skipping server %s, ipFamily is %s", a.NodeIP, ipFamily)
					continue
				}
				atype = "V4"
			} else {
				if ipFamily != "V6" {
					utils.AviLog.Infof("Skipping server %s, ipFamily is %s", a.NodeIP, ipFamily)
					continue
				}
				atype = "V6"
			}
			if (poolNode.TargetPort.Type == intstr.Int && a.PodPort == poolNode.TargetPort.IntValue()) ||
				a.PodPort == int(targetPort) {
				server := AviPoolMetaServer{
					Port: int32(a.NodePort),
					Ip: models.IPAddr{
						Addr: &a.NodeIP,
						Type: &atype,
					}}
				poolMeta = append(poolMeta, server)
			}
		}
	}
	utils.AviLog.Infof("key: %s, msg: servers for port: %v (%v), are: %v", key, poolNode.Port, poolNode.PortName, utils.Stringify(poolMeta))
	return poolMeta
}

func PopulateServersForNodePort(poolNode *AviPoolNode, ns string, serviceName string, ingress bool, key string) []AviPoolMetaServer {

	ipFamily := lib.GetIPFamily()
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
				return nil
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
			if ipFamily == "V4" {
				if nodeIP == "" {
					utils.AviLog.Warnf("key: %s,msg: NodeIP not found for node: %s", key, node.Name)
					return nil
				} else {
					atype = "V4"
					serverIP = avimodels.IPAddr{Type: &atype, Addr: &nodeIP}
				}
			} else {
				if nodeIP6 == "" {
					utils.AviLog.Warnf("key: %s,msg: NodeIP6 not found for node: %s", key, node.Name)
					return nil
				} else {
					atype = "V6"
					serverIP = avimodels.IPAddr{Type: &atype, Addr: &nodeIP6}
				}
			}

			server := AviPoolMetaServer{Ip: serverIP}
			poolMeta = append(poolMeta, server)
		}
	}

	return poolMeta
}

func PopulateServers(poolNode *AviPoolNode, ns string, serviceName string, ingress bool, key string) []AviPoolMetaServer {

	ipFamily := lib.GetIPFamily()
	// Find the servers that match the port.
	if ingress {
		// If it's an ingress case, check if the service of type clusterIP or not.
		found, _ := objects.SharedClusterIpLister().Get(ns + "/" + serviceName)
		if !found {
			utils.AviLog.Warnf("key: %s, msg: service pointed by the ingress object is not found in ClusterIP store", key)
			return nil
		}
	}
	epObj, err := utils.GetInformers().EpInformer.Lister().Endpoints(ns).Get(serviceName)
	if err != nil {
		utils.AviLog.Warnf("key: %s, msg: error while retrieving endpoints: %s", key, err)
		return nil
	}
	var pool_meta []AviPoolMetaServer
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
		if port_match {
			var atype string
			utils.AviLog.Infof("key: %s, msg: found port match for port %v", key, poolNode.Port)
			for _, addr := range ss.Addresses {

				ip := addr.IP
				if utils.IsV4(addr.IP) {
					if ipFamily != "V4" {
						utils.AviLog.Infof("Skipping server %s, ipFamily is %s", addr.IP, ipFamily)
						continue
					}
					atype = "V4"
				} else {
					if ipFamily != "V6" {
						utils.AviLog.Infof("Skipping server %s, ipFamily is %s", addr.IP, ipFamily)
						continue
					}
					atype = "V6"
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
	utils.AviLog.Infof("key: %s, msg: servers for port: %v, are: %v", key, poolNode.Port, utils.Stringify(pool_meta))
	return pool_meta
}

func PopulateServersForMultiClusterIngress(poolNode *AviPoolNode, ns, cluster, serviceNamespace, serviceName string, key string) []AviPoolMetaServer {

	ipFamily := lib.GetIPFamily()
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
					if ipFamily != "V4" {
						utils.AviLog.Infof("Skipping server %s, ipFamily is %s", addr, ipFamily)
						continue
					}
					addrType = "V4"
				} else {
					if ipFamily != "V6" {
						utils.AviLog.Infof("Skipping server %s, ipFamily is %s", addr, ipFamily)
						continue
					}
					addrType = "V6"
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

func getL4InfraSetting(key string, svc *corev1.Service, advl4GWClassName *string) (*akov1alpha1.AviInfraSetting, error) {
	var err error
	var infraSetting *akov1alpha1.AviInfraSetting

	if lib.UseServicesAPI() && advl4GWClassName != nil {
		gwClass, err := lib.AKOControlConfig().SvcAPIInformers().GatewayClassInformer.Lister().Get(*advl4GWClassName)
		if err != nil {
			utils.AviLog.Warnf("key: %s, msg: Unable to get corresponding GatewayClass %s", key, err.Error())
			return nil, err
		} else {
			if gwClass.Spec.ParametersRef != nil && gwClass.Spec.ParametersRef.Group == lib.AkoGroup && gwClass.Spec.ParametersRef.Kind == lib.AviInfraSetting {
				infraSetting, err = lib.AKOControlConfig().CRDInformers().AviInfraSettingInformer.Lister().Get(gwClass.Spec.ParametersRef.Name)
				if err != nil {
					utils.AviLog.Warnf("key: %s, msg: Unable to get corresponding AviInfraSetting via GatewayClass %s", key, err.Error())
					return nil, err
				}
			}
		}
	} else if infraSettingAnnotation, ok := svc.GetAnnotations()[lib.InfraSettingNameAnnotation]; ok && infraSettingAnnotation != "" {
		infraSetting, err = lib.AKOControlConfig().CRDInformers().AviInfraSettingInformer.Lister().Get(infraSettingAnnotation)
		if err != nil {
			utils.AviLog.Warnf("key: %s, msg: Unable to get corresponding AviInfraSetting via annotation %s", key, err.Error())
			return nil, err
		}
	}

	if infraSetting != nil && infraSetting.Status.Status != lib.StatusAccepted {
		utils.AviLog.Warnf("key: %s, msg: Referred AviInfraSetting %s is invalid", key, infraSetting.Name)
		return nil, fmt.Errorf("Referred AviInfraSetting %s is invalid", infraSetting.Name)
	}

	return infraSetting, nil
}
