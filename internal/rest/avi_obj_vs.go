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
	"errors"
	"fmt"
	"strconv"
	"strings"

	avicache "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/status"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	avimodels "github.com/avinetworks/sdk/go/models"
	"github.com/davecgh/go-spew/spew"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const VSVIP_NOTFOUND = "VsVip object not found"

func FindPoolGroupForPort(pgList []*nodes.AviPoolGroupNode, portToSearch int32) string {
	for _, pg := range pgList {
		if pg.Port == strconv.Itoa(int(portToSearch)) {
			return pg.Name
		}
	}
	return ""
}

func (rest *RestOperations) AviVsBuild(vs_meta *nodes.AviVsNode, rest_method utils.RestMethod, cache_obj *avicache.AviVsCache, key string) []*utils.RestOp {
	if vs_meta.IsSNIChild {
		rest_ops := rest.AviVsSniBuild(vs_meta, rest_method, cache_obj, key)
		return rest_ops
	} else {
		network_prof := "/api/networkprofile/?name=" + vs_meta.NetworkProfile
		app_prof := "/api/applicationprofile/?name=" + vs_meta.ApplicationProfile
		// TODO use PoolGroup and use policies if there are > 1 pool, etc.
		name := vs_meta.Name
		cksum := vs_meta.CloudConfigCksum
		checksumstr := strconv.Itoa(int(cksum))
		cr := lib.AKOUser
		cloudRef := "/api/cloud?name=" + utils.CloudName
		svc_mdata_json, _ := json.Marshal(&vs_meta.ServiceMetadata)
		svc_mdata := string(svc_mdata_json)
		vrfContextRef := "/api/vrfcontext?name=" + vs_meta.VrfContext
		seGroupRef := "/api/serviceenginegroup?name=" + vs_meta.ServiceEngineGroup
		enableRHI := lib.GetEnableRHI() // We don't impact the checksum of the VS since it's a global setting in AKO.
		vs := avimodels.VirtualService{
			Name:                  &name,
			NetworkProfileRef:     &network_prof,
			ApplicationProfileRef: &app_prof,
			CloudConfigCksum:      &checksumstr,
			CreatedBy:             &cr,
			CloudRef:              &cloudRef,
			ServiceMetadata:       &svc_mdata,
			SeGroupRef:            &seGroupRef,
			VrfContextRef:         &vrfContextRef,
		}
		if enableRHI {
			// If the value is set to false, we would simply remove it from the payload, which should default it to false.
			vs.EnableRhi = &enableRHI
		}
		if lib.GetAdvancedL4() {
			ignPool := true
			vs.IgnPoolNetReach = &ignPool
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
		tenant := fmt.Sprintf("/api/tenant/?name=%s", vs_meta.Tenant)
		vs.TenantRef = &tenant

		if vs_meta.SNIParent {
			// This is a SNI parent
			utils.AviLog.Debugf("key: %s, msg: vs %s is a SNI Parent", key, vs_meta.Name)
			vh_parent := utils.VS_TYPE_VH_PARENT
			vs.Type = &vh_parent
		}
		// TODO other fields like cloud_ref, mix of TCP & UDP protocols, etc.

		for i, pp := range vs_meta.PortProto {
			port := pp.Port
			svc := avimodels.Service{Port: &port, EnableSsl: &vs_meta.PortProto[i].EnableSSL}
			vs.Services = append(vs.Services, &svc)
		}

		if vs_meta.SharedVS {
			// This is a shared VS - which should have a datascript
			var i int32
			var vsdatascripts []*avimodels.VSDataScripts
			for _, ds := range vs_meta.HTTPDSrefs {
				var j int32
				j = i
				dsRef := "/api/vsdatascriptset/?name=" + ds.Name
				vsdatascript := &avimodels.VSDataScripts{Index: &j, VsDatascriptSetRef: &dsRef}
				vsdatascripts = append(vsdatascripts, vsdatascript)
				i = i + 1
			}
			vs.VsDatascripts = vsdatascripts
		}

		if len(vs_meta.HttpPolicyRefs) > 0 {
			var i int32
			i = 0
			var httpPolicyCollection []*avimodels.HTTPPolicies
			for _, http := range vs_meta.HttpPolicyRefs {
				// Update them on the VS object
				var j int32
				j = i + 11
				i = i + 1
				httpPolicy := fmt.Sprintf("/api/httppolicyset/?name=%s", http.Name)
				httpPolicies := &avimodels.HTTPPolicies{HTTPPolicySetRef: &httpPolicy, Index: &j}
				httpPolicyCollection = append(httpPolicyCollection, httpPolicies)
			}
			vs.HTTPPolicies = httpPolicyCollection
		}

		if strings.Contains(*vs.Name, lib.PassthroughPrefix) && !strings.HasSuffix(*vs.Name, lib.PassthroughInsecure) {
			// This is a passthrough secure VS, we want the VS to be down if all the pools are down.
			vsDownOnPoolDown := true
			vs.RemoveListeningPortOnVsDown = &vsDownOnPoolDown
		}
		if len(vs_meta.L4PolicyRefs) > 0 {
			vsDownOnPoolDown := true
			vs.RemoveListeningPortOnVsDown = &vsDownOnPoolDown
			var i int32
			i = 0
			var l4Policies []*avimodels.L4Policies
			for _, l4pol := range vs_meta.L4PolicyRefs {
				// Update them on the VS object
				var j int32
				i = i + 1
				l4PolicyRef := fmt.Sprintf("/api/l4policyset/?name=%s", l4pol.Name)
				l4Policy := &avimodels.L4Policies{L4PolicySetRef: &l4PolicyRef, Index: &j}
				l4Policies = append(l4Policies, l4Policy)
			}
			vs.L4Policies = l4Policies
		}

		var rest_ops []*utils.RestOp

		var rest_op utils.RestOp
		var path string

		// VS objects cache can be created by other objects and they would just set VS name and not uud
		// Do a POST call in that case
		if rest_method == utils.RestPut && cache_obj.Uuid != "" {
			path = "/api/virtualservice/" + cache_obj.Uuid
			rest_op = utils.RestOp{Path: path, Method: rest_method, Obj: vs,
				Tenant: vs_meta.Tenant, Model: "VirtualService", Version: utils.CtrlVersion}
			rest_ops = append(rest_ops, &rest_op)

		} else {
			rest_method = utils.RestPost
			macro := utils.AviRestObjMacro{ModelName: "VirtualService", Data: vs}
			path = "/api/macro"
			rest_op = utils.RestOp{Path: path, Method: rest_method, Obj: macro,
				Tenant: vs_meta.Tenant, Model: "VirtualService", Version: utils.CtrlVersion}
			rest_ops = append(rest_ops, &rest_op)

		}
		return rest_ops
	}
}

func (rest *RestOperations) AviVsSniBuild(vs_meta *nodes.AviVsNode, rest_method utils.RestMethod, cache_obj *avicache.AviVsCache, key string) []*utils.RestOp {
	name := vs_meta.Name
	cksum := vs_meta.CloudConfigCksum
	checksumstr := strconv.Itoa(int(cksum))
	cr := lib.AKOUser

	east_west := false
	var app_prof string
	app_prof = "/api/applicationprofile/?name=" + utils.DEFAULT_L7_SECURE_APP_PROFILE
	if vs_meta.AppProfileRef != "" {
		// hostrule ref overrides defaults
		app_prof = vs_meta.AppProfileRef
	}

	cloudRef := "/api/cloud?name=" + utils.CloudName
	network_prof := "/api/networkprofile/?name=" + "System-TCP-Proxy"
	vrfContextRef := "/api/vrfcontext?name=" + vs_meta.VrfContext
	seGroupRef := "/api/serviceenginegroup?name=" + lib.GetSEGName()
	svc_mdata_json, _ := json.Marshal(&vs_meta.ServiceMetadata)
	svc_mdata := string(svc_mdata_json)
	sniChild := &avimodels.VirtualService{
		Name:                  &name,
		CloudConfigCksum:      &checksumstr,
		CreatedBy:             &cr,
		NetworkProfileRef:     &network_prof,
		ApplicationProfileRef: &app_prof,
		EastWestPlacement:     &east_west,
		CloudRef:              &cloudRef,
		VrfContextRef:         &vrfContextRef,
		SeGroupRef:            &seGroupRef,
		ServiceMetadata:       &svc_mdata,
		WafPolicyRef:          &vs_meta.WafPolicyRef,
		SslProfileRef:         &vs_meta.SSLProfileRef,
		AnalyticsProfileRef:   &vs_meta.AnalyticsProfileRef,
		ErrorPageProfileRef:   &vs_meta.ErrorPageProfileRef,
		Enabled:               vs_meta.Enabled,
	}

	//This VS has a TLSKeyCert associated, we need to mark 'type': 'VS_TYPE_VH_PARENT'
	vh_type := utils.VS_TYPE_VH_CHILD
	sniChild.Type = &vh_type
	vhParentUuid := "/api/virtualservice/?name=" + vs_meta.VHParentName
	sniChild.VhParentVsUUID = &vhParentUuid
	sniChild.VhDomainName = vs_meta.VHDomainNames
	ignPool := false
	sniChild.IgnPoolNetReach = &ignPool

	if vs_meta.DefaultPool != "" {
		pool_ref := "/api/pool/?name=" + vs_meta.DefaultPool
		sniChild.PoolRef = &pool_ref
	}

	var datascriptCollection []*avimodels.VSDataScripts
	for i, script := range vs_meta.VsDatascriptRefs {
		j := int32(i)
		datascript := script
		datascripts := &avimodels.VSDataScripts{VsDatascriptSetRef: &datascript, Index: &j}
		datascriptCollection = append(datascriptCollection, datascripts)
	}
	sniChild.VsDatascripts = datascriptCollection

	// No need of HTTP rules for TLS passthrough.
	if vs_meta.TLSType != utils.TLS_PASSTHROUGH {
		// this overwrites the sslkeycert created from the Secret object, with the one mentioned in HostRule.TLS
		if vs_meta.SSLKeyCertAviRef != "" {
			sniChild.SslKeyAndCertificateRefs = append(sniChild.SslKeyAndCertificateRefs, vs_meta.SSLKeyCertAviRef)
		} else {
			for _, sslkeycert := range vs_meta.SSLKeyCertRefs {
				certName := "/api/sslkeyandcertificate/?name=" + sslkeycert.Name
				sniChild.SslKeyAndCertificateRefs = append(sniChild.SslKeyAndCertificateRefs, certName)
			}
		}

		var httpPolicyCollection []*avimodels.HTTPPolicies
		internalPolicyIndexBuffer := int32(11)
		for i, http := range vs_meta.HttpPolicyRefs {
			// Update them on the VS object
			var j int32
			j = int32(i) + internalPolicyIndexBuffer
			httpPolicy := fmt.Sprintf("/api/httppolicyset/?name=%s", http.Name)
			httpPolicies := &avimodels.HTTPPolicies{HTTPPolicySetRef: &httpPolicy, Index: &j}
			httpPolicyCollection = append(httpPolicyCollection, httpPolicies)
		}

		// from hostrule CRD
		bufferLen := int32(len(httpPolicyCollection)) + internalPolicyIndexBuffer + 5
		for i, policy := range vs_meta.HttpPolicySetRefs {
			var j int32
			j = int32(i) + bufferLen
			httpPolicy := policy
			httpPolicies := &avimodels.HTTPPolicies{HTTPPolicySetRef: &httpPolicy, Index: &j}
			httpPolicyCollection = append(httpPolicyCollection, httpPolicies)
		}

		sniChild.HTTPPolicies = httpPolicyCollection
	}

	var rest_ops []*utils.RestOp
	var rest_op utils.RestOp
	var path string
	if rest_method == utils.RestPut {

		path = "/api/virtualservice/" + cache_obj.Uuid
		rest_op = utils.RestOp{Path: path, Method: rest_method, Obj: sniChild,
			Tenant: vs_meta.Tenant, Model: "VirtualService", Version: utils.CtrlVersion}
		rest_ops = append(rest_ops, &rest_op)

	} else {

		macro := utils.AviRestObjMacro{ModelName: "VirtualService", Data: sniChild}
		path = "/api/macro"
		rest_op = utils.RestOp{Path: path, Method: rest_method, Obj: macro,
			Tenant: vs_meta.Tenant, Model: "VirtualService", Version: utils.CtrlVersion}
		rest_ops = append(rest_ops, &rest_op)

	}

	return rest_ops
}

func (rest *RestOperations) AviVsCacheAdd(rest_op *utils.RestOp, key string) error {
	if (rest_op.Err != nil) || (rest_op.Response == nil) {
		utils.AviLog.Warnf("key: %s, rest_op has err or no response for VS, err: %s, response: %s", key, rest_op.Err, rest_op.Response)
		return errors.New("Error rest_op")
	}

	resp_elems, ok := RestRespArrToObjByType(rest_op, "virtualservice", key)
	if ok != nil || resp_elems == nil {
		utils.AviLog.Warnf("key: %s, msg: unable to find vs obj in resp %v", key, rest_op.Response)
		return errors.New("vs not found")
	}

	for _, resp := range resp_elems {
		name, ok := resp["name"].(string)
		if !ok {
			utils.AviLog.Warnf("key: %s, msg: name not present in response %v", key, resp)
			return errors.New("Name not present in response")
		}

		uuid, ok := resp["uuid"].(string)
		if !ok {
			utils.AviLog.Warnf("key: %s, msg: Uuid not present in response %v", key, resp)
			return errors.New("Uuid not present in response")
		}

		cksum := resp["cloud_config_cksum"].(string)

		var lastModifiedStr string
		lastModifiedIntf, ok := resp["_last_modified"]
		if !ok {
			utils.AviLog.Warnf("key: %s, msg: last_modified not present in response %v", key, resp)
		} else {
			lastModifiedStr, ok = lastModifiedIntf.(string)
			if !ok {
				utils.AviLog.Warnf("key: %s, msg: last_modified is not of type string", key)
			}
		}

		vh_parent_uuid, found_parent := resp["vh_parent_vs_ref"]
		var parentVsObj *avicache.AviVsCache
		var vhParentKey interface{}
		if found_parent {
			// the uuid is expected to be in the format: "https://IP:PORT/api/virtualservice/virtualservice-88fd9718-f4f9-4e2b-9552-d31336330e0e#mygateway"
			vs_uuid := avicache.ExtractUuid(vh_parent_uuid.(string), "virtualservice-.*.#")
			utils.AviLog.Debugf("key: %s, msg: extracted the vs uuid from parent ref: %s", key, vs_uuid)
			// Now let's get the VS key from this uuid
			var foundvscache bool
			vhParentKey, foundvscache = rest.cache.VsCacheMeta.AviCacheGetKeyByUuid(vs_uuid)
			utils.AviLog.Infof("key: %s, msg: extracted the VS key from the uuid: %s", key, vhParentKey)
			if foundvscache {
				parentVsObj = rest.getVsCacheObj(vhParentKey.(avicache.NamespaceName), key)
				parentVsObj.AddToSNIChildCollection(uuid)
			} else {
				parentKey := avicache.NamespaceName{Namespace: rest_op.Tenant, Name: ExtractVsName(vh_parent_uuid.(string))}
				vs_cache_obj := rest.cache.VsCacheMeta.AviCacheAddVS(parentKey)
				vs_cache_obj.AddToSNIChildCollection(uuid)
				utils.AviLog.Info(spew.Sprintf("key: %s, msg: added VS cache key during SNI update %v val %v\n", key, vhParentKey,
					vs_cache_obj))
			}
		}

		k := avicache.NamespaceName{Namespace: rest_op.Tenant, Name: name}
		vs_cache, ok := rest.cache.VsCacheMeta.AviCacheGet(k)
		var svc_mdata_obj avicache.ServiceMetadataObj
		if resp["service_metadata"] != nil {
			utils.AviLog.Infof("key:%s, msg: Service Metadata: %s", key, resp["service_metadata"])
			if err := json.Unmarshal([]byte(resp["service_metadata"].(string)),
				&svc_mdata_obj); err != nil {
				utils.AviLog.Warnf("Error parsing service metadata :%v", err)
			}
		}

		if ok {
			vs_cache_obj, found := vs_cache.(*avicache.AviVsCache)
			if found {
				if _, ok := resp["vsvip_ref"].(string); ok {
					vsVipUuid := avicache.ExtractUuid(resp["vsvip_ref"].(string), "vsvip-.*.#")
					vsVipName, vipFound := rest.cache.VSVIPCache.AviCacheGetNameByUuid(vsVipUuid)
					if vipFound {
						vipKey := avicache.NamespaceName{Namespace: lib.GetTenant(), Name: vsVipName.(string)}
						vsvip_cache, found := rest.cache.VSVIPCache.AviCacheGet(vipKey)
						if found {
							vsvip_cache_obj, ok := vsvip_cache.(*avicache.AviVSVIPCache)
							if ok {
								if len(vsvip_cache_obj.Vips) > 0 {
									vip := vsvip_cache_obj.Vips[0]
									vs_cache_obj.Vip = vip
									utils.AviLog.Info(spew.Sprintf("key: %s, msg: updated vsvip to the cache: %s", key, vip))
								}
							}
						}
					}
				}
				vs_cache_obj.Uuid = uuid
				vs_cache_obj.CloudConfigCksum = cksum
				vs_cache_obj.ServiceMetadataObj = svc_mdata_obj
				if vhParentKey != nil {
					vs_cache_obj.ParentVSRef = vhParentKey.(avicache.NamespaceName)
				}

				vs_cache_obj.LastModified = lastModifiedStr
				if lastModifiedStr == "" {
					vs_cache_obj.InvalidData = true
				} else {
					vs_cache_obj.InvalidData = false
				}
				utils.AviLog.Debug(spew.Sprintf("key: %s, msg: updated VS cache key %v val %v\n", key, k,
					utils.Stringify(vs_cache_obj)))
				if svc_mdata_obj.Gateway != "" {
					if lib.UseServicesAPI() {
						status.UpdateSvcApiGatewayStatusAddress([]status.UpdateStatusOptions{{
							Vip:             vs_cache_obj.Vip,
							ServiceMetadata: svc_mdata_obj,
							Key:             key,
						}}, false)
					} else {
						status.UpdateGatewayStatusAddress([]status.UpdateStatusOptions{{
							Vip:             vs_cache_obj.Vip,
							ServiceMetadata: svc_mdata_obj,
							Key:             key,
						}}, false)
					}
				} else if len(svc_mdata_obj.NamespaceServiceName) > 0 {
					// This service needs an update of the status
					status.UpdateL4LBStatus([]status.UpdateStatusOptions{{
						Vip:             vs_cache_obj.Vip,
						ServiceMetadata: svc_mdata_obj,
						Key:             key,
					}}, false)
				} else if (svc_mdata_obj.IngressName != "" || len(svc_mdata_obj.NamespaceIngressName) > 0) && svc_mdata_obj.Namespace != "" && parentVsObj != nil {
					status.UpdateRouteIngressStatus([]status.UpdateStatusOptions{{
						Vip:             parentVsObj.Vip,
						ServiceMetadata: svc_mdata_obj,
						Key:             key,
					}}, false)
				}
				// This code is most likely hit when the first time a shard vs is created and the vs_cache_obj is populated from the pool update.
				// But before this a pool may have got created as a part of the macro operation, so update the ingress status here.
				if rest_op.Method == utils.RestPost || rest_op.Method == utils.RestDelete {
					for _, poolkey := range vs_cache_obj.PoolKeyCollection {
						// Fetch the pool object from cache and check the service metadata
						pool_cache, ok := rest.cache.PoolCache.AviCacheGet(poolkey)
						if ok {
							utils.AviLog.Infof("key: %s, msg: found pool: %s, will update status", key, poolkey.Name)
							pool_cache_obj, found := pool_cache.(*avicache.AviPoolCache)
							if found {
								if pool_cache_obj.ServiceMetadataObj.Namespace != "" {
									status.UpdateRouteIngressStatus([]status.UpdateStatusOptions{{
										Vip:             vs_cache_obj.Vip,
										ServiceMetadata: pool_cache_obj.ServiceMetadataObj,
										Key:             key,
									}}, false)
								}
							}
						}
					}
				}
			}
		} else {
			vs_cache_obj := avicache.AviVsCache{Name: name, Tenant: rest_op.Tenant,
				Uuid: uuid, CloudConfigCksum: cksum, ServiceMetadataObj: svc_mdata_obj,
				LastModified: lastModifiedStr,
			}
			if lastModifiedStr == "" {
				vs_cache_obj.InvalidData = true
			}
			if _, ok := resp["vsvip_ref"].(string); ok {
				vsVipUuid := avicache.ExtractUuid(resp["vsvip_ref"].(string), "vsvip-.*.#")
				vsVipName, vipFound := rest.cache.VSVIPCache.AviCacheGetNameByUuid(vsVipUuid)
				if vipFound {
					vipKey := avicache.NamespaceName{Namespace: lib.GetTenant(), Name: vsVipName.(string)}
					vsvip_cache, found := rest.cache.VSVIPCache.AviCacheGet(vipKey)
					if found {
						vsvip_cache_obj, ok := vsvip_cache.(*avicache.AviVSVIPCache)
						if ok {
							if len(vsvip_cache_obj.Vips) > 0 {
								vip := vsvip_cache_obj.Vips[0]
								vs_cache_obj.Vip = vip
								utils.AviLog.Info(spew.Sprintf("key: %s, msg: added vsvip to the cache: %s", key, vip))
							}
						}
					}
				}
			}

			if len(svc_mdata_obj.NamespaceServiceName) > 0 {
				// This service needs an update of the status
				status.UpdateL4LBStatus([]status.UpdateStatusOptions{{
					Vip:             vs_cache_obj.Vip,
					ServiceMetadata: svc_mdata_obj,
					Key:             key,
				}}, false)
			}
			rest.cache.VsCacheMeta.AviCacheAdd(k, &vs_cache_obj)
			utils.AviLog.Infof("key: %s, msg: added VS cache key %v val %v\n", key, k, utils.Stringify(&vs_cache_obj))
		}

	}

	return nil
}

func (rest *RestOperations) AviVsCacheDel(rest_op *utils.RestOp, vsKey avicache.NamespaceName, key string) error {
	// Delete the SNI Child ref
	vs_cache, ok := rest.cache.VsCacheMeta.AviCacheGet(vsKey)
	if ok {
		vs_cache_obj, found := vs_cache.(*avicache.AviVsCache)
		if found {
			hostFoundInParentPool := false
			parent_vs_cache, parent_ok := rest.cache.VsCacheMeta.AviCacheGet(vs_cache_obj.ParentVSRef)
			if parent_ok {
				parent_vs_cache_obj, parent_found := parent_vs_cache.(*avicache.AviVsCache)
				if parent_found {
					// Find the SNI child and then remove
					rest.findSNIRefAndRemove(vsKey, parent_vs_cache_obj, key)

					// if we find a L7Shared pool that has the secure VS host then don't delete status
					// update is also not required since the shard would not change, IP should remain same
					hostname := vs_cache_obj.ServiceMetadataObj.HostNames[0]
					hostFoundInParentPool = rest.isHostPresentInSharedPool(hostname, parent_vs_cache_obj, key)
				}
			}

			// try to delete the vsvip from cache only if the vs is not of type insecure passthrough
			// and if controller version is >= 20.1.1
			if vs_cache_obj.ServiceMetadataObj.PassthroughParentRef == "" {
				if lib.VSVipDelRequired() && len(vs_cache_obj.VSVipKeyCollection) > 0 {
					vsvip := vs_cache_obj.VSVipKeyCollection[0].Name
					vsvipKey := avicache.NamespaceName{Namespace: vsKey.Namespace, Name: vsvip}
					utils.AviLog.Infof("key: %s, msg: deleting vsvip cache for key: %s", key, vsvipKey)
					rest.cache.VSVIPCache.AviCacheDelete(vsvipKey)
				}
			}

			// Reset the LB status field as well.
			if vs_cache_obj.ServiceMetadataObj.Gateway != "" {
				status.DeleteGatewayStatusAddress(vs_cache_obj.ServiceMetadataObj, key)
				status.DeleteL4LBStatus(vs_cache_obj.ServiceMetadataObj, key)
			} else if len(vs_cache_obj.ServiceMetadataObj.NamespaceServiceName) > 0 {
				status.DeleteL4LBStatus(vs_cache_obj.ServiceMetadataObj, key)
			}

			if (vs_cache_obj.ServiceMetadataObj.IngressName != "" || len(vs_cache_obj.ServiceMetadataObj.NamespaceIngressName) > 0) && vs_cache_obj.ServiceMetadataObj.Namespace != "" {
				// SNI VS deletion related ingress status update
				if !hostFoundInParentPool {
					status.DeleteRouteIngressStatus(vs_cache_obj.ServiceMetadataObj, true, key)
				}
			} else {
				// Shared VS deletion related ingress status update
				for _, poolKey := range vs_cache_obj.PoolKeyCollection {
					rest.DeletePoolIngressStatus(poolKey, true, key)
				}
			}
		}
	}
	utils.AviLog.Infof("key: %s, msg: deleting vs cache for key: %s", key, vsKey)
	rest.cache.VsCacheMeta.AviCacheDelete(vsKey)

	return nil
}

func (rest *RestOperations) AviVSDel(uuid string, tenant string, key string) (*utils.RestOp, bool) {
	if uuid == "" {
		utils.AviLog.Warnf("key: %s, msg: empty uuid for VS, skipping delete", key)
		return nil, false
	}
	path := "/api/virtualservice/" + uuid
	rest_op := utils.RestOp{Path: path, Method: "DELETE",
		Tenant: tenant, Model: "VirtualService", Version: utils.CtrlVersion}
	utils.AviLog.Info(spew.Sprintf("key: %s, msg: VirtualService DELETE Restop %v \n",
		key, utils.Stringify(rest_op)))
	return &rest_op, true
}

func (rest *RestOperations) findSNIRefAndRemove(snichildkey avicache.NamespaceName, parentVsObj *avicache.AviVsCache, key string) {
	parentVsObj.VSCacheLock.Lock()
	defer parentVsObj.VSCacheLock.Unlock()
	for i, sni_uuid := range parentVsObj.SNIChildCollection {
		sni_vs_key, ok := rest.cache.VsCacheMeta.AviCacheGetKeyByUuid(sni_uuid)
		if ok {
			if sni_vs_key.(avicache.NamespaceName).Name == snichildkey.Name {
				parentVsObj.SNIChildCollection = append(parentVsObj.SNIChildCollection[:i], parentVsObj.SNIChildCollection[i+1:]...)
				utils.AviLog.Infof("key: %s, msg: removed sni key: %s", key, snichildkey.Name)
				break
			}
		}
	}
}

func (rest *RestOperations) isHostPresentInSharedPool(hostname string, parentVs *avicache.AviVsCache, key string) bool {
	for _, poolKey := range parentVs.PoolKeyCollection {
		if poolCache, found := rest.cache.PoolCache.AviCacheGet(poolKey); found {
			if pool, ok := poolCache.(*avicache.AviPoolCache); ok &&
				utils.HasElem(pool.ServiceMetadataObj.HostNames, hostname) {
				utils.AviLog.Debugf("key: %s, msg: hostname %v present in parentVS %s pool collection, will skip ingress status delete",
					key, parentVs.Name, hostname)
				return true
			}
		}
	}
	return false
}

func (rest *RestOperations) AviVsVipBuild(vsvip_meta *nodes.AviVSVIPNode, cache_obj *avicache.AviVSVIPCache, key string) (*utils.RestOp, error) {
	name := vsvip_meta.Name
	tenant := fmt.Sprintf("/api/tenant/?name=%s", vsvip_meta.Tenant)
	cloudRef := "/api/cloud?name=" + utils.CloudName
	var dns_info_arr []*avimodels.DNSInfo
	var path string
	var rest_op utils.RestOp
	var networkRef string
	vipId, ipType := "0", "V4"

	networkName := lib.GetNetworkName()
	if networkName != "" {
		networkRef = "/api/network/?name=" + networkName
	}
	subnetMask := lib.GetSubnetPrefixInt()
	subnetAddress := lib.GetSubnetIP()

	// all vsvip models would have auto_alloc set to true even in case of static IP programming
	autoAllocate := true

	if cache_obj != nil {
		vsvip, err := rest.AviVsVipGet(key, cache_obj.Uuid, name)
		if err != nil {
			return nil, err
		}
		for i, fqdn := range vsvip_meta.FQDNs {
			dns_info := avimodels.DNSInfo{Fqdn: &vsvip_meta.FQDNs[i]}
			foundFQDN := false
			// Verify this FQDN is already in the list or not.
			for _, dns := range dns_info_arr {
				if *dns.Fqdn == fqdn {
					foundFQDN = true
				}
			}
			if !foundFQDN {
				dns_info_arr = append(dns_info_arr, &dns_info)
			}
		}
		vsvip.DNSInfo = dns_info_arr

		// handling static IP updates, this would throw an error
		// for advl4 the error is propagated to the gateway status
		if vsvip_meta.IPAddress != "" {
			vip := &avimodels.Vip{
				VipID:          &vipId,
				AutoAllocateIP: &autoAllocate,
				IPAddress: &avimodels.IPAddr{
					Type: &ipType,
					Addr: &vsvip_meta.IPAddress,
				},
			}
			if networkName != "" {
				vip.IPAMNetworkSubnet = &avimodels.IPNetworkSubnet{
					NetworkRef: &networkRef,
				}
			}
			vsvip.Vip = []*avimodels.Vip{vip}
		}

		rest_op = utils.RestOp{
			Path:    "/api/vsvip/" + cache_obj.Uuid,
			Method:  utils.RestPut,
			Obj:     vsvip,
			Tenant:  vsvip_meta.Tenant,
			Model:   "VsVip",
			Version: utils.CtrlVersion,
		}
	} else {
		vip := avimodels.Vip{
			VipID:          &vipId,
			AutoAllocateIP: &autoAllocate,
		}

		// setting IPAMNetworkSubnet.Subnet value in case subnetCIDR is provided
		if lib.GetSubnetPrefix() == "" || subnetAddress == "" {
			utils.AviLog.Warnf("Incomplete values provided for subnetIP, will not use IPAMNetworkSubnet in vsvip")
		} else if lib.IsPublicCloud() && lib.GetCloudType() == lib.CLOUD_GCP {
			// add the IPAMNetworkSubnet
			vip.IPAMNetworkSubnet = &avimodels.IPNetworkSubnet{
				Subnet: &avimodels.IPAddrPrefix{
					IPAddr: &avimodels.IPAddr{Type: &ipType, Addr: &subnetAddress},
					Mask:   &subnetMask,
				},
			}
		} else if !lib.GetAdvancedL4() {
			vip.IPAMNetworkSubnet = &avimodels.IPNetworkSubnet{
				Subnet: &avimodels.IPAddrPrefix{
					IPAddr: &avimodels.IPAddr{Type: &ipType, Addr: &subnetAddress},
					Mask:   &subnetMask,
				},
			}
		}

		// configuring static IP, from gateway.Addresses (advl4) and service.loadBalancerIP (l4)
		if vsvip_meta.IPAddress != "" {
			vip.IPAddress = &avimodels.IPAddr{Type: &ipType, Addr: &vsvip_meta.IPAddress}
		}

		// selecting network with user input, in case user input is not provided AKO relies on
		// usable network configuration in ipamdnsproviderprofile
		if networkName != "" {
			if lib.IsPublicCloud() && lib.GetCloudType() != lib.CLOUD_GCP {
				vip.SubnetUUID = &networkName
			} else {
				// Set the IPAM network subnet for all clouds except AWS and Azure
				if vip.IPAMNetworkSubnet == nil {
					// initialize if not initialized earlier
					vip.IPAMNetworkSubnet = &avimodels.IPNetworkSubnet{}
				}
				vip.IPAMNetworkSubnet.NetworkRef = &networkRef
			}

		}

		addr := "172.18.0.0"
		ew_subnet := avimodels.IPAddrPrefix{
			IPAddr: &avimodels.IPAddr{Type: &ipType, Addr: &addr},
			Mask:   &subnetMask,
		}

		var east_west bool
		if vsvip_meta.EastWest == true {
			vip.Subnet = &ew_subnet
			east_west = true
		} else {
			east_west = false
		}

		for i, fqdn := range vsvip_meta.FQDNs {
			dns_info := avimodels.DNSInfo{Fqdn: &vsvip_meta.FQDNs[i]}
			foundFQDN := false
			// Verify this FQDN is already in the list or not.
			for _, dns := range dns_info_arr {
				if *dns.Fqdn == fqdn {
					foundFQDN = true
				}
			}
			if !foundFQDN {
				dns_info_arr = append(dns_info_arr, &dns_info)
			}
		}

		vrfContextRef := "/api/vrfcontext?name=" + vsvip_meta.VrfContext
		vsvip := avimodels.VsVip{
			Name:              &name,
			TenantRef:         &tenant,
			CloudRef:          &cloudRef,
			EastWestPlacement: &east_west,
			VrfContextRef:     &vrfContextRef,
			DNSInfo:           dns_info_arr,
			Vip:               []*avimodels.Vip{&vip},
		}

		macro := utils.AviRestObjMacro{ModelName: "VsVip", Data: vsvip}
		path = "/api/macro"
		// Patch an existing vsvip if it exists in the cache but not associated with this VS.
		vsvip_key := avicache.NamespaceName{Namespace: vsvip_meta.Tenant, Name: name}
		utils.AviLog.Debugf("key: %s, searching in cache for vsVip Key: %s", key, vsvip_key)
		vsvip_cache, ok := rest.cache.VSVIPCache.AviCacheGet(vsvip_key)
		if ok {
			vsvip_cache_obj, _ := vsvip_cache.(*avicache.AviVSVIPCache)
			vsvip_avi, err := rest.AviVsVipGet(key, vsvip_cache_obj.Uuid, name)
			if err != nil {
				if strings.Contains(err.Error(), VSVIP_NOTFOUND) {
					// Clear the cache for this key
					rest.cache.VSVIPCache.AviCacheDelete(vsvip_key)
					utils.AviLog.Warnf("key: %s, Removed the vsvip object from the cache", key)
					rest_op = utils.RestOp{Path: path, Method: utils.RestPost, Obj: macro,
						Tenant: vsvip_meta.Tenant, Model: "VsVip", Version: utils.CtrlVersion}
					return &rest_op, nil
				}
				// If it's not nil, return an error.
				utils.AviLog.Warnf("key: %s, Error in vsvip GET operation: %s", key, err)
				return nil, err
			}
			for i, fqdn := range vsvip_meta.FQDNs {
				dns_info := avimodels.DNSInfo{Fqdn: &vsvip_meta.FQDNs[i]}
				foundFQDN := false
				// Verify this FQDN is already in the list or not.
				for _, dns := range dns_info_arr {
					if *dns.Fqdn == fqdn {
						foundFQDN = true
					}
				}
				if !foundFQDN {
					dns_info_arr = append(dns_info_arr, &dns_info)
				}
			}
			vsvip_avi.DNSInfo = dns_info_arr
			vsvip_avi.VrfContextRef = &vrfContextRef

			path = "/api/vsvip/" + vsvip_cache_obj.Uuid
			rest_op = utils.RestOp{Path: path,
				Method:  utils.RestPut,
				Obj:     vsvip_avi,
				Tenant:  vsvip_meta.Tenant,
				Model:   "VsVip",
				Version: utils.CtrlVersion,
			}
		} else {
			rest_op = utils.RestOp{Path: path,
				Method:  utils.RestPost,
				Obj:     macro,
				Tenant:  vsvip_meta.Tenant,
				Model:   "VsVip",
				Version: utils.CtrlVersion,
			}
		}
	}

	return &rest_op, nil
}

func (rest *RestOperations) AviVsVipGet(key, uuid, name string) (*avimodels.VsVip, error) {
	if rest.aviRestPoolClient == nil {
		utils.AviLog.Warnf("key: %s, msg: aviRestPoolClient during vsvip not initialized\n", key)
		return nil, errors.New("client in aviRestPoolClient during vsvip not initialized")
	}
	if len(rest.aviRestPoolClient.AviClient) < 1 {
		utils.AviLog.Warnf("key: %s, msg: client in aviRestPoolClient during vsvip not initialized\n", key)
		return nil, errors.New("client in aviRestPoolClient during vsvip not initialized")
	}
	client := rest.aviRestPoolClient.AviClient[0]
	uri := "/api/vsvip/" + uuid

	rawData, err := client.AviSession.GetRaw(uri)
	if err != nil {
		utils.AviLog.Warnf("VsVip Get uri %v returned err %v", uri, err)
		webSyncErr := &utils.WebSyncError{
			Err: err, Operation: string(utils.RestGet),
		}
		return nil, webSyncErr
	}
	vsvip := avimodels.VsVip{}
	json.Unmarshal(rawData, &vsvip)

	return &vsvip, nil
}

func (rest *RestOperations) AviVsVipDel(uuid string, tenant string, key string) *utils.RestOp {
	path := "/api/vsvip/" + uuid
	rest_op := utils.RestOp{Path: path, Method: "DELETE",
		Tenant: tenant, Model: "VsVip", Version: utils.CtrlVersion}
	utils.AviLog.Info(spew.Sprintf("key: %s, msg: VSVIP DELETE Restop %v \n", key,
		utils.Stringify(rest_op)))
	return &rest_op
}

func (rest *RestOperations) AviVsVipCacheAdd(rest_op *utils.RestOp, vsKey avicache.NamespaceName, key string) error {
	if (rest_op.Err != nil) || (rest_op.Response == nil) {
		if rest_op.Message == "" {
			utils.AviLog.Warnf("key: %s, rest_op has err or no response for vsvip err: %s, response: %s", key, rest_op.Err, rest_op.Response)
			return errors.New("Errored vsvip rest_op")
		}

		vs_cache, ok := rest.cache.VsCacheMeta.AviCacheGet(vsKey)
		if ok {
			vs_cache_obj, found := vs_cache.(*avicache.AviVsCache)
			if found && vs_cache_obj.ServiceMetadataObj.Gateway != "" {
				gwNSName := strings.Split(vs_cache_obj.ServiceMetadataObj.Gateway, "/")
				if lib.GetAdvancedL4() {
					gw, err := lib.GetAdvL4Informers().GatewayInformer.Lister().Gateways(gwNSName[0]).Get(gwNSName[1])
					if err != nil {
						utils.AviLog.Warnf("key: %s, msg: Gateway object not found, skippig status update %v", key, err)
						return err
					}

					gwStatus := gw.Status.DeepCopy()
					status.UpdateGatewayStatusGWCondition(key, gwStatus, &status.UpdateGWStatusConditionOptions{
						Type:    "Pending",
						Status:  corev1.ConditionTrue,
						Reason:  "InvalidAddress",
						Message: rest_op.Message,
					})
					status.UpdateGatewayStatusObject(key, gw, gwStatus)

				} else if lib.UseServicesAPI() {
					gw, err := lib.GetSvcAPIInformers().GatewayInformer.Lister().Gateways(gwNSName[0]).Get(gwNSName[1])
					if err != nil {
						utils.AviLog.Warnf("key: %s, msg: Gateway object not found, skippig status update %v", key, err)
						return err
					}

					gwStatus := gw.Status.DeepCopy()
					status.UpdateSvcApiGatewayStatusGWCondition(key, gwStatus, &status.UpdateSvcApiGWStatusConditionOptions{
						Type:    "Pending",
						Status:  metav1.ConditionTrue,
						Reason:  "InvalidAddress",
						Message: rest_op.Message,
					})
					status.UpdateSvcApiGatewayStatusObject(key, gw, gwStatus)
				}
				utils.AviLog.Warnf("key: %s, msg: IPAddress Updates on gateway not supported, Please recreate gateway object with the new preferred IPAddress", key)
				return errors.New(rest_op.Message)
			}

		}
	}

	resp_elems, ok := RestRespArrToObjByType(rest_op, "vsvip", key)
	if ok != nil || resp_elems == nil {
		utils.AviLog.Warnf("key: %s, msg: unable to find vsvip obj in resp %v", key, rest_op.Response)
		return errors.New("vsvip not found")
	}

	for _, resp := range resp_elems {
		name, ok := resp["name"].(string)
		if !ok {
			utils.AviLog.Warnf("key: %s, msg: vsvip name not present in response %v", key, resp)
			continue
		}

		uuid, ok := resp["uuid"].(string)
		if !ok {
			utils.AviLog.Warnf("key: %s, msg: vsvip Uuid not present in response %v", key, resp)
			continue
		}

		var lastModifiedStr string
		lastModifiedIntf, ok := resp["_last_modified"]
		if !ok {
			utils.AviLog.Warnf("key: %s, msg: last_modified not present in response %v", key, resp)
		} else {
			lastModifiedStr, ok = lastModifiedIntf.(string)
			if !ok {
				utils.AviLog.Warnf("key: %s, msg: last_modified is not of type string", key)
			}
		}

		var vsvipFQDNs []string
		if _, found := resp["dns_info"]; found {
			if allDNSInfo, ok := resp["dns_info"].([]interface{}); ok {
				for _, dnsInfoIntf := range allDNSInfo {
					dnsinfo, valid := dnsInfoIntf.(map[string]interface{})
					if !valid {
						utils.AviLog.Infof("key: %s, msg: invalid type for dns_info in vsvip: %s", key, name)
						continue
					}
					fqdnIntf, valid := dnsinfo["fqdn"]
					if !valid {
						utils.AviLog.Infof("key: %s, msg: fqdn not found for dns_info in vsvip: %s", key, name)
						continue
					}
					fqdn, valid := fqdnIntf.(string)
					if valid {
						vsvipFQDNs = append(vsvipFQDNs, fqdn)
					}
				}
			}
		}

		var vsvipVips []string
		if _, found := resp["vip"]; found {
			if vips, ok := resp["vip"].([]interface{}); ok {
				for _, vipsIntf := range vips {
					vip, valid := vipsIntf.(map[string]interface{})
					if !valid {
						utils.AviLog.Infof("key: %s, msg: invalid type for vip in vsvip: %s", key, name)
						continue
					}
					ip_address, valid := vip["ip_address"].(map[string]interface{})
					if !valid {
						utils.AviLog.Infof("key: %s, msg: invalid type for ip_address in vsvip: %s", key, name)
						continue
					}
					addr, valid := ip_address["addr"].(string)
					if !valid {
						utils.AviLog.Infof("key: %s, msg: invalid type for addr in vsvip: %s", key, name)
						continue
					}
					vsvipVips = append(vsvipVips, addr)
				}
			}
		}

		vsvip_cache_obj := avicache.AviVSVIPCache{Name: name, Tenant: rest_op.Tenant,
			Uuid: uuid, LastModified: lastModifiedStr, FQDNs: vsvipFQDNs, Vips: vsvipVips}
		if lastModifiedStr == "" {
			vsvip_cache_obj.InvalidData = true
		}

		k := avicache.NamespaceName{Namespace: rest_op.Tenant, Name: name}
		rest.cache.VSVIPCache.AviCacheAdd(k, &vsvip_cache_obj)
		// Update the VS object
		vs_cache, ok := rest.cache.VsCacheMeta.AviCacheGet(vsKey)
		if ok {
			vs_cache_obj, found := vs_cache.(*avicache.AviVsCache)
			if found {
				vs_cache_obj.AddToVSVipKeyCollection(k)
				utils.AviLog.Debugf("key: %s, msg: modified the VS cache object for VSVIP collection. The cache now is :%v", key, utils.Stringify(vs_cache_obj))
			}

		} else {
			vs_cache_obj := rest.cache.VsCacheMeta.AviCacheAddVS(vsKey)
			vs_cache_obj.AddToVSVipKeyCollection(k)
			utils.AviLog.Info(spew.Sprintf("key: %s, msg: added VS cache key during vsvip update %v val %v\n", key, vsKey,
				vs_cache_obj))
		}
		utils.AviLog.Info(spew.Sprintf("key: %s, msg: added vsvip cache k %v val %v\n", key, k,
			vsvip_cache_obj))
	}

	return nil
}

func (rest *RestOperations) AviVsVipCacheDel(rest_op *utils.RestOp, vsKey avicache.NamespaceName, key string) error {
	vsvipkey := avicache.NamespaceName{Namespace: rest_op.Tenant, Name: rest_op.ObjName}
	rest.cache.VSVIPCache.AviCacheDelete(vsvipkey)
	if vsKey != (avicache.NamespaceName{}) {
		vs_cache, ok := rest.cache.VsCacheMeta.AviCacheGet(vsKey)
		if ok {
			vs_cache_obj, found := vs_cache.(*avicache.AviVsCache)
			if found {
				vs_cache_obj.RemoveFromVSVipKeyCollection(vsvipkey)
			}
		}
	}

	return nil

}
