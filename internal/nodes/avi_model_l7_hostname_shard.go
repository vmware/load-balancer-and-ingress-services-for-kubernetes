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
	"regexp"
	"strings"

	avicache "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	avimodels "github.com/avinetworks/sdk/go/models"
)

func (o *AviObjectGraph) BuildL7VSGraphHostNameShard(vsName, hostname string, routeIgrObj RouteIngressModel, pathsvc []IngressHostPathSvc, key string) {
	o.Lock.Lock()
	defer o.Lock.Unlock()
	// We create pools and attach servers to them here. Pools are created with a priorty label of host/path
	namespace := routeIgrObj.GetNamespace()
	ingName := routeIgrObj.GetName()
	utils.AviLog.Infof("key: %s, msg: Building the L7 pools for namespace: %s, hostname: %s", key, namespace, hostname)
	pgName := lib.GetL7SharedPGName(vsName)
	pgNode := o.GetPoolGroupByName(pgName)
	vsNode := o.GetAviVS()
	if len(vsNode) != 1 {
		utils.AviLog.Warnf("key: %s, msg: more than one vs in model.", key)
		return
	}
	var priorityLabel string
	var poolName string
	utils.AviLog.Infof("key: %s, msg: The pathsvc mapping: %v", key, pathsvc)
	for _, obj := range pathsvc {
		if obj.Path != "" {
			priorityLabel = hostname + obj.Path
		} else {
			priorityLabel = hostname
		}

		// Using servciename in poolname for routes, but not in ingress for consistency with existing naming convention.
		// If possible, we would make this uniform
		if routeIgrObj.GetType() == utils.Ingress {
			poolName = lib.GetL7PoolName(priorityLabel, namespace, ingName)
		} else {
			poolName = lib.GetL7PoolName(priorityLabel, namespace, ingName, obj.ServiceName)
		}

		// First check if there are pools related to this ingress present in the model already
		poolNodes := o.GetAviPoolNodesByIngress(namespace, ingName)
		utils.AviLog.Debugf("key: %s, msg: found pools in the model: %s", key, utils.Stringify(poolNodes))
		for _, pool := range poolNodes {
			if pool.Name == poolName {
				o.RemovePoolNodeRefs(pool.Name)
			}
		}
		// First retrieve the FQDNs from the cache and update the model

		var storedHosts []string
		storedHosts = append(storedHosts, hostname)
		RemoveFQDNsFromModel(vsNode[0], storedHosts, key)
		if pgNode != nil {
			//utils.AviLog.Infof("key: %s, msg: hostpathsvc list: %s", key, utils.Stringify(parsedIng))
			// Processsing insecure ingress
			if !utils.HasElem(vsNode[0].VSVIPRefs[0].FQDNs, hostname) {
				vsNode[0].VSVIPRefs[0].FQDNs = append(vsNode[0].VSVIPRefs[0].FQDNs, hostname)
			}

			poolNode := &AviPoolNode{
				Name:          poolName,
				IngressName:   ingName,
				PortName:      obj.PortName,
				Tenant:        lib.GetTenant(),
				PriorityLabel: priorityLabel,
				Port:          obj.Port,
				TargetPort:    obj.TargetPort,
				ServiceMetadata: avicache.ServiceMetadataObj{
					IngressName: ingName,
					Namespace:   namespace,
					HostNames:   storedHosts,
					PoolRatio:   obj.weight,
				},
			}
			poolNode.VrfContext = lib.GetVrf()
			serviceType := lib.GetServiceType()
			if serviceType == lib.NodePortLocal {
				if servers := PopulateServersForNPL(poolNode, namespace, obj.ServiceName, true, key); servers != nil {
					poolNode.Servers = servers
				}
			} else if serviceType == lib.NodePort {
				if servers := PopulateServersForNodePort(poolNode, namespace, obj.ServiceName, true, key); servers != nil {
					poolNode.Servers = servers
				}
			} else {
				if servers := PopulateServers(poolNode, namespace, obj.ServiceName, true, key); servers != nil {
					poolNode.Servers = servers
				}
			}
			poolNode.CalculateCheckSum()
			o.AddModelNode(poolNode)
			vsNode[0].PoolRefs = append(vsNode[0].PoolRefs, poolNode)
			utils.AviLog.Debugf("key: %s, msg: the pools after append are: %v", key, utils.Stringify(vsNode[0].PoolRefs))
		}

	}
	for _, obj := range pathsvc {
		BuildPoolHTTPRule(hostname, obj.Path, ingName, namespace, key, vsNode[0], false)
	}

	// Reset the PG Node members and rebuild them
	pgNode.Members = nil
	for _, poolNode := range vsNode[0].PoolRefs {
		ratio := poolNode.ServiceMetadata.PoolRatio
		pool_ref := fmt.Sprintf("/api/pool?name=%s", poolNode.Name)
		pgNode.Members = append(pgNode.Members, &avimodels.PoolGroupMember{PoolRef: &pool_ref, PriorityLabel: &poolNode.PriorityLabel, Ratio: &ratio})

	}
}

