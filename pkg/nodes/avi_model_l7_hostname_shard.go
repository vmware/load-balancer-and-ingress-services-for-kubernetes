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
	extensionv1beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
)

func (o *AviObjectGraph) BuildL7VSGraphHostNameShard(vsName string, namespace string, ingName string, hostname string, pathsvc []IngressHostPathSvc, key string) {
	o.Lock.Lock()
	defer o.Lock.Unlock()
	// We create pools and attach servers to them here. Pools are created with a priorty label of host/path
	utils.AviLog.Info.Printf("key: %s, msg: Building the L7 pools for namespace: %s, hostname: %s", key, namespace, hostname)
	pgName := lib.GetL7SharedPGName(vsName)
	pgNode := o.GetPoolGroupByName(pgName)
	vsNode := o.GetAviVS()
	if len(vsNode) != 1 {
		utils.AviLog.Warning.Printf("key: %s, msg: more than one vs in model.", key)
		return
	}
	var priorityLabel string
	utils.AviLog.Info.Printf("key: %s, msg: The pathsvc mapping: %v", key, pathsvc)
	for _, obj := range pathsvc {
		if obj.Path != "" {
			priorityLabel = hostname + obj.Path
		} else {
			priorityLabel = hostname
		}
		poolName := lib.GetL7PoolName(priorityLabel, namespace, ingName)
		// First check if there are pools related to this ingress present in the model already
		poolNodes := o.GetAviPoolNodesByIngress(namespace, ingName)
		utils.AviLog.Info.Printf("key: %s, msg: found pools in the model: %s", key, utils.Stringify(poolNodes))
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
			//utils.AviLog.Info.Printf("key: %s, msg: hostpathsvc list: %s", key, utils.Stringify(parsedIng))
			// Processsing insecure ingress
			if !utils.HasElem(vsNode[0].VSVIPRefs[0].FQDNs, hostname) {
				vsNode[0].VSVIPRefs[0].FQDNs = append(vsNode[0].VSVIPRefs[0].FQDNs, hostname)
			}

			poolNode := &AviPoolNode{Name: poolName, IngressName: ingName, Tenant: utils.ADMIN_NS, PriorityLabel: priorityLabel, Port: obj.Port, ServiceMetadata: avicache.ServiceMetadataObj{IngressName: ingName, Namespace: namespace, HostNames: storedHosts}}
			poolNode.VrfContext = lib.GetVrf()
			if servers := PopulateServers(poolNode, namespace, obj.ServiceName, key); servers != nil {
				poolNode.Servers = servers
			}
			poolNode.CalculateCheckSum()
			o.AddModelNode(poolNode)
			vsNode[0].PoolRefs = append(vsNode[0].PoolRefs, poolNode)
			utils.AviLog.Info.Printf("key: %s, msg: the pools after append are: %v", key, utils.Stringify(vsNode[0].PoolRefs))
		}

	}
	// Reset the PG Node members and rebuild them
	pgNode.Members = nil
	for _, poolNode := range vsNode[0].PoolRefs {
		pool_ref := fmt.Sprintf("/api/pool?name=%s", poolNode.Name)
		pgNode.Members = append(pgNode.Members, &avimodels.PoolGroupMember{PoolRef: &pool_ref, PriorityLabel: &poolNode.PriorityLabel})

	}
}

func (o *AviObjectGraph) DeletePoolForHostname(vsName, namespace, ingName, hostname, key string, secure bool) {
	o.Lock.Lock()
	defer o.Lock.Unlock()

	vsNode := o.GetAviVS()
	if !secure {
		// Fetch the ingress pools that are present in the model and delete them.
		poolNodes := o.GetAviPoolNodesByIngress(namespace, ingName)
		utils.AviLog.Info.Printf("key: %s, msg: Pool Nodes to delete for ingress:  %s", key, utils.Stringify(poolNodes))
		for _, pool := range poolNodes {
			// It might be safe to remove all the pools for this VS for this ingress in one shot.
			o.RemovePoolNodeRefs(pool.Name)
		}
	}
	// Generate SNI nodes and mark them for deletion. SNI node names: ingressname--namespace--secretname
	// Fetch all the secrets for this ingress
	if secure {
		found, secrets := objects.SharedSvcLister().IngressMappings(namespace).GetIngToSecret(ingName)
		utils.AviLog.Info.Printf("key: %s, msg: retrieved secrets for ingress: %s", key, secrets)
		if found {
			for _, secret := range secrets {
				sniNodeName := lib.GetSniNodeName(ingName, namespace, secret, hostname)
				utils.AviLog.Info.Printf("key: %s, msg: sni node to delete :%s", key, sniNodeName)
				RemoveSniInModel(sniNodeName, vsNode, key)
				RemoveRedirectHTTPPolicyInModel(vsNode[0], hostname, key)
			}
		}
	}
	var hosts []string
	hosts = append(hosts, hostname)

	// Remove these hosts from the overall FQDN list
	RemoveFQDNsFromModel(vsNode[0], hosts, key)
	// Reset the PG Node members and rebuild them
	if !secure {
		pgName := lib.GetL7SharedPGName(vsName)
		pgNode := o.GetPoolGroupByName(pgName)
		pgNode.Members = nil
		for _, poolNode := range vsNode[0].PoolRefs {
			pool_ref := fmt.Sprintf("/api/pool?name=%s", poolNode.Name)
			pgNode.Members = append(pgNode.Members, &avimodels.PoolGroupMember{PoolRef: &pool_ref, PriorityLabel: &poolNode.PriorityLabel})
		}
	}
}

