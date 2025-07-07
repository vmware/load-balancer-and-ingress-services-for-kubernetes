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

package rest

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/jinzhu/copier"
	avimodels "github.com/vmware/alb-sdk/go/models"
	"google.golang.org/protobuf/proto"

	avicache "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

func (rest *RestOperations) RestOperationForEvh(vsName string, namespace string, avimodel *nodes.AviObjectGraph, sniNode bool, vs_cache_obj *avicache.AviVsCache, key string) {
	var pools_to_delete []avicache.NamespaceName
	var pgs_to_delete []avicache.NamespaceName
	var vsvip_to_delete []avicache.NamespaceName
	var sni_to_delete []avicache.NamespaceName
	var httppol_to_delete []avicache.NamespaceName
	var l4pol_to_delete []avicache.NamespaceName
	var sslkey_cert_delete []avicache.NamespaceName
	var string_groups_to_delete []avicache.NamespaceName
	var vsvipErr error
	var publishKey string

	vsKey := avicache.NamespaceName{Namespace: namespace, Name: vsName}
	aviVsNode := avimodel.GetAviEvhVS()[0]
	if avimodel != nil && len(avimodel.GetAviEvhVS()) > 0 {
		publishKey = avimodel.GetAviEvhVS()[0].Name
	}
	if publishKey == "" {
		// This is a delete case for the virtualservice. Derive the virtualservice from the 'key'
		splitKeys := strings.Split(key, "/")
		if len(splitKeys) == 2 {
			publishKey = splitKeys[1]
		}
	}
	nsPublishKey := avicache.NamespaceName{Namespace: namespace, Name: publishKey}
	// Order would be this: 1. Pools 2. PGs  3. DS. 4. SSLKeyCert 5. VS
	if vs_cache_obj != nil {
		var rest_ops []*utils.RestOp
		vsvip_to_delete, rest_ops, vsvipErr = rest.VSVipCU(aviVsNode.VSVIPRefs, vs_cache_obj, namespace, rest_ops, key)
		if vsvipErr != nil {
			if rest.CheckAndPublishForRetry(vsvipErr, nsPublishKey, key, avimodel) {
				return
			}
		}
		sslkey_cert_delete, rest_ops = rest.CACertCU(aviVsNode.CACertRefs, vs_cache_obj.SSLKeyCertCollection, namespace, rest_ops, key)
		// SSLKeyCertCollection which did not match cacerts are present in the list sslkey_cert_delete,
		// which shuld be the new SSLKeyCertCollection
		sslkey_cert_delete, rest_ops = rest.SSLKeyCertCU(aviVsNode.SSLKeyCertRefs, sslkey_cert_delete, namespace, rest_ops, key)
		pools_to_delete, rest_ops = rest.PoolCU(aviVsNode.PoolRefs, vs_cache_obj, namespace, rest_ops, key)
		pgs_to_delete, rest_ops = rest.PoolGroupCU(aviVsNode.PoolGroupRefs, vs_cache_obj, namespace, rest_ops, key)
		string_groups_to_delete, rest_ops = rest.StringGroupVsCU(aviVsNode.StringGroupRefs, vs_cache_obj, namespace, rest_ops, key)
		httppol_to_delete, rest_ops = rest.HTTPPolicyCU(aviVsNode.HttpPolicyRefs, vs_cache_obj, namespace, rest_ops, key)
		utils.AviLog.Debugf("key: %s, msg: stored checksum for VS: %s, model checksum: %s", key, vs_cache_obj.CloudConfigCksum, strconv.Itoa(int(aviVsNode.GetCheckSum())))
		if vs_cache_obj.CloudConfigCksum == strconv.Itoa(int(aviVsNode.GetCheckSum())) {
			utils.AviLog.Debugf("key: %s, msg: the checksums are same for vs %s, not doing anything", key, vs_cache_obj.Name)
		} else {
			utils.AviLog.Debugf("key: %s, msg: the stored checksum for vs is %v, and the obtained checksum for VS is: %v", key, vs_cache_obj.CloudConfigCksum, strconv.Itoa(int(aviVsNode.GetCheckSum())))
			// The checksums are different, so it should be a PUT call.
			restOp := rest.AviVsBuildForEvh(aviVsNode, utils.RestPut, vs_cache_obj, key)
			if restOp != nil {
				rest_ops = append(rest_ops, restOp...)
			}

		}
		if success, _ := rest.ExecuteRestAndPopulateCache(rest_ops, vsKey, avimodel, key, true); !success {
			return
		}
	} else {
		var rest_ops []*utils.RestOp
		_, rest_ops, vsvipErr = rest.VSVipCU(aviVsNode.VSVIPRefs, nil, namespace, rest_ops, key)
		if vsvipErr != nil {
			if rest.CheckAndPublishForRetry(vsvipErr, nsPublishKey, key, avimodel) {
				return
			}
		}
		_, rest_ops = rest.CACertCU(aviVsNode.CACertRefs, []avicache.NamespaceName{}, namespace, rest_ops, key)
		_, rest_ops = rest.SSLKeyCertCU(aviVsNode.SSLKeyCertRefs, nil, namespace, rest_ops, key)
		_, rest_ops = rest.PoolCU(aviVsNode.PoolRefs, nil, namespace, rest_ops, key)
		_, rest_ops = rest.PoolGroupCU(aviVsNode.PoolGroupRefs, nil, namespace, rest_ops, key)
		_, rest_ops = rest.StringGroupVsCU(aviVsNode.StringGroupRefs, nil, namespace, rest_ops, key)
		_, rest_ops = rest.HTTPPolicyCU(aviVsNode.HttpPolicyRefs, nil, namespace, rest_ops, key)

		// The cache was not found - it's a POST call.
		restOp := rest.AviVsBuildForEvh(aviVsNode, utils.RestPost, nil, key)
		if restOp != nil {
			rest_ops = append(rest_ops, restOp...)
		}
		utils.AviLog.Debugf("POST key: %s, vsKey: %s", key, vsKey)
		utils.AviLog.Debugf("POST restops %s", utils.Stringify(rest_ops))
		if success, _ := rest.ExecuteRestAndPopulateCache(rest_ops, vsKey, avimodel, key, true); !success {
			return
		}
	}
	if vs_cache_obj != nil {
		for _, sni_uuid := range vs_cache_obj.SNIChildCollection {
			sni_vs_key, ok := rest.cache.VsCacheMeta.AviCacheGetKeyByUuid(sni_uuid)
			if ok {
				sni_to_delete = append(sni_to_delete, sni_vs_key.(avicache.NamespaceName))
			} else {
				utils.AviLog.Debugf("key: %s, msg: Couldn't get SNI key for uuid: %v", key, sni_uuid)
			}
		}
	}
	var rest_ops []*utils.RestOp
	vsKey = avicache.NamespaceName{Namespace: namespace, Name: vsName}
	rest_ops = rest.SSLKeyCertDelete(sslkey_cert_delete, namespace, rest_ops, key)
	rest_ops = rest.VSVipDelete(vsvip_to_delete, namespace, rest_ops, key)
	rest_ops = rest.HTTPPolicyDelete(httppol_to_delete, namespace, rest_ops, key)
	rest_ops = rest.StringGroupDelete(string_groups_to_delete, namespace, rest_ops, key)
	rest_ops = rest.L4PolicyDelete(l4pol_to_delete, namespace, rest_ops, key)
	rest_ops = rest.PoolGroupDelete(pgs_to_delete, namespace, rest_ops, key)
	rest_ops = rest.PoolDelete(pools_to_delete, namespace, rest_ops, nil, key)
	if success, _ := rest.ExecuteRestAndPopulateCache(rest_ops, vsKey, avimodel, key, true); !success {
		return
	}

	for _, evhNode := range aviVsNode.EvhNodes {
		utils.AviLog.Debugf("key: %s, msg: processing EVH node: %s", key, evhNode.Name)
		utils.AviLog.Debugf("key: %s, msg: probable EVH delete candidates: %s", key, sni_to_delete)
		var evh_rest_ops []*utils.RestOp
		vsKey = avicache.NamespaceName{Namespace: namespace, Name: evhNode.Name}
		if vs_cache_obj != nil {
			sni_to_delete, evh_rest_ops = rest.EvhNodeCU(evhNode, vs_cache_obj, namespace, sni_to_delete, evh_rest_ops, key)
		} else {
			_, evh_rest_ops = rest.EvhNodeCU(evhNode, nil, namespace, sni_to_delete, evh_rest_ops, key)
		}
		if success, processNextChild := rest.ExecuteRestAndPopulateCache(evh_rest_ops, vsKey, avimodel, key, true); !success {
			if !processNextChild {
				utils.AviLog.Infof("key: %s, msg: Failure in processing EVH node: %s. Not processing other child nodes.", key, evhNode.Name)
				return
			}
		}
	}

	// Let's populate all the DELETE entries
	if len(sni_to_delete) > 0 {
		utils.AviLog.Infof("key: %s, msg: EVH delete candidates are : %s", key, sni_to_delete)
		var rest_ops []*utils.RestOp
		for _, del_sni := range sni_to_delete {
			rest.SNINodeDelete(del_sni, namespace, rest_ops, avimodel, key)
			if success, _ := rest.ExecuteRestAndPopulateCache(rest_ops, vsKey, avimodel, key, true); !success {
				return
			}
		}

	}

}

