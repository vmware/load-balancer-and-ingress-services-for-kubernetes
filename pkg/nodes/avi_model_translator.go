/*
* [2013] - [2019] Avi Networks Incorporated
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

	avimodels "github.com/avinetworks/sdk/go/models"
	"gitlab.eng.vmware.com/orion/akc/pkg/objects"
	"gitlab.eng.vmware.com/orion/container-lib/utils"
	corev1 "k8s.io/api/core/v1"
)

const (
	HTTP            = "HTTP"
	HeaderMethod    = ":method"
	HeaderAuthority = ":authority"
	HeaderScheme    = ":scheme"
	TLS             = "TLS"
	HTTPS           = "HTTPS"
	TCP             = "TCP"
)

func contains(s []int32, e int32) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func (o *AviObjectGraph) ConstructAviL4VsNode(svcObj *corev1.Service) *AviVsNode {
	var avi_vs_meta *AviVsNode

	// FQDN should come from the cloud. Modify
	avi_vs_meta = &AviVsNode{Name: svcObj.ObjectMeta.Name, Tenant: svcObj.ObjectMeta.Namespace,
		EastWest: false}
	var portProtocols []AviPortHostProtocol
	for _, port := range svcObj.Spec.Ports {
		pp := AviPortHostProtocol{Port: int32(port.Port), Protocol: TCP}
		portProtocols = append(portProtocols, pp)
	}
	avi_vs_meta.PortProto = portProtocols
	// Default case.
	avi_vs_meta.ApplicationProfile = "System-TCP"
	// For HTTP it's always System-TCP-Proxy.
	avi_vs_meta.NetworkProfile = "System-TCP-Proxy"

	return avi_vs_meta
}

func (o *AviObjectGraph) ConstructAviTCPPGPoolNodes(svcObj *corev1.Service, vsNode *AviVsNode) {
	var prevTCPModelPoolGroupNodes []*AviPoolGroupNode
	var prevTCPPoolGroupNodesInCache []utils.NamespaceName
	model_name := svcObj.ObjectMeta.Namespace + "/" + svcObj.ObjectMeta.Name
	found, aviModel := objects.SharedAviGraphLister().Get(model_name)
	if found && aviModel != nil {
		if len(aviModel.(*AviObjectGraph).GetAviVS()) == 1 {
			prevTCPModelPoolGroupNodes = aviModel.(*AviObjectGraph).GetAviVS()[0].TCPPoolGroupRefs
			utils.AviLog.Info.Printf("Evaluating TCP Pool Groups. The prevModel PGs are: %v", prevTCPModelPoolGroupNodes)
		}
	}
	cache := utils.SharedAviObjCache()
	vsKey := utils.NamespaceName{Namespace: svcObj.ObjectMeta.Namespace, Name: svcObj.ObjectMeta.Name}
	vs_cache, ok := cache.VsCache.AviCacheGet(vsKey)
	vs_cache_obj, ok := vs_cache.(*utils.AviVsCache)
	if ok {
		// There's a VS Cache - let's check the PGs
		if vs_cache_obj.PGKeyCollection != nil {
			prevTCPPoolGroupNodesInCache = vs_cache_obj.PGKeyCollection
		}
	}
	for _, portProto := range vsNode.PortProto {
		filterPort := portProto.Port
		pgNamePrefix := "tcp-" + fmt.Sprint(filterPort) + "-"

		var pgName string
		// Check if the NamePrefix exists or not
		for _, pgNodeNsName := range prevTCPPoolGroupNodesInCache {
			if strings.HasPrefix(pgNodeNsName.Name, pgNamePrefix) {
				pgName = pgNodeNsName.Name
			}
		}
		// Check if the PG name is present in the model cache.
		if pgName == "" {
			for _, pgNode := range prevTCPModelPoolGroupNodes {
				if strings.HasPrefix(pgNode.Name, pgNamePrefix) {
					pgName = pgNode.Name
				}
			}
		}
		// If the PGName was not found in the cache, generate the name
		if pgName == "" {
			pgName = o.generateRandomStringName(pgNamePrefix)
		}
		pgNode := &AviPoolGroupNode{Name: pgName, Tenant: svcObj.ObjectMeta.Namespace, Port: fmt.Sprint(filterPort)}
		// For TCP - the PG to Pool relationship is 1x1
		poolNode := &AviPoolNode{Name: "pool-" + pgName, Tenant: svcObj.ObjectMeta.Namespace, Port: filterPort, Protocol: "TCP"}

		if servers := o.populateServers(poolNode, svcObj.ObjectMeta.Namespace, svcObj.ObjectMeta.Name); servers != nil {
			poolNode.Servers = servers
		}
		pool_ref := fmt.Sprintf("/api/pool?name=%s", poolNode.Name)
		pgNode.Members = append(pgNode.Members, &avimodels.PoolGroupMember{PoolRef: &pool_ref})

		vsNode.PoolRefs = append(vsNode.PoolRefs, poolNode)
		utils.AviLog.Info.Printf("Evaluated TCP pool group values :%v", utils.Stringify(pgNode))
		utils.AviLog.Info.Printf("Evaluated TCP pool values :%v", utils.Stringify(poolNode))
		vsNode.TCPPoolGroupRefs = append(vsNode.TCPPoolGroupRefs, pgNode)
		pgNode.CalculateCheckSum()
		poolNode.CalculateCheckSum()
		o.GraphChecksum = o.GraphChecksum + pgNode.GetCheckSum()
		o.GraphChecksum = o.GraphChecksum + poolNode.GetCheckSum()
	}
}

func (o *AviObjectGraph) populateServers(poolNode *AviPoolNode, ns string, serviceName string) []AviPoolMetaServer {
	// Find the servers that match the port.
	epObj, err := utils.GetInformers().EpInformer.Lister().Endpoints(ns).Get(serviceName)
	if err != nil {
		utils.AviLog.Info.Printf("Error while retrieving endpoints for Svc :%v in namespace :%s", serviceName, ns)
		return nil
	}
	//TODO: The POD based subsets will be handled subsequently.
	var pool_meta []AviPoolMetaServer
	for _, ss := range epObj.Subsets {
		//var epp_port int32
		port_match := false
		for _, epp := range ss.Ports {
			if (int32(poolNode.Port) == epp.Port) || (poolNode.Name == epp.Name) {
				port_match = true
				//epp_port = epp.Port
				break
			}
		}
		if port_match {
			var atype string
			utils.AviLog.Info.Printf("Found Port Match for port %v, for service: %v", poolNode.Port, serviceName)
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
	utils.AviLog.Info.Printf("Servers for port: %v, for service: %v are: %v", poolNode.Port, serviceName, utils.Stringify(pool_meta))
	return pool_meta
}

// Move this method to utils
func (o *AviObjectGraph) generateRandomStringName(name string) string {
	// TODO: Watch out for collisions, if need we can increase 10 below.
	random_string := utils.RandomSeq(5)
	// TODO: Find a way to avoid collisions
	utils.AviLog.Info.Printf("Random string generated :%s", random_string)
	name = name + "-" + random_string
	return name
}

func (o *AviObjectGraph) BuildL4LBGraph(namespace string, svcName string) {
	// We use the gateway fields to arrive at various AVI VS Node object.
	var VsNode *AviVsNode
	svcObj, err := utils.GetInformers().ServiceInformer.Lister().Services(namespace).Get(svcName)
	if err != nil {
		utils.AviLog.Warning.Printf("Error in obtaining the object for service: %s", svcName)
		return
	}
	VsNode = o.ConstructAviL4VsNode(svcObj)
	o.ConstructAviTCPPGPoolNodes(svcObj, VsNode)
	o.AddModelNode(VsNode)
	VsNode.CalculateCheckSum()
	o.GraphChecksum = o.GraphChecksum + VsNode.GetCheckSum()
	utils.AviLog.Info.Printf("Checksum  for AVI VS object %v", VsNode.GetCheckSum())
	utils.AviLog.Info.Printf("Computed Graph Checksum for VS is: %v", o.GraphChecksum)
}
