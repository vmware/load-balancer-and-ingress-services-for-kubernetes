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

	avicache "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	avimodels "github.com/avinetworks/sdk/go/models"
)

func (o *AviObjectGraph) BuildVSForPassthrough(vsName, namespace, hostname, key string) *AviVsNode {
	o.Lock.Lock()
	defer o.Lock.Unlock()
	var avi_vs_meta *AviVsNode

	// create the secured shared VS to listen on port 443
	avi_vs_meta = &AviVsNode{Name: vsName, Tenant: lib.GetTenant(),
		EastWest: false, SharedVS: true}
	if lib.GetSEGName() != lib.DEFAULT_GROUP {
		avi_vs_meta.ServiceEngineGroup = lib.GetSEGName()
	}
	var portProtocols []AviPortHostProtocol
	httpsPort := AviPortHostProtocol{Port: 443, Protocol: utils.HTTP}
	portProtocols = append(portProtocols, httpsPort)
	avi_vs_meta.PortProto = portProtocols

	avi_vs_meta.ApplicationProfile = utils.DEFAULT_L4_APP_PROFILE
	avi_vs_meta.NetworkProfile = utils.DEFAULT_TCP_NW_PROFILE

	vrfcontext := lib.GetVrf()
	avi_vs_meta.VrfContext = lib.GetVrf()

	o.AddModelNode(avi_vs_meta)
	o.ConstructL4DataScript(vsName, key, avi_vs_meta)
	var fqdns []string
	fqdns = append(fqdns, hostname)

	// VSvip node to be shared by the secure and insecure VS
	vsVipNode := &AviVSVIPNode{Name: lib.GetVsVipName(vsName), Tenant: lib.GetTenant(), FQDNs: fqdns,
		EastWest: false, VrfContext: vrfcontext}
	avi_vs_meta.VSVIPRefs = append(avi_vs_meta.VSVIPRefs, vsVipNode)
	return avi_vs_meta
}