func (rest *RestOperations) EvhNodeCU(sni_node *nodes.AviEvhVsNode, vs_cache_obj *avicache.AviVsCache, namespace string, cache_sni_nodes []avicache.NamespaceName, rest_ops []*utils.RestOp, key string) ([]avicache.NamespaceName, []*utils.RestOp) {
	var sni_pools_to_delete []avicache.NamespaceName
	var sni_pgs_to_delete []avicache.NamespaceName
	var http_policies_to_delete []avicache.NamespaceName
	var sslkey_cert_delete []avicache.NamespaceName
	var string_groups_to_delete []avicache.NamespaceName
	var sni_cache_obj *avicache.AviVsCache
	if vs_cache_obj != nil {
		sni_key := avicache.NamespaceName{Namespace: namespace, Name: sni_node.Name}
		// Search the VS cache and obtain the UUID of this VS. Then see if this UUID is part of the SNIChildCollection or not.
		found := utils.HasElem(cache_sni_nodes, sni_key)
		utils.AviLog.Debugf("key: %s, msg: processing node key: %v", key, sni_key)
		if found && cache_sni_nodes != nil {
			cache_sni_nodes = avicache.RemoveNamespaceName(cache_sni_nodes, sni_key)
			utils.AviLog.Debugf("key: %s, msg: the cache evh nodes are: %v", key, cache_sni_nodes)
			sni_cache_obj = rest.getVsCacheObj(sni_key, key)
			if sni_cache_obj != nil {
				// CAcerts have to be created first, as they are referred by the keycerts
				sslkey_cert_delete, rest_ops = rest.CACertCU(sni_node.CACertRefs, sni_cache_obj.SSLKeyCertCollection, namespace, rest_ops, key)
				// SSLKeyCertCollection which did not match cacerts are present in the list sslkey_cert_delete,
				// which shuld be the new SSLKeyCertCollection
				sslkey_cert_delete, rest_ops = rest.SSLKeyCertCU(sni_node.SSLKeyCertRefs, sslkey_cert_delete, namespace, rest_ops, key)
				sni_pools_to_delete, rest_ops = rest.PoolCU(sni_node.PoolRefs, sni_cache_obj, namespace, rest_ops, key)
				sni_pgs_to_delete, rest_ops = rest.PoolGroupCU(sni_node.PoolGroupRefs, sni_cache_obj, namespace, rest_ops, key)
				string_groups_to_delete, rest_ops = rest.StringGroupVsCU(sni_node.StringGroupRefs, sni_cache_obj, namespace, rest_ops, key)
				http_policies_to_delete, rest_ops = rest.HTTPPolicyCU(sni_node.HttpPolicyRefs, sni_cache_obj, namespace, rest_ops, key)

				// The checksums are different, so it should be a PUT call.
				if sni_cache_obj.CloudConfigCksum != strconv.Itoa(int(sni_node.GetCheckSum())) {
					restOp := rest.AviVsBuildForEvh(sni_node, utils.RestPut, sni_cache_obj, key)
					if restOp != nil {
						rest_ops = append(rest_ops, restOp...)
					}
					utils.AviLog.Infof("key: %s, msg: the checksums are different for evh child %s, operation: PUT", key, sni_node.Name)

				}
			}
		} else {
			utils.AviLog.Debugf("key: %s, msg: evh child %s not found in cache, operation: POST", key, sni_node.Name)
			_, rest_ops = rest.CACertCU(sni_node.CACertRefs, []avicache.NamespaceName{}, namespace, rest_ops, key)
			_, rest_ops = rest.SSLKeyCertCU(sni_node.SSLKeyCertRefs, nil, namespace, rest_ops, key)
			_, rest_ops = rest.PoolCU(sni_node.PoolRefs, nil, namespace, rest_ops, key)
			_, rest_ops = rest.PoolGroupCU(sni_node.PoolGroupRefs, nil, namespace, rest_ops, key)
			_, rest_ops = rest.StringGroupVsCU(sni_node.StringGroupRefs, nil, namespace, rest_ops, key)
			_, rest_ops = rest.HTTPPolicyCU(sni_node.HttpPolicyRefs, nil, namespace, rest_ops, key)

			// Not found - it should be a POST call.
			restOp := rest.AviVsBuildForEvh(sni_node, utils.RestPost, nil, key)
			if restOp != nil {
				rest_ops = append(rest_ops, restOp...)
			}
		}
		rest_ops = rest.SSLKeyCertDelete(sslkey_cert_delete, namespace, rest_ops, key)
		rest_ops = rest.HTTPPolicyDelete(http_policies_to_delete, namespace, rest_ops, key)
		rest_ops = rest.StringGroupDelete(string_groups_to_delete, namespace, rest_ops, key)
		rest_ops = rest.PoolGroupDelete(sni_pgs_to_delete, namespace, rest_ops, key)
		rest_ops = rest.PoolDelete(sni_pools_to_delete, namespace, rest_ops, sni_cache_obj, key)
		utils.AviLog.Debugf("key: %s, msg: the EVH VSes to be deleted are: %s", key, cache_sni_nodes)
	} else {
		utils.AviLog.Debugf("key: %s, msg: EVH child %s not found in cache and EVH parent also does not exist in cache", key, sni_node.Name)
		_, rest_ops = rest.CACertCU(sni_node.CACertRefs, []avicache.NamespaceName{}, namespace, rest_ops, key)
		_, rest_ops = rest.SSLKeyCertCU(sni_node.SSLKeyCertRefs, nil, namespace, rest_ops, key)
		_, rest_ops = rest.PoolCU(sni_node.PoolRefs, nil, namespace, rest_ops, key)
		_, rest_ops = rest.PoolGroupCU(sni_node.PoolGroupRefs, nil, namespace, rest_ops, key)
		_, rest_ops = rest.StringGroupVsCU(sni_node.StringGroupRefs, nil, namespace, rest_ops, key)
		_, rest_ops = rest.HTTPPolicyCU(sni_node.HttpPolicyRefs, nil, namespace, rest_ops, key)

		// Not found - it should be a POST call.
		restOp := rest.AviVsBuildForEvh(sni_node, utils.RestPost, nil, key)
		if restOp != nil {
			rest_ops = append(rest_ops, restOp...)
		}
	}
	return cache_sni_nodes, rest_ops
}

