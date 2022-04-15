/*
 * Copyright 2021 VMware, Inc.
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

package avirest

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"

	avimodels "github.com/vmware/alb-sdk/go/models"
	"google.golang.org/protobuf/proto"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/rest"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/third_party/github.com/vmware/alb-sdk/go/clients"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/third_party/github.com/vmware/alb-sdk/go/session"

	gocidr "github.com/apparentlymart/go-cidr/cidr"
	"github.com/vmware/alb-sdk/go/models"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var CloudCache *models.Cloud
var NetCache *models.Network
var IPAMCache *models.IPAMDNSProviderProfile

// SyncLSLRNetwork fetches all networkinfo CR objects, compares them with the data network configured in the cloud,
// and updates the cloud if any LS-LR data is missing. It also creates or updates the VCF network with the CIDRs
// Provided in the Networkinfo objects.
func SyncLSLRNetwork() {
	lslrmap, cidrs := lib.GetNetinfoCRData()
	utils.AviLog.Infof("Got data LS LR Map: %v, from NetworkInfo CR", lslrmap)

	client := InfraAviClientInstance()
	found, cloudModel := getAviCloudFromCache(client, utils.CloudName)
	if !found {
		utils.AviLog.Warnf("Failed to get Cloud data from cache")
		return
	}

	if cloudModel.NsxtConfiguration == nil {
		utils.AviLog.Warnf("NSX-T config not set in cloud, LS-LR mapping won't be updated")
		return
	}

	if len(cloudModel.NsxtConfiguration.DataNetworkConfig.VlanSegments) != 0 {
		utils.AviLog.Infof("NSX-T cloud is using Vlan Segments, LS-LR mapping won't be updated")
		return
	}

	if cloudModel.NsxtConfiguration.DataNetworkConfig.Tier1SegmentConfig.Manual == nil {
		utils.AviLog.Warnf("Tier1SegmentConfig is nil in NSX-T cloud, LS-LR mapping won't be updated")
		return
	}

	if len(lslrmap) > 0 {
		matched := matchSegmentInCloud(cloudModel.NsxtConfiguration.DataNetworkConfig.Tier1SegmentConfig.Manual.Tier1Lrs, lslrmap)
		if !matched {
			cloudModel.NsxtConfiguration.DataNetworkConfig.Tier1SegmentConfig.Manual.Tier1Lrs = constructLsLrInCloud(lslrmap)
			path := "/api/cloud/" + *cloudModel.UUID
			restOp := utils.RestOp{
				ObjName: utils.CloudName,
				Path:    path,
				Method:  utils.RestPut,
				Obj:     &cloudModel,
				Tenant:  "admin",
				Model:   "cloud",
			}
			executeRestOp("fullsync", client, &restOp)
		} else {
			utils.AviLog.Infof("LS LR update not required in cloud: %s", utils.CloudName)
		}
	}

	addNetworkInCloud("fullsync", cidrs, true)
	addNetworkInIPAM("fullsync")
}

func AddSegment(obj interface{}) bool {
	objKey := "Netinfo" + "/" + utils.ObjKey(obj)
	utils.AviLog.Debugf("key: %s, Network Info ADD Event", objKey)
	crd := obj.(*unstructured.Unstructured)

	spec := crd.Object["topology"].(map[string]interface{})
	lr, ok := spec["gatewayPath"].(string)
	if !ok {
		utils.AviLog.Infof("key: %s, lr not found from NetInfo CR", objKey)
		return false
	}
	ls, ok := spec["aviSegmentPath"].(string)
	if !ok {
		utils.AviLog.Infof("key: %s, ls not found from NetInfo CR", objKey)
		return false
	}
	cidrs := make(map[string]struct{})
	cidrIntf, ok := spec["ingressCIDRs"].([]interface{})
	if !ok {
		utils.AviLog.Infof("key: %s, cidr not found from NetInfo CR", objKey)
		return false
	} else {
		for _, cidr := range cidrIntf {
			cidrs[cidr.(string)] = struct{}{}
		}
	}

	utils.AviLog.Infof("key: %s, Adding LR %s, LS %s from networkinfo CR", objKey, lr, ls)
	client := InfraAviClientInstance()
	found, cloudModel := getAviCloudFromCache(client, utils.CloudName)
	if !found {
		utils.AviLog.Fatalf("key: %s, Failed to get Cloud data from cache", objKey)
		return false
	}

	if cloudModel.NsxtConfiguration == nil {
		utils.AviLog.Warnf("key: %s, NSX-T config not set in cloud, segment won't be added", objKey)
		return false
	}

	updateRequired, lslrList := addSegmentInCloud(cloudModel.NsxtConfiguration.DataNetworkConfig.Tier1SegmentConfig.Manual.Tier1Lrs, lr, ls)
	if updateRequired {
		cloudModel.NsxtConfiguration.DataNetworkConfig.Tier1SegmentConfig.Manual.Tier1Lrs = lslrList
		path := "/api/cloud/" + *cloudModel.UUID
		restOp := utils.RestOp{
			ObjName: utils.CloudName,
			Path:    path,
			Method:  utils.RestPut,
			Obj:     &cloudModel,
			Tenant:  "admin",
			Model:   "cloud",
		}
		executeRestOp(objKey, client, &restOp)
	} else {
		utils.AviLog.Infof("key: %s, LSLR update not required in cloud: %s", objKey, utils.CloudName)
	}

	addNetworkInCloud(objKey, cidrs, false)
	addNetworkInIPAM(objKey)
	return true
}

func DeleteSegment(obj interface{}) {
	objKey := "Netinfo" + utils.ObjKey(obj)
	utils.AviLog.Debugf("key:%s, Network Info DELETE Event", objKey)
	crd := obj.(*unstructured.Unstructured)

	spec := crd.Object["topology"].(map[string]interface{})
	lr, ok := spec["gatewayPath"].(string)
	if !ok {
		utils.AviLog.Infof("key: %s, lr not found from NetInfo CR", objKey)
		return
	}
	ls, ok := spec["aviSegmentPath"].(string)
	if !ok {
		utils.AviLog.Infof("key: %s, ls not found from NetInfo CR", objKey)
		return
	}
	cidrs := make(map[string]struct{})
	cidrIntf, ok := spec["ingressCIDRs"].([]interface{})
	if !ok {
		utils.AviLog.Infof("key: %s, cidr not found from NetInfo CR", objKey)
		return
	} else {
		for _, cidr := range cidrIntf {
			cidrs[cidr.(string)] = struct{}{}
		}
	}

	utils.AviLog.Infof("key: %s, Network Info CR deleted, removing LR %s, LS %s and CIDR %v from cloud", objKey, lr, ls, cidrs)

	client := InfraAviClientInstance()
	found, cloudModel := getAviCloudFromCache(client, utils.CloudName)
	if !found {
		utils.AviLog.Warnf("key: %s, Failed to get Cloud data from cache", objKey)
		return
	}

	if cloudModel.NsxtConfiguration == nil {
		utils.AviLog.Warnf("key: %s, NSX-T config not set in cloud, segment won't be deleted", objKey)
		return
	}

	updateRequired, lslrList := delSegmentInCloud(cloudModel.NsxtConfiguration.DataNetworkConfig.Tier1SegmentConfig.Manual.Tier1Lrs, lr, ls)
	if updateRequired {
		cloudModel.NsxtConfiguration.DataNetworkConfig.Tier1SegmentConfig.Manual.Tier1Lrs = lslrList
		path := "/api/cloud/" + *cloudModel.UUID
		restOp := utils.RestOp{
			ObjName: utils.CloudName,
			Path:    path,
			Method:  utils.RestPut,
			Obj:     &cloudModel,
			Tenant:  "admin",
			Model:   "cloud",
		}
		executeRestOp(objKey, client, &restOp)
	} else {
		utils.AviLog.Infof("key: %s, LSLR update not required in cloud: %s", objKey, utils.CloudName)
	}
	delCIDRFromNetwork(objKey, cidrs)
}

func matchCidrInNetwork(subnets []*models.Subnet, cidrs map[string]struct{}) bool {
	if len(subnets) != len(cidrs) {
		return false
	}
	for _, subnet := range subnets {
		addr := *subnet.Prefix.IPAddr.Addr
		mask := *subnet.Prefix.Mask
		cidr := fmt.Sprintf("%s/%d", addr, mask)
		if !utils.HasElem(cidrs, cidr) {
			utils.AviLog.Infof("could not find addr %s", cidr)
			return false
		}
	}
	return true
}

func findAndRemoveCidrInNetwork(subnets []*models.Subnet, cidrs map[string]struct{}) ([]*models.Subnet, bool) {
	subnetsCopy := make([]*models.Subnet, len(subnets))
	copy(subnetsCopy, subnets)
	cidrsLen := len(cidrs)
	if cidrsLen == 0 {
		return subnetsCopy, true
	}
	for i, subnet := range subnets {
		addr := *subnet.Prefix.IPAddr.Addr
		mask := *subnet.Prefix.Mask
		cidr := fmt.Sprintf("%s/%d", addr, mask)
		if _, found := cidrs[cidr]; found {
			subnetsCopy = append(subnetsCopy[:i], subnetsCopy[i+1:]...)
			cidrsLen -= 1
			if cidrsLen == 0 {
				return subnetsCopy, true
			}
		}
	}
	return subnetsCopy, false
}

func addNetworkInCloud(objKey string, cidrs map[string]struct{}, replaceAll bool) {
	client := InfraAviClientInstance()
	netName := lib.GetVCFNetworkName()
	method := utils.RestPost
	path := "/api/network/"
	found, netModel := getAviNetFromCache(client, netName)
	if !found {
		utils.AviLog.Warnf("key: %s, Failed to get Network data from cache", objKey)
		cloudRef := fmt.Sprintf("/api/cloud?name=%s", utils.CloudName)
		netModel = models.Network{
			Name:     &netName,
			CloudRef: &cloudRef,
		}
	} else {
		if replaceAll {
			if matchCidrInNetwork(netModel.ConfiguredSubnets, cidrs) {
				utils.AviLog.Infof("All CIDRs already present in the network, skipping network update")
				return
			}
			netModel.ConfiguredSubnets = []*models.Subnet{}
		} else {
			updatedSubnets, matched := findAndRemoveCidrInNetwork(netModel.ConfiguredSubnets, cidrs)
			if matched {
				utils.AviLog.Infof("All CIDRs already present in the network, skipping network update")
				return
			}
			netModel.ConfiguredSubnets = updatedSubnets
		}
		method = utils.RestPut
		path = "/api/network/" + *netModel.UUID
	}

	utils.AviLog.Infof("key: %s, list of CIDRs to be added: %v", objKey, cidrs)

	for cidr := range cidrs {
		_, ipnet, err := net.ParseCIDR(cidr)
		if err != nil {
			utils.AviLog.Warnf("key: %s, cidr %s is not of correct format, err %v", objKey, cidr, err)
			continue
		}
		startIP, endIP := gocidr.AddressRange(ipnet)
		startIPStr := gocidr.Inc(startIP).String()
		endIPStr := gocidr.Dec(endIP).String()
		ipStr := ipnet.IP.String()
		addrType := "V4"
		if !utils.IsV4(ipStr) {
			addrType = "V6"
		}
		mask, _ := ipnet.Mask.Size()
		mask32 := int32(mask)

		subnet := models.Subnet{
			Prefix: &models.IPAddrPrefix{
				IPAddr: &models.IPAddr{
					Addr: &ipStr,
					Type: &addrType,
				},
				Mask: &mask32,
			},
		}
		staticRange := models.StaticIPRange{
			Range: &models.IPAddrRange{
				Begin: &models.IPAddr{
					Addr: &startIPStr,
					Type: &addrType,
				},
				End: &models.IPAddr{
					Addr: &endIPStr,
					Type: &addrType,
				},
			},
		}
		subnet.StaticIPRanges = append(subnet.StaticIPRanges, &staticRange)
		netModel.ConfiguredSubnets = append(netModel.ConfiguredSubnets, &subnet)
	}

	restOp := utils.RestOp{
		ObjName: utils.CloudName,
		Path:    path,
		Method:  method,
		Obj:     &netModel,
		Tenant:  "admin",
		Model:   "network",
	}

	utils.AviLog.Infof("key: %s, Adding/Updating VCF network: %v", objKey, restOp)
	executeRestOp(objKey, client, &restOp)
}

func addNetworkInIPAM(key string) {
	client := InfraAviClientInstance()
	found, cloudModel := getAviCloudFromCache(client, utils.CloudName)
	if !found {
		utils.AviLog.Warnf("Failed to get Cloud data from cache")
		return
	}
	found, ipam := getAviIPAMFromCache(client, *cloudModel.IPAMProviderRef)
	if !found {
		utils.AviLog.Warnf("Failed to get IPAM data from cache")
		return
	}
	netName := lib.GetVCFNetworkName()
	networkRef := "/api/network/?name=" + netName

	if ipam.InternalProfile != nil &&
		ipam.InternalProfile.UsableNetworks != nil &&
		len(ipam.InternalProfile.UsableNetworks) > 0 {
		exists := false
		for _, ntw := range ipam.InternalProfile.UsableNetworks {
			if strings.Contains(*ntw.NwRef, netName) {
				exists = true
			}
		}
		if !exists {
			ipam.InternalProfile.UsableNetworks = append(ipam.InternalProfile.UsableNetworks, &models.IPAMUsableNetwork{
				NwRef: proto.String(networkRef),
			})
		}
	} else {
		ipam.InternalProfile = &avimodels.IPAMDNSInternalProfile{
			UsableNetworks: []*models.IPAMUsableNetwork{{NwRef: proto.String(networkRef)}},
		}
	}
	path := strings.Split(*cloudModel.IPAMProviderRef, "/ipamdnsproviderprofile/")[1]
	restOp := utils.RestOp{
		Path:   "/api/ipamdnsproviderprofile/" + path,
		Method: utils.RestPut,
		Obj:    &ipam,
		Tenant: "admin",
		Model:  "ipamdnsproviderprofile",
	}
	executeRestOp(key, client, &restOp)
}

func delCIDRFromNetwork(objKey string, cidrs map[string]struct{}) {
	client := InfraAviClientInstance()
	netName := lib.GetVCFNetworkName()
	method := utils.RestPut
	path := "/api/network/"
	found, netModel := getAviNetFromCache(client, netName)
	if !found {
		utils.AviLog.Infof("key: %s, Failed to get Network data from cache", objKey)
		return
	}

	updatedSubnets, _ := findAndRemoveCidrInNetwork(netModel.ConfiguredSubnets, cidrs)
	netModel.ConfiguredSubnets = updatedSubnets
	path = "/api/network/" + *netModel.UUID

	utils.AviLog.Infof("key: %s, list of CIDRs to be deleted: %v", objKey, cidrs)
	restOp := utils.RestOp{
		ObjName: utils.CloudName,
		Path:    path,
		Method:  method,
		Obj:     &netModel,
		Tenant:  "admin",
		Model:   "network",
	}

	utils.AviLog.Debugf("key: %s, executing restop to delete CIDR from vcf network: %v", objKey, restOp)
	executeRestOp(objKey, client, &restOp)
}

func addSegmentInCloud(lslrList []*models.Tier1LogicalRouterInfo, lr, ls string) (bool, []*models.Tier1LogicalRouterInfo) {
	listCopy := make([]*models.Tier1LogicalRouterInfo, len(lslrList))
	copy(listCopy, lslrList)
	for i := range listCopy {
		if *lslrList[i].SegmentID == ls {
			if *lslrList[i].Tier1LrID == lr {
				return false, lslrList
			}
			lslrList = append(lslrList[:i], lslrList[i+1:]...)
			break
		}
	}
	lrls := models.Tier1LogicalRouterInfo{
		SegmentID: &ls,
		Tier1LrID: &lr,
	}
	lslrList = append(lslrList, &lrls)
	return true, lslrList
}

func delSegmentInCloud(lslrList []*models.Tier1LogicalRouterInfo, lr, ls string) (bool, []*models.Tier1LogicalRouterInfo) {
	for i := range lslrList {
		if *lslrList[i].SegmentID == ls {
			lslrList = append(lslrList[:i], lslrList[i+1:]...)
			return true, lslrList
		}
	}
	return false, lslrList
}

func matchSegmentInCloud(lslrList []*models.Tier1LogicalRouterInfo, lslrMap map[string]string) bool {
	// if len(lslrMap) != len(lslrList) {
	// 	return false
	// }
	for i := range lslrList {
		if lslrMap[*lslrList[i].SegmentID] != *lslrList[i].Tier1LrID {
			return false
		}
	}
	return true
}

func constructLsLrInCloud(lslrMap map[string]string) []*models.Tier1LogicalRouterInfo {
	var lslrList []*models.Tier1LogicalRouterInfo
	for ls, lr := range lslrMap {
		lslr := models.Tier1LogicalRouterInfo{
			SegmentID: &ls,
			Tier1LrID: &lr,
		}
		lslrList = append(lslrList, &lslr)
	}
	return lslrList
}

func getAviCloudFromCache(client *clients.AviClient, cloudName string) (bool, models.Cloud) {
	if CloudCache == nil {
		if AviCloudCachePopulate(client, cloudName) != nil {
			return false, models.Cloud{}
		}
	}
	return true, *CloudCache
}

func getAviNetFromCache(client *clients.AviClient, netName string) (bool, models.Network) {
	if NetCache == nil {
		if AviNetCachePopulate(client, netName, utils.CloudName) != nil {
			return false, models.Network{}
		}
	}
	return true, *NetCache
}

func getAviIPAMFromCache(client *clients.AviClient, ipamRef string) (bool, models.IPAMDNSProviderProfile) {
	if IPAMCache == nil {
		ipamRefStr := strings.Split(ipamRef, "/")
		ipamUuid := ipamRefStr[len(ipamRefStr)-1]
		if AviIPAMCachePopulate(client, ipamUuid) != nil {
			return false, models.IPAMDNSProviderProfile{}
		}
	}
	return true, *IPAMCache
}

// AviCloudCachePopulate queries avi rest api to get cloud data and stores in CloudCache
func AviCloudCachePopulate(client *clients.AviClient, cloudName string) error {
	uri := "/api/cloud/?name=" + cloudName
	result, err := lib.AviGetCollectionRaw(client, uri)
	if err != nil {
		utils.AviLog.Warnf("Cloud Get uri %v returned err %v", uri, err)
		return err
	}

	elems := make([]json.RawMessage, result.Count)
	err = json.Unmarshal(result.Results, &elems)
	if err != nil {
		utils.AviLog.Warnf("Failed to unmarshal data, err: %v", err)
		return err
	}

	if result.Count != 1 {
		utils.AviLog.Errorf("Expected one cloud for cloud name: %s, found: %d", cloudName, result.Count)
		return fmt.Errorf("Expected one cloud for cloud name: %s, found: %d", cloudName, result.Count)
	}

	CloudCache = &models.Cloud{}
	if err = json.Unmarshal(elems[0], &CloudCache); err != nil {
		utils.AviLog.Warnf("Failed to unmarshal cloud data, err: %v", err)
		return err
	}
	return nil
}

// AviNetCachePopulate queries avi rest api to get network data for vcf and stores in NetCache
func AviNetCachePopulate(client *clients.AviClient, netName, cloudName string) error {
	uri := "/api/network/?name=" + netName + "&cloud_ref.name=" + cloudName
	result, err := lib.AviGetCollectionRaw(client, uri)
	if err != nil {
		utils.AviLog.Warnf("Cloud Get uri %v returned err %v", uri, err)
		return err
	}

	elems := make([]json.RawMessage, result.Count)
	err = json.Unmarshal(result.Results, &elems)
	if err != nil {
		utils.AviLog.Warnf("Failed to unmarshal data, err: %v", err)
		return err
	}

	if result.Count != 1 {
		utils.AviLog.Infof("Expected one network with network name: %s, found: %d", netName, result.Count)
		return fmt.Errorf("Expected one network with network name: %s, found: %d", netName, result.Count)
	}

	NetCache = &models.Network{}
	if err = json.Unmarshal(elems[0], &NetCache); err != nil {
		utils.AviLog.Warnf("Failed to unmarshal network data, err: %v", err)
		return err
	}
	return nil
}

// AviIPAMCachePopulate queries avi rest api to get IPAM data and stores in IPAMCache
func AviIPAMCachePopulate(client *clients.AviClient, ipamUuid string) error {
	uri := "/api/ipamdnsproviderprofile/" + ipamUuid
	result, err := lib.AviGetRaw(client, uri)
	if err != nil {
		utils.AviLog.Warnf("Cloud Get uri %v returned err %v", uri, err)
		return err
	}

	var data json.RawMessage
	err = json.Unmarshal(result, &data)
	if err != nil {
		utils.AviLog.Warnf("Failed to unmarshal data, err: %v", err)
		return err
	}

	if err = json.Unmarshal(data, &IPAMCache); err != nil {
		utils.AviLog.Warnf("Failed to unmarshal cloud data, err: %v", err)
		return err
	}
	return nil
}

func executeRestOp(key string, client *clients.AviClient, restOp *utils.RestOp, retryNum ...int) {
	utils.AviLog.Debugf("key: %s, Executing rest operation to sync object in cloud: %v", key, *restOp)
	retry := 0
	if len(retryNum) > 0 {
		utils.AviLog.Infof("key: %s, Retrying to execute rest request", key)
		retry = retryNum[0]
		if retry >= 3 {
			utils.AviLog.Errorf("key %s, rest request execution aborted after retrying 3 times", key)
			return
		}
	}
	restLayer := rest.NewRestOperations(nil, nil, true)
	err := restLayer.AviRestOperateWrapper(client, []*utils.RestOp{restOp}, key)
	if err != nil {
		if checkAndRetry(key, err) {
			executeRestOp(key, client, restOp, retry+1)
		} else if aviError, ok := err.(session.AviError); ok && aviError.HttpStatusCode == 403 {
			utils.AviLog.Debugf("Switching to admin context from  %s", lib.GetTenant())
			SetAdminTenant := session.SetTenant(lib.GetAdminTenant())
			SetTenant := session.SetTenant(lib.GetTenant())
			SetAdminTenant(client.AviSession)
			defer SetTenant(client.AviSession)
			executeRestOp(key, client, restOp)
		} else {
			utils.AviLog.Warnf("key %s, Got error in executing rest request: %v", key, err)
			return
		}
	}
	switch restOp.Model {
	case "cloud":
		AviCloudCachePopulate(client, utils.CloudName)
	case "network":
		AviNetCachePopulate(client, lib.GetVCFNetworkName(), utils.CloudName)
	}
	utils.AviLog.Infof("key: %s, Successfully executed rest operation to sync object: %v", key, *restOp)
}

func checkAndRetry(key string, err error) bool {
	if webSyncErr, ok := err.(*utils.WebSyncError); ok {
		if aviError, ok := webSyncErr.GetWebAPIError().(session.AviError); ok {
			switch aviError.HttpStatusCode {
			case 401:
				if strings.Contains(*aviError.Message, "Invalid credentials") {
					return false
				} else {
					utils.AviLog.Warnf("key %s, msg: got 401 error while executing rest request, adding to fast retry queue", key)
					return true
				}
			}
		}
	}
	return false
}
