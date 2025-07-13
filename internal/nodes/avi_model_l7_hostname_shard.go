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

	akov1beta1 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/apis/ako/v1beta1"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	avimodels "github.com/vmware/alb-sdk/go/models"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/util/sets"
)

func (o *AviObjectGraph) BuildDedicatedL7VSGraphHostNameShard(vsName, hostname string, routeIgrObj RouteIngressModel, insecureEdgeTermAllow bool, pathsvcMap HostMetadata, key string) {
	o.Lock.Lock()
	defer o.Lock.Unlock()
	var pathFQDNs []string

	namespace := routeIgrObj.GetNamespace()
	ingName := routeIgrObj.GetName()
	utils.AviLog.Infof("key: %s, msg: Building the L7 pools for namespace: %s, hostname: %s", key, namespace, hostname)
	vsNode := o.GetAviVS()

	if len(vsNode) != 1 {
		utils.AviLog.Warnf("key: %s, msg: more than one vs in model.", key)
		return
	}
	var infraSettingName string
	infraSetting := routeIgrObj.GetAviInfraSetting()
	if infraSetting != nil && !lib.IsInfraSettingNSScoped(infraSetting.Name, namespace) {
		infraSettingName = infraSetting.Name
	}
	pathFQDNs = append(pathFQDNs, hostname)

	// Populate the hostmap with empty secret for insecure ingress
	PopulateIngHostMap(namespace, hostname, ingName, "", pathsvcMap)
	_, ingressHostMap := SharedHostNameLister().Get(hostname)

	vsNode[0].ServiceMetadata.NamespaceIngressName = ingressHostMap.GetIngressesForHostName()
	vsNode[0].ServiceMetadata.Namespace = namespace
	vsNode[0].ServiceMetadata.HostNames = pathFQDNs
	vsNode[0].AviMarkers = lib.PopulateVSNodeMarkers(namespace, hostname, infraSettingName)

	// Update the VSVIP with the host information.
	if !utils.HasElem(vsNode[0].VSVIPRefs[0].FQDNs, hostname) {
		vsNode[0].VSVIPRefs[0].FQDNs = append(vsNode[0].VSVIPRefs[0].FQDNs, hostname)
	}

	found, gsFqdnCache := objects.SharedCRDLister().GetLocalFqdnToGSFQDNMapping(hostname)
	gslbHostHeader := pathsvcMap.gslbHostHeader
	if gslbHostHeader != "" {
		utils.AviLog.Debugf("key: %s, msg: GSLB host header: %v", key, gslbHostHeader)
		if gsFqdnCache != gslbHostHeader {
			RemoveFqdnFromVIP(vsNode[0], key, []string{gsFqdnCache})
		}
		// check if the Domain is already added, if not add it.
		if !utils.HasElem(pathFQDNs, gslbHostHeader) {
			pathFQDNs = append(pathFQDNs, gslbHostHeader)
		}
		objects.SharedCRDLister().UpdateLocalFQDNToGSFqdnMapping(hostname, gslbHostHeader)
	} else {
		if found {
			objects.SharedCRDLister().DeleteLocalFqdnToGsFqdnMap(hostname)
			RemoveFqdnFromVIP(vsNode[0], key, []string{gsFqdnCache})
		}
	}

	RemoveRedirectHTTPPolicyInModel(vsNode[0], pathFQDNs, key)
	vsNode[0].DeletSSLRefInDedicatedNode(key)
	vsNode[0].DeleteSSLPort(key)
	vsNode[0].Secure = false
	vsNode[0].DeleteSecureAppProfile(key)
	objType := routeIgrObj.GetType()
	isIngr := objType == utils.Ingress
	pathsvc := pathsvcMap.ingressHPSvc

	o.BuildPoolPGPolicyForDedicatedVS(vsNode, namespace, ingName, hostname, infraSetting, key, pathFQDNs, pathsvc, insecureEdgeTermAllow, isIngr)
	BuildL7HostRule(hostname, key, vsNode[0])
	// Compare and remove the deleted aliases from the FQDN list
	var hostsToRemove []string
	_, oldFQDNAliases := objects.SharedCRDLister().GetFQDNToAliasesMapping(hostname)
	for _, host := range oldFQDNAliases {
		if !utils.HasElem(vsNode[0].VHDomainNames, host) {
			hostsToRemove = append(hostsToRemove, host)
		}
	}
	vsNode[0].RemoveFQDNsFromModel(hostsToRemove, key)

	// Add FQDN aliases in the hostrule CRD to parent and child VSes
	vsNode[0].AddFQDNsToModel(vsNode[0].VHDomainNames, gslbHostHeader, key)
	vsNode[0].AddFQDNAliasesToHTTPPolicy(vsNode[0].VHDomainNames, key)
	vsNode[0].AviMarkers.Host = vsNode[0].VHDomainNames
	objects.SharedCRDLister().UpdateFQDNToAliasesMappings(hostname, vsNode[0].VHDomainNames)
}

