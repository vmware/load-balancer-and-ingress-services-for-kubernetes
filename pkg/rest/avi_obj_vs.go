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

	var vip avimodels.Vip
	if rest_method == utils.RestPost {
		auto_alloc := true
		vip = avimodels.Vip{AutoAllocateIP: &auto_alloc}
	} else {
		auto_alloc_put := true
		auto_allocate_floating_ip := false
		vip = avimodels.Vip{AutoAllocateIP: &auto_alloc_put, AutoAllocateFloatingIP: &auto_allocate_floating_ip}
	}
	// E/W placement subnet is don't care, just needs to be a valid subnet
	mask := int32(24)
	addr := "172.18.0.0"
	atype := "V4"
	sip := avimodels.IPAddr{Type: &atype, Addr: &addr}
	ew_subnet := avimodels.IPAddrPrefix{IPAddr: &sip, Mask: &mask}
	var east_west bool
	if vs_meta.EastWest == true {
		vip.Subnet = &ew_subnet
		east_west = true
	} else {
		east_west = false
	}
	network_prof := "/api/networkprofile/?name=" + vs_meta.NetworkProfile
	app_prof := "/api/applicationprofile/?name=" + vs_meta.ApplicationProfile
	// TODO use PoolGroup and use policies if there are > 1 pool, etc.
	name := vs_meta.Name
	var dns_info_arr []*avimodels.DNSInfo
	// Form the DNS_Info name_of_vs.namespace.<dns_ipam>
	cloud, _ := rest.cache.CloudKeyCache.AviCacheGet(utils.CloudName)
	utils.AviLog.Info.Printf("key: %s, msg: build vs objs | name: %v \n| vs_meta %v \n| cloud %v", key, name, vs_meta, cloud)
	fqdn := name + "." + vs_meta.Tenant + "." + cloud.(*avicache.AviCloudPropertyCache).NSIpamDNS
	dns_info := avimodels.DNSInfo{Fqdn: &fqdn}
	dns_info_arr = append(dns_info_arr, &dns_info)
	cksum := vs_meta.CloudConfigCksum
	checksumstr := fmt.Sprint(cksum)
	cr := utils.OSHIFT_K8S_CLOUD_CONNECTOR
	cloudRef := "/api/cloud?name=" + utils.CloudName
	vs := avimodels.VirtualService{Name: &name,
		NetworkProfileRef:     &network_prof,
		ApplicationProfileRef: &app_prof,
		CloudConfigCksum:      &checksumstr,
		CreatedBy:             &cr,
		DNSInfo:               dns_info_arr,
		EastWestPlacement:     &east_west,
		CloudRef:              &cloudRef}

	if vs_meta.DefaultPoolGroup != "" {
		pool_ref := "/api/poolgroup/?name=" + vs_meta.DefaultPoolGroup
		vs.PoolGroupRef = &pool_ref
	}
	vs.Vip = append(vs.Vip, &vip)
	tenant := fmt.Sprintf("/api/tenant/?name=%s", vs_meta.Tenant)
	vs.TenantRef = &tenant

	if vs_meta.SNIParent {
		// This is a SNI parent
		utils.AviLog.Info.Printf("key: %s, msg: vs %s is a SNI Parent", key, vs_meta.Name)
		vh_parent := "VS_TYPE_VH_PARENT"
		vs.Type = &vh_parent
	}
	// TODO other fields like cloud_ref, mix of TCP & UDP protocols, etc.

	for _, pp := range vs_meta.PortProto {
		port := pp.Port
		svc := avimodels.Service{Port: &port}
		if pp.Protocol == utils.TCP {
			utils.AviLog.Info.Printf("key: %s, msg: processing TCP ports for VS creation :%v", key, pp.Port)
			onw_profile := "/api/networkprofile/?name=System-TCP-Proxy"
			svc.OverrideNetworkProfileRef = &onw_profile
			port := pp.Port
			var sproto string
			sproto = "PROTOCOL_TYPE_TCP_PROXY"
			pg_name := FindPoolGroupForPort(vs_meta.TCPPoolGroupRefs, port)
			if pg_name != "" {
				utils.AviLog.Info.Printf("key: %s, msg: TCP ports for VS creation returned PG: %s", key, pg_name)
				oapp_profile := "/api/applicationprofile/?name=System-L4-Application"
				pg_ref := "/api/poolgroup/?name=" + pg_name
				sps := avimodels.ServicePoolSelector{ServicePoolGroupRef: &pg_ref,
					ServicePort: &port, ServiceProtocol: &sproto}
				vs.ServicePoolSelect = append(vs.ServicePoolSelect, &sps)
				svc.OverrideApplicationProfileRef = &oapp_profile
			} else {
				utils.AviLog.Info.Printf("key: %s, msg: TCP ports for VS creation returned no matching PGs", key)
			}

		} else if pp.Protocol == utils.UDP && vs_meta.NetworkProfile == "System-TCP-Proxy" {
			onw_profile := "/api/networkprofile/?name=System-UDP-Fast-Path"
			svc.OverrideNetworkProfileRef = &onw_profile
		}
		if pp.Secret != "" || pp.Passthrough {
			ssl_enabled := true
			svc.EnableSsl = &ssl_enabled
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

func (rest *RestOperations) AviVsCacheAdd(rest_op *utils.RestOp, key string) error {
	if (rest_op.Err != nil) || (rest_op.Response == nil) {
		utils.AviLog.Warning.Printf("key: %s, msg:rest_op has err or no reponse", key)
		return errors.New("Errored rest_op")
	}

	resp_elems, ok := RestRespArrToObjByType(rest_op, "virtualservice", key)
	if ok != nil || resp_elems == nil {
		utils.AviLog.Warning.Printf("key: %s, msg: unable to find pool obj in resp %v", key, rest_op.Response)
		return errors.New("pool not found")
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
			vs_uuid := ExtractVsUuid(vh_parent_uuid.(string))
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
		k := avicache.NamespaceName{Namespace: rest_op.Tenant, Name: name}
		vs_cache, ok := rest.cache.VsCache.AviCacheGet(k)
		if ok {
			vs_cache_obj, found := vs_cache.(*avicache.AviVsCache)
			if found {
				vs_cache_obj.Uuid = uuid
				vs_cache_obj.CloudConfigCksum = cksum
				utils.AviLog.Info.Print(spew.Sprintf("key: %s, msg: updated VS cache key %v val %v\n", key, k,
					utils.Stringify(vs_cache_obj)))
			}
		} else {
			vs_cache_obj := avicache.AviVsCache{Name: name, Tenant: rest_op.Tenant,
				Uuid: uuid, CloudConfigCksum: cksum}
			rest.cache.VsCache.AviCacheAdd(k, &vs_cache_obj)
			utils.AviLog.Info.Print(spew.Sprintf("key: %s, msg: added VS cache key %v val %v\n", key, k,
				vs_cache_obj))
		}

	}

	return nil
}

func (rest *RestOperations) AviVsCacheDel(vs_cache *avicache.AviCache, rest_op *utils.RestOp, key string) error {

	vsKey := avicache.NamespaceName{Namespace: rest_op.Tenant, Name: rest_op.ObjName}
	vs_cache.AviCacheDelete(vsKey)

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
