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
	"strings"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	akov1alpha1 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/apis/ako/v1alpha1"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	avimodels "github.com/vmware/alb-sdk/go/models"
	"google.golang.org/protobuf/proto"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/sets"
)

// TODO: Move to utils
const tlsCert = "tls.crt"

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

func (o *AviObjectGraph) ConstructAviL7VsNode(vsName string, key string, routeIgrObj RouteIngressModel, dedicatedVs, secureVS bool) {
	o.Lock.Lock()
	defer o.Lock.Unlock()

	var vrfcontext string
	avi_vs_meta := &AviVsNode{
		Name:               vsName,
		Tenant:             lib.GetTenant(),
		ServiceEngineGroup: lib.GetSEGName(),
		EnableRhi:          proto.Bool(lib.GetEnableRHI()),
		NetworkProfile:     utils.DEFAULT_TCP_NW_PROFILE,
		ApplicationProfile: utils.DEFAULT_L7_APP_PROFILE,
		PortProto: []AviPortHostProtocol{
			{Port: 80, Protocol: utils.HTTP},
		},
	}
	if !dedicatedVs {
		avi_vs_meta.SharedVS = true
		avi_vs_meta.SNIParent = true
	} else {
		avi_vs_meta.Dedicated = true
	}

	//For SNI, by default port 80 and 443 added
	//For dedicated, in secure ingress only port 443 added
	if !dedicatedVs || secureVS {
		httpsPort := AviPortHostProtocol{Port: 443, Protocol: utils.HTTP, EnableSSL: true}
		avi_vs_meta.PortProto = append(avi_vs_meta.PortProto, httpsPort)
	}
	if dedicatedVs && secureVS {
		avi_vs_meta.ApplicationProfile = utils.DEFAULT_L7_SECURE_APP_PROFILE
	}

	// If NSX-T LR path is empty, set vrfContext
	if lib.GetT1LRPath() == "" {
		vrfcontext = lib.GetVrf()
		avi_vs_meta.VrfContext = vrfcontext
	}
	if !dedicatedVs {
		o.ConstructShardVsPGNode(vsName, key, avi_vs_meta)
		o.ConstructHTTPDataScript(vsName, key, avi_vs_meta)
	}
	o.AddModelNode(avi_vs_meta)

	subDomains := GetDefaultSubDomain()
	fqdns, fqdn := lib.GetFqdns(vsName, key, subDomains)
	configuredSharedVSFqdn := fqdn

	vsVipNode := &AviVSVIPNode{
		Name:        lib.GetVsVipName(vsName),
		Tenant:      lib.GetTenant(),
		FQDNs:       fqdns,
		VrfContext:  vrfcontext,
		VipNetworks: lib.GetVipNetworkList(),
	}

	if lib.GetT1LRPath() != "" {
		vsVipNode.T1Lr = lib.GetT1LRPath()
	}

	if avi_vs_meta.EnableRhi != nil && *avi_vs_meta.EnableRhi {
		vsVipNode.BGPPeerLabels = lib.GetGlobalBgpPeerLabels()
	}

	if infraSetting := routeIgrObj.GetAviInfraSetting(); infraSetting != nil {
		buildWithInfraSetting(key, avi_vs_meta, vsVipNode, infraSetting)
	}

	avi_vs_meta.VSVIPRefs = append(avi_vs_meta.VSVIPRefs, vsVipNode)
	//Apply hostrule on shared Vs fqdn
	if avi_vs_meta.SharedVS && configuredSharedVSFqdn != "" {
		BuildL7HostRule(configuredSharedVSFqdn, key, avi_vs_meta)
	}
}

func (o *AviObjectGraph) ConstructShardVsPGNode(vsName string, key string, vsNode *AviVsNode) *AviPoolGroupNode {
	pgName := lib.GetL7SharedPGName(vsName)
	pgNode := &AviPoolGroupNode{Name: pgName, Tenant: lib.GetTenant(), ImplicitPriorityLabel: true}
	pgNode.AttachedToSharedVS = vsNode.SharedVS
	vsNode.PoolGroupRefs = append(vsNode.PoolGroupRefs, pgNode)
	o.AddModelNode(pgNode)
	return pgNode
}

