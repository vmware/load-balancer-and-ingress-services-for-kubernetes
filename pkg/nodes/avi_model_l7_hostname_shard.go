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
	pgName := vsName + utils.L7_PG_PREFIX
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
		poolName := priorityLabel + "--" + namespace + "--" + ingName
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

			poolNode := &AviPoolNode{Name: poolName, IngressName: ingName, Tenant: utils.ADMIN_NS, PriorityLabel: priorityLabel, Port: obj.Port, ServiceMetadata: avicache.ServiceMetadataObj{IngressName: ingName, Namespace: namespace, HostName: hostname}}
			poolNode.VrfContext = lib.GetVrf()
			if servers := PopulateServers(poolNode, namespace, obj.ServiceName, key); servers != nil {
				poolNode.Servers = servers
			}
			poolNode.CalculateCheckSum()
			o.AddModelNode(poolNode)
			utils.AviLog.Info.Printf("key: %s, msg: the pools before append are: %v", key, utils.Stringify(vsNode[0].PoolRefs))
			vsNode[0].PoolRefs = append(vsNode[0].PoolRefs, poolNode)
		}

	}
	// Reset the PG Node members and rebuild them
	pgNode.Members = nil
	for _, poolNode := range vsNode[0].PoolRefs {
		pool_ref := fmt.Sprintf("/api/pool?name=%s", poolNode.Name)
		pgNode.Members = append(pgNode.Members, &avimodels.PoolGroupMember{PoolRef: &pool_ref, PriorityLabel: &poolNode.PriorityLabel})

	}
}

func (o *AviObjectGraph) DeletePoolForHostname(vsName, namespace, ingName, hostname, key string) {
	o.Lock.Lock()
	defer o.Lock.Unlock()

	vsNode := o.GetAviVS()

	// Fetch the ingress pools that are present in the model and delete them.
	poolNodes := o.GetAviPoolNodesByIngress(namespace, ingName)
	utils.AviLog.Info.Printf("key: %s, msg: Pool Nodes to delete for ingress:  %s", key, utils.Stringify(poolNodes))

	for _, pool := range poolNodes {
		// It might be safe to remove all the pools for this VS for this ingress in one shot.
		o.RemovePoolNodeRefs(pool.Name)
	}
	// Generate SNI nodes and mark them for deletion. SNI node names: ingressname--namespace--secretname
	// Fetch all the secrets for this ingress
	found, secrets := objects.SharedSvcLister().IngressMappings(namespace).GetIngToSecret(ingName)
	utils.AviLog.Info.Printf("key: %s, msg: retrieved secrets for ingress: %s", key, secrets)
	if found {
		for _, secret := range secrets {
			sniNodeName := ingName + "--" + namespace + "--" + secret + "--" + hostname
			utils.AviLog.Info.Printf("key: %s, msg: sni node to delete :%s", key, sniNodeName)
			RemoveSniInModel(sniNodeName, vsNode, key)
		}
	}
	var hosts []string
	hosts = append(hosts, hostname)

	// Remove these hosts from the overall FQDN list
	RemoveFQDNsFromModel(vsNode[0], hosts, key)
	// Reset the PG Node members and rebuild them
	pgName := vsName + utils.L7_PG_PREFIX
	pgNode := o.GetPoolGroupByName(pgName)
	pgNode.Members = nil
	for _, poolNode := range vsNode[0].PoolRefs {
		pool_ref := fmt.Sprintf("/api/pool?name=%s", poolNode.Name)
		pgNode.Members = append(pgNode.Members, &avimodels.PoolGroupMember{PoolRef: &pool_ref, PriorityLabel: &poolNode.PriorityLabel})
	}
	utils.AviLog.Info.Printf("key: %s, msg: after removing fqdn refs in vs : %s", key, vsNode[0].VSVIPRefs[0].FQDNs)

}

