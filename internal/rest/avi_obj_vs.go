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

	"github.com/davecgh/go-spew/spew"
	avimodels "github.com/vmware/alb-sdk/go/models"
)

const VSVIP_NOTFOUND = "VsVip object not found"

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

		var enableRhi bool
		if vs_meta.EnableRhi != nil {
			enableRhi = *vs_meta.EnableRhi
		} else {
			enableRhi = lib.GetEnableRHI()
		}
		vs.EnableRhi = &enableRhi

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
		if lib.GetGRBACSupport() {
			vs.Labels = lib.GetLabels()
		}
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
				Tenant: vs_meta.Tenant, Model: "VirtualService", Version: utils.CtrlVersion, ObjName: *vs.Name}
			rest_ops = append(rest_ops, &rest_op)
		} else {
			path = "/api/virtualservice/"
			rest_op = utils.RestOp{Path: path, Method: utils.RestPost, Obj: vs,
				Tenant: vs_meta.Tenant, Model: "VirtualService", Version: utils.CtrlVersion, ObjName: *vs.Name}
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
	if lib.GetGRBACSupport() {
		sniChild.Labels = lib.GetLabels()
	}
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
		path = "/api/virtualservice"
		rest_op = utils.RestOp{Path: path, Method: rest_method, Obj: sniChild,
			Tenant: vs_meta.Tenant, Model: "VirtualService", Version: utils.CtrlVersion}
		rest_ops = append(rest_ops, &rest_op)
	}

	return rest_ops
}

func (rest *RestOperations) AviVsCacheAdd(rest_op *utils.RestOp, key string) error {
	if (rest_op.Err != nil) || (rest_op.Response == nil) {
		utils.AviLog.Warnf("key: %s, rest_op has err or no response for VS, err: %v, response: %v", key, rest_op.Err, rest_op.Response)
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
					updateOptions := status.UpdateOptions{
						Vip:             vs_cache_obj.Vip,
						ServiceMetadata: svc_mdata_obj,
						Key:             key,
					}
					statusOption := status.StatusOptions{
						ObjType: lib.Gateway,
						Op:      lib.UpdateStatus,
						Options: &updateOptions,
					}
					if lib.UseServicesAPI() {
						statusOption.ObjType = lib.SERVICES_API
					}
					status.PublishToStatusQueue(updateOptions.ServiceMetadata.Gateway, statusOption)

				} else if len(svc_mdata_obj.NamespaceServiceName) > 0 {
					// This service needs an update of the status
					updateOptions := status.UpdateOptions{
						Vip:                vs_cache_obj.Vip,
						ServiceMetadata:    svc_mdata_obj,
						Key:                key,
						VirtualServiceUUID: vs_cache_obj.Uuid,
					}
					statusOption := status.StatusOptions{
						ObjType: utils.L4LBService,
						Op:      lib.UpdateStatus,
						Options: &updateOptions,
					}
					status.PublishToStatusQueue(svc_mdata_obj.NamespaceServiceName[0], statusOption)
				} else if (svc_mdata_obj.IngressName != "" || len(svc_mdata_obj.NamespaceIngressName) > 0) && svc_mdata_obj.Namespace != "" && parentVsObj != nil {
					updateOptions := status.UpdateOptions{
						Vip:                parentVsObj.Vip,
						ServiceMetadata:    svc_mdata_obj,
						Key:                key,
						VirtualServiceUUID: vs_cache_obj.Uuid,
					}
					statusOption := status.StatusOptions{
						ObjType: utils.Ingress,
						Op:      lib.UpdateStatus,
						Options: &updateOptions,
					}
					if utils.GetInformers().RouteInformer != nil {
						statusOption.ObjType = utils.OshiftRoute
					}
					status.PublishToStatusQueue(updateOptions.ServiceMetadata.IngressName, statusOption)
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
									updateOptions := status.UpdateOptions{
										Vip:                vs_cache_obj.Vip,
										ServiceMetadata:    pool_cache_obj.ServiceMetadataObj,
										Key:                key,
										VirtualServiceUUID: vs_cache_obj.Uuid,
									}
									statusOption := status.StatusOptions{
										ObjType: utils.Ingress,
										Op:      lib.UpdateStatus,
										Options: &updateOptions,
									}
									if utils.GetInformers().RouteInformer != nil {
										statusOption.ObjType = utils.OshiftRoute
									}
									status.PublishToStatusQueue(updateOptions.ServiceMetadata.IngressName, statusOption)
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
				updateOptions := status.UpdateOptions{
					ServiceMetadata: vs_cache_obj.ServiceMetadataObj,
					Key:             key,
				}
				statusOption := status.StatusOptions{
					ObjType: utils.L4LBService,
					Op:      lib.DeleteStatus,
					Options: &updateOptions,
				}
				status.PublishToStatusQueue(svc_mdata_obj.NamespaceServiceName[0], statusOption)
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
					if len(vs_cache_obj.ServiceMetadataObj.HostNames) > 0 {
						hostname := vs_cache_obj.ServiceMetadataObj.HostNames[0]
						hostFoundInParentPool = rest.isHostPresentInSharedPool(hostname, parent_vs_cache_obj, key)
					}

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
				updateOptions := status.UpdateOptions{
					ServiceMetadata: vs_cache_obj.ServiceMetadataObj,
					Key:             key,
				}
				statusOption := status.StatusOptions{
					ObjType: lib.Gateway,
					Op:      lib.DeleteStatus,
					Options: &updateOptions,
				}
				status.PublishToStatusQueue(updateOptions.ServiceMetadata.Gateway, statusOption)

				statusOptionLBSvc := statusOption
				statusOptionLBSvc.ObjType = utils.L4LBService
				status.PublishToStatusQueue(updateOptions.ServiceMetadata.Gateway, statusOptionLBSvc)

			} else if len(vs_cache_obj.ServiceMetadataObj.NamespaceServiceName) > 0 {
				status.DeleteL4LBStatus(vs_cache_obj.ServiceMetadataObj, key)
			}

			if (vs_cache_obj.ServiceMetadataObj.IngressName != "" || len(vs_cache_obj.ServiceMetadataObj.NamespaceIngressName) > 0) && vs_cache_obj.ServiceMetadataObj.Namespace != "" {
				// SNI VS deletion related ingress status update
				if !hostFoundInParentPool {
					updateOptions := status.UpdateOptions{
						ServiceMetadata: vs_cache_obj.ServiceMetadataObj,
						Key:             key,
					}
					statusOption := status.StatusOptions{
						ObjType: utils.Ingress,
						Op:      lib.DeleteStatus,
						IsVSDel: true,
						Options: &updateOptions,
					}
					if utils.GetInformers().RouteInformer != nil {
						statusOption.ObjType = utils.OshiftRoute
					}
					status.PublishToStatusQueue(updateOptions.ServiceMetadata.IngressName, statusOption)
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
