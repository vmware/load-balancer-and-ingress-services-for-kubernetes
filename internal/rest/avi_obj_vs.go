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
	"sort"
	"strconv"
	"strings"

	avicache "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/status"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	"github.com/davecgh/go-spew/spew"
	avimodels "github.com/vmware/alb-sdk/go/models"
	"google.golang.org/protobuf/proto"
)

const VSVIP_NOTFOUND = "VsVip object not found"

// TODO: convert it to generalized function to be used by evh and sni both.
func setDedicatedVSNodeProperties(vs *avimodels.VirtualService, vs_meta *nodes.AviVsNode) {
	var datascriptCollection []*avimodels.VSDataScripts
	// this overwrites the sslkeycert created from the Secret object, with the one mentioned in HostRule.TLS
	if len(vs_meta.SSLKeyCertAviRef) != 0 {
		vs.SslKeyAndCertificateRefs = append(vs.SslKeyAndCertificateRefs, vs_meta.SSLKeyCertAviRef...)
	} else {
		for _, sslkeycert := range vs_meta.SSLKeyCertRefs {
			certName := "/api/sslkeyandcertificate/?name=" + sslkeycert.Name
			vs.SslKeyAndCertificateRefs = append(vs.SslKeyAndCertificateRefs, certName)
		}
	}
	vs.SslProfileRef = &vs_meta.SSLProfileRef
	//set datascripts to VS from hostrule crd
	for i, script := range vs_meta.VsDatascriptRefs {
		j := int32(i)
		datascript := script
		datascripts := &avimodels.VSDataScripts{VsDatascriptSetRef: &datascript, Index: &j}
		datascriptCollection = append(datascriptCollection, datascripts)
	}
	vs.VsDatascripts = datascriptCollection
	if vs_meta.AppProfileRef != "" {
		// hostrule ref overrides defaults
		vs.ApplicationProfileRef = &vs_meta.AppProfileRef
	}
	vs.WafPolicyRef = &vs_meta.WafPolicyRef
	vs.ErrorPageProfileRef = &vs_meta.ErrorPageProfileRef
	vs.AnalyticsProfileRef = &vs_meta.AnalyticsProfileRef
	vs.EastWestPlacement = proto.Bool(false)
	vs.Enabled = vs_meta.Enabled
	normal_vs_type := utils.VS_TYPE_NORMAL
	vs.Type = &normal_vs_type
}