func hostNameShardAndPublish(ingress, namespace, key string, fullsync bool, sharedQueue *utils.WorkerQueue) {
	var ingObj interface{}
	var err error
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
		processIng := true
		if lib.GetIngressApi() == utils.ExtV1IngressInformer {
			processIng = filterIngressOnClassExtV1(ingObj.(*extensionv1beta1.Ingress))
			if !processIng {
				// If the ingress class is not right, let's delete it.
				DeletePoolsByHostname(namespace, ingress, key, fullsync, sharedQueue)
			}
			parsedIng = parseHostPathForIngress(namespace, ingress, ingObj.(*extensionv1beta1.Ingress).Spec, key)
		} else {
			processIng = filterIngressOnClass(ingObj.(*v1beta1.Ingress))
			if !processIng {
				// If the ingress class is not right, let's delete it.
				DeletePoolsByHostname(namespace, ingress, key, fullsync, sharedQueue)
			}
			parsedIng = parseHostPathForIngressCoreV1(namespace, ingress, ingObj.(*v1beta1.Ingress).Spec, key)
		}
		if processIng {
			// Check if this ingress and had any previous mappings, if so - delete them first.
			ok, Storedhosts := objects.SharedSvcLister().IngressMappings(namespace).GetIngToHost(ingress)
			if ok {
				for _, host := range Storedhosts {
					shardVsName := DeriveHostNameShardVS(host, key)

					if shardVsName == "" {
						// If we aren't able to derive the ShardVS name, we should return
						return
					}
					model_name := utils.ADMIN_NS + "/" + shardVsName
					found, aviModel := objects.SharedAviGraphLister().Get(model_name)
					if !found || aviModel == nil {
						utils.AviLog.Warning.Printf("key :%s, msg: model not found during delete: %s", key, model_name)
						continue
					}
					// Delete the pool corresponding to this host
					aviModel.(*AviObjectGraph).DeletePoolForHostname(shardVsName, namespace, ingress, host, key)
					ok := saveAviModel(model_name, aviModel.(*AviObjectGraph), key)
					if ok && len(aviModel.(*AviObjectGraph).GetOrderedNodes()) != 0 && !fullsync {
						PublishKeyToRestLayer(aviModel.(*AviObjectGraph), model_name, key, sharedQueue)
					}
				}
			}
			// Process insecure routes first.
			var hosts []string
			for host, pathsvcmap := range parsedIng.IngressHostMap {
				hosts = append(hosts, host)
				shardVsName := DeriveHostNameShardVS(host, key)
				if shardVsName == "" {
					// If we aren't able to derive the ShardVS name, we should return
					return
				}
				model_name := utils.ADMIN_NS + "/" + shardVsName
				found, aviModel := objects.SharedAviGraphLister().Get(model_name)
				if !found || aviModel == nil {
					utils.AviLog.Info.Printf("key :%s, msg: model not found, generating new model with name: %s", key, model_name)
					aviModel = NewAviObjectGraph()
					aviModel.(*AviObjectGraph).ConstructAviL7VsNode(shardVsName, key)
				}
				aviModel.(*AviObjectGraph).BuildL7VSGraphHostNameShard(shardVsName, namespace, ingress, host, pathsvcmap, key)
				ok := saveAviModel(model_name, aviModel.(*AviObjectGraph), key)
				if ok && len(aviModel.(*AviObjectGraph).GetOrderedNodes()) != 0 && !fullsync {
					PublishKeyToRestLayer(aviModel.(*AviObjectGraph), model_name, key, sharedQueue)
				}
			}
			var sniHosts []string
			// Process secure routes next.
			for _, tlssetting := range parsedIng.TlsCollection {
				locSniHost := sniNodeHostName(tlssetting, ingress, namespace, key, fullsync, sharedQueue)
				sniHosts = append(sniHosts, locSniHost...)
			}
			hosts = append(hosts, sniHosts...)
			objects.SharedSvcLister().IngressMappings(namespace).UpdateIngToHostMapping(ingress, hosts)

		}
	}
}