func setDedicatedEvhVSNodeProperties(vs *avimodels.VirtualService, vs_meta *nodes.AviEvhVsNode) {
	var datascriptCollection []*avimodels.VSDataScripts
	// this overwrites the sslkeycert created from the Secret object, with the one mentioned in HostRule.TLS
	if len(vs_meta.SslKeyAndCertificateRefs) != 0 {
		vs.SslKeyAndCertificateRefs = append(vs.SslKeyAndCertificateRefs, vs_meta.SslKeyAndCertificateRefs...)
	} else {
		for _, sslkeycert := range vs_meta.SSLKeyCertRefs {
			certName := "/api/sslkeyandcertificate/?name=" + sslkeycert.Name
			vs.SslKeyAndCertificateRefs = append(vs.SslKeyAndCertificateRefs, certName)
		}
	}
	vs.SslProfileRef = vs_meta.SslProfileRef

	//set datascripts to VS from hostrule crd
	for i, script := range vs_meta.VsDatascriptRefs {
		j := int32(i)
		datascript := script
		datascripts := &avimodels.VSDataScripts{VsDatascriptSetRef: &datascript, Index: &j}
		datascriptCollection = append(datascriptCollection, datascripts)
	}
	vs.VsDatascripts = datascriptCollection

	if vs_meta.ApplicationProfileRef != nil {
		// hostrule ref overrides defaults
		vs.ApplicationProfileRef = vs_meta.ApplicationProfileRef
	}

	if len(vs_meta.ICAPProfileRefs) != 0 {
		vs.IcapRequestProfileRefs = vs_meta.ICAPProfileRefs
	}

	vs.WafPolicyRef = vs_meta.WafPolicyRef
	vs.ErrorPageProfileRef = &vs_meta.ErrorPageProfileRef
	vs.AnalyticsProfileRef = vs_meta.AnalyticsProfileRef
	vs.EastWestPlacement = proto.Bool(false)
	vs.Enabled = vs_meta.Enabled
	normal_vs_type := utils.VS_TYPE_NORMAL
	vs.Type = &normal_vs_type
}

