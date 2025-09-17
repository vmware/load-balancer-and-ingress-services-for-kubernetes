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
	"sync"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/rest"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/third_party/github.com/vmware/alb-sdk/go/clients"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/third_party/github.com/vmware/alb-sdk/go/session"

	gocidr "github.com/apparentlymart/go-cidr/cidr"
	"github.com/vmware/alb-sdk/go/models"
)

var CloudCache *models.Cloud
var NetCache map[string]*models.Network
var IPAMCache *models.IPAMDNSProviderProfile
var instantiateFullSyncWorker sync.Once
var worker *utils.FullSyncThread

// SyncLSLRNetwork fetches all networkinfo CR objects, compares them with the data network configured in the cloud,
// and updates the cloud if any LS-LR data is missing. It also creates or updates the VCF network with the CIDRs
// Provided in the Networkinfo objects.
func SyncLSLRNetwork() {
	lslrmap, cidrs := lib.GetNetworkInfoCRData(lib.GetDynamicClientSet())
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
		dataNetworkTier1Lrs, updateRequired := removeStaleLRLSEntries(client, cloudModel, lslrmap)
		if !matchSegmentInCloud(dataNetworkTier1Lrs, lslrmap) || updateRequired {
			cloudModel.NsxtConfiguration.DataNetworkConfig.Tier1SegmentConfig.Manual.Tier1Lrs = constructLsLrInCloud(dataNetworkTier1Lrs, lslrmap)
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

	addNetworkInCloud("fullsync", cidrs, client)
	networksToDelete := getNetworksToDeleteInCloud("fullsync", cidrs, client)
	addNetworkInIPAM("fullsync", networksToDelete, client)
	deleteNetworkInCloud("fullsync", networksToDelete, client)
}

func matchCidrInNetwork(subnets []*models.Subnet, cidrs map[string]struct{}) bool {
	if len(subnets) != len(cidrs) {
		return false
	}
	subnetCIDRs := make(map[string]struct{})
	for _, subnet := range subnets {
		addr := *subnet.Prefix.IPAddr.Addr
		mask := *subnet.Prefix.Mask
		cidr := fmt.Sprintf("%s/%d", addr, mask)
		if _, ok := cidrs[cidr]; !ok {
			utils.AviLog.Infof("could not find addr %s in list of cidrs", cidr)
			return false
		}
		subnetCIDRs[cidr] = struct{}{}
	}
	for cidr := range cidrs {
		if _, ok := subnetCIDRs[cidr]; !ok {
			utils.AviLog.Infof("could not find addr %s in subnet cidrs", cidr)
			return false
		}
	}
	return true
}

func addNetworkInCloud(objKey string, cidrToNS map[string]map[string]struct{}, client *clients.AviClient) {
	method := utils.RestPost
	path := "/api/network/"
	_, netModels := getAviNetFromCache(client)
	for ns, cidrs := range cidrToNS {
		netName := lib.GetVCFNetworkNameWithNS(ns)
		netModel, found := netModels[netName]
		if !found {
			utils.AviLog.Infof("key: %s, Failed to get Network data from cache", objKey)
			cloudRef := fmt.Sprintf("/api/cloud?name=%s", utils.CloudName)
			netModel = &models.Network{
				Name:     &netName,
				CloudRef: &cloudRef,
			}
		} else {
			if matchCidrInNetwork(netModel.ConfiguredSubnets, cidrs) {
				utils.AviLog.Infof("All CIDRs already present in the network, skipping network update")
				continue
			}
			netModel.ConfiguredSubnets = []*models.Subnet{}
			method = utils.RestPut
			path = "/api/network/" + *netModel.UUID
		}

		utils.AviLog.Infof("key: %s, list of cidrs to be added: %v", objKey, cidrs)
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

		utils.AviLog.Infof("key: %s, Adding/Updating VCF network: %v", objKey, utils.Stringify(restOp))
		executeRestOp(objKey, client, &restOp)
	}
}

func addNetworkInIPAM(key string, networksToDelete map[string]string, client *clients.AviClient) {
	found, cloudModel := getAviCloudFromCache(client, utils.CloudName)
	if !found {
		utils.AviLog.Warnf("Failed to get Cloud data from cache")
		return
	}

	found, ipam := getAviIPAMFromCache(client, strings.Split(*cloudModel.IPAMProviderRef, "#")[1])
	if !found {
		utils.AviLog.Warnf("Failed to get IPAM data from cache")
		return
	}

	_, netModels := getAviNetFromCache(client)
	if ipam.InternalProfile == nil {
		ipam.InternalProfile = &models.IPAMDNSInternalProfile{}
	}
	usableNetworks := make(map[string]struct{})
	updateIPAM := false
	for _, nw := range ipam.InternalProfile.UsableNetworks {
		netName := strings.Split(*nw.NwRef, "#")[1]
		if _, ok := netModels[netName]; !ok && strings.HasPrefix(netName, lib.GetVCFNetworkName()) {
			updateIPAM = true
			continue
		}
		if _, ok := networksToDelete[netName]; ok {
			updateIPAM = true
			continue
		}
		usableNetworks[netName] = struct{}{}
	}
	for netName := range netModels {
		if _, ok := networksToDelete[netName]; ok {
			continue
		}
		if _, exists := usableNetworks[netName]; !exists {
			updateIPAM = true
			usableNetworks[netName] = struct{}{}
		}
	}
	if !updateIPAM {
		return
	}

	ipamUsableNetworks := make([]*models.IPAMUsableNetwork, len(usableNetworks))
	i := 0
	for netName := range usableNetworks {
		networkRef := "/api/network/?name=" + netName
		ipamUsableNetworks[i] = &models.IPAMUsableNetwork{
			NwRef: proto.String(networkRef),
		}
		i += 1
	}
	ipam.InternalProfile.UsableNetworks = ipamUsableNetworks

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

func matchSegmentInCloud(lslrList []*models.Tier1LogicalRouterInfo, lslrMap map[string]string) bool {
	cloudLSLRMap := make(map[string]string)
	for i := range lslrList {
		cloudLSLRMap[*lslrList[i].SegmentID] = *lslrList[i].Tier1LrID
	}

	for ls, lr := range lslrMap {
		if val, ok := cloudLSLRMap[ls]; !ok || (ok && val != lr) {
			return false
		}
	}
	return true
}

func constructLsLrInCloud(lslrList []*models.Tier1LogicalRouterInfo, lslrMap map[string]string) []*models.Tier1LogicalRouterInfo {
	var cloudLSLRList []*models.Tier1LogicalRouterInfo
	cloudLSLRMap := make(map[string]string)
	addLRInfo := func(ls, lr string) {
		cloudLSLRList = append(cloudLSLRList, &models.Tier1LogicalRouterInfo{
			SegmentID: &ls,
			Tier1LrID: &lr,
		})
	}
	for i := range lslrList {
		cloudLSLRMap[*lslrList[i].SegmentID] = *lslrList[i].Tier1LrID
	}
	for ls, lr := range lslrMap {
		if val, ok := cloudLSLRMap[ls]; !ok || (ok && val != lr) {
			cloudLSLRMap[ls] = lr
		}
	}
	for ls, lr := range cloudLSLRMap {
		addLRInfo(ls, lr)
	}
	return cloudLSLRList
}

func getNetworksToDeleteInCloud(objKey string, cidrToNS map[string]map[string]struct{}, client *clients.AviClient) map[string]string {
	networks := make(map[string]string)
	_, netModels := getAviNetFromCache(client)
	for netName, net := range netModels {
		pfx := lib.GetVCFNetworkName() + "-"
		if !strings.HasPrefix(netName, pfx) {
			continue
		}
		ns := strings.Split(netName, pfx)[1]
		if _, ok := cidrToNS[ns]; !ok {
			networks[netName] = *net.UUID
		}
	}
	return networks
}

func deleteNetworkInCloud(objKey string, networksToDelete map[string]string, client *clients.AviClient) {
	wg := sync.WaitGroup{}
	for netName, netUUID := range networksToDelete {
		//delete the network
		path := "/api/network/" + netUUID
		restOp := utils.RestOp{
			ObjName: netName,
			Path:    path,
			Method:  utils.RestDelete,
			Tenant:  "admin",
			Model:   "network",
		}
		wg.Add(1)
		go func(restOp utils.RestOp) {
			utils.AviLog.Infof("key: %s, Deleting VCF network %s", objKey, restOp.ObjName)
			executeRestOp(objKey, client, &restOp)
			wg.Done()
		}(restOp)
	}
	wg.Wait()
}

func getAviCloudFromCache(client *clients.AviClient, cloudName string) (bool, models.Cloud) {
	if CloudCache == nil {
		if AviCloudCachePopulate(client, cloudName) != nil {
			return false, models.Cloud{}
		}
	}
	return true, *CloudCache
}

func getAviNetFromCache(client *clients.AviClient) (bool, map[string]*models.Network) {
	if len(NetCache) == 0 {
		if AviNetCachePopulate(client, utils.CloudName) != nil {
			return false, nil
		}
	}
	return true, NetCache
}

func getAviIPAMFromCache(client *clients.AviClient, ipamName string) (bool, models.IPAMDNSProviderProfile) {
	if IPAMCache == nil {
		if AviIPAMCachePopulate(client, ipamName) != nil {
			return false, models.IPAMDNSProviderProfile{}
		}
	}
	return true, *IPAMCache
}

// AviCloudCachePopulate queries avi rest api to get cloud data and stores in CloudCache
func AviCloudCachePopulate(client *clients.AviClient, cloudName string) error {
	uri := "/api/cloud/?include_name&name=" + cloudName
	result, err := lib.AviGetCollectionRaw(client, uri)
	if err != nil {
		utils.AviLog.Errorf("Cloud Get uri %v returned err %v", uri, err)
		return err
	}

	elems := make([]json.RawMessage, result.Count)
	err = json.Unmarshal(result.Results, &elems)
	if err != nil {
		utils.AviLog.Errorf("Failed to unmarshal data, err: %v", err)
		return err
	}

	if result.Count != 1 {
		err = fmt.Errorf("expected one cloud for cloud name: %s, found: %d", cloudName, result.Count)
		utils.AviLog.Errorf(err.Error())
		return err
	}

	CloudCache = &models.Cloud{}
	if err = json.Unmarshal(elems[0], &CloudCache); err != nil {
		utils.AviLog.Warnf("Failed to unmarshal cloud data, err: %v", err)
		return err
	}

	AviIPAMCachePopulate(client, strings.Split(*CloudCache.IPAMProviderRef, "#")[1])
	return nil
}

// AviNetCachePopulate queries avi rest api to get network data for vcf and stores in NetCaches
func AviNetCachePopulate(client *clients.AviClient, cloudName string) error {
	newNetCaches := make(map[string]*models.Network)
	err := aviNetCachePopulate(client, cloudName, newNetCaches)
	if err != nil {
		return err
	}
	NetCache = newNetCaches
	return nil
}

func aviNetCachePopulate(client *clients.AviClient, cloudName string, netCache map[string]*models.Network, nextPage ...lib.NextPage) error {
	uri := "/api/network/?include_name&cloud_ref.name=" + cloudName
	if len(nextPage) > 0 {
		uri = nextPage[0].NextURI
	}
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

	for i := 0; i < len(elems); i++ {
		net := &models.Network{}
		if err = json.Unmarshal(elems[i], &net); err != nil {
			return err
		}
		if strings.HasPrefix(*net.Name, lib.GetVCFNetworkName()) {
			netCache[*net.Name] = net
		}
	}
	if result.Next != "" {
		next_uri := strings.Split(result.Next, "/api/network")
		if len(next_uri) > 1 {
			overrideUri := "/api/network" + next_uri[1]
			nextPage := lib.NextPage{NextURI: overrideUri}
			aviNetCachePopulate(client, cloudName, netCache, nextPage)
		}
	}
	return nil
}

// AviIPAMCachePopulate queries avi rest api to get IPAM data and stores in IPAMCache
func AviIPAMCachePopulate(client *clients.AviClient, ipamName string) error {
	uri := "/api/ipamdnsproviderprofile/?include_name&name=" + ipamName
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
		utils.AviLog.Infof("Expected one IPAM with name: %s, found: %d", ipamName, result.Count)
		return fmt.Errorf("Expected one IPAM with name: %s, found: %d", ipamName, result.Count)
	}

	IPAMCache = &models.IPAMDNSProviderProfile{}
	if err = json.Unmarshal(elems[0], &IPAMCache); err != nil {
		utils.AviLog.Warnf("Failed to unmarshal IPAM data, err: %v", err)
		return err
	}
	return nil
}

func executeRestOp(key string, client *clients.AviClient, restOp *utils.RestOp, retryNum ...int) {
	utils.AviLog.Debugf("key: %s, Executing rest operation to sync object in cloud: %v", key, utils.Stringify(restOp))
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
	if restOp.Err != nil {
		err = restOp.Err
	}
	if err != nil {
		lib.CheckForInvalidCredentials(restOp.Path, err)
		if checkAndRetry(key, err) {
			executeRestOp(key, client, restOp, retry+1)
		} else if aviError, ok := err.(session.AviError); ok && aviError.HttpStatusCode == 403 {
			utils.AviLog.Debugf("Switching to admin context from  %s", lib.GetTenant())
			SetAdminTenant := session.SetTenant(lib.GetAdminTenant())
			SetTenant := session.SetTenant(lib.GetTenant())
			SetAdminTenant(client.AviSession)
			defer SetTenant(client.AviSession)
			executeRestOp(key, client, restOp)
		} else if strings.Contains(err.Error(), "Concurrent Update Error") {
			utils.AviLog.Infof("Got Concurrent Update Error, refreshing cache and scheduling periodic sync")
			refreshCache(restOp.Model, client)
			ScheduleQuickSync()
			return
		} else {
			utils.AviLog.Warnf("key %s, Got error in executing rest request: %v", key, err)
			refreshCache(restOp.Model, client)
			return
		}
	}
	refreshCache(restOp.Model, client)
	utils.AviLog.Infof("key: %s, Successfully executed rest operation to sync object: %v", key, utils.Stringify(restOp))
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

func NewLRLSFullSyncWorker() *utils.FullSyncThread {
	instantiateFullSyncWorker.Do(func() {
		worker = utils.NewFullSyncThread(time.Duration(lib.FullSyncInterval) * time.Second)
		worker.SyncFunction = SyncLSLRNetwork
		worker.QuickSyncFunction = func(qSync bool) error { return nil }
	})
	return worker
}

func ScheduleQuickSync() {
	if worker != nil {
		worker.QuickSync()
	}
}

func getClusterSpecificNSXTSegmentsinCloud(client *clients.AviClient, lsLRMap map[string]string, next ...string) error {
	uri := fmt.Sprintf("/api/nsxtsegmentruntime/?cloud_ref.name=%s", utils.CloudName)
	if len(next) > 0 {
		uri = next[0]
	}
	result, err := lib.AviGetCollectionRaw(client, uri)
	if err != nil {
		return err
	}
	elems := make([]json.RawMessage, result.Count)
	err = json.Unmarshal(result.Results, &elems)
	if err != nil {
		utils.AviLog.Warnf("Failed to unmarshal nsxt segment runtime result, err: %v", err)
		return err
	}
	for i := 0; i < len(elems); i++ {
		sg := models.NsxtSegmentRuntime{}
		err = json.Unmarshal(elems[i], &sg)
		if err != nil {
			utils.AviLog.Warnf("Failed to unmarshal nsxt segment runtime data, err: %v", err)
			return err
		}
		if strings.HasPrefix(*sg.Name, fmt.Sprintf("avi-%s", lib.GetClusterID())) {
			lsLRMap[*sg.SegmentID] = *sg.Tier1ID
		}
	}
	if result.Next != "" {
		next_uri := strings.Split(result.Next, "/api/nsxtsegmentruntime")
		if len(next_uri) > 1 {
			nextPage := "/api/nsxtsegmentruntime" + next_uri[1]
			err = getClusterSpecificNSXTSegmentsinCloud(client, lsLRMap, nextPage)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func removeStaleLRLSEntries(client *clients.AviClient, cloudModel models.Cloud, lslrmap map[string]string) ([]*models.Tier1LogicalRouterInfo, bool) {
	updatedRequired := false
	cloudTier1Lrs := cloudModel.NsxtConfiguration.DataNetworkConfig.Tier1SegmentConfig.Manual.Tier1Lrs
	dataNetworkTier1Lrs := make([]*models.Tier1LogicalRouterInfo, 0)
	cloudLRLSMap := make(map[string]string)
	err := getClusterSpecificNSXTSegmentsinCloud(client, cloudLRLSMap)
	if err != nil {
		utils.AviLog.Warnf("Failed to get LR to LS Map from cloud, err: %s", err)
		copy(dataNetworkTier1Lrs, cloudTier1Lrs)
		return dataNetworkTier1Lrs, updatedRequired
	}
	for i := 0; i < len(cloudTier1Lrs); i++ {
		if _, ok := lslrmap[*cloudTier1Lrs[i].SegmentID]; !ok {
			if _, present := cloudLRLSMap[*cloudTier1Lrs[i].SegmentID]; present {
				// Skipping this LS-LR entry, as it is present in cloud config, but not in WCP clutser
				updatedRequired = true
				continue
			}
		}
		dataNetworkTier1Lrs = append(dataNetworkTier1Lrs, cloudTier1Lrs[i])
	}
	return dataNetworkTier1Lrs, updatedRequired
}

func refreshCache(cacheModel string, client *clients.AviClient) {
	switch cacheModel {
	case "cloud":
		AviCloudCachePopulate(client, utils.CloudName)
	case "ipamdnsproviderprofile":
		AviCloudCachePopulate(client, utils.CloudName)
	case "network":
		AviNetCachePopulate(client, utils.CloudName)
	}
}
