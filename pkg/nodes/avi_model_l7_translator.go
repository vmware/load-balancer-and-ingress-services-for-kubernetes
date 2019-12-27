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

	avimodels "github.com/avinetworks/sdk/go/models"
	"gitlab.eng.vmware.com/orion/container-lib/utils"
	extensionv1beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
)

// Candidate for utils.
var shardSizeMap = map[string]uint32{
	"LARGE":  8,
	"MEDIUM": 4,
	"SMALL":  2,
}

func (o *AviObjectGraph) BuildL7VSGraph(vsName string, namespace string, ingName string, key string) {
	// We create pools and attach servers to them here. Pools are created with a priorty label of host/path
	ingObj, err := utils.GetInformers().IngressInformer.Lister().Ingresses(namespace).Get(ingName)
	if err != nil {
		// A case, where we detected in Layer 2 that the ingress has been deleted.
		if errors.IsNotFound(err) {
			utils.AviLog.Info.Printf("key: %s, msg: ingress not found:  %s", key, ingName)

			// Fetch the ingress pools that are present in the model and delete them.
			poolNodes := o.GetAviPoolNodesByIngress(namespace, ingName)
			utils.AviLog.Info.Printf("key: %s, msg: Pool Nodes to delete for ingress:  %s", key, utils.Stringify(poolNodes))

			for _, pool := range poolNodes {
				o.RemoveModelNode(pool.Name)
			}
		}
	} else {
		// First check if there are pools related to this ingress present in the model already
		poolNodes := o.GetAviPoolNodesByIngress(namespace, ingName)
		for _, pool := range poolNodes {
			o.RemoveModelNode(pool.Name)
		}
		// Create fresh pools
		pgName := vsName + utils.L7_PG_PREFIX
		// PGs are in 'admin' namespace right now.
		pgNode := o.GetPoolGroupByName(pgName)
		if pgNode != nil {
			hostPathSvcList := parseHostPathForIngress(ingName, ingObj.Spec, key)
			for _, obj := range hostPathSvcList {
				var priorityLabel string
				if obj.Path != "" {
					priorityLabel = obj.Host + "/" + obj.Path
				} else {
					priorityLabel = obj.Host
				}
				poolNode := &AviPoolNode{Name: "pool-" + priorityLabel, IngressName: ingName, Tenant: namespace, PriorityLabel: priorityLabel, Port: obj.Port}
				if servers := PopulateServers(poolNode, namespace, obj.ServiceName, key); servers != nil {
					poolNode.Servers = servers
				}
				pool_ref := fmt.Sprintf("/api/pool?name=%s", poolNode.Name)
				pgNode.Members = append(pgNode.Members, &avimodels.PoolGroupMember{PoolRef: &pool_ref, PriorityLabel: &priorityLabel})
				o.AddModelNode(poolNode)
			}
		}
	}
}

func parseHostPathForIngress(ingName string, ingSpec extensionv1beta1.IngressSpec, key string) []IngressHostPathSvc {
	// Figure out the service names that are part of this ingress
	var hostPathMapSvcList []IngressHostPathSvc

	for _, rule := range ingSpec.Rules {
		var hostName string
		if rule.Host == "" {
			// The Host field is empty. Generate a hostName using the sub-domain info from configmap
			hostName = ingName // (TODO): Add sub-domain
		} else {
			hostName = rule.Host
		}
		for _, path := range rule.IngressRuleValue.HTTP.Paths {
			hostPathMapSvc := IngressHostPathSvc{}
			hostPathMapSvc.Host = hostName
			hostPathMapSvc.Path = path.Path
			hostPathMapSvc.ServiceName = path.Backend.ServiceName
			hostPathMapSvc.Port = path.Backend.ServicePort.IntVal
			if hostPathMapSvc.Port == 0 {
				// Default to port 80 if not set in the ingress object
				hostPathMapSvc.Port = 80
			}
			hostPathMapSvcList = append(hostPathMapSvcList, hostPathMapSvc)
		}
	}
	utils.AviLog.Info.Printf("key: %s, msg: host path obtained from ingress:  %v", key, hostPathMapSvcList)
	return hostPathMapSvcList
}

func (o *AviObjectGraph) ConstructAviL7VsNode(vsName string, key string) *AviVsNode {
	var avi_vs_meta *AviVsNode
	// This is a shared VS - always created in the admin namespace for now.
	avi_vs_meta = &AviVsNode{Name: vsName, Tenant: utils.ADMIN_NS,
		EastWest: false}
	// Hard coded ports for the shared VS
	var portProtocols []AviPortHostProtocol
	httpPort := AviPortHostProtocol{Port: 80, Protocol: utils.HTTP}
	httpsPort := AviPortHostProtocol{Port: 443, Protocol: utils.HTTP}
	portProtocols = append(portProtocols, httpPort)
	portProtocols = append(portProtocols, httpsPort)
	avi_vs_meta.PortProto = portProtocols
	// Default case.
	avi_vs_meta.ApplicationProfile = utils.DEFAULT_L7_APP_PROFILE
	avi_vs_meta.NetworkProfile = utils.DEFAULT_TCP_NW_PROFILE
	o.AddModelNode(avi_vs_meta)
	o.ConstructShardVsPGNode(vsName, key, avi_vs_meta)
	o.ConstructHTTPDataScript(vsName, key, avi_vs_meta)
	return avi_vs_meta
}

func (o *AviObjectGraph) ConstructShardVsPGNode(vsName string, key string, vsNode *AviVsNode) *AviPoolGroupNode {
	pgName := vsName + utils.L7_PG_PREFIX
	pgNode := &AviPoolGroupNode{Name: pgName, Tenant: utils.ADMIN_NS}
	vsNode.PoolGroupRefs = append(vsNode.PoolGroupRefs, pgNode)
	o.AddModelNode(pgNode)
	return pgNode
}

func (o *AviObjectGraph) ConstructHTTPDataScript(vsName string, key string, vsNode *AviVsNode) *AviHTTPDataScriptNode {
	scriptStr := utils.HTTP_DS_SCRIPT
	evt := utils.VS_DATASCRIPT_EVT_HTTP_REQ
	var poolGroupRefs []string
	pgName := "/api/poolgroup?name=" + vsName + "-pg-l7"
	poolGroupRefs = append(poolGroupRefs, pgName)
	dsName := vsName + "-http-datascript"
	script := &DataScript{Script: scriptStr, Evt: evt}
	dsScriptNode := &AviHTTPDataScriptNode{Name: dsName, DataScript: script, PoolGroupRefs: poolGroupRefs}
	vsNode.HTTPDSrefs = append(vsNode.HTTPDSrefs, dsScriptNode)
	o.AddModelNode(dsScriptNode)
	return dsScriptNode
}
