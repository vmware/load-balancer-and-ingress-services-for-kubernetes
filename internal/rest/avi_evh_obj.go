package rest

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/avinetworks/sdk/go/clients"
	avimodels "github.com/avinetworks/sdk/go/models"
	"github.com/avinetworks/sdk/go/session"
	avicache "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/api/models"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
	"google.golang.org/protobuf/proto"
)

func (rest *RestOperations) RestOperationForEvh(vsName string, namespace string, avimodel *nodes.AviObjectGraph, sniNode bool, vs_cache_obj *avicache.AviVsCache, key string) {
	var pools_to_delete []avicache.NamespaceName
	var pgs_to_delete []avicache.NamespaceName
	var vsvip_to_delete []avicache.NamespaceName
	var sni_to_delete []avicache.NamespaceName
	var httppol_to_delete []avicache.NamespaceName
	var l4pol_to_delete []avicache.NamespaceName
	var sslkey_cert_delete []avicache.NamespaceName
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
	// Order would be this: 1. Pools 2. PGs  3. DS. 4. SSLKeyCert 5. VS
	if vs_cache_obj != nil {
		var rest_ops []*utils.RestOp
		vsvip_to_delete, rest_ops, vsvipErr = rest.VSVipCU(aviVsNode.VSVIPRefs, vs_cache_obj, namespace, rest_ops, key)
		if vsvipErr != nil {
			if rest.CheckAndPublishForRetry(vsvipErr, publishKey, key, avimodel) {
				return
			}
		}
		sslkey_cert_delete, rest_ops = rest.CACertCU(aviVsNode.CACertRefs, vs_cache_obj.SSLKeyCertCollection, namespace, rest_ops, key)
		// SSLKeyCertCollection which did not match cacerts are present in the list sslkey_cert_delete,
		// which shuld be the new SSLKeyCertCollection
		sslkey_cert_delete, rest_ops = rest.SSLKeyCertCU(aviVsNode.SSLKeyCertRefs, sslkey_cert_delete, namespace, rest_ops, key)
		pools_to_delete, rest_ops = rest.PoolCU(aviVsNode.PoolRefs, vs_cache_obj, namespace, rest_ops, key)
		pgs_to_delete, rest_ops = rest.PoolGroupCU(aviVsNode.PoolGroupRefs, vs_cache_obj, namespace, rest_ops, key)
		httppol_to_delete, rest_ops = rest.HTTPPolicyCU(aviVsNode.HttpPolicyRefs, vs_cache_obj, namespace, rest_ops, key)
		// ds_to_delete, rest_ops = rest.DatascriptCU(aviVsNode.HTTPDSrefs, vs_cache_obj, namespace, rest_ops, key)
		utils.AviLog.Debugf("key: %s, msg: stored checksum for VS: %s, model checksum: %s", key, vs_cache_obj.CloudConfigCksum, strconv.Itoa(int(aviVsNode.GetCheckSum())))
		if vs_cache_obj.CloudConfigCksum == strconv.Itoa(int(aviVsNode.GetCheckSum())) {
			utils.AviLog.Debugf("key: %s, msg: the checksums are same for vs %s, not doing anything", key, vs_cache_obj.Name)
		} else {
			utils.AviLog.Debugf("key: %s, msg: the stored checksum for vs is %v, and the obtained checksum for VS is: %v", key, vs_cache_obj.CloudConfigCksum, strconv.Itoa(int(aviVsNode.GetCheckSum())))
			// The checksums are different, so it should be a PUT call.
			restOp := rest.AviVsBuildForEvh(aviVsNode, utils.RestPut, vs_cache_obj, key)
			rest_ops = append(rest_ops, restOp...)

		}
		if success := rest.ExecuteRestAndPopulateCacheForEvh(rest_ops, vsKey, avimodel, key); !success {
			return
		}
	} else {
		var rest_ops []*utils.RestOp
		_, rest_ops, vsvipErr = rest.VSVipCU(aviVsNode.VSVIPRefs, nil, namespace, rest_ops, key)
		if vsvipErr != nil {
			if rest.CheckAndPublishForRetry(vsvipErr, publishKey, key, avimodel) {
				return
			}
		}
		_, rest_ops = rest.CACertCU(aviVsNode.CACertRefs, []avicache.NamespaceName{}, namespace, rest_ops, key)
		_, rest_ops = rest.SSLKeyCertCU(aviVsNode.SSLKeyCertRefs, nil, namespace, rest_ops, key)
		_, rest_ops = rest.PoolCU(aviVsNode.PoolRefs, nil, namespace, rest_ops, key)
		_, rest_ops = rest.PoolGroupCU(aviVsNode.PoolGroupRefs, nil, namespace, rest_ops, key)
		_, rest_ops = rest.HTTPPolicyCU(aviVsNode.HttpPolicyRefs, nil, namespace, rest_ops, key)

		// The cache was not found - it's a POST call.
		restOp := rest.AviVsBuildForEvh(aviVsNode, utils.RestPost, nil, key)
		rest_ops = append(rest_ops, restOp...)
		utils.AviLog.Debugf("POST key: %s, vsKey: %s", key, vsKey)
		utils.AviLog.Debugf("POST restops %s", utils.Stringify(rest_ops))
		if success := rest.ExecuteRestAndPopulateCacheForEvh(rest_ops, vsKey, avimodel, key); !success {
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
	rest_ops = rest.L4PolicyDelete(l4pol_to_delete, namespace, rest_ops, key)
	rest_ops = rest.PoolGroupDelete(pgs_to_delete, namespace, rest_ops, key)
	rest_ops = rest.PoolDelete(pools_to_delete, namespace, rest_ops, key)
	if success := rest.ExecuteRestAndPopulateCacheForEvh(rest_ops, vsKey, avimodel, key); !success {
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
		if success := rest.ExecuteRestAndPopulateCacheForEvh(evh_rest_ops, vsKey, avimodel, key); !success {
			return
		}
	}

	// Let's populate all the DELETE entries
	if len(sni_to_delete) > 0 {
		utils.AviLog.Infof("key: %s, msg: EVH delete candidates are : %s", key, sni_to_delete)
		var rest_ops []*utils.RestOp
		for _, del_sni := range sni_to_delete {
			rest.SNINodeDelete(del_sni, namespace, rest_ops, avimodel, key)
			if success := rest.ExecuteRestAndPopulateCacheForEvh(rest_ops, vsKey, avimodel, key); !success {
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
	if vs_cache_obj != nil {
		sni_key := avicache.NamespaceName{Namespace: namespace, Name: sni_node.Name}
		// Search the VS cache and obtain the UUID of this VS. Then see if this UUID is part of the SNIChildCollection or not.
		found := utils.HasElem(cache_sni_nodes, sni_key)
		utils.AviLog.Debugf("key: %s, msg: processing node key: %v", key, sni_key)
		if found && cache_sni_nodes != nil {
			cache_sni_nodes = Remove(cache_sni_nodes, sni_key)
			utils.AviLog.Debugf("key: %s, msg: the cache sni nodes are: %v", key, cache_sni_nodes)
			sni_cache_obj := rest.getVsCacheObj(sni_key, key)
			if sni_cache_obj != nil {
				// CAcerts have to be created first, as they are referred by the keycerts
				sslkey_cert_delete, rest_ops = rest.CACertCU(sni_node.CACertRefs, sni_cache_obj.SSLKeyCertCollection, namespace, rest_ops, key)
				// SSLKeyCertCollection which did not match cacerts are present in the list sslkey_cert_delete,
				// which shuld be the new SSLKeyCertCollection
				sslkey_cert_delete, rest_ops = rest.SSLKeyCertCU(sni_node.SSLKeyCertRefs, sslkey_cert_delete, namespace, rest_ops, key)
				sni_pools_to_delete, rest_ops = rest.PoolCU(sni_node.PoolRefs, sni_cache_obj, namespace, rest_ops, key)
				sni_pgs_to_delete, rest_ops = rest.PoolGroupCU(sni_node.PoolGroupRefs, sni_cache_obj, namespace, rest_ops, key)
				http_policies_to_delete, rest_ops = rest.HTTPPolicyCU(sni_node.HttpPolicyRefs, sni_cache_obj, namespace, rest_ops, key)

				// The checksums are different, so it should be a PUT call.
				if sni_cache_obj.CloudConfigCksum != strconv.Itoa(int(sni_node.GetCheckSum())) {
					restOp := rest.AviVsBuildForEvh(sni_node, utils.RestPut, sni_cache_obj, key)
					rest_ops = append(rest_ops, restOp...)
					utils.AviLog.Infof("key: %s, msg: the checksums are different for sni child %s, operation: PUT", key, sni_node.Name)

				}
			}
		} else {
			utils.AviLog.Debugf("key: %s, msg: sni child %s not found in cache, operation: POST", key, sni_node.Name)
			_, rest_ops = rest.CACertCU(sni_node.CACertRefs, []avicache.NamespaceName{}, namespace, rest_ops, key)
			_, rest_ops = rest.SSLKeyCertCU(sni_node.SSLKeyCertRefs, nil, namespace, rest_ops, key)
			_, rest_ops = rest.PoolCU(sni_node.PoolRefs, nil, namespace, rest_ops, key)
			_, rest_ops = rest.PoolGroupCU(sni_node.PoolGroupRefs, nil, namespace, rest_ops, key)
			_, rest_ops = rest.HTTPPolicyCU(sni_node.HttpPolicyRefs, nil, namespace, rest_ops, key)

			// Not found - it should be a POST call.
			restOp := rest.AviVsBuildForEvh(sni_node, utils.RestPost, nil, key)
			rest_ops = append(rest_ops, restOp...)
		}
		rest_ops = rest.SSLKeyCertDelete(sslkey_cert_delete, namespace, rest_ops, key)
		rest_ops = rest.HTTPPolicyDelete(http_policies_to_delete, namespace, rest_ops, key)
		rest_ops = rest.PoolGroupDelete(sni_pgs_to_delete, namespace, rest_ops, key)
		rest_ops = rest.PoolDelete(sni_pools_to_delete, namespace, rest_ops, key)
		utils.AviLog.Debugf("key: %s, msg: the SNI VSes to be deleted are: %s", key, cache_sni_nodes)
	} else {
		utils.AviLog.Debugf("key: %s, msg: sni child %s not found in cache and SNI parent also does not exist in cache", key, sni_node.Name)
		_, rest_ops = rest.CACertCU(sni_node.CACertRefs, []avicache.NamespaceName{}, namespace, rest_ops, key)
		_, rest_ops = rest.SSLKeyCertCU(sni_node.SSLKeyCertRefs, nil, namespace, rest_ops, key)
		_, rest_ops = rest.PoolCU(sni_node.PoolRefs, nil, namespace, rest_ops, key)
		_, rest_ops = rest.PoolGroupCU(sni_node.PoolGroupRefs, nil, namespace, rest_ops, key)
		_, rest_ops = rest.HTTPPolicyCU(sni_node.HttpPolicyRefs, nil, namespace, rest_ops, key)

		// Not found - it should be a POST call.
		restOp := rest.AviVsBuildForEvh(sni_node, utils.RestPost, nil, key)
		rest_ops = append(rest_ops, restOp...)
	}
	return cache_sni_nodes, rest_ops
}

func (rest *RestOperations) AviVsBuildForEvh(vs_meta *nodes.AviEvhVsNode, rest_method utils.RestMethod, cache_obj *avicache.AviVsCache, key string) []*utils.RestOp {
	if !vs_meta.EVHParent {
		rest_ops := rest.AviVsEvhBuildForEvh(vs_meta, rest_method, cache_obj, key)
		return rest_ops
	} else {
		// This is EVH Parent
		network_prof := "/api/networkprofile/?name=" + vs_meta.NetworkProfile
		app_prof := "/api/applicationprofile/?name=" + vs_meta.ApplicationProfile
		name := vs_meta.Name
		cksum := vs_meta.CloudConfigCksum
		checksumstr := strconv.Itoa(int(cksum))
		cr := lib.AKOUser
		cloudRef := "/api/cloud?name=" + utils.CloudName
		svc_mdata_json, _ := json.Marshal(&vs_meta.ServiceMetadata)
		svc_mdata := string(svc_mdata_json)
		vrfContextRef := "/api/vrfcontext?name=" + vs_meta.VrfContext
		seGroupRef := "/api/serviceenginegroup?name=" + lib.GetSEGName()
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
			VhType:                proto.String(utils.VS_TYPE_VH_ENHANCED),
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

		if vs_meta.EVHParent {
			// This is a EVH parent
			utils.AviLog.Debugf("key: %s, msg: vs %s is a EVH Parent", key, vs_meta.Name)
			vh_parent := utils.VS_TYPE_VH_PARENT
			vs.Type = &vh_parent
		}
		// TODO other fields like cloud_ref, mix of TCP & UDP protocols, etc.

		for i, pp := range vs_meta.PortProto {
			port := pp.Port
			svc := avimodels.Service{Port: &port, EnableSsl: &vs_meta.PortProto[i].EnableSSL}
			vs.Services = append(vs.Services, &svc)
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

		if vs_meta.TLSType != utils.TLS_PASSTHROUGH {
			// this overwrites the sslkeycert created from the Secret object, with the one mentioned in HostRule.TLS
			if vs_meta.SSLKeyCertAviRef != "" {
				vs.SslKeyAndCertificateRefs = append(vs.SslKeyAndCertificateRefs, vs_meta.SSLKeyCertAviRef)
			} else {
				for _, sslkeycert := range vs_meta.SSLKeyCertRefs {
					certName := "/api/sslkeyandcertificate/?name=" + sslkeycert.Name
					vs.SslKeyAndCertificateRefs = append(vs.SslKeyAndCertificateRefs, certName)
				}
			}

		}

		if strings.Contains(*vs.Name, lib.PassthroughPrefix) && !strings.HasSuffix(*vs.Name, lib.PassthroughInsecure) {
			// This is a passthrough secure VS, we want the VS to be down if all the pools are down.
			vsDownOnPoolDown := true
			vs.RemoveListeningPortOnVsDown = &vsDownOnPoolDown
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

func (rest *RestOperations) AviVsEvhBuildForEvh(vs_meta *nodes.AviEvhVsNode, rest_method utils.RestMethod, cache_obj *avicache.AviVsCache, key string) []*utils.RestOp {
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
	evhChild := &avimodels.VirtualService{
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
		VhType:                proto.String(utils.VS_TYPE_VH_ENHANCED),
	}

	//This VS has a TLSKeyCert associated, we need to mark 'type': 'VS_TYPE_VH_PARENT'
	vh_type := utils.VS_TYPE_VH_CHILD
	evhChild.Type = &vh_type
	vhParentUuid := "/api/virtualservice/?name=" + vs_meta.VHParentName
	evhChild.VhParentVsUUID = &vhParentUuid
	// evhChild.VhDomainName = vs_meta.VHDomainNames
	ignPool := false
	evhChild.IgnPoolNetReach = &ignPool
	// Fill vhmatch information
	match_case := "SENSITIVE"
	matchCriteria := "BEGINS_WITH"
	path_match := avimodels.PathMatch{
		MatchCriteria: &matchCriteria,
		MatchCase:     &match_case,
		MatchStr:      []string{"/"},
	}
	pathMatches := make([]*avimodels.PathMatch, 0)
	pathMatches = append(pathMatches, &path_match)
	vhMatch := &avimodels.VHMatch{Host: &vs_meta.EvhHostName, Path: pathMatches}
	vhMatches := make([]*avimodels.VHMatch, 0)
	vhMatches = append(vhMatches, vhMatch)
	evhChild.VhMatches = vhMatches

	if vs_meta.DefaultPool != "" {
		pool_ref := "/api/pool/?name=" + vs_meta.DefaultPool
		evhChild.PoolRef = &pool_ref
	}
	if len(vs_meta.PoolGroupRefs) > 0 {
		poolgroup_ref := "/api/poolgroup/?name=" + vs_meta.PoolGroupRefs[0].Name
		evhChild.PoolGroupRef = &poolgroup_ref
	}

	// No need of HTTP rules for TLS passthrough.
	if vs_meta.TLSType != utils.TLS_PASSTHROUGH {
		// this overwrites the sslkeycert created from the Secret object, with the one mentioned in HostRule.TLS
		if vs_meta.SSLKeyCertAviRef != "" {
			evhChild.SslKeyAndCertificateRefs = append(evhChild.SslKeyAndCertificateRefs, vs_meta.SSLKeyCertAviRef)
		} else {
			for _, sslkeycert := range vs_meta.SSLKeyCertRefs {
				certName := "/api/sslkeyandcertificate/?name=" + sslkeycert.Name
				evhChild.SslKeyAndCertificateRefs = append(evhChild.SslKeyAndCertificateRefs, certName)
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

		evhChild.HTTPPolicies = httpPolicyCollection
	}

	var rest_ops []*utils.RestOp
	var rest_op utils.RestOp
	var path string
	if rest_method == utils.RestPut {

		path = "/api/virtualservice/" + cache_obj.Uuid
		rest_op = utils.RestOp{Path: path, Method: rest_method, Obj: evhChild,
			Tenant: vs_meta.Tenant, Model: "VirtualService", Version: utils.CtrlVersion}
		rest_ops = append(rest_ops, &rest_op)

	} else {

		macro := utils.AviRestObjMacro{ModelName: "VirtualService", Data: evhChild}
		path = "/api/macro"
		rest_op = utils.RestOp{Path: path, Method: rest_method, Obj: macro,
			Tenant: vs_meta.Tenant, Model: "VirtualService", Version: utils.CtrlVersion}
		rest_ops = append(rest_ops, &rest_op)

	}

	return rest_ops
}

func (rest *RestOperations) ExecuteRestAndPopulateCacheForEvh(rest_ops []*utils.RestOp, aviObjKey avicache.NamespaceName, avimodel *nodes.AviObjectGraph, key string, sslKey ...utils.NamespaceName) bool {
	// Choose a avi client based on the model name hash. This would ensure that the same worker queue processes updates for a given VS all the time.
	shardSize := lib.GetshardSize()
	var retry, fastRetry bool
	if shardSize != 0 {
		bkt := utils.Bkt(key, shardSize)
		if len(rest.aviRestPoolClient.AviClient) > 0 && len(rest_ops) > 0 {
			utils.AviLog.Infof("key: %s, msg: processing in rest queue number: %v", key, bkt)
			aviclient := rest.aviRestPoolClient.AviClient[bkt]
			err := rest.AviRestOperateWrapper(aviclient, rest_ops)
			if err == nil {
				models.RestStatus.UpdateAviApiRestStatus(utils.AVIAPI_CONNECTED, nil)
				utils.AviLog.Debugf("key: %s, msg: rest call executed successfully, will update cache", key)

				// Add to local obj caches
				for _, rest_op := range rest_ops {
					rest.PopulateOneCache(rest_op, aviObjKey, key)
				}

			} else if aviObjKey.Name == lib.DummyVSForStaleData {
				utils.AviLog.Warnf("key: %s, msg: error in rest request %v, for %s, won't retry", key, err.Error(), lib.DummyVSForStaleData)
				return false
			} else {
				var publishKey string
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

				if rest.CheckAndPublishForRetry(err, publishKey, key, avimodel) {
					return false
				}
				utils.AviLog.Warnf("key: %s, msg: there was an error sending the macro %v", key, err.Error())
				models.RestStatus.UpdateAviApiRestStatus("", err)
				for i := len(rest_ops) - 1; i >= 0; i-- {
					// Go over each of the failed requests and enqueue them to the worker queue for retry.
					if rest_ops[i].Err != nil {
						// check for VSVIP errors for blocked IP address updates
						if lib.GetAdvancedL4() && checkVsVipUpdateErrors(key, rest_ops[i]) {
							rest.PopulateOneCache(rest_ops[i], aviObjKey, key)
							continue
						}

						// If it's for a SNI child, publish the parent VS's key
						if avimodel != nil && len(avimodel.GetAviEvhVS()) > 0 {
							utils.AviLog.Warnf("key: %s, msg: Retrieved key for Retry:%s, object: %s", key, publishKey, rest_ops[i].ObjName)
							if avimodel.GetRetryCounter() != 0 {
								aviError, ok := rest_ops[i].Err.(session.AviError)
								if !ok {
									utils.AviLog.Infof("key: %s, msg: Error is not of type AviError, err: %v, %T", key, rest_ops[i].Err, rest_ops[i].Err)
									continue
								}
								retryable, fastRetryable := rest.RefreshCacheForRetryLayerForEvh(publishKey, aviObjKey, rest_ops[i], aviError, aviclient, avimodel, key)
								fastRetry = fastRetry || fastRetryable
								retry = retry || retryable
							} else {
								utils.AviLog.Warnf("key: %s, msg: retry count exhausted, skipping", key)
							}
						} else {
							utils.AviLog.Warnf("key: %s, msg: Avi model not set, possibly a DELETE call", key)
							aviError, ok := rest_ops[i].Err.(session.AviError)
							// If it's 404, don't retry
							if ok {
								statuscode := aviError.HttpStatusCode
								if statuscode != 404 {
									rest.PublishKeyToSlowRetryLayer(publishKey, key)
									return false
								} else {
									rest.AviVsCacheDel(rest_ops[i], aviObjKey, key)
								}
							}
						}
					} else {
						rest.PopulateOneCache(rest_ops[i], aviObjKey, key)
					}
				}

				if retry {
					rest.PublishKeyToRetryLayer(publishKey, key)
				}
				return false
			}
		}
	}
	return true
}

func (rest *RestOperations) RefreshCacheForRetryLayerForEvh(parentVsKey string, aviObjKey avicache.NamespaceName, rest_op *utils.RestOp, aviError session.AviError, c *clients.AviClient, avimodel *nodes.AviObjectGraph, key string) (bool, bool) {
	var fastRetry bool
	statuscode := aviError.HttpStatusCode
	errorStr := aviError.Error()
	retry := true
	utils.AviLog.Warnf("key: %s, msg: problem in processing request for: %s", key, rest_op.Model)
	utils.AviLog.Infof("key: %s, msg: error str: %s", key, errorStr)
	aviObjCache := avicache.SharedAviObjCache()

	if statuscode >= 500 && statuscode < 599 {
		fastRetry = true
	} else if statuscode >= 400 && statuscode < 499 { // Will account for more error codes.*/
		fastRetry = true
		// 404 means the object exists in our cache but not on the controller.
		if statuscode == 404 {
			switch rest_op.Model {
			case "Pool":
				var poolObjName string
				switch rest_op.Obj.(type) {
				case utils.AviRestObjMacro:
					poolObjName = *rest_op.Obj.(utils.AviRestObjMacro).Data.(avimodels.Pool).Name
				case avimodels.Pool:
					poolObjName = *rest_op.Obj.(avimodels.Pool).Name
				}
				rest_op.ObjName = poolObjName
				rest.AviPoolCacheDel(rest_op, aviObjKey, key)
			case "PoolGroup":
				var pgObjName string
				switch rest_op.Obj.(type) {
				case utils.AviRestObjMacro:
					pgObjName = *rest_op.Obj.(utils.AviRestObjMacro).Data.(avimodels.PoolGroup).Name
				case avimodels.PoolGroup:
					pgObjName = *rest_op.Obj.(avimodels.PoolGroup).Name
				}
				rest_op.ObjName = pgObjName
				if strings.Contains(errorStr, "Pool object not found!") {
					// PG error with pool object not found.
					aviObjCache.AviPopulateOnePGCache(c, utils.CloudName, pgObjName)
					// After the refresh - get the members
					pgKey := avicache.NamespaceName{Namespace: lib.GetTenant(), Name: pgObjName}
					pgCache, ok := rest.cache.PgCache.AviCacheGet(pgKey)
					if ok {
						pgCacheObj, _ := pgCache.(*avicache.AviPGCache)
						// Iterate the pools
						vsNode := avimodel.GetAviEvhVS()[0]
						var pools []string
						for _, pgNode := range vsNode.PoolGroupRefs {
							if pgNode.Name == pgObjName {
								for _, poolInModel := range pgNode.Members {
									poolToken := strings.Split(*poolInModel.PoolRef, "?name=")
									if len(poolToken) > 1 {
										pools = append(pools, poolToken[1])
									}
								}
							}
						}
						utils.AviLog.Debugf("key: %s, msg: pools in model during retry: %s", key, pools)
						// Find out pool members that exist in the model but do not exist in the cache and delete them.

						poolsCopy := make([]string, len(pools))
						copy(poolsCopy, pools)
						for _, poolName := range pgCacheObj.Members {
							if utils.HasElem(pools, poolName) {
								poolsCopy = utils.Remove(poolsCopy, poolName)
							}
						}
						// Whatever is left it in poolsCopy - remove them from the avi pools cache
						for _, poolsToDel := range poolsCopy {
							rest_op.ObjName = poolsToDel
							utils.AviLog.Debugf("key: %s, msg: deleting pool from cache due to pool not found %s", key, poolsToDel)
							rest.AviPoolCacheDel(rest_op, aviObjKey, key)
						}
					} else {
						utils.AviLog.Infof("key: %s, msg: PG object not found during retry pgname: %s", key, pgObjName)
					}
				}
				rest.AviPGCacheDel(rest_op, aviObjKey, key)
			case "VsVip":
				var VsVip string
				switch rest_op.Obj.(type) {
				case utils.AviRestObjMacro:
					VsVip = *rest_op.Obj.(utils.AviRestObjMacro).Data.(avimodels.VsVip).Name
				case avimodels.VsVip:
					VsVip = *rest_op.Obj.(avimodels.VsVip).Name
				}
				rest_op.ObjName = VsVip
				rest.AviVsVipCacheDel(rest_op, aviObjKey, key)
			case "HTTPPolicySet":
				var HTTPPolicySet string
				switch rest_op.Obj.(type) {
				case utils.AviRestObjMacro:
					HTTPPolicySet = *rest_op.Obj.(utils.AviRestObjMacro).Data.(avimodels.HTTPPolicySet).Name
				case avimodels.HTTPPolicySet:
					HTTPPolicySet = *rest_op.Obj.(avimodels.HTTPPolicySet).Name
				}
				rest_op.ObjName = HTTPPolicySet
				rest.AviHTTPPolicyCacheDel(rest_op, aviObjKey, key)
			case "L4PolicySet":
				var L4PolicySet string
				switch rest_op.Obj.(type) {
				case utils.AviRestObjMacro:
					L4PolicySet = *rest_op.Obj.(utils.AviRestObjMacro).Data.(avimodels.L4PolicySet).Name
				case avimodels.L4PolicySet:
					L4PolicySet = *rest_op.Obj.(avimodels.L4PolicySet).Name
				}
				rest_op.ObjName = L4PolicySet
				rest.AviL4PolicyCacheDel(rest_op, aviObjKey, key)
			case "SSLKeyAndCertificate":
				var SSLKeyAndCertificate string
				switch rest_op.Obj.(type) {
				case utils.AviRestObjMacro:
					SSLKeyAndCertificate = *rest_op.Obj.(utils.AviRestObjMacro).Data.(avimodels.SSLKeyAndCertificate).Name
				case avimodels.SSLKeyAndCertificate:
					SSLKeyAndCertificate = *rest_op.Obj.(avimodels.SSLKeyAndCertificate).Name
				}
				rest_op.ObjName = SSLKeyAndCertificate
				rest.AviSSLCacheDel(rest_op, aviObjKey, key)
			case "PKIprofile":
				var PKIprofile string
				switch rest_op.Obj.(type) {
				case utils.AviRestObjMacro:
					PKIprofile = *rest_op.Obj.(utils.AviRestObjMacro).Data.(avimodels.PKIprofile).Name
				case avimodels.PKIprofile:
					PKIprofile = *rest_op.Obj.(avimodels.PKIprofile).Name
				}
				rest_op.ObjName = PKIprofile
				rest.AviPkiProfileCacheDel(rest_op, aviObjKey, key)
			case "VirtualService":
				rest.AviVsCacheDel(rest_op, aviObjKey, key)
			case "VSDataScriptSet":
				var VSDataScriptSet string
				switch rest_op.Obj.(type) {
				case utils.AviRestObjMacro:
					VSDataScriptSet = *rest_op.Obj.(utils.AviRestObjMacro).Data.(avimodels.VSDataScriptSet).Name
				case avimodels.VSDataScriptSet:
					VSDataScriptSet = *rest_op.Obj.(avimodels.VSDataScriptSet).Name
				}
				rest_op.ObjName = VSDataScriptSet
				rest.AviDSCacheDel(rest_op, aviObjKey, key)
			}
		} else if statuscode == 409 {

			// TODO (sudswas): if error code 400 happens, it means layer 2's model has issue - can re-trigger a model eval in that case?
			// If it's 409 it refers to a conflict. That means the cache should be refreshed for the particular object.

			utils.AviLog.Infof("key: %s, msg: Confict for object: %s of type :%s", key, rest_op.ObjName, rest_op.Model)
			switch rest_op.Model {
			case "Pool":
				var poolObjName string
				switch rest_op.Obj.(type) {
				case utils.AviRestObjMacro:
					poolObjName = *rest_op.Obj.(utils.AviRestObjMacro).Data.(avimodels.Pool).Name
				case avimodels.Pool:
					poolObjName = *rest_op.Obj.(avimodels.Pool).Name
				}
				aviObjCache.AviPopulateOnePoolCache(c, utils.CloudName, poolObjName)
			case "PoolGroup":
				var pgObjName string
				switch rest_op.Obj.(type) {
				case utils.AviRestObjMacro:
					pgObjName = *rest_op.Obj.(utils.AviRestObjMacro).Data.(avimodels.PoolGroup).Name
				case avimodels.PoolGroup:
					pgObjName = *rest_op.Obj.(avimodels.PoolGroup).Name
				}
				aviObjCache.AviPopulateOnePGCache(c, utils.CloudName, pgObjName)
			case "VsVip":
				var VsVip string
				switch rest_op.Obj.(type) {
				case utils.AviRestObjMacro:
					VsVip = *rest_op.Obj.(utils.AviRestObjMacro).Data.(avimodels.VsVip).Name
				case avimodels.VsVip:
					VsVip = *rest_op.Obj.(avimodels.VsVip).Name
				}
				aviObjCache.AviPopulateOneVsVipCache(c, utils.CloudName, VsVip)
			case "HTTPPolicySet":
				var HTTPPolicySet string
				switch rest_op.Obj.(type) {
				case utils.AviRestObjMacro:
					HTTPPolicySet = *rest_op.Obj.(utils.AviRestObjMacro).Data.(avimodels.HTTPPolicySet).Name
				case avimodels.HTTPPolicySet:
					HTTPPolicySet = *rest_op.Obj.(avimodels.HTTPPolicySet).Name
				}
				aviObjCache.AviPopulateOneVsHttpPolCache(c, utils.CloudName, HTTPPolicySet)
			case "L4PolicySet":
				var L4PolicySet string
				switch rest_op.Obj.(type) {
				case utils.AviRestObjMacro:
					L4PolicySet = *rest_op.Obj.(utils.AviRestObjMacro).Data.(avimodels.L4PolicySet).Name
				case avimodels.L4PolicySet:
					L4PolicySet = *rest_op.Obj.(avimodels.L4PolicySet).Name
				}
				aviObjCache.AviPopulateOneVsL4PolCache(c, utils.CloudName, L4PolicySet)
			case "SSLKeyAndCertificate":
				var SSLKeyAndCertificate string
				switch rest_op.Obj.(type) {
				case utils.AviRestObjMacro:
					SSLKeyAndCertificate = *rest_op.Obj.(utils.AviRestObjMacro).Data.(avimodels.SSLKeyAndCertificate).Name
				case avimodels.SSLKeyAndCertificate:
					SSLKeyAndCertificate = *rest_op.Obj.(avimodels.SSLKeyAndCertificate).Name
				}
				aviObjCache.AviPopulateOneSSLCache(c, utils.CloudName, SSLKeyAndCertificate)
			case "PKIprofile":
				var PKIprofile string
				switch rest_op.Obj.(type) {
				case utils.AviRestObjMacro:
					PKIprofile = *rest_op.Obj.(utils.AviRestObjMacro).Data.(avimodels.PKIprofile).Name
				case avimodels.PKIprofile:
					PKIprofile = *rest_op.Obj.(avimodels.PKIprofile).Name
				}
				aviObjCache.AviPopulateOnePKICache(c, utils.CloudName, PKIprofile)
			case "VirtualService":
				aviObjCache.AviObjOneVSCachePopulate(c, utils.CloudName, aviObjKey.Name)
				vsObjMeta, ok := rest.cache.VsCacheMeta.AviCacheGet(aviObjKey)
				if !ok {
					// Object deleted
					utils.AviLog.Warnf("key: %s, msg: VS object already deleted during retry", key)
				} else {
					vsCopy, done := vsObjMeta.(*avicache.AviVsCache).GetVSCopy()
					if done {
						rest.cache.VsCacheMeta.AviCacheAdd(aviObjKey, vsCopy)
					}
				}
			case "VSDataScriptSet":
				var VSDataScriptSet string
				switch rest_op.Obj.(type) {
				case utils.AviRestObjMacro:
					VSDataScriptSet = *rest_op.Obj.(utils.AviRestObjMacro).Data.(avimodels.VSDataScriptSet).Name
				case avimodels.VSDataScript:
					VSDataScriptSet = *rest_op.Obj.(avimodels.VSDataScriptSet).Name
				}
				aviObjCache.AviPopulateOneVsDSCache(c, utils.CloudName, VSDataScriptSet)
			}
		} else if statuscode == 408 {
			// This status code refers to a problem with the controller timeouts. We need to re-init the session object.
			utils.AviLog.Infof("key :%s, msg: Controller request timed out, will re-init session by retrying", key)

		} else {

			// We don't want to handle any other error code like 400 etc.
			utils.AviLog.Infof("key: %s, msg: Detected error code %d that we don't support, not going to retry", key, statuscode)
			retry = false
		}
	}

	return retry, fastRetry
}
