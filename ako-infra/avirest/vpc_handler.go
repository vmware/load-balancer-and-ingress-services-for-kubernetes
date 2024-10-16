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

var defaultProject string

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
	nsToVPCMap, err := lib.GetVPCs()
	if err != nil {
		utils.AviLog.Errorf("Failed to list VPCs, error: %s", err)
		return
	}
	utils.AviLog.Infof("Got NS to VPC Map: %v", nsToVPCMap)
	v.createInfraSettingAndAnnotateNS(nsToVPCMap)
}

func (v *VPCHandler) createInfraSettingAndAnnotateNS(nsToVPCMap map[string]string) {
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
	wg := sync.WaitGroup{}
	for ns, vpc := range nsToVPCMap {
		arr := strings.Split(vpc, "/vpcs/")
		infraSettingName := strings.ReplaceAll(arr[len(arr)-1], "_", "-")
		delete(staleInfraSettingCRSet, infraSettingName)
		project := strings.Split(arr[0], "/projects/")[1]
		tenant, err := getTenantForProject(project)
		if err != nil {
			utils.AviLog.Warnf("failed to fetch admin tenant from Avi, error: %s", err.Error())
			continue
		}
		// multiple namespaces can use the same vpc, and there will always be only 1 infrasetting per vpc
		// so no need to attempt Infrasetting creation
		// just annotate the namespace with the infrasetting and tenant info
		if _, ok := processedInfraSettingCRSet[infraSettingName]; ok {
			lib.AnnotateNamespaceWithInfraSetting(ns, infraSettingName)
			lib.AnnotateNamespaceWithTenant(ns, tenant)
			continue
		}
		processedInfraSettingCRSet[infraSettingName] = struct{}{}
		wg.Add(1)
		go func(vpc, ns string) {
			defer wg.Done()
			_, err := lib.CreateOrUpdateAviInfraSetting(infraSettingName, "", vpc)
			if err != nil {
				utils.AviLog.Errorf("failed to create aviInfraSetting, name: %s, error: %s", infraSettingName, err.Error())
			} else {
				lib.AnnotateNamespaceWithInfraSetting(ns, infraSettingName)
			}
			lib.AnnotateNamespaceWithTenant(ns, tenant)
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

	wg.Wait()
	lib.RemoveInfraSettingAnnotationFromNamespaces(staleInfraSettingCRSet)
}

func (v *VPCHandler) NewLRLSFullSyncWorker() *utils.FullSyncThread {
	worker = utils.NewFullSyncThread(time.Duration(lib.FullSyncInterval) * time.Second)
	worker.SyncFunction = v.SyncLSLRNetwork
	worker.QuickSyncFunction = func(qSync bool) error { return nil }
	return worker
}

func getTenantForProject(project string) (string, error) {
	if defaultProject == "" {
		c := InfraAviClientInstance()
		uri := "api/tenant/admin"
		response := models.Tenant{}
		err := lib.AviGet(c, uri, &response)
		if err != nil {
			return "", err
		}
		for _, attr := range response.Attrs {
			if *attr.Key == "path" {
				projectSlice := strings.Split(*attr.Value, "/projects/")
				defaultProject = projectSlice[len(projectSlice)-1]
			}
		}
	}
	if project == defaultProject {
		return "admin", nil
	}
	return project, nil
}