func (o *AviObjectGraph) DeletePoolForHostname(vsName, hostname string, routeIgrObj RouteIngressModel, pathSvc map[string][]string, key string, removeFqdn, removeRedir, secure bool) {
	o.Lock.Lock()
	defer o.Lock.Unlock()

	namespace := routeIgrObj.GetNamespace()
	ingName := routeIgrObj.GetName()
	vsNode := o.GetAviVS()
	var poolName string
	keepSni := false
	if !secure {
		// Fetch the ingress pools that are present in the model and delete them.
		poolNodes := o.GetAviPoolNodesByIngress(namespace, ingName)
		utils.AviLog.Debugf("key: %s, msg: Pool Nodes to delete for ingress: %s", key, utils.Stringify(poolNodes))
		for _, pool := range poolNodes {
			// Only delete the pools that belong to the host path combinations.
			var priorityLabel string
			for path, services := range pathSvc {
				if path != "" {
					priorityLabel = hostname + path
				} else {
					priorityLabel = hostname
				}
				for _, svcName := range services {
					if routeIgrObj.GetType() == utils.Ingress {
						poolName = lib.GetL7PoolName(priorityLabel, namespace, ingName)
					} else {
						poolName = lib.GetL7PoolName(priorityLabel, namespace, ingName, svcName)
					}
					if poolName == pool.Name {
						o.RemovePoolNodeRefs(poolName)
					}
				}
			}
			// It might be safe to remove all the pools for this VS for this ingress in one shot.
		}
		pgName := lib.GetL7SharedPGName(vsName)
		pgNode := o.GetPoolGroupByName(pgName)
		pgNode.Members = nil
		for _, poolNode := range vsNode[0].PoolRefs {
			ratio := poolNode.ServiceMetadata.PoolRatio
			pool_ref := fmt.Sprintf("/api/pool?name=%s", poolNode.Name)
			pgNode.Members = append(pgNode.Members, &avimodels.PoolGroupMember{PoolRef: &pool_ref, PriorityLabel: &poolNode.PriorityLabel, Ratio: &ratio})
		}
	} else {
		// Remove the ingress from the hostmap
		hostMapOk, ingressHostMap := SharedHostNameLister().Get(hostname)
		if hostMapOk {
			// Replace the ingress map for this host.
			keyToRemove := namespace + "/" + ingName
			delete(ingressHostMap.HostNameMap, keyToRemove)
			SharedHostNameLister().Save(hostname, ingressHostMap)
		}

		isIngr := routeIgrObj.GetType() == utils.Ingress
		// SNI VSes donot have secretname in their names
		sniNodeName := lib.GetSniNodeName(ingName, namespace, "", hostname)
		utils.AviLog.Infof("key: %s, msg: sni node to delete: %s", key, sniNodeName)
		keepSni = o.ManipulateSniNode(sniNodeName, ingName, namespace, hostname, pathSvc, vsNode, key, isIngr)
	}
	if removeFqdn && !keepSni {
		var hosts []string
		hosts = append(hosts, hostname)
		// Remove these hosts from the overall FQDN list
		RemoveFQDNsFromModel(vsNode[0], hosts, key)
	}
	if removeRedir && !keepSni {
		RemoveRedirectHTTPPolicyInModel(vsNode[0], hostname, key)
	}

}