func (o *AviObjectGraph) BuildPoolPGPolicyForDedicatedVS(vsNode []*AviVsNode, namespace, ingName, hostname string, infraSetting *akov1beta1.AviInfraSetting, key string, pathFQDNs []string, paths []IngressHostPathSvc, insecureEdgeTermAllow, isIngr bool) {
	localPGList := make(map[string]*AviPoolGroupNode)
	var policyNode *AviHttpPolicySetNode
	var pgfound bool
	var priorityLabel string
	var poolName string

	pathSet := sets.NewString(vsNode[0].Paths...)
	ingressNameSet := sets.NewString(vsNode[0].IngressNames...)
	ingressNameSet.Insert(ingName)

	var infraSettingName string
	if infraSetting != nil && !lib.IsInfraSettingNSScoped(infraSetting.Name, namespace) {
		infraSettingName = infraSetting.Name
	}

	httpPolName := lib.GetSniHttpPolName(namespace, hostname, infraSettingName)
	isHttpPolNameLengthExceedAviLimit := false
	if lib.CheckObjectNameLength(httpPolName, lib.HTTPPS) {
		isHttpPolNameLengthExceedAviLimit = true
	}
	for i, http := range vsNode[0].HttpPolicyRefs {
		if http.Name == httpPolName {
			if isHttpPolNameLengthExceedAviLimit {
				// replace- this is for existing httppolicyset on upgrade
				vsNode[0].HttpPolicyRefs = append(vsNode[0].HttpPolicyRefs[:i], vsNode[0].HttpPolicyRefs[i+1:]...)
			} else {
				policyNode = vsNode[0].HttpPolicyRefs[i]
			}
		}
	}
	// append only when name length don't exceed
	if policyNode == nil && !isHttpPolNameLengthExceedAviLimit {
		policyNode = &AviHttpPolicySetNode{Name: httpPolName, Tenant: vsNode[0].Tenant}
		vsNode[0].HttpPolicyRefs = append(vsNode[0].HttpPolicyRefs, policyNode)
	}

	utils.AviLog.Infof("key: %s, msg: The pathsvc mapping: %v", key, paths)
	for _, obj := range paths {
		isPoolNameLenExceedAviLimit := false
		isPGNameLenExceedAviLimit := false

		var pgNode *AviPoolGroupNode
		httpPGPath := AviHostPathPortPoolPG{Host: pathFQDNs}

		if obj.PathType == networkingv1.PathTypeExact {
			httpPGPath.MatchCriteria = "EQUALS"
		} else {
			// PathTypePrefix and PathTypeImplementationSpecific
			// default behaviour for AKO set be Prefix match on the path
			httpPGPath.MatchCriteria = "BEGINS_WITH"
		}

		if obj.Path != "" {
			httpPGPath.Path = append(httpPGPath.Path, obj.Path)
			priorityLabel = hostname + obj.Path
		} else {
			priorityLabel = hostname
		}
		if isIngr {
			poolName = lib.GetSniPoolName(ingName, namespace, hostname, obj.Path, infraSettingName, vsNode[0].Dedicated)
		} else {
			poolName = lib.GetSniPoolName(ingName, namespace, hostname, obj.Path, infraSettingName, vsNode[0].Dedicated, obj.ServiceName)
		}

		if lib.GetNoPGForSNI() && isIngr {
			// If this flag is switched on at a time when the pool is referred by a PG, then the httppolicyset cannot refer to the same pool unless the pool is detached from the poolgroup
			// first, and that is going to mess up the ordering. Hence creating a pool with a different name here. The previous pool will become stale in the process and will get deleted.
			// An AKO reboot would be required to clean up any stale pools if left behind.
			poolName = poolName + "--" + lib.PoolNameSuffixForHttpPolToPool
			if lib.CheckObjectNameLength(poolName, lib.Pool) {
				isPoolNameLenExceedAviLimit = true
			}
			if !isPoolNameLenExceedAviLimit {
				// Add only when pool name is < 255
				httpPGPath.Pool = poolName
			}
			utils.AviLog.Infof("key: %s, msg: using pool name: %s instead of poolgroups for http policy set", key, poolName)
		} else {
			pgName := lib.GetSniPGName(ingName, namespace, hostname, obj.Path, infraSettingName, vsNode[0].Dedicated)
			if lib.CheckObjectNameLength(pgName, lib.PG) {
				isPGNameLenExceedAviLimit = true
			}
			pgNode, pgfound = localPGList[pgName]
			if !pgfound {
				pgNode = &AviPoolGroupNode{Name: pgName, Tenant: vsNode[0].Tenant}
			}
			localPGList[pgName] = pgNode
			if !isPGNameLenExceedAviLimit {
				httpPGPath.PoolGroup = pgNode.Name
			}
			pgNode.AviMarkers = lib.PopulatePGNodeMarkers(namespace, hostname, infraSettingName, []string{ingName}, []string{obj.Path})
		}
		var storedHosts []string
		storedHosts = append(storedHosts, hostname)
		poolNode := buildPoolNode(key, poolName, ingName, namespace, priorityLabel, hostname, infraSetting, obj.ServiceName, storedHosts, insecureEdgeTermAllow, obj)
		isPoolNameLenExceedAviLimit = false
		if lib.CheckObjectNameLength(poolNode.Name, lib.Pool) {
			isPoolNameLenExceedAviLimit = true
		}
		if !lib.GetNoPGForSNI() || !isIngr {
			pool_ref := fmt.Sprintf("/api/pool?name=%s", poolNode.Name)

			ratio := obj.weight
			if !isPoolNameLenExceedAviLimit {
				pgNode.Members = append(pgNode.Members, &avimodels.PoolGroupMember{PoolRef: &pool_ref, Ratio: &ratio})
			}
			// if PG name exceeds limit, do not add it to vs node
			if isPGNameLenExceedAviLimit || vsNode[0].CheckPGNameNChecksum(pgNode.Name, pgNode.GetCheckSum()) {
				vsNode[0].ReplaceSniPGInSNINode(pgNode, key, isPGNameLenExceedAviLimit)
			}
		}
		if isPoolNameLenExceedAviLimit || vsNode[0].CheckPoolNChecksum(poolNode.Name, poolNode.GetCheckSum()) {
			// Replace the poolNode.
			vsNode[0].ReplaceSniPoolInSNINode(poolNode, key, isPoolNameLenExceedAviLimit)
		}
		if !pgfound {
			pathSet.Insert(obj.Path)
			isHPPNameLengthExceedAviLimit := false
			hppMapName := lib.GetSniHppMapName(ingName, namespace, hostname, obj.Path, infraSettingName, vsNode[0].Dedicated)
			if lib.CheckObjectNameLength(hppMapName, lib.HTTPPS) {
				isHPPNameLengthExceedAviLimit = true
			}
			httpPGPath.Name = hppMapName
			httpPGPath.IngName = ingName
			if !isHttpPolNameLengthExceedAviLimit {
				policyNode.AviMarkers = lib.PopulateHTTPPolicysetNodeMarkers(namespace, hostname, infraSettingName, ingressNameSet.List(), pathSet.List())
			}
			httpPGPath.CalculateCheckSum()
			if isHPPNameLengthExceedAviLimit || vsNode[0].CheckHttpPolNameNChecksum(httpPolName, hppMapName, httpPGPath.Checksum) {
				vsNode[0].ReplaceSniHTTPRefInSNINode(httpPGPath, httpPolName, key, isHPPNameLengthExceedAviLimit)
			}
		}
		BuildPoolHTTPRule(hostname, obj.Path, ingName, namespace, infraSettingName, key, vsNode[0], true, vsNode[0].Dedicated)
	}
	vsNode[0].Paths = pathSet.List()
	vsNode[0].IngressNames = ingressNameSet.List()
	utils.AviLog.Infof("key: %s, msg: added pools and poolgroups. NodeChecksum for Insecure Dedicated Vs :%s is :%v", key, vsNode[0].Name, vsNode[0].GetCheckSum())
}

