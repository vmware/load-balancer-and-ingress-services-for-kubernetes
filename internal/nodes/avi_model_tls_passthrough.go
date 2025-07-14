/*
 * Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
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
	akov1beta1 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/apis/ako/v1beta1"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	avimodels "github.com/vmware/alb-sdk/go/models"
)

func (o *AviObjectGraph) BuildVSForPassthrough(vsName, namespace, hostname, tenant, key string, infraSetting *akov1beta1.AviInfraSetting) *AviVsNode {
	o.Lock.Lock()
	defer o.Lock.Unlock()
	var avi_vs_meta *AviVsNode

	// create the secured shared VS to listen on port 443
	avi_vs_meta = &AviVsNode{
		Name:               vsName,
		Tenant:             tenant,
		SharedVS:           true,
		ServiceEngineGroup: lib.GetSEGName(),
	}

	enableRhi := lib.GetEnableRHI()
	avi_vs_meta.EnableRhi = &enableRhi

	var portProtocols []AviPortHostProtocol
	httpsPort := AviPortHostProtocol{Port: 443, Protocol: utils.HTTP}
	portProtocols = append(portProtocols, httpsPort)
	avi_vs_meta.PortProto = portProtocols

	avi_vs_meta.ApplicationProfile = utils.DEFAULT_L4_APP_PROFILE
	avi_vs_meta.NetworkProfile = utils.DEFAULT_TCP_NW_PROFILE

	vrfcontext := ""
	t1lr := lib.GetT1LRPath()
	if infraSetting != nil && infraSetting.Spec.NSXSettings.T1LR != nil {
		t1lr = *infraSetting.Spec.NSXSettings.T1LR
	}
	if t1lr == "" {
		vrfcontext = lib.GetVrf()
		avi_vs_meta.VrfContext = vrfcontext
	}

	o.AddModelNode(avi_vs_meta)
	o.ConstructL4DataScript(vsName, key, avi_vs_meta)
	var fqdns []string
	fqdns = append(fqdns, hostname)

	// VSvip node to be shared by the secure and insecure VS
	vsVipNode := &AviVSVIPNode{
		Name:        lib.GetVsVipName(vsName),
		Tenant:      tenant,
		FQDNs:       fqdns,
		VrfContext:  vrfcontext,
		VipNetworks: utils.GetVipNetworkList(),
	}

	if t1lr != "" {
		vsVipNode.T1Lr = t1lr
	}

	if avi_vs_meta.EnableRhi != nil && *avi_vs_meta.EnableRhi {
		vsVipNode.BGPPeerLabels = lib.GetGlobalBgpPeerLabels()
	}

	avi_vs_meta.VSVIPRefs = append(avi_vs_meta.VSVIPRefs, vsVipNode)
	return avi_vs_meta
}

func (o *AviObjectGraph) BuildGraphForPassthrough(svclist []IngressHostPathSvc, objName, hostname, namespace, tenant, key string, redirect bool, infraSetting *akov1beta1.AviInfraSetting) {
	o.Lock.Lock()
	defer o.Lock.Unlock()
	vsList := o.GetAviVS()
	if len(vsList) == 0 {
		utils.AviLog.Warnf("key: %s, msg: no VS found in the graph", key)
		return
	}
	secureSharedVS := vsList[0]
	datascriptList := o.GetAviHTTPDSNode()
	if len(datascriptList) == 0 {
		utils.AviLog.Warnf("key: %s, msg: no Datascript found in the graph for VS: %s", key, secureSharedVS.Name)
		return
	}
	dsNode := datascriptList[0]
	var infrasettingName string
	var infraSettingNameWithSuffix string
	var pgName string
	if infraSetting != nil {
		if !lib.IsInfraSettingNSScoped(infraSetting.Name, namespace) {
			infrasettingName = infraSetting.Name
			infraSettingNameWithSuffix = infrasettingName + "-"
		}
	}
	//Replace AVIINFRA with infrasettingname if present
	dsNode.Script = strings.Replace(dsNode.Script, "AVIINFRA", infraSettingNameWithSuffix, 1)
	// Get Poolgroup Node, create if not present
	pgName = lib.GetPassthroughPGName(hostname, infrasettingName)
	pgNode := o.GetPoolGroupByName(pgName)
	if pgNode == nil {
		pgNode = &AviPoolGroupNode{Name: pgName, Tenant: tenant}
		o.AddModelNode(pgNode)
		pgNode.AviMarkers = lib.PopulatePassthroughPGMarkers(hostname, infrasettingName)
		utils.AviLog.Infof("key: %s, msg: adding PG %s for the passthrough VS: %s", key, pgName, secureSharedVS.Name)
		utils.AviLog.Debugf("key: %s, Number of PGs %d, Added PG node %s", key, len(dsNode.PoolGroupRefs), utils.Stringify(pgNode.Members))
	}
	isPGNameExceedsAviLimit := false
	if lib.CheckObjectNameLength(pgName, lib.PG) {
		isPGNameExceedsAviLimit = true
	}
	// only add the pg node if not presesnt in the VS
	if !utils.HasElem(secureSharedVS.PoolGroupRefs, pgNode) && !isPGNameExceedsAviLimit {
		secureSharedVS.PoolGroupRefs = append(secureSharedVS.PoolGroupRefs, pgNode)
	}
	if !utils.HasElem(dsNode.PoolGroupRefs, pgName) && !isPGNameExceedsAviLimit {
		dsNode.PoolGroupRefs = append(dsNode.PoolGroupRefs, pgName)
	}

	if !utils.HasElem(secureSharedVS.VSVIPRefs[0].FQDNs, hostname) {
		secureSharedVS.VSVIPRefs[0].FQDNs = append(secureSharedVS.VSVIPRefs[0].FQDNs, hostname)
	}

	// store the Pools in the a temoprary list to be used for populating PG members
	tmpPoolList := []*AviPoolNode{}
	vrfContext := ""
	t1lr := lib.GetT1LRPath()
	if infraSetting != nil && infraSetting.Spec.NSXSettings.T1LR != nil {
		t1lr = *infraSetting.Spec.NSXSettings.T1LR
	}
	if t1lr == "" {
		vrfContext = lib.GetVrf()
	}
	for _, obj := range svclist {
		poolName := lib.GetPassthroughPoolName(hostname, obj.ServiceName, infrasettingName)
		poolNode := o.GetAviPoolNodeByName(poolName)
		if poolNode == nil {
			poolNode = &AviPoolNode{
				Name:       poolName,
				Tenant:     tenant,
				VrfContext: vrfContext,
			}
			poolNode.NetworkPlacementSettings = lib.GetNodeNetworkMap()
			if t1lr != "" {
				poolNode.T1Lr = t1lr
			}
			poolNode.AviMarkers = lib.PopulatePassthroughPoolMarkers(hostname, obj.ServiceName, infrasettingName)
		}
		poolNode.IngressName = objName
		poolNode.PortName = obj.PortName
		poolNode.Port = obj.Port
		poolNode.TargetPort = obj.TargetPort
		poolNode.ServiceMetadata = lib.ServiceMetadataObj{
			IngressName: objName, Namespace: namespace, PoolRatio: obj.weight,
			HostNames: []string{hostname},
		}

		poolNode.Servers = []AviPoolMetaServer{}
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
		buildPoolWithInfraSetting(key, poolNode, infraSetting)
		poolNode.CalculateCheckSum()
		tmpPoolList = append(tmpPoolList, poolNode)
	}

	// Remove existing Pool nodes first from the VS to make sure alternate backends are updated correctly
	for _, pgMember := range pgNode.Members {
		poolRef := pgMember.PoolRef
		poolName := strings.TrimPrefix(*poolRef, "/api/pool?name=")
		o.RemovePoolNodeRefs(poolName)
	}

	pgNode.Members = nil
	// add the pool in the vs and pg members
	for _, poolNode := range tmpPoolList {
		if lib.CheckObjectNameLength(poolNode.Name, lib.PG) {
			// Do not add if length is > limit
			continue
		}
		secureSharedVS.PoolRefs = append(secureSharedVS.PoolRefs, poolNode)
		ratio := poolNode.ServiceMetadata.PoolRatio
		poolRef := fmt.Sprintf("/api/pool?name=%s", poolNode.Name)
		pgNode.Members = append(pgNode.Members, &avimodels.PoolGroupMember{PoolRef: &poolRef, Ratio: &ratio})
	}

	// Check and create insecure shared VS for passthrough
	var passChildVS *AviVsNode
	if len(secureSharedVS.PassthroughChildNodes) > 0 {
		passChildVS = secureSharedVS.PassthroughChildNodes[0]
	}
	var hostnameSlice []string
	hostnameSlice = append(hostnameSlice, hostname)
	if !redirect {
		if passChildVS != nil {
			RemoveRedirectHTTPPolicyInModel(passChildVS, hostnameSlice, key)
		}
		return
	}

	if passChildVS == nil {
		passChildVS = &AviVsNode{
			Name:               secureSharedVS.Name + lib.PassthroughInsecure,
			Tenant:             tenant,
			VrfContext:         vrfContext,
			ServiceEngineGroup: lib.GetSEGName(),
			ApplicationProfile: utils.DEFAULT_L7_APP_PROFILE,
			NetworkProfile:     utils.DEFAULT_TCP_NW_PROFILE,
			PortProto:          []AviPortHostProtocol{{Port: 80, Protocol: utils.HTTP}},
			SharedVS:           true,
		}

		passChildVS.VSVIPRefs = append(passChildVS.VSVIPRefs, secureSharedVS.VSVIPRefs...)
		if infraSetting != nil {
			buildWithInfraSetting(key, namespace, passChildVS, passChildVS.VSVIPRefs[0], infraSetting)
		}
		secureSharedVS.PassthroughChildNodes = append(secureSharedVS.PassthroughChildNodes, passChildVS)

		passChildVS.ServiceMetadata.PassthroughParentRef = secureSharedVS.Name
		secureSharedVS.ServiceMetadata.PassthroughChildRef = passChildVS.Name
	}
	o.BuildPolicyRedirectForVS([]*AviVsNode{passChildVS}, hostnameSlice, namespace, "", hostname, key)
}

func (o *AviObjectGraph) ConstructL4DataScript(vsName string, key string, vsNode *AviVsNode) *AviHTTPDataScriptNode {
	dsScriptNode := &AviHTTPDataScriptNode{
		Name:   lib.GetL7InsecureDSName(vsName),
		Tenant: vsNode.GetTenant(),
		DataScript: &DataScript{
			Script: lib.PassthroughDatascript,
			Evt:    "VS_DATASCRIPT_EVT_L4_REQUEST",
		},
		ProtocolParsers: []string{"/api/protocolparser/?name=Default-TLS"},
	}

	dsScriptNode.Script = strings.Replace(dsScriptNode.Script, "CLUSTER", lib.GetClusterName(), 1)

	vsNode.HTTPDSrefs = append(vsNode.HTTPDSrefs, dsScriptNode)
	o.AddModelNode(dsScriptNode)
	return dsScriptNode
}

func (o *AviObjectGraph) DeleteObjectsForPassthroughHost(vsName, hostname string, routeIgrObj RouteIngressModel, pathSvc map[string][]string, infraSettingName, key string, removeFqdn, removeRedir, secure bool) {
	o.Lock.Lock()
	defer o.Lock.Unlock()
	pgName := lib.GetPassthroughPGName(hostname, infraSettingName)
	pgNode := o.GetPoolGroupByName(pgName)
	if pgNode == nil {
		return
	}
	utils.AviLog.Debugf("key: %s, msg: pg Nodes to delete for ingress: %s", key, utils.Stringify(pgName))
	for _, pgMember := range pgNode.Members {
		poolRef := pgMember.PoolRef
		poolName := strings.TrimPrefix(*poolRef, "/api/pool?name=")
		o.RemovePoolNodeRefs(poolName)
	}
	pgNode.Members = []*avimodels.PoolGroupMember{}

	vsNode := o.GetAviVS()
	if len(vsNode) == 0 {
		return
	}
	o.RemovePGNodeRefs(pgName, vsNode[0])

	dsNode := o.GetAviHTTPDSNode()
	if len(dsNode) == 0 {
		return
	}
	for i, pgref := range dsNode[0].PoolGroupRefs {
		if pgref == pgName {
			dsNode[0].PoolGroupRefs = append(dsNode[0].PoolGroupRefs[:i], dsNode[0].PoolGroupRefs[i+1:]...)
			break
		}
	}

	hosts := []string{hostname}
	if removeFqdn {
		vsNode[0].RemoveFQDNsFromModel(hosts, key)
	}

	if removeRedir {
		if len(vsNode[0].PassthroughChildNodes) > 0 {
			passChild := vsNode[0].PassthroughChildNodes[0]
			utils.AviLog.Infof("key: %s, msg: Removing redirect policy for %s from passthrough VS %s", key, hostname, passChild.Name)
			var hostnameSlice []string
			hostnameSlice = append(hostnameSlice, hostname)
			RemoveRedirectHTTPPolicyInModel(passChild, hostnameSlice, key)
		}
	}

	utils.AviLog.Debugf("key: %s, msg: passthrough pg refs %s", key, utils.Stringify(dsNode[0].PoolGroupRefs))
	utils.AviLog.Debugf("key: %s, msg: passthrough datascript %s", key, utils.Stringify(o.GetAviHTTPDSNode()))
}
