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
	"context"
	"fmt"
	"strings"

	avicache "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	avimodels "github.com/avinetworks/sdk/go/models"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TODO: Move to utils
const tlsCert = "tls.crt"

func (o *AviObjectGraph) BuildL7VSGraph(vsName string, namespace string, ingName string, key string) {
	o.Lock.Lock()
	defer o.Lock.Unlock()
	// We create pools and attach servers to them here. Pools are created with a priorty label of host/path
	utils.AviLog.Infof("key: %s, msg: Building the L7 pools for namespace: %s, ingName: %s", key, namespace, ingName)

	ingObj, err := utils.GetInformers().IngressInformer.Lister().Ingresses(namespace).Get(ingName)

	pgName := lib.GetL7SharedPGName(vsName)
	pgNode := o.GetPoolGroupByName(pgName)
	vsNode := o.GetAviVS()
	if len(vsNode) != 1 {
		utils.AviLog.Warnf("key: %s, msg: more than one vs in model.", key)
		return
	}
	if err != nil {
		o.DeletePoolForIngress(namespace, ingName, key, vsNode)
	} else {
		var parsedIng IngressConfig
		processIng := true
		processIng = validateIngressForClass(key, ingObj) && utils.CheckIfNamespaceAccepted(namespace, utils.GetGlobalNSFilter(), nil, true)
		if !processIng {
			// If the ingress class is not right, let's delete it.
			o.DeletePoolForIngress(namespace, ingName, key, vsNode)
		}
		parsedIng = o.Validator.ParseHostPathForIngress(namespace, ingName, ingObj.Spec, key)
		if processIng {
			// First check if there are pools related to this ingress present in the model already
			poolNodes := o.GetAviPoolNodesByIngress(namespace, ingName)
			utils.AviLog.Infof("key: %s, msg: found pools in the model: %s", key, utils.Stringify(poolNodes))
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
			// Generate SNI nodes and mark them for deletion. SNI node names: ingressname--namespace-secretname
			// Fetch all the secrets for this ingress
			found, secrets := objects.SharedSvcLister().IngressMappings(namespace).GetIngToSecret(ingName)
			utils.AviLog.Infof("key: %s, msg: retrieved secrets for ingress: %s", key, secrets)
			if found {
				for _, namespacedSecret := range secrets {
					_, secret := utils.ExtractNamespaceObjectName(namespacedSecret)
					if secret == "" {
						utils.AviLog.Warnf("key: %s, msg: got empty secret for ingress %s/%s", key, namespace, ingName)
						continue
					}
					sniNodeName := lib.GetSniNodeName(ingName, namespace, secret)
					RemoveSniInModel(sniNodeName, vsNode, key)
				}
			}

			utils.AviLog.Infof("key: %s, msg: parsedIng value: %v", key, parsedIng)
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

			// hostNamePathStore cache operation
			_, oldHostMap := objects.SharedSvcLister().IngressMappings(namespace).GetIngToHost(ingName)
			updateHostPathCache(namespace, ingName, oldHostMap, newHostMap)

			objects.SharedSvcLister().IngressMappings(namespace).UpdateIngToHostMapping(ingName, newHostMap)
			// PGs are in 'admin' namespace right now.
			if pgNode != nil {
				utils.AviLog.Infof("key: %s, msg: hostpathsvc list: %s", key, utils.Stringify(parsedIng))
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
						poolNode := &AviPoolNode{Name: lib.GetL7PoolName(priorityLabel, namespace, ingName), PortName: obj.PortName, IngressName: ingName, Tenant: lib.GetTenant(), PriorityLabel: priorityLabel, Port: obj.Port, ServiceMetadata: avicache.ServiceMetadataObj{IngressName: ingName, Namespace: namespace, HostNames: hostSlice}}
						poolNode.VrfContext = lib.GetVrf()
						if !lib.IsNodePortMode() {
							if servers := PopulateServers(poolNode, namespace, obj.ServiceName, true, key); servers != nil {
								poolNode.Servers = servers
							}
						} else {
							if servers := PopulateServersForNodePort(poolNode, namespace, obj.ServiceName, true, key); servers != nil {
								poolNode.Servers = servers
							}
						}
						poolNode.CalculateCheckSum()
						o.AddModelNode(poolNode)
						utils.AviLog.Infof("key: %s, msg: the pools before append are: %v", key, utils.Stringify(vsNode[0].PoolRefs))
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
						Tenant:       lib.GetTenant(),
						IsSNIChild:   true,
						ServiceMetadata: avicache.ServiceMetadataObj{
							IngressName: ingName,
							Namespace:   namespace,
						},
					}
					if lib.GetSEGName() != lib.DEFAULT_GROUP {
						sniNode.ServiceEngineGroup = lib.GetSEGName()
					}
					sniNode.VrfContext = lib.GetVrf()
					certsBuilt := o.BuildTlsCertNode(objects.SharedSvcLister(), sniNode, namespace, tlssetting, key)
					if certsBuilt {
						o.BuildPolicyPGPoolsForSNI(vsNode, sniNode, namespace, ingName, tlssetting, tlssetting.SecretName, key, true)
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
	utils.AviLog.Infof("key: %s, msg: ingress not found:  %s", key, ingName)

	// Fetch the ingress pools that are present in the model and delete them.
	poolNodes := o.GetAviPoolNodesByIngress(namespace, ingName)
	utils.AviLog.Infof("key: %s, msg: Pool Nodes to delete for ingress:  %s", key, utils.Stringify(poolNodes))

	for _, pool := range poolNodes {
		o.RemovePoolNodeRefs(pool.Name)
	}
	// Generate SNI nodes and mark them for deletion. SNI node names: ingressname--namespace-secretname
	// Fetch all the secrets for this ingress

	found, secrets := objects.SharedSvcLister().IngressMappings(namespace).GetIngToSecret(ingName)
	utils.AviLog.Infof("key: %s, msg: retrieved secrets for ingress: %s", key, secrets)
	if found {
		for _, namespacedSecret := range secrets {
			_, secret := utils.ExtractNamespaceObjectName(namespacedSecret)
			if secret == "" {
				utils.AviLog.Warnf("key: %s, msg: got empty secret while deleting pool for Ingress %s/%s", key, namespace, ingName)
				continue
			}
			sniNodeName := lib.GetSniNodeName(ingName, namespace, secret)
			utils.AviLog.Infof("key: %s, msg: sni node to delete :%s", key, sniNodeName)
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
	utils.AviLog.Infof("key: %s, msg: after removing fqdn refs in vs : %s", key, vsNode[0].VSVIPRefs[0].FQDNs)

	// Now remove the secret relationship
	objects.SharedSvcLister().IngressMappings(namespace).RemoveIngressSecretMappings(ingName)
	// Remove the hosts mapping for this ingress
	objects.SharedSvcLister().IngressMappings(namespace).DeleteIngToHostMapping(ingName)

	// remove hostpath mappings
	updateHostPathCache(namespace, ingName, hostMap, nil)
}

func RemoveFQDNsFromModel(vsNode *AviVsNode, hosts []string, key string) {
	if len(vsNode.VSVIPRefs) > 0 {
		for i, fqdn := range vsNode.VSVIPRefs[0].FQDNs {
			if utils.HasElem(hosts, fqdn) {
				// remove logic conainer-lib candidate
				vsNode.VSVIPRefs[0].FQDNs[i] = vsNode.VSVIPRefs[0].FQDNs[len(vsNode.VSVIPRefs[0].FQDNs)-1]
				vsNode.VSVIPRefs[0].FQDNs[len(vsNode.VSVIPRefs[0].FQDNs)-1] = ""
				vsNode.VSVIPRefs[0].FQDNs = vsNode.VSVIPRefs[0].FQDNs[:len(vsNode.VSVIPRefs[0].FQDNs)-1]
			}
		}
	}
}

func FindAndReplaceSniInModel(currentSniNode *AviVsNode, modelSniNodes []*AviVsNode, key string) bool {
	for i, modelSniNode := range modelSniNodes[0].SniNodes {
		if currentSniNode.Name == modelSniNode.Name {
			// Check if the checksums are same
			if !(modelSniNode.GetCheckSum() == currentSniNode.GetCheckSum()) {
				// The checksums are not same. Replace this sni node
				modelSniNodes[0].SniNodes = append(modelSniNodes[0].SniNodes[:i], modelSniNodes[0].SniNodes[i+1:]...)
				modelSniNodes[0].SniNodes = append(modelSniNodes[0].SniNodes, currentSniNode)
				utils.AviLog.Infof("key: %s, msg: replaced sni node in model: %s", key, currentSniNode.Name)
			}
			return true
		}
	}
	return false
}

func RemoveSniInModel(currentSniNodeName string, modelSniNodes []*AviVsNode, key string) {
	if len(modelSniNodes[0].SniNodes) > 0 {
		for i, modelSniNode := range modelSniNodes[0].SniNodes {
			if currentSniNodeName == modelSniNode.Name {
				modelSniNodes[0].SniNodes = append(modelSniNodes[0].SniNodes[:i], modelSniNodes[0].SniNodes[i+1:]...)
				utils.AviLog.Infof("key: %s, msg: deleted sni node in model: %s", key, currentSniNodeName)
				return
			}
		}
	}
}

func (o *AviObjectGraph) ConstructAviL7VsNode(vsName string, key string) *AviVsNode {
	o.Lock.Lock()
	defer o.Lock.Unlock()
	var avi_vs_meta *AviVsNode

	// This is a shared VS - always created in the admin namespace for now.
	avi_vs_meta = &AviVsNode{Name: vsName, Tenant: lib.GetTenant(),
		EastWest: false, SharedVS: true}
	if lib.GetSEGName() != lib.DEFAULT_GROUP {
		avi_vs_meta.ServiceEngineGroup = lib.GetSEGName()
	}
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
			fqdn = vsName + "." + lib.GetTenant() + subDomains[0]
		} else {
			fqdn = vsName + "." + lib.GetTenant() + "." + subDomains[0]
		}
		fqdns = append(fqdns, fqdn)
	} else {
		utils.AviLog.Warnf("key: %s, msg: there is no nsipamdns configured in the cloud, not configuring the default fqdn", key)
	}
	vsVipNode := &AviVSVIPNode{Name: lib.GetVsVipName(vsName), Tenant: lib.GetTenant(), FQDNs: fqdns,
		EastWest: false, VrfContext: vrfcontext}
	avi_vs_meta.VSVIPRefs = append(avi_vs_meta.VSVIPRefs, vsVipNode)
	return avi_vs_meta
}

func (o *AviObjectGraph) ConstructShardVsPGNode(vsName string, key string, vsNode *AviVsNode) *AviPoolGroupNode {
	pgName := lib.GetL7SharedPGName(vsName)
	pgNode := &AviPoolGroupNode{Name: pgName, Tenant: lib.GetTenant(), ImplicitPriorityLabel: true}
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
	dsScriptNode := &AviHTTPDataScriptNode{Name: dsName, Tenant: lib.GetTenant(), DataScript: script, PoolGroupRefs: poolGroupRefs}
	if len(dsScriptNode.PoolGroupRefs) > 0 {
		dsScriptNode.Script = strings.Replace(dsScriptNode.Script, "POOLGROUP", dsScriptNode.PoolGroupRefs[0], 1)
	}
	vsNode.HTTPDSrefs = append(vsNode.HTTPDSrefs, dsScriptNode)
	o.AddModelNode(dsScriptNode)
	return dsScriptNode
}

// BuildCACertNode : Build a new node to store CA cert, this would be referred by the corresponding keycert
func (o *AviObjectGraph) BuildCACertNode(tlsNode *AviVsNode, cacert, keycertname, key string) string {
	cacertNode := &AviTLSKeyCertNode{Name: lib.GetCACertNodeName(keycertname), Tenant: lib.GetTenant()}
	cacertNode.Type = lib.CertTypeCA
	cacertNode.Cert = []byte(cacert)

	if tlsNode.CheckCACertNodeNameNChecksum(cacertNode.Name, cacertNode.GetCheckSum()) {
		if len(tlsNode.CACertRefs) == 1 {
			tlsNode.CACertRefs[0] = cacertNode
			utils.AviLog.Warnf("key: %s, msg: duplicate cacerts detected for %s, overwriting", key, cacertNode.Name)
		} else {
			tlsNode.ReplaceCACertRefInSNINode(cacertNode, key)
		}
	}
	return cacertNode.Name
}

func (o *AviObjectGraph) BuildTlsCertNode(svcLister *objects.SvcLister, tlsNode *AviVsNode, namespace string, tlsData TlsSettings, key string, sniHost ...string) bool {
	mClient := utils.GetInformers().ClientSet
	secretName := tlsData.SecretName
	secretNS := tlsData.SecretNS
	if secretNS == "" {
		secretNS = namespace
	}

	var certNode *AviTLSKeyCertNode
	if len(sniHost) > 0 {
		certNode = &AviTLSKeyCertNode{Name: lib.GetTLSKeyCertNodeName(namespace, secretName, sniHost[0]), Tenant: lib.GetTenant()}
	} else {
		certNode = &AviTLSKeyCertNode{Name: lib.GetTLSKeyCertNodeName(namespace, secretName), Tenant: lib.GetTenant()}
	}
	certNode.Type = lib.CertTypeVS

	// Openshift Routes do not refer to a secret, instead key/cert values are mentioned in the route
	if strings.HasPrefix(secretName, lib.RouteSecretsPrefix) {
		if tlsData.cert != "" && tlsData.key != "" {
			certNode.Cert = []byte(tlsData.cert)
			certNode.Key = []byte(tlsData.key)
			if tlsData.cacert != "" {
				certNode.CACert = o.BuildCACertNode(tlsNode, tlsData.cacert, certNode.Name, key)
			} else {
				tlsNode.DeleteCACertRefInSNINode(lib.GetCACertNodeName(certNode.Name), key)
			}
		} else {
			ok, _ := svcLister.IngressMappings(namespace).GetSecretToIng(secretName)
			if ok {
				svcLister.IngressMappings(namespace).DeleteSecretToIngMapping(secretName)
			}
			utils.AviLog.Infof("key: %s, msg: no cert/key specified for TLS route")
			//To Do: use a Default secret if required
			return false
		}
	} else {
		secretObj, err := mClient.CoreV1().Secrets(secretNS).Get(context.TODO(), secretName, metav1.GetOptions{})
		if err != nil || secretObj == nil {
			// This secret has been deleted.
			ok, ingNames := svcLister.IngressMappings(namespace).GetSecretToIng(secretName)
			if ok {
				// Delete the secret key in the cache if it has no references
				if len(ingNames) == 0 {
					svcLister.IngressMappings(namespace).DeleteSecretToIngMapping(secretName)
				}
			}
			utils.AviLog.Infof("key: %s, msg: secret: %s has been deleted, err: %s", key, secretName, err)
			return false
		}
		keycertMap := secretObj.Data
		cert, ok := keycertMap[tlsCert]
		if ok {
			certNode.Cert = cert
		} else {
			utils.AviLog.Infof("key: %s, msg: certificate not found for secret: %s", key, secretObj.Name)
			return false
		}
		tlsKey, keyfound := keycertMap[utils.K8S_TLS_SECRET_KEY]
		if keyfound {
			certNode.Key = tlsKey
		} else {
			utils.AviLog.Infof("key: %s, msg: key not found for secret: %s", key, secretObj.Name)
			return false
		}
		utils.AviLog.Infof("key: %s, msg: Added the secret object to tlsnode: %s", key, secretObj.Name)
	}
	// If this SSLCertRef is already present don't add it.
	if len(sniHost) > 0 {
		if tlsNode.CheckSSLCertNodeNameNChecksum(lib.GetTLSKeyCertNodeName(namespace, secretName, sniHost[0]), certNode.GetCheckSum()) {
			if len(tlsNode.SSLKeyCertRefs) == 1 {
				// Overwrite if the secrets are different.
				tlsNode.SSLKeyCertRefs[0] = certNode
				utils.AviLog.Warnf("key: %s, msg: Duplicate secrets detected for the same hostname, overwrote the secret for hostname %s, with contents of secret :%s in ns: %s", key, sniHost[0], secretName, namespace)
			} else {
				tlsNode.ReplaceSniSSLRefInSNINode(certNode, key)
			}
		}
	} else {
		tlsNode.SSLKeyCertRefs = append(tlsNode.SSLKeyCertRefs, certNode)
	}
	return true
}

func (o *AviObjectGraph) BuildPolicyPGPoolsForSNI(vsNode []*AviVsNode, tlsNode *AviVsNode, namespace string, ingName string, hostpath TlsSettings, secretName string, key string, isIngr bool, hostName ...string) {
	localPGList := make(map[string]*AviPoolGroupNode)
	for host, paths := range hostpath.Hosts {
		if len(hostName) > 0 {
			if hostName[0] != host {
				// If a hostname is passed to this method, ensure we only process that hostname and nothing else.
				continue
			}
		}
		// Update the VSVIP with the host information.
		if !utils.HasElem(vsNode[0].VSVIPRefs[0].FQDNs, host) {
			vsNode[0].VSVIPRefs[0].FQDNs = append(vsNode[0].VSVIPRefs[0].FQDNs, host)
		}
		if !utils.HasElem(tlsNode.VHDomainNames, host) {
			tlsNode.VHDomainNames = append(tlsNode.VHDomainNames, host)
		}
		for _, path := range paths {
			var httpPolicySet []AviHostPathPortPoolPG

			httpPGPath := AviHostPathPortPoolPG{Host: host}

			if path.PathType == networkingv1beta1.PathTypeExact {
				httpPGPath.MatchCriteria = "EQUALS"
			} else {
				// PathTypePrefix and PathTypeImplementationSpecific
				// default behaviour for AKO set be Prefix match on the path
				httpPGPath.MatchCriteria = "BEGINS_WITH"
			}

			if path.Path != "" {
				httpPGPath.Path = append(httpPGPath.Path, path.Path)
			}

			pgName := lib.GetSniPGName(ingName, namespace, host, path.Path)
			var pgNode *AviPoolGroupNode
			// There can be multiple services for the same path in case of alternate backend.
			// In that case, make sure we are creating only one PG per path
			pgNode, pgfound := localPGList[pgName]
			if !pgfound {
				pgNode = &AviPoolGroupNode{Name: pgName, Tenant: lib.GetTenant()}
				localPGList[pgName] = pgNode
				httpPGPath.PoolGroup = pgNode.Name
				httpPGPath.Host = host
				httpPolicySet = append(httpPolicySet, httpPGPath)
			}

			var poolName string
			// Do not use serviceName in SNI Pool Name for ingress for backward compatibility
			if isIngr {
				poolName = lib.GetSniPoolName(ingName, namespace, host, path.Path)
			} else {
				poolName = lib.GetSniPoolName(ingName, namespace, host, path.Path, path.ServiceName)
			}
			hostSlice := []string{host}
			poolNode := &AviPoolNode{
				Name:       poolName,
				PortName:   path.PortName,
				Tenant:     lib.GetTenant(),
				VrfContext: lib.GetVrf(),
				ServiceMetadata: avicache.ServiceMetadataObj{
					IngressName: ingName,
					Namespace:   namespace,
					HostNames:   hostSlice,
				},
			}

			if hostpath.reencrypt == true {
				o.BuildPoolSecurity(poolNode, hostpath, key)
			}

			if !lib.IsNodePortMode() {
				if servers := PopulateServers(poolNode, namespace, path.ServiceName, true, key); servers != nil {
					poolNode.Servers = servers
				}
			} else {
				if servers := PopulateServersForNodePort(poolNode, namespace, path.ServiceName, true, key); servers != nil {
					poolNode.Servers = servers
				}
			}
			pool_ref := fmt.Sprintf("/api/pool?name=%s", poolNode.Name)
			ratio := path.weight
			pgNode.Members = append(pgNode.Members, &avimodels.PoolGroupMember{PoolRef: &pool_ref, Ratio: &ratio})

			if tlsNode.CheckPGNameNChecksum(pgNode.Name, pgNode.GetCheckSum()) {
				tlsNode.ReplaceSniPGInSNINode(pgNode, key)
			}
			if tlsNode.CheckPoolNChecksum(poolNode.Name, poolNode.GetCheckSum()) {
				// Replace the poolNode.
				tlsNode.ReplaceSniPoolInSNINode(poolNode, key)
			}
			o.AddModelNode(poolNode)
			if !pgfound {
				httppolname := lib.GetSniHttpPolName(ingName, namespace, host, path.Path)
				policyNode := &AviHttpPolicySetNode{Name: httppolname, HppMap: httpPolicySet, Tenant: lib.GetTenant()}
				if tlsNode.CheckHttpPolNameNChecksum(httppolname, policyNode.GetCheckSum()) {
					tlsNode.ReplaceSniHTTPRefInSNINode(policyNode, key)
				}
			}
		}
		for _, path := range paths {
			BuildPoolHTTPRule(host, path.Path, ingName, namespace, key, tlsNode, true)
		}
	}
	utils.AviLog.Infof("key: %s, msg: added pools and poolgroups. tlsNodeChecksum for tlsNode :%s is :%v", key, tlsNode.Name, tlsNode.GetCheckSum())

}

func (o *AviObjectGraph) BuildPoolSecurity(poolNode *AviPoolNode, tlsData TlsSettings, key string) {
	poolNode.SniEnabled = true
	poolNode.SslProfileRef = fmt.Sprintf("/api/sslprofile?name=%s", lib.DefaultPoolSSLProfile)

	utils.AviLog.Infof("key: %s, Added ssl profile for pool %s", poolNode.Name)
	if tlsData.destCA == "" {
		return
	}
	pkiProfile := AviPkiProfileNode{
		Name:   lib.GetPoolPKIProfileName(poolNode.Name),
		Tenant: lib.GetTenant(),
		CACert: tlsData.destCA,
	}
	utils.AviLog.Infof("key: %s, Added pki profile %s for pool %s", pkiProfile.Name, poolNode.Name)
	poolNode.PkiProfile = &pkiProfile
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
		Tenant:        lib.GetTenant(),
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
				utils.AviLog.Infof("key: %s, msg: replaced host %s for policy %s in model", key, hostname, policy.Name)
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
			utils.AviLog.Infof("key: %s, msg: removed host %s from policy %s in model %v", key, hostname, policy.Name, policy.RedirectPorts[0].Hosts)
			if len(policy.RedirectPorts[0].Hosts) == 0 {
				deletePolicy = true
			}

			if deletePolicy {
				vsNode.HttpPolicyRefs = append(vsNode.HttpPolicyRefs[:i], vsNode.HttpPolicyRefs[i+1:]...)
				utils.AviLog.Infof("key: %s, msg: removed policy %s in model", key, policy.Name)
			}
		}
	}
}
