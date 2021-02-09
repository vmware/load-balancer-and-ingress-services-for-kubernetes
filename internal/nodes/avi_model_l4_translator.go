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
	"strings"

	avicache "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	"github.com/avinetworks/sdk/go/models"
	avimodels "github.com/avinetworks/sdk/go/models"
	corev1 "k8s.io/api/core/v1"
)

func (o *AviObjectGraph) ConstructAviL4VsNode(svcObj *corev1.Service, key string) *AviVsNode {
	var avi_vs_meta *AviVsNode
	var fqdns []string
	autoFQDN := true
	if lib.GetL4FqdnFormat() == 3 {
		autoFQDN = false
	}
	annotations := svcObj.GetAnnotations()
	extDNS, ok := annotations[lib.ExternalDNSAnnotation]
	if ok {
		autoFQDN = false
		fqdns = append(fqdns, extDNS)
	}
	vsName := lib.GetL4VSName(svcObj.ObjectMeta.Name, svcObj.ObjectMeta.Namespace)
	subDomains := GetDefaultSubDomain()
	if subDomains != nil && autoFQDN {
		var fqdn string
		// honour defaultSubDomain from values.yaml if specified
		defaultSubDomain := lib.GetDomain()
		if defaultSubDomain != "" && utils.HasElem(subDomains, defaultSubDomain) {
			subDomains = []string{defaultSubDomain}
		}

		// subDomains[0] would either have the defaultSubDomain value
		// or would default to the first dns subdomain it gets from the dns profile
		subdomain := subDomains[0]
		if strings.HasPrefix(subDomains[0], ".") {
			subdomain = strings.Replace(subDomains[0], ".", "", -1)
		}
		if lib.GetL4FqdnFormat() == 1 {
			// Generate the FQDN based on the logic: <svc_name>.<namespace>.<sub-domain>
			fqdn = svcObj.Name + "." + svcObj.ObjectMeta.Namespace + "." + subdomain
		} else if lib.GetL4FqdnFormat() == 2 {
			// Generate the FQDN based on the logic: <svc_name>-<namespace>.<sub-domain>
			fqdn = svcObj.Name + "-" + svcObj.ObjectMeta.Namespace + "." + subdomain
		}

		fqdns = append(fqdns, fqdn)
	}
	avi_vs_meta = &AviVsNode{
		Name:     vsName,
		Tenant:   lib.GetTenant(),
		EastWest: false,
		ServiceMetadata: avicache.ServiceMetadataObj{
			NamespaceServiceName: []string{svcObj.ObjectMeta.Namespace + "/" + svcObj.ObjectMeta.Name},
			HostNames:            fqdns,
		},
	}

	if lib.GetSEGName() != lib.DEFAULT_SE_GROUP {
		avi_vs_meta.ServiceEngineGroup = lib.GetSEGName()
	}
	vrfcontext := lib.GetVrf()
	avi_vs_meta.VrfContext = vrfcontext

	isTCP := false
	var portProtocols []AviPortHostProtocol
	for _, port := range svcObj.Spec.Ports {
		pp := AviPortHostProtocol{Port: int32(port.Port), Protocol: fmt.Sprint(port.Protocol), Name: port.Name}
		portProtocols = append(portProtocols, pp)
		if port.Protocol == "" || port.Protocol == utils.TCP {
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

	vsVipName := lib.GetL4VSVipName(svcObj.ObjectMeta.Name, svcObj.ObjectMeta.Namespace)
	vsVipNode := &AviVSVIPNode{
		Name:       vsVipName,
		Tenant:     lib.GetTenant(),
		FQDNs:      fqdns,
		EastWest:   false,
		VrfContext: vrfcontext,
	}

	if svcObj.Spec.LoadBalancerIP != "" {
		vsVipNode.IPAddress = svcObj.Spec.LoadBalancerIP
	}

	avi_vs_meta.VSVIPRefs = append(avi_vs_meta.VSVIPRefs, vsVipNode)
	utils.AviLog.Infof("key: %s, msg: created vs object: %s", key, utils.Stringify(avi_vs_meta))
	return avi_vs_meta
}

func (o *AviObjectGraph) ConstructAviL4PolPoolNodes(svcObj *corev1.Service, vsNode *AviVsNode, key string) {
	var l4Policies []*AviL4PolicyNode
	var portPoolSet []AviHostPathPortPoolPG
	for _, portProto := range vsNode.PortProto {
		filterPort := portProto.Port
		poolNode := &AviPoolNode{Name: lib.GetL4PoolName(vsNode.Name, filterPort), Tenant: lib.GetTenant(), Protocol: portProto.Protocol, PortName: portProto.Name}
		poolNode.VrfContext = lib.GetVrf()

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

		pool_ref := fmt.Sprintf("/api/pool?name=%s", poolNode.Name)
		portPool := AviHostPathPortPoolPG{Port: uint32(filterPort), Pool: pool_ref, Protocol: portProto.Protocol}
		portPoolSet = append(portPoolSet, portPool)

		vsNode.PoolRefs = append(vsNode.PoolRefs, poolNode)
		utils.AviLog.Infof("key: %s, msg: evaluated L4 pool values :%v", key, utils.Stringify(poolNode))

		poolNode.CalculateCheckSum()
		o.AddModelNode(poolNode)
		o.GraphChecksum = o.GraphChecksum + poolNode.GetCheckSum()
	}
	l4policyNode := &AviL4PolicyNode{Name: vsNode.Name, Tenant: lib.GetTenant(), PortPool: portPoolSet}
	l4Policies = append(l4Policies, l4policyNode)
	l4policyNode.CalculateCheckSum()
	o.GraphChecksum = o.GraphChecksum + l4policyNode.GetCheckSum()
	vsNode.L4PolicyRefs = l4Policies
	utils.AviLog.Infof("key: %s, msg: evaluated L4 pool policies :%v", key, utils.Stringify(vsNode.L4PolicyRefs))

}

func PopulateServersForNPL(poolNode *AviPoolNode, ns string, serviceName string, ingress bool, key string) []AviPoolMetaServer {
	if ingress {
		found, _ := objects.SharedClusterIpLister().Get(ns + "/" + serviceName)
		if !found {
			utils.AviLog.Warnf("key: %s, msg: service pointed by the ingress object is not found in ClusterIP store", key)
			return nil
		}
	}
	pods := lib.GetPodsFromService(ns, serviceName)

	var poolMeta []AviPoolMetaServer
	svcObj, err := utils.GetInformers().ServiceInformer.Lister().Services(ns).Get(serviceName)
	if err != nil {
		utils.AviLog.Warnf("key: %s, msg: error in obtaining the object for service: %s", key, serviceName)
		return poolMeta
	}

	targetPorts := make(map[int]bool)
	for _, port := range svcObj.Spec.Ports {
		if port.Name != poolNode.PortName && len(svcObj.Spec.Ports) != 1 {
			// continue only if port name does not match and it is multiport svcobj
			continue
		}
		targetPorts[port.TargetPort.IntValue()] = true
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
			if utils.IsV4(a.NodeIP) {
				atype = "V4"
			} else {
				atype = "V6"
			}
			server := AviPoolMetaServer{
				Port: int32(a.NodePort),
				Ip: models.IPAddr{
					Addr: &a.NodeIP,
					Type: &atype,
				}}
			poolMeta = append(poolMeta, server)
		}
	}
	utils.AviLog.Infof("key: %s, msg: servers for port: %v, are: %v", key, poolNode.Port, utils.Stringify(poolMeta))
	return poolMeta
}

func PopulateServersForNodePort(poolNode *AviPoolNode, ns string, serviceName string, ingress bool, key string) []AviPoolMetaServer {

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
			addresses := node.Status.Addresses
			ip := ""
			var atype string
			for _, address := range addresses {
				if address.Type == corev1.NodeInternalIP {
					ip = address.Address
				}
			}
			if ip == "" {
				utils.AviLog.Warnf("key: %s,msg: NodeInternalIP not found for node: %s", key, node.Name)
				return nil
			}
			if utils.IsV4(ip) {
				atype = "V4"
			} else {
				atype = "V6"
			}

			a := avimodels.IPAddr{Type: &atype, Addr: &ip}
			server := AviPoolMetaServer{Ip: a}
			poolMeta = append(poolMeta, server)
		}
	}

	return poolMeta
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
	epObj, err := utils.GetInformers().EpInformer.Lister().Endpoints(ns).Get(serviceName)
	if err != nil {
		utils.AviLog.Warnf("key: %s, msg: error while retrieving endpoints: %s", key, err)
		return nil
	}
	var pool_meta []AviPoolMetaServer
	for _, ss := range epObj.Subsets {
		port_match := false
		for _, epp := range ss.Ports {
			if poolNode.PortName == epp.Name || poolNode.TargetPort == epp.Port {
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
					atype = "V4"
				} else {
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
	o.ConstructAviL4PolPoolNodes(svcObj, VsNode, key)
	o.AddModelNode(VsNode)
	VsNode.CalculateCheckSum()
	o.GraphChecksum = o.GraphChecksum + VsNode.GetCheckSum()
	utils.AviLog.Infof("key: %s, msg: checksum  for AVI VS object %v", key, VsNode.GetCheckSum())
	utils.AviLog.Infof("key: %s, msg: computed Graph checksum for VS is: %v", key, o.GraphChecksum)
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
