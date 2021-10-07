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
	"strings"

	avicache "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/rest"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/third_party/github.com/vmware/alb-sdk/go/clients"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/third_party/github.com/vmware/alb-sdk/go/session"

	"github.com/vmware/alb-sdk/go/models"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var CloudCache *models.Cloud

// SyncLSLR fetches all networkinfo CR objects, compares them with the data network configured in the cloud,
// and updates the cloud if any LS-LR data is missing.
func SyncLSLR() {
	lslrmap := lib.GetNetinfoCRData()
	utils.AviLog.Infof("Got data LS LR Map: %v, from NetworkInfo CR", lslrmap)

	client := avicache.SharedAVIClients().AviClient[0]
	found, cloudModel := getAviCloudFromCache(client, utils.CloudName)
	if !found {
		utils.AviLog.Warnf("Failed to get Cloud data from cache")
		return
	}
	matched := matchSegmentInCloud(cloudModel.NsxtConfiguration.DataNetworkConfig.Tier1SegmentConfig.Manual.Tier1Lrs, lslrmap)
	if !matched {
		cloudModel.NsxtConfiguration.DataNetworkConfig.Tier1SegmentConfig.Manual.Tier1Lrs = constructLsLrInCloud(lslrmap)
		path := "/api/cloud/" + *cloudModel.UUID
		restOp := utils.RestOp{ObjName: utils.CloudName, Path: path, Method: utils.RestPut, Obj: &cloudModel,
			Tenant: "admin", Model: "cloud", Version: utils.CtrlVersion}
		executeRestOp("fullsync", client, &restOp)
	} else {
		utils.AviLog.Infof("LS LR update not required in cloud: %s", utils.CloudName)
	}
}

func AddSegment(obj interface{}) {
	objKey := "Netinfo" + utils.ObjKey(obj)
	utils.AviLog.Debugf("key: %s, Network Info DELETE Event", objKey)
	crd := obj.(*unstructured.Unstructured)

	specJSON, found, err := unstructured.NestedStringMap(crd.UnstructuredContent(), "topology")
	if err != nil || !found {
		utils.AviLog.Warnf("NetworkInfo topology not found: %+v", err)
		return
	}
	lr := specJSON["gatewayPath"]
	ls := specJSON["aviSegmentPath"]
	utils.AviLog.Infof("key: %s, Adding LR %s, LS %s from networkinfo CR", objKey, lr, ls)
	client := avicache.SharedAVIClients().AviClient[0]
	found, cloudModel := getAviCloudFromCache(client, utils.CloudName)
	if !found {
		utils.AviLog.Warnf("key: %s, Failed to get Cloud data from cache", objKey)
		return
	}
	updateRequired, lslrList := addSegmentInCloud(cloudModel.NsxtConfiguration.DataNetworkConfig.Tier1SegmentConfig.Manual.Tier1Lrs, lr, ls)
	if updateRequired {
		cloudModel.NsxtConfiguration.DataNetworkConfig.Tier1SegmentConfig.Manual.Tier1Lrs = lslrList
		path := "/api/cloud/" + *cloudModel.UUID
		restOp := utils.RestOp{ObjName: utils.CloudName, Path: path, Method: utils.RestPut, Obj: &cloudModel,
			Tenant: "admin", Model: "cloud", Version: utils.CtrlVersion}
		executeRestOp(objKey, client, &restOp)
	} else {
		utils.AviLog.Infof("key: %s, LSLR update not required in cloud: %s", objKey, utils.CloudName)
	}
}

func DeleteSegment(obj interface{}) {
	objKey := "Netinfo" + utils.ObjKey(obj)
	utils.AviLog.Debugf("key:%s, Network Info DELETE Event", objKey)
	crd := obj.(*unstructured.Unstructured)

	specJSON, found, err := unstructured.NestedStringMap(crd.UnstructuredContent(), "topology")
	if err != nil || !found {
		utils.AviLog.Warnf("key: %s, NetworkInfo topology not found: %v", objKey, err)
		return
	}
	lr := specJSON["gatewayPath"]
	ls := specJSON["aviSegmentPath"]
	utils.AviLog.Infof("key: %s, Network Info CR deleted, removing LR %s, LS %s from cloud", objKey, lr, ls)
	client := avicache.SharedAVIClients().AviClient[0]
	found, cloudModel := getAviCloudFromCache(client, utils.CloudName)
	if !found {
		utils.AviLog.Warnf("key: %s, Failed to get Cloud data from cache", objKey)
		return
	}
	updateRequired, lslrList := delSegmentInCloud(cloudModel.NsxtConfiguration.DataNetworkConfig.Tier1SegmentConfig.Manual.Tier1Lrs, lr, ls)
	if updateRequired {
		cloudModel.NsxtConfiguration.DataNetworkConfig.Tier1SegmentConfig.Manual.Tier1Lrs = lslrList
		path := "/api/cloud/" + *cloudModel.UUID
		restOp := utils.RestOp{ObjName: utils.CloudName, Path: path, Method: utils.RestPut, Obj: &cloudModel,
			Tenant: "admin", Model: "cloud", Version: utils.CtrlVersion}
		executeRestOp(objKey, client, &restOp)
	} else {
		utils.AviLog.Infof("key: %s, LSLR update not required in cloud: %s", objKey, utils.CloudName)
	}
}

func addSegmentInCloud(lslrList []*models.Tier1LogicalRouterInfo, lr, ls string) (bool, []*models.Tier1LogicalRouterInfo) {
	for i := range lslrList {
		if *lslrList[i].SegmentID == ls {
			if *lslrList[i].Tier1LrID == lr {
				return false, lslrList
			}
			lslrList = append(lslrList[:i], lslrList[i+1:]...)
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
	if len(lslrMap) != len(lslrList) {
		return false
	}
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

func executeRestOp(key string, client *clients.AviClient, restOp *utils.RestOp, retryNum ...int) {
	utils.AviLog.Debugf("key: %s, Executing rest operation to sync object in cloud: %v", key, *restOp)
	retry := 0
	if len(retryNum) > 0 {
		utils.AviLog.Infof("key: %s, Retrying to execute rest request", key)
		retry = retryNum[0]
		if retry >= 3 {
			utils.AviLog.Errorf("key %s, rest request execution aborted after reqtrying 3 times", key)
			return
		}
	}
	err := rest.AviRestOperateWrapper(client, []*utils.RestOp{restOp})
	if err != nil {
		utils.AviLog.Infof("key %s, Got error in executing rest request: %v", key, err)
		if checkAndRetry(key, err) {
			executeRestOp(key, client, restOp, retry+1)
		}
	}
	AviCloudCachePopulate(client, utils.CloudName)
	utils.AviLog.Infof("key: %s, Successfully executed rest operation to sync object in cloud: %v", key, *restOp)
}

func checkAndRetry(key string, err error) bool {
	if err == nil {
		return false
	}
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