func (o *AviObjectGraph) BuildL7VSGraphHostNameShard(vsName, hostname string, routeIgrObj RouteIngressModel, pathsvc []IngressHostPathSvc, gslbHostHeader string, insecureEdgeTermAllow bool, key string) {
	// panic(vsName)
	o.Lock.Lock()
	defer o.Lock.Unlock()
	// We create pools and attach servers to them here. Pools are created with a priorty label of host/path
	namespace := routeIgrObj.GetNamespace()
	ingName := routeIgrObj.GetName()
	utils.AviLog.Infof("key: %s, msg: Building the L7 pools for namespace: %s, hostname: %s", key, namespace, hostname)
	pgName := lib.GetL7SharedPGName(vsName)
	pgNode := o.GetPoolGroupByName(pgName)
	vsNode := o.GetAviVS()
	utils.AviLog.Debugf("key: %s, msg: GSLB host header: %v", key, gslbHostHeader)

	o.BuildHeaderRewrite(vsNode, gslbHostHeader, hostname, key)

	if len(vsNode) != 1 {
		utils.AviLog.Warnf("key: %s, msg: more than one vs in model.", key)
		return
	}
	var priorityLabel string
	var poolName string
	var serviceName string
	var infraSetting *akov1beta1.AviInfraSetting
	var infraSettingName string
	infraSetting = routeIgrObj.GetAviInfraSetting()
	if infraSetting != nil && !lib.IsInfraSettingNSScoped(infraSetting.Name, namespace) {
		infraSettingName = infraSetting.Name
	}

	utils.AviLog.Infof("key: %s, msg: The pathsvc mapping: %v", key, pathsvc)
	for _, obj := range pathsvc {
		if obj.Path != "" {
			priorityLabel = hostname + obj.Path
		} else {
			priorityLabel = hostname
		}

		// Using servicename in poolname for routes, but not in ingress for consistency with existing naming convention.
		// If possible, we would make this uniform
		if routeIgrObj.GetType() == utils.Ingress {
			poolName = lib.GetL7PoolName(priorityLabel, namespace, ingName, infraSettingName)
			serviceName = ""
		} else {
			poolName = lib.GetL7PoolName(priorityLabel, namespace, ingName, infraSettingName, obj.ServiceName)
			serviceName = obj.ServiceName
		}

		// First check if there are pools related to this ingress present in the model already
		poolNodes := o.GetAviPoolNodesByIngress(namespace, ingName)
		utils.AviLog.Debugf("key: %s, msg: found pools in the model: %s", key, utils.Stringify(poolNodes))
		for _, pool := range poolNodes {
			if pool.Name == poolName {
				o.RemovePoolNodeRefs(pool.Name)
				break
			}
		}
		// First retrieve the FQDNs from the cache and update the model
		var storedHosts []string
		storedHosts = append(storedHosts, hostname)
		vsNode[0].RemoveFQDNsFromModel(storedHosts, key)
		if pgNode != nil {
			// Processsing insecure ingress
			if !utils.HasElem(vsNode[0].VSVIPRefs[0].FQDNs, hostname) {
				vsNode[0].VSVIPRefs[0].FQDNs = append(vsNode[0].VSVIPRefs[0].FQDNs, hostname)
				// combine maps of each hostname.
			}
			// Check poolname length, if >255, don't add it.
			if lib.CheckObjectNameLength(poolName, lib.Pool) {
				// as this object will not be created at AviController, continue from here.
				continue
			}
			poolNode := buildPoolNode(key, poolName, ingName, namespace, priorityLabel, hostname, infraSetting, serviceName, storedHosts, insecureEdgeTermAllow, obj)
			vsNode[0].PoolRefs = append(vsNode[0].PoolRefs, poolNode)
			utils.AviLog.Debugf("key: %s, msg: the pools after append are: %v", key, utils.Stringify(vsNode[0].PoolRefs))
		}

	}
	for _, obj := range pathsvc {
		BuildPoolHTTPRule(hostname, obj.Path, ingName, namespace, infraSettingName, key, vsNode[0], false, vsNode[0].Dedicated)
	}

	// Reset the PG Node members and rebuild them
	pgNode.Members = nil
	for _, poolNode := range vsNode[0].PoolRefs {
		ratio := poolNode.ServiceMetadata.PoolRatio
		pool_ref := fmt.Sprintf("/api/pool?name=%s", poolNode.Name)
		pgNode.Members = append(pgNode.Members, &avimodels.PoolGroupMember{PoolRef: &pool_ref, PriorityLabel: &poolNode.PriorityLabel, Ratio: &ratio})

	}
}

