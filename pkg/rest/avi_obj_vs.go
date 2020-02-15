/*
 * [2013] - [2018] Avi Networks Incorporated
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

	avimodels "github.com/avinetworks/sdk/go/models"
	"github.com/davecgh/go-spew/spew"
	avicache "gitlab.eng.vmware.com/orion/akc/pkg/cache"
	"gitlab.eng.vmware.com/orion/akc/pkg/nodes"
	"gitlab.eng.vmware.com/orion/container-lib/utils"
)

func FindPoolGroupForPort(pgList []*nodes.AviPoolGroupNode, portToSearch int32) string {
	for _, pg := range pgList {
		if pg.Port == fmt.Sprint(portToSearch) {
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
		checksumstr := fmt.Sprint(cksum)
		cr := utils.OSHIFT_K8S_CLOUD_CONNECTOR
		cloudRef := "/api/cloud?name=" + utils.CloudName
		svc_mdata_json, _ := json.Marshal(&vs_meta.ServiceMetadata)
		svc_mdata := string(svc_mdata_json)
		vrfContextRef := "/api/vrfcontext?name=" + vs_meta.VrfContext
		vs := avimodels.VirtualService{Name: &name,
			NetworkProfileRef:     &network_prof,
			ApplicationProfileRef: &app_prof,
			CloudConfigCksum:      &checksumstr,
			CreatedBy:             &cr,
			CloudRef:              &cloudRef,
			ServiceMetadata:       &svc_mdata,
			VrfContextRef:         &vrfContextRef,
		}

		if vs_meta.DefaultPoolGroup != "" {
			pool_ref := "/api/poolgroup/?name=" + vs_meta.DefaultPoolGroup
			vs.PoolGroupRef = &pool_ref
		}
		if len(vs_meta.VSVIPRefs) > 0 {
			vipref := "/api/vsvip/?name=" + vs_meta.VSVIPRefs[0].Name
			vs.VsvipRef = &vipref
		} else {
			utils.AviLog.Warning.Printf("key: %s, msg: unable to set the vsvip reference")
		}
		tenant := fmt.Sprintf("/api/tenant/?name=%s", vs_meta.Tenant)
		vs.TenantRef = &tenant

		if vs_meta.SNIParent {
			// This is a SNI parent
			utils.AviLog.Info.Printf("key: %s, msg: vs %s is a SNI Parent", key, vs_meta.Name)
			vh_parent := utils.VS_TYPE_VH_PARENT
			vs.Type = &vh_parent
		}
		// TODO other fields like cloud_ref, mix of TCP & UDP protocols, etc.

		for i, pp := range vs_meta.PortProto {
			port := pp.Port
			svc := avimodels.Service{Port: &port, EnableSsl: &vs_meta.PortProto[i].EnableSSL}
			if pp.Protocol == utils.TCP {
				utils.AviLog.Info.Printf("key: %s, msg: processing TCP ports for VS creation :%v", key, pp.Port)
				port := pp.Port
				var sproto string
				sproto = "PROTOCOL_TYPE_TCP_PROXY"
				pg_name := FindPoolGroupForPort(vs_meta.TCPPoolGroupRefs, port)
				if pg_name != "" {
					utils.AviLog.Info.Printf("key: %s, msg: TCP ports for VS creation returned PG: %s", key, pg_name)
					pg_ref := "/api/poolgroup/?name=" + pg_name
					sps := avimodels.ServicePoolSelector{ServicePoolGroupRef: &pg_ref,
						ServicePort: &port, ServiceProtocol: &sproto}
					vs.ServicePoolSelect = append(vs.ServicePoolSelect, &sps)
				} else {
					utils.AviLog.Info.Printf("key: %s, msg: TCP ports for VS creation returned no matching PGs", key)
				}

			} else if pp.Protocol == utils.UDP && vs_meta.NetworkProfile == "System-TCP-Proxy" {
				onw_profile := "/api/networkprofile/?name=System-UDP-Fast-Path"
				svc.OverrideNetworkProfileRef = &onw_profile
			}
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

		var rest_ops []*utils.RestOp

		var rest_op utils.RestOp
		var path string
		if rest_method == utils.RestPut {
			path = "/api/virtualservice/" + cache_obj.Uuid
			rest_op = utils.RestOp{Path: path, Method: rest_method, Obj: vs,
				Tenant: vs_meta.Tenant, Model: "VirtualService", Version: utils.CtrlVersion}
			rest_ops = append(rest_ops, &rest_op)

		} else {
			macro := utils.AviRestObjMacro{ModelName: "VirtualService", Data: vs}
			path = "/api/macro"
			rest_op = utils.RestOp{Path: path, Method: rest_method, Obj: macro,
				Tenant: vs_meta.Tenant, Model: "VirtualService", Version: utils.CtrlVersion}
			rest_ops = append(rest_ops, &rest_op)

		}

		utils.AviLog.Info.Print(spew.Sprintf("key: %s, msg: VS Restop %v K8sAviVsMeta %v\n", key, utils.Stringify(rest_op),
			*vs_meta))
		return rest_ops
	}
}

func (rest *RestOperations) AviVsSniBuild(vs_meta *nodes.AviVsNode, rest_method utils.RestMethod, cache_obj *avicache.AviVsCache, key string) []*utils.RestOp {
	name := vs_meta.Name
	cksum := vs_meta.CloudConfigCksum
	checksumstr := fmt.Sprint(cksum)
	cr := utils.OSHIFT_K8S_CLOUD_CONNECTOR

	east_west := false
	var app_prof string
	if vs_meta.TLSType != utils.TLS_PASSTHROUGH {
		app_prof = "/api/applicationprofile/?name=" + utils.DEFAULT_L7_APP_PROFILE
	} else {
		app_prof = "/api/applicationprofile/?name=" + utils.DEFAULT_L7_SECURE_APP_PROFILE
	}
	cloudRef := "/api/cloud?name=" + utils.CloudName
	network_prof := "/api/networkprofile/?name=" + "System-TCP-Proxy"
	vrfContextRef := "/api/vrfcontext?name=" + vs_meta.VrfContext
	sniChild := &avimodels.VirtualService{Name: &name, CloudConfigCksum: &checksumstr,
		CreatedBy:             &cr,
		NetworkProfileRef:     &network_prof,
		ApplicationProfileRef: &app_prof,
		EastWestPlacement:     &east_west,
		CloudRef:              &cloudRef,
		VrfContextRef:         &vrfContextRef,
	}

	//This VS has a TLSKeyCert associated, we need to mark 'type': 'VS_TYPE_VH_PARENT'
	vh_type := utils.VS_TYPE_VH_CHILD
	sniChild.Type = &vh_type
	vhParentUuid := "/api/virtualservice/?name=" + vs_meta.VHParentName
	sniChild.VhParentVsUUID = &vhParentUuid
	sniChild.VhDomainName = vs_meta.VHDomainNames
	ignPool := true
	sniChild.IgnPoolNetReach = &ignPool

	if vs_meta.DefaultPool != "" {
		pool_ref := "/api/pool/?name=" + vs_meta.DefaultPool
		sniChild.PoolRef = &pool_ref
	}
	var rest_ops []*utils.RestOp
	// No need of HTTP rules for TLS passthrough.
	if vs_meta.TLSType != utils.TLS_PASSTHROUGH {
		for _, sslkeycert := range vs_meta.SniNodes {
			certName := "/api/sslkeyandcertificate/?name=" + sslkeycert.Name
			sniChild.SslKeyAndCertificateRefs = append(sniChild.SslKeyAndCertificateRefs, certName)
		}
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
		sniChild.HTTPPolicies = httpPolicyCollection
	}
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
		utils.AviLog.Warning.Printf("key: %s, rest_op has err or no reponse for VS, err: %s, response: %s", key, rest_op.Err, rest_op.Response)
		return errors.New("Errored rest_op")
	}

	resp_elems, ok := RestRespArrToObjByType(rest_op, "virtualservice", key)
	if ok != nil || resp_elems == nil {
		utils.AviLog.Warning.Printf("key: %s, msg: unable to find vs obj in resp %v", key, rest_op.Response)
		return errors.New("vs not found")
	}

	for _, resp := range resp_elems {
		name, ok := resp["name"].(string)
		if !ok {
			utils.AviLog.Warning.Printf("key: %s, msg: name not present in response %v", key, resp)
			return errors.New("Name not present in response")
		}

		uuid, ok := resp["uuid"].(string)
		if !ok {
			utils.AviLog.Warning.Printf("key: %s, msg: Uuid not present in response %v", key, resp)
			return errors.New("Uuid not present in response")
		}

		cksum := resp["cloud_config_cksum"].(string)
		utils.AviLog.Info.Printf("key: %s, msg: vs information %s", key, utils.Stringify(resp))

		vh_parent_uuid, found_parent := resp["vh_parent_vs_ref"]
		if found_parent {
			// the uuid is expected to be in the format: "https://IP:PORT/api/virtualservice/virtualservice-88fd9718-f4f9-4e2b-9552-d31336330e0e#mygateway"
			vs_uuid := avicache.ExtractVsUuid(vh_parent_uuid.(string))
			utils.AviLog.Info.Printf("key: %s, msg: extracted the vs uuid from parent ref: %s", key, vs_uuid)
			// Now let's get the VS key from this uuid
			vsKey, foundvscache := rest.cache.VsCache.AviCacheGetKeyByUuid(vs_uuid)
			utils.AviLog.Info.Printf("key: %s, msg: extracted the VS key from the uuid :%s", key, vsKey)
			if foundvscache {
				vs_obj := rest.getVsCacheObj(vsKey.(avicache.NamespaceName), key)
				if !utils.HasElem(vs_obj.SNIChildCollection, uuid) {
					vs_obj.SNIChildCollection = append(vs_obj.SNIChildCollection, uuid)
				}
			} else {
				vs_cache_obj := avicache.AviVsCache{Name: ExtractVsName(vh_parent_uuid.(string)), Tenant: rest_op.Tenant,
					SNIChildCollection: []string{uuid}}
				rest.cache.VsCache.AviCacheAdd(vsKey, &vs_cache_obj)
				utils.AviLog.Info.Print(spew.Sprintf("key: %s, msg: added VS cache key during SNI update %v val %v\n", key, vsKey,
					vs_cache_obj))
			}
		}
		vsvip, vipExists := resp["vip"].([]interface{})
		k := avicache.NamespaceName{Namespace: rest_op.Tenant, Name: name}
		vs_cache, ok := rest.cache.VsCache.AviCacheGet(k)
		var svc_mdata interface{}
		var svc_mdata_map map[string]interface{}
		var svc_mdata_obj avicache.LBServiceMetadataObj
		if resp["service_metadata"] != nil {
			if err := json.Unmarshal([]byte(resp["service_metadata"].(string)),
				&svc_mdata); err == nil {
				var svcOk bool
				svc_mdata_map, svcOk = svc_mdata.(map[string]interface{})
				if !svcOk {
					utils.AviLog.Warning.Printf("resp %v svc_mdata %T has invalid service_metadata type", resp, svc_mdata)
				} else {
					LBSvcMdataMapToObj(&svc_mdata_map, &svc_mdata_obj)
				}
			}
		}
		if ok {
			vs_cache_obj, found := vs_cache.(*avicache.AviVsCache)
			if found {
				if vipExists && len(vsvip) > 0 {
					vip := (resp["vip"].([]interface{})[0].(map[string]interface{})["ip_address"]).(map[string]interface{})["addr"].(string)
					vs_cache_obj.Vip = vip
					utils.AviLog.Info.Print(spew.Sprintf("key: %s, msg: updated vsvip to the cache: %s", key, vip))
				}
				vs_cache_obj.Uuid = uuid
				vs_cache_obj.CloudConfigCksum = cksum
				vs_cache_obj.ServiceMetadataObj = svc_mdata_obj
				utils.AviLog.Info.Print(spew.Sprintf("key: %s, msg: updated VS cache key %v val %v\n", key, k,
					utils.Stringify(vs_cache_obj)))
				if svc_mdata_obj.ServiceName != "" && svc_mdata_obj.Namespace != "" {
					// This service needs an update of the status
					UpdateL4LBStatus(vs_cache_obj, svc_mdata_obj, key)
				}
				// This code is most likely hit when the first time a shard vs is created and the vs_cache_obj is populated from the pool update.
				// But before this a pool may have got created as a part of the macro operation, so update the ingress status here.
				for _, poolkey := range vs_cache_obj.PoolKeyCollection {
					// Fetch the pool object from cache and check the service metadata
					pool_cache, ok := rest.cache.PoolCache.AviCacheGet(poolkey)
					if ok {
						utils.AviLog.Info.Printf("key :%s, msg: found pool :%s, will update status", key, poolkey.Name)
						pool_cache_obj, found := pool_cache.(*avicache.AviPoolCache)
						if found {
							if pool_cache_obj.ServiceMetadataObj.Namespace != "" {
								UpdateIngressStatus(vs_cache_obj, pool_cache_obj.ServiceMetadataObj, key)
							}
						}
					}
				}
			}
		} else {
			vs_cache_obj := avicache.AviVsCache{Name: name, Tenant: rest_op.Tenant,
				Uuid: uuid, CloudConfigCksum: cksum, ServiceMetadataObj: svc_mdata_obj}
			if vipExists && len(vsvip) > 0 {
				vip := (resp["vip"].([]interface{})[0].(map[string]interface{})["ip_address"]).(map[string]interface{})["addr"].(string)
				vs_cache_obj.Vip = vip
				utils.AviLog.Info.Print(spew.Sprintf("key: %s, msg: added vsvip to the cache: %s", key, vip))
			}
			if svc_mdata_obj.ServiceName != "" && svc_mdata_obj.Namespace != "" {
				// This service needs an update of the status
				UpdateL4LBStatus(&vs_cache_obj, svc_mdata_obj, key)
			}
			rest.cache.VsCache.AviCacheAdd(k, &vs_cache_obj)
			utils.AviLog.Info.Print(spew.Sprintf("key: %s, msg: added VS cache key %v val %v\n", key, k,
				vs_cache_obj))
		}

	}

	return nil
}

func LBSvcMdataMapToObj(svc_mdata_map *map[string]interface{}, svc_mdata *avicache.LBServiceMetadataObj) {
	for k, val := range *svc_mdata_map {
		switch k {
		case "svc_name":
			svcName, ok := val.(string)
			if ok {
				svc_mdata.ServiceName = svcName
			} else {
				utils.AviLog.Warning.Printf("Incorrect type %T in lb_svc_mdata_map %v", val, *svc_mdata_map)
			}
		case "namespace":
			namespace, ok := val.(string)
			if ok {
				svc_mdata.Namespace = namespace
			} else {
				utils.AviLog.Warning.Printf("Incorrect type %T in lb_svc_mdata_map %v", val, *svc_mdata_map)
			}
		}
	}
}

func (rest *RestOperations) AviVsCacheDel(vsKey avicache.NamespaceName, rest_op *utils.RestOp, key string) error {
	// Delete the SNI Child ref
	vs_cache, ok := rest.cache.VsCache.AviCacheGet(vsKey)
	if ok {
		vs_cache_obj, found := vs_cache.(*avicache.AviVsCache)
		if found {
			parent_vs_cache, parent_ok := rest.cache.VsCache.AviCacheGet(vs_cache_obj.ParentVSRef)
			if parent_ok {
				parent_vs_cache_obj, parent_found := parent_vs_cache.(*avicache.AviVsCache)
				if parent_found {
					// Find the SNI child and then remove
					rest.findSNIRefAndRemove(vsKey, parent_vs_cache_obj, key)
				}
			}
			if len(vs_cache_obj.VSVipKeyCollection) > 0 {
				vsvip := vs_cache_obj.VSVipKeyCollection[0].Name
				vsvipKey := avicache.NamespaceName{Namespace: vsKey.Namespace, Name: vsvip}
				utils.AviLog.Info.Printf("key: %s, msg: deleting vsvip cache for key: %s", key, vsvipKey)
				rest.cache.VSVIPCache.AviCacheDelete(vsvipKey)
			}
		}
	}
	utils.AviLog.Info.Printf("key: %s, msg: deleting vs cache for key: %s", key, vsKey)
	rest.cache.VsCache.AviCacheDelete(vsKey)

	return nil
}

func (rest *RestOperations) AviVSDel(uuid string, tenant string, key string) *utils.RestOp {
	path := "/api/virtualservice/" + uuid
	rest_op := utils.RestOp{Path: path, Method: "DELETE",
		Tenant: tenant, Model: "VirtualService", Version: utils.CtrlVersion}
	utils.AviLog.Info.Print(spew.Sprintf("VirtualService DELETE Restop %v \n",
		utils.Stringify(rest_op)))
	return &rest_op
}

func (rest *RestOperations) findSNIRefAndRemove(snichildkey avicache.NamespaceName, parentVsObj *avicache.AviVsCache, key string) {
	for i, sni_uuid := range parentVsObj.SNIChildCollection {
		sni_vs_key, ok := rest.cache.VsCache.AviCacheGetKeyByUuid(sni_uuid)
		if ok {
			if sni_vs_key.(avicache.NamespaceName).Name == snichildkey.Name {
				parentVsObj.SNIChildCollection = append(parentVsObj.SNIChildCollection[:i], parentVsObj.SNIChildCollection[i+1:]...)
				utils.AviLog.Info.Printf("key: %s, msg: removed sni key :%s", key, snichildkey.Name)
				break
			}
		}
	}
}

func (rest *RestOperations) AviVsVipBuild(vsvip_meta *nodes.AviVSVIPNode, cache_obj *avicache.AviVSVIPCache, key string) *utils.RestOp {
	name := vsvip_meta.Name
	tenant := fmt.Sprintf("/api/tenant/?name=%s", vsvip_meta.Tenant)
	cloudRef := "/api/cloud?name=" + utils.CloudName
	var dns_info_arr []*avimodels.DNSInfo
	var path string
	var rest_op utils.RestOp

	if cache_obj != nil {

		vsvip := rest.AviVsVipGet(key, cache_obj.Uuid, name)
		if vsvip == nil {
			return nil
		}
		for i, _ := range vsvip_meta.FQDNs {
			dns_info := avimodels.DNSInfo{Fqdn: &vsvip_meta.FQDNs[i]}
			dns_info_arr = append(dns_info_arr, &dns_info)
		}
		vsvip.DNSInfo = dns_info_arr
		path = "/api/vsvip/" + cache_obj.Uuid
		rest_op = utils.RestOp{Path: path, Method: utils.RestPut, Obj: vsvip,
			Tenant: vsvip_meta.Tenant, Model: "VsVip", Version: utils.CtrlVersion}
	} else {
		auto_alloc := true
		var vips []*avimodels.Vip
		vip := avimodels.Vip{AutoAllocateIP: &auto_alloc}

		mask := int32(24)
		addr := "172.18.0.0"
		atype := "V4"
		sip := avimodels.IPAddr{Type: &atype, Addr: &addr}
		ew_subnet := avimodels.IPAddrPrefix{IPAddr: &sip, Mask: &mask}
		var east_west bool
		if vsvip_meta.EastWest == true {
			vip.Subnet = &ew_subnet
			east_west = true
		} else {
			east_west = false
		}

		for i, _ := range vsvip_meta.FQDNs {
			dns_info := avimodels.DNSInfo{Fqdn: &vsvip_meta.FQDNs[i]}
			dns_info_arr = append(dns_info_arr, &dns_info)
		}
		vrfContextRef := "/api/vrfcontext?name=" + vsvip_meta.VrfContext
		vsvip := avimodels.VsVip{Name: &name, TenantRef: &tenant, CloudRef: &cloudRef,
			EastWestPlacement: &east_west, VrfContextRef: &vrfContextRef}
		vsvip.DNSInfo = dns_info_arr
		vips = append(vips, &vip)
		vsvip.Vip = vips
		macro := utils.AviRestObjMacro{ModelName: "VsVip", Data: vsvip}
		path = "/api/macro"
		// Patch an existing vsvip if it exists in the cache but not associated with this VS.
		vsvip_key := avicache.NamespaceName{Namespace: vsvip_meta.Tenant, Name: name}
		utils.AviLog.Warning.Printf("key: %s, seaching in cache for vsVip Key: %s", key, vsvip_key)
		vsvip_cache, ok := rest.cache.VSVIPCache.AviCacheGet(vsvip_key)
		if ok {
			vsvip_cache_obj, _ := vsvip_cache.(*avicache.AviVSVIPCache)
			vsvip_avi := rest.AviVsVipGet(key, vsvip_cache_obj.Uuid, name)
			if vsvip_avi == nil {
				return nil
			}
			for i, _ := range vsvip_meta.FQDNs {
				dns_info := avimodels.DNSInfo{Fqdn: &vsvip_meta.FQDNs[i]}
				dns_info_arr = append(dns_info_arr, &dns_info)
			}
			vsvip_avi.DNSInfo = dns_info_arr
			vsvip_avi.VrfContextRef = &vrfContextRef
			path = "/api/vsvip/" + vsvip_cache_obj.Uuid
			rest_op = utils.RestOp{Path: path, Method: utils.RestPut, Obj: vsvip_avi,
				Tenant: vsvip_meta.Tenant, Model: "VsVip", Version: utils.CtrlVersion}
		} else {
			rest_op = utils.RestOp{Path: path, Method: utils.RestPost, Obj: macro,
				Tenant: vsvip_meta.Tenant, Model: "VsVip", Version: utils.CtrlVersion}
		}
	}

	return &rest_op
}

func (rest *RestOperations) AviVsVipGet(key, uuid, name string) *avimodels.VsVip {
	if rest.aviRestPoolClient == nil {
		utils.AviLog.Warning.Printf("key: %s, msg: aviRestPoolClient during vsvip not initialized\n", key)
		return nil
	}
	if len(rest.aviRestPoolClient.AviClient) < 1 {
		utils.AviLog.Warning.Printf("key: %s, msg: client in aviRestPoolClient during vsvip not initialized\n", key)
		return nil
	}
	client := rest.aviRestPoolClient.AviClient[0]
	uri := "/api/vsvip/" + uuid

	rawData, err := client.AviSession.GetRaw(uri)
	if err != nil {
		utils.AviLog.Warning.Printf("VsVip Get uri %v returned err %v", uri, err)
		return nil
	}
	vsvip := avimodels.VsVip{}
	json.Unmarshal(rawData, &vsvip)

	return &vsvip
}

func (rest *RestOperations) AviVsVipDel(uuid string, tenant string, key string) *utils.RestOp {
	path := "/api/vsvip/" + uuid
	rest_op := utils.RestOp{Path: path, Method: "DELETE",
		Tenant: tenant, Model: "VsVip", Version: utils.CtrlVersion}
	utils.AviLog.Info.Print(spew.Sprintf("key: %s, msg: VSVIP DELETE Restop %v \n", key,
		utils.Stringify(rest_op)))
	return &rest_op
}

func (rest *RestOperations) AviVsVipCacheAdd(rest_op *utils.RestOp, vsKey avicache.NamespaceName, key string) error {
	if (rest_op.Err != nil) || (rest_op.Response == nil) {
		utils.AviLog.Warning.Printf("key: %s, rest_op has err or no reponse for vsvip err: %s, response: %s", key, rest_op.Err, rest_op.Response)
		return errors.New("Errored vsvip rest_op")
	}

	resp_elems, ok := RestRespArrToObjByType(rest_op, "vsvip", key)
	if ok != nil || resp_elems == nil {
		utils.AviLog.Warning.Printf("key: %s, msg: unable to find vsvip obj in resp %v", key, rest_op.Response)
		return errors.New("vsvip not found")
	}

	for _, resp := range resp_elems {
		name, ok := resp["name"].(string)
		if !ok {
			utils.AviLog.Warning.Printf("key: %s, msg: vsvip name not present in response %v", key, resp)
			continue
		}

		uuid, ok := resp["uuid"].(string)
		if !ok {
			utils.AviLog.Warning.Printf("key: %s, msg: vsvip Uuid not present in response %v", key, resp)
			continue
		}

		vsvip_cache_obj := avicache.AviVSVIPCache{Name: name, Tenant: rest_op.Tenant,
			Uuid: uuid}

		k := avicache.NamespaceName{Namespace: rest_op.Tenant, Name: name}
		rest.cache.VSVIPCache.AviCacheAdd(k, &vsvip_cache_obj)
		// Update the VS object
		vs_cache, ok := rest.cache.VsCache.AviCacheGet(vsKey)
		if ok {
			vs_cache_obj, found := vs_cache.(*avicache.AviVsCache)
			if found {
				utils.AviLog.Info.Printf("key: %s, msg: the VS cache before modification by VSVIP creation is :%v", key, utils.Stringify(vs_cache_obj))
				if vs_cache_obj.VSVipKeyCollection == nil {
					vs_cache_obj.VSVipKeyCollection = []avicache.NamespaceName{k}
				} else {
					if !utils.HasElem(vs_cache_obj.VSVipKeyCollection, k) {
						vs_cache_obj.VSVipKeyCollection = append(vs_cache_obj.VSVipKeyCollection, k)
					}
				}
				utils.AviLog.Info.Printf("key: %s, msg: modified the VS cache object for VSVIP collection. The cache now is :%v", key, utils.Stringify(vs_cache_obj))
			}

		} else {
			vs_cache_obj := avicache.AviVsCache{Name: vsKey.Name, Tenant: vsKey.Namespace,
				VSVipKeyCollection: []avicache.NamespaceName{k}}
			rest.cache.VSVIPCache.AviCacheAdd(vsKey, &vs_cache_obj)
			utils.AviLog.Info.Print(spew.Sprintf("key: %s, msg: added VS cache key during vsvip update %v val %v\n", key, vsKey,
				vs_cache_obj))
		}
		utils.AviLog.Info.Print(spew.Sprintf("key: %s, msg: added vsvip cache k %v val %v\n", key, k,
			vsvip_cache_obj))
	}

	return nil
}

func (rest *RestOperations) AviVsVipCacheDel(rest_op *utils.RestOp, vsKey avicache.NamespaceName, key string) error {
	vsvipkey := avicache.NamespaceName{Namespace: rest_op.Tenant, Name: rest_op.ObjName}
	rest.cache.VSVIPCache.AviCacheDelete(vsvipkey)
	if vsKey != (avicache.NamespaceName{}) {
		vs_cache, ok := rest.cache.VsCache.AviCacheGet(vsKey)
		if ok {
			vs_cache_obj, found := vs_cache.(*avicache.AviVsCache)
			if found {
				vs_cache_obj.VSVipKeyCollection = Remove(vs_cache_obj.VSVipKeyCollection, vsvipkey)
			}
		}
	}

	return nil

}
