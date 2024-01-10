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
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"google.golang.org/protobuf/proto"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/third_party/github.com/vmware/alb-sdk/go/clients"

	gocidr "github.com/apparentlymart/go-cidr/cidr"
	"github.com/vmware/alb-sdk/go/models"
)

type T1LRNetworking struct {
}

func (t *T1LRNetworking) AddNetworkInfoEventHandler(stopCh <-chan struct{}) {
	networkInfoHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			utils.AviLog.Infof("NCP Network Info ADD Event")
			ScheduleQuickSync()
		},
		UpdateFunc: func(old, obj interface{}) {
			utils.AviLog.Infof("NCP Network Info Update Event")
			ScheduleQuickSync()
		},
		DeleteFunc: func(obj interface{}) {
			utils.AviLog.Infof("NCP Network Info Delete Event")
			ScheduleQuickSync()
		},
	}
	lib.GetDynamicInformers().VCFNetworkInfoInformer.Informer().AddEventHandler(networkInfoHandler)
	go lib.GetDynamicInformers().VCFNetworkInfoInformer.Informer().Run(stopCh)

	ClusterNetworkInfoHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			utils.AviLog.Infof("NCP Cluster Network Info ADD Event")
			ScheduleQuickSync()
		},
		UpdateFunc: func(old, obj interface{}) {
			utils.AviLog.Infof("NCP Cluster Network Info Update Event")
			ScheduleQuickSync()
		},
		DeleteFunc: func(obj interface{}) {
			utils.AviLog.Infof("NCP Cluster Network Info Delete Event")
			ScheduleQuickSync()
		},
	}
	lib.GetDynamicInformers().VCFClusterNetworkInformer.Informer().AddEventHandler(ClusterNetworkInfoHandler)
	go lib.GetDynamicInformers().VCFClusterNetworkInformer.Informer().Run(stopCh)
	if !cache.WaitForCacheSync(stopCh,
		lib.GetDynamicInformers().VCFNetworkInfoInformer.Informer().HasSynced,
		lib.GetDynamicInformers().VCFClusterNetworkInformer.Informer().HasSynced) {
		runtime.HandleError(fmt.Errorf("timed out waiting for cluster/namespace network info caches to sync"))
	} else {
		utils.AviLog.Infof("Caches synced for cluster/namespace network info informer")
	}
}