func buildPoolNode(key, poolName, ingName, namespace, priorityLabel, hostname string, infraSetting *akov1beta1.AviInfraSetting, serviceName string, storedHosts []string, insecureEdgeTermAllow bool, obj IngressHostPathSvc) *AviPoolNode {
	tenant := lib.GetTenantInNamespace(namespace)
	poolNode := &AviPoolNode{
		Name:          poolName,
		IngressName:   ingName,
		PortName:      obj.PortName,
		Tenant:        tenant,
		PriorityLabel: strings.ToLower(priorityLabel),
		Port:          obj.Port,
		TargetPort:    obj.TargetPort,
		ServiceMetadata: lib.ServiceMetadataObj{
			IngressName:           ingName,
			Namespace:             namespace,
			HostNames:             storedHosts,
			PoolRatio:             obj.weight,
			InsecureEdgeTermAllow: insecureEdgeTermAllow,
		},
		VrfContext: lib.GetVrf(),
	}

	poolNode.NetworkPlacementSettings = lib.GetNodeNetworkMap()

	t1lr := lib.GetT1LRPath()
	if infraSetting != nil && infraSetting.Spec.NSXSettings.T1LR != nil {
		t1lr = *infraSetting.Spec.NSXSettings.T1LR
	}
	if t1lr != "" {
		poolNode.T1Lr = t1lr
		// Unset the poolnode's vrfcontext.
		poolNode.VrfContext = ""
	}

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

	var infraSettingName string
	if infraSetting != nil && !lib.IsInfraSettingNSScoped(infraSetting.Name, namespace) {
		infraSettingName = infraSetting.Name
	}
	poolNode.AviMarkers = lib.PopulatePoolNodeMarkers(namespace, hostname, infraSettingName, serviceName, []string{ingName}, []string{obj.Path})

	buildPoolWithInfraSetting(key, poolNode, infraSetting)
	if lib.IsIstioEnabled() {
		poolNode.UpdatePoolNodeForIstio()
	}
	return poolNode
}

