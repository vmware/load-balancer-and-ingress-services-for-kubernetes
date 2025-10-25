package avirest

import (
	"context"
	"fmt"
	"strings"
	"time"

	avicache "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
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
	vpcNetworkConfigEventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			utils.AviLog.Infof("VpcNetworkConfig ADD Event")
			ScheduleQuickSync()
		},
		UpdateFunc: func(oldObj interface{}, newObj interface{}) {
			utils.AviLog.Infof("VpcNetworkConfig UPDATE Event")
			ScheduleQuickSync()

		},
		DeleteFunc: func(obj interface{}) {
			utils.AviLog.Infof("VpcNetworkConfig DELETE Event")
			ScheduleQuickSync()
		},
	}
	lib.GetDynamicInformers().VPCNetworkConfigurationInformer.Informer().AddEventHandler(vpcNetworkConfigEventHandler)
	go lib.GetDynamicInformers().VPCNetworkConfigurationInformer.Informer().Run(stopCh)
	if !cache.WaitForCacheSync(stopCh,
		lib.GetDynamicInformers().VPCNetworkConfigurationInformer.Informer().HasSynced) {
		runtime.HandleError(fmt.Errorf("timed out waiting for VPC caches to sync"))
	} else {
		utils.AviLog.Infof("Caches synced for VPC informer")
	}
}

func (v *VPCHandler) SyncLSLRNetwork() {
	// Check controller uptime - if controller was rebooted/upgraded, AKO-infra will restart
	aviObjCache := avicache.SharedAviObjCache()
	if err := aviObjCache.AviClusterStatusPopulate(InfraAviClientInstance()); err != nil {
		utils.AviLog.Warnf("Failed to check controller cluster status: %v", err)
	}

	nsToVPCMap, err := lib.GetVPCs()
	if err != nil {
		utils.AviLog.Errorf("Failed to list VPCs, error: %s", err)
		return
	}
	utils.AviLog.Infof("Got NS to VPC Map: %v", nsToVPCMap)
	nsToSEGMap, err := lib.GetNSToSEGMap()
	if err != nil {
		utils.AviLog.Errorf("Failed to get NS to SEG Map, error: %s", err.Error())
		return
	}
	utils.AviLog.Infof("Got NS to SEG Map: %v", nsToSEGMap)
	v.createInfraSettingAndAnnotateNS(nsToVPCMap, nsToSEGMap)
}

func (v *VPCHandler) createInfraSettingAndAnnotateNS(nsToVPCMap, nsToSEGMap map[string]string) {
	infraSettingCRs, err := lib.AKOControlConfig().CRDInformers().AviInfraSettingInformer.Lister().List(labels.Set(nil).AsSelector())
	if err != nil {
		utils.AviLog.Errorf("Failed to list AviInfraSetting CRs, error: %s", err.Error())
		return
	}

	staleInfraSettingCRSet := make(map[string]struct{})
	for _, infraSettingCR := range infraSettingCRs {
		staleInfraSettingCRSet[infraSettingCR.Name] = struct{}{}
	}

	processedInfraSettingCRSet := make(map[string]struct{})
	nsxProjectToTenantMap, err := lib.GetNSXProjectToTenantMap(InfraAviClientInstance())
	if err != nil {
		utils.AviLog.Errorf("Failed to get NSX project to tenant map, skipping reconcilliation, error: %s", err.Error())
		return
	}

	for ns, vpc := range nsToVPCMap {
		arr := strings.Split(vpc, "/vpcs/")
		projectArr := strings.Split(arr[0], "/projects/")
		project := projectArr[len(projectArr)-1]
		tenant, ok := nsxProjectToTenantMap[project]
		if !ok {
			utils.AviLog.Warnf("Tenant not found for project %s", project)
			continue
		}

		name := project + arr[len(arr)-1]
		segName := "Default-Group"
		if seg, ok := nsToSEGMap[ns]; ok {
			name = name + seg
			segName = seg
		}
		infraSettingName := lib.GetAviInfraSettingName(name)

		// multiple namespaces can use the same vpc, and there will always be only 1 infrasetting per vpc
		// so no need to attempt Infrasetting creation
		// just annotate the namespace with the infrasetting and tenant info
		if _, ok := processedInfraSettingCRSet[infraSettingName]; ok {
			lib.AnnotateNamespaceWithTenantAndInfraSetting(ns, tenant, infraSettingName)
			continue
		}

		processedInfraSettingCRSet[infraSettingName] = struct{}{}
		delete(staleInfraSettingCRSet, infraSettingName)

		_, err = lib.CreateOrUpdateAviInfraSetting(infraSettingName, "", vpc, segName)
		if err != nil {
			utils.AviLog.Errorf("failed to create aviInfraSetting, name: %s, error: %s", infraSettingName, err.Error())
			continue
		}
		lib.AnnotateNamespaceWithTenantAndInfraSetting(ns, tenant, infraSettingName)
	}

	for infraSettingName := range staleInfraSettingCRSet {
		err := lib.AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().AviInfraSettings().Delete(context.TODO(), infraSettingName, metav1.DeleteOptions{})
		if err != nil {
			utils.AviLog.Warnf("failed to delete aviInfraSetting, name: %s, error: %s", infraSettingName, err.Error())
		} else {
			utils.AviLog.Infof("deleted aviInfraSetting, name: %s", infraSettingName)
		}
	}

	lib.RemoveInfraSettingAnnotationFromNamespaces(staleInfraSettingCRSet)
}

func (v *VPCHandler) NewLRLSFullSyncWorker() *utils.FullSyncThread {
	worker = utils.NewFullSyncThread(time.Duration(lib.FullSyncInterval) * time.Second)
	worker.SyncFunction = v.SyncLSLRNetwork
	worker.QuickSyncFunction = func(qSync bool) error { return nil }
	return worker
}