func (rest *RestOperations) AviVsBuildForEvh(vs_meta *nodes.AviEvhVsNode, rest_method utils.RestMethod, cache_obj *avicache.AviVsCache, key string) []*utils.RestOp {

	if lib.CheckObjectNameLength(vs_meta.Name, lib.EVHVS) {
		utils.AviLog.Warnf("key: %s not Processing EVHVS object", key)
		return nil
	}

	if !vs_meta.EVHParent && !vs_meta.Dedicated {
		rest_ops := rest.AviVsChildEvhBuild(vs_meta, rest_method, cache_obj, key)
		return rest_ops
	} else {
		// This is EVH Parent or dedicated VS
		network_prof := "/api/networkprofile/?name=" + vs_meta.NetworkProfile
		app_prof := "/api/applicationprofile/?name=" + vs_meta.ApplicationProfile
		if vs_meta.ApplicationProfileRef != nil {
			// hostrule ref overrides defaults
			app_prof = *vs_meta.ApplicationProfileRef
		}

		name := vs_meta.Name
		cksum := vs_meta.CloudConfigCksum
		checksumstr := strconv.Itoa(int(cksum))
		cr := lib.AKOUser
		cloudRef := fmt.Sprintf("/api/cloud?name=%s", utils.CloudName)
		svc_mdata_json, _ := json.Marshal(&vs_meta.ServiceMetadata)
		svc_mdata := string(svc_mdata_json)

		seGroupRef := "/api/serviceenginegroup?name=" + vs_meta.ServiceEngineGroup
		vs := avimodels.VirtualService{
			Name:                  &name,
			NetworkProfileRef:     &network_prof,
			ApplicationProfileRef: &app_prof,
			CloudConfigCksum:      &checksumstr,
			CreatedBy:             &cr,
			CloudRef:              &cloudRef,
			ServiceMetadata:       &svc_mdata,
			SeGroupRef:            &seGroupRef,
		}

		if len(vs_meta.ICAPProfileRefs) != 0 {
			vs.IcapRequestProfileRefs = vs_meta.ICAPProfileRefs
		}

		if vs_meta.VrfContext != "" {
			vs.VrfContextRef = proto.String("/api/vrfcontext?name=" + vs_meta.VrfContext)
		}

		if vs_meta.ErrorPageProfileRef != "" {
			vs.ErrorPageProfileRef = &vs_meta.ErrorPageProfileRef
		}

		var enableRhi bool
		if vs_meta.EnableRhi != nil {
			enableRhi = *vs_meta.EnableRhi
		} else {
			enableRhi = lib.GetEnableRHI()
		}
		if enableRhi {
			vs.EnableRhi = &enableRhi
		}

		if vs_meta.SharedVS {
			vs.Markers = lib.GetMarkers()
		} else {
			vs.Markers = lib.GetAllMarkers(vs_meta.AviMarkers)
		}

		if vs_meta.DefaultPoolGroup != "" {
			pool_ref := "/api/poolgroup/?name=" + vs_meta.DefaultPoolGroup
			vs.PoolGroupRef = &pool_ref
		}
		if len(vs_meta.VSVIPRefs) > 0 {
			vipref := "/api/vsvip/?name=" + vs_meta.VSVIPRefs[0].Name
			vs.VsvipRef = &vipref
		} else {
			utils.AviLog.Warnf("key: %s, msg: unable to set the vsvip reference")
		}
		tenant := fmt.Sprintf("/api/tenant/?name=%s", lib.GetEscapedValue(vs_meta.Tenant))
		vs.TenantRef = &tenant

		if vs_meta.EVHParent {
			// This is a EVH parent
			utils.AviLog.Debugf("key: %s, msg: vs %s is a EVH Parent", key, vs_meta.Name)
			vh_parent := utils.VS_TYPE_VH_PARENT
			vs.Type = &vh_parent
		}
		// TODO other fields like cloud_ref, mix of TCP & UDP protocols, etc.

		for i, pp := range vs_meta.PortProto {
			port := uint32(pp.Port)
			svc := avimodels.Service{Port: &port, EnableSsl: &vs_meta.PortProto[i].EnableSSL, EnableHttp2: &vs_meta.PortProto[i].EnableHTTP2}
			vs.Services = append(vs.Services, &svc)
		}

		var httpPolicyCollection []*avimodels.HTTPPolicies
		internalPolicyIndexBuffer := int32(11)
		if len(vs_meta.HttpPolicyRefs) > 0 {
			for i, http := range vs_meta.HttpPolicyRefs {
				// Update them on the VS object
				j := int32(i) + internalPolicyIndexBuffer
				httpPolicy := fmt.Sprintf("/api/httppolicyset/?name=%s", http.Name)
				httpPolicies := &avimodels.HTTPPolicies{HTTPPolicySetRef: &httpPolicy, Index: &j}
				httpPolicyCollection = append(httpPolicyCollection, httpPolicies)
			}
		}

		bufferLen := int32(len(httpPolicyCollection)) + internalPolicyIndexBuffer + 5
		for i, policy := range vs_meta.HttpPolicySetRefs {
			j := int32(i) + bufferLen
			httpPolicy := policy
			httpPolicies := &avimodels.HTTPPolicies{HTTPPolicySetRef: &httpPolicy, Index: &j}
			httpPolicyCollection = append(httpPolicyCollection, httpPolicies)
		}
		vs.HTTPPolicies = httpPolicyCollection

		// Datascripts from hostrule.
		var datascriptCollection []*avimodels.VSDataScripts
		for i, script := range vs_meta.VsDatascriptRefs {
			j := int32(i)
			datascript := script
			datascripts := &avimodels.VSDataScripts{VsDatascriptSetRef: &datascript, Index: &j}
			datascriptCollection = append(datascriptCollection, datascripts)
		}
		if len(datascriptCollection) > 0 {
			vs.VsDatascripts = datascriptCollection
		}

		if vs_meta.TLSType != utils.TLS_PASSTHROUGH && !vs_meta.Dedicated {
			//Append cert from hostrule
			for _, evhNode := range vs_meta.EvhNodes {
				if len(evhNode.SslKeyAndCertificateRefs) != 0 {
					for _, evhcert := range evhNode.SslKeyAndCertificateRefs {
						if !utils.HasElem(vs.SslKeyAndCertificateRefs, evhcert) {
							vs.SslKeyAndCertificateRefs = append(vs.SslKeyAndCertificateRefs, evhcert)
						}
					}
				}
			}
			//Append cert present in ingress/route
			for _, sslkeycert := range vs_meta.SSLKeyCertRefs {
				certName := "/api/sslkeyandcertificate/?name=" + sslkeycert.Name
				vs.SslKeyAndCertificateRefs = append(vs.SslKeyAndCertificateRefs, certName)
			}
			if vs_meta.SslProfileRef != nil {
				vs.SslProfileRef = vs_meta.SslProfileRef
			}

		}
		//Dedicated VS
		if vs_meta.Dedicated {
			setDedicatedEvhVSNodeProperties(&vs, vs_meta)
		} else {
			vs.VhType = proto.String(utils.VS_TYPE_VH_ENHANCED)
		}
		if strings.Contains(*vs.Name, lib.PassthroughPrefix) && !strings.HasSuffix(*vs.Name, lib.PassthroughInsecure) {
			// This is a passthrough secure VS, we want the VS to be down if all the pools are down.
			vsDownOnPoolDown := true
			vs.RemoveListeningPortOnVsDown = &vsDownOnPoolDown
		}
		vs.AnalyticsPolicy = vs_meta.GetAnalyticsPolicy()

		if err := copier.CopyWithOption(&vs, &vs_meta.AviVsNodeGeneratedFields, copier.Option{IgnoreEmpty: true}); err != nil {
			utils.AviLog.Warnf("key: %s, msg: unable to set few parameters in the VS, err: %v", key, err)
		}

		var rest_ops []*utils.RestOp

		var rest_op utils.RestOp
		var path string

		// VS objects cache can be created by other objects and they would just set VS name and not uud
		// Do a POST call in that case
		if rest_method == utils.RestPut && cache_obj.Uuid != "" {
			path = "/api/virtualservice/" + cache_obj.Uuid
			rest_op = utils.RestOp{
				ObjName: vs_meta.Name,
				Path:    path,
				Method:  rest_method,
				Obj:     vs,
				Tenant:  vs_meta.Tenant,
				Model:   "VirtualService",
			}
			// This will be populated for GW in GwAPI deployment
			if vs_meta.Caller != "" {
				rest_op.Caller = vs_meta.Caller
			}
			rest_ops = append(rest_ops, &rest_op)

		} else {
			rest_method = utils.RestPost
			path = "/api/virtualservice/"
			rest_op = utils.RestOp{
				ObjName: vs_meta.Name,
				Path:    path,
				Method:  rest_method,
				Obj:     vs,
				Tenant:  vs_meta.Tenant,
				Model:   "VirtualService",
			}
			// This will be populated for GW in GwAPI deployment
			if vs_meta.Caller != "" {
				rest_op.Caller = vs_meta.Caller
			}
			rest_ops = append(rest_ops, &rest_op)

		}
		return rest_ops
	}
}

