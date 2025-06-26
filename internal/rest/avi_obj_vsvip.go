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
	"net"
	"reflect"
	"strconv"
	"strings"

	avicache "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/status"
	akov1beta1 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/apis/ako/v1beta1"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	"github.com/davecgh/go-spew/spew"
	avimodels "github.com/vmware/alb-sdk/go/models"
	"google.golang.org/protobuf/proto"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetIPAMProviderType() string {
	cache := avicache.SharedAviObjCache()
	cloud, ok := cache.CloudKeyCache.AviCacheGet(utils.CloudName)
	if !ok || cloud == nil {
		utils.AviLog.Warnf("Cloud object %s not found in cache", utils.CloudName)
		return ""
	}
	cloudProperty, ok := cloud.(*avicache.AviCloudPropertyCache)
	if !ok {
		utils.AviLog.Warnf("Cloud property object not found")
		return ""
	}
	return cloudProperty.IPAMType
}

func (rest *RestOperations) AviVsVipBuild(vsvip_meta *nodes.AviVSVIPNode, vsCache *avicache.AviVsCache, cache_obj *avicache.AviVSVIPCache, key string) (*utils.RestOp, error) {
	if lib.CheckObjectNameLength(vsvip_meta.Name, lib.VIP) {
		utils.AviLog.Warnf("key: %s not processing VSVIP object", key)
		return nil, nil
	}
	name := vsvip_meta.Name
	tenant := fmt.Sprintf("/api/tenant/?name=%s", lib.GetEscapedValue(vsvip_meta.Tenant))
	cloudRef := fmt.Sprintf("/api/cloud?name=%s", utils.CloudName)
	var dns_info_arr []*avimodels.DNSInfo
	var path string
	var networkRef string
	var rest_op utils.RestOp
	vipId, ipType, ip6Type := "0", "V4", "V6"

	cksum := vsvip_meta.CloudConfigCksum
	cksumstr := strconv.Itoa(int(cksum))

	// all vsvip models would have auto_alloc set to true even in case of static IP programming
	autoAllocate := true

	if cache_obj != nil {
		vsvip, err := rest.AviVsVipGet(key, cache_obj.Uuid, name, vsvip_meta.Tenant)
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

		noVipUpdatesAllowedForIPAMTypes := []string{
			lib.IPAMProviderInfoblox,
			lib.IPAMProviderCustom,
		}
		if !utils.HasElem(noVipUpdatesAllowedForIPAMTypes, GetIPAMProviderType()) {
			vip := &avimodels.Vip{
				VipID:                  &vipId,
				AutoAllocateIP:         &autoAllocate,
				AutoAllocateFloatingIP: vsvip_meta.EnablePublicIP,
			}

			// This would throw an error for advl4 the error is propagated to the gateway status.
			if vsvip_meta.IPAddress != "" {
				if utils.IsV4(vsvip_meta.IPAddress) {
					vip.IPAddress = &avimodels.IPAddr{Type: &ipType, Addr: &vsvip_meta.IPAddress}
				} else {
					vip.Ip6Address = &avimodels.IPAddr{Type: &ip6Type, Addr: &vsvip_meta.IPAddress}
				}
			}

			if lib.IsPublicCloud() && lib.GetCloudType() != lib.CLOUD_GCP {
				vips := networkNamesToVips(vsvip_meta.VipNetworks, vsvip_meta.EnablePublicIP)
				vsvip.Vip = []*avimodels.Vip{}
				vsvip.Vip = append(vsvip.Vip, vips...)
			} else if lib.GetCloudType() == lib.CLOUD_NSXT && lib.GetVPCMode() {
				vpcArr := strings.Split(vsvip_meta.T1Lr, "/vpcs/")
				projectArr := strings.Split(vpcArr[0], "/projects/")
				vipNetwork := fmt.Sprintf("%s_AVISEPARATOR_%s_AVISEPARATOR_PUBLIC", projectArr[len(projectArr)-1], vpcArr[len(vpcArr)-1])
				vip.SubnetUUID = &vipNetwork
				vsvip.Vip = []*avimodels.Vip{vip}
			} else {
				// Set the IPAM network subnet for all clouds except AWS and Azure
				if len(vsvip_meta.VipNetworks) != 0 {
					vip.IPAMNetworkSubnet = &avimodels.IPNetworkSubnet{}
					if len(vsvip.Vip) == 1 {
						if vsvip.Vip[0].IPAMNetworkSubnet != nil && vsvip.Vip[0].IPAMNetworkSubnet.NetworkRef != nil {
							ipamNetworkRefSplit := strings.Split(*vsvip.Vip[0].IPAMNetworkSubnet.NetworkRef, "#")
							if len(ipamNetworkRefSplit) > 1 && ipamNetworkRefSplit[1] == vsvip_meta.VipNetworks[0].NetworkName {
								vip.IPAMNetworkSubnet = vsvip.Vip[0].IPAMNetworkSubnet
							}
						}
					}
					// Shouldn't be required but kept for backup purpose
					networkRef = "/api/network/?name=" + vsvip_meta.VipNetworks[0].NetworkName
					if len(vsvip_meta.VipNetworks[0].NetworkUUID) != 0 {
						networkRef = "/api/network/" + vsvip_meta.VipNetworks[0].NetworkUUID
					}
					vip.IPAMNetworkSubnet.NetworkRef = &networkRef
					utils.AviLog.Debugf("Network: %s Network ref in rest layer: %s", vsvip_meta.VipNetworks[0].NetworkName, *vip.IPAMNetworkSubnet.NetworkRef)
					if vsvip_meta.VipNetworks[0].V6Cidr != "" {
						lib.UpdateV6(vip, &vsvip_meta.VipNetworks[0])
					}
					if lib.GetCloudType() == lib.CLOUD_NSXT &&
						lib.GetNSXTTransportZone() == lib.VLAN_TRANSPORT_ZONE {
						setVipPlacementNetwork(vip, vsvip_meta.VipNetworks[0].Cidr, &networkRef)
					}
					vsvip.Vip = []*avimodels.Vip{vip}
				}
			}

			if vsCache != nil && !vsCache.EnableRhi && len(vsvip_meta.BGPPeerLabels) > 0 {
				err = fmt.Errorf("to use selective vip advertisement, %s VS must advertise vips via BGP. Please recreate the VS", vsCache.Name)
				utils.AviLog.Errorf("To use selective vip advertisement, %s VS must advertise vips via BGP. Please recreate the VS", vsCache.Name)
				return nil, err
			}

			if len(vsvip_meta.BGPPeerLabels) > 0 {
				vsvip.BgpPeerLabels = vsvip_meta.BGPPeerLabels
			} else {
				vsvip.BgpPeerLabels = nil
			}
		}

		vsvip.Markers = lib.GetMarkers()

		rest_op = utils.RestOp{
			ObjName: name,
			Path:    "/api/vsvip/" + cache_obj.Uuid,
			Method:  utils.RestPut,
			Obj:     vsvip,
			Tenant:  vsvip_meta.Tenant,
			Model:   "VsVip",
		}
	} else {
		var vips []*avimodels.Vip
		vip := avimodels.Vip{
			VipID:                  &vipId,
			AutoAllocateIP:         &autoAllocate,
			AutoAllocateFloatingIP: vsvip_meta.EnablePublicIP,
		}

		// configuring static IP, from gateway.Addresses (advl4, svcapi) and service.loadBalancerIP (l4)
		if vsvip_meta.IPAddress != "" {
			if utils.IsV4(vsvip_meta.IPAddress) {
				vip.IPAddress = &avimodels.IPAddr{Type: &ipType, Addr: &vsvip_meta.IPAddress}
			} else {
				vip.Ip6Address = &avimodels.IPAddr{Type: &ip6Type, Addr: &vsvip_meta.IPAddress}
			}
		}

		// selecting network with user input, in case user input is not provided AKO relies on
		// usable network configuration in ipamdnsproviderprofile
		if lib.IsPublicCloud() && lib.GetCloudType() != lib.CLOUD_GCP {
			vips = networkNamesToVips(vsvip_meta.VipNetworks, vsvip_meta.EnablePublicIP)
		} else if lib.GetCloudType() == lib.CLOUD_NSXT && lib.GetVPCMode() {
			vpcArr := strings.Split(vsvip_meta.T1Lr, "/vpcs/")
			projectArr := strings.Split(vpcArr[0], "/projects/")
			vipNetwork := fmt.Sprintf("%s_AVISEPARATOR_%s_AVISEPARATOR_PUBLIC", projectArr[len(projectArr)-1], vpcArr[len(vpcArr)-1])
			vip.SubnetUUID = &vipNetwork
		} else {
			// Set the IPAM network subnet for all clouds except AWS and Azure
			if len(vsvip_meta.VipNetworks) != 0 {
				vipNetwork := vsvip_meta.VipNetworks[0]
				if vip.IPAMNetworkSubnet == nil {
					vip.IPAMNetworkSubnet = &avimodels.IPNetworkSubnet{}
				}
				networkRef := "/api/network/?name=" + vipNetwork.NetworkName
				if len(vipNetwork.NetworkUUID) != 0 {
					networkRef = "/api/network/" + vipNetwork.NetworkUUID
				}
				vip.IPAMNetworkSubnet.NetworkRef = &networkRef
				utils.AviLog.Debugf("Network: %s Network ref in rest layer: %s", vsvip_meta.VipNetworks[0].NetworkName, *vip.IPAMNetworkSubnet.NetworkRef)
				// setting IPAMNetworkSubnet.Subnet value in case subnetCIDR is provided
				if vipNetwork.Cidr == "" && vipNetwork.V6Cidr == "" {
					utils.AviLog.Warnf("key: %s, msg: Incomplete values provided for CIDR, will not use IPAMNetworkSubnet in vsvip", key)
				} else {
					var ipPrefixSlice, ip6PrefixSlice []string
					var mask, mask6 int
					if vipNetwork.Cidr != "" {
						ipPrefixSlice = strings.Split(vipNetwork.Cidr, "/")
						mask, _ = strconv.Atoi(ipPrefixSlice[1])
					}
					if vipNetwork.V6Cidr != "" {
						ip6PrefixSlice = strings.Split(vipNetwork.V6Cidr, "/")
						mask6, _ = strconv.Atoi(ip6PrefixSlice[1])
					}
					if (lib.IsPublicCloud() && lib.GetCloudType() == lib.CLOUD_GCP) || (!utils.IsWCP()) {
						if vipNetwork.Cidr != "" {
							vip.IPAMNetworkSubnet.Subnet = &avimodels.IPAddrPrefix{
								IPAddr: &avimodels.IPAddr{Type: &ipType, Addr: &ipPrefixSlice[0]},
								Mask:   proto.Int32(int32(mask)),
							}
						}
						if vipNetwork.V6Cidr != "" {
							vip.IPAMNetworkSubnet.Subnet6 = &avimodels.IPAddrPrefix{
								IPAddr: &avimodels.IPAddr{Type: &ip6Type, Addr: &ip6PrefixSlice[0]},
								Mask:   proto.Int32(int32(mask6)),
							}
						}
					}
					if lib.GetCloudType() == lib.CLOUD_NSXT &&
						lib.GetNSXTTransportZone() == lib.VLAN_TRANSPORT_ZONE {
						setVipPlacementNetwork(&vip, vipNetwork.Cidr, &networkRef)

					}
				}
				if vipNetwork.V6Cidr != "" {
					lib.UpdateV6(&vip, &vipNetwork)
				}
			}
		}

		if len(vips) == 0 {
			vips = append(vips, &vip)
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
		var vrfContextRef string
		vsvip := avimodels.VsVip{
			Name:                  &name,
			TenantRef:             &tenant,
			CloudRef:              &cloudRef,
			EastWestPlacement:     proto.Bool(false),
			DNSInfo:               dns_info_arr,
			Vip:                   vips,
			VsvipCloudConfigCksum: &cksumstr,
		}

		if vsvip_meta.VrfContext != "" {
			vrfContextRef = "/api/vrfcontext?name=" + vsvip_meta.VrfContext
			vsvip.VrfContextRef = &vrfContextRef
		}
		if vsvip_meta.T1Lr != "" {
			vsvip.Tier1Lr = &vsvip_meta.T1Lr
		}
		if len(vsvip_meta.BGPPeerLabels) > 0 {
			vsvip.BgpPeerLabels = vsvip_meta.BGPPeerLabels
		}

		vsvip.Markers = lib.GetMarkers()

		path = "/api/vsvip"
		// Patch an existing vsvip if it exists in the cache but not associated with this VS.
		vsvip_key := avicache.NamespaceName{Namespace: vsvip_meta.Tenant, Name: name}
		utils.AviLog.Debugf("key: %s, searching in cache for vsVip Key: %s", key, vsvip_key)
		vsvip_cache, ok := rest.cache.VSVIPCache.AviCacheGet(vsvip_key)
		if ok {
			vsvip_cache_obj, _ := vsvip_cache.(*avicache.AviVSVIPCache)
			vsvip_avi, err := rest.AviVsVipGet(key, vsvip_cache_obj.Uuid, name, vsvip_meta.Tenant)
			if err != nil {
				if strings.Contains(err.Error(), lib.VSVIPNotFoundError) {
					// Clear the cache for this key
					rest.cache.VSVIPCache.AviCacheDelete(vsvip_key)
					utils.AviLog.Warnf("key: %s, Removed the vsvip object from the cache", key)
					rest_op = utils.RestOp{
						ObjName: name,
						Path:    path,
						Method:  utils.RestPost,
						Obj:     vsvip,
						Tenant:  vsvip_meta.Tenant,
						Model:   "VsVip",
					}
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
			if len(vsvip_meta.BGPPeerLabels) > 0 {
				vsvip_avi.BgpPeerLabels = vsvip_meta.BGPPeerLabels
			} else {
				vsvip_avi.BgpPeerLabels = nil
			}

			vsvip.Markers = lib.GetMarkers()

			vsvip_avi.VsvipCloudConfigCksum = &cksumstr
			path = "/api/vsvip/" + vsvip_cache_obj.Uuid
			rest_op = utils.RestOp{
				ObjName: name,
				Path:    path,
				Method:  utils.RestPut,
				Obj:     vsvip_avi,
				Tenant:  vsvip_meta.Tenant,
				Model:   "VsVip",
			}
		} else {
			rest_op = utils.RestOp{
				ObjName: name,
				Path:    path,
				Method:  utils.RestPost,
				Obj:     vsvip,
				Tenant:  vsvip_meta.Tenant,
				Model:   "VsVip",
			}
		}
	}

	return &rest_op, nil
}

func (rest *RestOperations) AviVsVipGet(key, uuid, name, tenant string) (*avimodels.VsVip, error) {
	aviRestPoolClient := avicache.SharedAVIClients(tenant)
	if aviRestPoolClient == nil {
		utils.AviLog.Warnf("key: %s, msg: aviRestPoolClient during vsvip not initialized", key)
		return nil, errors.New("client in aviRestPoolClient during vsvip not initialized")
	}
	if len(aviRestPoolClient.AviClient) < 1 {
		utils.AviLog.Warnf("key: %s, msg: client in aviRestPoolClient during vsvip not initialized", key)
		return nil, errors.New("client in aviRestPoolClient during vsvip not initialized")
	}
	client := aviRestPoolClient.AviClient[0]
	uri := "/api/vsvip/" + uuid + "/?include_name"

	rawData, err := lib.AviGetRaw(client, uri)
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
	rest_op := utils.RestOp{
		Path:   path,
		Method: "DELETE",
		Tenant: tenant,
		Model:  "VsVip",
	}
	utils.AviLog.Infof(spew.Sprintf("key: %s, msg: VSVIP DELETE Restop %v ", key,
		utils.Stringify(rest_op)))
	return &rest_op
}

func (rest *RestOperations) AviVsVipPut(uuid string, vsvipObj *avimodels.VsVip, tenant string, key string) *utils.RestOp {
	path := "/api/vsvip/" + uuid
	rest_op := utils.RestOp{
		Path:   path,
		Method: utils.RestPut,
		Obj:    vsvipObj,
		Tenant: tenant,
		Model:  "VsVip",
	}
	utils.AviLog.Infof(spew.Sprintf("key: %s, msg: VSVIP PUT Restop %v ", key,
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
				if utils.IsWCP() {
					gw, err := lib.AKOControlConfig().AdvL4Informers().GatewayInformer.Lister().Gateways(gwNSName[0]).Get(gwNSName[1])
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
					gw, err := lib.AKOControlConfig().SvcAPIInformers().GatewayInformer.Lister().Gateways(gwNSName[0]).Get(gwNSName[1])
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

	resp_elems := rest.restOperator.RestRespArrToObjByType(rest_op, "vsvip", key)
	if resp_elems == nil {
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
		var vsvipFips []string
		var vsvipV6ips []string
		var networkNames []string
		if _, found := resp["vip"]; found {
			if vips, ok := resp["vip"].([]interface{}); ok {
				for _, vipsIntf := range vips {
					vip, valid := vipsIntf.(map[string]interface{})
					if !valid {
						utils.AviLog.Infof("key: %s, msg: invalid type for vip in vsvip: %s", key, name)
						continue
					}
					ipType := lib.IPTypeV4Only
					auto_allocate_ip_type, ok := vip["auto_allocate_ip_type"]
					if ok {
						ipType = auto_allocate_ip_type.(string)
					}
					if ipType == lib.IPTypeV4Only || ipType == lib.IPTypeV4V6 {
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
					if ipType == lib.IPTypeV6Only || ipType == lib.IPTypeV4V6 {
						ip6_address, valid := vip["ip6_address"].(map[string]interface{})
						if !valid {
							utils.AviLog.Warnf("key: %s, msg: invalid type for ip6_address in vsvip: %s", key, name)
							continue
						}
						v6_addr, valid := ip6_address["addr"].(string)
						if !valid {
							utils.AviLog.Warnf("key: %s, msg: invalid type for v6 addr in vsvip: %s", key, name)
							continue
						}
						vsvipV6ips = append(vsvipV6ips, v6_addr)
					}
					fipEnabled := false
					auto_allocate_floating_ip, ok := vip["auto_allocate_floating_ip"]
					if ok {
						fipEnabled = auto_allocate_floating_ip.(bool)
					}
					if fipEnabled {
						floating_ip, valid := vip["floating_ip"].(map[string]interface{})
						if !valid {
							utils.AviLog.Warnf("key: %s, msg: invalid type for floating_ip in vsvip: %s", key, name)
						} else {
							fip_addr, valid := floating_ip["addr"].(string)
							if !valid {
								utils.AviLog.Warnf("key: %s, msg: invalid type for addr in vsvip: %s", key, name)
								continue
							}
							vsvipFips = append(vsvipFips, fip_addr)
						}
					}

					if ipamNetworkSubnet, ipamOk := vip["ipam_network_subnet"].(map[string]interface{}); ipamOk {
						if networkRef, netRefOk := ipamNetworkSubnet["network_ref"].(string); netRefOk {
							if networkRefName := strings.Split(networkRef, "#"); len(networkRefName) == 2 {
								networkNames = append(networkNames, strings.Split(networkRef, "#")[1])
							}
							if lib.UsesNetworkRef() {
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
			Fips:         vsvipFips,
			V6IPs:        vsvipV6ips,
			NetworkNames: networkNames,
		}

		if lastModifiedStr == "" {
			vsvip_cache_obj.InvalidData = true
		}

		if resp["vsvip_cloud_config_cksum"] != nil {
			vsvip_cache_obj.CloudConfigCksum = resp["vsvip_cloud_config_cksum"].(string)
		}

		k := avicache.NamespaceName{Namespace: rest_op.Tenant, Name: name}
		oldVsVipCache, oldVsVipFound := rest.cache.VSVIPCache.AviCacheGet(k)
		rest.cache.VSVIPCache.AviCacheAdd(k, &vsvip_cache_obj)
		// Update the VS object
		vs_cache, ok := rest.cache.VsCacheMeta.AviCacheGet(vsKey)
		if ok {
			vs_cache_obj, found := vs_cache.(*avicache.AviVsCache)
			if found {
				var oldVsVips, oldVsFips, oldVsV6ips []string
				if oldVsVipFound {
					oldVsVipCacheObj, ok := oldVsVipCache.(*avicache.AviVSVIPCache)
					if ok {
						oldVsVips = oldVsVipCacheObj.Vips
						oldVsFips = oldVsVipCacheObj.Fips
						oldVsV6ips = oldVsVipCacheObj.V6IPs
					}
				}

				vs_cache_obj.AddToVSVipKeyCollection(k)
				utils.AviLog.Debugf("key: %s, msg: modified the VS cache object for VSVIP collection. The cache now is :%v", key, utils.Stringify(vs_cache_obj))
				if rest_op.Method == utils.RestPut {
					if len(vs_cache_obj.SNIChildCollection) > 0 {
						for _, childUuid := range vs_cache_obj.SNIChildCollection {
							childKey, childFound := rest.cache.VsCacheMeta.AviCacheGetKeyByUuid(childUuid)
							if childFound {
								childVSKey := childKey.(avicache.NamespaceName)
								childObj, _ := rest.cache.VsCacheMeta.AviCacheGet(childVSKey)
								child_cache_obj, vs_found := childObj.(*avicache.AviVsCache)
								if vs_found {
									if !reflect.DeepEqual(vsvip_cache_obj.Vips, oldVsVips) ||
										!reflect.DeepEqual(vsvip_cache_obj.Fips, oldVsFips) ||
										!reflect.DeepEqual(vsvip_cache_obj.V6IPs, oldVsV6ips) {
										rest.StatusUpdateForPool(rest_op.Method, child_cache_obj, key)
										// rest.StatusUpdateForVS(child_cache_obj, key)
									}
								}
							}
						}
					}
					if !reflect.DeepEqual(vsvip_cache_obj.Vips, oldVsVips) ||
						!reflect.DeepEqual(vsvip_cache_obj.Fips, oldVsFips) ||
						!reflect.DeepEqual(vsvip_cache_obj.V6IPs, oldVsV6ips) {
						rest.StatusUpdateForPool(rest_op.Method, vs_cache_obj, key)
						// rest.StatusUpdateForVS(vs_cache_obj, key)
					}
				}
			}

		} else {
			vs_cache_obj := rest.cache.VsCacheMeta.AviCacheAddVS(vsKey)
			vs_cache_obj.AddToVSVipKeyCollection(k)
			utils.AviLog.Infof("key: %s, msg: added VS cache key during vsvip update %v val %v", key, vsKey, utils.Stringify(vs_cache_obj))
			if rest_op.Method == utils.RestPut {
				if len(vs_cache_obj.SNIChildCollection) > 0 {
					for _, childUuid := range vs_cache_obj.SNIChildCollection {
						childKey, childFound := rest.cache.VsCacheMeta.AviCacheGetKeyByUuid(childUuid)
						if childFound {
							childVSKey := childKey.(avicache.NamespaceName)
							childObj, _ := rest.cache.VsCacheMeta.AviCacheGet(childVSKey)
							child_cache_obj, vs_found := childObj.(*avicache.AviVsCache)
							if vs_found {
								rest.StatusUpdateForPool(rest_op.Method, child_cache_obj, key)
								// rest.StatusUpdateForVS(child_cache_obj, key)
							}
						}
					}
				}
				rest.StatusUpdateForPool(rest_op.Method, vs_cache_obj, key)
				// rest.StatusUpdateForVS(vs_cache_obj, key)
			}
		}
		utils.AviLog.Infof(spew.Sprintf("key: %s, msg: added vsvip cache k %v val %v", key, k,
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

func networkNamesToVips(vipNetworks []akov1beta1.AviInfraSettingVipNetwork, enablePublicIP *bool) []*avimodels.Vip {
	var vipList []*avimodels.Vip
	autoAllocate := true

	for vipIDInt, vipNetwork := range vipNetworks {
		vipID := strconv.Itoa(vipIDInt + 1)
		newVip := &avimodels.Vip{
			VipID:                  &vipID,
			AutoAllocateIP:         &autoAllocate,
			AutoAllocateFloatingIP: enablePublicIP,
		}
		newVip.SubnetUUID = proto.String(vipNetwork.NetworkName)
		vipList = append(vipList, newVip)
		if vipNetwork.V6Cidr != "" {
			lib.UpdateV6(newVip, &vipNetwork)
		}
	}

	return vipList
}

func setVipPlacementNetwork(vip *avimodels.Vip, cidr string, networkRef *string) {
	_, ipnet, _ := net.ParseCIDR(cidr)
	addr := ipnet.IP.String()
	mask := strings.Split(cidr, "/")[1]
	intMask, _ := strconv.ParseInt(mask, 10, 32)
	int32Mask := int32(intMask)
	placementNetwork := &avimodels.VipPlacementNetwork{
		NetworkRef: networkRef,
		Subnet: &avimodels.IPAddrPrefix{
			IPAddr: &avimodels.IPAddr{
				Type: proto.String("V4"),
				Addr: &addr,
			},
			Mask: &int32Mask,
		},
	}
	vip.PlacementNetworks = []*avimodels.VipPlacementNetwork{placementNetwork}
}
