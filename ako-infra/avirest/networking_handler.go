package avirest

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/vmware/alb-sdk/go/clients"
	"github.com/vmware/alb-sdk/go/models"
	"github.com/vmware/alb-sdk/go/session"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/rest"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

type NetworkingHandler interface {
	AddNetworkInfoEventHandler(stopCh <-chan struct{})
	SyncLSLRNetwork()
	NewLRLSFullSyncWorker() *utils.FullSyncThread
}

var CloudCache *models.Cloud
var NetCache map[string]*models.Network
var IPAMCache *models.IPAMDNSProviderProfile

var worker *utils.FullSyncThread

func ScheduleQuickSync() {
	if worker != nil {
		worker.QuickSync()
	}
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

	if CloudCache.IPAMProviderRef == nil {
		utils.AviLog.Fatalf("IPAM Proivder not configured in the cloud %s", cloudName)
	}
	return AviIPAMCachePopulate(client, strings.Split(*CloudCache.IPAMProviderRef, "#")[1])
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

	restLayer := rest.NewRestOperations(nil, true)
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