func (o *AviObjectGraph) ManipulateSniNode(currentSniNodeName, ingName, namespace, hostname string, pathSvc map[string][]string, vsNode []*AviVsNode, key string, isIngr bool) bool {
	for _, modelSniNode := range vsNode[0].SniNodes {
		if currentSniNodeName != modelSniNode.Name {
			continue
		}
		for path, services := range pathSvc {
			pgName := lib.GetSniPGName(ingName, namespace, hostname, path)
			pgNode := modelSniNode.GetPGForVSByName(pgName)
			for _, svc := range services {
				var sniPool string
				if isIngr {
					sniPool = lib.GetSniPoolName(ingName, namespace, hostname, path)
				} else {
					sniPool = lib.GetSniPoolName(ingName, namespace, hostname, path, svc)
				}
				o.RemovePoolNodeRefsFromSni(sniPool, modelSniNode)
				o.RemovePoolRefsFromPG(sniPool, pgNode)
			}
			// Remove the SNI PG if it has no member
			if pgNode != nil {
				if len(pgNode.Members) == 0 {
					o.RemovePGNodeRefs(pgName, modelSniNode)
					httppolname := lib.GetSniHttpPolName(ingName, namespace, hostname, path)
					o.RemoveHTTPRefsFromSni(httppolname, modelSniNode)
				}
			}
		}
		// After going through the paths, if the SNI node does not have any PGs - then delete it.
		if len(modelSniNode.PoolRefs) == 0 {
			RemoveSniInModel(currentSniNodeName, vsNode, key)
			// Remove the snihost mapping
			SharedHostNameLister().Delete(hostname)
			return false
		}
	}

	return true
}

func updateHostPathCache(ns, ingress string, oldHostMap, newHostMap map[string]map[string][]string) {
	mmapval := ns + "/" + ingress

	// remove from oldHostMap
	for _, oldMap := range oldHostMap {
		for host, paths := range oldMap {
			for _, path := range paths {
				SharedHostNameLister().RemoveHostPathStore(host, path, mmapval)
			}
		}
	}

	// add from newHostMap
	if newHostMap != nil {
		for _, newMap := range newHostMap {
			for host, paths := range newMap {
				for _, path := range paths {
					SharedHostNameLister().SaveHostPathStore(host, path, mmapval)
				}
			}
		}
	}
}

// difference returns the elements in `a` that aren't in `b`.
func Difference(a, b []string) []string {
	mb := make(map[string]struct{}, len(b))
	for _, x := range b {
		mb[x] = struct{}{}
	}
	var diff []string
	for _, x := range a {
		if _, found := mb[x]; !found {
			diff = append(diff, x)
		}
	}
	return diff
}

func getPaths(pathMapArr []IngressHostPathSvc) []string {
	// Returns a list of paths for a given host
	paths := []string{}
	for _, pathmap := range pathMapArr {
		paths = append(paths, pathmap.Path)
	}
	return paths
}