func DeletePoolsByHostname(namespace, ingress, key string, fullsync bool, sharedQueue *utils.WorkerQueue) {
	ok, hosts := objects.SharedSvcLister().IngressMappings(namespace).GetIngToHost(ingress)
	if !ok {
		utils.AviLog.Warning.Printf("key :%s, msg: nothing to delete for ingress: %s", key, ingress)
		return
	}
	utils.AviLog.Info.Printf("key :%s, msg: hosts to delete are: :%s", key, hosts)
	for _, host := range hosts {
		shardVsName := DeriveHostNameShardVS(host, key)

		if shardVsName == "" {
			// If we aren't able to derive the ShardVS name, we should return
			return
		}
		model_name := utils.ADMIN_NS + "/" + shardVsName
		found, aviModel := objects.SharedAviGraphLister().Get(model_name)
		if !found || aviModel == nil {
			utils.AviLog.Warning.Printf("key :%s, msg: model not found during delete: %s", key, model_name)
			continue
		}
		// Delete the pool corresponding to this host
		aviModel.(*AviObjectGraph).DeletePoolForHostname(shardVsName, namespace, ingress, host, key)
		ok := saveAviModel(model_name, aviModel.(*AviObjectGraph), key)
		if ok && len(aviModel.(*AviObjectGraph).GetOrderedNodes()) != 0 && !fullsync {
			PublishKeyToRestLayer(aviModel.(*AviObjectGraph), model_name, key, sharedQueue)
		}
	}
	// Now remove the secret relationship
	objects.SharedSvcLister().IngressMappings(namespace).RemoveIngressSecretMappings(ingress)
	// Remove the hosts mapping for this ingress
	objects.SharedSvcLister().IngressMappings(namespace).DeleteIngToHostMapping(ingress)
}

func sniNodeHostName(tlssetting TlsSettings, ingName, namespace, key string, fullsync bool, sharedQueue *utils.WorkerQueue) []string {
	var sniHosts []string
	for sniHost, _ := range tlssetting.Hosts {
		sniHosts = append(sniHosts, sniHost)
		shardVsName := DeriveHostNameShardVS(sniHost, key)
		// For each host, create a SNI node with the secret giving us the key and cert.
		// construct a SNI VS node per tls setting which corresponds to one secret
		if shardVsName == "" {
			// If we aren't able to derive the ShardVS name, we should return
			return sniHosts
		}
		model_name := utils.ADMIN_NS + "/" + shardVsName
		found, aviModel := objects.SharedAviGraphLister().Get(model_name)
		if !found || aviModel == nil {
			utils.AviLog.Info.Printf("key :%s, msg: model not found, generating new model with name: %s", key, model_name)
			aviModel = NewAviObjectGraph()
			aviModel.(*AviObjectGraph).ConstructAviL7VsNode(shardVsName, key)
		}
		vsNode := aviModel.(*AviObjectGraph).GetAviVS()

		sniNode := &AviVsNode{Name: ingName + "--" + namespace + "--" + tlssetting.SecretName + "--" + sniHost, VHParentName: vsNode[0].Name, Tenant: utils.ADMIN_NS, IsSNIChild: true}
		sniNode.VrfContext = lib.GetVrf()
		certsBuilt := aviModel.(*AviObjectGraph).BuildTlsCertNode(sniNode, namespace, tlssetting.SecretName, key)
		if certsBuilt {
			aviModel.(*AviObjectGraph).BuildPolicyPGPoolsForSNI(vsNode, sniNode, namespace, ingName, tlssetting, tlssetting.SecretName, key)
			foundSniModel := FindAndReplaceSniInModel(sniNode, vsNode, key)
			if !foundSniModel {
				vsNode[0].SniNodes = append(vsNode[0].SniNodes, sniNode)
			}

		}
		ok := saveAviModel(model_name, aviModel.(*AviObjectGraph), key)
		if ok && len(aviModel.(*AviObjectGraph).GetOrderedNodes()) != 0 && !fullsync {
			PublishKeyToRestLayer(aviModel.(*AviObjectGraph), model_name, key, sharedQueue)
		}
	}
	return sniHosts
}