// SyncLSLRNetwork fetches all networkinfo CR objects, compares them with the data network configured in the cloud,
// and updates the cloud if any LS-LR data is missing. It also creates or updates the VCF network with the CIDRs
// Provided in the Networkinfo objects.
func (t *T1LRNetworking) SyncLSLRNetwork() {
	lslrmap, nsLRMap, cidrs := lib.GetNetworkInfoCRData(lib.GetDynamicClientSet())
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
		dataNetworkTier1Lrs, updateRequired := t.removeStaleLRLSEntries(client, cloudModel, lslrmap)
		if !t.matchSegmentInCloud(dataNetworkTier1Lrs, lslrmap) || updateRequired {
			cloudModel.NsxtConfiguration.DataNetworkConfig.Tier1SegmentConfig.Manual.Tier1Lrs = t.constructLsLrInCloud(dataNetworkTier1Lrs, lslrmap)
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

	t.addNetworkInCloud("fullsync", cidrs, client)
	networksToDelete := t.getNetworksToDeleteInCloud("fullsync", cidrs, client)
	t.addNetworkInIPAM("fullsync", networksToDelete, client)
	t.deleteNetworkInCloud("fullsync", networksToDelete, client)
	t.createInfraSettingAndAnnotateNS(nsLRMap, cidrs)
}

func (t *T1LRNetworking) createInfraSettingAndAnnotateNS(nsLRMap map[string]string, cidrs map[string]map[string]struct{}) {
	infraSettingCRs, err := lib.AKOControlConfig().CRDInformers().AviInfraSettingInformer.Lister().List(labels.Set(nil).AsSelector())
	if err != nil {
		utils.AviLog.Errorf("failed to list AviInfraSetting, error: %s", err.Error())
		return
	}

	oldInfraSettingCRSet := make(map[string]struct{})
	for _, infraSettingCR := range infraSettingCRs {
		oldInfraSettingCRSet[infraSettingCR.Name] = struct{}{}
	}

	newInfraSettingCRSet := make(map[string]struct{})
	wg := sync.WaitGroup{}
	for ns, lr := range nsLRMap {
		infraSettingName := getInfraSettingNameFromT1LR(lr)
		if _, ok := newInfraSettingCRSet[infraSettingName]; ok {
			lib.AnnotateNamespaceWithInfraSetting(ns, infraSettingName)
			continue
		}
		newInfraSettingCRSet[infraSettingName] = struct{}{}
		if _, ok := oldInfraSettingCRSet[infraSettingName]; ok {
			continue
		}
		netName := lib.GetVCFNetworkName()
		if _, ok := cidrs[ns]; ok {
			netName = lib.GetVCFNetworkNameWithNS(ns)
		}
		wg.Add(1)
		go func(lr, ns string) {
			_, err := lib.CreateAviInfraSetting(infraSettingName, netName, lr, "")
			if err != nil {
				utils.AviLog.Errorf("failed to create aviInfraSetting, name: %s, error: %s", infraSettingName, err.Error())
			} else {
				lib.AnnotateNamespaceWithInfraSetting(ns, infraSettingName)
			}
			wg.Done()
		}(lr, ns)
	}

	for infraSettingName := range oldInfraSettingCRSet {
		if _, ok := newInfraSettingCRSet[infraSettingName]; ok {
			continue
		}
		wg.Add(1)
		go func(name string) {
			err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().AviInfraSettings().Delete(context.TODO(), name, metav1.DeleteOptions{})
			if err != nil {
				utils.AviLog.Errorf("failed to delete aviInfraSetting, name: %s, error: %s", name, err.Error())
			}
			wg.Done()
		}(infraSettingName)
	}
	wg.Wait()
}

func (t *T1LRNetworking) matchCidrInNetwork(subnets []*models.Subnet, cidrs map[string]struct{}) bool {
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

func (t *T1LRNetworking) addNetworkInCloud(objKey string, cidrToNS map[string]map[string]struct{}, client *clients.AviClient) {
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
			if t.matchCidrInNetwork(netModel.ConfiguredSubnets, cidrs) {
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

func (t *T1LRNetworking) addNetworkInIPAM(key string, networksToDelete map[string]string, client *clients.AviClient) {
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

func (t *T1LRNetworking) matchSegmentInCloud(lslrList []*models.Tier1LogicalRouterInfo, lslrMap map[string]string) bool {
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

func (t *T1LRNetworking) constructLsLrInCloud(lslrList []*models.Tier1LogicalRouterInfo, lslrMap map[string]string) []*models.Tier1LogicalRouterInfo {
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

func (t *T1LRNetworking) getNetworksToDeleteInCloud(objKey string, cidrToNS map[string]map[string]struct{}, client *clients.AviClient) map[string]string {
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

func (t *T1LRNetworking) deleteNetworkInCloud(objKey string, networksToDelete map[string]string, client *clients.AviClient) {
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

func (t *T1LRNetworking) getClusterSpecificNSXTSegmentsinCloud(client *clients.AviClient, lsLRMap map[string]string, next ...string) error {
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
		if strings.HasPrefix(*sg.Name, fmt.Sprintf("avi-%s", lib.GetClusterName())) {
			lsLRMap[*sg.SegmentID] = *sg.Tier1ID
		}
	}
	if result.Next != "" {
		next_uri := strings.Split(result.Next, "/api/nsxtsegmentruntime")
		if len(next_uri) > 1 {
			nextPage := "/api/nsxtsegmentruntime" + next_uri[1]
			err = t.getClusterSpecificNSXTSegmentsinCloud(client, lsLRMap, nextPage)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (t *T1LRNetworking) removeStaleLRLSEntries(client *clients.AviClient, cloudModel models.Cloud, lslrmap map[string]string) ([]*models.Tier1LogicalRouterInfo, bool) {
	updatedRequired := false
	cloudTier1Lrs := cloudModel.NsxtConfiguration.DataNetworkConfig.Tier1SegmentConfig.Manual.Tier1Lrs
	dataNetworkTier1Lrs := make([]*models.Tier1LogicalRouterInfo, 0)
	cloudLRLSMap := make(map[string]string)
	err := t.getClusterSpecificNSXTSegmentsinCloud(client, cloudLRLSMap)
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

func (t *T1LRNetworking) NewLRLSFullSyncWorker() *utils.FullSyncThread {
	instantiateFullSyncWorker.Do(func() {
		worker = utils.NewFullSyncThread(time.Duration(lib.FullSyncInterval) * time.Second)
		worker.SyncFunction = t.SyncLSLRNetwork
		worker.QuickSyncFunction = func(qSync bool) error { return nil }
	})
	return worker
}

func getInfraSettingNameFromT1LR(lr string) string {
	arr := strings.Split(lr, "/")
	infraSettingName := arr[len(arr)-1]
	if strings.Contains(infraSettingName, ":") {
		infraSettingName = strings.Split(infraSettingName, ":")[1]
	}
	return strings.Replace(infraSettingName, "_", "-", -1)
}
