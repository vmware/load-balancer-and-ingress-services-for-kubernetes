/*
* [2013] - [2020] Avi Networks Incorporated
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

	avicache "ako/pkg/cache"
	"ako/pkg/lib"
	"ako/pkg/objects"

	"github.com/avinetworks/container-lib/utils"
	avimodels "github.com/avinetworks/sdk/go/models"
	"k8s.io/apimachinery/pkg/api/errors"
)

func (o *AviObjectGraph) BuildL7VSGraphHostNameShard(vsName string, namespace string, ingName string, hostname string, pathsvc []IngressHostPathSvc, key string) {
	o.Lock.Lock()
	defer o.Lock.Unlock()
	// We create pools and attach servers to them here. Pools are created with a priorty label of host/path
	utils.AviLog.Infof("key: %s, msg: Building the L7 pools for namespace: %s, hostname: %s", key, namespace, hostname)
	pgName := lib.GetL7SharedPGName(vsName)
	pgNode := o.GetPoolGroupByName(pgName)
	vsNode := o.GetAviVS()
	if len(vsNode) != 1 {
		utils.AviLog.Warnf("key: %s, msg: more than one vs in model.", key)
		return
	}
	var priorityLabel string
	utils.AviLog.Infof("key: %s, msg: The pathsvc mapping: %v", key, pathsvc)
	for _, obj := range pathsvc {
		if obj.Path != "" {
			priorityLabel = hostname + obj.Path
		} else {
			priorityLabel = hostname
		}
		poolName := lib.GetL7PoolName(priorityLabel, namespace, ingName)
		// First check if there are pools related to this ingress present in the model already
		poolNodes := o.GetAviPoolNodesByIngress(namespace, ingName)
		utils.AviLog.Infof("key: %s, msg: found pools in the model: %s", key, utils.Stringify(poolNodes))
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

			poolNode := &AviPoolNode{Name: poolName, IngressName: ingName, Tenant: lib.GetTenant(), PriorityLabel: priorityLabel, Port: obj.Port, ServiceMetadata: avicache.ServiceMetadataObj{IngressName: ingName, Namespace: namespace, HostNames: storedHosts}}
			poolNode.VrfContext = lib.GetVrf()
			if servers := PopulateServers(poolNode, namespace, obj.ServiceName, key); servers != nil {
				poolNode.Servers = servers
			}
			poolNode.CalculateCheckSum()
			o.AddModelNode(poolNode)
			vsNode[0].PoolRefs = append(vsNode[0].PoolRefs, poolNode)
			utils.AviLog.Infof("key: %s, msg: the pools after append are: %v", key, utils.Stringify(vsNode[0].PoolRefs))
		}

	}
	// Reset the PG Node members and rebuild them
	pgNode.Members = nil
	for _, poolNode := range vsNode[0].PoolRefs {
		pool_ref := fmt.Sprintf("/api/pool?name=%s", poolNode.Name)
		pgNode.Members = append(pgNode.Members, &avimodels.PoolGroupMember{PoolRef: &pool_ref, PriorityLabel: &poolNode.PriorityLabel})

	}
}

func (o *AviObjectGraph) DeletePoolForHostname(vsName, namespace, ingName, hostname string, paths []string, key string, removeFqdn, removeRedir, secure bool) {
	o.Lock.Lock()
	defer o.Lock.Unlock()

	vsNode := o.GetAviVS()
	keepSni := false
	if !secure {
		// Fetch the ingress pools that are present in the model and delete them.
		poolNodes := o.GetAviPoolNodesByIngress(namespace, ingName)
		utils.AviLog.Infof("key: %s, msg: Pool Nodes to delete for ingress:  %s", key, utils.Stringify(poolNodes))
		for _, pool := range poolNodes {
			// Only delete the pools that belong to the host path combinations.
			var priorityLabel string
			for _, path := range paths {
				if path != "" {
					priorityLabel = hostname + path
				} else {
					priorityLabel = hostname
				}
				poolName := lib.GetL7PoolName(priorityLabel, namespace, ingName)
				if poolName == pool.Name {
					o.RemovePoolNodeRefs(poolName)
				}
			}
			// It might be safe to remove all the pools for this VS for this ingress in one shot.
		}
		pgName := lib.GetL7SharedPGName(vsName)
		pgNode := o.GetPoolGroupByName(pgName)
		pgNode.Members = nil
		for _, poolNode := range vsNode[0].PoolRefs {
			pool_ref := fmt.Sprintf("/api/pool?name=%s", poolNode.Name)
			pgNode.Members = append(pgNode.Members, &avimodels.PoolGroupMember{PoolRef: &pool_ref, PriorityLabel: &poolNode.PriorityLabel})
		}
	} else {
		// Generate SNI nodes and mark them for deletion. SNI node names: ingressname--namespace--secretname
		// Fetch all the secrets for this ingress
		found, secrets := objects.SharedSvcLister().IngressMappings(namespace).GetIngToSecret(ingName)
		// Remove the ingress from the hostmap
		hostMapOk, ingressHostMap := SharedHostNameLister().Get(hostname)
		if hostMapOk {
			// Replace the ingress map for this host.
			keyToRemove := namespace + "/" + ingName
			delete(ingressHostMap.HostNameMap, keyToRemove)
			SharedHostNameLister().Save(hostname, ingressHostMap)
		}
		utils.AviLog.Infof("key: %s, msg: retrieved secrets for ingress: %s", key, secrets)
		if found {
			for _, secret := range secrets {
				sniNodeName := lib.GetSniNodeName(ingName, namespace, secret, hostname)
				// This is a bit tricky. We will get the hostname and the paths to be deleted from the sni node.
				// So first we get the SNI node from memory and we update the pool/pg/httppol of the sninode.
				// However if this update leads to 0 pools in the sni node, it means the SNI node can be fully deleted.
				utils.AviLog.Infof("key: %s, msg: sni node to delete :%s", key, sniNodeName)
				keepSni = o.ManipulateSniNode(sniNodeName, ingName, namespace, hostname, paths, vsNode, key)
			}
		}
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

func (o *AviObjectGraph) ManipulateSniNode(currentSniNodeName, ingName, namespace, hostname string, paths []string, vsNode []*AviVsNode, key string) bool {
	for _, modelSniNode := range vsNode[0].SniNodes {
		if currentSniNodeName == modelSniNode.Name {
			for _, path := range paths {
				sniPool := lib.GetSniPoolName(ingName, namespace, hostname, path)
				o.RemovePoolNodeRefsFromSni(sniPool, modelSniNode)
				pgName := lib.GetSniPGName(ingName, namespace, hostname, path)
				o.RemovePgNodeRefsFromSni(pgName, modelSniNode)
				httppolname := lib.GetSniHttpPolName(ingName, namespace, hostname, path)
				o.RemoveHTTPRefsFromSni(httppolname, modelSniNode)
			}
			// After going through the paths, if the SNI node does not have any PGs - then delete it.
			if len(modelSniNode.PoolRefs) == 0 {
				RemoveSniInModel(currentSniNodeName, vsNode, key)
				// Remove the snihost mapping
				SharedHostNameLister().Delete(hostname)
				return false
			}
		}
	}
	return true

}

func HostNameShardAndPublish(ingress, namespace, key string, fullsync bool, sharedQueue *utils.WorkerQueue) {
	var err error
	o := NewNodesValidator()
	myIng, err := utils.GetInformers().IngressInformer.Lister().ByNamespace(namespace).Get(ingress)

	if err != nil {
		utils.AviLog.Infof("key: %s, msg: Error :%v", key, err)
		// Detect a delete condition here.
		if errors.IsNotFound(err) {
			DeletePoolsByHostname(namespace, ingress, key, fullsync, sharedQueue)
		}
	} else {
		ingObj, ok := utils.ToNetworkingIngress(myIng)
		if !ok {
			utils.AviLog.Errorf("Unable to convert obj type interface to networking/v1beta1 ingress")
		}

		var parsedIng IngressConfig
		var modelList []string
		processIng := true

		processIng = filterIngressOnClass(ingObj)
		if !processIng {
			// If the ingress class is not right, let's delete it.
			DeletePoolsByHostname(namespace, ingress, key, fullsync, sharedQueue)
		}
		parsedIng = o.ParseHostPathForIngress(namespace, ingress, ingObj.Spec, key)

		if processIng {
			// Check if this ingress and had any previous mappings, if so - delete them first.
			storedHostsFound, Storedhosts := objects.SharedSvcLister().IngressMappings(namespace).GetIngToHost(ingress)
			// Process insecure routes first.
			hostsMap := make(map[string]map[string][]string)
			insecureHostPathMapArr := make(map[string][]string)
			for host, pathsvcmap := range parsedIng.IngressHostMap {
				if storedHostsFound {
					// Remove this entry from storedHosts. First check if the host exists in the stored map or not.
					storedPaths, found := Storedhosts["insecure"][host]
					if found {
						// TODO: StoredPaths might be empty if the host was not specified with any paths.
						// Verify the paths and take out the paths that are not need.
						diffStoredPaths := Difference(storedPaths, getPaths(pathsvcmap))
						if len(diffStoredPaths) == 0 {
							// There's no difference between the paths, we should delete the host entry in the stored Map
							delete(Storedhosts["insecure"], host)
						} else {
							// These paths are meant for deletion
							Storedhosts["insecure"][host] = diffStoredPaths
						}
					}
				}
				insecureHostPathMapArr[host] = getPaths(pathsvcmap)
				shardVsName := DeriveHostNameShardVS(host, key)
				if shardVsName == "" {
					// If we aren't able to derive the ShardVS name, we should return
					return
				}
				model_name := lib.GetModelName(lib.GetTenant(), shardVsName)
				found, aviModel := objects.SharedAviGraphLister().Get(model_name)
				if !found || aviModel == nil {
					utils.AviLog.Infof("key: %s, msg: model not found, generating new model with name: %s", key, model_name)
					aviModel = NewAviObjectGraph()
					aviModel.(*AviObjectGraph).ConstructAviL7VsNode(shardVsName, key)
				}
				aviModel.(*AviObjectGraph).BuildL7VSGraphHostNameShard(shardVsName, namespace, ingress, host, pathsvcmap, key)
				changedModel := saveAviModel(model_name, aviModel.(*AviObjectGraph), key)
				if !utils.HasElem(modelList, model_name) && changedModel {
					modelList = append(modelList, model_name)
				}
			}
			hostsMap["insecure"] = insecureHostPathMapArr
			// Process secure routes next.
			secureHostPathMapArr := make(map[string][]string)
			for _, tlssetting := range parsedIng.TlsCollection {
				locSniHostMap := sniNodeHostName(tlssetting, ingress, namespace, key, fullsync, sharedQueue, &modelList)
				for host, newPaths := range locSniHostMap {
					if storedHostsFound {
						// Remove this entry from storedHosts. First check if the host exists in the stored map or not.
						storedPaths, found := Storedhosts["secure"][host]
						if found {
							// TODO: StoredPaths might be empty if the host was not specified with any paths.
							// Verify the paths and take out the paths that are not need.
							diffStoredPaths := Difference(storedPaths, newPaths)
							if len(diffStoredPaths) == 0 {
								// There's no difference between the paths, we should delete the host entry in the stored Map
								delete(Storedhosts["secure"], host)
							} else {
								// These paths are meant for deletion
								Storedhosts["secure"][host] = diffStoredPaths
							}
						}
					}
					secureHostPathMapArr[host] = newPaths
				}
			}
			utils.AviLog.Infof("key: %s, msg: Stored hosts: %s", key, Storedhosts)
			hostsMap["secure"] = secureHostPathMapArr

			if storedHostsFound {
				for hostType, hosts := range Storedhosts {
					for host, paths := range hosts {
						shardVsName := DeriveHostNameShardVS(host, key)
						if shardVsName == "" {
							// If we aren't able to derive the ShardVS name, we should return
							return
						}
						model_name := lib.GetModelName(lib.GetTenant(), shardVsName)
						found, aviModel := objects.SharedAviGraphLister().Get(model_name)
						if !found || aviModel == nil {
							utils.AviLog.Warnf("key: %s, msg: model not found during delete: %s", key, model_name)
							continue
						}
						// By default remove both redirect and fqdn. So if the host isn't transitioning, then we will remove both.
						removeFqdn := true
						removeRedir := true
						for currentHost, _ := range hostsMap["secure"] {
							if hostType == "insecure" { // old host was insecure
								if currentHost == host {
									// Transition from insecure --> secure
									removeRedir = false
									removeFqdn = false
								}
							}
							if hostType == "secure" {
								if currentHost == host { // old host was secure
									// Transition from secure --> secure (path changed)
									removeRedir = false
									removeFqdn = false
								}
							}

						}
						for currentHost, _ := range hostsMap["insecure"] {
							if hostType == "secure" { // old host was secure
								if currentHost == host {
									// Transition from secure --> insecure
									removeRedir = true
									removeFqdn = false
								}
							}
							if hostType == "insecure" { // old host was insecure
								if currentHost == host {
									// Transition from insecure --> insecure (path changed)
									removeFqdn = false
								}
							}
						}
						// Delete the pool corresponding to this host
						if hostType == "secure" {
							aviModel.(*AviObjectGraph).DeletePoolForHostname(shardVsName, namespace, ingress, host, paths, key, removeFqdn, removeRedir, true)
						} else {
							aviModel.(*AviObjectGraph).DeletePoolForHostname(shardVsName, namespace, ingress, host, paths, key, removeFqdn, removeRedir, false)

						}
						changedModel := saveAviModel(model_name, aviModel.(*AviObjectGraph), key)
						if !utils.HasElem(modelList, model_name) && changedModel {
							modelList = append(modelList, model_name)
						}
					}
				}
			}

			objects.SharedSvcLister().IngressMappings(namespace).UpdateIngToHostMapping(ingress, hostsMap)
			if !fullsync {
				utils.AviLog.Infof("key: %s, msg: List of models to publish: %s", key, modelList)
				for _, modelName := range modelList {
					PublishKeyToRestLayer(modelName, key, sharedQueue)
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
	var paths []string
	for _, pathmap := range pathMapArr {
		paths = append(paths, pathmap.Path)
	}
	return paths
}

func DeletePoolsByHostname(namespace, ingress, key string, fullsync bool, sharedQueue *utils.WorkerQueue) {
	ok, hostMap := objects.SharedSvcLister().IngressMappings(namespace).GetIngToHost(ingress)
	if !ok {
		utils.AviLog.Warnf("key: %s, msg: nothing to delete for ingress: %s", key, ingress)
		return
	}

	utils.AviLog.Infof("key: %s, msg: hosts to delete are: :%s", key, hostMap)
	for hostType, hosts := range hostMap {
		for host, paths := range hosts {
			shardVsName := DeriveHostNameShardVS(host, key)

			if shardVsName == "" {
				// If we aren't able to derive the ShardVS name, we should return
				return
			}
			model_name := lib.GetModelName(lib.GetTenant(), shardVsName)
			found, aviModel := objects.SharedAviGraphLister().Get(model_name)
			if !found || aviModel == nil {
				utils.AviLog.Warnf("key: %s, msg: model not found during delete: %s", key, model_name)
				continue
			}
			// Delete the pool corresponding to this host
			if hostType == "secure" {
				aviModel.(*AviObjectGraph).DeletePoolForHostname(shardVsName, namespace, ingress, host, paths, key, true, true, true)
			} else {
				aviModel.(*AviObjectGraph).DeletePoolForHostname(shardVsName, namespace, ingress, host, paths, key, true, true, false)
			}
			ok := saveAviModel(model_name, aviModel.(*AviObjectGraph), key)
			if ok && len(aviModel.(*AviObjectGraph).GetOrderedNodes()) != 0 && !fullsync {
				PublishKeyToRestLayer(model_name, key, sharedQueue)
			}
		}
	}
	// Now remove the secret relationship
	objects.SharedSvcLister().IngressMappings(namespace).RemoveIngressSecretMappings(ingress)
	// Remove the hosts mapping for this ingress
	objects.SharedSvcLister().IngressMappings(namespace).DeleteIngToHostMapping(ingress)
}

func sniNodeHostName(tlssetting TlsSettings, ingName, namespace, key string, fullsync bool, sharedQueue *utils.WorkerQueue, modelList *[]string) map[string][]string {
	hostPathMap := make(map[string][]string)
	for sniHost, paths := range tlssetting.Hosts {
		var allSniHosts []string
		var sniHosts []string
		hostPathMap[sniHost] = getPaths(paths)
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
		allSniHosts = append(allSniHosts, sniHost)
		shardVsName := DeriveHostNameShardVS(sniHost, key)
		// For each host, create a SNI node with the secret giving us the key and cert.
		// construct a SNI VS node per tls setting which corresponds to one secret
		if shardVsName == "" {
			// If we aren't able to derive the ShardVS name, we should return
			return hostPathMap
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
		sniNode := vsNode[0].GetSniNodeForName(lib.GetSniNodeName(ingName, namespace, tlssetting.SecretName, sniHost))
		if sniNode == nil {
			sniNode = &AviVsNode{
				Name:         lib.GetSniNodeName(ingName, namespace, tlssetting.SecretName, sniHost),
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
			sniNode.ServiceMetadata = avicache.ServiceMetadataObj{
				NamespaceIngressName: ingressHostMap.GetIngressesForHostName(sniHost),
				Namespace:            namespace,
				HostNames:            sniHosts,
			}
		}
		sniNode.VrfContext = lib.GetVrf()
		certsBuilt := aviModel.(*AviObjectGraph).BuildTlsCertNode(sniNode, namespace, tlssetting.SecretName, key, sniHost)
		if certsBuilt {
			aviModel.(*AviObjectGraph).BuildPolicyPGPoolsForSNI(vsNode, sniNode, namespace, ingName, tlssetting, tlssetting.SecretName, key, sniHost)
			foundSniModel := FindAndReplaceSniInModel(sniNode, vsNode, key)
			if !foundSniModel {
				vsNode[0].SniNodes = append(vsNode[0].SniNodes, sniNode)
			}
			aviModel.(*AviObjectGraph).BuildPolicyRedirectForVS(vsNode, sniHost, namespace, ingName, key)

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

	return hostPathMap
}
