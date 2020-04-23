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

	avicache "ako/pkg/cache"
	"ako/pkg/lib"
	"ako/pkg/objects"

	"github.com/avinetworks/container-lib/utils"
	avimodels "github.com/avinetworks/sdk/go/models"
	extensionv1beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TODO: Move to utils
const tlsCert = "tls.crt"

func (o *AviObjectGraph) BuildL7VSGraph(vsName string, namespace string, ingName string, key string) {
	o.Lock.Lock()
	defer o.Lock.Unlock()
	// We create pools and attach servers to them here. Pools are created with a priorty label of host/path
	utils.AviLog.Info.Printf("key: %s, msg: Building the L7 pools for namespace: %s, ingName: %s", key, namespace, ingName)
	var err error
	var ingObj interface{}
	if lib.GetIngressApi() == utils.ExtV1IngressInformer {
		ingObj, err = utils.GetInformers().ExtV1IngressInformer.Lister().Ingresses(namespace).Get(ingName)
	} else {
		ingObj, err = utils.GetInformers().CoreV1IngressInformer.Lister().Ingresses(namespace).Get(ingName)
	}
	pgName := lib.GetL7SharedPGName(vsName)
	pgNode := o.GetPoolGroupByName(pgName)
	vsNode := o.GetAviVS()
	if len(vsNode) != 1 {
		utils.AviLog.Warning.Printf("key: %s, msg: more than one vs in model.", key)
		return
	}
	if err != nil {
		o.DeletePoolForIngress(namespace, ingName, key, vsNode)
	} else {
		var parsedIng IngressConfig
		processIng := true
		if lib.GetIngressApi() == utils.ExtV1IngressInformer {
			processIng = filterIngressOnClassExtV1(ingObj.(*extensionv1beta1.Ingress))
			if !processIng {
				// If the ingress class is not right, let's delete it.
				o.DeletePoolForIngress(namespace, ingName, key, vsNode)
			}
			parsedIng = o.Validator.ParseHostPathForIngress(namespace, ingName, ingObj.(*extensionv1beta1.Ingress).Spec, key)
		} else {
			processIng = filterIngressOnClass(ingObj.(*v1beta1.Ingress))
			if !processIng {
				// If the ingress class is not right, let's delete it.
				o.DeletePoolForIngress(namespace, ingName, key, vsNode)
			}
			parsedIng = o.Validator.ParseHostPathForIngressCoreV1(namespace, ingName, ingObj.(*v1beta1.Ingress).Spec, key)
		}
		if processIng {
			// First check if there are pools related to this ingress present in the model already
			poolNodes := o.GetAviPoolNodesByIngress(namespace, ingName)
			utils.AviLog.Info.Printf("key: %s, msg: found pools in the model: %s", key, utils.Stringify(poolNodes))
			for _, pool := range poolNodes {
				o.RemovePoolNodeRefs(pool.Name)
			}

			// First retrieve the FQDNs from the cache and update the model
			ok, hostMap := objects.SharedSvcLister().IngressMappings(namespace).GetIngToHost(ingName)
			var storedHosts []string
			// Mash the list of secure and insecure hosts.
			for mtype, hostsbytype := range hostMap {
				for host, _ := range hostsbytype {
					if mtype == "secure" {
						RemoveRedirectHTTPPolicyInModel(vsNode[0], host, key)
					}
					storedHosts = append(storedHosts, host)
				}
			}
			if ok {
				RemoveFQDNsFromModel(vsNode[0], storedHosts, key)
			}

			// Update the host mappings for this ingress
			// Generate SNI nodes and mark them for deletion. SNI node names: ingressname--namespace--secretname
			// Fetch all the secrets for this ingress
			found, secrets := objects.SharedSvcLister().IngressMappings(namespace).GetIngToSecret(ingName)
			utils.AviLog.Info.Printf("key: %s, msg: retrieved secrets for ingress: %s", key, secrets)
			if found {
				for _, secret := range secrets {
					sniNodeName := lib.GetSniNodeName(ingName, namespace, secret)
					RemoveSniInModel(sniNodeName, vsNode, key)
				}
			}

			utils.AviLog.Info.Printf("key: %s, msg: parsedIng value: %v", key, parsedIng)
			newHostMap := make(map[string]map[string][]string)
			insecureHostPathMapArr := make(map[string][]string)
			for host, pathmap := range parsedIng.IngressHostMap {
				insecureHostPathMapArr[host] = getPaths(pathmap)
			}
			newHostMap["insecure"] = insecureHostPathMapArr
			secureHostPathMapArr := make(map[string][]string)
			for _, tlssetting := range parsedIng.TlsCollection {
				for sniHost, paths := range tlssetting.Hosts {
					secureHostPathMapArr[sniHost] = getPaths(paths)
				}
			}
			newHostMap["secure"] = secureHostPathMapArr
			objects.SharedSvcLister().IngressMappings(namespace).UpdateIngToHostMapping(ingName, newHostMap)
			// PGs are in 'admin' namespace right now.
			if pgNode != nil {
				utils.AviLog.Info.Printf("key: %s, msg: hostpathsvc list: %s", key, utils.Stringify(parsedIng))
				// Processsing insecure ingress
				for host, val := range parsedIng.IngressHostMap {
					if !utils.HasElem(vsNode[0].VSVIPRefs[0].FQDNs, host) {
						vsNode[0].VSVIPRefs[0].FQDNs = append(vsNode[0].VSVIPRefs[0].FQDNs, host)
					}
					for _, obj := range val {
						var priorityLabel string
						var hostSlice []string
						if obj.Path != "" {
							priorityLabel = host + obj.Path
						} else {
							priorityLabel = host
						}
						hostSlice = append(hostSlice, host)
						poolNode := &AviPoolNode{Name: lib.GetL7PoolName(priorityLabel, namespace, ingName), IngressName: ingName, Tenant: utils.ADMIN_NS, PriorityLabel: priorityLabel, Port: obj.Port, ServiceMetadata: avicache.ServiceMetadataObj{IngressName: ingName, Namespace: namespace, HostNames: hostSlice}}
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

				// Processing the TLS nodes
				for _, tlssetting := range parsedIng.TlsCollection {
					// For each host, create a SNI node with the secret giving us the key and cert.
					// construct a SNI VS node per tls setting which corresponds to one secret
					sniNode := &AviVsNode{
						Name:         lib.GetSniNodeName(ingName, namespace, tlssetting.SecretName),
						VHParentName: vsNode[0].Name,
						Tenant:       utils.ADMIN_NS,
						IsSNIChild:   true,
						ServiceMetadata: avicache.ServiceMetadataObj{
							IngressName: ingName,
							Namespace:   namespace,
						},
					}
					sniNode.VrfContext = lib.GetVrf()
					certsBuilt := o.BuildTlsCertNode(sniNode, namespace, tlssetting.SecretName, key)
					if certsBuilt {
						o.BuildPolicyPGPoolsForSNI(vsNode, sniNode, namespace, ingName, tlssetting, tlssetting.SecretName, key)
						foundSniModel := FindAndReplaceSniInModel(sniNode, vsNode, key)
						if !foundSniModel {
							vsNode[0].SniNodes = append(vsNode[0].SniNodes, sniNode)
						}
						sniNode.ServiceMetadata = avicache.ServiceMetadataObj{IngressName: ingName, Namespace: namespace, HostNames: sniNode.VHDomainNames}
						for _, hostname := range sniNode.VHDomainNames {
							o.BuildPolicyRedirectForVS(vsNode, hostname, namespace, ingName, key)
						}
					}

				}
			}
		}
	}
	// Reset the PG Node members and rebuild them
	pgNode.Members = nil
	for _, poolNode := range vsNode[0].PoolRefs {
		pool_ref := fmt.Sprintf("/api/pool?name=%s", poolNode.Name)
		pgNode.Members = append(pgNode.Members, &avimodels.PoolGroupMember{PoolRef: &pool_ref, PriorityLabel: &poolNode.PriorityLabel})
	}

}

func (o *AviObjectGraph) DeletePoolForIngress(namespace, ingName, key string, vsNode []*AviVsNode) {
	// A case, where we detected in Layer 2 that the ingress has been deleted.
	utils.AviLog.Info.Printf("key: %s, msg: ingress not found:  %s", key, ingName)

	// Fetch the ingress pools that are present in the model and delete them.
	poolNodes := o.GetAviPoolNodesByIngress(namespace, ingName)
	utils.AviLog.Info.Printf("key: %s, msg: Pool Nodes to delete for ingress:  %s", key, utils.Stringify(poolNodes))

	for _, pool := range poolNodes {
		o.RemovePoolNodeRefs(pool.Name)
	}
	// Generate SNI nodes and mark them for deletion. SNI node names: ingressname--namespace--secretname
	// Fetch all the secrets for this ingress

	found, secrets := objects.SharedSvcLister().IngressMappings(namespace).GetIngToSecret(ingName)
	utils.AviLog.Info.Printf("key: %s, msg: retrieved secrets for ingress: %s", key, secrets)
	if found {
		for _, secret := range secrets {
			sniNodeName := lib.GetSniNodeName(ingName, namespace, secret)
			utils.AviLog.Info.Printf("key: %s, msg: sni node to delete :%s", key, sniNodeName)
			RemoveSniInModel(sniNodeName, vsNode, key)
		}
	}
	ok, hostMap := objects.SharedSvcLister().IngressMappings(namespace).GetIngToHost(ingName)
	var hosts []string
	// Mash the list of secure and insecure hosts.
	for mtype, hostsbytype := range hostMap {
		for host, _ := range hostsbytype {
			if mtype == "secure" {
				RemoveRedirectHTTPPolicyInModel(vsNode[0], host, key)
			}
			hosts = append(hosts, host)
		}
	}
	if ok {
		// Remove these hosts from the overall FQDN list
		RemoveFQDNsFromModel(vsNode[0], hosts, key)
	}
	utils.AviLog.Info.Printf("key: %s, msg: after removing fqdn refs in vs : %s", key, vsNode[0].VSVIPRefs[0].FQDNs)

	// Now remove the secret relationship
	objects.SharedSvcLister().IngressMappings(namespace).RemoveIngressSecretMappings(ingName)
	// Remove the hosts mapping for this ingress
	objects.SharedSvcLister().IngressMappings(namespace).DeleteIngToHostMapping(ingName)

}

func RemoveFQDNsFromModel(vsNode *AviVsNode, hosts []string, key string) {
	if len(vsNode.VSVIPRefs) > 0 {
		newFQDNs := make([]string, len(vsNode.VSVIPRefs[0].FQDNs))
		copy(newFQDNs, vsNode.VSVIPRefs[0].FQDNs)
		utils.AviLog.Info.Printf("key: %s, msg: found fqdn refs in vs : %s", key, vsNode.VSVIPRefs[0].FQDNs)
		for _, host := range hosts {
			var i int
			for _, fqdn := range vsNode.VSVIPRefs[0].FQDNs {
				if fqdn != host {
					// Gather this entry in the new list
					if len(newFQDNs) == 0 {
						break
					}
					newFQDNs[i] = fqdn
					i++
				}
			}
			// Empty unsed bytes.
			if len(newFQDNs) != 0 {
				newFQDNs = newFQDNs[:i]
			}
		}
		vsNode.VSVIPRefs[0].FQDNs = newFQDNs
	}
}

func FindAndReplaceSniInModel(currentSniNode *AviVsNode, modelSniNodes []*AviVsNode, key string) bool {
	if len(modelSniNodes[0].SniNodes) > 0 {
		for i, modelSniNode := range modelSniNodes[0].SniNodes {
			if currentSniNode.Name == modelSniNode.Name {
				// Check if the checksums are same
				if !(modelSniNode.GetCheckSum() == currentSniNode.GetCheckSum()) {
					// The checksums are not same. Replace this sni node
					modelSniNodes[0].SniNodes = append(modelSniNodes[0].SniNodes[:i], modelSniNodes[0].SniNodes[i+1:]...)
					modelSniNodes[0].SniNodes = append(modelSniNodes[0].SniNodes, currentSniNode)
					utils.AviLog.Info.Printf("key: %s, msg: replaced sni node in model: %s", key, currentSniNode.Name)
				}
				return true
			}
		}
	}
	return false
}

func RemoveSniInModel(currentSniNodeName string, modelSniNodes []*AviVsNode, key string) {
	if len(modelSniNodes[0].SniNodes) > 0 {
		for i, modelSniNode := range modelSniNodes[0].SniNodes {
			if currentSniNodeName == modelSniNode.Name {
				// Check if the checksums are same
				// The checksums are not same. Replace this sni node
				modelSniNodes[0].SniNodes = append(modelSniNodes[0].SniNodes[:i], modelSniNodes[0].SniNodes[i+1:]...)
				utils.AviLog.Info.Printf("key: %s, msg: deleted sni node in model: %s", key, currentSniNodeName)
			}
		}
	}
}

func (o *AviObjectGraph) ConstructAviL7VsNode(vsName string, key string) *AviVsNode {
	o.Lock.Lock()
	defer o.Lock.Unlock()
	var avi_vs_meta *AviVsNode

	// This is a shared VS - always created in the admin namespace for now.
	avi_vs_meta = &AviVsNode{Name: vsName, Tenant: utils.ADMIN_NS,
		EastWest: false, SharedVS: true}
	// Hard coded ports for the shared VS
	var portProtocols []AviPortHostProtocol
	httpPort := AviPortHostProtocol{Port: 80, Protocol: utils.HTTP}
	httpsPort := AviPortHostProtocol{Port: 443, Protocol: utils.HTTP, EnableSSL: true}
	portProtocols = append(portProtocols, httpPort)
	portProtocols = append(portProtocols, httpsPort)
	avi_vs_meta.PortProto = portProtocols
	// Default case.
	avi_vs_meta.ApplicationProfile = utils.DEFAULT_L7_APP_PROFILE
	avi_vs_meta.NetworkProfile = utils.DEFAULT_TCP_NW_PROFILE
	avi_vs_meta.SNIParent = true

	vrfcontext := lib.GetVrf()
	avi_vs_meta.VrfContext = vrfcontext

	o.AddModelNode(avi_vs_meta)
	o.ConstructShardVsPGNode(vsName, key, avi_vs_meta)
	o.ConstructHTTPDataScript(vsName, key, avi_vs_meta)
	var fqdns []string

	subDomains := GetDefaultSubDomain()
	if subDomains != nil {
		var fqdn string
		if strings.HasPrefix(subDomains[0], ".") {
			fqdn = vsName + "." + utils.ADMIN_NS + subDomains[0]
		} else {
			fqdn = vsName + "." + utils.ADMIN_NS + "." + subDomains[0]
		}
		fqdns = append(fqdns, fqdn)
	} else {
		utils.AviLog.Warning.Printf("key: %s, msg: there is no nsipamdns configured in the cloud, not configuring the default fqdn", key)
	}
	vsVipNode := &AviVSVIPNode{Name: lib.GetVsVipName(vsName), Tenant: utils.ADMIN_NS, FQDNs: fqdns,
		EastWest: false, VrfContext: vrfcontext}
	avi_vs_meta.VSVIPRefs = append(avi_vs_meta.VSVIPRefs, vsVipNode)
	return avi_vs_meta
}

func (o *AviObjectGraph) ConstructShardVsPGNode(vsName string, key string, vsNode *AviVsNode) *AviPoolGroupNode {
	pgName := lib.GetL7SharedPGName(vsName)
	pgNode := &AviPoolGroupNode{Name: pgName, Tenant: utils.ADMIN_NS, ImplicitPriorityLabel: true}
	vsNode.PoolGroupRefs = append(vsNode.PoolGroupRefs, pgNode)
	o.AddModelNode(pgNode)
	return pgNode
}

func (o *AviObjectGraph) ConstructHTTPDataScript(vsName string, key string, vsNode *AviVsNode) *AviHTTPDataScriptNode {
	scriptStr := utils.HTTP_DS_SCRIPT
	evt := utils.VS_DATASCRIPT_EVT_HTTP_REQ
	var poolGroupRefs []string
	pgName := lib.GetL7SharedPGName(vsName)
	poolGroupRefs = append(poolGroupRefs, pgName)
	dsName := lib.GetL7InsecureDSName(vsName)
	script := &DataScript{Script: scriptStr, Evt: evt}
	dsScriptNode := &AviHTTPDataScriptNode{Name: dsName, Tenant: utils.ADMIN_NS, DataScript: script, PoolGroupRefs: poolGroupRefs}
	if len(dsScriptNode.PoolGroupRefs) > 0 {
		dsScriptNode.Script = strings.Replace(dsScriptNode.Script, "POOLGROUP", dsScriptNode.PoolGroupRefs[0], 1)
	}
	vsNode.HTTPDSrefs = append(vsNode.HTTPDSrefs, dsScriptNode)
	o.AddModelNode(dsScriptNode)
	return dsScriptNode
}

func (o *AviObjectGraph) BuildTlsCertNode(tlsNode *AviVsNode, namespace string, secretName string, key string) bool {
	mClient := utils.GetInformers().ClientSet
	secretObj, err := mClient.CoreV1().Secrets(namespace).Get(secretName, metav1.GetOptions{})
	if err != nil || secretObj == nil {
		// This secret has been deleted.
		ok, ingNames := objects.SharedSvcLister().IngressMappings(namespace).GetSecretToIng(secretName)
		if ok {
			// Delete the secret key in the cache if it has no references
			if len(ingNames) == 0 {
				objects.SharedSvcLister().IngressMappings(namespace).DeleteSecretToIngMapping(secretName)
			}
		}
		utils.AviLog.Info.Printf("key: %s, msg: secret: %s has been deleted, err: %s", key, secretName, err)
		return false
	}
	certNode := &AviTLSKeyCertNode{Name: lib.GetTLSKeyCertNodeName(namespace, secretName), Tenant: utils.ADMIN_NS}
	keycertMap := secretObj.Data
	cert, ok := keycertMap[tlsCert]
	if ok {
		certNode.Cert = cert
	} else {
		utils.AviLog.Info.Printf("key: %s, msg: certificate not found for secret: %s", key, secretObj.Name)
		return false
	}
	tlsKey, keyfound := keycertMap[utils.K8S_TLS_SECRET_KEY]
	if keyfound {
		certNode.Key = tlsKey
	} else {
		utils.AviLog.Info.Printf("key: %s, msg: key not found for secret: %s", key, secretObj.Name)
		return false
	}
	utils.AviLog.Info.Printf("key: %s, msg: Added the secret object to tlsnode: %s", key, secretObj.Name)

	tlsNode.SSLKeyCertRefs = append(tlsNode.SSLKeyCertRefs, certNode)
	return true
}

func (o *AviObjectGraph) BuildPolicyPGPoolsForSNI(vsNode []*AviVsNode, tlsNode *AviVsNode, namespace string, ingName string, hostpath TlsSettings, secretName string, key string, hostName ...string) {
	for host, paths := range hostpath.Hosts {
		if len(hostName) > 0 {
			if hostName[0] != host {
				// If a hostname is passed to this method, ensure we only process that hostname and nothing else.
				continue
			}
		}
		// Update secret --> hostname mapping
		objects.SharedSvcLister().IngressMappings(namespace).UpdateSecretToHostNameMapping(secretName, host)
		// Update the VSVIP with the host information.
		if !utils.HasElem(vsNode[0].VSVIPRefs[0].FQDNs, host) {
			vsNode[0].VSVIPRefs[0].FQDNs = append(vsNode[0].VSVIPRefs[0].FQDNs, host)
		}
		tlsNode.VHDomainNames = append(tlsNode.VHDomainNames, host)
		for _, path := range paths {
			var httpPolicySet []AviHostPathPortPoolPG

			httpPGPath := AviHostPathPortPoolPG{Host: host}
			httpPGPath.Path = append(httpPGPath.Path, path.Path)
			httpPGPath.MatchCriteria = "BEGINS_WITH"
			pgName := lib.GetSniPGName(ingName, namespace, host, path.Path)
			pgNode := &AviPoolGroupNode{Name: pgName, Tenant: utils.ADMIN_NS}
			httpPGPath.PoolGroup = pgNode.Name
			httpPGPath.Host = host
			httpPolicySet = append(httpPolicySet, httpPGPath)

			tlsNode.PoolGroupRefs = append(tlsNode.PoolGroupRefs, pgNode)

			poolNode := &AviPoolNode{Name: lib.GetSniPoolName(ingName, namespace, host, path.Path), Tenant: utils.ADMIN_NS}
			poolNode.VrfContext = lib.GetVrf()

			if servers := PopulateServers(poolNode, namespace, path.ServiceName, key); servers != nil {
				poolNode.Servers = servers
			}
			pool_ref := fmt.Sprintf("/api/pool?name=%s", poolNode.Name)
			pgNode.Members = append(pgNode.Members, &avimodels.PoolGroupMember{PoolRef: &pool_ref})

			tlsNode.PoolRefs = append(tlsNode.PoolRefs, poolNode)
			httppolname := lib.GetSniHttpPolName(ingName, namespace, host, path.Path)
			policyNode := &AviHttpPolicySetNode{Name: httppolname, HppMap: httpPolicySet, Tenant: utils.ADMIN_NS}
			tlsNode.HttpPolicyRefs = append(tlsNode.HttpPolicyRefs, policyNode)
		}
	}
	utils.AviLog.Info.Printf("key: %s, msg: added pools and poolgroups to tlsNode: %s", key, tlsNode.Name)

}

func (o *AviObjectGraph) BuildPolicyRedirectForVS(vsNode []*AviVsNode, hostname string, namespace, ingName, key string) {
	policyname := lib.GetL7HttpRedirPolicy(vsNode[0].Name)
	myHppMap := AviRedirectPort{
		Hosts:        []string{hostname},
		RedirectPort: 443,
		StatusCode:   lib.STATUS_REDIRECT,
		VsPort:       80,
	}

	redirectPolicy := &AviHttpPolicySetNode{
		Tenant:        utils.ADMIN_NS,
		Name:          policyname,
		RedirectPorts: []AviRedirectPort{myHppMap},
	}

	if policyFound := FindAndReplaceRedirectHTTPPolicyInModel(vsNode[0], redirectPolicy, hostname, key); !policyFound {
		redirectPolicy.CalculateCheckSum()
		vsNode[0].HttpPolicyRefs = append(vsNode[0].HttpPolicyRefs, redirectPolicy)
	}

}

func FindAndReplaceRedirectHTTPPolicyInModel(vsNode *AviVsNode, httpPolicy *AviHttpPolicySetNode, hostname, key string) bool {
	for _, policy := range vsNode.HttpPolicyRefs {
		if policy.Name == httpPolicy.Name && policy.CloudConfigCksum != httpPolicy.CloudConfigCksum {
			if !utils.HasElem(policy.RedirectPorts[0].Hosts, hostname) {
				policy.RedirectPorts[0].Hosts = append(policy.RedirectPorts[0].Hosts, hostname)
				utils.AviLog.Info.Printf("key: %s, msg: replaced host %s for policy %s in model", key, hostname, policy.Name)
			}
			return true
		}
	}
	return false
}

func RemoveRedirectHTTPPolicyInModel(vsNode *AviVsNode, hostname, key string) {
	policyName := lib.GetL7HttpRedirPolicy(vsNode.Name)
	deletePolicy := false
	for i, policy := range vsNode.HttpPolicyRefs {
		if policy.Name == policyName {
			// one redirect policy per shard vs
			policy.RedirectPorts[0].Hosts = utils.Remove(policy.RedirectPorts[0].Hosts, hostname)
			utils.AviLog.Info.Printf("key: %s, msg: removed host %s from policy %s in model %v", key, hostname, policy.Name, policy.RedirectPorts[0].Hosts)
			if len(policy.RedirectPorts[0].Hosts) == 0 {
				deletePolicy = true
			}

			if deletePolicy {
				vsNode.HttpPolicyRefs = append(vsNode.HttpPolicyRefs[:i], vsNode.HttpPolicyRefs[i+1:]...)
				utils.AviLog.Info.Printf("key: %s, msg: removed policy %s in model", key, policy.Name)
			}
		}
	}
}