func (rest *RestOperations) AviVsBuild(vs_meta *nodes.AviVsNode, rest_method utils.RestMethod, cache_obj *avicache.AviVsCache, key string) []*utils.RestOp {
	if lib.CheckObjectNameLength(vs_meta.Name, lib.SNIVS) {
		utils.AviLog.Warnf("key: %s not processing VS object", key)
		return nil
	}
	if vs_meta.IsSNIChild {
		rest_ops := rest.AviVsSniBuild(vs_meta, rest_method, cache_obj, key)
		return rest_ops
	} else {
		svc_mdata_json, _ := json.Marshal(&vs_meta.ServiceMetadata)
		svc_mdata := string(svc_mdata_json)

		vs := avimodels.VirtualService{
			Name:                  proto.String(vs_meta.Name),
			CloudConfigCksum:      proto.String(strconv.Itoa(int(vs_meta.CloudConfigCksum))),
			CreatedBy:             proto.String(lib.AKOUser),
			CloudRef:              proto.String("/api/cloud?name=" + utils.CloudName),
			TenantRef:             proto.String(fmt.Sprintf("/api/tenant/?name=%s", vs_meta.Tenant)),
			ApplicationProfileRef: proto.String("/api/applicationprofile/?name=" + vs_meta.ApplicationProfile),
			SeGroupRef:            proto.String("/api/serviceenginegroup?name=" + vs_meta.ServiceEngineGroup),
			VrfContextRef:         proto.String("/api/vrfcontext?name=" + vs_meta.VrfContext),
			WafPolicyRef:          &vs_meta.WafPolicyRef,
			AnalyticsProfileRef:   &vs_meta.AnalyticsProfileRef,
			ErrorPageProfileRef:   &vs_meta.ErrorPageProfileRef,
			Enabled:               vs_meta.Enabled,
			ServiceMetadata:       &svc_mdata,
		}

		if vs_meta.AppProfileRef != "" {
			// hostrule ref overrides defaults
			vs.ApplicationProfileRef = proto.String(vs_meta.AppProfileRef)
		}

		if lib.GetT1LRPath() != "" {
			// Clear the vrfContextRef
			vs.VrfContextRef = nil
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

		if lib.GetAdvancedL4() {
			vs.IgnPoolNetReach = proto.Bool(true)
		}

		if vs_meta.DefaultPoolGroup != "" {
			vs.PoolGroupRef = proto.String("/api/poolgroup/?name=" + vs_meta.DefaultPoolGroup)
		}

		if len(vs_meta.VSVIPRefs) > 0 {
			vs.VsvipRef = proto.String("/api/vsvip/?name=" + vs_meta.VSVIPRefs[0].Name)
		} else {
			utils.AviLog.Warnf("key: %s, msg: unable to set the vsvip reference")
		}

		if vs_meta.SNIParent {
			// This is a SNI parent
			utils.AviLog.Debugf("key: %s, msg: vs %s is a SNI Parent", key, vs_meta.Name)
			vh_parent := utils.VS_TYPE_VH_PARENT
			vs.Type = &vh_parent
		}

		for i, pp := range vs_meta.PortProto {
			port := pp.Port
			svc := avimodels.Service{
				Port:         &port,
				EnableSsl:    &vs_meta.PortProto[i].EnableSSL,
				PortRangeEnd: &port,
			}
			if vs_meta.NetworkProfile == utils.MIXED_NET_PROFILE && pp.Protocol == utils.UDP {
				svc.OverrideNetworkProfileRef = proto.String("/api/networkprofile/?name=" + utils.SYSTEM_UDP_FAST_PATH)
			}
			vs.Services = append(vs.Services, &svc)
		}

		// In case the VS has services that are a mix of TCP and UDP sockets,
		// we create the VS with global network profile TCP Fast Path,
		// and override required services with UDP Fast Path.
		if vs_meta.NetworkProfile == utils.MIXED_NET_PROFILE {
			vs_meta.NetworkProfile = utils.TCP_NW_FAST_PATH
		}
		vs.NetworkProfileRef = proto.String("/api/networkprofile/?name=" + vs_meta.NetworkProfile)

		var datascriptCollection []*avimodels.VSDataScripts
		if vs_meta.SharedVS {
			// This is a shared VS - which should have a datascript
			for i, ds := range vs_meta.HTTPDSrefs {
				j := int32(i)
				dsRef := "/api/vsdatascriptset/?name=" + ds.Name
				vsdatascript := &avimodels.VSDataScripts{Index: &j, VsDatascriptSetRef: &dsRef}
				datascriptCollection = append(datascriptCollection, vsdatascript)
			}
		}

		// Overwrite datascript policies from hostrule to the Parent VS.
		if len(vs_meta.VsDatascriptRefs) > 0 {
			datascriptCollection = make([]*avimodels.VSDataScripts, len(vs_meta.VsDatascriptRefs))
			for i, script := range vs_meta.VsDatascriptRefs {
				j := int32(i)
				datascript := script
				datascripts := &avimodels.VSDataScripts{VsDatascriptSetRef: &datascript, Index: &j}
				datascriptCollection = append(datascriptCollection, datascripts)
			}
		}
		vs.VsDatascripts = datascriptCollection

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

		//Dedicated VS
		if vs_meta.Dedicated {
			setDedicatedVSNodeProperties(&vs, vs_meta)
		}

		bufferLen := int32(len(httpPolicyCollection)) + internalPolicyIndexBuffer + 5
		for i, policy := range vs_meta.HttpPolicySetRefs {
			j := int32(i) + bufferLen
			httpPolicy := policy
			httpPolicies := &avimodels.HTTPPolicies{HTTPPolicySetRef: &httpPolicy, Index: &j}
			httpPolicyCollection = append(httpPolicyCollection, httpPolicies)
		}
		vs.HTTPPolicies = httpPolicyCollection

		if strings.Contains(*vs.Name, lib.PassthroughPrefix) && !strings.HasSuffix(*vs.Name, lib.PassthroughInsecure) {
			// This is a passthrough secure VS, we want the VS to be down if all the pools are down.
			vsDownOnPoolDown := true
			vs.RemoveListeningPortOnVsDown = &vsDownOnPoolDown
		}

		if vs_meta.SharedVS {
			vs.Markers = lib.GetMarkers()
		} else {
			vs.Markers = lib.GetAllMarkers(vs_meta.AviMarkers)
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
		vs.AnalyticsPolicy = vs_meta.GetAnalyticsPolicy()

		var rest_ops []*utils.RestOp

		var rest_op utils.RestOp
		var path string

		// VS objects cache can be created by other objects and they would just set VS name and not uud
		// Do a POST call in that case
		if rest_method == utils.RestPut && cache_obj.Uuid != "" {
			path = "/api/virtualservice/" + cache_obj.Uuid
			rest_op = utils.RestOp{
				Path:    path,
				Method:  rest_method,
				Obj:     vs,
				Tenant:  vs_meta.Tenant,
				Model:   "VirtualService",
				ObjName: *vs.Name,
			}
			rest_ops = append(rest_ops, &rest_op)
		} else {
			path = "/api/virtualservice/"
			rest_op = utils.RestOp{
				Path:    path,
				Method:  utils.RestPost,
				Obj:     vs,
				Tenant:  vs_meta.Tenant,
				Model:   "VirtualService",
				ObjName: *vs.Name,
			}
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
		EastWestPlacement:     proto.Bool(false),
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
	sniChild.AnalyticsPolicy = vs_meta.GetAnalyticsPolicy()
	if lib.GetT1LRPath() != "" {
		// Clear the vrfContextRef
		sniChild.VrfContextRef = nil
	}
	//This VS has a TLSKeyCert associated, we need to mark 'type': 'VS_TYPE_VH_PARENT'
	vh_type := utils.VS_TYPE_VH_CHILD
	sniChild.Type = &vh_type
	vhParentUuid := "/api/virtualservice/?name=" + vs_meta.VHParentName
	sniChild.VhParentVsUUID = &vhParentUuid
	sniChild.VhDomainName = vs_meta.VHDomainNames
	ignPool := false
	sniChild.IgnPoolNetReach = &ignPool

	sniChild.Markers = lib.GetAllMarkers(vs_meta.AviMarkers)

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
		if len(vs_meta.SSLKeyCertAviRef) != 0 {
			sniChild.SslKeyAndCertificateRefs = append(sniChild.SslKeyAndCertificateRefs, vs_meta.SSLKeyCertAviRef...)
		} else {
			for _, sslkeycert := range vs_meta.SSLKeyCertRefs {
				certName := "/api/sslkeyandcertificate/?name=" + sslkeycert.Name
				sniChild.SslKeyAndCertificateRefs = append(sniChild.SslKeyAndCertificateRefs, certName)
			}
		}
		sniChild.HTTPPolicies = AviVsHttpPSAdd(vs_meta, false)
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
			Obj:     sniChild,
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
			Obj:     sniChild,
			Tenant:  vs_meta.Tenant,
			Model:   "VirtualService",
		}
		rest_ops = append(rest_ops, &rest_op)
	}

	return rest_ops
}

func AviVsHttpPSAdd(vs_meta interface{}, isEVH bool) []*avimodels.HTTPPolicies {
	var httpPolicyRef []*nodes.AviHttpPolicySetNode
	var httpPSRef []string
	var httpPolicyCollection []*avimodels.HTTPPolicies
	if isEVH {
		vsMeta := vs_meta.(*nodes.AviEvhVsNode)
		httpPolicyRef = vsMeta.HttpPolicyRefs
		httpPSRef = vsMeta.HttpPolicySetRefs
	} else {
		vsMeta := vs_meta.(*nodes.AviVsNode)
		httpPolicyRef = vsMeta.HttpPolicyRefs
		httpPSRef = vsMeta.HttpPolicySetRefs
	}

	internalPolicyIndexBuffer := int32(11)

	var httpsWithHppMap []*nodes.AviHttpPolicySetNode
	var httpsNoHppMap []*nodes.AviHttpPolicySetNode
	for _, http := range httpPolicyRef {
		if http.HppMap != nil && http.HppMap[0].Path != nil {
			httpsWithHppMap = append(httpsWithHppMap, http)
		} else {
			httpsNoHppMap = append(httpsNoHppMap, http)
		}
	}

	sort.Slice(httpsWithHppMap, func(i, j int) bool {
		return httpsWithHppMap[i].HppMap[0].Path[0] < httpsWithHppMap[j].HppMap[0].Path[0]
	})

	var j int32
	for i, http := range httpsNoHppMap {
		j = int32(i) + internalPolicyIndexBuffer
		k := j
		httpPolicy := fmt.Sprintf("/api/httppolicyset/?name=%s", http.Name)
		httpPolicies := &avimodels.HTTPPolicies{HTTPPolicySetRef: &httpPolicy, Index: &k}
		httpPolicyCollection = append(httpPolicyCollection, httpPolicies)
	}
	if len(httpsNoHppMap) == 0 {
		j = internalPolicyIndexBuffer
	} else {
		j = j + 1
	}
	for _, http := range httpsWithHppMap {
		k := j
		j = j + 1
		httpPolicy := fmt.Sprintf("/api/httppolicyset/?name=%s", http.Name)
		httpPolicies := &avimodels.HTTPPolicies{HTTPPolicySetRef: &httpPolicy, Index: &k}
		httpPolicyCollection = append(httpPolicyCollection, httpPolicies)
	}

	// from hostrule CRD
	bufferLen := int32(len(httpPolicyCollection)) + internalPolicyIndexBuffer + 5
	for i, policy := range httpPSRef {
		var j int32
		j = int32(i) + bufferLen
		httpPolicy := policy
		httpPolicies := &avimodels.HTTPPolicies{HTTPPolicySetRef: &httpPolicy, Index: &j}
		httpPolicyCollection = append(httpPolicyCollection, httpPolicies)
	}

	return httpPolicyCollection
}

func (rest *RestOperations) StatusUpdateForPool(restMethod utils.RestMethod, vs_cache_obj *avicache.AviVsCache, key string) {
	if restMethod == utils.RestPost || restMethod == utils.RestDelete || restMethod == utils.RestPut {
		for _, poolkey := range vs_cache_obj.PoolKeyCollection {
			// Fetch the pool object from cache and check the service metadata
			pool_cache, ok := rest.cache.PoolCache.AviCacheGet(poolkey)
			if ok {
				pool_cache_obj, found := pool_cache.(*avicache.AviPoolCache)
				if found {
					utils.AviLog.Infof("key: %s, msg: found pool: %s, will update status", key, poolkey.Name)
					IPAddrs := rest.GetIPAddrsFromCache(vs_cache_obj)
					if len(IPAddrs) == 0 {
						utils.AviLog.Warnf("key: %s, msg: Unable to find VIP corresponding to Pool %s vsCache %v", key, pool_cache_obj.Name, utils.Stringify(vs_cache_obj))
						continue
					}
					switch pool_cache_obj.ServiceMetadataObj.ServiceMetadataMapping("Pool") {
					case lib.GatewayPool:
						updateOptions := status.UpdateOptions{
							Vip:                IPAddrs,
							ServiceMetadata:    pool_cache_obj.ServiceMetadataObj,
							Key:                key,
							VirtualServiceUUID: vs_cache_obj.Uuid,
							VSName:             vs_cache_obj.Name,
						}
						statusOption := status.StatusOptions{
							ObjType: utils.L4LBService,
							Op:      lib.UpdateStatus,
							Options: &updateOptions,
						}
						utils.AviLog.Infof("key: %s Publishing to status queue, options: %v", updateOptions.ServiceMetadata.NamespaceServiceName[0], utils.Stringify(statusOption))
						status.PublishToStatusQueue(updateOptions.ServiceMetadata.NamespaceServiceName[0], statusOption)
					case lib.SNIInsecureOrEVHPool:
						updateOptions := status.UpdateOptions{
							Vip:                IPAddrs,
							ServiceMetadata:    pool_cache_obj.ServiceMetadataObj,
							Key:                key,
							VirtualServiceUUID: vs_cache_obj.Uuid,
							VSName:             vs_cache_obj.Name,
						}
						statusOption := status.StatusOptions{
							ObjType: utils.Ingress,
							Op:      lib.UpdateStatus,
							Options: &updateOptions,
						}
						if utils.GetInformers().RouteInformer != nil {
							statusOption.ObjType = utils.OshiftRoute
						}
						if pool_cache_obj.ServiceMetadataObj.IsMCIIngress {
							statusOption.ObjType = lib.MultiClusterIngress
						}
						utils.AviLog.Debugf("key: %s Publishing to status queue, options: %v", updateOptions.ServiceMetadata.IngressName, utils.Stringify(statusOption))
						status.PublishToStatusQueue(updateOptions.ServiceMetadata.IngressName, statusOption)
					}
				}
			}
		}
	}
}

func (rest *RestOperations) StatusUpdateForVS(restMethod utils.RestMethod, vsCacheObj *avicache.AviVsCache, key string) {
	IPAddrs := rest.GetIPAddrsFromCache(vsCacheObj)
	serviceMetadataObj := vsCacheObj.ServiceMetadataObj
	switch serviceMetadataObj.ServiceMetadataMapping("VS") {
	case lib.GatewayVS:
		updateOptions := status.UpdateOptions{
			Vip:             IPAddrs,
			ServiceMetadata: serviceMetadataObj,
			Key:             key,
			VSName:          vsCacheObj.Name,
		}
		statusOption := status.StatusOptions{
			ObjType: lib.Gateway,
			Op:      lib.UpdateStatus,
			Options: &updateOptions,
		}
		if lib.UseServicesAPI() {
			statusOption.ObjType = lib.SERVICES_API
		}
		utils.AviLog.Infof("key: %s Publishing to status queue, options: %v", updateOptions.ServiceMetadata.Gateway, utils.Stringify(statusOption))
		status.PublishToStatusQueue(updateOptions.ServiceMetadata.Gateway, statusOption)
	case lib.ServiceTypeLBVS:
		updateOptions := status.UpdateOptions{
			Vip:                IPAddrs,
			ServiceMetadata:    serviceMetadataObj,
			Key:                key,
			VirtualServiceUUID: vsCacheObj.Uuid,
			VSName:             vsCacheObj.Name,
		}
		statusOption := status.StatusOptions{
			ObjType: utils.L4LBService,
			Op:      lib.UpdateStatus,
			Options: &updateOptions,
		}
		utils.AviLog.Infof("key: %s Publishing to status queue, options: %v", updateOptions.ServiceMetadata.NamespaceServiceName[0], utils.Stringify(statusOption))
		status.PublishToStatusQueue(updateOptions.ServiceMetadata.NamespaceServiceName[0], statusOption)
	case lib.ChildVS:
		rest.StatusUpdateForPool(restMethod, vsCacheObj, key)
		// updateOptions := status.UpdateOptions{
		// 	Vip:                IPAddrs,
		// 	ServiceMetadata:    serviceMetadataObj,
		// 	Key:                key,
		// 	VirtualServiceUUID: vsCacheObj.Uuid,
		// 	VSName:             vsCacheObj.Name,
		// }
		// statusOption := status.StatusOptions{
		// 	ObjType: utils.Ingress,
		// 	Op:      lib.UpdateStatus,
		// 	Options: &updateOptions,
		// }
		// if utils.GetInformers().RouteInformer != nil {
		// 	statusOption.ObjType = utils.OshiftRoute
		// }
		// utils.AviLog.Infof("key: %s Publishing to status queue, options: %v", updateOptions.ServiceMetadata.IngressName, utils.Stringify(statusOption))
		// status.PublishToStatusQueue(updateOptions.ServiceMetadata.IngressName, statusOption)

	}
}

func (rest *RestOperations) AviVsCacheAdd(rest_op *utils.RestOp, key string) error {
	if (rest_op.Err != nil) || (rest_op.Response == nil) {
		utils.AviLog.Warnf("key: %s, rest_op has err or no response for VS, err: %v, response: %v", key, rest_op.Err, rest_op.Response)
		return errors.New("Error rest_op")
	}

	resp_elems := rest.restOperator.RestRespArrToObjByType(rest_op, "virtualservice", key)
	if resp_elems == nil {
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
				utils.AviLog.Info(spew.Sprintf("key: %s, msg: added VS cache key during SNI update %v val %v", key, parentKey,
					vs_cache_obj))
			}
		}
		var svc_mdata_obj lib.ServiceMetadataObj
		if resp["service_metadata"] != nil {
			utils.AviLog.Infof("key:%s, msg: Service Metadata: %s", key, resp["service_metadata"])
			if err := json.Unmarshal([]byte(resp["service_metadata"].(string)),
				&svc_mdata_obj); err != nil {
				utils.AviLog.Warnf("Error parsing service metadata :%v", err)
			}
		}

		k := avicache.NamespaceName{Namespace: rest_op.Tenant, Name: name}
		vs_cache, ok := rest.cache.VsCacheMeta.AviCacheGet(k)
		var vs_cache_obj *avicache.AviVsCache
		var found bool
		if ok {
			vs_cache_obj, found = vs_cache.(*avicache.AviVsCache)
			if found {
				vs_cache_obj.Uuid = uuid
				vs_cache_obj.CloudConfigCksum = cksum

				status.HostRuleEventBroadcast(vs_cache_obj.Name, vs_cache_obj.ServiceMetadataObj.CRDStatus, svc_mdata_obj.CRDStatus)
				vs_cache_obj.ServiceMetadataObj = svc_mdata_obj
				if val, ok := resp["enable_rhi"].(bool); ok {
					vs_cache_obj.EnableRhi = val
				}
				if vhParentKey != nil {
					vs_cache_obj.ParentVSRef = vhParentKey.(avicache.NamespaceName)
				}

				vs_cache_obj.LastModified = lastModifiedStr
				if lastModifiedStr == "" {
					vs_cache_obj.InvalidData = true
				} else {
					vs_cache_obj.InvalidData = false
				}
				utils.AviLog.Debug(spew.Sprintf("key: %s, msg: updated VS cache key %v val %v", key, k,
					utils.Stringify(vs_cache_obj)))

				// This code is most likely hit when the first time a shard vs is created and the vs_cache_obj is populated from the pool update.
				// But before this a pool may have got created as a part of the macro operation, so update the ingress status here.
				// rest.StatusUpdateForPool(rest_op.Method, vs_cache_obj, key)
			}

		} else {
			vs_cache_obj = &avicache.AviVsCache{
				Name:               name,
				Tenant:             rest_op.Tenant,
				Uuid:               uuid,
				CloudConfigCksum:   cksum,
				ServiceMetadataObj: svc_mdata_obj,
				LastModified:       lastModifiedStr,
			}
			if val, ok := resp["enable_rhi"].(bool); ok {
				vs_cache_obj.EnableRhi = val
			}
			if lastModifiedStr == "" {
				vs_cache_obj.InvalidData = true
			}

			rest.cache.VsCacheMeta.AviCacheAdd(k, vs_cache_obj)
			status.HostRuleEventBroadcast(vs_cache_obj.Name, lib.CRDMetadata{}, svc_mdata_obj.CRDStatus)
			utils.AviLog.Infof("key: %s, msg: added VS cache key %v val %v", key, k, utils.Stringify(vs_cache_obj))
		}

		rest.StatusUpdateForVS(rest_op.Method, vs_cache_obj, key)
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
				if len(vs_cache_obj.VSVipKeyCollection) > 0 {
					vsvip := vs_cache_obj.VSVipKeyCollection[0].Name
					vsvipKey := avicache.NamespaceName{Namespace: vsKey.Namespace, Name: vsvip}
					utils.AviLog.Infof("key: %s, msg: deleting vsvip cache for key: %s", key, vsvipKey)
					rest.cache.VSVIPCache.AviCacheDelete(vsvipKey)
				}
			}

			switch vs_cache_obj.ServiceMetadataObj.ServiceMetadataMapping("VS") {
			case lib.GatewayVS:
				updateOptions := status.UpdateOptions{
					ServiceMetadata: vs_cache_obj.ServiceMetadataObj,
					Key:             key,
					VSName:          vs_cache_obj.Name,
				}
				statusOption := status.StatusOptions{
					ObjType: lib.Gateway,
					Op:      lib.DeleteStatus,
					Options: &updateOptions,
				}
				utils.AviLog.Infof("key: %s Publishing to status queue, options: %v", updateOptions.ServiceMetadata.Gateway, utils.Stringify(statusOption))
				status.PublishToStatusQueue(updateOptions.ServiceMetadata.Gateway, statusOption)
				// The pools would have service metadata for backend services, corresponding to which
				// statuses need to be deleted.
				for _, poolKey := range vs_cache_obj.PoolKeyCollection {
					rest.DeletePoolIngressStatus(poolKey, true, vs_cache_obj.Name, key)
				}
			case lib.ServiceTypeLBVS:
				updateOptions := status.UpdateOptions{
					ServiceMetadata:    vs_cache_obj.ServiceMetadataObj,
					Key:                key,
					VirtualServiceUUID: vs_cache_obj.Uuid,
					VSName:             vs_cache_obj.Name,
				}
				statusOption := status.StatusOptions{
					ObjType: utils.L4LBService,
					Op:      lib.DeleteStatus,
					Options: &updateOptions,
				}
				utils.AviLog.Infof("key: %s Publishing to status queue, options: %v", vs_cache_obj.ServiceMetadataObj.NamespaceServiceName[0], utils.Stringify(statusOption))
				status.PublishToStatusQueue(vs_cache_obj.ServiceMetadataObj.NamespaceServiceName[0], statusOption)
			case lib.ChildVS:
				if !hostFoundInParentPool {
					// TODO: revisit
					// updateOptions := status.UpdateOptions{
					// 	ServiceMetadata: vs_cache_obj.ServiceMetadataObj,
					// 	Key:             key,
					// 	VSName:          vs_cache_obj.Name,
					// }
					// statusOption := status.StatusOptions{
					// 	ObjType: utils.Ingress,
					// 	Op:      lib.DeleteStatus,
					// 	IsVSDel: true,
					// 	Options: &updateOptions,
					// }
					// if utils.GetInformers().RouteInformer != nil {
					// 	statusOption.ObjType = utils.OshiftRoute
					// }
					// status.PublishToStatusQueue(updateOptions.ServiceMetadata.IngressName, statusOption)

					for _, poolKey := range vs_cache_obj.PoolKeyCollection {
						rest.DeletePoolIngressStatus(poolKey, true, vs_cache_obj.Name, key)
					}
				}

				status.HostRuleEventBroadcast(vs_cache_obj.Name, vs_cache_obj.ServiceMetadataObj.CRDStatus, lib.CRDMetadata{})
			default:
				// insecure ingress status updates in regular AKO.
				for _, poolKey := range vs_cache_obj.PoolKeyCollection {
					rest.DeletePoolIngressStatus(poolKey, true, vs_cache_obj.Name, key)
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
	rest_op := utils.RestOp{
		Path:   path,
		Method: "DELETE",
		Tenant: tenant,
		Model:  "VirtualService",
	}
	utils.AviLog.Info(spew.Sprintf("key: %s, msg: VirtualService DELETE Restop %v ",
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
				utils.AviLog.Debugf("key: %s, msg: hostname %v present in %s pool collection, will skip ingress status delete",
					key, parentVs.Name, hostname)
				return true
			}
		}
	}
	return false
}

func (rest *RestOperations) GetIPAddrsFromCache(vsCache *avicache.AviVsCache) []string {
	var IPAddrs []string
	if len(vsCache.VSVipKeyCollection) == 0 {
		parentVSKey := vsCache.ParentVSRef
		parentCache, ok := rest.cache.VsCacheMeta.AviCacheGet(parentVSKey)
		if ok {
			parentCacheObj, _ := parentCache.(*avicache.AviVsCache)
			if parentCacheObj != nil && len(parentCacheObj.VSVipKeyCollection) > 0 {
				// This essentially assigns the value of parent VS Cache to `vsCache`
				// meaning that if we were originally unable to find VSVIP attached to the
				// original vsCache (it was a child VS!), then we check whether a parent for
				// the original vsCache exists. If the parent exists and the parent has VSVIP
				// references (which it should), going forward we would traverse through it's
				// VSVIP Collection and fetch the IP Addresses.
				// If the original vsCache had VSVIP collection (it was a parent VS), we simply
				// donot arrive at this step and go ahead fetching IP addresses from it's VSVIP
				// Collection itself.
				utils.AviLog.Infof("Getting IP Address from parent VS %v", parentCacheObj.Name)
				vsCache = parentCacheObj
			}
		}
	}

	for _, vsvipkey := range vsCache.VSVipKeyCollection {
		vsvip_cache, ok := rest.cache.VSVIPCache.AviCacheGet(vsvipkey)
		if ok {
			vsvip_cache_obj, found := vsvip_cache.(*avicache.AviVSVIPCache)
			if found {
				if len(vsvip_cache_obj.Fips) != 0 {
					IPAddrs = vsvip_cache_obj.Fips
				} else if len(vsvip_cache_obj.V6IPs) != 0 {
					IPAddrs = vsvip_cache_obj.V6IPs
				} else {
					IPAddrs = vsvip_cache_obj.Vips
				}
			}
		}
	}

	return IPAddrs
}