func (o *AviObjectGraph) BuildGraphForPassthrough(svclist []IngressHostPathSvc, objName, hostname, namesapce, key string, redirect bool) {
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

	// Get Poolgroup Node, create if not present
	pgName := lib.GetClusterName() + "--" + hostname
	pgNode := o.GetPoolGroupByName(pgName)
	if pgNode == nil {
		pgNode = &AviPoolGroupNode{Name: pgName, Tenant: lib.GetTenant()}
		o.AddModelNode(pgNode)

		utils.AviLog.Infof("key: %s, msg: adding PG %s for the passthrough VS: %s", key, pgName, secureSharedVS.Name)
		utils.AviLog.Debugf("key: %s, Number of PGs %d, Added PG node %s", len(dsNode.PoolGroupRefs), utils.Stringify(pgNode.Members))
	}

	// only add the pg node if not presesnt in the VS
	if !utils.HasElem(secureSharedVS.PoolGroupRefs, pgNode) {
		secureSharedVS.PoolGroupRefs = append(secureSharedVS.PoolGroupRefs, pgNode)
	}
	if !utils.HasElem(dsNode.PoolGroupRefs, pgName) {
		dsNode.PoolGroupRefs = append(dsNode.PoolGroupRefs, pgName)
	}

	if !utils.HasElem(secureSharedVS.VSVIPRefs[0].FQDNs, hostname) {
		secureSharedVS.VSVIPRefs[0].FQDNs = append(secureSharedVS.VSVIPRefs[0].FQDNs, hostname)
	}

	// store the Pools in the a temoprary list to be used for populating PG members
	tmpPoolList := []*AviPoolNode{}
	for _, obj := range svclist {
		poolName := lib.GetClusterName() + "--" + hostname + "-" + obj.ServiceName
		poolNode := o.GetAviPoolNodeByName(poolName)
		if poolNode == nil {
			poolNode = &AviPoolNode{Name: poolName,
				Tenant:     lib.GetTenant(),
				VrfContext: lib.GetVrf(),
			}
			o.AddModelNode(poolNode)
		}
		poolNode.IngressName = objName
		poolNode.PortName = obj.PortName
		poolNode.Port = obj.Port
		poolNode.TargetPort = obj.TargetPort
		poolNode.ServiceMetadata = avicache.ServiceMetadataObj{
			IngressName: objName, Namespace: namesapce, PoolRatio: obj.weight,
			HostNames: []string{hostname},
		}

		poolNode.Servers = []AviPoolMetaServer{}
		if !lib.IsNodePortMode() {
			if servers := PopulateServers(poolNode, namesapce, obj.ServiceName, true, key); servers != nil {
				poolNode.Servers = servers
			}
		} else {
			if servers := PopulateServersForNodePort(poolNode, namesapce, obj.ServiceName, true, key); servers != nil {
				poolNode.Servers = servers
			}
		}
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

	if !redirect {
		if passChildVS != nil {
			RemoveRedirectHTTPPolicyInModel(passChildVS, hostname, key)
		}
		return
	}

	if passChildVS == nil {
		passChildVS = &AviVsNode{
			Name: secureSharedVS.Name + lib.PassthroughInsecure, Tenant: lib.GetTenant(), EastWest: false, VrfContext: lib.GetVrf(),
		}
		if lib.GetSEGName() != lib.DEFAULT_GROUP {
			passChildVS.ServiceEngineGroup = lib.GetSEGName()
		}
		passChildVS.ApplicationProfile = utils.DEFAULT_L7_APP_PROFILE
		passChildVS.NetworkProfile = utils.DEFAULT_TCP_NW_PROFILE
		httpPort := AviPortHostProtocol{Port: 80, Protocol: utils.HTTP}
		passChildVS.PortProto = []AviPortHostProtocol{httpPort}
		passChildVS.VSVIPRefs = append(passChildVS.VSVIPRefs, secureSharedVS.VSVIPRefs...)
		secureSharedVS.PassthroughChildNodes = append(secureSharedVS.PassthroughChildNodes, passChildVS)

		passChildVS.ServiceMetadata.PassthroughParentRef = secureSharedVS.Name
		secureSharedVS.ServiceMetadata.PassthroughChildRef = passChildVS.Name
	}

	o.BuildPolicyRedirectForVS([]*AviVsNode{passChildVS}, hostname, namesapce, "", key)
}

func (o *AviObjectGraph) ConstructL4DataScript(vsName string, key string, vsNode *AviVsNode) *AviHTTPDataScriptNode {
	scriptStr := lib.PassthroughDatascript
	evt := "VS_DATASCRIPT_EVT_L4_REQUEST"
	dsName := lib.GetL7InsecureDSName(vsName)
	script := &DataScript{Script: scriptStr, Evt: evt}
	dsScriptNode := &AviHTTPDataScriptNode{Name: dsName, Tenant: lib.GetTenant(), DataScript: script}
	dsScriptNode.Script = strings.Replace(dsScriptNode.Script, "CLUSTER", lib.GetClusterName(), 1)
	dsScriptNode.ProtocolParsers = []string{"/api/protocolparser/?name=Default-TLS"}

	vsNode.HTTPDSrefs = append(vsNode.HTTPDSrefs, dsScriptNode)
	o.AddModelNode(dsScriptNode)
	return dsScriptNode
}

func (o *AviObjectGraph) DeleteObjectsForPassthroughHost(vsName, hostname string, routeIgrObj RouteIngressModel, pathSvc map[string][]string, key string, removeFqdn, removeRedir, secure bool) {
	o.Lock.Lock()
	defer o.Lock.Unlock()
	pgName := lib.GetClusterName() + "--" + hostname
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
		RemoveFQDNsFromModel(vsNode[0], hosts, key)
	}

	if removeRedir {
		if len(vsNode[0].PassthroughChildNodes) > 0 {
			passChild := vsNode[0].PassthroughChildNodes[0]
			utils.AviLog.Infof("key: %s, msg: Removing redierct policy for %s from passthrough VS %s", key, hostname, passChild.Name)
			RemoveRedirectHTTPPolicyInModel(passChild, hostname, key)
		}
	}

	utils.AviLog.Debugf("key: %s, msg: passthrough pg refs %s", utils.Stringify(dsNode[0].PoolGroupRefs))
	utils.AviLog.Debugf("key: %s, msg: passthrough datascript %s", utils.Stringify(o.GetAviHTTPDSNode()))
}