func (o *AviObjectGraph) DeletePoolForHostname(vsName, hostname string, routeIgrObj RouteIngressModel, pathSvc map[string][]string, key string, infraSettingName string, removeFqdn, removeRedir, secure bool) bool {
	o.Lock.Lock()
	defer o.Lock.Unlock()

	namespace := routeIgrObj.GetNamespace()
	ingName := routeIgrObj.GetName()
	vsNode := o.GetAviVS()
	var poolName string

	keepSni := false
	if !secure && !vsNode[0].Dedicated {
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
						poolName = lib.GetL7PoolName(priorityLabel, namespace, ingName, infraSettingName)
					} else {
						poolName = lib.GetL7PoolName(priorityLabel, namespace, ingName, infraSettingName, svcName)
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
		if pgNode != nil {
			pgNode.Members = nil
			for _, poolNode := range vsNode[0].PoolRefs {
				ratio := poolNode.ServiceMetadata.PoolRatio
				pool_ref := fmt.Sprintf("/api/pool?name=%s", poolNode.Name)
				pgNode.Members = append(pgNode.Members, &avimodels.PoolGroupMember{PoolRef: &pool_ref, PriorityLabel: &poolNode.PriorityLabel, Ratio: &ratio})
			}
		}
		// Remove the httpredirect policy if any
		if len(vsNode) > 0 {
			RemoveHeaderRewriteHTTPPolicyInModel(vsNode[0], hostname, key)
		}
	} else {
		isIngr := routeIgrObj.GetType() == utils.Ingress
		// SNI VSes donot have secretname in their names
		// Remove the ingress from the hostmap if it secure or dedicated vs
		hostMapOk, ingressHostMap := SharedHostNameLister().Get(hostname)
		if hostMapOk {
			// Replace the ingress map for this host.
			keyToRemove := namespace + "/" + ingName
			delete(ingressHostMap.HostNameMap, keyToRemove)
			SharedHostNameLister().Save(hostname, ingressHostMap)
		}
		sniNodeName := lib.GetSniNodeName(infraSettingName, hostname)
		utils.AviLog.Infof("key: %s, msg: sni node to delete: %s", key, sniNodeName)
		keepSni = o.ManipulateSniNode(sniNodeName, ingName, namespace, hostname, pathSvc, vsNode, key, isIngr, infraSettingName)
	}
	_, FQDNAliases := objects.SharedCRDLister().GetFQDNToAliasesMapping(hostname)
	if removeFqdn && !keepSni {
		var hosts []string
		hosts = append(hosts, hostname)
		hosts = append(hosts, FQDNAliases...)

		// Remove these hosts from the overall FQDN list
		vsNode[0].RemoveFQDNsFromModel(hosts, key)
	}
	if removeRedir && !keepSni {
		var hostnames []string
		found, gsFqdnCache := objects.SharedCRDLister().GetLocalFqdnToGSFQDNMapping(hostname)
		if found {
			hostnames = append(hostnames, gsFqdnCache)
		}
		hostnames = append(hostnames, hostname)
		hostnames = append(hostnames, FQDNAliases...)
		RemoveRedirectHTTPPolicyInModel(vsNode[0], hostnames, key)
	}
	if vsNode[0].Dedicated && !keepSni {
		return true
	}
	return false
}
func (o *AviObjectGraph) manipulateVsNode(vsNode *AviVsNode, ingName, namespace, hostname, infraSettingName string, pathSvc map[string][]string, isIngr bool) {
	for path, services := range pathSvc {
		pgName := lib.GetSniPGName(ingName, namespace, hostname, path, infraSettingName, vsNode.Dedicated)
		pgNode := vsNode.GetPGForVSByName(pgName)
		if pgNode == nil {
			pgNode = vsNode.GetPGForVSByName(lib.GetEncodedSniPGPoolNameforRegex(pgName))
			if pgNode != nil {
				pgName = lib.GetEncodedSniPGPoolNameforRegex(pgName)
			}
		}
		for _, svc := range services {
			var sniPool string
			if isIngr {
				sniPool = lib.GetSniPoolName(ingName, namespace, hostname, path, infraSettingName, vsNode.Dedicated)
			} else {
				sniPool = lib.GetSniPoolName(ingName, namespace, hostname, path, infraSettingName, vsNode.Dedicated, svc)
			}
			// Pls decprecate when PGs have http caching
			if lib.GetNoPGForSNI() && isIngr {
				sniPool = sniPool + "--" + lib.PoolNameSuffixForHttpPolToPool
			}
			if lib.IsNameEncoded(pgName) {
				sniPool = lib.GetEncodedSniPGPoolNameforRegex(sniPool)
			}
			o.RemovePoolNodeRefsFromSni(sniPool, vsNode)
			o.RemovePoolRefsFromPG(sniPool, pgNode)
		}
		// Remove the SNI PG if it has no member
		if pgNode != nil {
			if len(pgNode.Members) == 0 {
				o.RemovePGNodeRefs(pgName, vsNode)
				hppmapname := lib.GetSniHppMapName(ingName, namespace, hostname, path, infraSettingName, vsNode.Dedicated)
				httppolname := lib.GetSniHttpPolName(namespace, hostname, infraSettingName)
				o.RemoveHTTPRefsStringGroupsFromSni(httppolname, hppmapname, vsNode)
			}
		}
		// Keeping this block separate for deprecation later.
		if lib.GetNoPGForSNI() && isIngr {
			hppmapname := lib.GetSniHppMapName(ingName, namespace, hostname, path, infraSettingName, vsNode.Dedicated)
			httppolname := lib.GetSniHttpPolName(namespace, hostname, infraSettingName)
			o.RemoveHTTPRefsStringGroupsFromSni(httppolname, hppmapname, vsNode)
		}
	}
}

