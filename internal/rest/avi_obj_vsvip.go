/*
 * Copyright 2020-2021 VMware, Inc.
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

func (rest *RestOperations) AviVsVipBuild(vsvip_meta *nodes.AviVSVIPNode, cache_obj *avicache.AviVSVIPCache, key string) (*utils.RestOp, error) {
	name := vsvip_meta.Name
	tenant := fmt.Sprintf("/api/tenant/?name=%s", vsvip_meta.Tenant)
	cloudRef := "/api/cloud?name=" + utils.CloudName
	var dns_info_arr []*avimodels.DNSInfo
	var path string
	var rest_op utils.RestOp
	vipId, ipType := "0", "V4"

	cksum := vsvip_meta.CloudConfigCksum
	cksumstr := strconv.Itoa(int(cksum))

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
		vsvip.VsvipCloudConfigCksum = &cksumstr

		// handling static IP and networkName (infraSetting) updates.
		vip := &avimodels.Vip{
			VipID:          &vipId,
			AutoAllocateIP: &autoAllocate,
		}

		// This would throw an error for advl4 the error is propagated to the gateway status.
		if vsvip_meta.IPAddress != "" {
			vip.IPAddress = &avimodels.IPAddr{Type: &ipType, Addr: &vsvip_meta.IPAddress}
		}
		if lib.IsPublicCloud() && lib.GetCloudType() != lib.CLOUD_GCP {
			vips := networkNamesToVips(vsvip_meta.NetworkNames)
			vsvip.Vip = []*avimodels.Vip{}
			vsvip.Vip = append(vsvip.Vip, vips...)
		} else {
			// Set the IPAM network subnet for all clouds except AWS and Azure
			if len(vsvip_meta.NetworkNames) != 0 {
				if vip.IPAMNetworkSubnet == nil {
					vip.IPAMNetworkSubnet = &avimodels.IPNetworkSubnet{}
				}
				networkRef := "/api/network/?name=" + vsvip_meta.NetworkNames[0]
				vip.IPAMNetworkSubnet.NetworkRef = &networkRef
				vsvip.Vip = []*avimodels.Vip{vip}
			}
		}

		// Override the Vip in VsVip tto bring in updates, keeping everything else as is.

		if lib.GetEnableCtrl2014Features() {
			vsvip.Labels = lib.GetLabels()
		}
		rest_op = utils.RestOp{
			ObjName: name,
			Path:    "/api/vsvip/" + cache_obj.Uuid,
			Method:  utils.RestPut,
			Obj:     vsvip,
			Tenant:  vsvip_meta.Tenant,
			Model:   "VsVip",
			Version: utils.CtrlVersion,
		}
	} else {
		var vips []*avimodels.Vip
		vip := avimodels.Vip{
			VipID:          &vipId,
			AutoAllocateIP: &autoAllocate,
		}

		// setting IPAMNetworkSubnet.Subnet value in case subnetCIDR is provided
		subnetMask := lib.GetSubnetPrefixInt()
		subnetAddress := lib.GetSubnetIP()
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
		if lib.IsPublicCloud() && lib.GetCloudType() != lib.CLOUD_GCP {
			vips = networkNamesToVips(vsvip_meta.NetworkNames)
		} else {
			// Set the IPAM network subnet for all clouds except AWS and Azure
			if len(vsvip_meta.NetworkNames) != 0 {
				if vip.IPAMNetworkSubnet == nil {
					vip.IPAMNetworkSubnet = &avimodels.IPNetworkSubnet{}
				}
				networkRef := "/api/network/?name=" + vsvip_meta.NetworkNames[0]
				vip.IPAMNetworkSubnet.NetworkRef = &networkRef
			}
		}

		if len(vips) == 0 {
			vips = append(vips, &vip)
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
			Vip:               vips,
		}
		if lib.GetEnableCtrl2014Features() {
			vsvip.Labels = lib.GetLabels()
		}
		vsvip.VsvipCloudConfigCksum = &cksumstr
		path = "/api/vsvip"
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
					rest_op = utils.RestOp{Path: path, Method: utils.RestPost, Obj: vsvip,
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
			if lib.GetEnableCtrl2014Features() {
				vsvip_avi.Labels = lib.GetLabels()
			}
			vsvip_avi.VsvipCloudConfigCksum = &cksumstr
			path = "/api/vsvip/" + vsvip_cache_obj.Uuid
			rest_op = utils.RestOp{
				ObjName: name,
				Path:    path,
				Method:  utils.RestPut,
				Obj:     vsvip_avi,
				Tenant:  vsvip_meta.Tenant,
				Model:   "VsVip",
				Version: utils.CtrlVersion,
			}
		} else {
			rest_op = utils.RestOp{
				ObjName: name,
				Path:    path,
				Method:  utils.RestPost,
				Obj:     vsvip,
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
	uri := "/api/vsvip/" + uuid + "/?include_name"

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
			utils.AviLog.Warnf("key: %s, rest_op has err or no response for vsvip err: %v, response: %v", key, rest_op.Err, rest_op.Response)
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
		var networkNames []string
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
					if ipamNetworkSubnet, ipamOk := vip["ipam_network_subnet"].(map[string]interface{}); ipamOk {
						if networkRef, netRefOk := ipamNetworkSubnet["network_ref"].(string); netRefOk {
							if networkRefName := strings.Split(networkRef, "#"); len(networkRefName) == 2 {
								networkNames = append(networkNames, strings.Split(networkRef, "#")[1])
							}
							if lib.GetCloudType() == lib.CLOUD_AWS {
								networkRefNameSplit := strings.Split(networkRef, "/")
								networkNames = append(networkNames, networkRefNameSplit[len(networkRefNameSplit)-1])
							}
						}
					}
				}
			}
		}

		vsvip_cache_obj := avicache.AviVSVIPCache{
			Name:         name,
			Tenant:       rest_op.Tenant,
			Uuid:         uuid,
			LastModified: lastModifiedStr,
			FQDNs:        vsvipFQDNs,
			Vips:         vsvipVips,
			NetworkNames: networkNames,
		}

		if lastModifiedStr == "" {
			vsvip_cache_obj.InvalidData = true
		}

		if resp["VsvipCloudConfigCksum"] != nil {
			vsvip_cache_obj.CloudConfigCksum = resp["VsvipCloudConfigCksum"].(string)
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

func networkNamesToVips(networkNames []string) []*avimodels.Vip {
	var vipList []*avimodels.Vip
	autoAllocate := true
	for vipIDInt, networkName := range networkNames {
		vipID := strconv.Itoa(vipIDInt + 1)
		newVip := &avimodels.Vip{
			VipID:          &vipID,
			AutoAllocateIP: &autoAllocate,
		}
		newVip.SubnetUUID = &networkName
		vipList = append(vipList, newVip)
	}
	return vipList
}