func (rest *RestOperations) AviVsChildEvhBuild(vs_meta *nodes.AviEvhVsNode, rest_method utils.RestMethod, cache_obj *avicache.AviVsCache, key string) []*utils.RestOp {
	name := vs_meta.Name
	cksum := vs_meta.CloudConfigCksum
	checksumstr := strconv.Itoa(int(cksum))
	cr := lib.AKOUser

	var app_prof string
	app_prof = "/api/applicationprofile/?name=" + vs_meta.ApplicationProfile
	if vs_meta.ApplicationProfileRef != nil {
		// hostrule ref overrides defaults
		app_prof = *vs_meta.ApplicationProfileRef
	}

	cloudRef := fmt.Sprintf("/api/cloud?name=%s", utils.CloudName)
	network_prof := "/api/networkprofile/?name=" + "System-TCP-Proxy"
	seGroupRef := fmt.Sprintf("/api/serviceenginegroup?name=%s", lib.GetSEGName())
	svc_mdata_json, _ := json.Marshal(&vs_meta.ServiceMetadata)
	svc_mdata := string(svc_mdata_json)
	evhChild := &avimodels.VirtualService{
		Name:                  &name,
		CloudConfigCksum:      &checksumstr,
		CreatedBy:             &cr,
		NetworkProfileRef:     &network_prof,
		ApplicationProfileRef: &app_prof,
		EastWestPlacement:     proto.Bool(false),
		CloudRef:              &cloudRef,
		SeGroupRef:            &seGroupRef,
		ServiceMetadata:       &svc_mdata,
		WafPolicyRef:          vs_meta.WafPolicyRef,
		SslProfileRef:         vs_meta.SslProfileRef,
		AnalyticsProfileRef:   vs_meta.AnalyticsProfileRef,
		ErrorPageProfileRef:   &vs_meta.ErrorPageProfileRef,
		Enabled:               vs_meta.Enabled,
		VhType:                proto.String(utils.VS_TYPE_VH_ENHANCED),
	}

	if len(vs_meta.ICAPProfileRefs) != 0 {
		evhChild.IcapRequestProfileRefs = vs_meta.ICAPProfileRefs
	}

	if vs_meta.VrfContext != "" {
		evhChild.VrfContextRef = proto.String("/api/vrfcontext?name=" + vs_meta.VrfContext)
	}
	//This VS has a TLSKeyCert associated, we need to mark 'type': 'VS_TYPE_VH_PARENT'
	vh_type := utils.VS_TYPE_VH_CHILD
	evhChild.Type = &vh_type
	vhParentUuid := "/api/virtualservice/?name=" + vs_meta.VHParentName
	evhChild.VhParentVsRef = &vhParentUuid

	ignPool := false
	evhChild.IgnPoolNetReach = &ignPool
	// Fill vhmatch information
	vhMatches := make([]*avimodels.VHMatch, 0)
	for _, Vhostname := range vs_meta.VHDomainNames {
		match_case := "SENSITIVE"
		matchCriteria := "BEGINS_WITH"
		name := "VHMATCHRULENAME"
		path_match := avimodels.PathMatch{
			MatchCriteria: &matchCriteria,
			MatchCase:     &match_case,
			MatchStr:      []string{"/"},
		}
		vHMatchRules := make([]*avimodels.VHMatchRule, 0)
		matchTarget := &avimodels.MatchTarget{
			Path: &path_match,
		}
		vHMatchRule := &avimodels.VHMatchRule{
			Name:    &name,
			Matches: matchTarget,
		}
		vHMatchRules = append(vHMatchRules, vHMatchRule)

		hostname := Vhostname
		vhMatch := &avimodels.VHMatch{Host: &hostname, Rules: vHMatchRules}
		vhMatches = append(vhMatches, vhMatch)
	}

	evhChild.VhMatches = vhMatches

	if vs_meta.VHMatches != nil {
		evhChild.VhMatches = vs_meta.VHMatches
	}

	evhChild.Markers = lib.GetAllMarkers(vs_meta.AviMarkers)

	if vs_meta.DefaultPool != "" {
		pool_ref := "/api/pool/?name=" + vs_meta.DefaultPool
		evhChild.PoolRef = &pool_ref
	}

	if vs_meta.DefaultPoolGroup != "" {
		pg_ref := "/api/poolgroup/?name=" + vs_meta.DefaultPoolGroup
		evhChild.PoolGroupRef = &pg_ref
	}
	var datascriptCollection []*avimodels.VSDataScripts

	//DS from hostrule
	for i, script := range vs_meta.VsDatascriptRefs {
		j := int32(i)
		datascript := script
		datascripts := &avimodels.VSDataScripts{VsDatascriptSetRef: &datascript, Index: &j}
		datascriptCollection = append(datascriptCollection, datascripts)
	}
	evhChild.VsDatascripts = datascriptCollection

	// No need of HTTP rules for TLS passthrough.
	if vs_meta.TLSType != utils.TLS_PASSTHROUGH {
		evhChild.HTTPPolicies = AviVsHttpPSAdd(vs_meta, true)
	}
	evhChild.AnalyticsPolicy = vs_meta.GetAnalyticsPolicy()
	if err := copier.CopyWithOption(&evhChild, &vs_meta.AviVsNodeGeneratedFields, copier.Option{IgnoreEmpty: true}); err != nil {
		utils.AviLog.Warnf("key: %s, msg: unable to set few parameters in the child VS, err: %v", key, err)
	}

	var rest_ops []*utils.RestOp
	var rest_op utils.RestOp
	var path string
	if rest_method == utils.RestPut {

		path = "/api/virtualservice/" + cache_obj.Uuid
		rest_op = utils.RestOp{
			ObjName: vs_meta.Name,
			Path:    path,
			Method:  rest_method,
			Obj:     evhChild,
			Tenant:  vs_meta.Tenant,
			Model:   "VirtualService",
		}
		rest_ops = append(rest_ops, &rest_op)

	} else {
		path = "/api/virtualservice"
		rest_op = utils.RestOp{
			ObjName: vs_meta.Name,
			Path:    path,
			Method:  rest_method,
			Obj:     evhChild,
			Tenant:  vs_meta.Tenant,
			Model:   "VirtualService",
		}
		rest_ops = append(rest_ops, &rest_op)

	}

	return rest_ops
}