func sniNodeHostName(routeIgrObj RouteIngressModel, tlssetting TlsSettings, ingName, namespace, key string, fullsync bool, sharedQueue *utils.WorkerQueue, modelList *[]string) map[string][]IngressHostPathSvc {
	hostPathSvcMap := make(map[string][]IngressHostPathSvc)
	for sniHost, paths := range tlssetting.Hosts {
		var sniHosts []string
		hostPathSvcMap[sniHost] = paths
		hostMap := HostNamePathSecrets{paths: getPaths(paths), secretName: tlssetting.SecretName}
		found, ingressHostMap := SharedHostNameLister().Get(sniHost)
		if found {
			// Replace the ingress map for this host.
			ingressHostMap.HostNameMap[namespace+"/"+ingName] = hostMap
			ingressHostMap.GetIngressesForHostName(sniHost)
		} else {
			// Create the map
			ingressHostMap = NewSecureHostNameMapProp()
			ingressHostMap.HostNameMap[namespace+"/"+ingName] = hostMap
		}
		SharedHostNameLister().Save(sniHost, ingressHostMap)
		sniHosts = append(sniHosts, sniHost)
		shardVsName := DeriveHostNameShardVS(sniHost, key)
		// For each host, create a SNI node with the secret giving us the key and cert.
		// construct a SNI VS node per tls setting which corresponds to one secret
		if shardVsName == "" {
			// If we aren't able to derive the ShardVS name, we should return
			//return hostPathMap
			return hostPathSvcMap
		}
		model_name := lib.GetModelName(lib.GetTenant(), shardVsName)
		found, aviModel := objects.SharedAviGraphLister().Get(model_name)
		if !found || aviModel == nil {
			utils.AviLog.Infof("key: %s, msg: model not found, generating new model with name: %s", key, model_name)
			aviModel = NewAviObjectGraph()
			aviModel.(*AviObjectGraph).ConstructAviL7VsNode(shardVsName, key)
		}
		vsNode := aviModel.(*AviObjectGraph).GetAviVS()

		if len(vsNode) < 1 {
			return nil
		}

		certsBuilt := false
		sniSecretName := tlssetting.SecretName
		re := regexp.MustCompile(fmt.Sprintf(`^%s.*`, lib.DummySecret))
		if re.MatchString(sniSecretName) {
			sniSecretName = strings.Split(sniSecretName, "/")[1]
			certsBuilt = true
		}

		sniNode := vsNode[0].GetSniNodeForName(lib.GetSniNodeName(ingName, namespace, sniSecretName, sniHost))
		if sniNode == nil {
			sniNode = &AviVsNode{
				Name:         lib.GetSniNodeName(ingName, namespace, sniSecretName, sniHost),
				VHParentName: vsNode[0].Name,
				Tenant:       lib.GetTenant(),
				IsSNIChild:   true,
				ServiceMetadata: avicache.ServiceMetadataObj{
					NamespaceIngressName: ingressHostMap.GetIngressesForHostName(sniHost),
					Namespace:            namespace,
					HostNames:            sniHosts,
				},
			}
		} else {
			// The SNI node exists, just update the svc metadata
			sniNode.ServiceMetadata.NamespaceIngressName = ingressHostMap.GetIngressesForHostName(sniHost)
			sniNode.ServiceMetadata.Namespace = namespace
			sniNode.ServiceMetadata.HostNames = sniHosts
			if sniNode.SSLKeyCertAviRef != "" {
				certsBuilt = true
			}
		}
		if lib.GetSEGName() != lib.DEFAULT_SE_GROUP {
			sniNode.ServiceEngineGroup = lib.GetSEGName()
		}
		sniNode.VrfContext = lib.GetVrf()
		if !certsBuilt {
			certsBuilt = aviModel.(*AviObjectGraph).BuildTlsCertNode(routeIgrObj.GetSvcLister(), sniNode, namespace, tlssetting, key, sniHost)
		}
		if certsBuilt {
			isIngr := routeIgrObj.GetType() == utils.Ingress
			aviModel.(*AviObjectGraph).BuildPolicyPGPoolsForSNI(vsNode, sniNode, namespace, ingName, tlssetting, sniSecretName, key, isIngr, sniHost)
			foundSniModel := FindAndReplaceSniInModel(sniNode, vsNode, key)
			if !foundSniModel {
				vsNode[0].SniNodes = append(vsNode[0].SniNodes, sniNode)
			}
			RemoveRedirectHTTPPolicyInModel(vsNode[0], sniHost, key)
			if tlssetting.redirect == true {
				aviModel.(*AviObjectGraph).BuildPolicyRedirectForVS(vsNode, sniHost, namespace, ingName, key)
			}
			BuildL7HostRule(sniHost, namespace, ingName, key, sniNode)
		} else {
			hostMapOk, ingressHostMap := SharedHostNameLister().Get(sniHost)
			if hostMapOk {
				// Replace the ingress map for this host.
				keyToRemove := namespace + "/" + ingName
				delete(ingressHostMap.HostNameMap, keyToRemove)
				SharedHostNameLister().Save(sniHost, ingressHostMap)
			}
			// Since the cert couldn't be built, check if this SNI is affected by only in ingress if so remove the sni node from the model
			if len(ingressHostMap.GetIngressesForHostName(sniHost)) == 0 {
				RemoveSniInModel(sniNode.Name, vsNode, key)
				RemoveRedirectHTTPPolicyInModel(vsNode[0], sniHost, key)
			}
		}
		// Only add this node to the list of models if the checksum has changed.
		modelChanged := saveAviModel(model_name, aviModel.(*AviObjectGraph), key)
		if !utils.HasElem(*modelList, model_name) && modelChanged {
			*modelList = append(*modelList, model_name)
		}
	}

	//return hostPathMap
	return hostPathSvcMap
}