func HostNameShardAndPublish(ingress, namespace, key string, fullsync bool, sharedQueue *utils.WorkerQueue) {
	var ingObj interface{}
	var err error
	o := NewNodesValidator()
	if lib.GetIngressApi() == utils.ExtV1IngressInformer {
		ingObj, err = utils.GetInformers().ExtV1IngressInformer.Lister().Ingresses(namespace).Get(ingress)
	} else {
		ingObj, err = utils.GetInformers().CoreV1IngressInformer.Lister().Ingresses(namespace).Get(ingress)
	}
	if err != nil {
		utils.AviLog.Info.Printf("key :%s, msg: Error :%v", key, err)
		// Detect a delete condition here.
		if errors.IsNotFound(err) {
			DeletePoolsByHostname(namespace, ingress, key, fullsync, sharedQueue)
		}
	} else {
		var parsedIng IngressConfig
		var modelList []string
		processIng := true
		if lib.GetIngressApi() == utils.ExtV1IngressInformer {
			processIng = filterIngressOnClassExtV1(ingObj.(*extensionv1beta1.Ingress))
			if !processIng {
				// If the ingress class is not right, let's delete it.
				DeletePoolsByHostname(namespace, ingress, key, fullsync, sharedQueue)
			}
			parsedIng = o.ParseHostPathForIngress(namespace, ingress, ingObj.(*extensionv1beta1.Ingress).Spec, key)
		} else {
			processIng = filterIngressOnClass(ingObj.(*v1beta1.Ingress))
			if !processIng {
				// If the ingress class is not right, let's delete it.
				DeletePoolsByHostname(namespace, ingress, key, fullsync, sharedQueue)
			}
			parsedIng = o.ParseHostPathForIngressCoreV1(namespace, ingress, ingObj.(*v1beta1.Ingress).Spec, key)
		}
		if processIng {
			// Check if this ingress and had any previous mappings, if so - delete them first.
			storedHostsFound, Storedhosts := objects.SharedSvcLister().IngressMappings(namespace).GetIngToHost(ingress)
			// Process insecure routes first.
			hostsMap := make(map[string][]string)
			var insecureHosts []string
			for host, pathsvcmap := range parsedIng.IngressHostMap {
				if storedHostsFound {
					// Remove this entry from storedHosts
					Storedhosts["insecure"] = utils.Remove(Storedhosts["insecure"], host)
				}
				insecureHosts = append(insecureHosts, host)
				shardVsName := DeriveHostNameShardVS(host, key)
				if shardVsName == "" {
					// If we aren't able to derive the ShardVS name, we should return
					return
				}
				model_name := lib.GetModelName(utils.ADMIN_NS, shardVsName)
				found, aviModel := objects.SharedAviGraphLister().Get(model_name)
				if !found || aviModel == nil {
					utils.AviLog.Info.Printf("key :%s, msg: model not found, generating new model with name: %s", key, model_name)
					aviModel = NewAviObjectGraph()
					aviModel.(*AviObjectGraph).ConstructAviL7VsNode(shardVsName, key)
				}
				aviModel.(*AviObjectGraph).BuildL7VSGraphHostNameShard(shardVsName, namespace, ingress, host, pathsvcmap, key)
				changedModel := saveAviModel(model_name, aviModel.(*AviObjectGraph), key)
				if !utils.HasElem(modelList, model_name) && changedModel {
					modelList = append(modelList, model_name)
				}
			}
			hostsMap["insecure"] = insecureHosts
			// Process secure routes next.
			var sniHosts []string
			for _, tlssetting := range parsedIng.TlsCollection {
				locSniHost := sniNodeHostName(tlssetting, ingress, namespace, key, fullsync, sharedQueue, &modelList)
				sniHosts = append(sniHosts, locSniHost...)
				if storedHostsFound {
					for _, hostToRemove := range locSniHost {
						// Remove this entry from storedHosts
						Storedhosts["secure"] = utils.Remove(Storedhosts["secure"], hostToRemove)
					}
				}
			}
			utils.AviLog.Info.Printf("key :%s, msg: Stored hosts: %s", key, Storedhosts)
			//hosts = append(hosts, sniHosts...)
			hostsMap["secure"] = sniHosts

			if storedHostsFound {
				for hostType, hosts := range Storedhosts {
					for _, host := range hosts {
						shardVsName := DeriveHostNameShardVS(host, key)
						if shardVsName == "" {
							// If we aren't able to derive the ShardVS name, we should return
							return
						}
						model_name := lib.GetModelName(utils.ADMIN_NS, shardVsName)
						found, aviModel := objects.SharedAviGraphLister().Get(model_name)
						if !found || aviModel == nil {
							utils.AviLog.Warning.Printf("key :%s, msg: model not found during delete: %s", key, model_name)
							continue
						}
						// Delete the pool corresponding to this host
						if hostType == "secure" {
							aviModel.(*AviObjectGraph).DeletePoolForHostname(shardVsName, namespace, ingress, host, key, true)
						} else {
							aviModel.(*AviObjectGraph).DeletePoolForHostname(shardVsName, namespace, ingress, host, key, false)

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
				utils.AviLog.Info.Printf("key :%s, msg: List of models to publish: %s", key, modelList)
				for _, modelName := range modelList {
					PublishKeyToRestLayer(modelName, key, sharedQueue)
				}
			}
		}
	}
}

func DeletePoolsByHostname(namespace, ingress, key string, fullsync bool, sharedQueue *utils.WorkerQueue) {
	ok, hostMap := objects.SharedSvcLister().IngressMappings(namespace).GetIngToHost(ingress)
	if !ok {
		utils.AviLog.Warning.Printf("key :%s, msg: nothing to delete for ingress: %s", key, ingress)
		return
	}

	utils.AviLog.Info.Printf("key :%s, msg: hosts to delete are: :%s", key, hostMap)
	for hostType, hosts := range hostMap {
		for _, host := range hosts {
			shardVsName := DeriveHostNameShardVS(host, key)

			if shardVsName == "" {
				// If we aren't able to derive the ShardVS name, we should return
				return
			}
			model_name := lib.GetModelName(utils.ADMIN_NS, shardVsName)
			found, aviModel := objects.SharedAviGraphLister().Get(model_name)
			if !found || aviModel == nil {
				utils.AviLog.Warning.Printf("key :%s, msg: model not found during delete: %s", key, model_name)
				continue
			}
			// Delete the pool corresponding to this host
			if hostType == "secure" {
				aviModel.(*AviObjectGraph).DeletePoolForHostname(shardVsName, namespace, ingress, host, key, true)
			} else {
				aviModel.(*AviObjectGraph).DeletePoolForHostname(shardVsName, namespace, ingress, host, key, false)
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

func sniNodeHostName(tlssetting TlsSettings, ingName, namespace, key string, fullsync bool, sharedQueue *utils.WorkerQueue, modelList *[]string) []string {
	var allSniHosts []string
	for sniHost, _ := range tlssetting.Hosts {
		var sniHosts []string
		sniHosts = append(sniHosts, sniHost)
		allSniHosts = append(allSniHosts, sniHost)
		shardVsName := DeriveHostNameShardVS(sniHost, key)
		// For each host, create a SNI node with the secret giving us the key and cert.
		// construct a SNI VS node per tls setting which corresponds to one secret
		if shardVsName == "" {
			// If we aren't able to derive the ShardVS name, we should return
			return allSniHosts
		}
		model_name := lib.GetModelName(utils.ADMIN_NS, shardVsName)
		found, aviModel := objects.SharedAviGraphLister().Get(model_name)
		if !found || aviModel == nil {
			utils.AviLog.Info.Printf("key :%s, msg: model not found, generating new model with name: %s", key, model_name)
			aviModel = NewAviObjectGraph()
			aviModel.(*AviObjectGraph).ConstructAviL7VsNode(shardVsName, key)
		}
		vsNode := aviModel.(*AviObjectGraph).GetAviVS()

		sniNode := &AviVsNode{
			Name:         lib.GetSniNodeName(ingName, namespace, tlssetting.SecretName, sniHost),
			VHParentName: vsNode[0].Name,
			Tenant:       utils.ADMIN_NS,
			IsSNIChild:   true,
			ServiceMetadata: avicache.ServiceMetadataObj{
				IngressName: ingName,
				Namespace:   namespace,
				HostNames:   sniHosts,
			},
		}
		sniNode.VrfContext = lib.GetVrf()
		certsBuilt := aviModel.(*AviObjectGraph).BuildTlsCertNode(sniNode, namespace, tlssetting.SecretName, key)
		if certsBuilt {
			aviModel.(*AviObjectGraph).BuildPolicyPGPoolsForSNI(vsNode, sniNode, namespace, ingName, tlssetting, tlssetting.SecretName, key, sniHost)
			foundSniModel := FindAndReplaceSniInModel(sniNode, vsNode, key)
			if !foundSniModel {
				vsNode[0].SniNodes = append(vsNode[0].SniNodes, sniNode)
			}
			aviModel.(*AviObjectGraph).BuildPolicyRedirectForVS(vsNode, allSniHosts, namespace, ingName, key)
		} else {
			// Since the cert couldn't be built, remove the sni node from the model
			RemoveSniInModel(sniNode.Name, vsNode, key)
		}
		// Only add this node to the list of models if the checksum has changed.
		modelChanged := saveAviModel(model_name, aviModel.(*AviObjectGraph), key)
		if !utils.HasElem(*modelList, model_name) && modelChanged {
			*modelList = append(*modelList, model_name)
		}
	}

	return allSniHosts
}