func (o *AviObjectGraph) ConstructHTTPDataScript(vsName string, key string, vsNode *AviVsNode) *AviHTTPDataScriptNode {
	scriptStr := utils.HTTP_DS_SCRIPT
	if lib.GetEnableCtrl2014Features() {
		scriptStr = utils.HTTP_DS_SCRIPT_MODIFIED
	}
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
func (o *AviObjectGraph) BuildCACertNode(tlsNode *AviVsNode, cacert, infraSettingName, host, key string) string {
	cacertNode := &AviTLSKeyCertNode{Name: lib.GetCACertNodeName(infraSettingName, host), Tenant: lib.GetTenant()}
	cacertNode.Type = lib.CertTypeCA
	cacertNode.Cert = []byte(cacert)
	cacertNode.AviMarkers = lib.PopulateTLSKeyCertNode(host, infraSettingName)
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

func (o *AviObjectGraph) BuildTlsCertNode(svcLister *objects.SvcLister, tlsNode *AviVsNode, namespace string, tlsData TlsSettings, key, infraSettingName, sniHost string) bool {
	secretName := tlsData.SecretName
	secretNS := tlsData.SecretNS
	if secretNS == "" {
		secretNS = namespace
	}

	certNode := &AviTLSKeyCertNode{
		Name:   lib.GetTLSKeyCertNodeName(infraSettingName, sniHost),
		Tenant: lib.GetTenant(),
		Type:   lib.CertTypeVS,
	}
	certNode.AviMarkers = lib.PopulateTLSKeyCertNode(sniHost, infraSettingName)
	// Openshift Routes do not refer to a secret, instead key/cert values are mentioned in the route
	if strings.HasPrefix(secretName, lib.RouteSecretsPrefix) {
		if tlsData.cert != "" && tlsData.key != "" {
			certNode.Cert = []byte(tlsData.cert)
			certNode.Key = []byte(tlsData.key)
			if tlsData.cacert != "" {
				certNode.CACert = o.BuildCACertNode(tlsNode, tlsData.cacert, infraSettingName, sniHost, key)
			} else {
				tlsNode.DeleteCACertRefInSNINode(lib.GetCACertNodeName(infraSettingName, sniHost), key)
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
		secretObj, err := utils.GetInformers().SecretInformer.Lister().Secrets(secretNS).Get(secretName)
		if err != nil || secretObj == nil {
			// This secret has been deleted.
			ok, ingNames := svcLister.IngressMappings(namespace).GetSecretToIng(secretName)
			if ok {
				// Delete the secret key in the cache if it has no references
				if len(ingNames) == 0 {
					svcLister.IngressMappings(namespace).DeleteSecretToIngMapping(secretName)
				}
			}
			utils.AviLog.Warnf("key: %s, msg: secret: %s has been deleted, err: %s", key, secretName, err)
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
	if tlsNode.CheckSSLCertNodeNameNChecksum(lib.GetTLSKeyCertNodeName(infraSettingName, sniHost), certNode.GetCheckSum()) {
		if len(tlsNode.SSLKeyCertRefs) == 1 {
			// Overwrite if the secrets are different.
			tlsNode.SSLKeyCertRefs[0] = certNode
			utils.AviLog.Warnf("key: %s, msg: Duplicate secrets detected for the same hostname, overwrote the secret for hostname %s, with contents of secret :%s in ns: %s", key, sniHost, secretName, namespace)
		} else {
			tlsNode.ReplaceSniSSLRefInSNINode(certNode, key)
		}
	}
	return true
}

func (o *AviObjectGraph) BuildPolicyPGPoolsForSNI(vsNode []*AviVsNode, tlsNode *AviVsNode, namespace string, ingName string, hostpath TlsSettings, secretName string, key string, isIngr bool, infraSetting *akov1alpha1.AviInfraSetting, hostName string) {
	localPGList := make(map[string]*AviPoolGroupNode)
	var sniFQDNs []string
	var priorityLabel string
	var policyNode *AviHttpPolicySetNode
	pathSet := sets.NewString(tlsNode.Paths...)

	var infraSettingName string
	if infraSetting != nil {
		infraSettingName = infraSetting.Name
	}

	ingressNameSet := sets.NewString(tlsNode.IngressNames...)
	ingressNameSet.Insert(ingName)
	for host, paths := range hostpath.Hosts {
		var pathFQDNs []string
		pathFQDNs = append(pathFQDNs, host)
		if hostName != host {
			// Ensure that we only process provided hostname and nothing else.
			continue
		}

		// Update the VSVIP with the host information.
		if !utils.HasElem(vsNode[0].VSVIPRefs[0].FQDNs, host) {
			vsNode[0].VSVIPRefs[0].FQDNs = append(vsNode[0].VSVIPRefs[0].FQDNs, host)
		}
		if paths.gslbHostHeader != "" {
			// check if the VHDomain is already added, if not add it.
			if !utils.HasElem(pathFQDNs, paths.gslbHostHeader) {
				pathFQDNs = append(pathFQDNs, paths.gslbHostHeader)
			}

			if vsNode[0].Dedicated && !utils.HasElem(vsNode[0].VSVIPRefs[0].FQDNs, paths.gslbHostHeader) {
				vsNode[0].VSVIPRefs[0].FQDNs = append(vsNode[0].VSVIPRefs[0].FQDNs, paths.gslbHostHeader)
			}

		}
		httpPolName := lib.GetSniHttpPolName(namespace, host, infraSettingName)
		for i, http := range tlsNode.HttpPolicyRefs {
			if http.Name == httpPolName {
				policyNode = tlsNode.HttpPolicyRefs[i]
			}
		}
		if policyNode == nil {
			policyNode = &AviHttpPolicySetNode{Name: httpPolName, Tenant: lib.GetTenant()}
			tlsNode.HttpPolicyRefs = append(tlsNode.HttpPolicyRefs, policyNode)
		}

		for _, path := range paths.ingressHPSvc {

			httpPGPath := AviHostPathPortPoolPG{Host: pathFQDNs}

			if path.PathType == networkingv1.PathTypeExact {
				httpPGPath.MatchCriteria = "EQUALS"
			} else {
				// PathTypePrefix and PathTypeImplementationSpecific
				// default behaviour for AKO set be Prefix match on the path
				httpPGPath.MatchCriteria = "BEGINS_WITH"
			}

			if path.Path != "" {
				httpPGPath.Path = append(httpPGPath.Path, path.Path)
				priorityLabel = host + path.Path
			} else {
				priorityLabel = host
			}
			var poolName string
			var pgfound bool
			var pgNode *AviPoolGroupNode
			// Do not use serviceName in SNI Pool Name for ingress for backward compatibility
			if isIngr {
				poolName = lib.GetSniPoolName(ingName, namespace, host, path.Path, infraSettingName, vsNode[0].Dedicated)
			} else {
				poolName = lib.GetSniPoolName(ingName, namespace, host, path.Path, infraSettingName, vsNode[0].Dedicated, path.ServiceName)
			}
			httpPGPath.Host = pathFQDNs
			// There can be multiple services for the same path in case of alternate backend.
			// In that case, make sure we are creating only one PG per path
			if lib.GetNoPGForSNI() && isIngr {
				// If this flag is switched on at a time when the pool is referred by a PG, then the httppolicyset cannot refer to the same pool unless the pool is detached from the poolgroup
				// first, and that is going to mess up the ordering. Hence creating a pool with a different name here. The previous pool will become stale in the process and will get deleted.
				// An AKO reboot would be required to clean up any stale pools if left behind.
				poolName = poolName + "--" + lib.PoolNameSuffixForHttpPolToPool
				httpPGPath.Pool = poolName
				utils.AviLog.Infof("key: %s, msg: using pool name: %s instead of poolgroups for http policy set", key, poolName)
			} else {
				pgName := lib.GetSniPGName(ingName, namespace, host, path.Path, infraSettingName, vsNode[0].Dedicated)
				pgNode, pgfound = localPGList[pgName]
				if !pgfound {
					pgNode = &AviPoolGroupNode{Name: pgName, Tenant: lib.GetTenant()}
				}
				localPGList[pgName] = pgNode
				httpPGPath.PoolGroup = pgNode.Name
				pgNode.AviMarkers = lib.PopulatePGNodeMarkers(namespace, host, infraSettingName, []string{ingName}, []string{path.Path})
			}

			hostSlice := []string{host}
			poolNode := &AviPoolNode{
				Name:          poolName,
				IngressName:   ingName,
				PortName:      path.PortName,
				Tenant:        lib.GetTenant(),
				PriorityLabel: priorityLabel,
				Port:          path.Port,
				TargetPort:    intstr.FromInt(int(path.TargetPort)),
				ServiceMetadata: lib.ServiceMetadataObj{
					IngressName: ingName,
					Namespace:   namespace,
					HostNames:   hostSlice,
					PoolRatio:   path.weight,
				},
				VrfContext: lib.GetVrf(),
			}

			poolNode.NetworkPlacementSettings, _ = lib.GetNodeNetworkMap()

			if lib.GetT1LRPath() != "" {
				poolNode.T1Lr = lib.GetT1LRPath()
				// Unset the poolnode's vrfcontext.
				poolNode.VrfContext = ""
			}

			poolNode.AviMarkers = lib.PopulatePoolNodeMarkers(namespace, host, infraSettingName,
				path.ServiceName, []string{ingName}, []string{path.Path})
			if hostpath.reencrypt {
				o.BuildPoolSecurity(poolNode, hostpath, key, poolNode.AviMarkers)
			}

			serviceType := lib.GetServiceType()
			if serviceType == lib.NodePortLocal {
				if servers := PopulateServersForNPL(poolNode, namespace, path.ServiceName, true, key); servers != nil {
					poolNode.Servers = servers
				}
			} else if serviceType == lib.NodePort {
				if servers := PopulateServersForNodePort(poolNode, namespace, path.ServiceName, true, key); servers != nil {
					poolNode.Servers = servers
				}
			} else {
				if servers := PopulateServers(poolNode, namespace, path.ServiceName, true, key); servers != nil {
					poolNode.Servers = servers
				}
			}

			buildPoolWithInfraSetting(key, poolNode, infraSetting)

			if !lib.GetNoPGForSNI() || !isIngr {
				pool_ref := fmt.Sprintf("/api/pool?name=%s", poolNode.Name)
				ratio := path.weight
				pgNode.Members = append(pgNode.Members, &avimodels.PoolGroupMember{PoolRef: &pool_ref, Ratio: &ratio})

				if tlsNode.CheckPGNameNChecksum(pgNode.Name, pgNode.GetCheckSum()) {
					tlsNode.ReplaceSniPGInSNINode(pgNode, key)
				}
			}

			if tlsNode.CheckPoolNChecksum(poolNode.Name, poolNode.GetCheckSum()) {
				// Replace the poolNode.
				tlsNode.ReplaceSniPoolInSNINode(poolNode, key)
			}
			if !pgfound {
				pathSet.Insert(path.Path)
				hppMapName := lib.GetSniHppMapName(ingName, namespace, host, path.Path, infraSettingName, vsNode[0].Dedicated)
				httpPGPath.Name = hppMapName
				httpPGPath.IngName = ingName
				policyNode.AviMarkers = lib.PopulateHTTPPolicysetNodeMarkers(namespace, host, infraSettingName, ingressNameSet.List(), pathSet.List())
				httpPGPath.CalculateCheckSum()

				if tlsNode.CheckHttpPolNameNChecksum(httpPolName, hppMapName, httpPGPath.Checksum) {
					tlsNode.ReplaceSniHTTPRefInSNINode(httpPGPath, httpPolName, key)
				}
			}
			BuildPoolHTTPRule(host, path.Path, ingName, namespace, infraSettingName, key, tlsNode, true, vsNode[0].Dedicated)
		}
		sniFQDNs = append(sniFQDNs, pathFQDNs...)
	}
	tlsNode.Paths = pathSet.List()
	tlsNode.IngressNames = ingressNameSet.List()

	// Whatever is there in sniFQDNs should be in the VHDomain
	tlsNode.VHDomainNames = sniFQDNs
	utils.AviLog.Infof("key: %s, msg: added pools and poolgroups. tlsNodeChecksum for tlsNode :%s is :%v", key, tlsNode.Name, tlsNode.GetCheckSum())
}

func (o *AviObjectGraph) BuildPoolSecurity(poolNode *AviPoolNode, tlsData TlsSettings, key string, aviMarkers utils.AviObjectMarkers) {
	poolNode.SniEnabled = true
	poolNode.SslProfileRef = fmt.Sprintf("/api/sslprofile?name=%s", lib.DefaultPoolSSLProfile)

	utils.AviLog.Infof("key: %s, Added ssl profile for pool %s", key, poolNode.Name)
	if tlsData.destCA == "" {
		return
	}
	pkiProfile := AviPkiProfileNode{
		Name:   lib.GetPoolPKIProfileName(poolNode.Name),
		Tenant: lib.GetTenant(),
		CACert: tlsData.destCA,
	}
	pkiProfile.AviMarkers = lib.PopulatePoolNodeMarkers(aviMarkers.Namespace, aviMarkers.Host[0],
		aviMarkers.InfrasettingName, aviMarkers.ServiceName, aviMarkers.IngressName, aviMarkers.Path)
	utils.AviLog.Infof("key: %s, Added pki profile %s for pool %s", pkiProfile.Name, poolNode.Name)
	poolNode.PkiProfile = &pkiProfile
}

func (o *AviObjectGraph) BuildPolicyRedirectForVS(vsNode []*AviVsNode, hostnames []string, namespace, infrasettingName, host, key string) {
	policyname := lib.GetL7HttpRedirPolicy(vsNode[0].Name)
	myHppMap := AviRedirectPort{
		Hosts:        hostnames,
		RedirectPort: 443,
		StatusCode:   lib.STATUS_REDIRECT,
		VsPort:       80,
	}

	redirectPolicy := &AviHttpPolicySetNode{
		Tenant:        lib.GetTenant(),
		Name:          policyname,
		RedirectPorts: []AviRedirectPort{myHppMap},
	}
	redirectPolicy.AttachedToSharedVS = vsNode[0].SharedVS
	if vsNode[0].Dedicated {
		//path and ingressname will be empty for redirect policy
		redirectPolicy.AviMarkers = lib.PopulateHTTPPolicysetNodeMarkers(namespace, host, infrasettingName, []string{}, []string{})
	}
	if policyFound := FindAndReplaceRedirectHTTPPolicyInModel(vsNode[0], redirectPolicy, hostnames, key); !policyFound {
		redirectPolicy.CalculateCheckSum()
		vsNode[0].HttpPolicyRefs = append(vsNode[0].HttpPolicyRefs, redirectPolicy)
	}

}

func (o *AviObjectGraph) BuildHeaderRewrite(vsNode []*AviVsNode, gslbHost, localHost, key string) {
	policyname := lib.GetHeaderRewritePolicy(vsNode[0].Name, localHost)
	rewriteRule := AviHostHeaderRewrite{
		SourceHost: gslbHost,
		TargetHost: localHost,
	}

	rewritePolicy := &AviHttpPolicySetNode{
		Tenant:        lib.GetTenant(),
		Name:          policyname,
		HeaderReWrite: &rewriteRule,
	}
	rewritePolicy.AttachedToSharedVS = vsNode[0].SharedVS
	if policyFound := FindAndReplaceHeaderRewriteHTTPPolicyInModel(vsNode[0], rewritePolicy, gslbHost, key); !policyFound && gslbHost != "" {
		rewritePolicy.CalculateCheckSum()
		vsNode[0].HttpPolicyRefs = append(vsNode[0].HttpPolicyRefs, rewritePolicy)
	}

}

func FindAndReplaceRedirectHTTPPolicyInModel(vsNode *AviVsNode, httpPolicy *AviHttpPolicySetNode, hostnames []string, key string) bool {
	// The hostnames slice can at max have 2 elements.
	var policyFound bool
	for _, hostname := range hostnames {
		for _, policy := range vsNode.HttpPolicyRefs {
			if policy.Name == httpPolicy.Name {
				if !utils.HasElem(policy.RedirectPorts[0].Hosts, hostname) {
					policy.RedirectPorts[0].Hosts = append(policy.RedirectPorts[0].Hosts, hostname)
					utils.AviLog.Debugf("key: %s, msg: replaced host %s for policy %s in model", key, hostname, policy.Name)
				}
				policyFound = true
			}
		}
	}
	return policyFound
}

func FindAndReplaceHeaderRewriteHTTPPolicyInModel(vsNode *AviVsNode, httpPolicy *AviHttpPolicySetNode, gslbHost, key string) bool {
	for i, policy := range vsNode.HttpPolicyRefs {
		if policy.Name == httpPolicy.Name {
			if gslbHost == "" {
				// This means the policy should be deleted.
				vsNode.HttpPolicyRefs = append(vsNode.HttpPolicyRefs[:i], vsNode.HttpPolicyRefs[i+1:]...)
				return true
			}
			vsNode.HttpPolicyRefs = append(vsNode.HttpPolicyRefs[:i], vsNode.HttpPolicyRefs[i+1:]...)
			vsNode.HttpPolicyRefs = append(vsNode.HttpPolicyRefs, httpPolicy)
			return true
		}
	}
	return false
}

func RemoveHeaderRewriteHTTPPolicyInModel(vsNode *AviVsNode, hostname, key string) {
	policyName := lib.GetHeaderRewritePolicy(vsNode.Name, hostname)
	for i, policy := range vsNode.HttpPolicyRefs {
		if policy.Name == policyName {
			// one redirect policy per shard vs
			vsNode.HttpPolicyRefs = append(vsNode.HttpPolicyRefs[:i], vsNode.HttpPolicyRefs[i+1:]...)
			utils.AviLog.Infof("key: %s, msg: removed http header rewrite policy %s in model", key, policy.Name)
			return
		}
	}
}
func DeleteDedicatedVSNode(vsNode *AviVsNode, hostsToRemove []string, key string) {
	vsNode.PoolGroupRefs = []*AviPoolGroupNode{}
	vsNode.PoolRefs = []*AviPoolNode{}
	vsNode.HttpPolicyRefs = []*AviHttpPolicySetNode{}
	RemoveFqdnFromVIP(vsNode, key, hostsToRemove)
	vsNode.DeletSSLRefInDedicatedNode(key)
	utils.AviLog.Infof("key: %s, msg: Deleted Dedicated node vs: %s", key, vsNode.Name)
}

func RemoveRedirectHTTPPolicyInModel(vsNode *AviVsNode, hostnames []string, key string) {
	policyName := lib.GetL7HttpRedirPolicy(vsNode.Name)
	for _, hostname := range hostnames {
		for i, policy := range vsNode.HttpPolicyRefs {
			if policy.Name == policyName {
				// one redirect policy per shard vs
				if utils.HasElem(policy.RedirectPorts[0].Hosts, hostname) {
					policy.RedirectPorts[0].Hosts = utils.Remove(policy.RedirectPorts[0].Hosts, hostname)
					utils.AviLog.Debugf("key: %s, msg: removed host %s from policy %s in model", key, hostname, policy.Name)
				}
				if len(policy.RedirectPorts[0].Hosts) == 0 {
					vsNode.HttpPolicyRefs = append(vsNode.HttpPolicyRefs[:i], vsNode.HttpPolicyRefs[i+1:]...)
					utils.AviLog.Infof("key: %s, msg: removed redirect policy %s in model", key, policy.Name)
				}
			}
		}
	}
}
func RemoveFqdnFromVIP(vsNode *AviVsNode, key string, Fqdns []string) {
	if len(vsNode.VSVIPRefs) > 0 {
		for _, fqdn := range Fqdns {
			for i, vipFqdn := range vsNode.VSVIPRefs[0].FQDNs {
				if fqdn == vipFqdn {
					utils.AviLog.Debugf("key: %s, msg: Removed FQDN %s from vs node %s", key, fqdn, vsNode.Name)
					vsNode.VSVIPRefs[0].FQDNs = append(vsNode.VSVIPRefs[0].FQDNs[:i], vsNode.VSVIPRefs[0].FQDNs[i+1:]...)
				}
			}
		}
	}
}
func buildWithInfraSetting(key string, vs *AviVsNode, vsvip *AviVSVIPNode, infraSetting *akov1alpha1.AviInfraSetting) {
	if infraSetting != nil && infraSetting.Status.Status == lib.StatusAccepted {
		if infraSetting.Spec.SeGroup.Name != "" {
			// This assumes that the SeGroup has the appropriate labels configured
			vs.ServiceEngineGroup = infraSetting.Spec.SeGroup.Name
		} else {
			vs.ServiceEngineGroup = lib.GetSEGName()
		}

		if infraSetting.Spec.Network.EnableRhi != nil {
			vs.EnableRhi = infraSetting.Spec.Network.EnableRhi
		} else {
			enableRhi := lib.GetEnableRHI()
			vs.EnableRhi = &enableRhi
		}

		if vs.EnableRhi != nil && *vs.EnableRhi {
			if infraSetting.Spec.Network.BgpPeerLabels != nil {
				vsvip.BGPPeerLabels = infraSetting.Spec.Network.BgpPeerLabels
			} else {
				vsvip.BGPPeerLabels = lib.GetGlobalBgpPeerLabels()
			}
		} else {
			vsvip.BGPPeerLabels = nil
		}

		if infraSetting.Spec.Network.VipNetworks != nil && len(infraSetting.Spec.Network.VipNetworks) > 0 {
			vsvip.VipNetworks = infraSetting.Spec.Network.VipNetworks
		} else {
			vsvip.VipNetworks = lib.GetVipNetworkList()
		}
		if lib.IsPublicCloud() {
			vsvip.EnablePublicIP = infraSetting.Spec.Network.EnablePublicIP
		}
		utils.AviLog.Debugf("key: %s, msg: Applied AviInfraSetting configuration over VSNode %s", key, vs.Name)
	}
}

func buildPoolWithInfraSetting(key string, pool *AviPoolNode, infraSetting *akov1alpha1.AviInfraSetting) {
	if infraSetting != nil && infraSetting.Status.Status == lib.StatusAccepted {
		if infraSetting.Spec.Network.NodeNetworks != nil && len(infraSetting.Spec.Network.NodeNetworks) > 0 {
			nodeNetworkMap := make(map[string][]string)
			for _, nodeNetwork := range infraSetting.Spec.Network.NodeNetworks {
				nodeNetworkMap[nodeNetwork.NetworkName] = nodeNetwork.Cidrs
			}
			pool.NetworkPlacementSettings = nodeNetworkMap
		} else {
			pool.NetworkPlacementSettings, _ = lib.GetNodeNetworkMap()
		}

		utils.AviLog.Debugf("key: %s, msg: Applied AviInfraSetting configuration over PoolNode %s", key, pool.Name)
	}
}