func (o *AviObjectGraph) ManipulateSniNode(currentSniNodeName, ingName, namespace, hostname string, pathSvc map[string][]string, vsNode []*AviVsNode, key string, isIngr bool, infraSettingName string) bool {
	if vsNode[0].Dedicated {
		o.manipulateVsNode(vsNode[0], ingName, namespace, hostname, infraSettingName, pathSvc, isIngr)
		if len(vsNode[0].PoolRefs) == 0 {
			// Remove the host mapping
			SharedHostNameLister().Delete(hostname)
			vsNode[0].DeletSSLRefInDedicatedNode(key)
			return false
		}
	} else {
		for _, modelSniNode := range vsNode[0].SniNodes {
			if currentSniNodeName != modelSniNode.Name {
				continue
			}
			o.manipulateVsNode(modelSniNode, ingName, namespace, hostname, infraSettingName, pathSvc, isIngr)
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

func (vsNode *AviVsNode) DeleteSSLPort(key string) {
	for i, port := range vsNode.PortProto {
		if port.Port == lib.SSLPort {
			vsNode.PortProto = append(vsNode.PortProto[:i], vsNode.PortProto[i+1:]...)
		}
	}
}
func (vsNode *AviVsNode) AddSSLPort(key string) {
	for _, port := range vsNode.PortProto {
		if port.Port == lib.SSLPort {
			return
		}
	}
	httpsPort := AviPortHostProtocol{Port: 443, Protocol: utils.HTTP, EnableSSL: true}
	vsNode.PortProto = append(vsNode.PortProto, httpsPort)
}
func (vsNode *AviVsNode) DeletSSLRefInDedicatedNode(key string) {
	vsNode.SSLKeyCertRefs = []*AviTLSKeyCertNode{}
	vsNode.SslProfileRef = nil
	vsNode.CACertRefs = []*AviTLSKeyCertNode{}
}
func (vsNode *AviVsNode) DeleteSecureAppProfile(key string) {
	if vsNode.ApplicationProfile == utils.DEFAULT_L7_SECURE_APP_PROFILE {
		vsNode.ApplicationProfile = utils.DEFAULT_L7_APP_PROFILE
	}
}
func getPaths(pathMapArr []IngressHostPathSvc) []string {
	// Returns a list of paths for a given host
	paths := []string{}
	for _, pathmap := range pathMapArr {
		paths = append(paths, pathmap.Path)
	}
	return paths
}

func sniNodeHostName(routeIgrObj RouteIngressModel, tlssetting TlsSettings, ingName, namespace, key string, fullsync bool, sharedQueue *utils.WorkerQueue, modelList *[]string) (map[string][]IngressHostPathSvc, bool) {
	hostPathSvcMap := make(map[string][]IngressHostPathSvc)
	infraSetting := routeIgrObj.GetAviInfraSetting()
	dedicated := false

	for sniHost, paths := range tlssetting.Hosts {
		var sniHosts []string
		hostPathSvcMap[sniHost] = paths.ingressHPSvc
		PopulateIngHostMap(namespace, sniHost, ingName, tlssetting.SecretName, paths)

		_, ingressHostMap := SharedHostNameLister().Get(sniHost)
		sniHosts = append(sniHosts, sniHost)
		_, shardVsName := DeriveShardVS(sniHost, key, routeIgrObj)
		dedicated = shardVsName.Dedicated
		// For each host, create a SNI node with the secret giving us the key and cert.
		// construct a SNI VS node per tls setting which corresponds to one secret
		model_name := lib.GetModelName(shardVsName.Tenant, shardVsName.Name)
		found, aviModel := objects.SharedAviGraphLister().Get(model_name)
		if !found || aviModel == nil {
			utils.AviLog.Infof("key: %s, msg: model not found, generating new model with name: %s", key, model_name)
			aviModel = NewAviObjectGraph()
			aviModel.(*AviObjectGraph).ConstructAviL7VsNode(shardVsName.Name, shardVsName.Tenant, key, routeIgrObj, shardVsName.Dedicated, true)
		}

		vsNode := aviModel.(*AviObjectGraph).GetAviVS()
		if len(vsNode) < 1 {
			return nil, dedicated
		}

		modelGraph := aviModel.(*AviObjectGraph)
		modelGraph.BuildModelGraphForSNI(routeIgrObj, ingressHostMap, sniHosts, tlssetting, ingName, namespace, infraSetting, sniHost, paths.gslbHostHeader, key)
		if found {
			// if vsNode already exists, check for updates via AviInfraSetting
			if infraSetting != nil {
				buildWithInfraSetting(key, namespace, vsNode[0], vsNode[0].VSVIPRefs[0], infraSetting)
				if vsNode[0].IsSharedVS() {
					for _, sni := range vsNode[0].SniNodes {
						if len(sni.GetVHDomainNames()) > 0 {
							sni.SetPortProtocols(vsNode[0].GetPortProtocols())
							BuildOnlyRegexAppRoot(sni.GetVHDomainNames()[0], key, sni)
						}
					}
				}
			}
		}
		// For dedicated vs we always need to reprocess app-root since we are not building app-root for the portProto that is added as part of same BuildL7HostRule call.
		// This is because for non-dedicated mode, we need the portProto from the parent vs node so it is ready by the time we apply app-root settings to child vs.
		// Where as for dedicated vs, hostrule tcp listener ports will override the existing portProto for the same vs node and later aviinfrasetting may update the listener ports as well.
		if vsNode[0].IsDedicatedVS() {
			BuildOnlyRegexAppRoot(sniHost, key, vsNode[0])
		}
		// Only add this node to the list of models if the checksum has changed.
		modelChanged := saveAviModel(model_name, modelGraph, key)
		if !utils.HasElem(*modelList, model_name) && modelChanged {
			*modelList = append(*modelList, model_name)
		}
	}

	return hostPathSvcMap, dedicated
}

func (o *AviObjectGraph) BuildModelGraphForSNI(routeIgrObj RouteIngressModel, ingressHostMap SecureHostNameMapProp, sniHosts []string, tlssetting TlsSettings, ingName, namespace string, infraSetting *akov1beta1.AviInfraSetting, sniHost, gsFqdn string, key string) {
	o.Lock.Lock()
	defer o.Lock.Unlock()
	var sniNode *AviVsNode
	vsNode := o.GetAviVS()

	certsBuilt := false
	sniSecretName := tlssetting.SecretName
	if lib.IsSecretAviCertRef(sniSecretName) {
		sniSecretName = strings.Split(sniSecretName, "/")[1]
		certsBuilt = true
	}

	var infraSettingName string
	if infraSetting != nil && !lib.IsInfraSettingNSScoped(infraSetting.Name, namespace) {
		infraSettingName = infraSetting.Name
	}

	isDedicated := vsNode[0].Dedicated
	if !isDedicated {
		sniNodeName := lib.GetSniNodeName(infraSettingName, sniHost)
		sniNode = vsNode[0].GetSniNodeForName(sniNodeName)
		if sniNode == nil {
			sniNode = &AviVsNode{
				Name:         sniNodeName,
				VHParentName: vsNode[0].Name,
				Tenant:       vsNode[0].Tenant,
				IsSNIChild:   true,
				ServiceMetadata: lib.ServiceMetadataObj{
					NamespaceIngressName: ingressHostMap.GetIngressesForHostName(),
					Namespace:            namespace,
					HostNames:            sniHosts,
				},
			}
		} else {
			// The SNI node exists, just update the svc metadata
			sniNode.ServiceMetadata.NamespaceIngressName = ingressHostMap.GetIngressesForHostName()
			sniNode.ServiceMetadata.Namespace = namespace
			sniNode.ServiceMetadata.HostNames = sniHosts
		}

		sniNode.ServiceEngineGroup = lib.GetSEGName()
		sniNode.VrfContext = lib.GetVrf()
		sniNode.AviMarkers = lib.PopulateVSNodeMarkers(namespace, sniHost, infraSettingName)
	} else {
		//For dedicated VS
		vsNode[0].ServiceMetadata.NamespaceIngressName = ingressHostMap.GetIngressesForHostName()
		vsNode[0].ServiceMetadata.Namespace = namespace
		vsNode[0].ServiceMetadata.HostNames = sniHosts
		vsNode[0].AddSSLPort(key)
		vsNode[0].Secure = true
		vsNode[0].ApplicationProfile = utils.DEFAULT_L7_SECURE_APP_PROFILE
		vsNode[0].AviMarkers = lib.PopulateVSNodeMarkers(namespace, sniHost, infraSettingName)
	}

	var sniHostToRemove []string
	sniHostToRemove = append(sniHostToRemove, sniHost)
	found, gsFqdnCache := objects.SharedCRDLister().GetLocalFqdnToGSFQDNMapping(sniHost)
	if gsFqdn == "" {
		// If the gslbHostHeader is empty but it is present in the in memory cache, then add it as a candidate for removal and  remove the in memory cache relationship
		if found {
			sniHostToRemove = append(sniHostToRemove, gsFqdnCache)
			objects.SharedCRDLister().DeleteLocalFqdnToGsFqdnMap(sniHost)
			if vsNode[0].Dedicated {
				RemoveFqdnFromVIP(vsNode[0], key, []string{gsFqdnCache})
			}
		}
	} else {
		if gsFqdn != gsFqdnCache {
			sniHostToRemove = append(sniHostToRemove, gsFqdnCache)
			if vsNode[0].Dedicated {
				RemoveFqdnFromVIP(vsNode[0], key, []string{gsFqdnCache})
			}
		}
		objects.SharedCRDLister().UpdateLocalFQDNToGSFqdnMapping(sniHost, gsFqdn)
	}
	if isDedicated {
		sniNode = vsNode[0]
	}
	if !certsBuilt {
		certsBuilt = o.BuildTlsCertNode(routeIgrObj.GetSvcLister(), sniNode, namespace, tlssetting, key, infraSettingName, sniHost)
	}
	if certsBuilt {
		isIngr := routeIgrObj.GetType() == utils.Ingress
		o.BuildPolicyPGPoolsForSNI(vsNode, sniNode, namespace, ingName, tlssetting, sniSecretName, key, isIngr, infraSetting, sniHost)
		if !isDedicated {
			foundSniModel := FindAndReplaceSniInModel(sniNode, vsNode, key)
			if !foundSniModel {
				vsNode[0].SniNodes = append(vsNode[0].SniNodes, sniNode)
			}
		}
		RemoveRedirectHTTPPolicyInModel(vsNode[0], sniHostToRemove, key)
		if !isDedicated {
			RemoveRedirectHTTPPolicyInSniNode(sniNode)
		}
		if tlssetting.redirect {
			if gsFqdn != "" {
				sniHosts = append(sniHosts, gsFqdn)
			}
			o.BuildPolicyRedirectForVS(vsNode, sniHosts, namespace, infraSettingName, sniHost, key)
		}
		// setting child node portProto with same value as parent node so that redirect rules are added for all front-end ports if app-root is set
		sniNode.SetPortProtocols(vsNode[0].PortProto)
		BuildL7HostRule(sniHost, key, sniNode)

		// Compare and remove the deleted aliases from the FQDN list
		var hostsToRemove []string
		_, oldFQDNAliases := objects.SharedCRDLister().GetFQDNToAliasesMapping(sniHost)
		for _, host := range oldFQDNAliases {
			if !utils.HasElem(sniNode.VHDomainNames, host) {
				hostsToRemove = append(hostsToRemove, host)
			}
		}
		vsNode[0].RemoveFQDNsFromModel(hostsToRemove, key)
		vsNode[0].RemoveFQDNAliasesFromHTTPPolicy(hostsToRemove, key)
		sniNode.RemoveFQDNAliasesFromHTTPPolicy(hostsToRemove, key)

		// Add FQDN aliases in the hostrule CRD to parent and child VSes
		vsNode[0].AddFQDNsToModel(sniNode.VHDomainNames, gsFqdn, key)
		vsNode[0].AddFQDNAliasesToHTTPPolicy(sniNode.VHDomainNames, key)
		sniNode.AddFQDNAliasesToHTTPPolicy(sniNode.VHDomainNames, key)
		sniNode.AviMarkers.Host = sniNode.VHDomainNames
		objects.SharedCRDLister().UpdateFQDNToAliasesMappings(sniHost, sniNode.VHDomainNames)
	} else {
		hostMapOk, ingressHostMap := SharedHostNameLister().Get(sniHost)
		if hostMapOk {
			// Replace the ingress map for this host.
			keyToRemove := namespace + "/" + ingName
			delete(ingressHostMap.HostNameMap, keyToRemove)
			SharedHostNameLister().Save(sniHost, ingressHostMap)
		}
		// Since the cert couldn't be built, check if this SNI is affected by only in ingress if so remove the sni node from the model
		if len(ingressHostMap.GetIngressesForHostName()) == 0 {
			sniHostToRemove = append(sniHostToRemove, sniNode.VHDomainNames...)
			if !isDedicated {
				RemoveSniInModel(sniNode.Name, vsNode, key)
				RemoveRedirectHTTPPolicyInModel(vsNode[0], sniHostToRemove, key)
			} else {
				DeleteDedicatedVSNode(vsNode[0], sniHostToRemove, key)
			}
			vsNode[0].RemoveFQDNsFromModel(sniHostToRemove, key)
		}
	}
}
