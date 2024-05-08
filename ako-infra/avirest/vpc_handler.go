package avirest

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/vmware/alb-sdk/go/models"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
)

type VPCHandler struct {
}

func (v *VPCHandler) AddNetworkInfoEventHandler(stopCh <-chan struct{}) {
	vpcEventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			utils.AviLog.Infof("VPC ADD Event")
			ScheduleQuickSync()
		},
		UpdateFunc: func(oldObj interface{}, newObj interface{}) {
			utils.AviLog.Infof("VPC UPDATE Event")
			ScheduleQuickSync()

		},
		DeleteFunc: func(obj interface{}) {
			utils.AviLog.Infof("VPC DELETE Event")
			ScheduleQuickSync()
		},
	}
	lib.GetDynamicInformers().VPCInformer.Informer().AddEventHandler(vpcEventHandler)
	go lib.GetDynamicInformers().VPCInformer.Informer().Run(stopCh)
	if !cache.WaitForCacheSync(stopCh,
		lib.GetDynamicInformers().VPCInformer.Informer().HasSynced) {
		runtime.HandleError(fmt.Errorf("timed out waiting for VPC caches to sync"))
	} else {
		utils.AviLog.Infof("Caches synced for VPC informer")
	}
}

func (v *VPCHandler) SyncLSLRNetwork() {
	vpcToSubnetMap, vpcToNSMap, err := lib.GetVPCs(lib.GetDynamicClientSet())
	if err != nil {
		utils.AviLog.Errorf("Failed to list VPCs, error: %s", err)
		return
	}
	utils.AviLog.Infof("Got VPC to NS Map: %v", vpcToNSMap)
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

	if len(vpcToSubnetMap) > 0 {
		cloudVPCToSubnetMap := make(map[string]string)
		for _, t1lr := range cloudModel.NsxtConfiguration.DataNetworkConfig.Tier1SegmentConfig.Manual.Tier1Lrs {
			cloudVPCToSubnetMap[*t1lr.Tier1LrID] = *t1lr.SegmentID
		}
		updateCloud := false
		//TODO: Remove Stale VPCs in Cloud
		for vpc, subnet := range vpcToSubnetMap {
			if val, ok := cloudVPCToSubnetMap[vpc]; !ok || val != subnet {
				updateCloud = true
				cloudVPCToSubnetMap[vpc] = subnet
			}
		}
		if !updateCloud {
			v.createInfraSettingAndAnnotateNS(vpcToNSMap)
			return
		}
		cloudLSLRList := make([]*models.Tier1LogicalRouterInfo, len(cloudVPCToSubnetMap))
		addLRInfo := func(vpc, subnet string, index int) {
			cloudLSLRList[index] = &models.Tier1LogicalRouterInfo{
				SegmentID: &subnet,
				Tier1LrID: &vpc,
			}
		}
		index := 0
		for vpc, subnet := range cloudVPCToSubnetMap {
			addLRInfo(vpc, subnet, index)
			index++
		}
		cloudModel.NsxtConfiguration.DataNetworkConfig.Tier1SegmentConfig.Manual.Tier1Lrs = cloudLSLRList
		vpcMode := true
		cloudModel.NsxtConfiguration.VpcMode = &vpcMode
		path := "/api/cloud/" + *cloudModel.UUID
		restOp := utils.RestOp{
			ObjName: utils.CloudName,
			Path:    path,
			Method:  utils.RestPut,
			Obj:     &cloudModel,
			Tenant:  lib.GetTenant(),
			Model:   "cloud",
		}
		executeRestOp("fullsync", client, &restOp)
	}
	v.createInfraSettingAndAnnotateNS(vpcToNSMap)
}

func (v *VPCHandler) createInfraSettingAndAnnotateNS(vpcToNSMap map[string]string) {
	infraSettingCRs, err := lib.AKOControlConfig().CRDInformers().AviInfraSettingInformer.Lister().List(labels.Set(nil).AsSelector())
	if err != nil {
		utils.AviLog.Errorf("Failed to list AviInfraSetting CRs, error: %s", err.Error())
		return
	}

	staleInfraSettingCRSet := make(map[string]struct{})
	for _, infraSettingCR := range infraSettingCRs {
		staleInfraSettingCRSet[infraSettingCR.Name] = struct{}{}
	}

	wg := sync.WaitGroup{}
	for vpc, ns := range vpcToNSMap {
		arr := strings.Split(vpc, "/vpcs/")
		infraSettingName := arr[len(arr)-1]
		delete(staleInfraSettingCRSet, infraSettingName)
		project := strings.Split(arr[0], "/projects/")[1]
		wg.Add(1)
		go func(vpc, ns string) {
			defer wg.Done()
			_, err := lib.CreateOrUpdateAviInfraSetting(infraSettingName, "", vpc, project)
			if err != nil {
				utils.AviLog.Errorf("failed to create aviInfraSetting, name: %s, error: %s", infraSettingName, err.Error())
			}
			utils.AviLog.Infof("Created AviInfraSetting: %s", infraSettingName)
			lib.AnnotateNamespaceWithInfraSetting(ns, infraSettingName)
		}(vpc, ns)
	}

	for infraSettingName := range staleInfraSettingCRSet {
		wg.Add(1)
		go func(name string) {
			err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().AviInfraSettings().Delete(context.TODO(), name, metav1.DeleteOptions{})
			if err != nil {
				utils.AviLog.Warnf("failed to delete aviInfraSetting, name: %s, error: %s", name, err.Error())
			}
			wg.Done()
		}(infraSettingName)
	}

	wg.Add(1)
	go func() {
		lib.AnnotateSystemNamespaceWithInfraSetting()
		wg.Done()
	}()
	wg.Wait()
	lib.RemoveInfraSettingAnnotationFromNamespaces(staleInfraSettingCRSet)
}

func (v *VPCHandler) NewLRLSFullSyncWorker() *utils.FullSyncThread {
	worker = utils.NewFullSyncThread(time.Duration(lib.FullSyncInterval) * time.Second)
	worker.SyncFunction = v.SyncLSLRNetwork
	worker.QuickSyncFunction = func(qSync bool) error { return nil }
	return worker
}
